package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of SerialLink",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("SerialLink version %s\n", Version)
	},
}

// RegisterVersionCommand adds the version command to the root command
func RegisterVersionCommand(root *cobra.Command) {
	root.AddCommand(versionCmd)
}
