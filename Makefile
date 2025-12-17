BUILD_DIR := build

.PHONY: help build test coverage lint install-tools ci fmt clean

help:
	@echo "SerialLink - cross-platform serial port agent"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary to $(BUILD_DIR)/"
	@echo "  install        - Install the binary to \$$GOPATH/bin"
	@echo "  test           - Run unit tests"
	@echo "  coverage       - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt"
	@echo "  vet            - Run go vet"
	@echo "  install-tools  - Install development tools (golangci-lint)"
	@echo "  ci             - Run all CI checks (fmt, vet, lint, test)"
	@echo "  clean          - Remove built binaries and build directory"

build:
	powershell -Command "if (!(Test-Path $(BUILD_DIR))) { New-Item -ItemType Directory -Path $(BUILD_DIR) }"
	go build -o $(BUILD_DIR)\\seriallink.exe ./

install:
	go install -ldflags "-X main.Version=$(shell git describe --tags --always)" ./

test:
	go test ./... -v

coverage:
	powershell -Command "if (!(Test-Path $(BUILD_DIR))) { New-Item -ItemType Directory -Path $(BUILD_DIR) }"
	go test ./... -v -coverprofile=$(BUILD_DIR)\\coverage.out
	go tool cover -html=$(BUILD_DIR)\\coverage.out -o $(BUILD_DIR)\\coverage.html
	@echo "Coverage report generated: $(BUILD_DIR)\\coverage.html"

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
	powershell -Command "if (Test-Path $(BUILD_DIR)) { Remove-Item -Recurse -Force $(BUILD_DIR) }"
