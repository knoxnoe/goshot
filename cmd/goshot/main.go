package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
	"github.com/watzon/goshot/pkg/syntax"
	"github.com/watzon/goshot/pkg/version"
)

type Config struct {
	// Interactive mode
	Interactive bool
	Input       string

	// Output options
	OutputFile    string
	ToClipboard   bool
	FromClipboard bool
	ToStdout      bool

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
	ShowLineNumbers    bool
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
		Use:   "goshot [flags] [file]",
		Short: "Goshot is a powerful tool for creating beautiful code screenshots with customizable window chrome, syntax highlighting, and backgrounds.",
		Run: func(cmd *cobra.Command, args []string) {
			if config.Interactive {
				fmt.Println("Interactive mode is coming soon!")
				os.Exit(1)
			}

			if err := renderImage(&config, args); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	// Interactive mode
	rootCmd.Flags().BoolVarP(&config.Interactive, "interactive", "i", false, "Interactive mode")

	// Output flags
	rootCmd.Flags().StringVarP(&config.OutputFile, "output", "o", "output.png", "Write output image to specific location instead of cwd")
	rootCmd.Flags().BoolVarP(&config.ToClipboard, "to-clipboard", "c", false, "Copy the output image to clipboard")
	rootCmd.Flags().BoolVar(&config.FromClipboard, "from-clipboard", false, "Read input from clipboard")
	rootCmd.Flags().BoolVarP(&config.ToStdout, "to-stdout", "s", false, "Write output to stdout")

	// Appearance flags
	rootCmd.Flags().StringVarP(&config.WindowChrome, "chrome", "C", "mac", "Chrome style. Available styles: mac, windows, gnome")
	rootCmd.Flags().StringVarP(&config.ChromeThemeName, "chrome-theme", "T", "", "Chrome theme name")
	rootCmd.Flags().BoolVarP(&config.DarkMode, "dark-mode", "d", false, "Use dark mode")
	rootCmd.Flags().StringVarP(&config.Theme, "theme", "t", "dracula", "The syntax highlight theme. It can be a theme name or path to a .tmTheme file")
	rootCmd.Flags().StringVarP(&config.Language, "language", "l", "", "The language for syntax highlighting. You can use full name (\"Rust\") or file extension (\"rs\")")
	rootCmd.Flags().StringVarP(&config.Font, "font", "f", "", "The fallback font list. eg. 'Hack; SimSun=31'")
	rootCmd.Flags().StringVarP(&config.BackgroundColor, "background", "b", "#aaaaff", "Background color of the image")
	rootCmd.Flags().StringVar(&config.BackgroundImage, "background-image", "", "Background image")
	rootCmd.Flags().StringVar(&config.BackgroundImageFit, "background-image-fit", "cover", "Background image fit mode. Available modes: contain, cover, fill, stretch, tile")
	rootCmd.Flags().BoolVar(&config.ShowLineNumbers, "no-line-number", false, "Hide the line number")
	rootCmd.Flags().Float64Var(&config.CornerRadius, "corner-radius", 10.0, "Corner radius of the image")
	rootCmd.Flags().BoolVar(&config.NoWindowControls, "no-window-controls", false, "Hide the window controls")
	rootCmd.Flags().StringVar(&config.WindowTitle, "window-title", "", "Show window title")
	rootCmd.Flags().Float64Var(&config.WindowCornerRadius, "window-corner-radius", 10, "Corner radius of the window")

	// Gradient flags
	rootCmd.Flags().StringVar(&config.GradientType, "gradient-type", "", "Gradient type. Available types: linear, radial, angular, diamond, spiral, square, star")
	rootCmd.Flags().StringArrayVar(&config.GradientStops, "gradient-stop", []string{"#232323;0", "#383838;100"}, "Gradient stops. eg. '--gradient-stop '#ff0000;0' --gradient-stop '#00ff00;100'")
	rootCmd.Flags().Float64Var(&config.GradientAngle, "gradient-angle", 45, "Gradient angle in degrees")
	rootCmd.Flags().Float64Var(&config.GradientCenterX, "gradient-center-x", 0.5, "Center X of the gradient")
	rootCmd.Flags().Float64Var(&config.GradientCenterY, "gradient-center-y", 0.5, "Center Y of the gradient")
	rootCmd.Flags().Float64Var(&config.GradientIntensity, "gradient-intensity", 5, "Intensity modifier for special gradients")

	// Padding and layout flags
	rootCmd.Flags().IntVar(&config.TabWidth, "tab-width", 4, "Tab width")
	rootCmd.Flags().IntVar(&config.StartLine, "start-line", 1, "Line to start from")
	rootCmd.Flags().IntVar(&config.EndLine, "end-line", 0, "Line to end at")
	rootCmd.Flags().IntVar(&config.LinePadding, "line-pad", 2, "Pad between lines")
	rootCmd.Flags().IntVar(&config.PadHoriz, "pad-horiz", 80, "Pad horiz")
	rootCmd.Flags().IntVar(&config.PadVert, "pad-vert", 100, "Pad vert")
	rootCmd.Flags().IntVar(&config.CodePadTop, "code-pad-top", 10, "Add padding to the top of the code")
	rootCmd.Flags().IntVar(&config.CodePadBottom, "code-pad-bottom", 10, "Add padding to the bottom of the code")
	rootCmd.Flags().IntVar(&config.CodePadLeft, "code-pad-left", 10, "Add padding to the X axis of the code")
	rootCmd.Flags().IntVar(&config.CodePadRight, "code-pad-right", 10, "Add padding to the X axis of the code")

	// Shadow flags
	rootCmd.Flags().Float64Var(&config.ShadowBlurRadius, "shadow-blur", 0, "Blur radius of the shadow. (set it to 0 to hide shadow)")
	rootCmd.Flags().StringVar(&config.ShadowColor, "shadow-color", "#00000033", "Color of shadow")
	rootCmd.Flags().Float64Var(&config.ShadowSpread, "shadow-spread", 0, "Spread radius of the shadow")
	rootCmd.Flags().Float64Var(&config.ShadowOffsetX, "shadow-offset-x", 0, "Shadow's offset in X axis")
	rootCmd.Flags().Float64Var(&config.ShadowOffsetY, "shadow-offset-y", 0, "Shadow's offset in Y axis")

	// Highlighting flags
	rootCmd.Flags().StringVar(&config.HighlightLines, "highlight-lines", "", "Lines to highlight. eg. '1-3;4'")

	// Additional utility commands
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "themes",
			Short: "List available themes",
			Run: func(cmd *cobra.Command, args []string) {
				themes := syntax.GetAvailableStyles()
				for _, theme := range themes {
					fmt.Println(theme)
				}
			},
		},
		&cobra.Command{
			Use:   "languages",
			Short: "List available languages",
			Run: func(cmd *cobra.Command, args []string) {
				languages := syntax.GetAvailableLanguages(false)
				for _, lang := range languages {
					fmt.Println(lang)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func renderImage(config *Config, args []string) error {
	// Get the input code
	var code string
	var err error

	switch {
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

		window = window.SetTitle(config.WindowTitle).SetCornerRadius(config.WindowCornerRadius)
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
		Language:        config.Language,
		Theme:           strings.ToLower(config.Theme),
		FontFamily:      requestedFont,
		FontSize:        fontSize,
		TabWidth:        config.TabWidth,
		PaddingLeft:     config.CodePadLeft,
		PaddingRight:    config.CodePadRight,
		PaddingTop:      config.CodePadTop,
		PaddingBottom:   config.CodePadBottom,
		ShowLineNumbers: !config.ShowLineNumbers,
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
		}

		if config.ToStdout {
			_, err := os.Stdout.Write(pngBuf.Bytes())
			if err != nil {
				return fmt.Errorf("failed to write image to stdout: %v", err)
			}
		}
		return nil
	}

	if err := render.SaveAsPNG(img, config.OutputFile); err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}
