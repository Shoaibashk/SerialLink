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

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pb "github.com/Shoaibashk/SerialLink-Proto/gen/go/seriallink/v1"
	"github.com/Shoaibashk/SerialLink/api"
	"github.com/Shoaibashk/SerialLink/config"
	"github.com/Shoaibashk/SerialLink/internal/serial"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SerialLink gRPC server",
	Long: `Start the SerialLink gRPC server to manage serial port connections.

The server listens for gRPC requests and provides:
  • Port discovery and enumeration
  • Port open/close with exclusive locking
  • Read/Write operations
  • Bidirectional streaming
  • Port configuration

Example:
  seriallink serve                          # Start with default settings
  seriallink serve --address 0.0.0.0:50052  # Custom address
  seriallink serve --tls                    # Enable TLS`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server flags
	serveCmd.Flags().StringP("address", "a", "", "gRPC server address (default: 0.0.0.0:50051)")
	serveCmd.Flags().Bool("tls", false, "enable TLS")
	serveCmd.Flags().String("cert", "", "TLS certificate file")
	serveCmd.Flags().String("key", "", "TLS key file")
	serveCmd.Flags().Bool("reflection", true, "enable gRPC reflection")

	// Bind flags to viper with error logging
	if err := viper.BindPFlag("server.grpc_address", serveCmd.Flags().Lookup("address")); err != nil {
		log.Warn("failed to bind address flag", "error", err)
	}
	if err := viper.BindPFlag("tls.enabled", serveCmd.Flags().Lookup("tls")); err != nil {
		log.Warn("failed to bind tls flag", "error", err)
	}
	if err := viper.BindPFlag("tls.cert_file", serveCmd.Flags().Lookup("cert")); err != nil {
		log.Warn("failed to bind cert flag", "error", err)
	}
	if err := viper.BindPFlag("tls.key_file", serveCmd.Flags().Lookup("key")); err != nil {
		log.Warn("failed to bind key flag", "error", err)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger based on config
	logger := initLogger(cfg)

	// Override with command line flags if provided
	if addr, _ := cmd.Flags().GetString("address"); addr != "" {
		cfg.Server.GRPCAddress = addr
	}

	logger.Info("Starting SerialLink server",
		"version", Version,
		"address", cfg.Server.GRPCAddress,
		"tls", cfg.TLS.Enabled)

	// Validate TLS certificates if TLS is enabled
	if cfg.TLS.Enabled {
		if err := validateTLSConfig(cfg.TLS, logger); err != nil {
			return fmt.Errorf("TLS validation failed: %w", err)
		}
	}

	// Create serial manager with default config
	defaultSerialConfig, err := cfg.Serial.Defaults.ToPortConfig()
	if err != nil {
		return fmt.Errorf("failed to build serial defaults: %w", err)
	}

	manager := serial.NewManager(cfg.Serial.AllowSharedAccess, defaultSerialConfig)
	defer manager.CloseAll()

	// Create scanner
	scanner, err := serial.NewScanner(cfg.Serial.ExcludePatterns, manager)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	// Create gRPC server options with logging interceptors
	var opts []grpc.ServerOption
	opts = append(opts,
		grpc.UnaryInterceptor(api.UnaryLoggingInterceptor(logger)),
		grpc.StreamInterceptor(api.StreamLoggingInterceptor(logger)),
	)

	// Configure TLS if enabled
	if cfg.TLS.Enabled {
		tlsConfig, tlsErr := loadTLSConfig(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if tlsErr != nil {
			return fmt.Errorf("failed to load TLS config: %w", tlsErr)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		logger.Info("TLS enabled", "cert", cfg.TLS.CertFile)
	}

	// Add server options for connection limits
	opts = append(opts,
		grpc.MaxConcurrentStreams(uint32(cfg.Server.MaxConnections)),
	)

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Create and register the serial service
	serialServer := api.NewSerialServer(manager, scanner, cfg, logger)
	pb.RegisterSerialServiceServer(grpcServer, serialServer)

	// Enable reflection for debugging
	if enabled, _ := cmd.Flags().GetBool("reflection"); enabled {
		reflection.Register(grpcServer)
	}

	// Start listening
	listener, err := net.Listen("tcp", cfg.Server.GRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Server.GRPCAddress, err)
	}

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("SerialLink gRPC server listening", "address", cfg.Server.GRPCAddress)
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down gracefully...")
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// initLogger creates and configures a charmbracelet logger based on config
func initLogger(cfg *config.Config) *log.Logger {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		ReportCaller:    true,
	})

	// Set log level from config
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	return logger
}

// validateTLSConfig validates that TLS certificate files exist and are readable
func validateTLSConfig(tlsCfg config.TLSConfig, logger *log.Logger) error {
	// Validate certificate file
	if tlsCfg.CertFile != "" {
		if _, err := os.Stat(tlsCfg.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS certificate file not found: %s", tlsCfg.CertFile)
		} else if err != nil {
			return fmt.Errorf("cannot access TLS certificate file: %w", err)
		}
		logger.Debug("TLS certificate file validated", "path", tlsCfg.CertFile)
	}

	// Validate key file
	if tlsCfg.KeyFile != "" {
		if _, err := os.Stat(tlsCfg.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS key file not found: %s", tlsCfg.KeyFile)
		} else if err != nil {
			return fmt.Errorf("cannot access TLS key file: %w", err)
		}
		logger.Debug("TLS key file validated", "path", tlsCfg.KeyFile)
	}

	// Validate CA file (optional)
	if tlsCfg.CAFile != "" {
		if _, err := os.Stat(tlsCfg.CAFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS CA file not found: %s", tlsCfg.CAFile)
		} else if err != nil {
			return fmt.Errorf("cannot access TLS CA file: %w", err)
		}
		logger.Debug("TLS CA file validated", "path", tlsCfg.CAFile)
	}

	return nil
}

func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
