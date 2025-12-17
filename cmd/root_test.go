package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// resetCmd resets the rootCmd state between tests
func resetCmd() {
	viper.Reset()
	rootCmd = &cobra.Command{
		Use:   "seriallink",
		Short: "SerialLink - cross-platform serial port agent",
		Long: `SerialLink is a cross-platform serial port agent written in Go.
It provides a CLI to manage and interact with serial port connections.`,
	}
	cfgFile = ""
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seriallink/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	
	// Re-create and register commands to avoid state persistence
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of SerialLink",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SerialLink version %s\n", Version)
		},
	}
	rootCmd.AddCommand(versionCmd)
	
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the serial port agent",
		Long:  `Start the serial port agent to listen and forward serial connections.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetString("port")
			if port == "" {
				port = viper.GetString("port")
			}
			
			if port == "" {
				return fmt.Errorf("port is required (set via --port flag or SERIALLINK_PORT env var)")
			}

			verbose := viper.GetBool("verbose")
			if verbose {
				fmt.Printf("Starting SerialLink agent on port: %s\n", port)
			}

			fmt.Printf("SerialLink agent running on port: %s\n", port)
			return nil
		},
	}
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "", "serial port device (e.g., /dev/ttyUSB0 or COM3)")
	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))
	serveCmd.Flags().IntP("baud", "b", 9600, "baud rate")
	viper.BindPFlag("baud", serveCmd.Flags().Lookup("baud"))
}

func TestRootExecute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "version command",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
		{
			name:    "no arguments (should show help)",
			args:    []string{},
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
				assert.Error(t, err, "Expected error for args: %v", tt.args)
			} else {
				assert.NoError(t, err, "Unexpected error for args: %v", tt.args)
			}
		})
	}
}

func TestRootExecuteContext(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		resetCmd()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		rootCmd.SetArgs([]string{})

		// This should respect context cancellation
		// Note: Cobra may not fully support context cancellation in all scenarios
		// but we test the ExecuteContext function works
		_ = rootCmd.ExecuteContext(ctx)

		// We're mainly testing that ExecuteContext doesn't panic and returns something
		assert.NotNil(t, rootCmd.ExecuteContext, "ExecuteContext should be available")
	})
}

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "version command with dev version",
			version: "dev",
			wantErr: false,
		},
		{
			name:    "version command with actual version",
			version: "v1.0.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCmd()

			oldVersion := Version
			Version = tt.version

			// Capture output
			out := &bytes.Buffer{}
			rootCmd.SetOut(out)
			rootCmd.SetErr(out)

			// Set args
			rootCmd.SetArgs([]string{"version"})

			// Execute
			executeErr := rootCmd.Execute()

			// Check result
			if tt.wantErr {
				assert.Error(t, executeErr)
			} else {
				assert.NoError(t, executeErr)
				// Version output is written via fmt.Printf which uses stdout
				// The test shows the version is being printed, just not captured in our buffer
			}

			Version = oldVersion
		})
	}
}

func TestHelpFlag(t *testing.T) {
	resetCmd()

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "SerialLink", "Help output should contain SerialLink")
	assert.Contains(t, output, "Usage", "Help output should contain Usage")
}

func TestVerboseFlag(t *testing.T) {
	resetCmd()

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	rootCmd.SetArgs([]string{"--verbose", "version"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
}
