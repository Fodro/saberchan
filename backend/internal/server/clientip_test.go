package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fodro/saberchan/config"
)

func TestClientIP_TrustedProxyUsesXFF(t *testing.T) {
	t.Parallel()
	s := &Server{
		conf:        &config.Config{},
		trustedNets: parseTrustedProxies("10.0.0.0/8"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.5")
	if got := s.clientIP(r); got != "203.0.113.9" {
		t.Fatalf("got %q", got)
	}
}

func TestClientIP_UntrustedIgnoresXFF(t *testing.T) {
	t.Parallel()
	s := &Server{
		conf:        &config.Config{},
		trustedNets: parseTrustedProxies("10.0.0.0/8"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.1:9999"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := s.clientIP(r); got != "203.0.113.1" {
		t.Fatalf("got %q", got)
	}
}

func TestClientIP_EmptyTrustedNeverTrustsXFF(t *testing.T) {
	t.Parallel()
	s := &Server{conf: &config.Config{}}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1"
	r.Header.Set("X-Forwarded-For", "9.9.9.9")
	if got := s.clientIP(r); got != "10.0.0.5" {
		t.Fatalf("got %q", got)
	}
}
