package main

import (
	"github.com/Shoaibashk/SerialLink/cmd"
)

var (
	// Version is set during build with ldflags
	Version = "dev"
)

func main() {
	cmd.Version = Version
	cmd.Execute()
}
