package config

import (
	"fmt"
	"os"
	"strings"
)

// Config contains runtime configuration required by the API application.
type Config struct {
	DatabaseURL             string
	ClerkWebhookSecret      string
	ClerkSecretKey          string
	QStashToken             string
	QStashCurrentSigningKey string
	QStashNextSigningKey    string
	AppBaseURL              string
	// JobBaseURL is the public origin QStash should POST callbacks to
	// (e.g. https://api-staging.askatlas.study). Defaults to AppBaseURL
	// when JOB_BASE_URL is unset, but the two differ in practice:
	// AppBaseURL is the Next.js frontend, while QStash callbacks must
	// land on the Go API where /jobs/* is registered. Folding them into
	// one var was a latent bug -- every dispatch was hitting the
	// frontend, which 404s.
	JobBaseURL   string
	AppEnv       string
	S3Bucket     string
	Port         string
	OpenAIAPIKey string
}

// Load reads and validates runtime configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:             strings.TrimSpace(os.Getenv("DATABASE_URL")),
		ClerkWebhookSecret:      strings.TrimSpace(os.Getenv("CLERK_WEBHOOK_SECRET")),
		ClerkSecretKey:          strings.TrimSpace(os.Getenv("CLERK_SECRET_KEY")),
		QStashToken:             strings.TrimSpace(os.Getenv("QSTASH_TOKEN")),
		QStashCurrentSigningKey: strings.TrimSpace(os.Getenv("QSTASH_CURRENT_SIGNING_KEY")),
		QStashNextSigningKey:    strings.TrimSpace(os.Getenv("QSTASH_NEXT_SIGNING_KEY")),
		AppBaseURL:              strings.TrimSpace(os.Getenv("APP_BASE_URL")),
		JobBaseURL:              strings.TrimSpace(os.Getenv("JOB_BASE_URL")),
		S3Bucket:                strings.TrimSpace(os.Getenv("S3_BUCKET")),
		Port:                    strings.TrimSpace(os.Getenv("PORT")),
		OpenAIAPIKey:            strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	var err error
	cfg.AppEnv, err = resolveAppEnv()
	if err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	if cfg.ClerkWebhookSecret == "" {
		return Config{}, fmt.Errorf("CLERK_WEBHOOK_SECRET environment variable is not set")
	}
	if cfg.ClerkSecretKey == "" {
		return Config{}, fmt.Errorf("CLERK_SECRET_KEY environment variable is not set")
	}
	if cfg.QStashToken == "" {
		return Config{}, fmt.Errorf("QSTASH_TOKEN environment variable is not set")
	}
	if cfg.QStashCurrentSigningKey == "" {
		return Config{}, fmt.Errorf("QSTASH_CURRENT_SIGNING_KEY environment variable is not set")
	}
	if cfg.QStashNextSigningKey == "" {
		return Config{}, fmt.Errorf("QSTASH_NEXT_SIGNING_KEY environment variable is not set")
	}
	if cfg.AppBaseURL == "" {
		return Config{}, fmt.Errorf("APP_BASE_URL environment variable is not set")
	}
	// JobBaseURL is intentionally optional -- when unset, QStash
	// callbacks fall back to AppBaseURL so envs that haven't been
	// updated keep their (broken-but-existing) behavior. Net-new envs
	// should set JOB_BASE_URL to the API origin explicitly.
	if cfg.JobBaseURL == "" {
		cfg.JobBaseURL = cfg.AppBaseURL
	}
	if cfg.S3Bucket == "" {
		return Config{}, fmt.Errorf("S3_BUCKET environment variable is not set")
	}
	if cfg.OpenAIAPIKey == "" {
		return Config{}, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return cfg, nil
}

func resolveAppEnv() (string, error) {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("APPENV")))
	}

	switch env {
	case "dev", "staging", "prod":
		return env, nil
	case "":
		return "", fmt.Errorf("APP_ENV (or APPENV) must be set to one of: dev, staging, prod")
	default:
		return "", fmt.Errorf("APP_ENV has invalid value %q; expected one of: dev, staging, prod", env)
	}
}
