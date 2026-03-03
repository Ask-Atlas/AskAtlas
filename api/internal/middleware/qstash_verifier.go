package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/upstash/qstash-go"
)

// maxQStashBodySize is the maximum allowed request body size for QStash webhooks (1 MB).
const maxQStashBodySize = 1 << 20

// QStashVerifier creates an HTTP middleware that validates QStash webhook signatures.
// currentSigningKey and nextSigningKey are the Upstash signing keys (injected from main).
func QStashVerifier(currentSigningKey, nextSigningKey string) func(next http.Handler) http.Handler {
	receiver := qstash.NewReceiver(currentSigningKey, nextSigningKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxQStashBodySize)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("qstash: failed to read request body", "error", err)
				apperrors.RespondWithError(w, apperrors.NewBadRequest("Bad Request", nil))
				return
			}

			if err := receiver.Verify(qstash.VerifyOptions{
				Signature: r.Header.Get("Upstash-Signature"),
				Body:      string(body),
				Url:       r.Header.Get("Upstash-Forwarded-Url"),
			}); err != nil {
				slog.Error("qstash: invalid signature", "error", err)
				apperrors.RespondWithError(w, apperrors.NewUnauthorized())
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)
		})
	}
}
