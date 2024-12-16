package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/fonts"
)

var fontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "List available fonts",
	Run: func(cmd *cobra.Command, args []string) {
		fontList := fonts.ListFonts()
		sort.Strings(fontList)

		fmt.Println(config.Styles.Subtitle.Render("Available Fonts:"))
		for _, font := range fontList {
			fmt.Printf("  %s\n", config.Styles.Info.Render(font))
		}
	},
}
