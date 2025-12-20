package serial

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.bug.st/serial"
)

// Session represents an active serial port session
type Session struct {
	ID         string
	PortName   string
	ClientID   string
	Exclusive  bool
	Config     PortConfig
	Statistics PortStatistics
	port       serial.Port
	mu         sync.Mutex
	closed     atomic.Bool
	readers    []chan []byte
	readersMu  sync.RWMutex
}

// IsClosed returns whether the session has been closed
func (s *Session) IsClosed() bool {
	return s.closed.Load()
}

// Manager handles serial port sessions and operations
type Manager struct {
	mu                sync.RWMutex
	sessions          map[string]*Session // key: port name
	sessionsByID      map[string]*Session // key: session ID
	allowSharedAccess bool
	defaultConfig     PortConfig
}

// NewManager creates a new serial port manager
func NewManager(allowSharedAccess bool, defaultConfig PortConfig) *Manager {
	return &Manager{
		sessions:          make(map[string]*Session),
		sessionsByID:      make(map[string]*Session),
		allowSharedAccess: allowSharedAccess,
		defaultConfig:     defaultConfig,
	}
}

// OpenPort opens a serial port and creates a new session
func (m *Manager) OpenPort(portName string, config PortConfig, clientID string, exclusive bool) (*Session, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if port is already open
	if existingSession, exists := m.sessions[portName]; exists {
		if existingSession.Exclusive || exclusive || !m.allowSharedAccess {
			return nil, ErrPortLocked
		}
	}

	// Open the serial port
	port, err := serial.Open(portName, config.ToSerialMode())
	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %w", portName, err)
	}

	// Set read timeout
	if config.ReadTimeoutMs > 0 {
		if err := port.SetReadTimeout(time.Duration(config.ReadTimeoutMs) * time.Millisecond); err != nil {
			port.Close()
			return nil, fmt.Errorf("failed to set read timeout: %w", err)
		}
	}

	// Create session
	session := &Session{
		ID:        uuid.New().String(),
		PortName:  portName,
		ClientID:  clientID,
		Exclusive: exclusive,
		Config:    config,
		Statistics: PortStatistics{
			OpenedAt:     time.Now(),
			LastActivity: time.Now(),
		},
		port:    port,
		readers: make([]chan []byte, 0),
	}

	m.sessions[portName] = session
	m.sessionsByID[session.ID] = session

	return session, nil
}

// ClosePort closes a serial port session
func (m *Manager) ClosePort(portName string, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[portName]
	if !exists {
		return ErrPortNotOpen
	}

	if session.ID != sessionID {
		return ErrInvalidSession
	}

	return m.closeSessionLocked(session)
}

// closeSessionLocked closes a session (must be called with lock held)
func (m *Manager) closeSessionLocked(session *Session) error {
	session.closed.Store(true)

	// Close all reader channels
	session.readersMu.Lock()
	for _, ch := range session.readers {
		close(ch)
	}
	session.readers = nil
	session.readersMu.Unlock()

	// Close the port
	var err error
	if session.port != nil {
		err = session.port.Close()
	}

	delete(m.sessions, session.PortName)
	delete(m.sessionsByID, session.ID)

	return err
}

// GetSession returns the session for a port
func (m *Manager) GetSession(portName string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[portName]
}

// GetSessionByID returns a session by its ID
func (m *Manager) GetSessionByID(sessionID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionsByID[sessionID]
}

// ValidateSession checks if a session is valid
func (m *Manager) ValidateSession(portName string, sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[portName]
	if !exists {
		return nil, ErrPortNotOpen
	}

	if session.ID != sessionID {
		return nil, ErrInvalidSession
	}

	if session.closed.Load() {
		return nil, ErrPortClosed
	}

	return session, nil
}

// Write writes data to a port
func (m *Manager) Write(portName string, sessionID string, data []byte) (int, error) {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return 0, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	n, err := session.port.Write(data)
	if err != nil {
		atomic.AddUint64(&session.Statistics.Errors, 1)
		return n, fmt.Errorf("write failed: %w", err)
	}

	atomic.AddUint64(&session.Statistics.BytesSent, uint64(n))
	session.Statistics.LastActivity = time.Now()

	return n, nil
}

// Read reads data from a port
func (m *Manager) Read(portName string, sessionID string, maxBytes int) ([]byte, error) {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	buffer := make([]byte, maxBytes)
	n, err := session.port.Read(buffer)
	if err != nil {
		atomic.AddUint64(&session.Statistics.Errors, 1)
		return nil, fmt.Errorf("read failed: %w", err)
	}

	atomic.AddUint64(&session.Statistics.BytesReceived, uint64(n))
	session.Statistics.LastActivity = time.Now()

	// Broadcast to all subscribed readers
	if n > 0 {
		data := buffer[:n]
		session.readersMu.RLock()
		for _, ch := range session.readers {
			select {
			case ch <- data:
			default:
				// Channel full, skip
			}
		}
		session.readersMu.RUnlock()
	}

	return buffer[:n], nil
}

// Configure updates port configuration
func (m *Manager) Configure(portName string, sessionID string, config PortConfig) error {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return err
	}

	if err := config.Validate(); err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if err := session.port.SetMode(config.ToSerialMode()); err != nil {
		return fmt.Errorf("failed to configure port: %w", err)
	}

	if config.ReadTimeoutMs > 0 {
		if err := session.port.SetReadTimeout(time.Duration(config.ReadTimeoutMs) * time.Millisecond); err != nil {
			return fmt.Errorf("failed to set read timeout: %w", err)
		}
	}

	session.Config = config
	return nil
}

// GetStatus returns the status of a port
func (m *Manager) GetStatus(portName string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[portName]
	if !exists {
		return nil, ErrPortNotOpen
	}

	return session, nil
}

// ListOpenPorts returns all open port names
func (m *Manager) ListOpenPorts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ports := make([]string, 0, len(m.sessions))
	for portName := range m.sessions {
		ports = append(ports, portName)
	}
	return ports
}

// CloseAll closes all open ports
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		_ = m.closeSessionLocked(session)
	}
}

// SubscribeToReads creates a channel that receives data read from the port
func (m *Manager) SubscribeToReads(portName string, sessionID string) (<-chan []byte, error) {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte, 100)

	session.readersMu.Lock()
	session.readers = append(session.readers, ch)
	session.readersMu.Unlock()

	return ch, nil
}

// UnsubscribeFromReads removes a read subscription
func (m *Manager) UnsubscribeFromReads(portName string, sessionID string, ch <-chan []byte) error {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return err
	}

	session.readersMu.Lock()
	defer session.readersMu.Unlock()

	for i, sub := range session.readers {
		if sub == ch {
			close(sub)
			session.readers = append(session.readers[:i], session.readers[i+1:]...)
			return nil
		}
	}

	return nil
}

// Flush drains both input and output buffers
func (m *Manager) Flush(portName string, sessionID string) error {
	session, err := m.ValidateSession(portName, sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if err := session.port.ResetInputBuffer(); err != nil {
		return fmt.Errorf("failed to reset input buffer: %w", err)
	}

	if err := session.port.ResetOutputBuffer(); err != nil {
		return fmt.Errorf("failed to reset output buffer: %w", err)
	}

	return nil
}

// GetDefaultConfig returns the manager's default port configuration
func (m *Manager) GetDefaultConfig() PortConfig {
	return m.defaultConfig
}
