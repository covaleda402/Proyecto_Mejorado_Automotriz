package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"workshop/internal/domain"
)

type contextKey string

const callerContextKey contextKey = "caller"

// caller is the authenticated identity of the current request.
type caller struct {
	UserID string
	Role   domain.Role
}

// userReader is the narrow port the middleware needs to check server-side account active state.
type userReader interface {
	FindByID(ctx context.Context, id string) (domain.User, error)
}

// authMiddleware rejects a request without a valid session token, verifies that the account
// is still active in the database (BR-ACTIVE-02), and enforces Fail-Closed policy on DB errors.
func authMiddleware(issuer TokenIssuer, users userReader, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			header := request.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				failure(writer, domain.ErrUnauthorized)
				return
			}
			claim, err := issuer.Verify(strings.TrimPrefix(header, "Bearer "), now())
			if err != nil {
				failure(writer, domain.ErrUnauthorized)
				return
			}

			// BR-ACTIVE-02: Server-side check against primary store with Fail-Closed semantics
			user, err := users.FindByID(request.Context(), claim.UserID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					failure(writer, domain.ErrAccountInactive)
					return
				}
				failure(writer, domain.ErrAuthorizationStateUnavailable)
				return
			}
			if !user.IsActive {
				failure(writer, domain.ErrAccountInactive)
				return
			}

			identity := caller{UserID: user.ID, Role: user.Role}
			next.ServeHTTP(writer, request.WithContext(
				context.WithValue(request.Context(), callerContextKey, identity),
			))
		})
	}
}

// callerFrom reads the authenticated identity a handler runs under.
func callerFrom(ctx context.Context) (caller, error) {
	identity, ok := ctx.Value(callerContextKey).(caller)
	if !ok || identity.UserID == "" {
		return caller{}, domain.ErrUnauthorized
	}
	return identity, nil
}

// requireAdministrator refuses a caller who is not the workshop manager.
func requireAdministrator(ctx context.Context) (caller, error) {
	identity, err := callerFrom(ctx)
	if err != nil {
		return caller{}, err
	}
	if identity.Role != domain.RoleAdministrator {
		return caller{}, fmt.Errorf("%w: this operation belongs to the administrator", domain.ErrForbidden)
	}
	return identity, nil
}
