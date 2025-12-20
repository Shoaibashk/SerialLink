// Package serial provides serial port management and communication functionality.
package serial

import (
	"regexp"
	"runtime"
	"sort"
	"sync"
	"time"

	"go.bug.st/serial/enumerator"
)

// PortType represents the type of serial port
type PortType int

const (
	PortTypeUnknown PortType = iota
	PortTypeUSB
	PortTypeNative
	PortTypeBluetooth
	PortTypeVirtual
)

// String returns the string representation of PortType
func (p PortType) String() string {
	switch p {
	case PortTypeUSB:
		return "USB"
	case PortTypeNative:
		return "Native"
	case PortTypeBluetooth:
		return "Bluetooth"
	case PortTypeVirtual:
		return "Virtual"
	default:
		return "Unknown"
	}
}

// PortInfo contains information about a serial port
type PortInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	HardwareID   string   `json:"hardware_id"`
	Manufacturer string   `json:"manufacturer"`
	Product      string   `json:"product"`
	SerialNumber string   `json:"serial_number"`
	VID          string   `json:"vid"`
	PID          string   `json:"pid"`
	PortType     PortType `json:"port_type"`
	IsOpen       bool     `json:"is_open"`
	LockedBy     string   `json:"locked_by"`
}

// Scanner handles serial port discovery and enumeration
type Scanner struct {
	mu              sync.RWMutex
	excludePatterns []*regexp.Regexp
	cachedPorts     []PortInfo
	manager         *Manager
}

// NewScanner creates a new port scanner
func NewScanner(excludePatterns []string, manager *Manager) (*Scanner, error) {
	s := &Scanner{
		manager: manager,
	}

	for _, pattern := range excludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		s.excludePatterns = append(s.excludePatterns, re)
	}

	return s, nil
}

// Scan discovers all available serial ports
func (s *Scanner) Scan() ([]PortInfo, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	var result []PortInfo

	for _, port := range ports {
		// Check if port should be excluded
		if s.isExcluded(port.Name) {
			continue
		}

		info := PortInfo{
			Name:         port.Name,
			Product:      port.Product,
			SerialNumber: port.SerialNumber,
			VID:          port.VID,
			PID:          port.PID,
			PortType:     s.detectPortType(port),
		}

		// Build hardware ID
		if port.VID != "" && port.PID != "" {
			info.HardwareID = "USB\\VID_" + port.VID + "&PID_" + port.PID
		}

		// Set description based on available info
		info.Description = s.buildDescription(port)

		// Check if port is currently open/locked
		if s.manager != nil {
			if session := s.manager.GetSession(port.Name); session != nil {
				info.IsOpen = true
				info.LockedBy = session.ClientID
			}
		}

		result = append(result, info)
	}

	// Sort ports by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	// Cache the results
	s.mu.Lock()
	s.cachedPorts = result
	s.mu.Unlock()

	return result, nil
}

// GetCached returns the last cached port list
func (s *Scanner) GetCached() []PortInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent data races
	if s.cachedPorts == nil {
		return nil
	}
	result := make([]PortInfo, len(s.cachedPorts))
	copy(result, s.cachedPorts)
	return result
}

// GetPort returns information about a specific port
func (s *Scanner) GetPort(name string) (*PortInfo, error) {
	ports, err := s.Scan()
	if err != nil {
		return nil, err
	}

	for _, port := range ports {
		if port.Name == name {
			return &port, nil
		}
	}

	return nil, ErrPortNotFound
}

// isExcluded checks if a port should be excluded based on patterns
func (s *Scanner) isExcluded(name string) bool {
	for _, pattern := range s.excludePatterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

// detectPortType determines the type of port
func (s *Scanner) detectPortType(port *enumerator.PortDetails) PortType {
	if port.IsUSB {
		return PortTypeUSB
	}

	// Check for Bluetooth ports
	switch runtime.GOOS {
	case "windows":
		// Windows Bluetooth COM ports often have specific names
		if matched, _ := regexp.MatchString(`(?i)bluetooth|bth`, port.Name); matched {
			return PortTypeBluetooth
		}
	case "linux":
		if matched, _ := regexp.MatchString(`/dev/rfcomm`, port.Name); matched {
			return PortTypeBluetooth
		}
	case "darwin":
		if matched, _ := regexp.MatchString(`/dev/.*Bluetooth`, port.Name); matched {
			return PortTypeBluetooth
		}
	}

	// Check for virtual/pseudo terminals
	if runtime.GOOS == "linux" {
		if matched, _ := regexp.MatchString(`/dev/pts/|/dev/pty`, port.Name); matched {
			return PortTypeVirtual
		}
	}

	return PortTypeNative
}

// buildDescription creates a human-readable description for the port
func (s *Scanner) buildDescription(port *enumerator.PortDetails) string {
	if port.Product != "" {
		return port.Product
	}
	if port.IsUSB {
		return "USB Serial Device"
	}
	return "Serial Port"
}

// PortChangeCallback is called when ports change
type PortChangeCallback func(added, removed []PortInfo, current []PortInfo)

// WatchPorts starts watching for port changes and calls the callback when ports change
func (s *Scanner) WatchPorts(intervalSeconds int, callback PortChangeCallback) chan struct{} {
	stop := make(chan struct{})

	if intervalSeconds <= 0 {
		intervalSeconds = 5 // Default 5 seconds
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		lastPorts := make(map[string]PortInfo)

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ports, err := s.Scan()
				if err != nil {
					continue
				}

				currentPorts := make(map[string]PortInfo)
				for _, p := range ports {
					currentPorts[p.Name] = p
				}

				// Find added ports
				var added []PortInfo
				for name, port := range currentPorts {
					if _, exists := lastPorts[name]; !exists {
						added = append(added, port)
					}
				}

				// Find removed ports
				var removed []PortInfo
				for name, port := range lastPorts {
					if _, exists := currentPorts[name]; !exists {
						removed = append(removed, port)
					}
				}

				// Notify if there are changes
				if len(added) > 0 || len(removed) > 0 {
					callback(added, removed, ports)
				}

				lastPorts = currentPorts
			}
		}
	}()

	return stop
}

// StopWatch stops watching for port changes
func (s *Scanner) StopWatch(stopChan chan struct{}) {
	if stopChan != nil {
		close(stopChan)
	}
}
