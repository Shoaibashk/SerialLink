// Package serial provides serial port management and communication functionality for SerialLink.
package serial

import "errors"

// Common errors for serial port operations
var (
	// ErrNotOpen is returned when operations are attempted on a closed connection
	ErrNotOpen = errors.New("serial port is not open")

	// ErrPortNotFound is returned when a port cannot be found
	ErrPortNotFound = errors.New("port not found")

	// ErrPortAlreadyOpen is returned when trying to open an already open port
	ErrPortAlreadyOpen = errors.New("port is already open")

	// ErrPortNotOpen is returned when operations are attempted on a closed port
	ErrPortNotOpen = errors.New("port is not open")

	// ErrPortLocked is returned when port is locked by another client
	ErrPortLocked = errors.New("port is locked by another client")

	// ErrInvalidSession is returned when session ID doesn't match
	ErrInvalidSession = errors.New("invalid session ID")

	// ErrInvalidConfig is returned when port configuration is invalid
	ErrInvalidConfig = errors.New("invalid port configuration")

	// ErrWriteTimeout is returned when write operation times out
	ErrWriteTimeout = errors.New("write timeout")

	// ErrReadTimeout is returned when read operation times out
	ErrReadTimeout = errors.New("read timeout")

	// ErrPortClosed is returned when port has been closed during operation
	ErrPortClosed = errors.New("port has been closed")
)
