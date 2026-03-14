.PHONY: help build test run docker-build docker-up docker-down docker-logs clean lint fmt vet

# Variables
APP_NAME=url-shortener
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_REGISTRY?=docker.io
GO=go
PLATFORMS?=linux/amd64,linux/arm64

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the Go application locally
	$(GO) build -o bin/$(APP_NAME) .

clean: ## Remove build artifacts
	rm -rf bin/
	$(GO) clean

# Test targets
test: ## Run unit tests
	$(GO) test -v -cover ./...

test-short: ## Run unit tests (short mode)
	$(GO) test -short -v ./...

# Lint & Format targets
fmt: ## Format code with gofmt
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Run all linters (fmt + vet)

# Dependency management
deps: ## Download and tidy dependencies
	$(GO) mod download
	$(GO) mod tidy

# Docker targets
docker-build: ## Build Docker image using docker buildx
	@# Use buildx (multi-platform) when available; otherwise fall back to docker build (single-platform)
	@if docker buildx version >/dev/null 2>&1; then \
		docker buildx build \
			--platform $(PLATFORMS) \
			-t $(DOCKER_IMAGE) \
			-f Dockerfile \
			--load \
			.; \
	else \
		echo "docker buildx not available, falling back to 'docker build' (single-arch)"; \
		docker build -t $(DOCKER_IMAGE) -f Dockerfile .; \
	fi

docker-build-push: ## Build and push Docker image to registry (requires credentials)
	@# Use buildx (multi-platform push) when available; otherwise fall back to docker build + docker push
	@if docker buildx version >/dev/null 2>&1; then \
		docker buildx build \
			--platform $(PLATFORMS) \
			-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE) \
			-f Dockerfile \
			--push \
			.; \
	else \
		echo "docker buildx not available, falling back to 'docker build' + 'docker push' (single-arch)"; \
		docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE) -f Dockerfile . && docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE); \
	fi

docker-up: ## Start Docker Compose stack
	docker-compose up -d

docker-down: ## Stop and remove Docker Compose stack
	docker-compose down

docker-logs: ## View Docker Compose logs (app & redis)
	docker-compose logs -f

docker-logs-app: ## View application logs
	docker-compose logs -f app

docker-logs-redis: ## View Redis logs
	docker-compose logs -f redis

docker-restart: ## Restart Docker Compose stack
	docker-compose restart

docker-clean: ## Remove Docker containers, volumes, and images
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# Development targets
dev-up: ## Start development environment with docker-compose
	docker-compose up

dev-down: ## Stop development environment
	docker-compose down

dev-logs: ## View development logs
	docker-compose logs -f

# Local run (requires Redis on localhost:6379)
run: build ## Build and run the application locally
	./bin/$(APP_NAME)

# All-in-one targets
all: clean lint test build ## Run clean, lint, test, and build

docker-all: docker-clean docker-build ## Clean and rebuild Docker image

.env:
	cp .env.example .env
	@echo ".env file created from .env.example"

# Information targets
info: ## Display project information
	@echo "Project: $(APP_NAME)"
	@echo "Go Version: $$($(GO) version | awk '{print $$3}')"
	@echo "Platforms: $(PLATFORMS)"
	@echo "Docker Image: $(DOCKER_IMAGE)"

version: ## Show Go version
	$(GO) version
