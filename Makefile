.PHONY: all build clean dev test lint help service-start service-stop service-status

# Variables
BINARY_NAME=mission-control
BUILD_DIR=bin
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
SERVICE_PID_FILE=.mission-control.pid
GO_FILES=$(shell find . -name '*.go' -not -path './ui/*')

# Default target
all: build

## help: Show this help message
help:
	@echo "Claw Agent Mission Control"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## dev: Start development mode with hot reload
dev:
	@echo "Starting development servers..."
	@trap 'kill 0' EXIT; \
	(cd ui && npm run dev) & \
	(air) & \
	wait

## dev-ui: Start frontend dev server only
dev-ui:
	cd ui && npm run dev

## dev-api: Start backend with Air (hot reload)
dev-api:
	air

## build: Build production binary with embedded frontend
build: build-ui build-server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-ui: Build frontend for production
build-ui:
	@echo "Building frontend..."
	cd ui && npm ci && npm run build

## build-server: Build Go server (requires built frontend)
build-server:
	@echo "Copying UI assets for embedding..."
	@rm -rf internal/ui/assets
	@cp -r ui/out internal/ui/assets
	@echo "Building server..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Cleaning up copied assets..."
	@rm -rf internal/ui/assets

## run: Run the production binary
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

## service-start: Run binary as background service (uses bin/mission-control)
service-start:
	@if [ ! -f $(BINARY_PATH) ]; then \
		echo "Binary not found at $(BINARY_PATH). Run 'make build' first."; \
		exit 1; \
	fi
	@if [ -f $(SERVICE_PID_FILE) ]; then \
		pid=$$(cat $(SERVICE_PID_FILE)); \
		if kill -0 $$pid 2>/dev/null; then \
			echo "Service already running (PID $$pid). Use 'make service-stop' first."; \
			exit 1; \
		fi; \
		rm -f $(SERVICE_PID_FILE); \
	fi
	@echo "Starting mission-control service in background..."
	@env HOST=0.0.0.0 nohup ./$(BINARY_PATH) > mission-control.log 2>&1 & echo $$! > $(SERVICE_PID_FILE)
	@echo "Service started (PID $$(cat $(SERVICE_PID_FILE))). Logs: mission-control.log"

## service-stop: Stop the background service
service-stop:
	@if [ ! -f $(SERVICE_PID_FILE) ]; then \
		echo "No PID file found. Service may not be running."; \
		exit 0; \
	fi
	@pid=$$(cat $(SERVICE_PID_FILE)); \
	if kill -0 $$pid 2>/dev/null; then \
		kill $$pid && echo "Service stopped (PID $$pid)."; \
	else \
		echo "Process $$pid not running."; \
	fi; \
	rm -f $(SERVICE_PID_FILE)

## service-status: Show whether the service is running
service-status:
	@if [ ! -f $(SERVICE_PID_FILE) ]; then \
		echo "Service not running (no PID file)."; \
		exit 0; \
	fi
	@pid=$$(cat $(SERVICE_PID_FILE)); \
	if kill -0 $$pid 2>/dev/null; then \
		echo "Service running (PID $$pid)."; \
	else \
		echo "Service not running (stale PID file)."; \
		rm -f $(SERVICE_PID_FILE); \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf ui/.next
	rm -rf ui/out
	rm -rf ui/node_modules
	rm -rf internal/ui/assets
	rm -f mission-control server claw-agent-mission-control

## test: Run all tests
test: test-go test-ui

## test-go: Run Go tests
test-go:
	go test -v ./...

## test-ui: Run frontend tests
test-ui:
	cd ui && npm test

## lint: Run linters
lint: lint-go lint-ui

## lint-go: Run Go linter
lint-go:
	golangci-lint run

## lint-ui: Run frontend linter
lint-ui:
	cd ui && npm run lint

## fmt: Format code
fmt:
	go fmt ./...
	cd ui && npm run format

## deps: Install dependencies
deps:
	go mod download
	cd ui && npm ci

## deps-update: Update dependencies
deps-update:
	go get -u ./...
	go mod tidy
	cd ui && npm update

## migrate: Run database migrations
migrate:
	go run ./cmd/migrate

## migrate-new: Create new migration (usage: make migrate-new name=create_users)
migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=migration_name"; exit 1; fi
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

## docker: Build Docker image
docker:
	docker build -t claw-agent-mission-control .

## docker-run: Run Docker container
docker-run:
	docker run -p 8080:8080 \
		-e OPENCLAW_GATEWAY_URL=ws://host.docker.internal:18789 \
		-e OPENCLAW_GATEWAY_TOKEN=$${OPENCLAW_GATEWAY_TOKEN} \
		-v $$(pwd)/data:/app/data \
		claw-agent-mission-control

## install-tools: Install development tools
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## sqlc: Generate sqlc code
sqlc:
	sqlc generate

service-rebuild: build service-stop service-start