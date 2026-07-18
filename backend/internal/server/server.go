package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/ban"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/captcha"
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
}

func NewServer(conf *config.Config, board board.Service, captcha captcha.Service, health health.Service, ban ban.Service) *Server {
	return &Server{
		srv: &http.Server{
			Addr: ":" + conf.Port,
		},
		conf:    conf,
		board:   board,
		captcha: captcha,
		health:  health,
		ban:     ban,
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
		})
	})

	return s.srv.ListenAndServe()
}

func (s *Server) CreateBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	var boardIn board.BoardInput
	if err := json.NewDecoder(r.Body).Decode(&boardIn); err != nil {
		log.Printf("failed to decode board: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	err := s.board.CreateBoard(r.Context(), &boardIn)
	if err != nil {
		log.Printf("failed to create board: %v", err)
		if errors.Is(err, board.ErrNotImplemented) {
			writeJSONError(w, http.StatusNotImplemented, err, "not_implemented")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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

func (s *Server) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.DeleteBoard(r.Context(), id); err != nil {
		log.Printf("failed to delete board %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) RestoreBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.RestoreBoard(r.Context(), id); err != nil {
		log.Printf("failed to restore board %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) DeleteThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.DeleteThread(r.Context(), id); err != nil {
		log.Printf("failed to delete thread %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) RestoreThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.RestoreThread(r.Context(), id); err != nil {
		log.Printf("failed to restore thread %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.DeletePost(r.Context(), id); err != nil {
		log.Printf("failed to delete post %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) RestorePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.board.RestorePost(r.Context(), id); err != nil {
		log.Printf("failed to restore post %s: %v", id, err)
		writeSoftDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type banPostRequest struct {
	Reason   string `json:"reason"`
	Duration string `json:"duration"`
}

// BanPost bans the author (IP if known, else fingerprint) of a post and
// soft-deletes it. Admin only.
func (s *Server) BanPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if !s.requireAdmin(w, r) {
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	var req banPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode ban request: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	duration, err := ban.ParseDuration(req.Duration)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if err := s.ban.BanPost(r.Context(), id, req.Reason, duration); err != nil {
		log.Printf("failed to ban post %s: %v", id, err)
		writeBanPostError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// writeBanPostError maps ban.Service.BanPost's sentinel errors to HTTP
// status codes.
func writeBanPostError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, ban.ErrPostNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, ban.ErrAlreadyDeleted):
		status = http.StatusConflict
		code = "already_deleted"
	case errors.Is(err, ban.ErrReasonRequired), errors.Is(err, ban.ErrNoTarget), errors.Is(err, ban.ErrNoIdentifier):
		status = http.StatusBadRequest
		code = "bad_request"
	}
	writeJSONError(w, status, err, code)
}

func (s *Server) GetBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	includeDeleted := s.isAdminRequest(r)
	boards, err := s.board.GetBoards(r.Context(), includeDeleted)
	if err != nil {
		log.Printf("failed to get boards: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(boards); err != nil {
		log.Printf("failed to encode boards: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) GetBoardByAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	alias := chi.URLParam(r, "alias")

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	includeDeleted := s.isAdminRequest(r)
	boardRes, err := s.board.GetBoardWithThreads(r.Context(), alias, limit, offset, includeDeleted)
	if err != nil {
		log.Printf("failed to get board: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(boardRes); err != nil {
		log.Printf("failed to encode board: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var thread board.Thread
	if isMultipart(r) {
		parsed, err := parseMultipartThread(r)
		if err != nil {
			log.Printf("failed to parse multipart thread: %v", err)
			writeJSONError(w, http.StatusBadRequest, err, "bad_request")
			return
		}
		thread = *parsed
	} else if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		log.Printf("failed to decode thread: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	// Never trust client is_admin — only the shared admin token grants privileges.
	thread.IsAdmin = s.isAdminRequest(r)
	var fingerprint string
	if thread.OriginalPost != nil {
		fingerprint = thread.OriginalPost.BrowserFingerprint
	}
	if s.checkBanned(w, r, fingerprint) {
		return
	}
	s.applyClientIP(r, thread.OriginalPost)
	res, err := s.board.CreateThread(r.Context(), &thread)
	if err != nil {
		log.Printf("failed to create thread: %v", err)
		status := http.StatusInternalServerError
		code := "internal_error"
		if errors.Is(err, board.ErrBoardLocked) {
			status = http.StatusForbidden
			code = "board_locked"
		} else if errors.Is(err, pgx.ErrNoRows) {
			status = http.StatusNotFound
			code = "not_found"
		} else if errors.Is(err, board.ErrNotImplemented) {
			status = http.StatusNotImplemented
			code = "not_implemented"
		}
		writeJSONError(w, status, err, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (s *Server) GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	id := chi.URLParam(r, "id")
	convertedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	includeDeleted := s.isAdminRequest(r)
	thread, err := s.board.GetThreadWithPosts(r.Context(), convertedId, includeDeleted)
	if err != nil {
		log.Printf("failed to get thread: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(thread); err != nil {
		log.Printf("failed to encode thread: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	threadID := chi.URLParam(r, "thread_id")
	convertedThreadID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("failed to parse thread_id: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	var post board.Post
	if isMultipart(r) {
		parsed, err := parseMultipartPost(r)
		if err != nil {
			log.Printf("failed to parse multipart post: %v", err)
			writeJSONError(w, http.StatusBadRequest, err, "bad_request")
			return
		}
		post = *parsed
	} else if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Printf("failed to decode post: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if s.checkBanned(w, r, post.BrowserFingerprint) {
		return
	}
	s.applyClientIP(r, &post)
	if err := s.board.CreatePost(r.Context(), convertedThreadID, &post); err != nil {
		log.Printf("failed to create post: %v", err)
		status := http.StatusInternalServerError
		code := "internal_error"
		if errors.Is(err, board.ErrBoardLocked) {
			status = http.StatusForbidden
			code = "board_locked"
		} else if errors.Is(err, board.ErrNotImplemented) {
			status = http.StatusNotImplemented
			code = "not_implemented"
		}
		writeJSONError(w, status, err, code)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GenerateCaptcha(w http.ResponseWriter, r *http.Request) {
	data, token, err := s.captcha.Generate(r.Context())
	if err != nil {
		log.Printf("failed to generate captcha: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Add("x-captcha-token", token)
	data.WriteImage(w)
}

func (s *Server) ValidateCaptcha(w http.ResponseWriter, r *http.Request) {
	var captchaReq captcha.CaptchaValidateReq
	if err := json.NewDecoder(r.Body).Decode(&captchaReq); err != nil {
		log.Printf("failed to decode captcha req: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}

	passed, _ := s.captcha.Validate(r.Context(), captchaReq.Input, captchaReq.Token)

	resp := captcha.CaptchaValidateResp{
		Passed: passed,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
