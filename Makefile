.PHONY: dev db db-stop migrate migrate-down build-api build-worker tidy

# Start local PostgreSQL
db:
	docker compose up -d postgres

# Stop local PostgreSQL
db-stop:
	docker compose down

# Run database migrations (requires migrate CLI: brew install golang-migrate)
migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

# Run API server (with live reload — requires: go install github.com/air-verse/air@latest)
dev-api:
	air -c .air.api.toml

# Run worker
dev-worker:
	go run ./cmd/worker

# Build binaries
build-api:
	go build -o bin/api ./cmd/api

build-worker:
	go build -o bin/worker ./cmd/worker

build: build-api build-worker

# Tidy dependencies
tidy:
	go mod tidy

# Run tests
test:
	go test ./...
