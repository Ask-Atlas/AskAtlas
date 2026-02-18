package middleware

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// ClerkAuth validates the Clerk JWT, resolves the Clerk user ID to an internal
// UUID, and stores it in the request context. Returns 401 on failure.
func ClerkAuth(resolver authctx.UserIDResolver) func(http.Handler) http.Handler {
	clerkMiddleware := clerkhttp.RequireHeaderAuthorization()

	return func(next http.Handler) http.Handler {
		return clerkMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				slog.Warn("ClerkAuth: session claims missing from context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			clerkUserID := claims.Subject
			if clerkUserID == "" {
				slog.Warn("ClerkAuth: empty subject in session claims")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := resolver.GetUserIDByClerkID(r.Context(), clerkUserID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					slog.Warn("ClerkAuth: user not found in database",
						"clerk_id", clerkUserID,
					)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				slog.Error("ClerkAuth: failed to resolve user ID",
					"clerk_id", clerkUserID,
					"error", err,
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := authctx.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}
