package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeUserStore struct {
	user map[string]domain.User
}

func newFakeUserStore(t *testing.T, username, password string) *fakeUserStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hashing the fixture password failed: %v", err)
	}
	user, err := domain.NewUser("user-1", username, string(hash), "Administrador del taller", domain.RoleAdministrator, testMoment)
	if err != nil {
		t.Fatalf("building the user fixture failed: %v", err)
	}
	return &fakeUserStore{user: map[string]domain.User{username: user}}
}

func (f *fakeUserStore) FindByUsername(_ context.Context, username string) (domain.User, error) {
	found, ok := f.user[username]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return found, nil
}

func (f *fakeUserStore) FindByID(_ context.Context, id string) (domain.User, error) {
	for _, item := range f.user {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func signIn(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewAuthHandler(usecase.NewAuthenticateUser(
		newFakeUserStore(t, "admin", "Admin2026"), testIssuer(), testClock(),
	), nil)
	request := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.SignIn(recorder, request)
	return recorder
}

func TestSignInReturnsATokenForValidCredentials(t *testing.T) {
	recorder := signIn(t, `{"username":"admin","password":"Admin2026"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("valid credentials must answer 200, got %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload sessionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("the response must be a session payload: %v", err)
	}
	if payload.Token == "" || payload.Role != string(domain.RoleAdministrator) {
		t.Fatalf("the session must carry a token and the role, got %+v", payload)
	}
	if _, err := testIssuer().Verify(payload.Token, testMoment.Add(time.Minute)); err != nil {
		t.Fatalf("the issued token must verify: %v", err)
	}
}

func TestSignInRejectsAWrongPasswordInSpanishWithoutAToken(t *testing.T) {
	recorder := signIn(t, `{"username":"admin","password":"wrong"}`)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("a wrong password must answer 401, got %d", recorder.Code)
	}
	var payload errorPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("the response must be an error payload: %v", err)
	}
	if payload.Message != "Usuario o contrasena incorrectos." {
		t.Fatalf("the message shown to the user must be in Spanish, got %q", payload.Message)
	}
	if strings.Contains(recorder.Body.String(), "token") {
		t.Fatal("a failed sign in must not return a token")
	}
}

func TestSignInRejectsAMalformedBody(t *testing.T) {
	recorder := signIn(t, "not-json")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("a malformed body must answer 400, got %d", recorder.Code)
	}
	var payload errorPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("the response must be an error payload: %v", err)
	}
	if payload.Message == "" || strings.Contains(payload.Message, "json") {
		t.Fatalf("the message must be sanitized and in Spanish, got %q", payload.Message)
	}
}

func TestSignInNeverReturnsThePasswordHash(t *testing.T) {
	recorder := signIn(t, `{"username":"admin","password":"Admin2026"}`)
	if strings.Contains(recorder.Body.String(), "$2a$") {
		t.Fatal("the session response must never carry the password hash")
	}
}

func TestSignInSetsSessionCookie(t *testing.T) {
	recorder := signIn(t, `{"username":"admin","password":"Admin2026"}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	cookie := recorder.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookie {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}
	if !sessionCookie.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if sessionCookie.Value == "" {
		t.Fatal("session cookie value must not be empty")
	}
}

func TestSignInRateLimiterBlocksAfterFiveFailedAttempts(t *testing.T) {
	limiter := NewLoginRateLimiter(5, 15*time.Minute, testClock())
	handler := NewAuthHandler(usecase.NewAuthenticateUser(
		newFakeUserStore(t, "admin", "Admin2026"), testIssuer(), testClock(),
	), limiter)

	// 5 failed attempts allowed
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(`{"username":"admin","password":"wrong"}`))
		rec := httptest.NewRecorder()
		handler.SignIn(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d should return 401, got %d", i, rec.Code)
		}
	}

	// 6th attempt should be blocked with 429 Too Many Requests
	req := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(`{"username":"admin","password":"wrong"}`))
	rec := httptest.NewRecorder()
	handler.SignIn(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th attempt must be blocked with 429, got %d", rec.Code)
	}

	var payload errorPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode error payload: %v", err)
	}
	if payload.Code != "too_many_requests" {
		t.Fatalf("expected code too_many_requests, got %q", payload.Code)
	}
}
