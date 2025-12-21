# 🔌 SerialLink
<!-- cSpell:ignore UART Shoaibashk SerialLink seriallink SERIALLINK -->

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/Shoaibashk/SerialLink/pulls)

> **A cross-platform serial port management service with gRPC API**

SerialLink runs as a background service, managing serial port connections on
**Windows, Linux, and Raspberry Pi**.
It exposes a high-performance gRPC API so any language—Python, C#, Node.js,
Go—can talk to your hardware over the network.

```text
Your App (any language)  ──gRPC──▶  SerialLink  ──USB/UART──▶  Hardware
```

---

## ✨ Why SerialLink?

| Problem | SerialLink Solution |
| --------- | --------------------- |
| Port locked to one process | 🔒 Session locks + clean handoffs |
| Different APIs per language/platform | ⚡ One gRPC API, works everywhere |
| No remote access to serial devices | 🌐 Network-accessible serial ports |
| Complex port configuration | 🛠️ Simple CLI + YAML config |

---

## 🚀 Quick Start

```bash
# Install
go install github.com/Shoaibashk/SerialLink@latest

# Scan for ports
seriallink scan

# Start the server
seriallink serve

# In another terminal: open a port
seriallink open COM1 --baud 115200
```

That's it. Your serial port is now accessible via gRPC at `localhost:50051`.

---

## 📦 Installation

### Go Install (recommended)

```bash
go install github.com/Shoaibashk/SerialLink@latest
```

### Build from Source

```bash
git clone https://github.com/Shoaibashk/SerialLink.git
cd SerialLink && make build
```

### Download Binary

[Releases](https://github.com/Shoaibashk/SerialLink/releases)

---

## 🛠️ CLI Reference

| Command | Description |
| --------- | ------------- |
| `seriallink serve` | Start the gRPC server |
| `seriallink scan` | List available serial ports |
| `seriallink open <port>` | Open a port with config |
| `seriallink close <port>` | Close and release a port |
| `seriallink read <port>` | Read data from port |
| `seriallink write <port> <data>` | Write data to port |
| `seriallink config <port>` | View/modify port settings |
| `seriallink status <port>` | Get port statistics |
| `seriallink info` | Service information |
| `seriallink version` | Version info |

### Common Examples

```bash
# Start server with custom address
seriallink serve --address 0.0.0.0:50052

# Scan with JSON output (great for scripts)
seriallink scan --json

# Open with full config
seriallink open /dev/ttyUSB0 --baud 115200 --data-bits 8 --parity none

# Read with timeout
seriallink read COM1 --timeout 5000 --format hex

# Write hex data
seriallink write COM1 --hex "48454C4C4F"
```

> 💡 **Tip:** Set `SERIALLINK_ADDRESS` env var to skip `--address` on every command.

---

## 🌐 gRPC API

Connect from any language:

```python
# Python
import grpc
channel = grpc.insecure_channel('localhost:50051')
stub = SerialServiceStub(channel)
ports = stub.ListPorts(ListPortsRequest())
```

```javascript
// Node.js
const client = new SerialService('localhost:50051', grpc.credentials.createInsecure());
client.ListPorts({}, (err, response) => console.log(response.ports));
```

```csharp
// C#
var channel = GrpcChannel.ForAddress("http://localhost:50051");
var client = new SerialService.SerialServiceClient(channel);
var ports = await client.ListPortsAsync(new ListPortsRequest());
```

**Key Methods:**

- `ListPorts`
- `OpenPort`
- `ClosePort`
- `Read`
- `Write`
- `StreamRead`
- `BiDirectionalStream`

📖 **Full API docs:** [docs/API.md](docs/API.md)

---

## ⚙️ Configuration

```yaml
# ~/.seriallink/config.yaml
server:
  grpc_address: "0.0.0.0:50051"

serial:
  defaults:
    baud_rate: 9600
    data_bits: 8
    parity: "none"

logging:
  level: "info"    # debug | info | warn | error
```

Config locations: `~/.seriallink/config.yaml` → `./config.yaml` → `/etc/seriallink/config.yaml`

---

## 📚 Documentation

| Doc | Description |
| ----- | ------------- |
| [**API Reference**](docs/API.md) | gRPC methods, client examples, error codes |
| [**Deployment Guide**](docs/DEPLOYMENT.md) | systemd, Windows service, Docker, TLS |
| [**Development Guide**](docs/DEVELOPMENT.md) | Building, architecture, contributing |

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

```bash
# Clone and build
git clone https://github.com/Shoaibashk/SerialLink.git
cd SerialLink
make build

# Run checks before submitting
make ci
```

**Ways to contribute:**

- 🐛 Report bugs via [Issues](https://github.com/Shoaibashk/SerialLink/issues)
- 💡 Suggest features
- 📖 Improve documentation
- 🔧 Submit pull requests

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed setup.

---

## 📄 License

[Apache License 2.0](LICENSE) — use it freely in personal and commercial projects.

---

Made with ❤️ by Shoaibashk for hardware hackers, IoT builders, and serial port wranglers.

[⭐ Star us on GitHub](https://github.com/Shoaibashk/SerialLink/stargazers)
