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

	pb "github.com/Shoaibashk/SerialLink/api/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var configCmd = &cobra.Command{
	Use:   "config PORT [flags]",
	Short: "Manage port configuration",
	Long: `Get or configure settings for a serial port.

This command allows you to view current configuration or apply new configuration settings.

Example:
  seriallink config COM1                              # View current configuration
  seriallink config COM1 --baud 115200                # Change baud rate
  seriallink config COM1 --parity even --data-bits 7  # Change multiple settings`,
	Args: cobra.ExactArgs(1),
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().String("session-id", "", "session ID")
	configCmd.Flags().Bool("json", false, "output in JSON format")

	// Configuration flags
	configCmd.Flags().Uint32P("baud", "b", 0, "baud rate (0 = don't change)")
	configCmd.Flags().String("data-bits", "", "data bits (5, 6, 7, 8)")
	configCmd.Flags().String("stop-bits", "", "stop bits (1, 1.5, 2)")
	configCmd.Flags().String("parity", "", "parity (none, odd, even, mark, space)")
	configCmd.Flags().String("flow-control", "", "flow control (none, hardware, software)")
}

func runConfig(cmd *cobra.Command, args []string) error {
	portName := args[0]
	sessionID, _ := cmd.Flags().GetString("session-id")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Check if we're modifying or just viewing
	baud, _ := cmd.Flags().GetUint32("baud")
	dataBits, _ := cmd.Flags().GetString("data-bits")
	stopBits, _ := cmd.Flags().GetString("stop-bits")
	parity, _ := cmd.Flags().GetString("parity")
	flowControl, _ := cmd.Flags().GetString("flow-control")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	// If configuration flags are provided, apply them
	if baud > 0 || dataBits != "" || stopBits != "" || parity != "" || flowControl != "" {
		return applyConfig(client, ctx, portName, sessionID, baud, dataBits, stopBits, parity, flowControl)
	}

	// Otherwise, just get the current configuration
	resp, err := client.GetPortConfig(ctx, &pb.GetPortConfigRequest{
		PortName: portName,
	})
	if err != nil {
		return fmt.Errorf("failed to get port config: %w", err)
	}

	if jsonOutput {
		return printConfigJSON(resp)
	}

	return printConfigTable(resp)
}

func applyConfig(client pb.SerialServiceClient, ctx context.Context, portName, sessionID string, baud uint32, dataBits, stopBits, parity, flowControl string) error {
	// Start with current config
	currentResp, err := client.GetPortConfig(ctx, &pb.GetPortConfigRequest{
		PortName: portName,
	})
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	config := currentResp

	// Apply updates
	if baud > 0 {
		config.BaudRate = baud
	}
	if dataBits != "" {
		config.DataBits = parseDataBits(dataBits)
	}
	if stopBits != "" {
		config.StopBits = parseStopBits(stopBits)
	}
	if parity != "" {
		config.Parity = parseParity(parity)
	}
	if flowControl != "" {
		config.FlowControl = parseFlowControl(flowControl)
	}

	// Apply configuration
	resp, err := client.ConfigurePort(ctx, &pb.ConfigurePortRequest{
		PortName:  portName,
		SessionId: sessionID,
		Config:    config,
	})
	if err != nil {
		return fmt.Errorf("failed to configure port: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("configuration failed: %s", resp.Message)
	}

	if IsVerbose() {
		fmt.Printf("Successfully configured %s\n", portName)
		fmt.Printf("  Baud Rate:      %d\n", config.BaudRate)
		fmt.Printf("  Data Bits:      %s\n", getDataBitsString(config.DataBits))
		fmt.Printf("  Stop Bits:      %s\n", getStopBitsString(config.StopBits))
		fmt.Printf("  Parity:         %s\n", getParityString(config.Parity))
		fmt.Printf("  Flow Control:   %s\n", getFlowControlString(config.FlowControl))
	} else {
		fmt.Printf("Configured %s\n", portName)
	}

	return nil
}

func printConfigTable(config *pb.PortConfig) error {
	fmt.Println("Port Configuration:")
	fmt.Printf("  Baud Rate:      %d\n", config.BaudRate)
	fmt.Printf("  Data Bits:      %s\n", getDataBitsString(config.DataBits))
	fmt.Printf("  Stop Bits:      %s\n", getStopBitsString(config.StopBits))
	fmt.Printf("  Parity:         %s\n", getParityString(config.Parity))
	fmt.Printf("  Flow Control:   %s\n", getFlowControlString(config.FlowControl))
	if config.ReadTimeoutMs > 0 {
		fmt.Printf("  Read Timeout:   %d ms\n", config.ReadTimeoutMs)
	}
	if config.WriteTimeoutMs > 0 {
		fmt.Printf("  Write Timeout:  %d ms\n", config.WriteTimeoutMs)
	}
	return nil
}

func printConfigJSON(config *pb.PortConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
