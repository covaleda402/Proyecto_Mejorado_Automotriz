package http

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LoginRateLimiter enforces rate limiting on login attempts to prevent brute force attacks.
// Standard configuration: 5 failed attempts per 15-minute window.
type LoginRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
	now         func() time.Time
}

// NewLoginRateLimiter creates a new rate limiter instance.
func NewLoginRateLimiter(maxAttempts int, window time.Duration, now func() time.Time) *LoginRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &LoginRateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
		now:         now,
	}
}

// IsBlocked checks if the key has exceeded the maximum allowed failed attempts.
func (l *LoginRateLimiter) IsBlocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-l.window)
	times := l.attempts[key]
	var active []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			active = append(active, t)
		}
	}
	if len(active) == 0 {
		delete(l.attempts, key)
		return false
	}
	l.attempts[key] = active
	return len(active) >= l.maxAttempts
}

// RecordFailure records a failed login attempt for the key.
func (l *LoginRateLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-l.window)
	times := l.attempts[key]
	var active []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			active = append(active, t)
		}
	}
	l.attempts[key] = append(active, l.now())
}

// Reset clears the record of failed attempts for the key (called upon successful login).
func (l *LoginRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

// RateLimitKey creates a tracking key based on client IP and username.
func RateLimitKey(request *http.Request, username string) string {
	ip := clientIP(request)
	return ip + ":" + strings.TrimSpace(strings.ToLower(username))
}

func clientIP(request *http.Request) string {
	if xff := request.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xrip := request.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return request.RemoteAddr
}
