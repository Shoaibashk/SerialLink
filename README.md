# SerialLink

A professional-grade, cross-platform serial port management service providing robust hardware abstraction and remote access capabilities.

## Overview

SerialLink is an enterprise-ready background service that provides unified management of serial port devices across Windows, Linux, and embedded systems. It exposes a high-performance gRPC API enabling seamless integration with applications written in any language, including Python, C#, Node.js, JavaScript, and Go.

The service architecture separates hardware management from client applications, enabling centralized port control, concurrent access management, and secure network exposure.

```text
           Client Applications (Python | C# | Node.js | Go | CLI)
                              |
                              | gRPC / Standard Protocol
                              |
                    +-------------------+
                    |   SerialLink       |
                    |   Agent Service   |
                    | (Background)      |
                    +-------------------+
                              |
                              | USB / COM / UART
                              |
                    Serial Port Devices
```

## Key Capabilities

### Serial Port Management

- **Port Discovery** - Automated detection of USB, native COM, Bluetooth, and virtual serial ports
- **Lifecycle Control** - Secure open/close operations with exclusive port access locking
- **Data Operations** - Configurable read/write operations with timeout handling
- **Real-time Streaming** - Bidirectional data streaming with server-initiated notifications
- **Hot-swap Detection** - Dynamic port availability monitoring

### Remote Access API

- **gRPC Protocol** - High-performance, language-agnostic service interface
- **Streaming Capabilities** - Server, client, and bidirectional streaming models
- **Protocol Flexibility** - Support for multiple transport layers

### Security & Access Control

- **Transport Security** - TLS/SSL encryption for network communications
- **Access Control** - Exclusive port locking mechanisms
- **Network Configuration** - Granular service endpoint binding
- **Audit Logging** - Complete operational event logging

### System Integration

- **Windows Service** - Native Windows service installation and management
- **systemd Support** - Linux and Raspberry Pi daemon integration
- **Auto-launch** - System boot integration
- **Operational Logging** - Detailed logging for troubleshooting and monitoring

## Architecture & Usage

### Service Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                         │
│  (Python | C# | Node.js | Go | Web | Mobile | CLI)    │
└────────────────────┬────────────────────────────────────┘
                     │ gRPC Protocol
                     ▼
┌─────────────────────────────────────────────────────────┐
│              SerialLink Agent Service                   │
│  ┌─────────────────────────────────────────────────┐   │
│  │         gRPC Server (api/grpc_server.go)        │   │
│  │  - ListPorts, OpenPort, ClosePort               │   │
│  │  - Read, Write Operations                       │   │
│  │  - Bidirectional Streaming                      │   │
│  │  - Port Configuration & Status                  │   │
│  └────────────┬────────────────────────────────────┘   │
│               │                                          │
│  ┌────────────▼────────────────────────────────────┐   │
│  │     Serial Port Manager (internal/serial/)      │   │
│  │  - Session Management                           │   │
│  │  - Port Lifecycle Control                       │   │
│  │  - Exclusive Access Locking                     │   │
│  │  - Configuration & Statistics                   │   │
│  └────────────┬────────────────────────────────────┘   │
│               │                                          │
│  ┌────────────▼────────────────────────────────────┐   │
│  │   Port Scanner (internal/serial/scanner.go)     │   │
│  │  - Port Discovery                               │   │
│  │  - Hardware Detection                           │   │
│  │  - Port Metadata Collection                     │   │
│  └────────────┬────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │ Serial Interface
                     ▼
         ┌──────────────────────────┐
         │   Hardware Devices       │
         │  USB | COM | UART | BLE  │
         └──────────────────────────┘
```

### Core Components

**API Layer** (`api/`)
- `grpc_server.go` - gRPC service implementation with all RPC handlers
- `proto/serial.proto` - Protocol Buffer definitions for service interface
- `proto/serial.pb.go` & `serial_grpc.pb.go` - Generated protobuf code

**Serial Port Management** (`internal/serial/`)
- `manager.go` - Session and connection lifecycle management
- `serial.go` - Low-level serial port operations
- `reader.go` - Streaming read operations
- `scanner.go` - Port discovery and hardware detection
- `errors.go` - Error definitions and handling

**Command Interface** (`cmd/`)
- `serve.go` - Service startup and lifecycle
- `open.go` / `close.go` - Port open/close CLI commands
- `read.go` / `write.go` - Data transfer operations
- `scan.go` - Port scanning and discovery
- `config.go` - Configuration management
- `status.go` / `info.go` - Status and information queries

**Configuration** (`config/`)
- `config.go` - Configuration schema and loading
- `agent.yaml` - Service configuration file

### Usage Workflow

#### 1. Service Deployment
```
Install → Configure → Start Agent Service → Listen on gRPC Port
```

#### 2. Client Interaction Flow
```
Client → Connect to gRPC → Authenticate → Discover Ports
    ↓
    → Open Port (Session Created) → Configure Port Settings
    ↓
    → Read/Write Data or Stream Data
    ↓
    → Close Port (Session Terminated) → Cleanup Resources
```

#### 3. Session Management
- **Exclusive Access**: Ports support exclusive locking to prevent concurrent access
- **Session ID**: Each port operation is tracked with a unique session identifier
- **Resource Cleanup**: Automatic cleanup of abandoned sessions and port handles

#### 4. API Endpoints

**Discovery Operations**
- `ListPorts` - Enumerate all available serial ports
- `GetPortInfo` - Get detailed information about a specific port

**Port Management**
- `OpenPort` - Acquire exclusive or shared access to a port
- `ClosePort` - Release port resources
- `GetPortStatus` - Query current port state
- `ConfigurePort` - Set port parameters (baud rate, parity, etc.)

**Data Transfer**
- `Write` - Single write operation
- `Read` - Single read operation with timeout
- `StreamRead` - Server-side streaming of incoming data
- `StreamWrite` - Client-side streaming of outgoing data
- `BiDirectionalStream` - Full-duplex streaming

**Diagnostics**
- `Ping` - Service health check
- `GetAgentInfo` - Service version and uptime information

### Concurrency Model

- **Thread-safe Operations**: All port operations use mutex protection
- **Session Isolation**: Each session maintains independent state
- **Streaming Support**: Concurrent readers/writers on single port via multiplexing
- **Atomic Operations**: Statistics and status updates use atomic operations
