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

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get service information",
	Long: `Get information about the SerialLink service including version, configuration, and features.

Example:
  seriallink info                # Display service information
  seriallink info --json         # Output as JSON`,
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().Bool("json", false, "output in JSON format")
}

func runInfo(cmd *cobra.Command, args []string) error {
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

	resp, err := client.GetAgentInfo(ctx, &pb.GetAgentInfoRequest{})
	if err != nil {
		return fmt.Errorf("failed to get agent info: %w", err)
	}

	if jsonOutput {
		return printInfoJSON(resp)
	}

	return printInfoTable(resp)
}

func printInfoTable(info *pb.AgentInfo) error {
	fmt.Println("SerialLink Service Information:")
	fmt.Printf("\nVersion:\n")
	fmt.Printf("  Version:        %s\n", info.Version)
	fmt.Printf("  Build Commit:   %s\n", info.BuildCommit)
	fmt.Printf("  Build Date:     %s\n", info.BuildDate)

	fmt.Printf("\nSystem:\n")
	fmt.Printf("  OS:             %s\n", info.Os)
	fmt.Printf("  Architecture:   %s\n", info.Arch)

	if info.UptimeSeconds > 0 {
		uptime := formatUptime(info.UptimeSeconds)
		fmt.Printf("  Uptime:         %s\n", uptime)
	}

	if info.Config != nil {
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  gRPC Address:   %s\n", info.Config.GrpcAddress)
		fmt.Printf("  TLS Enabled:    %v\n", info.Config.TlsEnabled)
		fmt.Printf("  Max Connections: %d\n", info.Config.MaxConnections)
	}

	if len(info.SupportedFeatures) > 0 {
		fmt.Printf("\nSupported Features:\n")
		for _, feature := range info.SupportedFeatures {
			fmt.Printf("  • %s\n", feature)
		}
	}

	return nil
}

func printInfoJSON(info *pb.AgentInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func formatUptime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
