export PATH := $(PATH):$(shell go env GOPATH)/bin

COMPOSE := docker compose -f docker/docker-compose.yml --env-file .env

.PHONY: help build run dev lint format docker-up docker-down docker-logs docker-ps docker-reset

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
