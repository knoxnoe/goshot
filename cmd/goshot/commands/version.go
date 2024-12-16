package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/watzon/goshot/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Revision: %s\n", version.Revision)
		fmt.Printf("Date: %s\n", version.Date)
	},
}
