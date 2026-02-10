.PHONY: all build test test-coverage lint clean install

BINARY_NAME = nuther
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

all: lint test build

build:
	go build $(LDFLAGS) ./cmd/nuther

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	go vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe coverage.out coverage.html

install:
	go install $(LDFLAGS) ./cmd/nuther
