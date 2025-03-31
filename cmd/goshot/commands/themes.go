package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/content/code"
)

var rootThemesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		themes := code.GetAvailableStyles()
		sort.Strings(themes)

		fmt.Println(config.Styles.Subtitle.Render("Available Themes:"))
		fmt.Println()

		for _, theme := range themes {
			fmt.Printf("  %s\n", config.Styles.Info.Render(theme))
		}
	},
}
