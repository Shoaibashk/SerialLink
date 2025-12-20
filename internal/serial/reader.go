package serial

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Reader provides continuous reading from a serial port with streaming support
type Reader struct {
	manager     *Manager
	portName    string
	sessionID   string
	bufferSize  int
	running     atomic.Bool
	stopChan    chan struct{}
	subscribers []chan DataEvent
	subMu       sync.RWMutex
}

// DataEvent represents a data read event
type DataEvent struct {
	Data      []byte
	Timestamp time.Time
	Sequence  uint32
	Error     error
}

// NewReader creates a new continuous reader for a port
func NewReader(manager *Manager, portName, sessionID string, bufferSize int) *Reader {
	if bufferSize <= 0 {
		bufferSize = 1024
	}

	return &Reader{
		manager:     manager,
		portName:    portName,
		sessionID:   sessionID,
		bufferSize:  bufferSize,
		stopChan:    make(chan struct{}),
		subscribers: make([]chan DataEvent, 0),
	}
}

// Start begins continuous reading from the port
func (r *Reader) Start(ctx context.Context) error {
	if r.running.Load() {
		return nil // Already running
	}

	// Validate session
	_, err := r.manager.ValidateSession(r.portName, r.sessionID)
	if err != nil {
		return err
	}

	r.running.Store(true)
	r.stopChan = make(chan struct{})

	go r.readLoop(ctx)

	return nil
}

// Stop stops the continuous reader
func (r *Reader) Stop() {
	if !r.running.Load() {
		return
	}

	r.running.Store(false)

	// Close stop channel (safe to close multiple times with this pattern)
	select {
	case <-r.stopChan:
		// Already closed
	default:
		close(r.stopChan)
	}

	// Close all subscriber channels
	r.subMu.Lock()
	for _, ch := range r.subscribers {
		close(ch)
	}
	r.subscribers = nil
	r.subMu.Unlock()
}

// Subscribe creates a new subscription to read events
func (r *Reader) Subscribe() <-chan DataEvent {
	ch := make(chan DataEvent, 100)

	r.subMu.Lock()
	r.subscribers = append(r.subscribers, ch)
	r.subMu.Unlock()

	return ch
}

// Unsubscribe removes a subscription
func (r *Reader) Unsubscribe(ch <-chan DataEvent) {
	r.subMu.Lock()
	defer r.subMu.Unlock()

	for i, sub := range r.subscribers {
		if sub == ch {
			close(sub)
			r.subscribers = append(r.subscribers[:i], r.subscribers[i+1:]...)
			return
		}
	}
}

// SubscriberCount returns the number of active subscribers
func (r *Reader) SubscriberCount() int {
	r.subMu.RLock()
	defer r.subMu.RUnlock()
	return len(r.subscribers)
}

// readLoop continuously reads from the port
func (r *Reader) readLoop(ctx context.Context) {
	var sequence uint32

	for r.running.Load() {
		select {
		case <-ctx.Done():
			r.Stop()
			return
		case <-r.stopChan:
			return
		default:
			data, err := r.manager.Read(r.portName, r.sessionID, r.bufferSize)

			// Skip if no data (timeout with no data is normal)
			if err == nil && len(data) == 0 {
				time.Sleep(1 * time.Millisecond) // Small sleep to prevent busy loop
				continue
			}

			event := DataEvent{
				Data:      data,
				Timestamp: time.Now(),
				Sequence:  atomic.AddUint32(&sequence, 1),
				Error:     err,
			}

			r.broadcast(event)

			if err != nil {
				// Check if it's a fatal error
				if err == ErrPortClosed || err == ErrInvalidSession {
					r.Stop()
					return
				}
				// Non-fatal errors - continue reading with small delay
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// broadcast sends an event to all subscribers
func (r *Reader) broadcast(event DataEvent) {
	r.subMu.RLock()
	defer r.subMu.RUnlock()

	for _, ch := range r.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, drop the event to prevent blocking
		}
	}
}

// IsRunning returns whether the reader is currently running
func (r *Reader) IsRunning() bool {
	return r.running.Load()
}

// WriteWithTimeout writes data with a specific timeout
func WriteWithTimeout(manager *Manager, portName, sessionID string, data []byte, timeout time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type writeResult struct {
		n   int
		err error
	}

	resultChan := make(chan writeResult, 1)

	go func() {
		n, err := manager.Write(portName, sessionID, data)
		resultChan <- writeResult{n: n, err: err}
	}()

	select {
	case result := <-resultChan:
		return result.n, result.err
	case <-ctx.Done():
		return 0, ErrWriteTimeout
	}
}

// LineReader reads complete lines from the port
type LineReader struct {
	reader    *Reader
	delimiter byte
	buffer    []byte
	maxLine   int
}

// NewLineReader creates a new line-based reader
func NewLineReader(reader *Reader, delimiter byte, maxLineSize int) *LineReader {
	if maxLineSize <= 0 {
		maxLineSize = 4096
	}

	return &LineReader{
		reader:    reader,
		delimiter: delimiter,
		buffer:    make([]byte, 0, maxLineSize),
		maxLine:   maxLineSize,
	}
}

// ReadLine reads a complete line from the subscription channel
func (lr *LineReader) ReadLine(dataChan <-chan DataEvent) ([]byte, error) {
	for {
		// Check buffer for existing line
		for i, b := range lr.buffer {
			if b == lr.delimiter {
				line := make([]byte, i)
				copy(line, lr.buffer[:i])
				lr.buffer = lr.buffer[i+1:]
				return line, nil
			}
		}

		// Wait for more data
		event, ok := <-dataChan
		if !ok {
			// Channel closed
			if len(lr.buffer) > 0 {
				line := lr.buffer
				lr.buffer = nil
				return line, nil
			}
			return nil, ErrPortClosed
		}

		if event.Error != nil {
			return nil, event.Error
		}

		// Append to buffer
		lr.buffer = append(lr.buffer, event.Data...)

		// Check for buffer overflow
		if len(lr.buffer) > lr.maxLine {
			// Return partial line and reset
			line := lr.buffer
			lr.buffer = nil
			return line, nil
		}
	}
}

// Reset clears the line buffer
func (lr *LineReader) Reset() {
	lr.buffer = lr.buffer[:0]
}
