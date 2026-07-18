package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Fodro/saberchan/internal/board"
	chi "github.com/go-chi/chi/v5"
)

func (s *Server) CreateBoard(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) DeleteBoard(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) GetBoards(w http.ResponseWriter, r *http.Request) {
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
