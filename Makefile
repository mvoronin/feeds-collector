.PHONY: all generate clean goenv-setup

# Go version to use
GO_VERSION = 1.22.0

# Path to the sqlc configuration file
SQLC_CONFIG = sqlc.yaml

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Default target: generate
all: generate

# Setup goenv to use the specified Go version
goenv-setup:
	@echo "Setting up goenv..."
	@if [ -z "$(goenv versions --bare | grep '^$(GO_VERSION)$$')" ]; then \
		~/.local/bin/goenv install $(GO_VERSION); \
	fi
	@goenv local $(GO_VERSION)
	@export PATH="$(goenv root)/shims:$(PATH)"

# Generate Go code from SQL schema
generate: goenv-setup
	@echo "Generating Go code from SQL schema..."
	sqlc generate -f $(SQLC_CONFIG)

# Clean up generated files (optional)
clean:
	@echo "Cleaning up generated files..."
	rm -rf internal/models

mod-tidy:
	go mod tidy

lint:
	@echo "Linting the project..."
	cd $(ROOT_DIR)/src && golangci-lint --config .golangci.yml run -v --fix

# Build the project
build:
	@echo "Building the project..."
	cd $(ROOT_DIR)/src && go build -o ${ROOT_DIR}/bin/feedscollector $(ROOT_DIR)/src/cmd/collector/main.go

# Run the gatherer
run:
	@echo "Running the project..."
	./bin/gatherer --config conf1.yaml

# Run the API
run-api:
	@echo "Running the API..."
	./bin/api --config conf1.yaml
