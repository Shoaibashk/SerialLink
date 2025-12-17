# SerialLink

SerialLink is a cross-platform serial port agent written in Go. It provides a CLI to manage and interact with serial port connections across Windows, Linux, and macOS.

## Features

- **Cross-platform**: Works on Windows, Linux, and macOS
- **Easy to use**: Simple CLI interface powered by Cobra
- **Configurable**: Support for YAML config files and environment variables via Viper
- **Well-tested**: Comprehensive unit tests with table-driven test patterns
- **Linted**: Enforced code quality with golangci-lint
- **CI/CD ready**: GitHub Actions workflow for automated testing

## Prerequisites

- Go 1.21 or later

## Installation

### From source

```bash
git clone https://github.com/Shoaibashk/SerialLink.git
cd SerialLink
make install
```

Or directly with Go:

```bash
go install github.com/Shoaibashk/SerialLink@latest
```

### Build locally

```bash
make build
```

This creates a `seriallink` binary in the current directory.

## Usage

### Version

Display the version:

```bash
seriallink version
```

### Serve

Start the serial port agent:

```bash
seriallink serve --port /dev/ttyUSB0 --baud 115200
```

Short flags:

```bash
seriallink serve -p COM3 -b 9600
```

With verbose output:

```bash
seriallink --verbose serve --port /dev/ttyUSB0
```

### Configuration

SerialLink supports configuration via:

1. **Command-line flags** (highest priority)
   ```bash
   seriallink serve --port /dev/ttyUSB0 --baud 115200
   ```

2. **Environment variables** (with `SERIALLINK_` prefix)
   ```bash
   export SERIALLINK_PORT=/dev/ttyUSB0
   export SERIALLINK_BAUD=115200
   seriallink serve
   ```

3. **Config file** (lowest priority)
   ```bash
   seriallink --config ~/.seriallink/config.yaml serve
   ```

Example config file (`~/.seriallink/config.yaml`):

```yaml
port: /dev/ttyUSB0
baud: 115200
verbose: false
```

## Development

### Setup

```bash
# Install development tools
make install-tools

# Run all checks (format, vet, lint, test)
make ci
```

### Available Targets

- `make build` - Build the binary
- `make test` - Run unit tests
- `make coverage` - Generate coverage report (coverage.html)
- `make lint` - Run golangci-lint
- `make fmt` - Format code with gofmt
- `make vet` - Run go vet
- `make ci` - Run all CI checks
- `make clean` - Remove built binaries and reports

### Running Tests

```bash
# Run all tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test ./cmd -run TestServeCommand -v
```

### Linting

```bash
# Run linter
golangci-lint run ./...

# Available linters (configured in .golangci.yml):
# - govet: Go's static analyzer
# - staticcheck: Static analysis for Go
# - gofmt: Code formatting
# - errcheck: Unchecked errors
# - ineffassign: Ineffective assignments
# - gocyclo: Cyclomatic complexity (threshold: 15)
# - unused: Unused code
# - typecheck: Type checking
```

## Project Structure

```
SerialLink/
├── cmd/                        # CLI commands
│   ├── root.go                # Root command with Viper config
│   ├── version.go             # Version subcommand
│   ├── serve.go               # Serve subcommand
│   ├── root_test.go           # Tests for root command
│   └── serve_test.go          # Tests for serve command
├── internal/
│   └── serial/
│       ├── serial.go          # Serial interface and implementation
│       └── errors.go          # Custom errors
├── .github/
│   └── workflows/
│       └── ci.yml             # GitHub Actions CI workflow
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── Makefile                   # Build automation
├── .golangci.yml              # Linter configuration
└── README.md                  # This file
```

## Code Structure

### Commands

Commands are organized under the `cmd/` package:
- **root.go**: Root command with Viper configuration initialization
- **version.go**: Version subcommand
- **serve.go**: Serve subcommand with serial port parameters

### Testing

Comprehensive tests included:
- **root_test.go**: Tests for `Execute()`, help flag, version command, and context handling
- **serve_test.go**: Table-driven tests for serve command with various flags and configurations

### Configuration

- Viper integration with automatic environment variable binding (`SERIALLINK_` prefix)
- Support for YAML config files in `~/.seriallink/config.yaml`
- Command-line flag precedence over config file values

## Contributing

Contributions are welcome! Please ensure:

1. Code passes `make ci` checks
2. Tests are included for new features
3. Commit messages are clear and descriptive

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details. 
