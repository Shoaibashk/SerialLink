package config

import (
	"testing"

	"github.com/Shoaibashk/SerialLink/internal/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerialDefaultsToPortConfig(t *testing.T) {
	defaults := SerialDefaults{
		BaudRate:       115200,
		DataBits:       8,
		StopBits:       1,
		Parity:         "none",
		FlowControl:    "hardware",
		ReadTimeoutMs:  250,
		WriteTimeoutMs: 300,
	}

	cfg, err := defaults.ToPortConfig()
	require.NoError(t, err)

	assert.Equal(t, serial.PortConfig{
		BaudRate:       115200,
		DataBits:       8,
		StopBits:       serial.StopBits1,
		Parity:         serial.ParityNone,
		FlowControl:    serial.FlowControlHardware,
		ReadTimeoutMs:  250,
		WriteTimeoutMs: 300,
	}, cfg)
}

func TestSerialDefaultsToPortConfigInvalid(t *testing.T) {
	defaults := SerialDefaults{
		BaudRate:    9600,
		DataBits:    8,
		StopBits:    1,
		Parity:      "invalid",
		FlowControl: "none",
	}

	_, err := defaults.ToPortConfig()
	require.Error(t, err)
}

func TestDefaultConfigValidate(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.Validate())
}

func TestValidateUsesSerialDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Serial.Defaults.FlowControl = "broken"

	err := cfg.Validate()
	require.Error(t, err)
}
