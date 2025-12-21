SHELL := cmd.exe
.SHELLFLAGS := /c

BUILD_DIR := build
BINARY_NAME := seriallink.exe
VERSION := dev
COMMIT := dev
BUILD_DATE := dev
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)

.PHONY: help build lint install-tools ci fmt clean proto vet install

help:
	@echo "SerialLink - Cross-platform serial port agent with gRPC API"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the server binary to $(BUILD_DIR)/"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt"
	@echo "  vet            - Run go vet"
	@echo "  proto          - Generate protobuf Go code"
	@echo "  install-tools  - Install development tools"
	@echo "  ci             - Run all CI checks (fmt, vet, lint)"
	@echo "  clean          - Remove built binaries and build directory"

build:
	if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)\$(BINARY_NAME) ./
	@echo Build complete: $(BUILD_DIR)\$(BINARY_NAME)

install:
	go install -ldflags "$(LDFLAGS)" ./

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/serial.proto
	@echo Proto files generated

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	scoop install protobuf
	scoop install grpcurl
	scoop install make
	@echo Development tools installed

ci: fmt vet lint
	@echo All CI checks passed!

clean:
	if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@echo Clean complete
