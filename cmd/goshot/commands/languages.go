package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/content/code"
)

var languagesCmd = &cobra.Command{
	Use:   "languages",
	Short: "List available languages",
	Run: func(cmd *cobra.Command, args []string) {
		languages := code.GetAvailableLanguages(false)
		sort.Strings(languages)

		fmt.Println(config.Styles.Subtitle.Render("Available Languages:"))
		fmt.Println()

		// Group languages by first letter for better readability
		grouped := make(map[string][]string)
		for _, lang := range languages {
			firstChar := strings.ToUpper(string(lang[0]))
			grouped[firstChar] = append(grouped[firstChar], lang)
		}

		// Get sorted keys
		var keys []string
		for k := range grouped {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Print grouped languages
		for _, key := range keys {
			fmt.Printf("%s\n", config.Styles.Info.Render(key))
			for _, lang := range grouped[key] {
				fmt.Printf("  %s\n", lang)
			}
			fmt.Println()
		}
	},
}
