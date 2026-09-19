package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorsMiddleware_Wildcard(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware("*")(dummy)

	t.Run("OPTIONS preflight returns 204 with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/session", nil)
		req.Header.Set("Origin", "https://sistema-taller-automotriz-1.onrender.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://sistema-taller-automotriz-1.onrender.com" {
			t.Errorf("expected origin https://sistema-taller-automotriz-1.onrender.com, got %s", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("expected Access-Control-Allow-Methods header")
		}
	})

	t.Run("POST request includes CORS header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
		req.Header.Set("Origin", "https://sistema-taller-automotriz-1.onrender.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://sistema-taller-automotriz-1.onrender.com" {
			t.Errorf("expected origin https://sistema-taller-automotriz-1.onrender.com, got %s", got)
		}
	})
}

func TestCorsMiddleware_SpecificOrigin(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware("https://frontend.com")(dummy)

	t.Run("Matching origin allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/session", nil)
		req.Header.Set("Origin", "https://frontend.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://frontend.com" {
			t.Errorf("expected origin https://frontend.com, got %s", got)
		}
	})

	t.Run("Mismatched origin ignored", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/session", nil)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("expected no CORS header for untrusted origin, got %s", got)
		}
	})
}
