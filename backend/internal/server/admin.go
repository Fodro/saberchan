package server

import (
	"crypto/subtle"
	"errors"
	"net/http"
)

const adminTokenHeader = "X-Admin-Token"

var errUnauthorized = errors.New("unauthorized")

// isAdminRequest is true only when the shared ADMIN_API_TOKEN is configured
// and matches the request header. Client-supplied is_admin fields are ignored.
func (s *Server) isAdminRequest(r *http.Request) bool {
	want := s.conf.AdminAPIToken
	if want == "" {
		return false
	}
	got := r.Header.Get(adminTokenHeader)
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if s.isAdminRequest(r) {
		return true
	}
	writeJSONError(w, http.StatusUnauthorized, errUnauthorized, "unauthorized")
	return false
}
