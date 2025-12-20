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

var openCmd = &cobra.Command{
	Use:   "open PORT [flags]",
	Short: "Open a serial port",
	Long: `Open a serial port with the specified configuration.

Example:
  seriallink open COM1                           # Open with defaults (9600 baud)
  seriallink open COM1 --baud 115200             # Open with specific baud rate
  seriallink open /dev/ttyUSB0 --baud 9600 --data-bits 8 --stop-bits 1 --parity none`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().Uint32("baud", 9600, "baud rate")
	openCmd.Flags().String("data-bits", "8", "data bits (5, 6, 7, 8)")
	openCmd.Flags().String("stop-bits", "1", "stop bits (1, 1.5, 2)")
	openCmd.Flags().String("parity", "none", "parity (none, odd, even, mark, space)")
	openCmd.Flags().String("flow-control", "none", "flow control (none, hardware, software)")
	openCmd.Flags().String("client-id", "", "client ID for locking (auto-generated if not provided)")
}

func runOpen(cmd *cobra.Command, args []string) error {
	portName := args[0]

	baud, _ := cmd.Flags().GetUint32("baud")
	dataBits, _ := cmd.Flags().GetString("data-bits")
	stopBits, _ := cmd.Flags().GetString("stop-bits")
	parity, _ := cmd.Flags().GetString("parity")
	flowControl, _ := cmd.Flags().GetString("flow-control")
	clientID, _ := cmd.Flags().GetString("client-id")

	if clientID == "" {
		clientID = fmt.Sprintf("cli-%d", time.Now().UnixNano())
	}

	// Map string values to protobuf enums
	dataBitsEnum := parseDataBits(dataBits)
	stopBitsEnum := parseStopBits(stopBits)
	parityEnum := parseParity(parity)
	flowControlEnum := parseFlowControl(flowControl)

	config := &pb.PortConfig{
		BaudRate:    baud,
		DataBits:    dataBitsEnum,
		StopBits:    stopBitsEnum,
		Parity:      parityEnum,
		FlowControl: flowControlEnum,
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

	resp, err := client.OpenPort(ctx, &pb.OpenPortRequest{
		PortName:  portName,
		Config:    config,
		ClientId:  clientID,
		Exclusive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open port: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to open port: %s", resp.Message)
	}

	if IsVerbose() {
		fmt.Printf("Successfully opened %s\n", portName)
		fmt.Printf("  Baud Rate:    %d\n", baud)
		fmt.Printf("  Data Bits:    %s\n", dataBits)
		fmt.Printf("  Stop Bits:    %s\n", stopBits)
		fmt.Printf("  Parity:       %s\n", parity)
		fmt.Printf("  Flow Control: %s\n", flowControl)
		fmt.Printf("  Session ID:   %s\n", resp.SessionId)
	} else {
		fmt.Printf("Opened %s (Session: %s)\n", portName, resp.SessionId)
	}

	return nil
}

func parseDataBits(s string) pb.DataBits {
	switch s {
	case "5":
		return pb.DataBits_DATA_BITS_5
	case "6":
		return pb.DataBits_DATA_BITS_6
	case "7":
		return pb.DataBits_DATA_BITS_7
	case "8":
		return pb.DataBits_DATA_BITS_8
	default:
		return pb.DataBits_DATA_BITS_8
	}
}

func parseStopBits(s string) pb.StopBits {
	switch s {
	case "1":
		return pb.StopBits_STOP_BITS_1
	case "1.5":
		return pb.StopBits_STOP_BITS_1_5
	case "2":
		return pb.StopBits_STOP_BITS_2
	default:
		return pb.StopBits_STOP_BITS_1
	}
}

func parseParity(s string) pb.Parity {
	switch s {
	case "none":
		return pb.Parity_PARITY_NONE
	case "odd":
		return pb.Parity_PARITY_ODD
	case "even":
		return pb.Parity_PARITY_EVEN
	case "mark":
		return pb.Parity_PARITY_MARK
	case "space":
		return pb.Parity_PARITY_SPACE
	default:
		return pb.Parity_PARITY_NONE
	}
}

func parseFlowControl(s string) pb.FlowControl {
	switch s {
	case "none":
		return pb.FlowControl_FLOW_CONTROL_NONE
	case "hardware":
		return pb.FlowControl_FLOW_CONTROL_HARDWARE
	case "software":
		return pb.FlowControl_FLOW_CONTROL_SOFTWARE
	default:
		return pb.FlowControl_FLOW_CONTROL_NONE
	}
}
