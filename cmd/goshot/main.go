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
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/syntax"
	"github.com/watzon/goshot/pkg/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	cliflag "k8s.io/component-base/cli/flag"
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
	ShowPrompt    bool
	AutoTitle     bool

	// Appearance
	WindowChrome       string
	ChromeThemeName    string
	DarkMode           bool
	Theme              string
	Language           string
	Font               string
	BackgroundColor    string
	BackgroundImage    string
	BackgroundImageFit string
	NoLineNumbers      bool
	CornerRadius       float64
	NoWindowControls   bool
	WindowTitle        string
	WindowCornerRadius float64

	// Gradient options
	GradientType      string
	GradientStops     []string
	GradientAngle     float64
	GradientCenterX   float64
	GradientCenterY   float64
	GradientIntensity float64

	// Padding and layout
	TabWidth      int
	StartLine     int
	EndLine       int
	LinePadding   int
	PadHoriz      int
	PadVert       int
	CodePadTop    int
	CodePadBottom int
	CodePadLeft   int
	CodePadRight  int

	// Shadow options
	ShadowBlurRadius float64
	ShadowColor      string
	ShadowSpread     float64
	ShadowOffsetX    float64
	ShadowOffsetY    float64

	// Highlighting
	HighlightLines string
}

func main() {
	var config Config

	rootCmd := &cobra.Command{
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

	execCommand := &cobra.Command{
		Use:   "exec [command]",
		Short: styles.subtitle.Render("Execute a command and create a screenshot of its output"),
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

			code := string(input)
			// if err := renderTerminalOutput(&config, false, code); err != nil {
			// 	fmt.Println(styles.error.Render(err.Error()))
			// 	os.Exit(1)
			// }
			fmt.Print(code)
		},
	}

	themesCmd := &cobra.Command{
		Use:   "themes",
		Short: "List available themes",
		Run: func(cmd *cobra.Command, args []string) {
			themes := syntax.GetAvailableStyles()
			sort.Strings(themes)

			fmt.Println(styles.subtitle.Render("Available Themes:"))
			fmt.Println()

			for _, theme := range themes {
				fmt.Printf("  %s\n", styles.info.Render(theme))
			}
		},
	}

	fontsCmd := &cobra.Command{
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

	languagesCmd := &cobra.Command{
		Use:   "languages",
		Short: "List available languages",
		Run: func(cmd *cobra.Command, args []string) {
			languages := syntax.GetAvailableLanguages(false)
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Revision: %s\n", version.Revision)
			fmt.Printf("Date: %s\n", version.Date)
		},
	}

	rfs := cliflag.NamedFlagSets{}
	rfg := map[*pflag.FlagSet]string{}

	appearanceFlagSet := rfs.FlagSet("appearance")
	outputFlagSet := rfs.FlagSet("output")
	layoutFlagSet := rfs.FlagSet("layout")
	gradientFlagSet := rfs.FlagSet("gradient")
	shadowFlagSet := rfs.FlagSet("shadow")

	// Output flags
	outputFlagSet.StringVarP(&config.OutputFile, "output", "o", "output.png", "Write output image to specific location instead of cwd")
	outputFlagSet.BoolVarP(&config.ToClipboard, "to-clipboard", "c", false, "Copy the output image to clipboard")
	outputFlagSet.BoolVar(&config.FromClipboard, "from-clipboard", false, "Read input from clipboard")
	outputFlagSet.BoolVarP(&config.ToStdout, "to-stdout", "s", false, "Write output to stdout")
	outputFlagSet.BoolVar(&config.ShowPrompt, "show-prompt", false, "Show the prompt used to generate the screenshot")
	outputFlagSet.BoolVar(&config.AutoTitle, "auto-title", false, "Automatically set the window title to the filename or command")
	rootCmd.Flags().AddFlagSet(outputFlagSet)
	execCommand.Flags().AddFlagSet(outputFlagSet)
	rfg[outputFlagSet] = "output"

	// Interactive mode
	appearanceFlagSet.BoolVarP(&config.Interactive, "interactive", "i", false, "Interactive mode")

	// Appearance flags
	appearanceFlagSet.StringVarP(&config.WindowChrome, "chrome", "C", "mac", "Chrome style (mac, windows, gnome)")
	appearanceFlagSet.StringVarP(&config.ChromeThemeName, "chrome-theme", "T", "", "Chrome theme name")
	appearanceFlagSet.BoolVarP(&config.DarkMode, "dark-mode", "d", false, "Use dark mode")
	appearanceFlagSet.StringVarP(&config.Theme, "theme", "t", "dracula", "Syntax highlight theme (name or .tmTheme file)")
	appearanceFlagSet.StringVarP(&config.Language, "language", "l", "", "Language for syntax highlighting (e.g., 'Rust' or 'rs')")
	appearanceFlagSet.StringVarP(&config.Font, "font", "f", "", "Fallback font list (e.g., 'Hack; SimSun=31')")
	appearanceFlagSet.StringVarP(&config.BackgroundColor, "background", "b", "#aaaaff", "Background color")
	appearanceFlagSet.StringVar(&config.BackgroundImage, "background-image", "", "Background image path")
	appearanceFlagSet.StringVar(&config.BackgroundImageFit, "background-image-fit", "cover", "Background image fit (contain, cover, fill, stretch, tile)")
	appearanceFlagSet.BoolVar(&config.NoLineNumbers, "no-line-number", false, "Hide line numbers")
	appearanceFlagSet.Float64Var(&config.CornerRadius, "corner-radius", 10.0, "Corner radius of the image")
	appearanceFlagSet.BoolVar(&config.NoWindowControls, "no-window-controls", false, "Hide window controls")
	appearanceFlagSet.StringVar(&config.WindowTitle, "window-title", "", "Window title")
	appearanceFlagSet.Float64Var(&config.WindowCornerRadius, "window-corner-radius", 10, "Corner radius of the window")
	appearanceFlagSet.StringVar(&config.HighlightLines, "highlight-lines", "", "Lines to highlight (e.g., '1-3;4')")
	rootCmd.Flags().AddFlagSet(appearanceFlagSet)
	execCommand.Flags().AddFlagSet(appearanceFlagSet)
	rfg[appearanceFlagSet] = "appearance"

	// Layout flags
	layoutFlagSet.IntVar(&config.TabWidth, "tab-width", 4, "Tab width")
	layoutFlagSet.IntVar(&config.StartLine, "start-line", 1, "Start line number")
	layoutFlagSet.IntVar(&config.EndLine, "end-line", 0, "End line number")
	layoutFlagSet.IntVar(&config.LinePadding, "line-pad", 2, "Padding between lines")
	layoutFlagSet.IntVar(&config.PadHoriz, "pad-horiz", 80, "Horizontal padding")
	layoutFlagSet.IntVar(&config.PadVert, "pad-vert", 100, "Vertical padding")
	layoutFlagSet.IntVar(&config.CodePadTop, "code-pad-top", 10, "Code top padding")
	layoutFlagSet.IntVar(&config.CodePadBottom, "code-pad-bottom", 10, "Code bottom padding")
	layoutFlagSet.IntVar(&config.CodePadLeft, "code-pad-left", 10, "Code left padding")
	layoutFlagSet.IntVar(&config.CodePadRight, "code-pad-right", 10, "Code right padding")
	rootCmd.Flags().AddFlagSet(layoutFlagSet)
	execCommand.Flags().AddFlagSet(layoutFlagSet)
	rfg[layoutFlagSet] = "layout"

	// Gradient flags
	gradientFlagSet.StringVarP(&config.GradientType, "gradient-type", "g", "", "Gradient type (linear, radial, angular, diamond, spiral, square, star)")
	gradientFlagSet.StringArrayVarP(&config.GradientStops, "gradient-stop", "G", []string{"#232323;0", "#383838;100"}, "Gradient stops (-G '#ff0000;0' -G '#00ff00;100')")
	gradientFlagSet.Float64Var(&config.GradientAngle, "gradient-angle", 45, "Gradient angle in degrees")
	gradientFlagSet.Float64Var(&config.GradientCenterX, "gradient-center-x", 0.5, "Gradient center X")
	gradientFlagSet.Float64Var(&config.GradientCenterY, "gradient-center-y", 0.5, "Gradient center Y")
	gradientFlagSet.Float64Var(&config.GradientIntensity, "gradient-intensity", 5, "Gradient intensity")
	rootCmd.Flags().AddFlagSet(gradientFlagSet)
	execCommand.Flags().AddFlagSet(gradientFlagSet)
	rfg[gradientFlagSet] = "gradient"

	// Shadow flags
	shadowFlagSet.Float64Var(&config.ShadowBlurRadius, "shadow-blur", 0, "Shadow blur radius (0 to disable)")
	shadowFlagSet.StringVar(&config.ShadowColor, "shadow-color", "#00000033", "Shadow color")
	shadowFlagSet.Float64Var(&config.ShadowSpread, "shadow-spread", 0, "Shadow spread radius")
	shadowFlagSet.Float64Var(&config.ShadowOffsetX, "shadow-offset-x", 0, "Shadow X offset")
	shadowFlagSet.Float64Var(&config.ShadowOffsetY, "shadow-offset-y", 0, "Shadow Y offset")
	rootCmd.Flags().AddFlagSet(shadowFlagSet)
	execCommand.Flags().AddFlagSet(shadowFlagSet)
	rfg[shadowFlagSet] = "shadow"

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(styles.subtitle.Render("Usage:"))
		fmt.Printf("  %s", cmd.Name())

		if cmd == rootCmd {
			fmt.Printf(" [flags] [file]")
		} else if cmd.HasAvailableLocalFlags() {
			fmt.Printf(" [flags]")
		}
		fmt.Println()
		fmt.Println()

		// Only show flags for root command
		if cmd == rootCmd {
			// Flags by group
			fmt.Println(styles.subtitle.Render("Flags:"))

			for fs, name := range rfg {
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
		} else if cmd.HasAvailableLocalFlags() {
			// For subcommands, use cobra's default flag display
			fmt.Println(styles.subtitle.Render("Flags:"))
			fmt.Println(cmd.LocalFlags().FlagUsages())
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

	// Additional utility commands
	rootCmd.AddCommand(execCommand, themesCmd, fontsCmd, languagesCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}
