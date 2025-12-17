package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is the application version
	Version = "dev"

	// cfgFile is the path to the config file
	cfgFile string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "seriallink",
		Short: "SerialLink - cross-platform serial port agent",
		Long: `SerialLink is a cross-platform serial port agent written in Go.
It provides a CLI to manage and interact with serial port connections.`,
	}
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext executes the root command with a context
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seriallink/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")

	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic(err)
	}

	// Register subcommands
	RegisterVersionCommand(rootCmd)
	RegisterServeCommand(rootCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			return
		}

		// Search config in home directory with name ".seriallink" (without extension)
		configDir := filepath.Join(home, ".seriallink")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read environment variables with SERIALLINK_ prefix
	viper.SetEnvPrefix("SERIALLINK")
	viper.AutomaticEnv()

	// If a config file is found, read it but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		// Only log error if it's not a config file not found error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
}
