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
	"fmt"
	"time"

	pb "github.com/Shoaibashk/SerialLink/api/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var readCmd = &cobra.Command{
	Use:   "read PORT [flags]",
	Short: "Read data from a serial port",
	Long: `Read data from an open serial port.

Example:
  seriallink read COM1                     # Read available data
  seriallink read COM1 --max-bytes 256     # Read up to 256 bytes
  seriallink read COM1 --timeout 5000      # Read with 5 second timeout`,
	Args: cobra.ExactArgs(1),
	RunE: runRead,
}

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.Flags().Uint32("max-bytes", 1024, "maximum bytes to read")
	readCmd.Flags().Uint32("timeout", 1000, "timeout in milliseconds")
	readCmd.Flags().String("session-id", "", "session ID")
	readCmd.Flags().String("format", "text", "output format (text, hex, json)")
}

func runRead(cmd *cobra.Command, args []string) error {
	portName := args[0]
	maxBytes, _ := cmd.Flags().GetUint32("max-bytes")
	timeout, _ := cmd.Flags().GetUint32("timeout")
	sessionID, _ := cmd.Flags().GetString("session-id")
	format, _ := cmd.Flags().GetString("format")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+2000)*time.Millisecond)
	defer cancel()

	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	resp, err := client.Read(ctx, &pb.ReadRequest{
		PortName:  portName,
		SessionId: sessionID,
		MaxBytes:  maxBytes,
		TimeoutMs: timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to read from port: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("read operation failed: %s", resp.Message)
	}

	if len(resp.Data) == 0 {
		if IsVerbose() {
			fmt.Println("No data available")
		}
		return nil
	}

	switch format {
	case "hex":
		for i, b := range resp.Data {
			if i > 0 && i%16 == 0 {
				fmt.Println()
			}
			fmt.Printf("%02x ", b)
		}
		fmt.Println()
	case "json":
		fmt.Printf("{\"data\":\"%x\",\"bytes_read\":%d}\n", resp.Data, resp.BytesRead)
	default: // text
		fmt.Print(string(resp.Data))
	}

	if IsVerbose() {
		fmt.Printf("\nRead %d bytes\n", resp.BytesRead)
	}

	return nil
}
