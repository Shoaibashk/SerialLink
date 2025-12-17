package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestServeCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "serve without port flag",
			args:    []string{"serve"},
			wantErr: true,
			errMsg:  "port is required",
		},
		{
			name:    "serve with port flag",
			args:    []string{"serve", "--port", "/dev/ttyUSB0"},
			wantErr: false,
		},
		{
			name:    "serve with short port flag",
			args:    []string{"serve", "-p", "COM3"},
			wantErr: false,
		},
		{
			name:    "serve with baud rate",
			args:    []string{"serve", "-p", "/dev/ttyUSB0", "--baud", "115200"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCmd()

			// Capture output
			out := &bytes.Buffer{}
			rootCmd.SetOut(out)
			rootCmd.SetErr(out)

			// Set args
			rootCmd.SetArgs(tt.args)

			// Execute
			err := rootCmd.Execute()

			// Check result
			if tt.wantErr {
				assert.Error(t, err, "Expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error for args: %v", tt.args)
			}

			viper.Reset()
		})
	}
}

func TestServeCommandWithVerbose(t *testing.T) {
	t.Run("serve with verbose flag", func(t *testing.T) {
		resetCmd()

		out := &bytes.Buffer{}
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		rootCmd.SetArgs([]string{"--verbose", "serve", "--port", "/dev/ttyUSB0"})
		err := rootCmd.Execute()

		assert.NoError(t, err)

		viper.Reset()
	})
}

func TestServeCommandBaudRate(t *testing.T) {
	t.Run("serve with valid baud rate", func(t *testing.T) {
		resetCmd()

		out := &bytes.Buffer{}
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		rootCmd.SetArgs([]string{"serve", "-p", "/dev/ttyUSB0", "-b", "115200"})
		err := rootCmd.Execute()

		assert.NoError(t, err)

		viper.Reset()
	})

	t.Run("serve with default baud rate", func(t *testing.T) {
		resetCmd()

		out := &bytes.Buffer{}
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		rootCmd.SetArgs([]string{"serve", "-p", "/dev/ttyUSB0"})
		err := rootCmd.Execute()

		assert.NoError(t, err)

		viper.Reset()
	})
}

// TestTableServeCommand demonstrates table-driven tests for serve command variations
func TestTableServeCommand(t *testing.T) {
	testCases := []struct {
		description     string
		args            []string
		expectError     bool
		expectErrSubstr string
	}{
		{"No port specified", []string{"serve"}, true, "port is required"},
		{"Port via --port flag", []string{"serve", "--port", "/dev/ttyUSB0"}, false, ""},
		{"Port via -p short flag", []string{"serve", "-p", "COM3"}, false, ""},
		{"With custom baud", []string{"serve", "--port", "/dev/ttyUSB0", "--baud", "115200"}, false, ""},
		{"With short flags", []string{"serve", "-p", "/dev/ttyUSB0", "-b", "38400"}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resetCmd()

			out := &bytes.Buffer{}
			rootCmd.SetOut(out)
			rootCmd.SetErr(out)

			rootCmd.SetArgs(tc.args)
			err := rootCmd.Execute()

			if tc.expectError {
				assert.Error(t, err, fmt.Sprintf("Expected error for: %v", tc.args))
				if tc.expectErrSubstr != "" {
					assert.Contains(t, err.Error(), tc.expectErrSubstr)
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("Unexpected error for: %v", tc.args))
			}

			viper.Reset()
		})
	}
}
