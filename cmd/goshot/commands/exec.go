package commands

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/cmd/goshot/utils"
	"github.com/watzon/goshot/pkg/content/term"
)

var execCmd = &cobra.Command{
	Use:   "exec [args...] -- [command] [command args...]",
	Short: "Execute a command and create a screenshot of its output",
	Long:  config.Styles.Info.Render("Execute a command and create a beautiful screenshot of its output"),
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var input []byte
		var err error
		if len(args) == 0 {
			input, err = io.ReadAll(cmd.InOrStdin())
		} else {
			config.Default.Args = args
			input, err = utils.ExecuteCommand(cmd.Context(), args)
		}
		if err != nil {
			fmt.Println(config.Styles.Error.Render(err.Error()))
			os.Exit(1)
		}

		if err := utils.RenderTerm(&config.Default, true, args, input); err != nil {
			fmt.Println(config.Styles.Error.Render("Failed to render image: " + err.Error()))
			os.Exit(1)
		}
	},
}

var execThemesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		themes := term.ListThemes()
		sort.Strings(themes)

		fmt.Println(config.Styles.Subtitle.Render("Available Themes:"))
		fmt.Println()

		for _, theme := range themes {
			fmt.Printf("  %s\n", config.Styles.Info.Render(theme))
		}
	},
}

func init() {
	// Appearance flags
	rfg := map[*pflag.FlagSet]string{}

	// Gradient flags
	gradientFlags := makeGradientFlagSet()
	execCmd.Flags().AddFlagSet(gradientFlags)
	rfg[gradientFlags] = "gradient"

	// Shadow flags
	shadowFlags := makeShadowFlagSet()
	execCmd.Flags().AddFlagSet(shadowFlags)
	rfg[shadowFlags] = "shadow"

	// Output flags
	outputFlags := makeOutputFlagSet()
	outputFlags.BoolVar(&config.Default.AutoTitle, "auto-title", false, "Automatically set the window title to the filename or command")
	execCmd.Flags().AddFlagSet(outputFlags)
	rfg[outputFlags] = "output"

	// Appearance flags
	appearanceFlags := makeAppearanceFlagSet()
	appearanceFlags.StringVarP(&config.Default.PromptTemplate, "prompt-template", "P", "\x1b[1;35m❯ \x1b[0;32m[command]\x1b[0m\n", "Prompt template")
	execCmd.Flags().AddFlagSet(appearanceFlags)
	rfg[appearanceFlags] = "appearance"

	// Layout flags
	layoutFlags := makeLayoutFlagSet()
	layoutFlags.IntVarP(&config.Default.CellWidth, "width", "w", 120, "Terminal width in cells")
	layoutFlags.IntVarP(&config.Default.CellHeight, "height", "H", 40, "Terminal height in cells")
	layoutFlags.BoolVarP(&config.Default.AutoSize, "auto-size", "A", false, "Resize terminal to fit content (width and height must already be larger than content)")
	layoutFlags.IntVar(&config.Default.CellPadLeft, "pad-left", 1, "Left padding in cells")
	layoutFlags.IntVar(&config.Default.CellPadRight, "pad-right", 1, "Right padding in cells")
	layoutFlags.IntVar(&config.Default.CellPadTop, "pad-top", 1, "Top padding in cells")
	layoutFlags.IntVar(&config.Default.CellPadBottom, "pad-bottom", 1, "Bottom padding in cells")
	layoutFlags.IntVar(&config.Default.CellSpacing, "cell-spacing", 0, "Cell spacing in cells")
	layoutFlags.BoolVarP(&config.Default.ShowPrompt, "show-prompt", "p", false, "Show the prompt used to generate the screenshot")
	execCmd.Flags().AddFlagSet(layoutFlags)
	rfg[layoutFlags] = "layout"

	// Additional utility commands
	execCmd.AddCommand(execThemesCmd)

	// Set usage function
	setUsageFunc(execCmd, rfg)
}
