# InstantGate API Makefile

.PHONY: help build run test clean docker-build docker-up docker-down fmt lint

# Variables
APP_NAME=instantgate
CMD_DIR=./cmd/instantgate
BUILD_DIR=./bin
CONFIG_FILE=./config/config.yaml

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development
run: ## Run the application locally
	$(GORUN) $(CMD_DIR)/main.go -config $(CONFIG_FILE)

dev: ## Run with hot reload (requires air)
	air

## Build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go

build-linux: ## Build for Linux
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)/main.go

build-darwin: ## Build for macOS
	@echo "Building $(APP_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)/main.go

build-windows: ## Build for Windows
	@echo "Building $(APP_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)/main.go

build-all: build-linux build-darwin build-windows ## Build for all platforms

## Testing
test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	$(GOTEST) -v -race ./...

## Dependencies
deps: ## Download dependencies
	$(GOMOD) download

deps-update: ## Update dependencies
	$(GOMOD) tidy

## Code Quality
fmt: ## Format code
	$(GOFMT) -s -w .

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

vet: ## Run go vet
	$(GOCMD) vet ./...

## Docker
docker-build: ## Build Docker image
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

docker-build-multi: ## Build multi-arch Docker image
	docker buildx build --platform linux/amd64,linux/arm64 -f deployments/Dockerfile -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

docker-up: ## Start services with docker-compose
	docker-compose -f deployments/docker-compose.yml up -d

docker-down: ## Stop services with docker-compose
	docker-compose -f deployments/docker-compose.yml down

docker-logs: ## Show docker-compose logs
	docker-compose -f deployments/docker-compose.yml logs -f

## Database
db-migrate: ## Run database migrations
	mysql -h localhost -u root -p instantgate < scripts/sample-schema.sql

db-shell: ## Open MySQL shell
	mysql -h localhost -u root -p instantgate

## Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

clean-docker: ## Clean Docker resources
	docker-compose -f deployments/docker-compose.yml down -v
	docker system prune -f

## Install tools
install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Default
.DEFAULT_GOAL := help
