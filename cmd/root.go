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

// Package cmd provides the CLI commands for SerialLink using Cobra.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Shoaibashk/SerialLink/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is the application version (set at build time)
	Version = "dev"

	// Commit is the git commit (set at build time)
	Commit = "none"

	// BuildDate is the build date (set at build time)
	BuildDate = "unknown"

	// cfgFile is the path to the config file
	cfgFile string

	// verbose enables verbose output
	verbose bool

	// address is the gRPC service address
	address string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "seriallink",
	Short: "SerialLink - Cross-platform serial port agent",
	Long: `SerialLink is a cross-platform serial port background service that runs on 
Windows, Linux, and Raspberry Pi. It manages all serial hardware and exposes 
a public gRPC API for any client - Python, C#, Node.js, Web, Mobile, or CLI.

Features:
  • Auto-detect serial ports (USB, native, Bluetooth, virtual)
  • Open/Close ports with exclusive locking
  • Read/Write with timeout support
  • Real-time bidirectional streaming
  • Hot-plug detection
  • TLS encryption support
  • Cross-language gRPC API

Example usage:
  seriallink serve                    Start the gRPC server
  seriallink scan                     List available serial ports
  seriallink version                  Show version information`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext executes the root command with a context
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: $HOME/.seriallink/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&address, "address", "localhost:50051", "gRPC service address (can also be set via SERIALLINK_ADDRESS env var)")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))

	// Bind environment variables
	_ = viper.BindEnv("address", "SERIALLINK_ADDRESS")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if err := config.InitViper(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if verbose {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// GetConfig returns the loaded configuration
func GetConfig() (*config.Config, error) {
	return config.Load()
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose || viper.GetBool("verbose")
}

// GetAddress returns the gRPC service address
func GetAddress() string {
	addr := viper.GetString("address")
	if addr == "" {
		addr = "localhost:50051"
	}
	return addr
}
