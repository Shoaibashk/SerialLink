# 🚢 Deployment Guide

Deploy SerialLink as a background service on your target platform.

---

## Linux (systemd)

### 1. Install Binary

```bash
# Option A: Build from source
make build
sudo cp build/seriallink /usr/local/bin/

# Option B: Go install
go install github.com/Shoaibashk/SerialLink@latest
sudo cp $(go env GOPATH)/bin/SerialLink /usr/local/bin/seriallink
```

### 2. Create Service File

Create `/etc/systemd/system/seriallink.service`:

```ini
[Unit]
Description=SerialLink Serial Port Service
Documentation=https://github.com/Shoaibashk/SerialLink
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/seriallink serve
Restart=on-failure
RestartSec=5
User=root
StandardOutput=journal
StandardError=journal

# Security hardening (optional)
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/dev

[Install]
WantedBy=multi-user.target
```

### 3. Enable and Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable seriallink
sudo systemctl start seriallink

# Check status
sudo systemctl status seriallink

# View logs
sudo journalctl -u seriallink -f
```

### 4. Serial Port Permissions

```bash
# Add service user to dialout group (if not running as root)
sudo usermod -a -G dialout <service-user>
```

---

## Windows (NSSM)

### 1. Download NSSM

Download [NSSM](https://nssm.cc/download) and extract to a permanent location.

### 2. Install Service

```powershell
# Install as service
nssm install SerialLink "C:\path\to\seriallink.exe" serve

# Configure service (optional)
nssm set SerialLink DisplayName "SerialLink Agent"
nssm set SerialLink Description "Serial port management service"
nssm set SerialLink Start SERVICE_AUTO_START

# Start the service
nssm start SerialLink
```

### 3. Service Management

```powershell
# Status
nssm status SerialLink

# Stop
nssm stop SerialLink

# Restart
nssm restart SerialLink

# Remove
nssm remove SerialLink confirm
```

### Alternative: Windows Service (sc.exe)

```powershell
# Create service
sc.exe create SerialLink binPath= "C:\path\to\seriallink.exe serve" start= auto

# Start
sc.exe start SerialLink

# Stop
sc.exe stop SerialLink
```

---

## Raspberry Pi

Same as Linux, with additional considerations:

### 1. Enable Serial Port

```bash
sudo raspi-config
# Interface Options → Serial Port → Enable
```

### 2. User Permissions

```bash
sudo usermod -a -G dialout $USER
# Log out and back in
```

### 3. Common Port Paths

| Port | Path |
| ------ | ------ |
| GPIO UART | `/dev/ttyAMA0` or `/dev/serial0` |
| USB Serial | `/dev/ttyUSB0` |
| USB ACM | `/dev/ttyACM0` |

---

## Docker

### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o seriallink .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/seriallink .
EXPOSE 50051
CMD ["./seriallink", "serve"]
```

### Run with Device Access

```bash
docker build -t seriallink .
docker run -d \
  --name seriallink \
  --device=/dev/ttyUSB0 \
  -p 50051:50051 \
  seriallink
```

> ⚠️ **Note:** Docker containers need `--device` flag to access serial ports.

---

## TLS Configuration

### Generate Self-Signed Certificate (Testing)

```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout key.pem \
  -out cert.pem \
  -days 365 \
  -nodes \
  -subj "/CN=seriallink"
```

### Start with TLS

```bash
seriallink serve --tls --tls-cert cert.pem --tls-key key.pem
```

### Production TLS

For production, use certificates from a trusted CA:

1. **Let's Encrypt** (free, automated)
2. **Your organization's CA**
3. **Commercial CA** (DigiCert, etc.)

Example with Let's Encrypt:

```bash
# Install certbot
sudo apt install certbot

# Get certificate
sudo certbot certonly --standalone -d serial.yourdomain.com

# Use certificates
seriallink serve --tls \
  --tls-cert /etc/letsencrypt/live/serial.yourdomain.com/fullchain.pem \
  --tls-key /etc/letsencrypt/live/serial.yourdomain.com/privkey.pem
```

---

## Configuration

### Config File Locations

SerialLink searches for config in order:

1. `$HOME/.seriallink/config.yaml`
2. `./config.yaml`
3. `/etc/seriallink/config.yaml`

### Production Configuration

```yaml
server:
  grpc_address: "0.0.0.0:50051"
  max_connections: 100
  connection_timeout: 30

tls:
  enabled: true
  cert_file: "/etc/seriallink/cert.pem"
  key_file: "/etc/seriallink/key.pem"

serial:
  defaults:
    baud_rate: 9600
    data_bits: 8
    stop_bits: 1
    parity: "none"
    flow_control: "none"
  scan_interval: 5
  allow_shared_access: false

logging:
  level: "info"
  format: "json"           # JSON for log aggregation
  file: "/var/log/seriallink.log"
  max_size: 100            # MB
  max_backups: 3
  max_age: 30              # days
  compress: true

metrics:
  enabled: true
  address: "0.0.0.0:9090"
  path: "/metrics"
```

---

## Health Checks

### gRPC Health Check

```bash
grpcurl -plaintext localhost:50051 serial.SerialService/Ping
```

### HTTP Health Endpoint (if metrics enabled)

```bash
curl http://localhost:9090/metrics
```

### Kubernetes Probes

```yaml
livenessProbe:
  exec:
    command:
      - grpcurl
      - -plaintext
      - localhost:50051
      - serial.SerialService/Ping
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  exec:
    command:
      - grpcurl
      - -plaintext
      - localhost:50051
      - serial.SerialService/Ping
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u seriallink -n 50

# Check binary permissions
ls -la /usr/local/bin/seriallink

# Test manually
/usr/local/bin/seriallink serve
```

### Port Access Denied

```bash
# Check port permissions
ls -la /dev/ttyUSB0

# Add user to dialout group
sudo usermod -a -G dialout $USER
```

### Connection Refused

```bash
# Check if service is listening
ss -tlnp | grep 50051

# Check firewall
sudo ufw status
sudo ufw allow 50051/tcp
```
