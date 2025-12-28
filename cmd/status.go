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
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/Shoaibashk/SerialLink-Proto/gen/go/seriallink/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var statusCmd = &cobra.Command{
	Use:   "status PORT [flags]",
	Short: "Get port status and statistics",
	Long: `Get the current status and statistics of a serial port.

Example:
  seriallink status COM1                  # Get port status
  seriallink status COM1 --json           # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().Bool("json", false, "output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) error {
	portName := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	resp, err := client.GetPortStatus(ctx, &pb.GetPortStatusRequest{
		PortName: portName,
	})
	if err != nil {
		return fmt.Errorf("failed to get port status: %w", err)
	}

	if jsonOutput {
		return printStatusJSON(resp.Status)
	}

	return printStatusTable(resp.Status)
}

func printStatusTable(status *pb.PortStatus) error {
	fmt.Printf("Port: %s\n", status.PortName)
	fmt.Printf("  Status:         %s\n", getStatusString(status.IsOpen))
	fmt.Printf("  Locked:         %v\n", status.IsLocked)
	if status.LockedBy != "" {
		fmt.Printf("  Locked By:      %s\n", status.LockedBy)
	}
	if status.SessionId != "" {
		fmt.Printf("  Session ID:     %s\n", status.SessionId)
	}

	if status.CurrentConfig != nil {
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  Baud Rate:      %d\n", status.CurrentConfig.BaudRate)
		fmt.Printf("  Data Bits:      %s\n", getDataBitsString(status.CurrentConfig.DataBits))
		fmt.Printf("  Stop Bits:      %s\n", getStopBitsString(status.CurrentConfig.StopBits))
		fmt.Printf("  Parity:         %s\n", getParityString(status.CurrentConfig.Parity))
		fmt.Printf("  Flow Control:   %s\n", getFlowControlString(status.CurrentConfig.FlowControl))
	}

	if status.Statistics != nil {
		stats := status.Statistics
		fmt.Printf("\nStatistics:\n")
		fmt.Printf("  Bytes Sent:     %d\n", stats.BytesSent)
		fmt.Printf("  Bytes Received: %d\n", stats.BytesReceived)
		fmt.Printf("  Errors:         %d\n", stats.Errors)
		if stats.OpenedAt > 0 {
			openTime := time.Unix(0, stats.OpenedAt)
			fmt.Printf("  Opened At:      %s\n", openTime.Format(time.RFC3339))
		}
		if stats.LastActivity > 0 {
			actTime := time.Unix(0, stats.LastActivity)
			fmt.Printf("  Last Activity:  %s\n", actTime.Format(time.RFC3339))
		}
	}

	return nil
}

func printStatusJSON(status *pb.PortStatus) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func getStatusString(isOpen bool) string {
	if isOpen {
		return "open"
	}
	return "closed"
}

func getDataBitsString(db pb.DataBits) string {
	switch db {
	case pb.DataBits_DATA_BITS_5:
		return "5"
	case pb.DataBits_DATA_BITS_6:
		return "6"
	case pb.DataBits_DATA_BITS_7:
		return "7"
	case pb.DataBits_DATA_BITS_8:
		return "8"
	default:
		return "unknown"
	}
}

func getStopBitsString(sb pb.StopBits) string {
	switch sb {
	case pb.StopBits_STOP_BITS_1:
		return "1"
	case pb.StopBits_STOP_BITS_1_5:
		return "1.5"
	case pb.StopBits_STOP_BITS_2:
		return "2"
	default:
		return "unknown"
	}
}

func getParityString(p pb.Parity) string {
	switch p {
	case pb.Parity_PARITY_NONE:
		return "none"
	case pb.Parity_PARITY_ODD:
		return "odd"
	case pb.Parity_PARITY_EVEN:
		return "even"
	case pb.Parity_PARITY_MARK:
		return "mark"
	case pb.Parity_PARITY_SPACE:
		return "space"
	default:
		return "unknown"
	}
}

func getFlowControlString(fc pb.FlowControl) string {
	switch fc {
	case pb.FlowControl_FLOW_CONTROL_NONE:
		return "none"
	case pb.FlowControl_FLOW_CONTROL_HARDWARE:
		return "hardware (RTS/CTS)"
	case pb.FlowControl_FLOW_CONTROL_SOFTWARE:
		return "software (XON/XOFF)"
	default:
		return "unknown"
	}
}
