dev:
	go run cmd/api/main.go

docker-build:
	docker build -t askatlas-api .

docker-run:
	docker run -p 8080:8080 --env-file .env.local askatlas-api

.PHONY: dev docker-build docker-run