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

// SerialLink is a cross-platform serial port background service that provides
// a gRPC API for managing serial port connections.
package main

import (
	"fmt"
	"os"

	"github.com/Shoaibashk/SerialLink/api"
	"github.com/Shoaibashk/SerialLink/cmd"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	// Set version info for cmd package
	cmd.Version = version
	cmd.Commit = commit
	cmd.BuildDate = buildDate

	// Set version info for api package
	api.Version = version
	api.Commit = commit
	api.BuildDate = buildDate

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
