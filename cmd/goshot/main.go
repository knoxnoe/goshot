package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/content/term"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
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
	CellWidth      int
	CellHeight     int
	AutoSize       bool
	CellPadLeft    int
	CellPadRight   int
	CellPadTop     int
	CellPadBottom  int
	CellSpacing    int
	ShowPrompt     bool
	AutoTitle      bool
	PromptTemplate string

	// Redaction settings
	RedactionEnabled    bool
	RedactionStyle      string
	RedactionBlurRadius float64
	RedactionPatterns   []string
	RedactionAreas      []string // Format: "x,y,width,height"
}

var config Config

func init() {
	// Initialize Viper first
	initConfig()

	// Initialize root command flags
	initRootConfig()

	// Initialize exec command flags
	initExecConfig()
}

func getConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error getting home directory: %v", err)))
			return ""
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configHome, "goshot")
}

func saveDefaultConfig() error {
	configDir := getConfigDir()
	if configDir == "" {
		return fmt.Errorf("could not determine config directory")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Skip if config file already exists
	if _, err := os.Stat(configFile); err == nil {
		return nil
	}

	// Get all settings from viper
	settings := viper.AllSettings()

	// Marshal to YAML
	yamlData, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the file
	if err := os.WriteFile(configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set config paths
	configDir := getConfigDir()
	if configDir != "" {
		viper.AddConfigPath(configDir)
	}
	viper.AddConfigPath(".")

	// Input/Output options
	viper.SetDefault("io.output_file", "output.png")
	viper.SetDefault("io.copy_to_clipboard", false)

	// Appearance
	viper.SetDefault("appearance.window_chrome", "mac")
	viper.SetDefault("appearance.chrome_theme", "")
	viper.SetDefault("appearance.light_mode", false)
	viper.SetDefault("appearance.theme", "ayu-dark")
	viper.SetDefault("appearance.font", "JetBrainsMonoNerdFont")
	viper.SetDefault("appearance.line_height", 1.0)
	viper.SetDefault("appearance.background.color", "#ABB8C3")
	viper.SetDefault("appearance.background.image.source", "")
	viper.SetDefault("appearance.background.image_fit", "cover")
	viper.SetDefault("appearance.background.gradient.type", "")
	viper.SetDefault("appearance.background.gradient.stops", []string{"#232323;0", "#383838;100"})
	viper.SetDefault("appearance.background.gradient.angle", 45.0)
	viper.SetDefault("appearance.background.gradient.center.x", 0.5)
	viper.SetDefault("appearance.background.gradient.center.y", 0.5)
	viper.SetDefault("appearance.background.gradient.intensity", 5.0)
	viper.SetDefault("appearance.line_numbers", true)
	viper.SetDefault("appearance.corner_radius", 10.0)
	viper.SetDefault("appearance.window.controls", true)
	viper.SetDefault("appearance.window.title", "")
	viper.SetDefault("appearance.window.corner_radius", 10.0)
	viper.SetDefault("appearance.lines.ranges", []string{})
	viper.SetDefault("appearance.lines.highlight", []string{})
	viper.SetDefault("appearance.shadow.blur_radius", 0.0)
	viper.SetDefault("appearance.shadow.color", "#00000033")
	viper.SetDefault("appearance.shadow.spread", 0.0)
	viper.SetDefault("appearance.shadow.offset.x", 0.0)
	viper.SetDefault("appearance.shadow.offset.y", 0.0)
	viper.SetDefault("appearance.layout.line_pad", 2)
	viper.SetDefault("appearance.layout.padding.horizontal", 60)
	viper.SetDefault("appearance.layout.padding.vertical", 50)
	viper.SetDefault("appearance.layout.code_padding.top", 10)
	viper.SetDefault("appearance.layout.code_padding.bottom", 10)
	viper.SetDefault("appearance.layout.code_padding.left", 10)
	viper.SetDefault("appearance.layout.code_padding.right", 10)
	viper.SetDefault("appearance.layout.line_number_padding", 10)
	viper.SetDefault("appearance.layout.width.min", 0)
	viper.SetDefault("appearance.layout.width.max", 0)
	viper.SetDefault("appearance.layout.tab_width", 4)

	// Terminal options
	viper.SetDefault("terminal.width", 120)
	viper.SetDefault("terminal.height", 40)
	viper.SetDefault("terminal.auto_size", false)
	viper.SetDefault("terminal.padding.left", 1)
	viper.SetDefault("terminal.padding.right", 1)
	viper.SetDefault("terminal.padding.top", 1)
	viper.SetDefault("terminal.padding.bottom", 1)
	viper.SetDefault("terminal.cell_spacing", 0)
	viper.SetDefault("terminal.show_prompt", false)
	viper.SetDefault("terminal.prompt_template", "\x1b[1;35m❯ \x1b[0;32m[command]\x1b[0m\n")

	// Redaction options
	viper.SetDefault("redaction.enabled", false)
	viper.SetDefault("redaction.style", "block")
	viper.SetDefault("redaction.blur_radius", 5.0)
	viper.SetDefault("redaction.patterns", []string{})
	viper.SetDefault("redaction.areas", []string{})

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; save default config
			if err := saveDefaultConfig(); err != nil {
				fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error saving default config: %v", err)))
			}
		} else {
			fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error reading config file: %v", err)))
		}
	}

	// Set up environment variable support
	viper.SetEnvPrefix("GOSHOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

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
			fileName := args[0]
			config.Input = fileName
			content, err := os.ReadFile(fileName)
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
	Use:   "exec [args...] -- [command] [command args...]",
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func makeOutputFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("output", pflag.ContinueOnError)
	fs.StringVarP(&config.OutputFile, "output", "o", viper.GetString("io.output_file"), "Write output image to specific location instead of cwd")
	fs.BoolVarP(&config.ToClipboard, "to-clipboard", "c", viper.GetBool("io.copy_to_clipboard"), "Copy the output image to clipboard")
	fs.BoolVar(&config.FromClipboard, "from-clipboard", viper.GetBool("io.from_clipboard"), "Read input from clipboard")
	fs.BoolVarP(&config.ToStdout, "to-stdout", "s", viper.GetBool("io.to_stdout"), "Write output to stdout")
	fs.BoolVar(&config.ShowPrompt, "show-prompt", viper.GetBool("terminal.show_prompt"), "Show the prompt used to generate the screenshot")
	return fs
}

func makeGradientFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("gradient", pflag.ExitOnError)

	fs.StringVar(&config.GradientType, "gradient-type", viper.GetString("appearance.background.gradient.type"), "Gradient type (linear, radial, angular, diamond, spiral, square, star)")
	fs.StringSliceVar(&config.GradientStops, "gradient-stops", viper.GetStringSlice("appearance.background.gradient.stops"), "Gradient stops (e.g., '#232323;0', '#383838;100')")
	fs.Float64Var(&config.GradientAngle, "gradient-angle", viper.GetFloat64("appearance.background.gradient.angle"), "Gradient angle (degrees)")
	fs.Float64Var(&config.GradientCenterX, "gradient-center-x", viper.GetFloat64("appearance.background.gradient.center.x"), "Gradient center X")
	fs.Float64Var(&config.GradientCenterY, "gradient-center-y", viper.GetFloat64("appearance.background.gradient.center.y"), "Gradient center Y")
	fs.Float64Var(&config.GradientIntensity, "gradient-intensity", viper.GetFloat64("appearance.background.gradient.intensity"), "Gradient intensity")

	return fs
}

func makeShadowFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("shadow", pflag.ExitOnError)

	fs.Float64Var(&config.ShadowBlurRadius, "shadow-blur", viper.GetFloat64("appearance.shadow.blur_radius"), "Shadow blur radius")
	fs.StringVar(&config.ShadowColor, "shadow-color", viper.GetString("appearance.shadow.color"), "Shadow color")
	fs.Float64Var(&config.ShadowSpread, "shadow-spread", viper.GetFloat64("appearance.shadow.spread"), "Shadow spread")
	fs.Float64Var(&config.ShadowOffsetX, "shadow-offset-x", viper.GetFloat64("appearance.shadow.offset.x"), "Shadow offset X")
	fs.Float64Var(&config.ShadowOffsetY, "shadow-offset-y", viper.GetFloat64("appearance.shadow.offset.y"), "Shadow offset Y")

	return fs
}

func makeAppearanceFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("appearance", pflag.ExitOnError)
	fs.StringVarP(&config.WindowChrome, "chrome", "C", viper.GetString("appearance.window_chrome"), "Chrome style (mac, windows, gnome)")
	fs.StringVarP(&config.ChromeThemeName, "chrome-theme", "T", viper.GetString("appearance.chrome_theme"), "Chrome theme name")
	fs.BoolVarP(&config.LightMode, "light-mode", "L", viper.GetBool("appearance.light_mode"), "Use light mode")
	fs.StringVarP(&config.Theme, "theme", "t", viper.GetString("appearance.theme"), "Syntax highlight theme name")
	fs.StringVarP(&config.Font, "font", "f", viper.GetString("appearance.font"), "Fallback font list (e.g., 'Hack; SimSun=31')")
	fs.Float64Var(&config.LineHeight, "line-height", viper.GetFloat64("appearance.line_height"), "Line height")
	fs.StringVarP(&config.BackgroundColor, "background", "b", viper.GetString("appearance.background.color"), "Background color")
	fs.StringVar(&config.BackgroundImage, "background-image", viper.GetString("appearance.background.image.source"), "Background image path")
	fs.StringVar(&config.BackgroundImageFit, "background-image-fit", viper.GetString("appearance.background.image_fit"), "Background image fit (contain, cover, fill, stretch, tile)")
	fs.Float64Var(&config.CornerRadius, "corner-radius", viper.GetFloat64("appearance.corner_radius"), "Corner radius of the image")
	fs.BoolVar(&config.NoWindowControls, "no-window-controls", !viper.GetBool("appearance.window.controls"), "Hide window controls")
	fs.StringVar(&config.WindowTitle, "window-title", viper.GetString("appearance.window.title"), "Window title")
	fs.Float64Var(&config.WindowCornerRadius, "window-corner-radius", viper.GetFloat64("appearance.window.corner_radius"), "Corner radius of the window")
	fs.StringSliceVar(&config.LineRanges, "line-range", viper.GetStringSlice("appearance.lines.ranges"), "Line range (e.g. 1-10)")
	fs.StringSliceVar(&config.HighlightLines, "highlight-lines", viper.GetStringSlice("appearance.lines.highlight"), "Highlight lines")

	// Add redaction flags
	redactionFlags := makeRedactionFlagSet()
	fs.AddFlagSet(redactionFlags)

	return fs
}

func makeLayoutFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("layout", pflag.ExitOnError)
	fs.IntVar(&config.LinePadding, "line-pad", viper.GetInt("appearance.layout.line_pad"), "Line padding")
	fs.IntVar(&config.PadHoriz, "pad-horiz", viper.GetInt("appearance.layout.padding.horizontal"), "Horizontal padding")
	fs.IntVar(&config.PadVert, "pad-vert", viper.GetInt("appearance.layout.padding.vertical"), "Vertical padding")
	fs.IntVar(&config.CodePadTop, "code-pad-top", viper.GetInt("appearance.layout.code_padding.top"), "Code top padding")
	fs.IntVar(&config.CodePadBottom, "code-pad-bottom", viper.GetInt("appearance.layout.code_padding.bottom"), "Code bottom padding")
	fs.IntVar(&config.CodePadLeft, "code-pad-left", viper.GetInt("appearance.layout.code_padding.left"), "Code left padding")
	fs.IntVar(&config.CodePadRight, "code-pad-right", viper.GetInt("appearance.layout.code_padding.right"), "Code right padding")
	fs.IntVar(&config.LineNumberPadding, "line-number-pad", viper.GetInt("appearance.layout.line_number_padding"), "Line number padding")
	fs.IntVar(&config.MinWidth, "min-width", viper.GetInt("appearance.layout.width.min"), "Minimum width")
	fs.IntVar(&config.MaxWidth, "max-width", viper.GetInt("appearance.layout.width.max"), "Maximum width")
	fs.IntVar(&config.TabWidth, "tab-width", viper.GetInt("appearance.layout.tab_width"), "Tab width")

	return fs
}

func makeRedactionFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("Redaction", pflag.ExitOnError)
	fs.BoolVar(&config.RedactionEnabled, "redact", viper.GetBool("redaction.enabled"), "Enable redaction of sensitive information")
	fs.StringVar(&config.RedactionStyle, "redact-style", viper.GetString("redaction.style"), "Redaction style (block or blur)")
	fs.Float64Var(&config.RedactionBlurRadius, "redact-blur", viper.GetFloat64("redaction.blur_radius"), "Blur radius for redacted areas")
	fs.StringSliceVar(&config.RedactionPatterns, "redact-pattern", viper.GetStringSlice("redaction.patterns"), "Additional regex patterns for redaction (can be specified multiple times)")
	fs.StringSliceVar(&config.RedactionAreas, "redact-area", viper.GetStringSlice("redaction.areas"), "Manual redaction areas in format 'x,y,width,height' (can be specified multiple times)")
	return fs
}

func initRootConfig() {
	// Add flags to root command
	flags := rootCmd.PersistentFlags()

	// Input/Output flags
	flags.StringVarP(&config.Input, "input", "i", "", "Input file")
	flags.StringVarP(&config.OutputFile, "output", "o", viper.GetString("io.output_file"), "Output file")
	flags.BoolVarP(&config.ToClipboard, "clipboard", "c", viper.GetBool("io.copy_to_clipboard"), "Copy to clipboard")
	flags.BoolVarP(&config.FromClipboard, "from-clipboard", "C", viper.GetBool("io.from_clipboard"), "Read from clipboard")
	flags.BoolVarP(&config.ToStdout, "stdout", "s", viper.GetBool("io.to_stdout"), "Write to stdout")

	// Appearance flags
	flags.StringVar(&config.WindowChrome, "chrome", viper.GetString("appearance.window_chrome"), "Chrome style (mac, windows, gnome)")
	flags.StringVar(&config.ChromeThemeName, "chrome-theme", viper.GetString("appearance.chrome_theme"), "Chrome theme name")
	flags.BoolVar(&config.LightMode, "light-mode", viper.GetBool("appearance.light_mode"), "Use light mode")
	flags.StringVar(&config.Theme, "theme", viper.GetString("appearance.theme"), "Theme name")
	flags.StringVar(&config.Language, "language", "", "Language override")
	flags.StringVar(&config.Font, "font", viper.GetString("appearance.font"), "Font family")
	flags.Float64Var(&config.LineHeight, "line-height", viper.GetFloat64("appearance.line_height"), "Line height")
	flags.StringVar(&config.BackgroundColor, "background-color", viper.GetString("appearance.background.color"), "Background color")
	flags.StringVar(&config.BackgroundImage, "background-image", viper.GetString("appearance.background.image.source"), "Background image")
	flags.StringVar(&config.BackgroundImageFit, "background-image-fit", viper.GetString("appearance.background.image_fit"), "Background image fit (fill, contain, cover, scale-down, none)")
	flags.BoolVar(&config.NoLineNumbers, "no-line-numbers", !viper.GetBool("appearance.line_numbers"), "Hide line numbers")
	flags.Float64Var(&config.CornerRadius, "corner-radius", viper.GetFloat64("appearance.corner_radius"), "Corner radius")
	flags.BoolVar(&config.NoWindowControls, "no-window-controls", !viper.GetBool("appearance.window.controls"), "Hide window controls")
	flags.StringVar(&config.WindowTitle, "window-title", viper.GetString("appearance.window.title"), "Window title")
	flags.Float64Var(&config.WindowCornerRadius, "window-corner-radius", viper.GetFloat64("appearance.window.corner_radius"), "Window corner radius")
	flags.StringSliceVar(&config.LineRanges, "line-ranges", viper.GetStringSlice("appearance.lines.ranges"), "Line ranges to show (e.g. 1-10,20-30)")
	flags.StringSliceVar(&config.HighlightLines, "highlight-lines", viper.GetStringSlice("appearance.lines.highlight"), "Line ranges to highlight")

	// Add flag sets
	rootCmd.PersistentFlags().AddFlagSet(makeGradientFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeShadowFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeLayoutFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(makeRedactionFlagSet())

	// Bind flags to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Add commands
	rootCmd.AddCommand(execCommand)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(rootThemesCmd)
	rootCmd.AddCommand(fontsCmd)
	rootCmd.AddCommand(languagesCmd)
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
	outputFlags.BoolVar(&config.AutoTitle, "auto-title", viper.GetBool("terminal.show_prompt"), "Automatically set the window title to the filename or command")
	execCommand.Flags().AddFlagSet(outputFlags)
	rfg[outputFlags] = "output"

	// Appearance flags
	appearanceFlags := makeAppearanceFlagSet()
	appearanceFlags.StringVarP(&config.PromptTemplate, "prompt-template", "P", viper.GetString("terminal.prompt_template"), "Prompt template")
	execCommand.Flags().AddFlagSet(appearanceFlags)
	rfg[appearanceFlags] = "appearance"

	// Layout flags
	layoutFlags := makeLayoutFlagSet()
	layoutFlags.IntVarP(&config.CellWidth, "width", "w", viper.GetInt("terminal.width"), "Terminal width in cells")
	layoutFlags.IntVarP(&config.CellHeight, "height", "H", viper.GetInt("terminal.height"), "Terminal height in cells")
	layoutFlags.BoolVarP(&config.AutoSize, "auto-size", "A", viper.GetBool("terminal.auto_size"), "Resize terminal to fit content (width and height must already be larger than content)")
	layoutFlags.IntVar(&config.CellPadLeft, "pad-left", viper.GetInt("terminal.padding.left"), "Left padding in cells")
	layoutFlags.IntVar(&config.CellPadRight, "pad-right", viper.GetInt("terminal.padding.right"), "Right padding in cells")
	layoutFlags.IntVar(&config.CellPadTop, "pad-top", viper.GetInt("terminal.padding.top"), "Top padding in cells")
	layoutFlags.IntVar(&config.CellPadBottom, "pad-bottom", viper.GetInt("terminal.padding.bottom"), "Bottom padding in cells")
	layoutFlags.IntVar(&config.CellSpacing, "cell-spacing", viper.GetInt("terminal.cell_spacing"), "Cell spacing in cells")
	layoutFlags.BoolVarP(&config.ShowPrompt, "show-prompt", "p", viper.GetBool("terminal.show_prompt"), "Show the prompt used to generate the screenshot")
	execCommand.Flags().AddFlagSet(layoutFlags)
	rfg[layoutFlags] = "layout"

	// Additional utility commands
	execCommand.AddCommand(execThemesCmd)

	// Set usage function
	setUsageFunc(rootCmd, rfg)
}

func setUsageFunc(cmd *cobra.Command, rfg map[*pflag.FlagSet]string) {
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(styles.subtitle.Render("Usage:"))
		fmt.Printf("  %s", cmd.Name())

		fmt.Printf(" %s", cmd.Use)
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
