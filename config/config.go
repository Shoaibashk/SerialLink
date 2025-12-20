/*
Copyright 2024 SerialLink Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package config provides configuration loading and management for SerialLink agent.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Shoaibashk/SerialLink/internal/serial"
	"github.com/spf13/viper"
)

// Config represents the complete agent configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server" yaml:"server"`
	TLS     TLSConfig     `mapstructure:"tls" yaml:"tls"`
	Serial  SerialConfig  `mapstructure:"serial" yaml:"serial"`
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`
	Service ServiceConfig `mapstructure:"service" yaml:"service"`
	Metrics MetricsConfig `mapstructure:"metrics" yaml:"metrics"`
}

// ServerConfig holds server-related settings
type ServerConfig struct {
	GRPCAddress       string `mapstructure:"grpc_address" yaml:"grpc_address"`
	WebSocketAddress  string `mapstructure:"websocket_address" yaml:"websocket_address"`
	WebSocketEnabled  bool   `mapstructure:"websocket_enabled" yaml:"websocket_enabled"`
	MaxConnections    int    `mapstructure:"max_connections" yaml:"max_connections"`
	ConnectionTimeout int    `mapstructure:"connection_timeout" yaml:"connection_timeout"`
}

// TLSConfig holds TLS/SSL settings
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
	KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
	CAFile   string `mapstructure:"ca_file" yaml:"ca_file"`
}

// SerialConfig holds serial port settings
type SerialConfig struct {
	Defaults          SerialDefaults `mapstructure:"defaults" yaml:"defaults"`
	ScanInterval      int            `mapstructure:"scan_interval" yaml:"scan_interval"`
	ExcludePatterns   []string       `mapstructure:"exclude_patterns" yaml:"exclude_patterns"`
	AllowSharedAccess bool           `mapstructure:"allow_shared_access" yaml:"allow_shared_access"`
}

// SerialDefaults holds default serial port parameters
type SerialDefaults struct {
	BaudRate       int    `mapstructure:"baud_rate" yaml:"baud_rate"`
	DataBits       int    `mapstructure:"data_bits" yaml:"data_bits"`
	StopBits       int    `mapstructure:"stop_bits" yaml:"stop_bits"`
	Parity         string `mapstructure:"parity" yaml:"parity"`
	FlowControl    string `mapstructure:"flow_control" yaml:"flow_control"`
	ReadTimeoutMs  int    `mapstructure:"read_timeout_ms" yaml:"read_timeout_ms"`
	WriteTimeoutMs int    `mapstructure:"write_timeout_ms" yaml:"write_timeout_ms"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level"`
	Format     string `mapstructure:"format" yaml:"format"`
	File       string `mapstructure:"file" yaml:"file"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age"`
	Compress   bool   `mapstructure:"compress" yaml:"compress"`
}

// ServiceConfig holds system service settings
type ServiceConfig struct {
	Name          string `mapstructure:"name" yaml:"name"`
	DisplayName   string `mapstructure:"display_name" yaml:"display_name"`
	Description   string `mapstructure:"description" yaml:"description"`
	AutoStart     bool   `mapstructure:"auto_start" yaml:"auto_start"`
	RestartPolicy string `mapstructure:"restart_policy" yaml:"restart_policy"`
	RestartDelay  int    `mapstructure:"restart_delay" yaml:"restart_delay"`
}

// MetricsConfig holds metrics/monitoring settings
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Address string `mapstructure:"address" yaml:"address"`
	Path    string `mapstructure:"path" yaml:"path"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCAddress:       "0.0.0.0:50051",
			WebSocketAddress:  "0.0.0.0:8080",
			WebSocketEnabled:  false,
			MaxConnections:    100,
			ConnectionTimeout: 30,
		},
		TLS: TLSConfig{
			Enabled: false,
		},
		Serial: SerialConfig{
			Defaults: SerialDefaults{
				BaudRate:       9600,
				DataBits:       8,
				StopBits:       1,
				Parity:         "none",
				FlowControl:    "none",
				ReadTimeoutMs:  1000,
				WriteTimeoutMs: 1000,
			},
			ScanInterval:      5,
			AllowSharedAccess: false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
		},
		Service: ServiceConfig{
			Name:          "seriallink",
			DisplayName:   "SerialLink Agent",
			Description:   "Cross-platform serial port background service",
			AutoStart:     true,
			RestartPolicy: "on-failure",
			RestartDelay:  5,
		},
		Metrics: MetricsConfig{
			Enabled: false,
			Address: "0.0.0.0:9090",
			Path:    "/metrics",
		},
	}
}

// ToPortConfig converts SerialDefaults into a concrete serial.PortConfig.
func (d SerialDefaults) ToPortConfig() (serial.PortConfig, error) {
	parity, err := serial.ParseParity(d.Parity)
	if err != nil {
		return serial.PortConfig{}, err
	}

	flowControl, err := serial.ParseFlowControl(d.FlowControl)
	if err != nil {
		return serial.PortConfig{}, err
	}

	stopBits, err := serial.ParseStopBits(d.StopBits)
	if err != nil {
		return serial.PortConfig{}, err
	}

	return serial.PortConfig{
		BaudRate:       d.BaudRate,
		DataBits:       d.DataBits,
		StopBits:       stopBits,
		Parity:         parity,
		FlowControl:    flowControl,
		ReadTimeoutMs:  d.ReadTimeoutMs,
		WriteTimeoutMs: d.WriteTimeoutMs,
	}, nil
}

// SetDefaults sets default values in viper
func SetDefaults() {
	defaults := DefaultConfig()

	// Server defaults
	viper.SetDefault("server.grpc_address", defaults.Server.GRPCAddress)
	viper.SetDefault("server.websocket_address", defaults.Server.WebSocketAddress)
	viper.SetDefault("server.websocket_enabled", defaults.Server.WebSocketEnabled)
	viper.SetDefault("server.max_connections", defaults.Server.MaxConnections)
	viper.SetDefault("server.connection_timeout", defaults.Server.ConnectionTimeout)

	// TLS defaults
	viper.SetDefault("tls.enabled", defaults.TLS.Enabled)

	// Serial defaults
	viper.SetDefault("serial.defaults.baud_rate", defaults.Serial.Defaults.BaudRate)
	viper.SetDefault("serial.defaults.data_bits", defaults.Serial.Defaults.DataBits)
	viper.SetDefault("serial.defaults.stop_bits", defaults.Serial.Defaults.StopBits)
	viper.SetDefault("serial.defaults.parity", defaults.Serial.Defaults.Parity)
	viper.SetDefault("serial.defaults.flow_control", defaults.Serial.Defaults.FlowControl)
	viper.SetDefault("serial.defaults.read_timeout_ms", defaults.Serial.Defaults.ReadTimeoutMs)
	viper.SetDefault("serial.defaults.write_timeout_ms", defaults.Serial.Defaults.WriteTimeoutMs)
	viper.SetDefault("serial.scan_interval", defaults.Serial.ScanInterval)
	viper.SetDefault("serial.allow_shared_access", defaults.Serial.AllowSharedAccess)

	// Logging defaults
	viper.SetDefault("logging.level", defaults.Logging.Level)
	viper.SetDefault("logging.format", defaults.Logging.Format)
	viper.SetDefault("logging.max_size", defaults.Logging.MaxSize)
	viper.SetDefault("logging.max_backups", defaults.Logging.MaxBackups)
	viper.SetDefault("logging.max_age", defaults.Logging.MaxAge)
	viper.SetDefault("logging.compress", defaults.Logging.Compress)

	// Service defaults
	viper.SetDefault("service.name", defaults.Service.Name)
	viper.SetDefault("service.display_name", defaults.Service.DisplayName)
	viper.SetDefault("service.description", defaults.Service.Description)
	viper.SetDefault("service.auto_start", defaults.Service.AutoStart)
	viper.SetDefault("service.restart_policy", defaults.Service.RestartPolicy)
	viper.SetDefault("service.restart_delay", defaults.Service.RestartDelay)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", defaults.Metrics.Enabled)
	viper.SetDefault("metrics.address", defaults.Metrics.Address)
	viper.SetDefault("metrics.path", defaults.Metrics.Path)
}

// Load reads configuration from viper and returns a Config struct
func Load() (*Config, error) {
	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadFromFile reads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return Load()
}

// LoadOrDefault loads configuration from file, or returns default if file doesn't exist
func LoadOrDefault(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	return LoadFromFile(path)
}

// Save writes configuration to a YAML file
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write using viper
	for key, value := range c.toMap() {
		viper.Set(key, value)
	}

	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// toMap converts config to a map for viper
func (c *Config) toMap() map[string]interface{} {
	return map[string]interface{}{
		"server":  c.Server,
		"tls":     c.TLS,
		"serial":  c.Serial,
		"logging": c.Logging,
		"service": c.Service,
		"metrics": c.Metrics,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.GRPCAddress == "" {
		return fmt.Errorf("grpc_address is required")
	}

	if c.Server.MaxConnections < 1 {
		return fmt.Errorf("max_connections must be at least 1")
	}

	if c.TLS.Enabled {
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
			return fmt.Errorf("TLS cert_file and key_file are required when TLS is enabled")
		}
	}

	if c.Serial.Defaults.BaudRate < 1 {
		return fmt.Errorf("baud_rate must be positive")
	}

	if c.Serial.Defaults.DataBits < 5 || c.Serial.Defaults.DataBits > 8 {
		return fmt.Errorf("data_bits must be between 5 and 8")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	if _, err := c.Serial.Defaults.ToPortConfig(); err != nil {
		return fmt.Errorf("invalid serial defaults: %w", err)
	}

	return nil
}

// DefaultConfigPath returns the default configuration file path for the current OS
func DefaultConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "SerialLink", "config.yaml")
	case "darwin":
		return "/usr/local/etc/seriallink/config.yaml"
	default:
		return "/etc/seriallink/config.yaml"
	}
}

// UserConfigPath returns the user-specific configuration file path
func UserConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, ".seriallink", "config.yaml")
	default:
		return filepath.Join(home, ".config", "seriallink", "config.yaml")
	}
}

// InitViper initializes viper with default configuration paths
func InitViper(configFile string) error {
	SetDefaults()

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Search in multiple locations
		home, _ := os.UserHomeDir()
		if home != "" {
			viper.AddConfigPath(filepath.Join(home, ".seriallink"))
			viper.AddConfigPath(filepath.Join(home, ".config", "seriallink"))
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/seriallink")

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variable support
	viper.SetEnvPrefix("SERIALLINK")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults
	}

	return nil
}
