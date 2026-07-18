package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Fodro/saberchan/internal/ban"
	"github.com/Fodro/saberchan/internal/board"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

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
