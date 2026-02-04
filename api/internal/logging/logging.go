package logging

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/lmittmann/tint"
)

type loggerConfig struct {
	env       string
	level     slog.Level
	addSource *bool
	writer    io.Writer
	handler   slog.Handler
}

type Option func(*loggerConfig)

func WithEnv(env string) Option {
	return func(config *loggerConfig) {
		config.env = env
	}
}

func WithLevel(level slog.Level) Option {
	return func(config *loggerConfig) {
		config.level = level
	}
}

func WithAddSource(addSource bool) Option {
	return func(config *loggerConfig) {
		config.addSource = &addSource
	}
}

func WithWriter(writer io.Writer) Option {
	return func(config *loggerConfig) {
		config.writer = writer
	}
}

func WithHandler(handler slog.Handler) Option {
	return func(config *loggerConfig) {
		config.handler = handler
	}
}

func NewLogger(opts ...Option) *slog.Logger {
	cfg := loggerConfig{
		env:    os.Getenv("APP_ENV"),
		level:  slog.LevelInfo,
		writer: os.Stdout,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	env := cfg.env
	if env == "" {
		env = "dev"
	}
	if cfg.handler != nil {
		return slog.New(cfg.handler)
	}

	addSource := env == "dev"
	if cfg.addSource != nil {
		addSource = *cfg.addSource
	}

	if env == "dev" {
		h := tint.NewHandler(cfg.writer, &tint.Options{
			Level:      cfg.level,
			TimeFormat: time.Kitchen,
			AddSource:  addSource,
		})
		return slog.New(h)
	}

	h := slog.NewJSONHandler(cfg.writer, &slog.HandlerOptions{
		Level:     cfg.level,
		AddSource: addSource,
	})
	return slog.New(h)

}

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"remote", r.RemoteAddr,
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
