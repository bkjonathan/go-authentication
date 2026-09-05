export PATH := $(PATH):$(shell go env GOPATH)/bin

COMPOSE := docker compose -f docker/docker-compose.yml --env-file .env

# The same file the application and compose read, with the same fallbacks
# internal/config declares. Optional: a fresh clone has no .env.
-include .env

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= password
DB_NAME ?= go_auth
DB_SSLMODE ?= disable

# Where the models are materialised for Atlas to read. Dropped and rebuilt on
# every diff, so nothing else may live in it.
SHADOW_SCHEMA := atlas_shadow

DB_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# The three states atlas.hcl compares - see the comments there.
DB_URL     := $(DB_DSN)&search_path=public
SHADOW_URL := $(DB_DSN)&search_path=$(SHADOW_SCHEMA)
DEV_URL    := docker://postgres/17-alpine/dev?search_path=public
export DB_URL SHADOW_URL DEV_URL

# Atlas has no "revert everything" flag, and asking to revert more versions
# than are applied is not an error - so ask for more than this project will have.
MIGRATE_ALL := 1000

.PHONY: help build run dev lint format \
	docker-up docker-down docker-logs docker-ps docker-reset \
	db-shadow db-diff db-inspect db-hash \
	migrate-up migrate-down migrate-reset migrate-status

help:
	@echo "Available Commands:"
	@echo "make build              - Build the application"
	@echo "make run                - Run the application"
	@echo "make dev                - Run the application in development mode"
	@echo "make lint               - Run linter on the codebase"
	@echo "make format             - Format the code and re-arrange imports"
	@echo "make docker-up          - Start the docker services in the background"
	@echo "make docker-down        - Stop the docker services"
	@echo "make docker-logs        - Follow the docker service logs"
	@echo "make docker-ps          - Show the docker service status"
	@echo "make docker-reset       - Stop the services and delete the database volume"
	@echo "make db-shadow          - Rebuild the shadow schema from the GORM models"
	@echo "make db-diff name=xxx   - Generate a migration from the GORM models"
	@echo "make db-inspect         - Print the DDL the models currently describe"
	@echo "make db-hash            - Re-hash atlas.sum after hand-editing a migration"
	@echo "make migrate-up         - Apply pending migrations"
	@echo "make migrate-down       - Roll back the last migration"
	@echo "make migrate-reset      - Roll back every migration"
	@echo "make migrate-status     - Show the currently applied version"

build:
	@go build -o bin/app ./cmd/api

run:
	@go run ./cmd/api

dev:
	@go run ./cmd/api

lint: format
	@golangci-lint run ./...

format:
	@gofmt -s -w .

docker-up:
	@$(COMPOSE) up -d

docker-down:
	@$(COMPOSE) down

docker-logs:
	@$(COMPOSE) logs -f

docker-ps:
	@$(COMPOSE) ps

docker-reset:
	@$(COMPOSE) down -v

db-shadow:
	@go run ./cmd/schema -schema $(SHADOW_SCHEMA)

db-diff: db-shadow
	@test -n "$(name)" || { echo "Usage: make db-diff name=add_password_reset" >&2; exit 1; }
	@atlas migrate diff $(name) --env local

db-inspect: db-shadow
	@atlas schema inspect --env local --url "$(SHADOW_URL)" --format '{{ sql . "  " }}'

db-hash:
	@atlas migrate hash --env local

migrate-up:
	@atlas migrate apply --env local

migrate-down:
	@atlas migrate down --env local

migrate-reset:
	@atlas migrate down --env local $(MIGRATE_ALL)

migrate-status:
	@atlas migrate status --env local
