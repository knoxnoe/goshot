package main

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
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/content/term"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Styles defines our CLI styles
var (
	styles = struct {
		title        lipgloss.Style
		subtitle     lipgloss.Style
		error        lipgloss.Style
		info         lipgloss.Style
		groupTitle   lipgloss.Style
		successBox   lipgloss.Style
		infoBox      lipgloss.Style
		flagStyle    lipgloss.Style
		descStyle    lipgloss.Style
		defaultStyle lipgloss.Style
	}{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#d7008f", Dark: "#FF79C6"}).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#7e3ff2", Dark: "#BD93F9"}).
			MarginBottom(1),
		error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#d70000", Dark: "#FF5555"}),
		successBox: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#388E3C"}),
		infoBox: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(lipgloss.AdaptiveColor{Light: "#0087af", Dark: "#8BE9FD"}),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#0087af", Dark: "#8BE9FD"}),
		groupTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#d75f00", Dark: "#FFB86C"}),

		flagStyle:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#50FA7B"}),
		descStyle:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#303030", Dark: "#F8F8F2"}),
		defaultStyle: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#6272A4"}).Italic(true),
	}
)

type Config struct {
	// Interactive mode
	Interactive bool

	// Input/Output options
	Input         string
	OutputFile    string
	ToClipboard   bool
	FromClipboard bool
	ToStdout      bool

	// Appearance
	WindowChrome       string
	ChromeThemeName    string
	LightMode          bool
	Theme              string
	Language           string
	Font               string
	LineHeight         float64
	BackgroundColor    string
	BackgroundImage    string
	BackgroundImageFit string
	NoLineNumbers      bool
	CornerRadius       float64
	NoWindowControls   bool
	WindowTitle        string
	WindowCornerRadius float64
	LineRanges         []string
	HighlightLines     []string

	// Gradient options
	GradientType      string
	GradientStops     []string
	GradientAngle     float64
	GradientCenterX   float64
	GradientCenterY   float64
	GradientIntensity float64

	// Padding and layout
	TabWidth          int
	StartLine         int
	EndLine           int
	LinePadding       int
	PadHoriz          int
	PadVert           int
	CodePadTop        int
	CodePadBottom     int
	CodePadLeft       int
	CodePadRight      int
	LineNumberPadding int
	MinWidth          int
	MaxWidth          int

	// Shadow options
	ShadowBlurRadius float64
	ShadowColor      string
	ShadowSpread     float64
	ShadowOffsetX    float64
	ShadowOffsetY    float64

	// Exec specific
	CellWidth     int
	CellHeight    int
	AutoSize      bool
	CellPadLeft   int
	CellPadRight  int
	CellPadTop    int
	CellPadBottom int
	CellSpacing   int
	ShowPrompt    bool
	AutoTitle     bool
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "goshot [file] [flags]",
	Short: styles.subtitle.Render("Create beautiful code screenshots with customizable styling"),
	Long: styles.title.Render("Goshot - Code Screenshot Generator") + "\n" +
		styles.info.Render("A powerful tool for creating beautiful code screenshots with customizable window chrome,\n"+
			"syntax highlighting, and backgrounds."),
	Args:               cobra.MaximumNArgs(1),
	DisableFlagParsing: false,
	TraverseChildren:   true,
	Run: func(cmd *cobra.Command, args []string) {
		if config.Interactive {
			fmt.Println(styles.error.Render("Interactive mode is coming soon!"))
			os.Exit(1)
		}

		// Skip processing for subcommands
		if cmd.Name() != "goshot" {
			return
		}

		var code string
		var err error
		switch {
		case config.FromClipboard:
			// Read from clipboard
			code, err = clipboard.ReadAll()
			if err != nil {
				fmt.Println(styles.error.Render("Failed to read from clipboard: " + err.Error()))
				os.Exit(1)
			}
		case len(args) > 0:
			// Read from file
			content, err := os.ReadFile(args[0])
			if err != nil {
				fmt.Println(styles.error.Render("Failed to read file: " + err.Error()))
				os.Exit(1)
			}
			code = string(content)
		default:
			// Read from stdin
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println(styles.error.Render("Failed to read from stdin: " + err.Error()))
				os.Exit(1)
			}
			code = string(content)
		}

		if err := renderCode(&config, !config.ToStdout, code); err != nil {
			fmt.Println(styles.error.Render("Failed to render image: " + err.Error()))
			os.Exit(1)
		}
	},
}

var execCommand = &cobra.Command{
	Use:   "exec [command]",
	Short: "Execute a command and create a screenshot of its output",
	Long:  styles.info.Render("Execute a command and create a beautiful screenshot of its output"),
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var input []byte
		var err error
		if len(args) == 0 {
			input, err = io.ReadAll(cmd.InOrStdin())
		} else {
			input, err = executeComamand(cmd.Context(), args)
		}
		if err != nil {
			fmt.Println(styles.error.Render(err.Error()))
			os.Exit(1)
		}

		if err := renderTerm(&config, true, args, input); err != nil {
			fmt.Println(styles.error.Render("Failed to render image: " + err.Error()))
			os.Exit(1)
		}
	},
}

var rootThemesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		themes := code.GetAvailableStyles()
		sort.Strings(themes)

		fmt.Println(styles.subtitle.Render("Available Themes:"))
		fmt.Println()

		for _, theme := range themes {
			fmt.Printf("  %s\n", styles.info.Render(theme))
		}
	},
}

var execThemesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		themes := term.ListThemes()
		sort.Strings(themes)

		fmt.Println(styles.subtitle.Render("Available Themes:"))
		fmt.Println()

		for _, theme := range themes {
			fmt.Printf("  %s\n", styles.info.Render(theme))
		}
	},
}

var fontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "List available fonts",
	Run: func(cmd *cobra.Command, args []string) {
		fonts := fonts.ListFonts()
		sort.Strings(fonts)

		fmt.Println(styles.subtitle.Render("Available Fonts:"))
		for _, font := range fonts {
			fmt.Printf("  %s\n", styles.info.Render(font))
		}
	},
}

var languagesCmd = &cobra.Command{
	Use:   "languages",
	Short: "List available languages",
	Run: func(cmd *cobra.Command, args []string) {
		languages := code.GetAvailableLanguages(false)
		sort.Strings(languages)

		fmt.Println(styles.subtitle.Render("Available Languages:"))
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
			fmt.Printf("%s\n", styles.info.Render(key))
			for _, lang := range grouped[key] {
				fmt.Printf("  %s\n", lang)
			}
			fmt.Println()
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Revision: %s\n", version.Revision)
		fmt.Printf("Date: %s\n", version.Date)
	},
}

func main() {
	initRootConfig()
	initExecConfig()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}
func makeOutputFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("output", pflag.ContinueOnError)
	fs.StringVarP(&config.OutputFile, "output", "o", "output.png", "Write output image to specific location instead of cwd")
	fs.BoolVarP(&config.ToClipboard, "to-clipboard", "c", false, "Copy the output image to clipboard")
	fs.BoolVar(&config.FromClipboard, "from-clipboard", false, "Read input from clipboard")
	fs.BoolVarP(&config.ToStdout, "to-stdout", "s", false, "Write output to stdout")
	fs.BoolVar(&config.ShowPrompt, "show-prompt", false, "Show the prompt used to generate the screenshot")
	return fs
}

func makeGradientFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("gradient", pflag.ContinueOnError)
	fs.StringVarP(&config.GradientType, "gradient-type", "g", "", "Gradient type (linear, radial, angular, diamond, spiral, square, star)")
	fs.StringArrayVarP(&config.GradientStops, "gradient-stop", "G", []string{"#232323;0", "#383838;100"}, "Gradient stops (-G '#ff0000;0' -G '#00ff00;100')")
	fs.Float64Var(&config.GradientAngle, "gradient-angle", 45, "Gradient angle in degrees")
	fs.Float64Var(&config.GradientCenterX, "gradient-center-x", 0.5, "Gradient center X")
	fs.Float64Var(&config.GradientCenterY, "gradient-center-y", 0.5, "Gradient center Y")
	fs.Float64Var(&config.GradientIntensity, "gradient-intensity", 5, "Gradient intensity")
	return fs
}

func makeShadowFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("shadow", pflag.ContinueOnError)
	fs.Float64Var(&config.ShadowBlurRadius, "shadow-blur", 0, "Shadow blur radius (0 to disable)")
	fs.StringVar(&config.ShadowColor, "shadow-color", "#00000033", "Shadow color")
	fs.Float64Var(&config.ShadowSpread, "shadow-spread", 0, "Shadow spread radius")
	fs.Float64Var(&config.ShadowOffsetX, "shadow-offset-x", 0, "Shadow X offset")
	fs.Float64Var(&config.ShadowOffsetY, "shadow-offset-y", 0, "Shadow Y offset")
	return fs
}

func makeAppearanceFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("appearance", pflag.ContinueOnError)
	fs.StringVarP(&config.WindowChrome, "chrome", "C", "mac", "Chrome style (mac, windows, gnome)")
	fs.StringVarP(&config.ChromeThemeName, "chrome-theme", "T", "", "Chrome theme name")
	fs.BoolVarP(&config.LightMode, "light-mode", "L", false, "Use light mode")
	fs.StringVarP(&config.Theme, "theme", "t", "seti", "Syntax highlight theme name")
	fs.StringVarP(&config.Font, "font", "f", "JetBrainsMonoNerdFont", "Fallback font list (e.g., 'Hack; SimSun=31')")
	fs.Float64Var(&config.LineHeight, "line-height", 1.0, "Line height")
	fs.StringVarP(&config.BackgroundColor, "background", "b", "#ABB8C3", "Background color")
	fs.StringVar(&config.BackgroundImage, "background-image", "", "Background image path")
	fs.StringVar(&config.BackgroundImageFit, "background-image-fit", "cover", "Background image fit (contain, cover, fill, stretch, tile)")
	fs.Float64Var(&config.CornerRadius, "corner-radius", 10.0, "Corner radius of the image")
	fs.BoolVar(&config.NoWindowControls, "no-window-controls", false, "Hide window controls")
	fs.StringVar(&config.WindowTitle, "window-title", "", "Window title")
	fs.Float64Var(&config.WindowCornerRadius, "window-corner-radius", 10, "Corner radius of the window")
	return fs
}

func makeLayoutFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("layout", pflag.ContinueOnError)
	fs.IntVar(&config.MinWidth, "tab-width", 4, "Tab width")
	return fs
}

func initRootConfig() {
	rfg := map[*pflag.FlagSet]string{}

	// Gradient flags
	gradientFlags := makeGradientFlagSet()
	rootCmd.Flags().AddFlagSet(gradientFlags)
	rfg[gradientFlags] = "gradient"

	// Shadow flags
	shadowFlags := makeShadowFlagSet()
	rootCmd.Flags().AddFlagSet(shadowFlags)
	rfg[shadowFlags] = "shadow"

	// Output flags
	outputFlags := makeOutputFlagSet()
	rootCmd.Flags().AddFlagSet(outputFlags)
	rfg[outputFlags] = "output"

	// Appearance flags
	appearanceFlags := makeAppearanceFlagSet()
	appearanceFlags.StringVarP(&config.Language, "language", "l", "", "Language for syntax highlighting (e.g., 'Rust' or 'rs')")
	appearanceFlags.BoolVar(&config.NoLineNumbers, "no-line-number", false, "Hide line numbers")
	appearanceFlags.StringArrayVarP(&config.LineRanges, "line-range", "r", []string{}, "Line ranges to render (e.g., '1..5' or '..10')")
	appearanceFlags.StringArrayVarP(&config.HighlightLines, "line-highlight", "R", []string{}, "Line ranges to highlight (e.g., '1..5' or '..10')")
	rootCmd.Flags().AddFlagSet(appearanceFlags)
	rfg[appearanceFlags] = "appearance"

	// Layout flags
	layoutFlags := makeLayoutFlagSet()
	layoutFlags.IntVar(&config.LinePadding, "line-pad", 2, "Padding between lines")
	layoutFlags.IntVar(&config.PadHoriz, "pad-horiz", 60, "Horizontal padding")
	layoutFlags.IntVar(&config.PadVert, "pad-vert", 50, "Vertical padding")
	layoutFlags.IntVar(&config.CodePadTop, "code-pad-top", 10, "Code top padding")
	layoutFlags.IntVar(&config.CodePadBottom, "code-pad-bottom", 10, "Code bottom padding")
	layoutFlags.IntVar(&config.CodePadLeft, "code-pad-left", 10, "Code left padding")
	layoutFlags.IntVar(&config.CodePadRight, "code-pad-right", 10, "Code right padding")
	layoutFlags.IntVar(&config.LineNumberPadding, "line-number-pad", 10, "Line number padding")
	layoutFlags.IntVar(&config.MinWidth, "min-width", 0, "Minimum width")
	layoutFlags.IntVar(&config.MaxWidth, "max-width", 0, "Maximum width")
	rootCmd.Flags().AddFlagSet(layoutFlags)
	rfg[layoutFlags] = "layout"

	// Additional utility commands
	rootCmd.AddCommand(execCommand, rootThemesCmd, fontsCmd, languagesCmd, versionCmd)

	// Set usage function
	setUsageFunc(rootCmd, rfg, "[flags] [file]")
}

func initExecConfig() {
	// Appearance flags
	rfg := map[*pflag.FlagSet]string{}

	// Gradient flags
	gradientFlags := makeGradientFlagSet()
	execCommand.Flags().AddFlagSet(gradientFlags)
	rfg[gradientFlags] = "gradient"

	// Shadow flags
	shadowFlags := makeShadowFlagSet()
	execCommand.Flags().AddFlagSet(shadowFlags)
	rfg[shadowFlags] = "shadow"

	// Output flags
	outputFlags := makeOutputFlagSet()
	outputFlags.BoolVar(&config.AutoTitle, "auto-title", false, "Automatically set the window title to the filename or command")
	execCommand.Flags().AddFlagSet(outputFlags)
	rfg[outputFlags] = "output"

	// Appearance flags
	appearanceFlags := makeAppearanceFlagSet()
	execCommand.Flags().AddFlagSet(appearanceFlags)
	rfg[appearanceFlags] = "appearance"

	// Layout flags
	layoutFlags := makeLayoutFlagSet()
	layoutFlags.IntVarP(&config.CellWidth, "width", "w", 120, "Terminal width in cells")
	layoutFlags.IntVarP(&config.CellHeight, "height", "H", 40, "Terminal height in cells")
	layoutFlags.BoolVarP(&config.AutoSize, "auto-size", "A", false, "Resize terminal to fit content (width and height must already be larger than content)")
	layoutFlags.IntVar(&config.CellPadLeft, "pad-left", 0, "Left padding in cells")
	layoutFlags.IntVar(&config.CellPadRight, "pad-right", 0, "Right padding in cells")
	layoutFlags.IntVar(&config.CellPadTop, "pad-top", 0, "Top padding in cells")
	layoutFlags.IntVar(&config.CellPadBottom, "pad-bottom", 0, "Bottom padding in cells")
	layoutFlags.IntVar(&config.CellSpacing, "cell-spacing", 0, "Cell spacing in cells")
	layoutFlags.BoolVarP(&config.ShowPrompt, "show-prompt", "p", false, "Show the prompt used to generate the screenshot")
	execCommand.Flags().AddFlagSet(layoutFlags)
	rfg[layoutFlags] = "layout"

	// Additional utility commands
	execCommand.AddCommand(execThemesCmd)

	// Set usage function
	setUsageFunc(rootCmd, rfg, "[flags] [command] [args...]")
}

func setUsageFunc(cmd *cobra.Command, rfg map[*pflag.FlagSet]string, usageStr string) {
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(styles.subtitle.Render("Usage:"))
		fmt.Printf("  %s", cmd.Name())

		if usageStr != "" {
			fmt.Printf(" %s", usageStr)
		} else if cmd.HasAvailableLocalFlags() {
			fmt.Printf(" [flags]")
		}
		fmt.Println()
		fmt.Println()

		// Flags by group
		fmt.Println(styles.subtitle.Render("Flags:"))

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
			fmt.Println(styles.groupTitle.Render("  " + cases.Title(language.English).String(name) + ":"))

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
					namePart += " " + styles.defaultStyle.Render(f.Value.Type())
				}

				// Format description and default value
				desc := f.Usage
				if f.DefValue != "" && f.DefValue != "false" {
					desc += styles.defaultStyle.Render(fmt.Sprintf(" (default %q)", f.DefValue))
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
					styles.descStyle.Render(desc),
				)
			})
			fmt.Println()
		}

		// Additional commands
		if cmd.HasAvailableSubCommands() {
			fmt.Println(styles.subtitle.Render("Additional Commands:"))
			for _, subCmd := range cmd.Commands() {
				if !subCmd.Hidden {
					fmt.Printf("  %s%s%s\n",
						styles.flagStyle.Render(subCmd.Name()),
						strings.Repeat(" ", 20-len(subCmd.Name())),
						styles.descStyle.Render(subCmd.Short),
					)
				}
			}
			fmt.Println()
		}

		return nil
	})
}
