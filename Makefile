.PHONY: build test run clean lint fmt coverage help

.DEFAULT_GOAL := help

build:
	@echo "Building the Otel Collector..."
	go build -o bin ./...
	@echo "Build complete: bin/otel-collector"

test:
	@echo "Running tests..."
	go test -count=1 -v ./...

run:
	@if [ -z "$$ATTRIBUTE_KEY" ]; then \
		echo "Error: ATTRIBUTE_KEY environment variable must be set"; \
		echo "Example: ATTRIBUTE_KEY=foo make run"; \
		exit 1; \
	fi
	@echo "Starting application..."
	go run ./...

run-example:
	@echo "Starting with example configuration..."
	OTEL_ENABLED=false ATTRIBUTE_KEY=foo WINDOW_DURATION=30s go run ./...

clean:
	@echo "Cleaning up..."
	rm -rf bin/otel-collector
	@echo "Cleanup complete."

run-client:
	@echo "Starting example client to send logs..."
	go run example_client.go

help:
	@echo "OTLP Log Parser - Make Targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

