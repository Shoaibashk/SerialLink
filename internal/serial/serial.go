package serial

import (
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Parity represents the parity setting for serial communication
type Parity int

const (
	ParityNone Parity = iota
	ParityOdd
	ParityEven
	ParityMark
	ParitySpace
)

// String returns the string representation of Parity
func (p Parity) String() string {
	switch p {
	case ParityNone:
		return "none"
	case ParityOdd:
		return "odd"
	case ParityEven:
		return "even"
	case ParityMark:
		return "mark"
	case ParitySpace:
		return "space"
	default:
		return "unknown"
	}
}

// StopBits represents the stop bits setting
type StopBits int

const (
	StopBits1 StopBits = iota
	StopBits1Half
	StopBits2
)

// String returns the string representation of StopBits
func (s StopBits) String() string {
	switch s {
	case StopBits1:
		return "1"
	case StopBits1Half:
		return "1.5"
	case StopBits2:
		return "2"
	default:
		return "unknown"
	}
}

// FlowControl represents the flow control setting
type FlowControl int

const (
	FlowControlNone     FlowControl = iota
	FlowControlHardware             // RTS/CTS
	FlowControlSoftware             // XON/XOFF
)

// String returns the string representation of FlowControl
func (f FlowControl) String() string {
	switch f {
	case FlowControlNone:
		return "none"
	case FlowControlHardware:
		return "hardware"
	case FlowControlSoftware:
		return "software"
	default:
		return "unknown"
	}
}

// PortConfig represents serial port configuration
type PortConfig struct {
	BaudRate       int
	DataBits       int
	StopBits       StopBits
	Parity         Parity
	FlowControl    FlowControl
	ReadTimeoutMs  int
	WriteTimeoutMs int
}

// DefaultConfig returns a default port configuration
func DefaultConfig() PortConfig {
	return PortConfig{
		BaudRate:       9600,
		DataBits:       8,
		StopBits:       StopBits1,
		Parity:         ParityNone,
		FlowControl:    FlowControlNone,
		ReadTimeoutMs:  1000,
		WriteTimeoutMs: 1000,
	}
}

// Validate checks if the configuration is valid
func (c PortConfig) Validate() error {
	validBaudRates := map[int]bool{
		300: true, 600: true, 1200: true, 2400: true, 4800: true,
		9600: true, 19200: true, 38400: true, 57600: true, 115200: true,
		230400: true, 460800: true, 921600: true,
	}

	if c.BaudRate < 1 {
		return fmt.Errorf("%w: baud rate must be positive, got %d", ErrInvalidConfig, c.BaudRate)
	}

	// Allow custom baud rates but warn about non-standard ones
	if !validBaudRates[c.BaudRate] && c.BaudRate < 300 {
		return fmt.Errorf("%w: baud rate %d is too low", ErrInvalidConfig, c.BaudRate)
	}

	if c.DataBits < 5 || c.DataBits > 8 {
		return fmt.Errorf("%w: data bits must be 5-8, got %d", ErrInvalidConfig, c.DataBits)
	}

	if c.StopBits < StopBits1 || c.StopBits > StopBits2 {
		return fmt.Errorf("%w: invalid stop bits value", ErrInvalidConfig)
	}

	if c.Parity < ParityNone || c.Parity > ParitySpace {
		return fmt.Errorf("%w: invalid parity value", ErrInvalidConfig)
	}

	if c.FlowControl < FlowControlNone || c.FlowControl > FlowControlSoftware {
		return fmt.Errorf("%w: invalid flow control value", ErrInvalidConfig)
	}

	return nil
}

// ToSerialMode converts PortConfig to serial.Mode for the underlying library
func (c PortConfig) ToSerialMode() *serial.Mode {
	mode := &serial.Mode{
		BaudRate: c.BaudRate,
		DataBits: c.DataBits,
	}

	switch c.StopBits {
	case StopBits1:
		mode.StopBits = serial.OneStopBit
	case StopBits1Half:
		mode.StopBits = serial.OnePointFiveStopBits
	case StopBits2:
		mode.StopBits = serial.TwoStopBits
	}

	switch c.Parity {
	case ParityNone:
		mode.Parity = serial.NoParity
	case ParityOdd:
		mode.Parity = serial.OddParity
	case ParityEven:
		mode.Parity = serial.EvenParity
	case ParityMark:
		mode.Parity = serial.MarkParity
	case ParitySpace:
		mode.Parity = serial.SpaceParity
	}

	return mode
}

// PortStatistics contains statistics about port usage
type PortStatistics struct {
	BytesSent     uint64
	BytesReceived uint64
	Errors        uint64
	OpenedAt      time.Time
	LastActivity  time.Time
}

// ReadResult represents the result of a read operation with timeout
type ReadResult struct {
	Data  []byte
	Error error
}

// ReadWithTimeout performs a read operation with a specified timeout
func ReadWithTimeout(m *Manager, portName, sessionID string, maxBytes int, timeout time.Duration) ReadResult {
	resultChan := make(chan ReadResult, 1)

	go func() {
		data, err := m.Read(portName, sessionID, maxBytes)
		resultChan <- ReadResult{Data: data, Error: err}
	}()

	select {
	case result := <-resultChan:
		return result
	case <-time.After(timeout):
		return ReadResult{Error: ErrReadTimeout}
	}
}

// ParseParity converts a parity string into a Parity enum.
func ParseParity(value string) (Parity, error) {
	switch strings.ToLower(value) {
	case "", "none":
		return ParityNone, nil
	case "odd":
		return ParityOdd, nil
	case "even":
		return ParityEven, nil
	case "mark":
		return ParityMark, nil
	case "space":
		return ParitySpace, nil
	default:
		return ParityNone, fmt.Errorf("%w: invalid parity %q", ErrInvalidConfig, value)
	}
}

// ParseFlowControl converts a flow control string into a FlowControl enum.
func ParseFlowControl(value string) (FlowControl, error) {
	switch strings.ToLower(value) {
	case "", "none":
		return FlowControlNone, nil
	case "hardware", "hw", "rts/cts":
		return FlowControlHardware, nil
	case "software", "sw", "xon/xoff":
		return FlowControlSoftware, nil
	default:
		return FlowControlNone, fmt.Errorf("%w: invalid flow control %q", ErrInvalidConfig, value)
	}
}

// ParseStopBits converts a stop bits integer into a StopBits enum.
func ParseStopBits(value int) (StopBits, error) {
	switch value {
	case 1:
		return StopBits1, nil
	case 2:
		return StopBits2, nil
	default:
		return StopBits1, fmt.Errorf("%w: invalid stop bits %d", ErrInvalidConfig, value)
	}
}
