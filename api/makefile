BINARY_NAME := api
MAIN_PATH   := ./cmd/api/main.go
SEED_PATH   := ./scripts/seed_files.go
SEED_SCHOOLS_PATH := ./scripts/seed_schools
SEED_COURSES_PATH := ./scripts/seed_courses
SEED_STUDY_GUIDES_PATH := ./scripts/seed_study_guides

ENV     ?= dev
ACTION  ?= seed
COUNT   ?= 50
MAX_ROWS ?= 1000
CSV     ?= scripts/data/schools.csv
COURSES_CSV ?= scripts/data/courses.csv
SECTIONS_CSV ?= scripts/data/course_sections.csv
STUDY_GUIDES_CSV ?= scripts/data/study_guides.csv

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
	sqlc/generate docker-build docker-run mockery seed seed-schools seed-courses \
	seed-study-guides seed-demo-content e2e generate-docs bruno

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

## make bruno
##
## Regenerates the Bruno collection under bruno/ from openapi.yaml. Wipes
## every subfolder except bruno/environments/ (which is hand-maintained)
## and re-emits one .bru per operation, grouped by the first non-variable
## path segment. Run after any openapi.yaml change so the Bruno sidebar
## stays in sync. Safe to run repeatedly -- output is deterministic.
bruno:
	go run ./scripts/brunogen

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

## make seed-courses [ENV=dev|staging|prod] [COURSES_CSV=scripts/data/courses.csv] [SECTIONS_CSV=scripts/data/course_sections.csv]
##
## Two-phase loader for courses + course_sections. Resolves parent schools
## by ipeds_id and parent courses by (school_ipeds_id, department, number),
## so seed-schools must run first. Both phases run inside one transaction
## with ON CONFLICT DO NOTHING against the natural unique keys
## (uq_courses_school_dept_num, uq_sections_course_term_code), so re-runs
## are safe. ENV=prod requires interactive [y] confirmation.
seed-courses:
	$(call prod-confirm)
	infisical run --env=$(ENV) -- go run $(SEED_COURSES_PATH) -courses=$(COURSES_CSV) -sections=$(SECTIONS_CSV)

## make seed-study-guides [ENV=dev|staging|prod] [STUDY_GUIDES_CSV=scripts/data/study_guides.csv]
##
## Seeds study_guides + the supporting aggregate rows (2 upvotes per
## guide, 1 recommendation on every other guide, 1 quiz per guide) for
## ASK-104 staging-matrix verification. Upserts 3 stable "seed bot"
## users with clerk_ids seed_bot_v1_1..3 so the seed is self-contained
## and does not depend on the Clerk webhook pipeline. All inserts use
## ON CONFLICT DO NOTHING so re-runs are safe. Parent courses must be
## seeded first (see seed-courses). ENV=prod requires interactive [y]
## confirmation.
seed-study-guides:
	$(call prod-confirm)
	infisical run --env=$(ENV) -- go run $(SEED_STUDY_GUIDES_PATH) -guides=$(STUDY_GUIDES_CSV)

## make seed-demo-content [ENV=dev|staging|prod] [ACTION=seed|clean] [SYNTH_USERS=1000] [SEED=42]
##
## Phase 3 demo-data loader. ACTION=seed (default) reads the Phase 1+2
## validated YAML/MD fixtures at scripts/seed_demo/fixtures/ and writes
## them into the target database in three layers:
##   1. Identity: bot + demo + N synthetic users (direct INSERT with
##      fake clerk_ids seed_bot, seed_demo, seed_synth_NNNN).
##   2. Fixtures: 100 files, 5 resources, 82 guides, 43 quizzes
##      (with their 258 questions + 716 options), plus join tables.
##      Guide body placeholders ({{FILE:slug}} etc.) are rewritten
##      in-place once all UUIDs are known.
##   3. Activity: course_sections + memberships (2k+), Zipf view
##      counts, long-tail votes, recommendations, favorites, last-
##      viewed entries. Demo user gets curated heavy population.
## All three layers + pass-2 rewrite run in a single transaction;
## INSERTs are ON CONFLICT DO NOTHING / SELECT-fallback so re-runs
## are idempotent. Runs in ~3 minutes against Neon dev.
##
## ACTION=clean reverses everything: removes every Garage object
## under the seed-demo/ prefix, then removes every DB row owned by
## a seed_* clerk_id user (FK cascades handle join tables, per-user
## votes/favorites/sessions/memberships, and all content). Real
## users (clerk_id prefixed user_*) and the Phase 0 schools/courses
## catalog are never touched.
##
## ENV=prod requires interactive [y] confirmation on both actions.
seed-demo-content:
	$(call prod-confirm)
	infisical run --env=$(ENV) -- go run ./scripts/seed_demo_content/ \
		--action=$(or $(ACTION),seed) \
		--synth-users=$(or $(SYNTH_USERS),1000) \
		--seed=$(or $(SEED),42)

## make e2e E2E_TOKEN=<token> [ENV=dev|staging]
e2e:
	@test -n "$(E2E_TOKEN)" || (echo "Error: E2E_TOKEN is required"; exit 1)
	@echo "Running E2E tests against $(ENV)"
	cd e2e && E2E_BASE_URL=$(call e2e_url,$(ENV)) E2E_TOKEN=$(E2E_TOKEN) npx playwright test

e2e_url = $(if $(filter staging,$(1)),https://api-staging.askatlas.study,https://api-dev.askatlas.study)