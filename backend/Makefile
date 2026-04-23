.PHONY: build run dev lint migrate migrate-up migrate-down format docker-up docker-down test swag

MIGRATIONS_DIR = db/migrations
DB_URL = "postgresql://postgres:password@localhost:5432/nyarikos?sslmode=disable"

build: migrate-up
	go build -o bin/app ./cmd/api

run: migrate-up
	go run ./cmd/api

dev:
	air

lint: format
	golangci-lint run ./...

format:
	@gofumpt -l -w .
	@goimports -w .

migrate:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

test:
# 	go test -coverprofile=coverage.out ./internal/...
# 	go tool cover -html=coverage.out
	go test -cover ./...

swag:
	swag fmt -g cmd/api/main.go
	swag init -g main.go --dir ./cmd/api,./internal/server,./internal/dto,./internal/models,./internal/utils --output ./docs