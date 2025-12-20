package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the serial port agent",
	Long:  `Start the serial port agent to listen and forward serial connections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get port from flags first, then try viper
		port, _ := cmd.Flags().GetString("port")
		if port == "" {
			port = viper.GetString("port")
		}

		if port == "" {
			return fmt.Errorf("port is required (set via --port flag or SERIALLINK_PORT env var)")
		}

		verbose := viper.GetBool("verbose")
		if verbose {
			fmt.Printf("Starting SerialLink agent on port: %s\n", port)
		}

		// TODO: Implement actual serial port agent logic
		fmt.Printf("SerialLink agent running on port: %s\n", port)
		return nil
	},
}

// RegisterServeCommand adds the serve command to the root command
func RegisterServeCommand(root *cobra.Command) {
	root.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "", "serial port device (e.g., /dev/ttyUSB0 or COM3)")
	if err := viper.BindPFlag("port", serveCmd.Flags().Lookup("port")); err != nil {
		panic(fmt.Sprintf("Failed to bind port flag: %v", err))
	}

	serveCmd.Flags().IntP("baud", "b", 9600, "baud rate")
	if err := viper.BindPFlag("baud", serveCmd.Flags().Lookup("baud")); err != nil {
		panic(fmt.Sprintf("Failed to bind baud flag: %v", err))
	}
}
