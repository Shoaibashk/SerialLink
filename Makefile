.PHONY: help build test lint install-tools ci fmt clean

help:
	@echo "SerialLink - cross-platform serial port agent"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  install        - Install the binary to \$$GOPATH/bin"
	@echo "  test           - Run unit tests"
	@echo "  coverage       - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt"
	@echo "  vet            - Run go vet"
	@echo "  install-tools  - Install development tools (golangci-lint)"
	@echo "  ci             - Run all CI checks (fmt, vet, lint, test)"
	@echo "  clean          - Remove built binaries"

build:
	go build -o seriallink ./

install:
	go install -ldflags "-X main.Version=$(shell git describe --tags --always)" ./

test:
	go test ./... -v -race

coverage:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

ci: fmt vet lint test
	@echo "All CI checks passed!"

clean:
	rm -f seriallink coverage.out coverage.html
