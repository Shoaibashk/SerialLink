package serial

import "io"

// Serial defines the interface for serial port operations
type Serial interface {
	Open() error
	Close() error
	Write(data []byte) (int, error)
	Read(p []byte) (int, error)
}

// DefaultSerial is the default implementation of the Serial interface
type DefaultSerial struct {
	port     string
	baudRate int
	conn     io.ReadWriteCloser
}

// NewDefaultSerial creates a new DefaultSerial instance
func NewDefaultSerial(port string, baudRate int) *DefaultSerial {
	return &DefaultSerial{
		port:     port,
		baudRate: baudRate,
	}
}

// Open opens the serial port connection
func (s *DefaultSerial) Open() error {
	// TODO: Implement actual serial port opening
	// This would use go.bug.st/serial or similar library
	return nil
}

// Close closes the serial port connection
func (s *DefaultSerial) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Write writes data to the serial port
func (s *DefaultSerial) Write(data []byte) (int, error) {
	if s.conn == nil {
		return 0, ErrNotOpen
	}
	return s.conn.Write(data)
}

// Read reads data from the serial port
func (s *DefaultSerial) Read(p []byte) (int, error) {
	if s.conn == nil {
		return 0, ErrNotOpen
	}
	return s.conn.Read(p)
}
