package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	svix "github.com/svix/svix-webhooks/go"
)

// SVIXVerifier creates an HTTP middleware that validates Svix webhook signatures.
// It ensures that incoming webhook requests genuinely originate from Clerk.
func SVIXVerifier(secret string) func(next http.Handler) http.Handler {
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		panic("failed to create svix webhook: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("failed to read request body", "error", err)
				apperrors.RespondWithError(w, apperrors.NewBadRequest("Bad Request", nil))
				return
			}

			if err := wh.Verify(body, r.Header); err != nil {
				slog.Error("failed to verify svix signature", "error", err)
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			next.ServeHTTP(w, r)
		})
	}

}
