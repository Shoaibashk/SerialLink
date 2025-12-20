package serial

import "errors"

var (
	// ErrNotOpen is returned when operations are attempted on a closed connection
	ErrNotOpen = errors.New("serial port is not open")
)
