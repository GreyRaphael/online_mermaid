package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoginLimiter struct {
	mu          sync.Mutex
	failures    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

func NewLoginLimiter(maxAttempts int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{
		failures:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

func (l *LoginLimiter) Window() time.Duration {
	return l.window
}

func (l *LoginLimiter) clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func (l *LoginLimiter) Allow(r *http.Request) bool {
	ip := l.clientIP(r)
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	times := l.failures[ip]
	valid := make([]time.Time, 0, len(times))
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	l.failures[ip] = valid

	return len(valid) < l.maxAttempts
}

func (l *LoginLimiter) RecordFailure(r *http.Request) {
	ip := l.clientIP(r)
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.failures[ip] = append(l.failures[ip], now)
}
