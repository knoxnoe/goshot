package main

import (
	"fmt"
	"os"

	"github.com/watzon/goshot/cmd/goshot/commands"
	"github.com/watzon/goshot/cmd/goshot/config"
)

func main() {
	// Initialize configuration
	config.Initialize()

	// Execute root command
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, config.Styles.Error.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}
