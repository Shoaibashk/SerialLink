# 🔧 Development Guide

This guide covers everything you need to contribute to SerialLink.

---

## Prerequisites

| Tool | Version | Purpose |
| ------ | --------- | --------- |
| Go | 1.24+ | Core language |
| protoc | 3.x+ | Protocol Buffers compiler |
| golangci-lint | latest | Linting |
| make | any | Build automation |

### Installing Prerequisites

**Go:** Download from [go.dev](https://go.dev/dl/)

**protoc:**

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt install -y protobuf-compiler

# Windows (Chocolatey)
choco install protoc
```

**golangci-lint:**

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Build Commands

| Command | Description |
| --------- | ------------- |
| `make build` | Build binary to `./build/seriallink` |
| `make install` | Install to `$GOPATH/bin` |
| `make proto` | Regenerate protobuf code |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofmt |
| `make vet` | Run go vet |
| `make ci` | Run all CI checks (fmt, vet, lint) |
| `make clean` | Remove build artifacts |
| `make tools` | Install dev tools |

### Build with Version Info

```bash
make build VERSION=1.0.0 \
  COMMIT=$(git rev-parse --short HEAD) \
  BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
```

---

## Project Structure

```json
SerialLink/
├── main.go                 # Entry point
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command & global flags
│   ├── serve.go           # Start gRPC server
│   ├── scan.go            # Port discovery
│   ├── open.go            # Open port
│   ├── close.go           # Close port
│   ├── read.go            # Read data
│   ├── write.go           # Write data
│   ├── config.go          # Port configuration
│   ├── status.go          # Port status
│   ├── info.go            # Service info
│   └── version.go         # Version info
├── api/
│   ├── grpc_server.go     # gRPC service implementation
│   └── proto/
│       ├── serial.proto   # Protocol Buffer definitions
│       ├── serial.pb.go   # Generated message types
│       └── serial_grpc.pb.go  # Generated service stubs
├── internal/
│   └── serial/
│       ├── manager.go     # Session & connection management
│       ├── serial.go      # Low-level port operations
│       ├── reader.go      # Streaming read operations
│       ├── scanner.go     # Port discovery & detection
│       └── errors.go      # Error definitions
├── config/
│   ├── config.go          # Configuration schema & loading
│   └── agent.yaml         # Default configuration
├── build/                  # Build output directory
└── docs/                   # Documentation
```

---

## Architecture

```json
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                         │
│  (Python | C# | Node.js | Go | Web | Mobile | CLI)      │
└────────────────────┬────────────────────────────────────┘
                     │ gRPC Protocol
                     ▼
┌─────────────────────────────────────────────────────────┐
│              SerialLink Agent Service                   │
│  ┌─────────────────────────────────────────────────┐    │
│  │         gRPC Server (api/grpc_server.go)        │    │
│  │  - ListPorts, OpenPort, ClosePort               │    │
│  │  - Read, Write Operations                       │    │
│  │  - Bidirectional Streaming                      │    │
│  │  - Port Configuration & Status                  │    │
│  └────────────┬────────────────────────────────────┘    │
│               │                                         │
│  ┌────────────▼────────────────────────────────────┐    │
│  │     Serial Port Manager (internal/serial/)      │    │
│  │  - Session Management                           │    │
│  │  - Port Lifecycle Control                       │    │
│  │  - Exclusive Access Locking                     │    │
│  │  - Configuration & Statistics                   │    │
│  └────────────┬────────────────────────────────────┘    │
│               │                                         │
│  ┌────────────▼────────────────────────────────────┐    │
│  │   Port Scanner (internal/serial/scanner.go)     │    │
│  │  - Port Discovery                               │    │
│  │  - Hardware Detection                           │    │
│  │  - Port Metadata Collection                     │    │
│  └────────────┬────────────────────────────────────┘    │
└────────────────────┬────────────────────────────────────┘
                     │ Serial Interface
                     ▼
         ┌──────────────────────────┐
         │   Hardware Devices       │
         │  USB | COM | UART | BLE  │
         └──────────────────────────┘
```

---

## Core Concepts

### Session Management

- **Exclusive Access**: Ports support exclusive locking to prevent
  concurrent access conflicts
- **Session ID**: Each `OpenPort` call generates a unique session identifier
- **Resource Cleanup**: Automatic cleanup of abandoned sessions and port handles

### Concurrency Model

- **Thread-safe Operations**: All port operations use mutex protection
- **Session Isolation**: Each session maintains independent state
- **Streaming Support**: Concurrent readers/writers via multiplexing
- **Atomic Operations**: Statistics and status updates use atomic operations

### Error Handling

Errors are defined in `internal/serial/errors.go`:

| Error | Description |
| ------- | ------------- |
| `ErrPortNotFound` | Port cannot be found |
| `ErrPortAlreadyOpen` | Port already opened |
| `ErrPortClosed` | Operations on closed port |
| `ErrPortInUse` | Port locked by another client |
| `ErrInvalidSession` | Session ID mismatch |
| `ErrInvalidConfig` | Invalid port configuration |
| `ErrReadTimeout` / `ErrWriteTimeout` | Operation timeouts |
| `ErrPortDisconnected` | Port closed during operation |

---

## Working with Protobuf

### Regenerating Protobuf Code

After modifying `api/proto/serial.proto`:

```bash
make proto
```

This requires:

- `protoc` compiler
- `protoc-gen-go` plugin
- `protoc-gen-go-grpc` plugin

Install plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## Testing

### Running Tests

```bash
go test ./...
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o build/coverage.html
```

---

## Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Run `make lint` to catch issues
- Keep functions small and focused
- Document exported types and functions

---

## Debugging Tips

### Enable Debug Logging

```yaml
# config.yaml
logging:
  level: "debug"
```

### Use gRPC Reflection

gRPC reflection is enabled by default. Use tools like `grpcurl`:

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# Describe a service
grpcurl -plaintext localhost:50051 describe serial.SerialService

# Call a method
grpcurl -plaintext localhost:50051 serial.SerialService/ListPorts
```

---

## Release Process

1. Install GoReleaser locally (one time): `go install github.com/goreleaser/goreleaser/v2/cmd/goreleaser@latest`
2. Verify packaging locally without publishing: `goreleaser release --snapshot --clean`
3. Run full CI checks: `make ci`
4. Tag the version you just tested: `git tag vX.Y.Z`
5. Push the tag to trigger the GitHub Action release (builds Windows/Linux/macOS/FreeBSD; 386, amd64, arm64 where supported): `git push origin vX.Y.Z`
