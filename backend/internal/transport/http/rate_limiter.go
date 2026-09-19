package http

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LoginRateLimiter protects the authentication endpoint against credential stuffing
// and brute force attacks using a sliding window counter by client IP and combined IP:username.
type LoginRateLimiter struct {
	mu          sync.Mutex
	failures    map[string][]time.Time
	window      time.Duration
	maxFailures int
	ticker      *time.Ticker
	stopChan    chan struct{}
	now         func() time.Time
}

// NewLoginRateLimiter creates a LoginRateLimiter with the default 15-minute window,
// 5 maximum failed attempts, and a 5-minute cleanup ticker.
func NewLoginRateLimiter() *LoginRateLimiter {
	return NewLoginRateLimiterCustom(15*time.Minute, 5, 5*time.Minute, time.Now)
}

// NewLoginRateLimiterCustom builds a LoginRateLimiter with custom parameters for testing.
func NewLoginRateLimiterCustom(
	window time.Duration,
	maxFailures int,
	cleanupInterval time.Duration,
	now func() time.Time,
) *LoginRateLimiter {
	if now == nil {
		now = time.Now
	}
	limiter := &LoginRateLimiter{
		failures:    make(map[string][]time.Time),
		window:      window,
		maxFailures: maxFailures,
		now:         now,
		stopChan:    make(chan struct{}),
	}
	if cleanupInterval > 0 {
		limiter.ticker = time.NewTicker(cleanupInterval)
		go func() {
			for {
				select {
				case <-limiter.ticker.C:
					limiter.cleanup()
				case <-limiter.stopChan:
					return
				}
			}
		}()
	}
	return limiter
}

// Close stops the background cleanup ticker cleanly.
func (l *LoginRateLimiter) Close() {
	if l.ticker != nil {
		l.ticker.Stop()
	}
	select {
	case <-l.stopChan:
	default:
		close(l.stopChan)
	}
}

// cleanup purges expired failure records to prevent memory leaks.
func (l *LoginRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-l.window)
	for key, times := range l.failures {
		valid := make([]time.Time, 0, len(times))
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(l.failures, key)
		} else {
			l.failures[key] = valid
		}
	}
}

// AllowAttempt performs a pre-check before authenticating. If either the client IP
// or the combined IP:username key has reached or exceeded 5 failures in the 15-minute window,
// it returns false and the duration until the rate limit expires.
func (l *LoginRateLimiter) AllowAttempt(ip, username string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	var maxRetryAfter time.Duration
	blocked := false

	keys := make([]string, 0, 2)
	if ip != "" {
		keys = append(keys, ip)
	}
	if ip != "" && username != "" {
		keys = append(keys, ip+":"+username)
	}

	for _, key := range keys {
		times, exists := l.failures[key]
		if !exists {
			continue
		}

		// Filter active attempts within window
		valid := make([]time.Time, 0, len(times))
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		l.failures[key] = valid

		if len(valid) >= l.maxFailures {
			blocked = true
			oldestBlocking := valid[len(valid)-l.maxFailures]
			retryAfter := oldestBlocking.Add(l.window).Sub(now)
			if retryAfter > maxRetryAfter {
				maxRetryAfter = retryAfter
			}
		}
	}

	if blocked {
		if maxRetryAfter < time.Second {
			maxRetryAfter = time.Second
		}
		return false, maxRetryAfter
	}

	return true, 0
}

// RecordFailure is called ONLY when an authentication attempt fails due to invalid credentials.
func (l *LoginRateLimiter) RecordFailure(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	keys := make([]string, 0, 2)
	if ip != "" {
		keys = append(keys, ip)
	}
	if ip != "" && username != "" {
		keys = append(keys, ip+":"+username)
	}

	for _, key := range keys {
		times := l.failures[key]
		valid := make([]time.Time, 0, len(times)+1)
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		valid = append(valid, now)
		l.failures[key] = valid
	}
}

// Reset clears failure counters for the IP and combined key upon a successful login.
func (l *LoginRateLimiter) Reset(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if ip != "" && username != "" {
		delete(l.failures, ip+":"+username)
	}
	if ip != "" {
		delete(l.failures, ip)
	}
}

// ExtractClientIP extracts the real client IP. If r.RemoteAddr is a trusted proxy
// (e.g. 127.0.0.1, ::1, docker internal network 172.x, private IPs, or cloud CGNAT 100.64.0.0/10),
// it reads the client IP from Cloudflare/Render headers (CF-Connecting-IP, Render-Client-IP) or X-Forwarded-For.
func (l *LoginRateLimiter) ExtractClientIP(r *http.Request, trustedProxies []string) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = strings.TrimSpace(r.RemoteAddr)
	}

	if isTrustedProxy(remoteHost, trustedProxies) {
		for _, header := range []string{"CF-Connecting-IP", "Render-Client-IP", "X-Real-IP"} {
			if val := strings.TrimSpace(r.Header.Get(header)); val != "" {
				if ip := net.ParseIP(val); ip != nil {
					return val
				}
			}
		}

		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			parts := strings.Split(xff, ",")
			for _, part := range parts {
				candidate := strings.TrimSpace(part)
				if ip := net.ParseIP(candidate); ip != nil {
					return candidate
				}
			}
		}
	}

	return remoteHost
}

func isTrustedProxy(host string, trustedProxies []string) bool {
	for _, trusted := range trustedProxies {
		trusted = strings.TrimSpace(trusted)
		if trusted == "" {
			continue
		}
		if host == trusted {
			return true
		}
		if _, ipNet, err := net.ParseCIDR(trusted); err == nil {
			if ip := net.ParseIP(host); ip != nil && ipNet.Contains(ip) {
				return true
			}
		}
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	if ip.IsLoopback() || host == "127.0.0.1" || host == "::1" {
		return true
	}

	if strings.HasPrefix(host, "172.") {
		return true
	}

	if ip.IsPrivate() {
		return true
	}

	// 100.64.0.0/10 Carrier-Grade NAT (used by Render, AWS, container fabrics)
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 100 && (ip4[1]&0xC0) == 64 {
		return true
	}

	return false
}
