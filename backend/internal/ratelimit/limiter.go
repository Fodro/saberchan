package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a simple fixed-window per-key rate limiter (in-process).
// Fine for a single backend replica on a VPS.
type Limiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	buckets map[string]*bucket
}

type bucket struct {
	count int
	start time.Time
}

func New(window time.Duration, max int) *Limiter {
	if window <= 0 {
		window = time.Minute
	}
	if max <= 0 {
		max = 30
	}
	return &Limiter{
		window:  window,
		max:     max,
		buckets: make(map[string]*bucket),
	}
}

// Allow reports whether key may proceed and records the hit when allowed.
func (l *Limiter) Allow(key string) bool {
	if key == "" {
		key = "unknown"
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.Sub(b.start) >= l.window {
		l.buckets[key] = &bucket{count: 1, start: now}
		return true
	}
	if b.count >= l.max {
		return false
	}
	b.count++
	return true
}
