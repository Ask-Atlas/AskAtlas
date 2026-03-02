// Package middleware provides reusable HTTP interceptors for routing and authentication.
package middleware

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// ClerkAuth creates an HTTP middleware that extracts the Clerk user session,
// validates it, and resolves the database user ID injecting it into the context.
func ClerkAuth(resolver authctx.UserIDResolver) func(http.Handler) http.Handler {
	clerkMiddleware := clerkhttp.WithHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apperrors.RespondWithError(w, apperrors.NewUnauthorized())
		})),
	)

	return func(next http.Handler) http.Handler {
		return clerkMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				slog.Warn("ClerkAuth: session claims missing from context")
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

			clerkUserID := claims.Subject
			if clerkUserID == "" {
				slog.Warn("ClerkAuth: empty subject in session claims")
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

			userID, err := resolver.GetUserIDByClerkID(r.Context(), clerkUserID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					slog.Warn("ClerkAuth: user not found in database", "clerk_id", clerkUserID)
					apperrors.RespondWithError(w, apperrors.NewUnauthorized())
					return
				}
				slog.Error("ClerkAuth: failed to resolve user ID", "clerk_id", clerkUserID, "error", err)
				apperrors.RespondWithError(w, apperrors.NewInternalError())
				return
			}

			ctx := authctx.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}
