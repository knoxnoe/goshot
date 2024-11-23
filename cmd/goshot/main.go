package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
	"github.com/watzon/goshot/pkg/syntax"
	"github.com/watzon/goshot/pkg/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	cliflag "k8s.io/component-base/cli/flag"
)

// Styles defines our CLI styles
var (
	styles = struct {
		title      lipgloss.Style
		subtitle   lipgloss.Style
		error      lipgloss.Style
		info       lipgloss.Style
		groupTitle lipgloss.Style
		successBox lipgloss.Style
		infoBox    lipgloss.Style
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
	}
)

type Config struct {
	// Interactive mode
	Interactive bool

	// Input/Output options
	Input          string
	OutputFile     string
	ToClipboard    bool
	FromClipboard  bool
	ToStdout       bool
	ExecuteCommand string
	ShowPrompt     bool
	AutoTitle      bool

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

// logMessage prints a styled message with consistent alignment
func logMessage(box lipgloss.Style, tag string, message string) {
	// Set a consistent width for the tag box and center the text
	const boxWidth = 11 // 9 characters + 2 padding spaces
	paddedTag := fmt.Sprintf("%*s", -boxWidth, tag)
	centeredBox := box.Width(boxWidth).Align(lipgloss.Center)
	fmt.Println(centeredBox.Render(paddedTag) + " " + styles.info.Render(message))
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

			if err := renderImage(&config, true, args); err != nil {
				fmt.Println(styles.error.Render(err.Error()))
				os.Exit(1)
			}
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
	outputFlagSet.StringVar(&config.ExecuteCommand, "execute", "", "Execute command and use output as input")
	outputFlagSet.BoolVar(&config.ShowPrompt, "show-prompt", false, "Show the prompt used to generate the screenshot")
	outputFlagSet.BoolVar(&config.AutoTitle, "auto-title", false, "Automatically set the window title to the filename or command")

	rootCmd.Flags().AddFlagSet(outputFlagSet)
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
	rfg[layoutFlagSet] = "layout"

	// Gradient flags
	gradientFlagSet.StringVarP(&config.GradientType, "gradient-type", "g", "", "Gradient type (linear, radial, angular, diamond, spiral, square, star)")
	gradientFlagSet.StringArrayVarP(&config.GradientStops, "gradient-stop", "G", []string{"#232323;0", "#383838;100"}, "Gradient stops (-G '#ff0000;0' -G '#00ff00;100')")
	gradientFlagSet.Float64Var(&config.GradientAngle, "gradient-angle", 45, "Gradient angle in degrees")
	gradientFlagSet.Float64Var(&config.GradientCenterX, "gradient-center-x", 0.5, "Gradient center X")
	gradientFlagSet.Float64Var(&config.GradientCenterY, "gradient-center-y", 0.5, "Gradient center Y")
	gradientFlagSet.Float64Var(&config.GradientIntensity, "gradient-intensity", 5, "Gradient intensity")
	rootCmd.Flags().AddFlagSet(gradientFlagSet)
	rfg[gradientFlagSet] = "gradient"

	// Shadow flags
	shadowFlagSet.Float64Var(&config.ShadowBlurRadius, "shadow-blur", 0, "Shadow blur radius (0 to disable)")
	shadowFlagSet.StringVar(&config.ShadowColor, "shadow-color", "#00000033", "Shadow color")
	shadowFlagSet.Float64Var(&config.ShadowSpread, "shadow-spread", 0, "Shadow spread radius")
	shadowFlagSet.Float64Var(&config.ShadowOffsetX, "shadow-offset-x", 0, "Shadow X offset")
	shadowFlagSet.Float64Var(&config.ShadowOffsetY, "shadow-offset-y", 0, "Shadow Y offset")
	rootCmd.Flags().AddFlagSet(shadowFlagSet)
	rfg[shadowFlagSet] = "shadow"

	rootCmd.MarkFlagsMutuallyExclusive("output", "to-clipboard", "to-stdout")
	rootCmd.MarkFlagsMutuallyExclusive("from-clipboard", "execute")

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(styles.subtitle.Render("Usage:"))
		fmt.Printf("  %s [flags] [file]\n", cmd.Name())
		fmt.Println()

		// Flags by group
		fmt.Println(styles.subtitle.Render("Flags:"))
		flagStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#50FA7B"})
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#303030", Dark: "#F8F8F2"})
		defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#6272A4"}).Italic(true)

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
					namePart += " " + defaultStyle.Render(f.Value.Type())
				}

				// Format description and default value
				desc := f.Usage
				if f.DefValue != "" && f.DefValue != "false" {
					desc += defaultStyle.Render(fmt.Sprintf(" (default %q)", f.DefValue))
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
					descStyle.Render(desc),
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
						flagStyle.Render(subCmd.Name()),
						strings.Repeat(" ", 20-len(subCmd.Name())),
						descStyle.Render(subCmd.Short),
					)
				}
			}
			fmt.Println()
		}

		return nil
	})

	// Additional utility commands
	rootCmd.AddCommand(
		&cobra.Command{
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
		},
		&cobra.Command{
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
		},
		&cobra.Command{
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
		},
		&cobra.Command{
			Use:   "version",
			Short: "Print version information",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("Version: %s\n", version.Version)
				fmt.Printf("Revision: %s\n", version.Revision)
				fmt.Printf("Date: %s\n", version.Date)
			},
		},
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.error.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func renderImage(config *Config, echo bool, args []string) error {
	// Get the input code
	var code string
	var err error

	switch {
	case config.ExecuteCommand != "":
		var stdout bytes.Buffer
		env := append(os.Environ(),
			"TERM=xterm-256color", // Support 256 colors
			"COLORTERM=truecolor", // Support 24-bit true color
			"FORCE_COLOR=1",       // Generic force color
			"CLICOLOR_FORCE=1",    // BSD apps
			"CLICOLOR=1",          // BSD apps
			"NO_COLOR=",           // Clear NO_COLOR
			"COLUMNS=120",         // Set terminal width
			"LINES=40",            // Set terminal height
		)

		logMessage(styles.infoBox, "EXECUTING", config.ExecuteCommand)

		// Try script first as it's most reliable for TTY emulation
		cmd := exec.Command("script", "-qfec", config.ExecuteCommand, "/dev/null")
		cmd.Env = env
		cmd.Dir, _ = os.Getwd() // Set working directory to current directory

		// Create pipes for stdout
		r, w := io.Pipe()
		bufR := bufio.NewReaderSize(r, 32*1024) // 32KB buffer
		cmd.Stdout = w
		cmd.Stderr = w

		// Start the command
		if err := cmd.Start(); err != nil {
			// If script fails, try stdbuf
			cmd = exec.Command("stdbuf", "-o0", "-e0", "sh", "-c", config.ExecuteCommand)
			cmd.Stdout = w
			cmd.Stderr = w
			cmd.Env = env

			if err := cmd.Start(); err != nil {
				// If stdbuf fails, try unbuffer
				cmd = exec.Command("unbuffer", "-p", config.ExecuteCommand)
				cmd.Stdout = w
				cmd.Stderr = w
				cmd.Env = env

				if err := cmd.Start(); err != nil {
					// Last resort: direct execution
					cmd = exec.Command("sh", "-c", config.ExecuteCommand)
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Env = env

					if err := cmd.Start(); err != nil {
						return fmt.Errorf("failed to start command: %v", err)
					}
				}
			}
		}

		// Copy output in a goroutine
		doneChan := make(chan struct{})
		go func() {
			defer w.Close()
			defer close(doneChan)
			cmd.Wait()
		}()

		// Read the output
		if _, err := io.Copy(&stdout, bufR); err != nil {
			return fmt.Errorf("failed to read command output: %v", err)
		}

		<-doneChan

		code = stdout.String()
		config.NoLineNumbers = true
	case config.FromClipboard:
		// Read from clipboard
		code, err = clipboard.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read from clipboard: %v", err)
		}
	case len(args) > 0:
		// Read from file
		content, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
		code = string(content)
	default:
		// Read from stdin
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
		code = string(content)
	}

	// Create canvas
	canvas := render.NewCanvas()

	// Set window chrome
	themeVariant := chrome.ThemeVariantLight
	if config.DarkMode {
		themeVariant = chrome.ThemeVariantDark
	}

	if config.NoWindowControls {
		window := chrome.NewBlankChrome().
			SetCornerRadius(config.WindowCornerRadius)
		canvas.SetChrome(window)
	} else {
		var window chrome.Chrome
		switch config.WindowChrome {
		case "mac":
			window = chrome.NewMacChrome(chrome.MacStyleSequoia)
		case "windows":
			window = chrome.NewWindowsChrome(chrome.WindowsStyleWin11)
		case "gnome":
			window = chrome.NewGNOMEChrome(chrome.GNOMEStyleAdwaita)
		default:
			return fmt.Errorf("invalid chrome style: %s", config.WindowChrome)
		}

		if config.ChromeThemeName == "" {
			window = window.SetVariant(themeVariant)
		} else {
			window = window.SetThemeByName(config.ChromeThemeName, themeVariant)
		}

		if config.AutoTitle {
			if len(args) > 0 {
				window = window.SetTitle(filepath.Base(args[0]))
			} else if config.ExecuteCommand != "" {
				window = window.SetTitle(config.ExecuteCommand)
			}
		} else {
			window = window.SetTitle(config.WindowTitle)
		}

		window = window.SetCornerRadius(config.WindowCornerRadius)
		canvas.SetChrome(window)
	}

	// Set background
	var bg background.Background
	if config.BackgroundImage != "" {
		file, err := os.Open(config.BackgroundImage)
		if err != nil {
			return fmt.Errorf("failed to open background image: %v", err)
		}
		defer file.Close()
		backgroundImage, _, err := image.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode background image: %v", err)
		}
		var fit background.ImageScaleMode
		switch config.BackgroundImageFit {
		case "fit":
			fit = background.ImageScaleFit
		case "cover":
			fit = background.ImageScaleCover
		case "fill":
			fit = background.ImageScaleFill
		case "stretch":
			fit = background.ImageScaleStretch
		case "tile":
			fit = background.ImageScaleTile
		default:
			return fmt.Errorf("invalid background image fit mode: %s", config.BackgroundImageFit)
		}
		bg = background.
			NewImageBackground(backgroundImage).
			SetScaleMode(fit).
			SetPaddingDetailed(config.PadHoriz, config.PadVert, config.PadHoriz, config.PadVert)
	} else if config.GradientType != "" {
		stops, err := parseGradientStops(config.GradientStops)
		if err != nil {
			return fmt.Errorf("invalid gradient stops: %v", err)
		}

		var gradient background.GradientType
		switch config.GradientType {
		case "linear":
			gradient = background.LinearGradient
		case "radial":
			gradient = background.RadialGradient
		case "angular":
			gradient = background.AngularGradient
		case "diamond":
			gradient = background.DiamondGradient
		case "spiral":
			gradient = background.SpiralGradient
		case "square":
			gradient = background.SquareGradient
		case "star":
			gradient = background.StarGradient
		default:
			return fmt.Errorf("invalid gradient type: %s", config.GradientType)
		}

		bg = background.NewGradientBackground(gradient, stops...).
			SetAngle(config.GradientAngle).
			SetCenter(config.GradientCenterX, config.GradientCenterY).
			SetIntensity(config.GradientIntensity).
			SetCenter(config.GradientCenterX, config.GradientCenterY).
			SetPaddingDetailed(config.PadHoriz, config.PadVert, config.PadHoriz, config.PadVert)
	} else if config.BackgroundColor != "" {
		// Parse background color
		var bgColor color.Color
		if config.BackgroundColor == "transparent" {
			bgColor = color.Transparent
		} else {
			bgColor, err = parseHexColor(config.BackgroundColor)
		}

		if err != nil {
			return fmt.Errorf("invalid background color: %v", err)
		}

		bg = background.NewColorBackground().
			SetColor(bgColor).
			SetPaddingDetailed(config.PadHoriz, config.PadVert, config.PadHoriz, config.PadVert)
	}

	if bg != nil {
		// Configure shadow if enabled
		if config.ShadowBlurRadius > 0 {
			shadowCol, err := parseHexColor(config.ShadowColor)
			if err != nil {
				return fmt.Errorf("invalid shadow color: %v", err)
			}
			bg = bg.SetShadow(background.NewShadow().
				SetBlur(config.ShadowBlurRadius).
				SetOffset(config.ShadowOffsetX, config.ShadowOffsetY).
				SetColor(shadowCol).
				SetSpread(config.ShadowSpread))
		}

		// Configure corner radius
		if config.CornerRadius > 0 {
			bg = bg.SetCornerRadius(config.CornerRadius)
		}

		// Set background
		canvas.SetBackground(bg)
	}

	// Configure highlighted lines
	highlightedLines := []render.LineRange{}
	if config.HighlightLines != "" {
		lines, err := parseHighlightLines(config.HighlightLines)
		if err != nil {
			return fmt.Errorf("invalid highlight lines: %v", err)
		}
		for _, line := range lines {
			highlightedLines = append(highlightedLines, render.LineRange{Start: line, End: line})
		}
	}

	// Get font
	fontSize := 14.0
	var requestedFont *fonts.Font
	if config.Font != "" {
		var fontStr string
		fontStr, fontSize = parseFonts(config.Font)
		if fontStr != "" {
			requestedFont, err = fonts.GetFont(fontStr, nil)
			if err != nil {
				return fmt.Errorf("failed to get font: %v", err)
			}
		}
	}

	// Configure code style
	canvas.SetCodeStyle(&render.CodeStyle{
		UseANSI:         config.ExecuteCommand != "",
		ShowPrompt:      config.ShowPrompt,
		PromptCommand:   config.ExecuteCommand,
		Language:        config.Language,
		Theme:           strings.ToLower(config.Theme),
		FontFamily:      requestedFont,
		FontSize:        fontSize,
		TabWidth:        config.TabWidth,
		PaddingLeft:     config.CodePadLeft,
		PaddingRight:    config.CodePadRight,
		PaddingTop:      config.CodePadTop,
		PaddingBottom:   config.CodePadBottom,
		ShowLineNumbers: !config.NoLineNumbers,
		LineNumberRange: render.LineRange{
			Start: config.StartLine,
			End:   config.EndLine,
		},
		LineHighlightRanges: highlightedLines,
	})

	// Render the image
	img, err := canvas.RenderToImage(code)
	if err != nil {
		return fmt.Errorf("failed to render image: %v", err)
	}

	if config.ToClipboard || config.ToStdout {
		pngBuf := bytes.NewBuffer(nil)
		if err := png.Encode(pngBuf, img); err != nil {
			return fmt.Errorf("failed to encode image to png: %v", err)
		}

		if config.ToClipboard {
			err := clipboard.WriteAll(pngBuf.String())
			if err != nil {
				return fmt.Errorf("failed to copy image to clipboard: %v", err)
			}

			if echo {
				logMessage(styles.successBox, "COPIED", "to clipboard")
			}
		}

		if config.ToStdout {
			_, err := os.Stdout.Write(pngBuf.Bytes())
			if err != nil {
				return fmt.Errorf("failed to write image to stdout: %v", err)
			}
		}
		return nil
	}

	err = saveImage(img, config)
	if err == nil {
		if echo {
			logMessage(styles.successBox, "WROTE", config.OutputFile)
		}
	} else {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}

func saveImage(img image.Image, config *Config) error {
	// If no output file is specified, use png as default
	if config.OutputFile == "" {
		config.OutputFile = "output.png"
	}

	// Get the extension from the filename
	ext := strings.ToLower(filepath.Ext(config.OutputFile))
	if ext == "" {
		ext = ".png"
		config.OutputFile += ext
	}

	// Save in the format matching the extension
	switch ext {
	case ".png":
		return render.SaveAsPNG(img, config.OutputFile)
	case ".jpg", ".jpeg":
		return render.SaveAsJPEG(img, config.OutputFile)
	case ".bmp":
		return render.SaveAsBMP(img, config.OutputFile)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}
