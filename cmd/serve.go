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
	"syscall"

	"github.com/Shoaibashk/SerialLink/api"
	pb "github.com/Shoaibashk/SerialLink/api/proto"
	"github.com/Shoaibashk/SerialLink/config"
	"github.com/Shoaibashk/SerialLink/internal/serial"
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

	// Bind flags to viper
	_ = viper.BindPFlag("server.grpc_address", serveCmd.Flags().Lookup("address"))
	_ = viper.BindPFlag("tls.enabled", serveCmd.Flags().Lookup("tls"))
	_ = viper.BindPFlag("tls.cert_file", serveCmd.Flags().Lookup("cert"))
	_ = viper.BindPFlag("tls.key_file", serveCmd.Flags().Lookup("key"))
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with command line flags if provided
	if addr, _ := cmd.Flags().GetString("address"); addr != "" {
		cfg.Server.GRPCAddress = addr
	}

	if IsVerbose() {
		fmt.Printf("Starting SerialLink server v%s\n", Version)
		fmt.Printf("  Address: %s\n", cfg.Server.GRPCAddress)
		fmt.Printf("  TLS:     %v\n", cfg.TLS.Enabled)
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

	// Create gRPC server options
	var opts []grpc.ServerOption

	// Configure TLS if enabled
	if cfg.TLS.Enabled {
		tlsConfig, tlsErr := loadTLSConfig(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if tlsErr != nil {
			return fmt.Errorf("failed to load TLS config: %w", tlsErr)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	// Add server options for connection limits
	opts = append(opts,
		grpc.MaxConcurrentStreams(uint32(cfg.Server.MaxConnections)),
	)

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Create and register the serial service
	serialServer := api.NewSerialServer(manager, scanner, cfg)
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
		fmt.Printf("SerialLink gRPC server listening on %s\n", cfg.Server.GRPCAddress)
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		fmt.Println("\nShutting down gracefully...")
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
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
