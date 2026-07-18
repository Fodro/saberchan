package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Fodro/saberchan/internal/follow"
	"github.com/google/uuid"
)

func (s *Server) FollowThread(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	if s.follow == nil {
		writeJSONError(w, http.StatusNotImplemented, errors.New("follow not configured"), "not_implemented")
		return
	}
	err = s.follow.Follow(r.Context(), id)
	if errors.Is(err, follow.ErrDead) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"alive": false})
		return
	}
	if err != nil {
		log.Printf("failed to follow thread %s: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"alive": true})
}

func (s *Server) GetFollowStatus(w http.ResponseWriter, r *http.Request) {
	if s.follow == nil {
		writeJSONError(w, http.StatusNotImplemented, errors.New("follow not configured"), "not_implemented")
		return
	}
	raw := r.URL.Query().Get("ids")
	var ids []uuid.UUID
	if raw != "" {
		parts := strings.Split(raw, ",")
		ids = make([]uuid.UUID, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := uuid.Parse(part)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, err, "bad_request")
				return
			}
			ids = append(ids, id)
		}
	}
	statuses, err := s.follow.Status(r.Context(), ids)
	if err != nil {
		log.Printf("failed to get follow status: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		log.Printf("failed to encode follow status: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
	}
}
