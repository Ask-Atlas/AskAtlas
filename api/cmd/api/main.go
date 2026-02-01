package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := logging.NewLogger()
	slog.SetDefault(logger)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logging.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	// TODO: Add timeout to a config file instead of hardcoding
	r.Use(middleware.Timeout(60 * time.Second))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	addr := ":8080"
	slog.Info("Server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
