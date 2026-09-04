export PATH := $(PATH):$(shell go env GOPATH)/bin


help:
	@echo "Available Commands:"
	@echo "make build              - Build the application"
	@echo "make run                - Run the application"
	@echo "make dev                - Run the application in development mode"
	@echo "make lint               - Run linter on the codebase"
	@echo "make format             - Format the code and re-arrange imports"

build:
	@go build -o bin/app ./cmd/api

run:
	@go run ./cmd/api

dev:
	@go run ./cmd/api

lint:format
	@golangci-lint run ./...

format:
	@gofmt -s -w .