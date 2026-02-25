# HBF Agent Makefile

# Variables
BINARY_NAME=hbf-agent
VERSION?=1.0.0
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/hbf-agent

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

.PHONY: all build clean test install uninstall run deps help

all: clean deps build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/hbf-agent

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## install: Install the binary and configuration
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo mkdir -p $(CONFIG_DIR)
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		sudo cp config/config.example.yaml $(CONFIG_DIR)/config.yaml; \
	fi
	@sudo cp deploy/systemd/hbf-agent.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@echo "Installation complete!"

## uninstall: Uninstall the binary and configuration
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo systemctl stop hbf-agent || true
	@sudo systemctl disable hbf-agent || true
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo rm -f /etc/systemd/system/hbf-agent.service
	@sudo systemctl daemon-reload
	@echo "Uninstall complete!"

## run: Run the agent locally
run: build
	@echo "Running $(BINARY_NAME)..."
	@sudo $(BUILD_DIR)/$(BINARY_NAME) --config config/config.example.yaml

## run-dev: Run the agent in development mode
run-dev:
	@echo "Running $(BINARY_NAME) in development mode..."
	@sudo $(GOCMD) run ./cmd/hbf-agent --config config/config.example.yaml --log-level debug

## lint: Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run --rm --privileged --network host \
		-v /etc/hbf-agent:/etc/hbf-agent \
		$(BINARY_NAME):$(VERSION)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
