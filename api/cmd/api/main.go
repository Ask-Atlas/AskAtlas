// Package main is the entry point for the AskAtlas API.
// It initializes the database connection, configures middleware,
// sets up routes, and starts the HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
	"github.com/Ask-Atlas/AskAtlas/api/internal/config"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	"github.com/Ask-Atlas/AskAtlas/api/internal/logging"
	"github.com/Ask-Atlas/AskAtlas/api/internal/middleware"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/Ask-Atlas/AskAtlas/api/internal/user"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	middleware_oapi "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(logging.WithEnv(cfg.AppEnv))
	slog.SetDefault(logger)
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(logging.RequestLogger(logger))
	r.Use(chiMiddleware.Recoverer)

	ctx := context.Background()
	connPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create database connection pool", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	queries := db.New(connPool)
	userRepository := user.NewSQLCRepository(queries)
	userService := user.NewService(userRepository)
	clerkService := clerk.NewClerkService(userService)
	clerkWebhookHandler := handlers.NewClerkWebhookHandler(clerkService)

	clerkSignatureVerifier := middleware.SVIXVerifier(cfg.ClerkWebhookSecret)
	clerkSDK.SetKey(cfg.ClerkSecretKey)

	s3Client, err := s3client.New(ctx, cfg.S3Bucket)
	if err != nil {
		slog.Error("failed to create S3 client", "error", err)
		os.Exit(1)
	}
	qstashClient := qstashclient.New(cfg.QStashToken, cfg.AppBaseURL+"/jobs", cfg.AppEnv)
	qstashVerifier := middleware.QStashVerifier(cfg.QStashCurrentSigningKey, cfg.QStashNextSigningKey)
	jobHandler := handlers.NewJobHandler(s3Client, queries)

	fileRepo := files.NewSQLCRepository(connPool, queries)
	fileService := files.NewService(fileRepo)
	fileHandler := handlers.NewFileHandler(fileService, qstashClient)

	clerkAuth := middleware.ClerkAuth(userService)

	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	r.Route("/webhooks", func(r chi.Router) {
		r.With(clerkSignatureVerifier).Post("/clerk", clerkWebhookHandler.Webhook)
	})

	r.Route("/jobs", func(r chi.Router) {
		r.With(qstashVerifier).Post("/delete-file", jobHandler.DeleteFileJob)
		r.With(qstashVerifier).Post("/delete-file-failed", jobHandler.DeleteFileFailedJob)
	})

	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load swagger spec", "error", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	oapiOptions := middleware_oapi.Options{
		ErrorHandler: api.OAPIValidatorErrorHandler,
	}

	r.Route("/api", func(r chi.Router) {
		r.Use(clerkAuth)
		r.Use(middleware_oapi.OapiRequestValidatorWithOptions(swagger, &oapiOptions))
		api.HandlerWithOptions(fileHandler, api.ChiServerOptions{
			BaseRouter:       r,
			ErrorHandlerFunc: api.OAPIStrictErrorHandler,
		})
	})

	addr := ":" + cfg.Port
	slog.Info("Server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
