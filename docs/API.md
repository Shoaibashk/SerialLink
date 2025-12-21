# 🌐 gRPC API Reference

SerialLink exposes a high-performance gRPC API for serial port management.

---

## Quick Start

### Connect with grpcurl

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# List available RPC methods
grpcurl -plaintext localhost:50051 list seriallink.v1.SerialService

# List available serial ports
grpcurl -plaintext localhost:50051 seriallink.v1.SerialService/ListPorts

# Open a serial port (115200 baud, 8 data bits, 1 stop bit)
grpcurl -plaintext -d '{
  "port_name": "COM3",
  "config": {
    "baud_rate": 115200,
    "data_bits": 8,
    "stop_bits": 1,
    "parity": 0,
    "flow_control": 0
  }
}' localhost:50051 seriallink.v1.SerialService/OpenPort

# Write data to port (returns session_id from OpenPort response)
grpcurl -plaintext -d '{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca",
  "data": "SGVsbG8gQXJkdWlubyE=",
  "flush": true
}' localhost:50051 seriallink.v1.SerialService/Write

# Read data from port
grpcurl -plaintext -d '{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca",
  "max_bytes": 1024,
  "timeout_ms": 2000
}' localhost:50051 seriallink.v1.SerialService/Read

# Close the port
grpcurl -plaintext -d '{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca"
}' localhost:50051 seriallink.v1.SerialService/ClosePort
```

### Proto File Location

The complete service definition is in [`api/proto/serial.proto`](../api/proto/serial.proto).

---

## Service Methods

### Port Discovery

#### `ListPorts`

Enumerate all available serial ports.

```protobuf
rpc ListPorts(ListPortsRequest) returns (ListPortsResponse)
```

**Request:** `ListPortsRequest` (empty)

**Response:**

```json
{
  "ports": [
    {
      "name": "COM3",
      "description": "Arduino Uno (COM3)",
      "hardwareId": "USB\\VID_2341&PID_0043",
      "product": "Arduino Uno (COM3)",
      "serialNumber": "95632313234351211231",
      "portType": "PORT_TYPE_USB"
    }
  ]
}
```

---

#### `GetPortInfo`

Get detailed information about a specific port.

```protobuf
rpc GetPortInfo(GetPortInfoRequest) returns (PortInfo)
```

**Request:**

```json
{
  "port_name": "COM1"
}
```

---

### Port Management

#### `OpenPort`

Open a serial port with exclusive or shared access.

```protobuf
rpc OpenPort(OpenPortRequest) returns (OpenPortResponse)
```

**Request:**

```json
{
  "port_name": "COM3",
  "config": {
    "baud_rate": 115200,
    "data_bits": 8,
    "stop_bits": 1,
    "parity": 0,
    "flow_control": 0
  }
}
```

**Response:**

```json
{
  "success": true,
  "message": "port opened successfully",
  "sessionId": "24189592-1c7f-4147-8679-87bf033c2bca"
}
```

> ⚠️ Save the `sessionId` — you'll need it for subsequent operations.

---

#### `ClosePort`

Close a port and release resources.

```protobuf
rpc ClosePort(ClosePortRequest) returns (ClosePortResponse)
```

**Request:**

```json
{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca"
}
```

---

#### `GetPortStatus`

Query current port state and statistics.

```protobuf
rpc GetPortStatus(GetPortStatusRequest) returns (PortStatus)
```

**Response:**

```json
{
  "portName": "COM3",
  "isOpen": true,
  "lockedBy": "default-client",
  "sessionId": "24189592-1c7f-4147-8679-87bf033c2bca",
  "currentConfig": {
    "baudRate": 115200,
    "dataBits": "DATA_BITS_8",
    "stopBits": "STOP_BITS_1",
    "parity": "PARITY_NONE",
    "flowControl": "FLOW_CONTROL_NONE"
  },
  "statistics": {
    "bytesSent": "14",
    "bytesReceived": "565",
    "openedAt": "1766343135",
    "lastActivity": "1766343150"
  }
}
```

---

#### `ConfigurePort`

Change port parameters on an open port.

```protobuf
rpc ConfigurePort(ConfigurePortRequest) returns (ConfigurePortResponse)
```

**Request:**

```json
{
  "port_name": "COM1",
  "session_id": "...",
  "baud_rate": 115200
}
```

---

### Data Transfer

#### `Write`

Write data to a port (single operation).

```protobuf
rpc Write(WriteRequest) returns (WriteResponse)
```

**Request:**

```json
{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca",
  "data": "SGVsbG8gQXJkdWlubyE=",
  "flush": true
}
```

**Response:**

```json
{
  "success": true,
  "bytesWritten": 14,
  "message": "data written successfully"
}
```

---

#### `Read`

Read data from a port with timeout.

```protobuf
rpc Read(ReadRequest) returns (ReadResponse)
```

**Request:**

```json
{
  "port_name": "COM3",
  "session_id": "24189592-1c7f-4147-8679-87bf033c2bca",
  "max_bytes": 1024,
  "timeout_ms": 2000
}
```

**Response:**

```json
{
  "success": true,
  "data": "MC4wMCA0LjI0DQowLjAwIDMuMzkNCjAuMDAgMi43Mg0KMC4wMCAyLjE3DQowLjAwIDEuNzQNCjAuMDAgMS4zOQ0KMC4wMCAxLjExDQowLjAwIDAuODkNCjAuMDAgMC43MQ0K...",
  "bytesRead": 565,
  "message": "data read successfully"
}
```

---

### Streaming

#### `StreamRead`

Server-side streaming of incoming data.

```protobuf
rpc StreamRead(StreamReadRequest) returns (stream ReadResponse)
```

Continuous stream of data as it arrives on the port.

---

#### `StreamWrite`

Client-side streaming for batch writes.

```protobuf
rpc StreamWrite(stream WriteRequest) returns (StreamWriteResponse)
```

Send multiple data chunks efficiently.

---

#### `BiDirectionalStream`

Full-duplex communication.

```protobuf
rpc BiDirectionalStream(stream WriteRequest) returns (stream ReadResponse)
```

Simultaneously send and receive data.

---

### Diagnostics

#### `Ping`

Service health check.

```protobuf
rpc Ping(PingRequest) returns (PingResponse)
```

**Response:**

```json
{
  "message": "pong",
  "timestamp": "2025-12-21T10:30:00Z"
}
```

---

#### `GetAgentInfo`

Get service version and status.

```protobuf
rpc GetAgentInfo(AgentInfoRequest) returns (AgentInfo)
```

**Response:**

```json
{
  "version": "1.0.0",
  "commit": "abc1234",
  "build_date": "2025-12-21",
  "uptime_seconds": 3600,
  "open_ports": 2,
  "features": ["tls", "streaming", "reflection"]
}
```

---

## Client Examples

### Codegen: Python

```python
import grpc
from serial_pb2_grpc import SerialServiceStub
from serial_pb2 import (
    ListPortsRequest,
    OpenPortRequest,
    WriteRequest,
    ReadRequest,
    ClosePortRequest
)

# Connect
channel = grpc.insecure_channel('localhost:50051')
stub = SerialServiceStub(channel)

# List ports
ports = stub.ListPorts(ListPortsRequest())
for port in ports.ports:
    print(f"{port.name}: {port.description}")

# Open port
response = stub.OpenPort(OpenPortRequest(
    port_name="COM1",
    baud_rate=9600
))
session_id = response.session_id

# Write data
stub.Write(WriteRequest(
    port_name="COM1",
    session_id=session_id,
    data=b"Hello, Arduino!",
    flush=True
))

# Read response
response = stub.Read(ReadRequest(
    port_name="COM1",
    session_id=session_id,
    max_bytes=1024,
    timeout_ms=1000
))
print(f"Received: {response.data}")

# Close port
stub.ClosePort(ClosePortRequest(
    port_name="COM1",
    session_id=session_id
))
```

---

### Codegen: Node.js

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const packageDefinition = protoLoader.loadSync('serial.proto');
const serial = grpc.loadPackageDefinition(packageDefinition).serial;

const client = new serial.SerialService(
  'localhost:50051',
  grpc.credentials.createInsecure()
);

// List ports
client.ListPorts({}, (err, response) => {
  response.ports.forEach(port => {
    console.log(`${port.name}: ${port.description}`);
  });
});

// Open port
client.OpenPort({
  port_name: 'COM1',
  baud_rate: 9600
}, (err, response) => {
  const sessionId = response.session_id;
  
  // Write data
  client.Write({
    port_name: 'COM1',
    session_id: sessionId,
    data: Buffer.from('Hello!'),
    flush: true
  }, (err, response) => {
    console.log(`Wrote ${response.bytes_written} bytes`);
  });
});
```

---

### Codegen: Go

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "github.com/Shoaibashk/SerialLink/api/proto"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewSerialServiceClient(conn)
    
    // List ports
    ports, err := client.ListPorts(context.Background(), &pb.ListPortsRequest{})
    if err != nil {
        log.Fatal(err)
    }
    
    for _, port := range ports.Ports {
        log.Printf("%s: %s", port.Name, port.Description)
    }
    
    // Open port
    resp, err := client.OpenPort(context.Background(), &pb.OpenPortRequest{
        PortName: "COM1",
        BaudRate: 9600,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    sessionID := resp.SessionId
    log.Printf("Opened with session: %s", sessionID)
}
```

---

### C\#

```csharp
using Grpc.Net.Client;
using Serial;

var channel = GrpcChannel.ForAddress("http://localhost:50051");
var client = new SerialService.SerialServiceClient(channel);

// List ports
var ports = await client.ListPortsAsync(new ListPortsRequest());
foreach (var port in ports.Ports)
{
    Console.WriteLine($"{port.Name}: {port.Description}");
}

// Open port
var openResponse = await client.OpenPortAsync(new OpenPortRequest
{
    PortName = "COM1",
    BaudRate = 9600
});

var sessionId = openResponse.SessionId;

// Write data
await client.WriteAsync(new WriteRequest
{
    PortName = "COM1",
    SessionId = sessionId,
    Data = Google.Protobuf.ByteString.CopyFromUtf8("Hello!"),
    Flush = true
});
```

---

## Error Codes

|Code|Status|Description|
|------|--------|-------------|
|`NOT_FOUND`|Port not found|The specified port doesn't exist|
|`ALREADY_EXISTS`|Port already open|Port is locked by another session|
|`PERMISSION_DENIED`|Invalid session|Session ID doesn't match|
|`INVALID_ARGUMENT`|Invalid config|Bad port configuration|
|`DEADLINE_EXCEEDED`|Timeout|Read/write operation timed out|
|`UNAVAILABLE`|Port disconnected|Port was disconnected|

---

## Generating Client Code

### Python

```bash
pip install grpcio grpcio-tools
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. serial.proto
```

### Node.js

```bash
npm install @grpc/grpc-js @grpc/proto-loader
# Dynamic loading, no code generation needed
```

### Go

```bash
protoc --go_out=. --go-grpc_out=. serial.proto
```

### Codegen: C\#

```bash
dotnet add package Grpc.Net.Client
dotnet add package Google.Protobuf
dotnet add package Grpc.Tools
# Add .proto to .csproj with <Protobuf Include="serial.proto" />
```
