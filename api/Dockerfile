#syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.19 AS production

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

RUN apk add --no-cache bash sudo curl

RUN curl -1sLf \
    'https://dl.cloudsmith.io/public/infisical/infisical-cli/setup.alpine.sh' \
    | bash

RUN apk update && sudo apk add infisical=0.43.32

WORKDIR /

COPY --from=builder /api /api
COPY start.sh /start.sh
RUN chmod +x start.sh

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/start.sh"]
