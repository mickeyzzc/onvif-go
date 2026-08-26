# onvif-go — build/test/lint entrypoints (MiBeeNvr-aligned conventions)

BINARY_DIR := build

.PHONY: all build test test-verbose lint lint-install fmt vet check tools cross clean

all: build

build:
	go build ./...

test:
	go test -race ./...

test-verbose:
	go test -race -v ./...

lint:
	golangci-lint run

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

# Mechanical formatting: gofumpt + goimports (the formatters enabled in .golangci.yml)
fmt:
	golangci-lint fmt

vet:
	go vet ./...

check: lint test

# Build the helper CLIs into build/
tools:
	@mkdir -p $(BINARY_DIR)
	go build -trimpath -o $(BINARY_DIR)/ ./cmd/...

cross:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o $(BINARY_DIR)/ ./cmd/...

clean:
	rm -rf $(BINARY_DIR)
