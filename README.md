# SerialLink

SerialLink is a **cross-platform serial port background service** that runs on Windows, Linux, and Raspberry Pi. It manages all serial hardware and exposes a public gRPC API for any client - Python, C#, Node.js, Web, Mobile, or CLI.

```text
           Any Client (Python | C# | Web | Mobile | CLI)
                              |
                              | gRPC / WebSocket
                              |
                    +-------------------+
                    |   SerialLink Agent  |
                    | (Background Svc)  |
                    +-------------------+
                              |
                              | USB / COM / UART
                              |
                       Hardware Devices
```

**No UI. No frontend. Just a rock-solid hardware agent.** 💪

## Features

### 🔌 Serial Port Management

- **Auto-detect ports** - Discover all USB, native, Bluetooth, and virtual serial ports
- **Open/Close** - Manage port lifecycle with exclusive locking
- **Read/Write** - Send and receive data with timeout support
- **Streaming** - Real-time bidirectional data streaming
- **Hot-plug support** - Detect port changes on the fly

### 🌐 Network API

- **gRPC API** - High-performance, strongly-typed API
- **Streaming support** - Server, client, and bidirectional streaming
- **Cross-language** - Use from any language with gRPC support

### 🔐 Security

- **TLS encryption** - Secure transport layer
- **Port locking** - Exclusive access control
- **Network binding** - Control service exposure

### ⚙️ System Integration

- **Windows Service** - Run as Windows background service
- **systemd service** - Run as Linux/Raspberry Pi daemon
- **Auto-start** - Start on system boot
- **Logging** - Comprehensive audit logging
