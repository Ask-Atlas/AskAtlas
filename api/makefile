BINARY_NAME=api
MAIN_PATH=./cmd/api/main.go

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

tidy:
	go mod tidy
	go mod verify

tidy-check:
	./scripts/check-tidy.sh

sqlc/generate:
	go tool sqlc generate

docker-build:
	docker build -t askatlas-api .

docker-run:
	docker run -p 8080:8080 --env-file .env.local askatlas-api

.PHONY: dev stage prod docker-build docker-run format format-check lint test