package main

import (
	"os"

	"github.com/Shoaibashk/SerialLink/cmd"
)

var (
	// Version is set during build with ldflags
	Version = "dev"
)

func main() {
	cmd.Version = Version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
