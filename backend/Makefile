.PHONY: build run dev lint migrate-create migrate-up migrate-down format docker-up docker-down

MIGRATIONS_DIR = db/migrations
DB_URL = "postgresql://postgres:password@localhost:5432/nyarikos?sslmode=disable"

build:
	go build -o bin/app ./cmd/api

run:
	go run ./cmd/api

dev:
	go run ./cmd/api

lint: format
	golangci-lint run ./...

format:
	@gofumpt -l -w .
	@goimports -w .

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down