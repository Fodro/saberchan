package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Fodro/saberchan/internal/board"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Server) DeleteThread(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	if !s.checkRateLimit(w, r, s.limitWrite) {
		return
	}
	var thread board.Thread
	var captchaInput, captchaToken string
	var err error
	if isMultipart(r) {
		parsed, err := parseMultipartThread(r)
		if err != nil {
			log.Printf("failed to parse multipart thread: %v", err)
			writeJSONError(w, http.StatusBadRequest, err, "bad_request")
			return
		}
		thread = *parsed
		captchaInput = r.FormValue("captcha_input")
		captchaToken = r.FormValue("captcha_token")
	} else if captchaInput, captchaToken, err = decodeJSONWithCaptcha(r, &thread); err != nil {
		log.Printf("failed to decode thread: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if !s.requireCaptcha(w, r, captchaInput, captchaToken) {
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
