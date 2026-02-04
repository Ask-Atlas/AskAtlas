package middleware

import (
	"io"
	"log/slog"
	"net/http"

	svix "github.com/svix/svix-webhooks/go"
)

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
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if err := wh.Verify(body, r.Header); err != nil {
				slog.Error("failed to verify svix signature", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

}
