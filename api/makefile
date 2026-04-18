BINARY_NAME := api
MAIN_PATH   := ./cmd/api/main.go
SEED_PATH   := ./scripts/seed_files.go
SEED_SCHOOLS_PATH := ./scripts/seed_schools

ENV     ?= dev
ACTION  ?= seed
COUNT   ?= 50
MAX_ROWS ?= 1000
CSV     ?= scripts/data/schools.csv

GOMARKDOC_VERSION        ?= v1.1.0
GOMARKDOC_CMD            := go run github.com/princjef/gomarkdoc/cmd/gomarkdoc@$(GOMARKDOC_VERSION)

# prod-confirm prompts for interactive [y] confirmation when ENV=prod, mirroring
# the migrations/makefile pattern. Targets that mutate data should `$(call prod-confirm)`
# before doing anything destructive or wide-reaching.
define prod-confirm
	@if [ "$(ENV)" = "prod" ]; then \
		echo "Are you sure you want to run this command in production?"; \
		read -p "Type 'y' to continue: " confirm; \
		if [ "$$confirm" != "y" ]; then \
			echo "Aborting..."; \
			exit 1; \
		fi; \
	fi
endef

.PHONY: dev stage prod build install format format-check lint test tidy tidy-check \
	sqlc/generate docker-build docker-run mockery seed seed-schools e2e generate-docs

dev:
	infisical run --env=dev -- go run $(MAIN_PATH)

stage:
	infisical run --env=staging -- go run $(MAIN_PATH)

prod:
	infisical run --env=prod -- go run $(MAIN_PATH)

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

install:
	go mod download

format:
	go tool goimports -w .

format-check:
	go tool goimports -l .

lint:
	go tool golangci-lint run

test:
	go test ./...

mockery:
	go tool mockery

generate-api:
	mkdir -p internal/api
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 -config oapi-codegen.yaml openapi.yaml

generate-docs:
	rm -rf ../docs/docs/backend
	mkdir -p ../docs/docs/backend
	go list ./cmd/... ./internal/... ./pkg/... \
		| grep -v -E '/mock|/mocks' \
		| sed 's|github.com/Ask-Atlas/AskAtlas/api/|./|' \
		| xargs $(GOMARKDOC_CMD) -o ../docs/docs/backend/{{.Dir}}.md
tidy:
	go mod tidy && go mod verify

tidy-check:
	./scripts/check-tidy.sh

sqlc/generate:
	go tool sqlc generate

docker-build:
	docker build -t askatlas-api .

docker-run:
	docker run -p 8080:8080 --env-file .env.local askatlas-api

## make seed USER_ID=<uuid> [ENV=dev|staging|prod] [ACTION=seed|clean] [COUNT=50] [MAX_ROWS=1000] [REPORT=true]
seed:
	@test -n "$(USER_ID)" || (echo "Error: USER_ID is required"; exit 1)
	infisical run --env=$(ENV) -- go run $(SEED_PATH) \
		-user-id=$(USER_ID) \
		-action=$(ACTION) \
		-count=$(COUNT) \
		-max-rows=$(MAX_ROWS) \
		$(if $(filter true,$(REPORT)),-report)

## make seed-schools [ENV=dev|staging|prod] [CSV=scripts/data/schools.csv]
##
## Loads the schools seed CSV via INSERT ... ON CONFLICT DO NOTHING. Idempotent
## on the schools table's partial unique indexes (ipeds_id, domain), so re-runs
## are safe. Use a custom CSV via CSV=path/to/your.csv (path is relative to
## the api/ directory). ENV=prod requires interactive [y] confirmation.
seed-schools:
	$(call prod-confirm)
	infisical run --env=$(ENV) -- go run $(SEED_SCHOOLS_PATH) -csv=$(CSV)

## make e2e E2E_TOKEN=<token> [ENV=dev|staging]
e2e:
	@test -n "$(E2E_TOKEN)" || (echo "Error: E2E_TOKEN is required"; exit 1)
	@echo "Running E2E tests against $(ENV)"
	cd e2e && E2E_BASE_URL=$(call e2e_url,$(ENV)) E2E_TOKEN=$(E2E_TOKEN) npx playwright test

e2e_url = $(if $(filter staging,$(1)),https://api-staging.askatlas.study,https://api-dev.askatlas.study)