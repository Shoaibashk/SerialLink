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
	"os"
	"text/tabwriter"
	"time"

	pb "github.com/Shoaibashk/SerialLink/api/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan and list available serial ports",
	Long: `Scan the system for available serial ports and display their information.

This command discovers all serial ports including:
  • USB serial devices
  • Native serial ports
  • Bluetooth serial ports
  • Virtual serial ports

Example:
  seriallink scan              # List all ports
  seriallink scan --json       # Output as JSON
  seriallink scan -v           # Show detailed port information`,
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().Bool("json", false, "output in JSON format")
	scanCmd.Flags().BoolP("verbose", "v", false, "show detailed port information")
}

func runScan(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to the gRPC service
	addr := GetAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewSerialServiceClient(conn)

	// List ports
	resp, err := client.ListPorts(ctx, &pb.ListPortsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list ports: %w", err)
	}

	if len(resp.Ports) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No serial ports found.")
		}
		return nil
	}

	if jsonOutput {
		return printPortsJSON(resp.Ports, verbose)
	}

	return printPortsTable(resp.Ports, verbose)
}

func printPortsTable(ports []*pb.PortInfo, verbose bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if verbose {
		fmt.Fprintln(w, "PORT\tDESCRIPTION\tHARDWARE ID\tMANUFACTURER\tPRODUCT\tSERIAL\tTYPE\tSTATUS")
		fmt.Fprintln(w, "----\t-----------\t-----------\t------------\t-------\t------\t----\t------")
		for _, port := range ports {
			status := "available"
			if port.IsOpen {
				status = fmt.Sprintf("open (by %s)", port.LockedBy)
			}
			portType := port.PortType.String()
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				port.Name,
				truncate(port.Description, 20),
				truncate(port.HardwareId, 15),
				truncate(port.Manufacturer, 12),
				truncate(port.Product, 15),
				truncate(port.SerialNumber, 15),
				portType,
				status,
			)
		}
	} else {
		fmt.Fprintln(w, "PORT\tDESCRIPTION\tTYPE")
		fmt.Fprintln(w, "----\t-----------\t----")
		for _, port := range ports {
			portType := port.PortType.String()
			status := ""
			if port.IsOpen {
				status = " [OPEN]"
			}
			fmt.Fprintf(w, "%s%s\t%s\t%s\n",
				port.Name,
				status,
				truncate(port.Description, 40),
				portType,
			)
		}
	}

	return w.Flush()
}

func printPortsJSON(ports []*pb.PortInfo, verbose bool) error {
	// Convert to JSON-friendly format
	type PortData struct {
		Name         string `json:"name"`
		Description  string `json:"description,omitempty"`
		HardwareID   string `json:"hardware_id,omitempty"`
		Manufacturer string `json:"manufacturer,omitempty"`
		Product      string `json:"product,omitempty"`
		SerialNumber string `json:"serial_number,omitempty"`
		PortType     string `json:"port_type"`
		IsOpen       bool   `json:"is_open"`
		LockedBy     string `json:"locked_by,omitempty"`
	}

	var data []PortData
	for _, port := range ports {
		portData := PortData{
			Name:     port.Name,
			PortType: port.PortType.String(),
			IsOpen:   port.IsOpen,
		}
		if verbose {
			portData.Description = port.Description
			portData.HardwareID = port.HardwareId
			portData.Manufacturer = port.Manufacturer
			portData.Product = port.Product
			portData.SerialNumber = port.SerialNumber
			portData.LockedBy = port.LockedBy
		}
		data = append(data, portData)
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
