package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/ban"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/captcha"
	"github.com/Fodro/saberchan/internal/follow"
	"github.com/Fodro/saberchan/internal/health"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	srv     *http.Server
	conf    *config.Config
	board   board.Service
	captcha captcha.Service
	health  health.Service
	ban     ban.Service
	follow  follow.Service
}

func NewServer(conf *config.Config, board board.Service, captcha captcha.Service, health health.Service, ban ban.Service, followSvc follow.Service) *Server {
	return &Server{
		srv: &http.Server{
			Addr: ":" + conf.Port,
		},
		conf:    conf,
		board:   board,
		captcha: captcha,
		health:  health,
		ban:     ban,
		follow:  followSvc,
	}
}

func writeJSONError(w http.ResponseWriter, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
		"code":  code,
	})
}

// writeBannedError responds 403 with the ban's reason/until alongside the
// usual error/code fields, so clients can show why they were blocked.
func writeBannedError(w http.ResponseWriter, rec *ban.Record) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":  "banned",
		"code":   "banned",
		"reason": rec.Reason,
		"until":  rec.Until.Format(time.RFC3339),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// checkBanned reports whether the request's client IP or fingerprint is
// banned, writing the 403 banned response and returning true if so. The IP
// used for the ban check is always the real client IP (regardless of
// StoreIp), since the ban gate must work even when IPs aren't persisted.
func (s *Server) checkBanned(w http.ResponseWriter, r *http.Request, fingerprint string) bool {
	if s.ban == nil {
		return false
	}
	rec, err := s.ban.Check(r.Context(), clientIP(r), fingerprint)
	if err != nil {
		log.Printf("failed to check ban: %v", err)
		return false
	}
	if rec == nil {
		return false
	}
	writeBannedError(w, rec)
	return true
}

func (s *Server) applyClientIP(r *http.Request, post *board.Post) {
	if post == nil {
		return
	}
	if s.conf.StoreIp {
		post.IP = clientIP(r)
	} else {
		post.IP = ""
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	s.srv.Handler = r

	// healtcheck
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		if err := s.health.Readiness(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {

			r.Route("/board", func(r chi.Router) {
				r.Post("/", s.CreateBoard)
				r.Get("/", s.GetBoards)
				r.Delete("/{id}", s.DeleteBoard)
				r.Post("/{id}/restore", s.RestoreBoard)
				r.Get("/{alias}", s.GetBoardByAlias)
			})

			r.Route("/thread", func(r chi.Router) {
				r.Post("/", s.CreateThread)
				r.Get("/{id}", s.GetThread)
				r.Delete("/{id}", s.DeleteThread)
				r.Post("/{id}/restore", s.RestoreThread)
			})

			r.Route("/post", func(r chi.Router) {
				r.Post("/{thread_id}", s.CreatePost)
				r.Delete("/{id}", s.DeletePost)
				r.Post("/{id}/restore", s.RestorePost)
				r.Post("/{id}/ban", s.BanPost)
			})

			r.Route("/captcha", func(r chi.Router) {
				r.Get("/", s.GenerateCaptcha)
				r.Post("/", s.ValidateCaptcha)
			})

			r.Route("/follow", func(r chi.Router) {
				r.Get("/", s.GetFollowStatus)
				r.Post("/{id}", s.FollowThread)
			})
		})
	})

	return s.srv.ListenAndServe()
}

// writeSoftDeleteError maps the shared soft-delete/restore sentinel errors to
// HTTP status codes for the DELETE/restore admin endpoints.
func writeSoftDeleteError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, board.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, board.ErrRestoreExpired):
		status = http.StatusConflict
		code = "restore_expired"
	case errors.Is(err, board.ErrAlreadyDeleted):
		status = http.StatusConflict
		code = "already_deleted"
	}
	writeJSONError(w, status, err, code)
}

func parseIDParam(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
