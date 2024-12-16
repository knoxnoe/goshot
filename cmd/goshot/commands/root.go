package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/cmd/goshot/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var rootCmd = &cobra.Command{
	Use:   "goshot [file] [flags]",
	Short: config.Styles.Subtitle.Render("Create beautiful code screenshots with customizable styling"),
	Long: config.Styles.Title.Render("Goshot - Code Screenshot Generator") + "\n" +
		config.Styles.Info.Render("A powerful tool for creating beautiful code screenshots with customizable window chrome,\n"+
			"syntax highlighting, and backgrounds."),
	Args:               cobra.MaximumNArgs(1),
	DisableFlagParsing: false,
	TraverseChildren:   true,
	Run: func(cmd *cobra.Command, args []string) {
		if config.Default.Interactive {
			fmt.Println(config.Styles.Error.Render("Interactive mode is coming soon!"))
			os.Exit(1)
		}

		// Skip processing for subcommands
		if cmd.Name() != "goshot" {
			return
		}

		var code string
		var err error
		switch {
		case config.Default.FromClipboard:
			// Read from clipboard
			code, err = clipboard.ReadAll()
			if err != nil {
				fmt.Println(config.Styles.Error.Render("Failed to read from clipboard: " + err.Error()))
				os.Exit(1)
			}
		case len(args) > 0:
			// Read from file
			fileName := args[0]
			config.Default.Input = fileName
			content, err := os.ReadFile(fileName)
			if err != nil {
				fmt.Println(config.Styles.Error.Render("Failed to read file: " + err.Error()))
				os.Exit(1)
			}
			code = string(content)
		default:
			// Read from stdin
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println(config.Styles.Error.Render("Failed to read from stdin: " + err.Error()))
				os.Exit(1)
			}
			code = string(content)
		}

		if err := utils.RenderCode(&config.Default, !config.Default.ToStdout, code); err != nil {
			fmt.Println(config.Styles.Error.Render("Failed to render image: " + err.Error()))
			os.Exit(1)
		}
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add flags to root command
	flags := rootCmd.PersistentFlags()

	flags.StringVar(&config.Default.Language, "language", "", "Language override")

	// Add flag sets
	rootCmd.PersistentFlags().AddFlagSet(makeOutputFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeGradientFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeShadowFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeLayoutFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeRedactionFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeAppearanceFlagSet())

	// Add commands
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(rootThemesCmd)
	rootCmd.AddCommand(fontsCmd)
	rootCmd.AddCommand(languagesCmd)

	// Set usage function
	setUsageFunc(rootCmd, map[*pflag.FlagSet]string{
		makeOutputFlagSet():     "output",
		makeGradientFlagSet():   "gradient",
		makeShadowFlagSet():     "shadow",
		makeLayoutFlagSet():     "layout",
		makeRedactionFlagSet():  "redaction",
		makeAppearanceFlagSet(): "appearance",
	})
}

func makeOutputFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("output", pflag.ContinueOnError)
	fs.StringVarP(&config.Default.OutputFile, "output", "o", "", "Write output image to specific location instead of cwd")
	fs.BoolVarP(&config.Default.ToClipboard, "to-clipboard", "c", false, "Copy the output image to clipboard")
	fs.BoolVar(&config.Default.FromClipboard, "from-clipboard", false, "Read input from clipboard")
	fs.BoolVarP(&config.Default.ToStdout, "to-stdout", "s", false, "Write output to stdout")
	fs.BoolVar(&config.Default.ShowPrompt, "show-prompt", false, "Show the prompt used to generate the screenshot")
	return fs
}

func makeGradientFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("gradient", pflag.ExitOnError)
	fs.StringVar(&config.Default.GradientType, "gradient-type", "", "Gradient type (linear, radial, angular, diamond, spiral, square, star)")
	fs.StringSliceVar(&config.Default.GradientStops, "gradient-stops", []string{}, "Gradient stops (e.g., '#232323;0', '#383838;100')")
	fs.Float64Var(&config.Default.GradientAngle, "gradient-angle", 45.0, "Gradient angle (degrees)")
	fs.Float64Var(&config.Default.GradientCenterX, "gradient-center-x", 0.5, "Gradient center X")
	fs.Float64Var(&config.Default.GradientCenterY, "gradient-center-y", 0.5, "Gradient center Y")
	fs.Float64Var(&config.Default.GradientIntensity, "gradient-intensity", 5.0, "Gradient intensity")
	return fs
}

func makeShadowFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("shadow", pflag.ExitOnError)
	fs.Float64Var(&config.Default.ShadowBlurRadius, "shadow-blur", 0.0, "Shadow blur radius")
	fs.StringVar(&config.Default.ShadowColor, "shadow-color", "#00000033", "Shadow color")
	fs.Float64Var(&config.Default.ShadowSpread, "shadow-spread", 0.0, "Shadow spread")
	fs.Float64Var(&config.Default.ShadowOffsetX, "shadow-offset-x", 0.0, "Shadow offset X")
	fs.Float64Var(&config.Default.ShadowOffsetY, "shadow-offset-y", 0.0, "Shadow offset Y")
	return fs
}

func makeLayoutFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("layout", pflag.ExitOnError)
	fs.IntVar(&config.Default.LinePadding, "line-pad", 2, "Line padding")
	fs.IntVar(&config.Default.PadHoriz, "pad-horiz", 100, "Horizontal padding")
	fs.IntVar(&config.Default.PadVert, "pad-vert", 80, "Vertical padding")
	fs.IntVar(&config.Default.CodePadTop, "code-pad-top", 10, "Code top padding")
	fs.IntVar(&config.Default.CodePadBottom, "code-pad-bottom", 10, "Code bottom padding")
	fs.IntVar(&config.Default.CodePadLeft, "code-pad-left", 10, "Code left padding")
	fs.IntVar(&config.Default.CodePadRight, "code-pad-right", 10, "Code right padding")
	fs.IntVar(&config.Default.LineNumberPadding, "line-number-pad", 10, "Line number padding")
	fs.IntVar(&config.Default.MinWidth, "min-width", 0, "Minimum width")
	fs.IntVar(&config.Default.MaxWidth, "max-width", 0, "Maximum width")
	fs.IntVar(&config.Default.TabWidth, "tab-width", 4, "Tab width")
	return fs
}

func makeRedactionFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("redaction", pflag.ExitOnError)
	fs.BoolVar(&config.Default.RedactionEnabled, "redact", false, "Enable redaction of sensitive information")
	fs.StringVar(&config.Default.RedactionStyle, "redact-style", "block", "Redaction style (block or blur)")
	fs.Float64Var(&config.Default.RedactionBlurRadius, "redact-blur", 5.0, "Blur radius for redacted areas")
	fs.StringSliceVar(&config.Default.RedactionPatterns, "redact-pattern", []string{}, "Additional regex patterns for redaction (can be specified multiple times)")
	fs.StringSliceVar(&config.Default.RedactionAreas, "redact-area", []string{}, "Manual redaction areas in format 'x,y,width,height' (can be specified multiple times)")
	return fs
}

func makeAppearanceFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("appearance", pflag.ExitOnError)
	fs.StringVarP(&config.Default.WindowChrome, "chrome", "C", "mac", "Chrome style (mac, windows, gnome)")
	fs.StringVarP(&config.Default.ChromeThemeName, "chrome-theme", "T", "", "Chrome theme name")
	fs.BoolVarP(&config.Default.LightMode, "light-mode", "L", false, "Use light mode")
	fs.StringVarP(&config.Default.Theme, "theme", "t", "ayu-dark", "Syntax highlight theme name")
	fs.StringVarP(&config.Default.Font, "font", "f", "JetBrainsMonoNerdFont", "Fallback font list (e.g., 'Hack; SimSun=31')")
	fs.Float64Var(&config.Default.LineHeight, "line-height", 1.0, "Line height")
	fs.StringVarP(&config.Default.BackgroundColor, "background", "b", "#ABB8C3", "Background color")
	fs.StringVar(&config.Default.BackgroundImage, "background-image", "", "Background image path")
	fs.StringVar(&config.Default.BackgroundImageFit, "background-image-fit", "cover", "Background image fit (contain, cover, fill, stretch, tile)")
	fs.Float64Var(&config.Default.CornerRadius, "corner-radius", 10.0, "Corner radius of the image")
	fs.BoolVar(&config.Default.NoWindowControls, "no-window-controls", false, "Hide window controls")
	fs.StringVar(&config.Default.WindowTitle, "window-title", "", "Window title")
	fs.Float64Var(&config.Default.WindowCornerRadius, "window-corner-radius", 10.0, "Corner radius of the window")
	fs.StringSliceVar(&config.Default.LineRanges, "line-range", []string{}, "Line range (e.g. 1-10)")
	fs.StringSliceVar(&config.Default.HighlightLines, "highlight-lines", []string{}, "Highlight lines")
	return fs
}

func setUsageFunc(cmd *cobra.Command, rfg map[*pflag.FlagSet]string) {
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(config.Styles.Subtitle.Render("Usage:"))
		fmt.Printf("  %s", cmd.Name())

		fmt.Printf(" %s", cmd.Use)
		fmt.Println()
		fmt.Println()

		// Flags by group
		fmt.Println(config.Styles.Subtitle.Render("Flags:"))

		// Disable sorting to have consistent output
		for fs := range rfg {
			fs.SortFlags = false
		}

		// Sort flags by group
		var keys []*pflag.FlagSet
		for k := range rfg {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return rfg[keys[i]] < rfg[keys[j]]
		})

		for _, fs := range keys {
			// Get group name
			name := rfg[fs]

			// Print group title
			fmt.Println(config.Styles.GroupTitle.Render("  " + cases.Title(language.English).String(name) + ":"))

			// Get all flags in this group
			fs.VisitAll(func(f *pflag.Flag) {
				// Format the flag name part
				namePart := "  "
				if f.Shorthand != "" {
					namePart += lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
						Render(fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name))
				} else {
					namePart += lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
						Render(fmt.Sprintf("--%s", f.Name))
				}

				// Add type if not a boolean
				if f.Value.Type() != "bool" {
					namePart += " " + config.Styles.DefaultStyle.Render(f.Value.Type())
				}

				// Format description and default value
				desc := f.Usage
				if f.DefValue != "" && f.DefValue != "false" {
					desc += config.Styles.DefaultStyle.Render(fmt.Sprintf(" (default %q)", f.DefValue))
				}

				// Calculate padding
				padding := 40 - lipgloss.Width(namePart)
				if padding < 2 {
					padding = 2
				}

				// Print the formatted flag line
				fmt.Printf("%s%s%s\n",
					namePart,
					strings.Repeat(" ", padding),
					config.Styles.DescStyle.Render(desc),
				)
			})
			fmt.Println()
		}

		// Additional commands
		if cmd.HasAvailableSubCommands() {
			fmt.Println(config.Styles.Subtitle.Render("Additional Commands:"))
			for _, subCmd := range cmd.Commands() {
				if !subCmd.Hidden {
					fmt.Printf("  %s%s%s\n",
						config.Styles.FlagStyle.Render(subCmd.Name()),
						strings.Repeat(" ", 20-len(subCmd.Name())),
						config.Styles.DescStyle.Render(subCmd.Short),
					)
				}
			}
			fmt.Println()
		}

		return nil
	})
}
