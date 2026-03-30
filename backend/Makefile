.PHONY: build run dev lint migrate-up migrate-down

help:
	@echo ""

build:
	go build -o bin/app ./cmd/api

run:
	go run ./cmd/api

dev:
	go run ./cmd/api

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path db/migrations -database "postgresql://postgresql:password@localhost:5432/nyarikos?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgresql://postgresql:password@localhost:5432/nyarikos?sslmode=disable" down

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down 