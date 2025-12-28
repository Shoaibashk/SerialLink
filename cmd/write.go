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

	pb "github.com/Shoaibashk/SerialLink-Proto/gen/go/seriallink/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var writeCmd = &cobra.Command{
	Use:   "write PORT DATA [flags]",
	Short: "Write data to a serial port",
	Long: `Write data to an open serial port.

Example:
  seriallink write COM1 "Hello"            # Write text
  seriallink write COM1 "A\nB\nC"           # Write with newlines
  seriallink write COM1 --hex "48656C6C6F" # Write hex data`,
	Args: cobra.MinimumNArgs(2),
	RunE: runWrite,
}

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().Bool("flush", true, "flush buffer after write")
	writeCmd.Flags().String("session-id", "", "session ID")
	writeCmd.Flags().Bool("hex", false, "interpret data as hex string")
}

func runWrite(cmd *cobra.Command, args []string) error {
	portName := args[0]
	data := args[1]

	flush, _ := cmd.Flags().GetBool("flush")
	sessionID, _ := cmd.Flags().GetString("session-id")
	hexMode, _ := cmd.Flags().GetBool("hex")

	// Convert data
	var dataBytes []byte
	if hexMode {
		// Parse hex string
		_, err := fmt.Sscanf(data, "%x", &dataBytes)
		if err != nil {
			return fmt.Errorf("failed to parse hex data: %w", err)
		}
	} else {
		dataBytes = []byte(data)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	resp, err := client.Write(ctx, &pb.WriteRequest{
		PortName:  portName,
		SessionId: sessionID,
		Data:      dataBytes,
		Flush:     flush,
	})
	if err != nil {
		return fmt.Errorf("failed to write to port: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("write operation failed: %s", resp.Message)
	}

	if IsVerbose() {
		fmt.Printf("Wrote %d bytes to %s\n", resp.BytesWritten, portName)
	} else {
		fmt.Printf("Wrote %d bytes\n", resp.BytesWritten)
	}

	return nil
}
