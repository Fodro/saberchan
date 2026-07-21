package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type MetricResponse struct {
	Boards []BoardMetrics `json:"boards"`
}

type BoardMetrics struct {
	Alias         string `json:"alias"`
	PostCount     uint64 `json:"post_count"`
	DeletedCount  uint64 `json:"deleted_count"`
	SageCount     uint64 `json:"sage_count"`
	ThreadCount   uint64 `json:"thread_count"`
}

func (s *Server) GetMetricPosts(w http.ResponseWriter, r *http.Request) {
	// Admin auth check
	if !s.isAdminRequest(r) {
		writeJSONError(w, http.StatusUnauthorized, errUnauthorized, "unauthorized")
		return
	}

	// Parse query params
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("missing from/to params"), "bad_request")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}

	// Call service
	metrics, err := s.board.GetBoardMetrics(r.Context(), from, to)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}

	// Convert to response format
	resp := MetricResponse{Boards: make([]BoardMetrics, len(metrics))}
	for i, m := range metrics {
		resp.Boards[i] = BoardMetrics{
			Alias:         m.BoardAlias,
			PostCount:     m.PostCount,
			DeletedCount:  m.DeletedCount,
			SageCount:     m.SageCount,
			ThreadCount:   m.ThreadCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode metric response: %v", err)
	}
}