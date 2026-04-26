// Package main is the entry point for the AskAtlas API.
// It initializes the database connection, configures middleware,
// sets up routes, and starts the HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/aiedits"
	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
	"github.com/Ask-Atlas/AskAtlas/api/internal/config"
	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	"github.com/Ask-Atlas/AskAtlas/api/internal/dashboard"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/favorites"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	"github.com/Ask-Atlas/AskAtlas/api/internal/logging"
	"github.com/Ask-Atlas/AskAtlas/api/internal/middleware"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	"github.com/Ask-Atlas/AskAtlas/api/internal/quizzes"
	"github.com/Ask-Atlas/AskAtlas/api/internal/recents"
	"github.com/Ask-Atlas/AskAtlas/api/internal/refs"
	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/Ask-Atlas/AskAtlas/api/internal/schools"
	"github.com/Ask-Atlas/AskAtlas/api/internal/sessions"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/internal/user"
	"github.com/getkin/kin-openapi/openapi3filter"
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
	jobBaseURL, err := url.JoinPath(cfg.AppBaseURL, "jobs")
	if err != nil {
		slog.Error("failed to construct job base URL", "error", err)
		os.Exit(1)
	}
	qstashClient := qstashclient.New(cfg.QStashToken, jobBaseURL, cfg.AppEnv)
	qstashVerifier := middleware.QStashVerifier(cfg.QStashCurrentSigningKey, cfg.QStashNextSigningKey)

	extractWorker := files.NewExtractWorker(files.NewExtractRepository(queries), s3Client)
	jobHandler := handlers.NewJobHandler(s3Client, queries, extractWorker)

	fileRepo := files.NewSQLCRepository(connPool, queries)
	fileService := files.NewService(fileRepo, files.WithDownloadURLGenerator(s3Client))
	fileHandler := handlers.NewFileHandler(fileService, qstashClient)
	grantHandler := handlers.NewGrantHandler(fileService)

	schoolsRepo := schools.NewSQLCRepository(queries)
	schoolsService := schools.NewService(schoolsRepo)
	schoolsHandler := handlers.NewSchoolsHandler(schoolsService)

	coursesRepo := courses.NewSQLCRepository(queries)
	coursesService := courses.NewService(coursesRepo)
	coursesHandler := handlers.NewCoursesHandler(coursesService)

	studyGuidesRepo := studyguides.NewSQLCRepository(connPool, queries)
	studyGuidesService := studyguides.NewService(studyGuidesRepo)
	studyGuidesHandler := handlers.NewStudyGuideHandler(studyGuidesService)
	studyGuideGrantHandler := handlers.NewStudyGuideGrantHandler(studyGuidesService)

	quizzesRepo := quizzes.NewSQLCRepository(connPool, queries)
	quizzesService := quizzes.NewService(quizzesRepo)
	quizzesHandler := handlers.NewQuizzesHandler(quizzesService)

	sessionsRepo := sessions.NewSQLCRepository(connPool, queries)
	sessionsService := sessions.NewService(sessionsRepo)
	sessionsHandler := handlers.NewSessionsHandler(sessionsService)

	recentsRepo := recents.NewSQLCRepository(queries)
	recentsService := recents.NewService(recentsRepo)
	recentsHandler := handlers.NewRecentsHandler(recentsService)

	favoritesRepo := favorites.NewSQLCRepository(queries)
	favoritesService := favorites.NewService(favoritesRepo)
	favoritesHandler := handlers.NewFavoritesHandler(favoritesService)

	dashboardRepo := dashboard.NewSQLCRepository(queries)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	refsRepo := refs.NewSQLCRepository(queries)
	refsService := refs.NewService(refsRepo)
	refsHandler := handlers.NewRefsHandler(refsService)

	// QuotaService is both the per-user daily-cap gate (called by
	// the AIQuota middleware before each request) and the
	// UsageRecorder that ai.Client uses to write ai_usage rows from
	// its cost-log hook -- so partial usage on cancellation still
	// counts against the user's quota.
	aiQuotaService := ai.NewQuotaService(queries, ai.QuotasFromEnv())
	aiClient := ai.NewClient(cfg.OpenAIAPIKey, logger, ai.WithRecorder(aiQuotaService))
	aiHandler := handlers.NewAIHandler(aiClient)

	aiEditsRepo := aiedits.NewSQLCRepository(queries)
	aiEditsService := aiedits.NewService(aiEditsRepo)
	aiEditHandler := handlers.NewAIEditHandler(studyGuidesService, aiEditsService, aiClient)

	clerkAuth := middleware.ClerkAuth(userService)

	// Default 60s timeout for non-streaming endpoints. AI routes
	// under /api/ai/* opt out below because chiMiddleware.Timeout
	// uses http.TimeoutHandler, which buffers the response and
	// would defeat SSE flush-per-chunk semantics.
	defaultTimeout := chiMiddleware.Timeout(60 * time.Second)

	r.With(defaultTimeout).Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	r.Route("/webhooks", func(r chi.Router) {
		r.Use(defaultTimeout)
		r.With(clerkSignatureVerifier).Post("/clerk", clerkWebhookHandler.Webhook)
	})

	r.Route("/jobs", func(r chi.Router) {
		r.Use(defaultTimeout)
		r.With(qstashVerifier).Post("/delete-file", jobHandler.DeleteFileJob)
		r.With(qstashVerifier).Post("/delete-file-failed", jobHandler.DeleteFileFailedJob)
		r.With(qstashVerifier).Post("/extract-file", jobHandler.ExtractFileJob)
		r.With(qstashVerifier).Post("/extract-file-failed", jobHandler.ExtractFileFailedJob)
	})

	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load swagger spec", "error", err)
		os.Exit(1)
	}

	oapiOptions := middleware_oapi.Options{
		ErrorHandler: api.OAPIValidatorErrorHandler,
		Options: openapi3filter.Options{
			AuthenticationFunc: api.BearerAuthFunc,
		},
	}

	compositeHandler := handlers.NewCompositeHandler(
		fileHandler,
		grantHandler,
		schoolsHandler,
		coursesHandler,
		studyGuidesHandler,
		studyGuideGrantHandler,
		quizzesHandler,
		sessionsHandler,
		recentsHandler,
		favoritesHandler,
		dashboardHandler,
		refsHandler,
		aiHandler,
		aiEditHandler,
	)

	r.Route("/api", func(r chi.Router) {
		r.Use(clerkAuth)
		r.Use(middleware_oapi.OapiRequestValidatorWithOptions(swagger, &oapiOptions))
		// Both middlewares consult the same FeatureForPath mapper so
		// they can never disagree about whether a request is "AI".
		// chiMiddleware.Timeout wraps the response writer in an
		// http.TimeoutHandler that buffers the body -- which breaks
		// SSE flush-per-chunk -- so AI routes opt out; the AI
		// client's request-context cancel is the real shutdown path.
		r.Use(skipTimeoutForAIRoutes(defaultTimeout, nil))
		r.Use(middleware.AIQuota(aiQuotaService, nil))
		api.HandlerWithOptions(compositeHandler, api.ChiServerOptions{
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

// skipTimeoutForAIRoutes returns a middleware that applies `timeout`
// to every request EXCEPT those that the FeatureForPath mapper
// classifies as AI. Used so SSE streaming routes bypass the default
// http.TimeoutHandler -- which buffers and breaks flush-per-chunk --
// while every other CRUD route keeps the 60s safety net.
//
// Sharing the mapper with middleware.AIQuota guarantees the timeout
// skip and the quota gate agree on which routes are "AI".
func skipTimeoutForAIRoutes(timeout func(http.Handler) http.Handler, mapper middleware.FeatureForPath) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrapped := timeout(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if middleware.IsAIRoute(mapper, r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			wrapped.ServeHTTP(w, r)
		})
	}
}
