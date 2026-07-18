package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	t.Parallel()
	l := New(time.Minute, 2)
	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("first two should allow")
	}
	if l.Allow("a") {
		t.Fatal("third should deny")
	}
	if !l.Allow("b") {
		t.Fatal("other key should allow")
	}
}
