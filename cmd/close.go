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

var closeCmd = &cobra.Command{
	Use:   "close PORT [flags]",
	Short: "Close a serial port",
	Long: `Close an open serial port.

Example:
  seriallink close COM1                    # Close port by name`,
	Args: cobra.ExactArgs(1),
	RunE: runClose,
}

func init() {
	rootCmd.AddCommand(closeCmd)

	closeCmd.Flags().String("session-id", "", "session ID (required if not the opener)")
}

func runClose(cmd *cobra.Command, args []string) error {
	portName := args[0]
	sessionID, _ := cmd.Flags().GetString("session-id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	resp, err := client.ClosePort(ctx, &pb.ClosePortRequest{
		PortName:  portName,
		SessionId: sessionID,
	})
	if err != nil {
		return fmt.Errorf("failed to close port: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to close port: %s", resp.Message)
	}

	if IsVerbose() {
		fmt.Printf("Successfully closed %s\n", portName)
	} else {
		fmt.Printf("Closed %s\n", portName)
	}

	return nil
}
