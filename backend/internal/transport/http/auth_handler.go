package http

import (
	"net/http"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

// signInRequest is the credential payload the login screen sends.
type signInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// sessionResponse is what a successful sign in returns. It never carries the
// password hash.
type sessionResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
}

// AuthHandler exposes the sign in and sign out operations.
type AuthHandler struct {
	authenticate usecase.AuthenticateUser
	rateLimiter  *LoginRateLimiter
}

// NewAuthHandler wires the authentication handler.
func NewAuthHandler(authenticate usecase.AuthenticateUser, rateLimiter *LoginRateLimiter) AuthHandler {
	return AuthHandler{authenticate: authenticate, rateLimiter: rateLimiter}
}

// SignIn verifies the credentials, enforces rate limits, sets a secure session cookie and returns session data.
func (h AuthHandler) SignIn(writer http.ResponseWriter, request *http.Request) {
	var payload signInRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}

	key := RateLimitKey(request, payload.Username)
	if h.rateLimiter != nil && h.rateLimiter.IsBlocked(key) {
		failure(writer, domain.ErrTooManyRequests)
		return
	}

	session, err := h.authenticate.Execute(request.Context(), payload.Username, payload.Password)
	if err != nil {
		if h.rateLimiter != nil {
			h.rateLimiter.RecordFailure(key)
		}
		failure(writer, err)
		return
	}

	if h.rateLimiter != nil {
		h.rateLimiter.Reset(key)
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	respond(writer, http.StatusOK, sessionResponse{
		Token:     session.Token,
		ExpiresAt: formatTime(session.ExpiresAt),
		UserID:    session.UserID,
		Username:  session.Username,
		FullName:  session.FullName,
		Role:      string(session.Role),
	})
}

// SignOut clears the session cookie.
func (h AuthHandler) SignOut(writer http.ResponseWriter, request *http.Request) {
	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	respond(writer, http.StatusOK, map[string]string{"message": "Sesion cerrada."})
}
