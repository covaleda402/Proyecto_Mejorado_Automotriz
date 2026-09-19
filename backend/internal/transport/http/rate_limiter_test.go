package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterExtractClientIP(t *testing.T) {
	limiter := NewLoginRateLimiterCustom(15*time.Minute, 5, 0, time.Now)
	defer limiter.Close()

	// 1. Untrusted remote address ignores X-Forwarded-For
	req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "198.51.100.25:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	ip := limiter.ExtractClientIP(req, nil)
	if ip != "198.51.100.25" {
		t.Fatalf("expected remote IP 198.51.100.25, got %q", ip)
	}

	// 2. Loopback remote address reads first IP from X-Forwarded-For
	req = httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.1")
	ip = limiter.ExtractClientIP(req, nil)
	if ip != "203.0.113.50" {
		t.Fatalf("expected XFF IP 203.0.113.50, got %q", ip)
	}

	// 3. Docker network remote address reads X-Forwarded-For
	req = httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "172.19.0.2:4000"
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	ip = limiter.ExtractClientIP(req, nil)
	if ip != "192.0.2.1" {
		t.Fatalf("expected XFF IP 192.0.2.1 from docker proxy, got %q", ip)
	}

	// 4. Custom trusted proxy
	req = httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	ip = limiter.ExtractClientIP(req, []string{"1.2.3.4"})
	if ip != "8.8.8.8" {
		t.Fatalf("expected XFF IP 8.8.8.8 from custom trusted proxy, got %q", ip)
	}
}

func TestRateLimiterWindowSlidingAndExpiry(t *testing.T) {
	currTime := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return currTime }

	limiter := NewLoginRateLimiterCustom(15*time.Minute, 5, 0, clock)
	defer limiter.Close()

	ip := "192.168.1.100"
	user := "alice"

	// 4 failures -> still allowed
	for i := 0; i < 4; i++ {
		limiter.RecordFailure(ip, user)
		allowed, _ := limiter.AllowAttempt(ip, user)
		if !allowed {
			t.Fatalf("should be allowed after %d failures", i+1)
		}
	}

	// 5th failure -> blocked
	limiter.RecordFailure(ip, user)
	allowed, retryAfter := limiter.AllowAttempt(ip, user)
	if allowed {
		t.Fatalf("expected blocked after 5 failures")
	}
	if retryAfter <= 0 {
		t.Fatalf("retryAfter should be positive, got %v", retryAfter)
	}

	// Advance time by 16 minutes -> should expire and be allowed
	currTime = currTime.Add(16 * time.Minute)
	allowed, _ = limiter.AllowAttempt(ip, user)
	if !allowed {
		t.Fatalf("expected allowed after window expiration")
	}
}
