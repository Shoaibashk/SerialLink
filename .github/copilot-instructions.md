# SerialLink Copilot Instructions

## Project Overview

SerialLink is a cross-platform Go service that manages serial port connections via gRPC API. Clients in any language (Python, C#, Node.js) connect over the network to interact with serial hardware.

```
Client Apps ──gRPC──▶ SerialLink Agent ──USB/UART──▶ Hardware
```

## Architecture

| Layer | Location | Responsibility |
|-------|----------|----------------|
| CLI | `cmd/` | Cobra commands, flags via Viper |
| gRPC API | `api/grpc_server.go` | Service implementation, interceptors |
| Serial Manager | `internal/serial/manager.go` | Session lifecycle, exclusive locking |
| Scanner | `internal/serial/scanner.go` | Port discovery, hardware detection |
| Proto definitions | `api/proto/proto/seriallink/v1/serial.proto` | gRPC service contract |

**Key flow:** `cmd/serve.go` → creates `Manager` + `Scanner` → injects into `SerialServer` → registers gRPC handlers.

## Build & Development

```powershell
# Build binary (outputs to build/seriallink.exe)
make build

# Run all CI checks (format, vet, lint)
make ci

# Regenerate protobuf after editing serial.proto
make proto
# Or directly: cd api/proto && .\generate.ps1 -Target go

# Install dev tools (golangci-lint, protoc plugins, buf)
make install-tools
```

**Version injection:** Build-time variables set via ldflags in Makefile (`VERSION`, `COMMIT`, `BUILD_DATE`).

## Code Conventions

### Error Handling
- Define domain errors in `internal/serial/errors.go` as sentinel values
- gRPC handlers convert to `status.Errorf(codes.X, ...)` in `api/grpc_server.go`
- Example: `ErrPortLocked` → `codes.FailedPrecondition`

### Session Pattern
```go
// OpenPort returns a Session with unique ID for all subsequent operations
session, err := manager.OpenPort(portName, config, clientID, exclusive)
// ClosePort requires matching session ID
err := manager.ClosePort(portName, sessionID)
```

### Configuration
- YAML config via Viper (`config/config.go`)
- Environment override: `SERIALLINK_ADDRESS`
- Flag binding pattern in `cmd/root.go` and `cmd/serve.go`

### Proto Module
- Located at `api/proto/` as a separate Go module (see `replace` directive in root `go.mod`)
- Import as: `pb "github.com/Shoaibashk/SerialLink-Proto/gen/go/seriallink/v1"`
- Uses buf for generation (see `api/proto/buf.gen.yaml`)

## Adding New Features

### New CLI Command
1. Create `cmd/<command>.go` with Cobra command struct
2. Register in `init()` with `rootCmd.AddCommand()`
3. Bind flags to Viper for config file support

### New gRPC Method
1. Add RPC to `api/proto/proto/seriallink/v1/serial.proto`
2. Run `make proto` to regenerate
3. Implement handler in `api/grpc_server.go`

### New Serial Operation
1. Add method to `Manager` in `internal/serial/manager.go`
2. Ensure thread-safety with `mu.Lock()`/`mu.RLock()`
3. Update session statistics atomically

## Streaming Handler Patterns

### Server-side Streaming (StreamRead)
Use when client needs continuous data from serial port:
```go
func (s *SerialServer) StreamRead(req *pb.StreamReadRequest, stream pb.SerialService_StreamReadServer) error {
    // 1. Create Reader with pub/sub pattern
    reader := serial.NewReader(s.manager, req.PortName, req.SessionID, chunkSize)
    defer reader.Stop()
    
    // 2. Start background read loop
    if err := reader.Start(stream.Context()); err != nil {
        return status.Errorf(codes.Internal, "failed to start reader: %v", err)
    }
    
    // 3. Subscribe and forward events to gRPC stream
    subscription := reader.Subscribe()
    for {
        select {
        case <-stream.Context().Done():
            return nil  // Client disconnected
        case event, ok := <-subscription:
            if !ok { return nil }
            if err := stream.Send(&pb.StreamReadResponse{Chunk: chunk}); err != nil {
                return err
            }
        }
    }
}
```

### Client-side Streaming (StreamWrite)
Use when client sends multiple chunks:
```go
func (s *SerialServer) StreamWrite(stream pb.SerialService_StreamWriteServer) error {
    var totalBytes uint64
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            // Client done - send final response
            return stream.SendAndClose(&pb.StreamWriteResponse{
                TotalBytesWritten: totalBytes,
            })
        }
        if err != nil { return err }
        
        // Process chunk
        n, _ := s.manager.Write(portName, sessionID, chunk.Data)
        atomic.AddUint64(&totalBytes, uint64(n))
    }
}
```

### Bidirectional Streaming
Split into goroutines - one for reads, one for writes:
```go
func (s *SerialServer) BiDirectionalStream(stream pb.SerialService_BiDirectionalStreamServer) error {
    errChan := make(chan error, 2)
    
    // Goroutine 1: Handle incoming writes from client
    go s.handleBiDirectionalWrites(stream, &portName, &sessionID, errChan)
    
    // Main: Handle outgoing reads to client
    return s.handleBiDirectionalReads(stream, ctx, errChan, reader, portName)
}
```

**Key pattern:** Use `serial.Reader` for pub/sub data distribution (see [internal/serial/reader.go](../internal/serial/reader.go)).

## Testing Locally

```powershell
# Start server
.\build\seriallink.exe serve

# In another terminal - test with grpcurl
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 seriallink.v1.SerialService/ListPorts
```

## Key Files Reference

- Entry point: [main.go](../main.go)
- gRPC implementation: [api/grpc_server.go](../api/grpc_server.go)
- Session management: [internal/serial/manager.go](../internal/serial/manager.go)
- Port configuration types: [internal/serial/serial.go](../internal/serial/serial.go)
- Error definitions: [internal/serial/errors.go](../internal/serial/errors.go)
