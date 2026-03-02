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
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	"github.com/Ask-Atlas/AskAtlas/api/internal/logging"
	"github.com/Ask-Atlas/AskAtlas/api/internal/middleware"
	"github.com/Ask-Atlas/AskAtlas/api/internal/user"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	middleware_oapi "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	logger := logging.NewLogger()
	slog.SetDefault(logger)
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(logging.RequestLogger(logger))
	r.Use(chiMiddleware.Recoverer)

	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is not set")
		os.Exit(1)
	}

	connPool, err := pgxpool.New(ctx, databaseURL)
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

	clerkWebhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if clerkWebhookSecret == "" {
		slog.Error("CLERK_WEBHOOK_SECRET environment variable is not set")
		os.Exit(1)
	}
	clerkSignatureVerifier := middleware.SVIXVerifier(clerkWebhookSecret)

	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		slog.Error("CLERK_SECRET_KEY environment variable is not set")
		os.Exit(1)
	}
	clerkSDK.SetKey(clerkSecretKey)

	fileRepo := files.NewSQLCRepository(queries)
	fileService := files.NewService(fileRepo)
	fileHandler := handlers.NewFileHandler(fileService)

	clerkAuth := middleware.ClerkAuth(userService)

	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	r.Route("/webhooks", func(r chi.Router) {
		r.With(clerkSignatureVerifier).Post("/clerk", clerkWebhookHandler.Webhook)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	slog.Info("Server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
