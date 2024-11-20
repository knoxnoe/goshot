package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
	"github.com/watzon/goshot/pkg/syntax"
)

var (
	// Output options
	outputFile    string
	toClipboard   bool
	fromClipboard bool

	// Appearance
	windowChrome       string
	darkMode           bool
	theme              string
	language           string
	font               string
	backgroundColor    string
	backgroundImage    string
	backgroundImageFit string
	showLineNumbers    bool
	cornerRadius       float64
	windowControls     bool
	windowTitle        string
	windowCornerRadius float64

	// Padding and layout
	tabWidth     int
	startLine    int
	endLine      int
	linePadding  int
	padHoriz     int
	padVert      int
	codePadVert  int
	codePadHoriz int

	// Shadow options
	shadowBlurRadius float64
	shadowColor      string
	shadowOffsetX    float64
	shadowOffsetY    float64

	// Highlighting
	highlightLines string
)

var rootCmd = &cobra.Command{
	Use:   "goshot [FLAGS] [OPTIONS] [FILE]",
	Short: "Create beautiful code screenshots",
	Long: `Goshot is a powerful tool for creating beautiful code screenshots with
customizable window chrome, syntax highlighting, and backgrounds.`,
}

func init() {
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render the code to an image",
		Run:   renderImage,
	}

	// Output flags
	renderCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output image to specific location instead of cwd")
	renderCmd.Flags().BoolVarP(&toClipboard, "to-clipboard", "c", false, "Copy the output image to clipboard")
	renderCmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read input from clipboard")

	// Appearance flags
	renderCmd.Flags().StringVarP(&windowChrome, "chrome", "C", "macos", "Chrome style. Available styles: macos, windows11")
	renderCmd.Flags().BoolVarP(&darkMode, "dark-mode", "d", false, "Use dark mode")
	renderCmd.Flags().StringVarP(&theme, "theme", "t", "dracula", "The syntax highlight theme. It can be a theme name or path to a .tmTheme file")
	renderCmd.Flags().StringVarP(&language, "language", "l", "", "The language for syntax highlighting. You can use full name (\"Rust\") or file extension (\"rs\")")
	renderCmd.Flags().StringVarP(&font, "font", "f", "", "The fallback font list. eg. 'Hack; SimSun=31'")
	renderCmd.Flags().StringVarP(&backgroundColor, "background", "b", "#aaaaff", "Background color of the image")
	renderCmd.Flags().StringVar(&backgroundImage, "background-image", "", "Background image")
	renderCmd.Flags().StringVar(&backgroundImageFit, "background-image-fit", "cover", "Background image fit mode. Available modes: contain, cover, fill, stretch, tile")
	renderCmd.Flags().BoolVar(&showLineNumbers, "no-line-number", false, "Hide the line number")
	renderCmd.Flags().Float64Var(&cornerRadius, "corner-radius", 10.0, "Corner radius of the image")
	renderCmd.Flags().BoolVar(&windowControls, "no-window-controls", false, "Hide the window controls")
	renderCmd.Flags().StringVar(&windowTitle, "window-title", "", "Show window title")
	renderCmd.Flags().Float64Var(&windowCornerRadius, "window-corner-radius", 10, "Corner radius of the window")

	// Padding and layout flags
	renderCmd.Flags().IntVar(&tabWidth, "tab-width", 4, "Tab width")
	renderCmd.Flags().IntVar(&startLine, "start-line", 1, "Line to start from")
	renderCmd.Flags().IntVar(&endLine, "end-line", 0, "Line to end at")
	renderCmd.Flags().IntVar(&linePadding, "line-pad", 2, "Pad between lines")
	renderCmd.Flags().IntVar(&padHoriz, "pad-horiz", 80, "Pad horiz")
	renderCmd.Flags().IntVar(&padVert, "pad-vert", 100, "Pad vert")
	renderCmd.Flags().IntVar(&codePadVert, "code-pad-vert", 10, "Add padding to the X axis of the code")
	renderCmd.Flags().IntVar(&codePadHoriz, "code-pad-horiz", 10, "Add padding to the Y axis of the code")

	// Shadow flags
	renderCmd.Flags().Float64Var(&shadowBlurRadius, "shadow-blur-radius", 0, "Blur radius of the shadow. (set it to 0 to hide shadow)")
	renderCmd.Flags().StringVar(&shadowColor, "shadow-color", "#555555", "Color of shadow")
	renderCmd.Flags().Float64Var(&shadowOffsetX, "shadow-offset-x", 0, "Shadow's offset in X axis")
	renderCmd.Flags().Float64Var(&shadowOffsetY, "shadow-offset-y", 0, "Shadow's offset in Y axis")

	// Highlighting flags
	renderCmd.Flags().StringVar(&highlightLines, "highlight-lines", "", "Lines to highlight. eg. '1-3;4'")

	// Additional utility commands
	rootCmd.AddCommand(
		renderCmd,
		&cobra.Command{
			Use:   "list-themes",
			Short: "List all available themes",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Available themes:")
				styles := syntax.GetAvailableStyles()
				for _, style := range styles {
					fmt.Printf("  %s\n", style)
				}
			},
		},
		&cobra.Command{
			Use:   "list-languages",
			Short: "List all available languages",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Available languages:")
				languages := syntax.GetAvailableLanguages(false)
				for _, lang := range languages {
					fmt.Printf("  %s\n", lang)
				}
			},
		},
		&cobra.Command{
			Use:   "list-fonts",
			Short: "List all available fonts in your system",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Available fonts:")
				fonts := fonts.ListFonts()
				for _, font := range fonts {
					fmt.Printf("  %s\n", font)
				}
			},
		},
		&cobra.Command{
			Use:   "config-file",
			Short: "Show the path of goshot config file",
			Run: func(cmd *cobra.Command, args []string) {
				// Implementation will go here
			},
		},
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func renderImage(cmd *cobra.Command, args []string) {
	// Get the input code
	var code string
	var err error

	switch {
	case fromClipboard:
		// Read from clipboard
		code, err = clipboard.ReadAll()
		if err != nil {
			fmt.Printf("failed to read from clipboard: %v", err)
			return
		}
	case len(args) > 0:
		// Read from file
		content, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("failed to read file: %v", err)
			return
		}
		code = string(content)
	default:
		// Read from stdin
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("failed to read from stdin: %v", err)
			return
		}
		code = string(content)
	}

	// Parse background color
	bgColor, err := parseHexColor(backgroundColor)
	if err != nil {
		fmt.Printf("invalid background color: %v", err)
		return
	}

	// Create canvas
	canvas := render.NewCanvas()

	// Set window chrome
	if !windowControls {
		switch windowChrome {
		case "macos":
			canvas.SetChrome(
				chrome.NewMacOSChrome(chrome.WithDarkMode(darkMode),
					chrome.WithTitle(windowTitle),
					chrome.WithCornerRadius(windowCornerRadius)),
			)
		case "windows11":
			canvas.SetChrome(
				chrome.NewWindows11Chrome(chrome.WithDarkMode(darkMode),
					chrome.WithTitle(windowTitle),
					chrome.WithCornerRadius(windowCornerRadius)),
			)
		default:
			fmt.Printf("invalid chrome style: %s", windowChrome)
			return
		}
	}

	// Set background
	var bg background.Background
	if backgroundImage != "" {
		file, err := os.Open(backgroundImage)
		if err != nil {
			fmt.Printf("failed to open background image: %v", err)
			return
		}
		defer file.Close()
		backgroundImage, _, err := image.Decode(file)
		if err != nil {
			fmt.Printf("failed to decode background image: %v", err)
			return
		}
		var fit background.ImageScaleMode
		switch backgroundImageFit {
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
			fmt.Printf("invalid background image fit mode: %s", backgroundImageFit)
			return
		}
		bg = background.NewImageBackground(backgroundImage).SetScaleMode(fit)
	} else {
		bg = background.NewColorBackground().
			SetColor(bgColor).
			SetPaddingDetailed(padHoriz, padVert, padVert, padHoriz)
	}

	// Configure shadow if enabled
	if shadowBlurRadius > 0 {
		shadowCol, err := parseHexColor(shadowColor)
		if err != nil {
			fmt.Printf("invalid shadow color: %v", err)
			return
		}
		bg = bg.SetShadow(background.NewShadow().
			SetBlur(shadowBlurRadius).
			SetOffset(shadowOffsetX, shadowOffsetY).
			SetColor(shadowCol))
	}

	if cornerRadius > 0 {
		bg = bg.SetCornerRadius(cornerRadius)
	}

	canvas.SetBackground(bg)

	// Configure highlighted lines
	highlightedLines := []render.LineRange{}
	if highlightLines != "" {
		lines, err := parseHighlightLines(highlightLines)
		if err != nil {
			fmt.Printf("invalid highlight lines: %v", err)
			return
		}
		for _, line := range lines {
			highlightedLines = append(highlightedLines, render.LineRange{Start: line, End: line})
		}
	}

	// Get font
	fontSize := 14.0
	var requestedFont *fonts.Font
	if font != "" {
		var fontStr string
		fontStr, fontSize = parseFonts(font)
		if fontStr != "" {
			requestedFont, err = fonts.GetFont(fontStr, nil)
			if err != nil {
				fmt.Printf("failed to get font: %v", err)
				return
			}
		}
	}

	// Configure code style
	canvas.SetCodeStyle(&render.CodeStyle{
		Language:        language,
		Theme:           strings.ToLower(theme),
		FontFamily:      requestedFont,
		FontSize:        fontSize,
		TabWidth:        tabWidth,
		PaddingX:        codePadHoriz,
		PaddingY:        codePadVert,
		ShowLineNumbers: !showLineNumbers,
		LineNumberRange: render.LineRange{
			Start: startLine,
			End:   endLine,
		},
		LineHighlightRanges: highlightedLines,
	})

	// Render the image
	img, err := canvas.RenderToImage(code)
	if err != nil {
		fmt.Printf("failed to render image: %v", err)
		return
	}

	if outputFile != "" {
		if err := render.SaveAsPNG(img, outputFile); err != nil {
			fmt.Printf("failed to save image: %v", err)
			return
		}
	}

	if toClipboard {
		pngBuf := bytes.NewBuffer(nil)
		if err := png.Encode(pngBuf, img); err != nil {
			fmt.Printf("failed to encode image to png: %v", err)
			return
		}
		clipboard.WriteAll(pngBuf.String())
	}
}

// Helper functions
func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return nil, err
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func parseHighlightLines(input string) ([]int, error) {
	var result []int
	parts := strings.Split(input, ";")

	for _, part := range parts {
		if strings.Contains(part, "-") {
			// Handle range (e.g., "1-3")
			var start, end int
			if _, err := fmt.Sscanf(part, "%d-%d", &start, &end); err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			// Handle single line
			var line int
			if _, err := fmt.Sscanf(part, "%d", &line); err != nil {
				return nil, err
			}
			result = append(result, line)
		}
	}

	return result, nil
}

// parseFonts takes in a string of fonts and returns the first font
// that is available on the system.
// Ex. "JetBrains Mono; DejaVu Sans=30"
func parseFonts(input string) (string, float64) {
	for _, fontSpec := range strings.Split(input, ";") {
		parts := strings.Split(strings.TrimSpace(fontSpec), "=")
		fontName := strings.TrimSpace(parts[0])
		fontSize := 14.0
		if len(parts) > 1 {
			if parsedSize, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
				fontSize = parsedSize
			}
		}

		if fonts.IsFontAvailable(fontName) {
			return fontName, fontSize
		}
	}

	return "", 14.0
}
