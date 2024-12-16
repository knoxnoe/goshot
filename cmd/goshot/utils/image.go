package utils

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/atotto/clipboard"
	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/content/code"
	content_term "github.com/watzon/goshot/content/term"
	"github.com/watzon/goshot/fonts"
	"github.com/watzon/goshot/render"
)

// RenderCode renders code content to an image with the given configuration
func RenderCode(cfg *config.Config, echo bool, input string) error {
	canvas, err := makeCanvas(cfg, []string{})
	if err != nil {
		return err
	}

	// Get font
	fontSize := 14.0
	var requestedFont *fonts.Font
	if cfg.Font == "" {
		requestedFont, err = fonts.GetFallback(fonts.FallbackMono)
		if err != nil {
			return err
		}
	} else {
		var fontStr string
		fontStr, fontSize = ParseFonts(cfg.Font)
		if fontStr == "" {
			return fmt.Errorf("invalid font: %s", cfg.Font)
		} else {
			requestedFont, err = fonts.GetFont(fontStr, nil)
			if err != nil {
				return err
			}
		}
	}

	// Configure content
	content := code.DefaultRenderer(input).
		WithLanguage(DetectLanguage(cfg.Input, cfg.Language)).
		WithTheme(cfg.Theme).
		WithFontSize(fontSize).
		WithLineHeight(cfg.LineHeight).
		WithPadding(cfg.CodePadLeft, cfg.CodePadRight, cfg.CodePadTop, cfg.CodePadBottom).
		WithLineNumberPadding(cfg.LineNumberPadding).
		WithTabWidth(cfg.TabWidth).
		WithMinWidth(cfg.MinWidth).
		WithMaxWidth(cfg.MaxWidth).
		WithLineNumbers(!cfg.NoLineNumbers).
		WithFont(requestedFont)

	// Configure redaction if enabled
	if cfg.RedactionEnabled {
		content.WithRedactionEnabled(true).
			WithRedactionBlurRadius(cfg.RedactionBlurRadius)

		// Set redaction style
		var style code.RedactionStyle
		switch cfg.RedactionStyle {
		case "blur":
			style = code.RedactionStyleBlur
		case "block":
			style = code.RedactionStyleBlock
		default:
			return fmt.Errorf("invalid redaction style: %s (must be 'block' or 'blur')", cfg.RedactionStyle)
		}
		content.WithRedactionStyle(style)

		// Add custom redaction patterns
		if len(cfg.RedactionPatterns) > 0 {
			for _, pattern := range cfg.RedactionPatterns {
				content.WithRedactionPattern(pattern, "Custom Pattern")
			}
		}

		// Add manual redaction areas
		if len(cfg.RedactionAreas) > 0 {
			areas, err := ParseRedactionAreas(cfg.RedactionAreas)
			if err != nil {
				return err
			}
			for _, area := range areas {
				content.WithManualRedaction(area.X, area.Y, area.Width, area.Height)
			}
		}
	}

	// Configure highlighted lines
	highlightedLines, err := ParseLineRanges(cfg.HighlightLines)
	if err != nil {
		return err
	}
	for _, lr := range highlightedLines {
		content.WithLineHighlightRange(lr.Start, lr.End)
	}

	// Configure line ranges
	lineRanges, err := ParseLineRanges(cfg.LineRanges)
	if err != nil {
		return err
	}
	for _, lr := range lineRanges {
		content.WithLineRange(lr.Start, lr.End)
	}

	canvas.WithContent(content)

	return renderAndSave(canvas, cfg, echo)
}

// RenderTerm renders terminal content to an image with the given configuration
func RenderTerm(cfg *config.Config, echo bool, args []string, input []byte) error {
	canvas, err := makeCanvas(cfg, args)
	if err != nil {
		return err
	}

	// Get font
	fontSize := 14.0
	var requestedFont *fonts.Font
	if cfg.Font == "" {
		requestedFont, err = fonts.GetFallback(fonts.FallbackMono)
		if err != nil {
			return err
		}
	} else {
		var fontStr string
		fontStr, fontSize = ParseFonts(cfg.Font)
		if fontStr == "" {
			return fmt.Errorf("invalid font: %s", cfg.Font)
		} else {
			requestedFont, err = fonts.GetFont(fontStr, nil)
			if err != nil {
				return err
			}
		}
	}

	renderer := content_term.NewRenderer(input, &content_term.TermStyle{
		Args:          args,
		Theme:         cfg.Theme,
		Font:          requestedFont,
		FontSize:      fontSize,
		LineHeight:    cfg.LineHeight,
		PaddingLeft:   cfg.CellPadLeft,
		PaddingRight:  cfg.CellPadRight,
		PaddingTop:    cfg.CellPadTop,
		PaddingBottom: cfg.CellPadBottom,
		Width:         cfg.CellWidth,
		Height:        cfg.CellHeight,
		AutoSize:      cfg.AutoSize,
		CellSpacing:   cfg.CellSpacing,
		ShowPrompt:    cfg.ShowPrompt,
		PromptFunc:    NewPromptFunc(cfg.PromptTemplate, cfg),
	})

	canvas.WithContent(renderer)

	return renderAndSave(canvas, cfg, echo)
}

// makeCanvas creates a new canvas with the given configuration
func makeCanvas(cfg *config.Config, args []string) (*render.Canvas, error) {
	var err error

	// Create canvas
	canvas := render.NewCanvas()

	// Set window chrome
	themeVariant := chrome.ThemeVariantDark
	if cfg.LightMode {
		themeVariant = chrome.ThemeVariantLight
	}

	if cfg.NoWindowControls {
		window := chrome.NewBlankChrome().
			WithCornerRadius(cfg.WindowCornerRadius)
		canvas.WithChrome(window)
	} else {
		var window chrome.Chrome
		switch cfg.WindowChrome {
		case "mac":
			window = chrome.NewMacChrome(chrome.MacStyleSequoia)
		case "windows":
			window = chrome.NewWindowsChrome(chrome.WindowsStyleWin11)
		case "gnome":
			window = chrome.NewGNOMEChrome(chrome.GNOMEStyleAdwaita)
		default:
			return nil, fmt.Errorf("invalid chrome style: %s", cfg.WindowChrome)
		}

		if cfg.ChromeThemeName == "" {
			window = window.WithVariant(themeVariant)
		} else {
			window = window.WithThemeByName(cfg.ChromeThemeName, themeVariant)
		}

		if cfg.AutoTitle {
			if len(args) > 0 {
				window = window.WithTitle(args[0])
			}
		} else {
			window = window.WithTitle(cfg.WindowTitle)
		}

		window = window.WithCornerRadius(cfg.WindowCornerRadius)
		canvas.WithChrome(window)
	}

	// Set background
	var bg background.Background
	if cfg.BackgroundImage != "" {
		file, err := os.Open(cfg.BackgroundImage)
		if err != nil {
			return nil, fmt.Errorf("failed to open background image: %v", err)
		}
		defer file.Close()
		backgroundImage, _, err := image.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode background image: %v", err)
		}
		var fit background.ImageScaleMode
		switch cfg.BackgroundImageFit {
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
			return nil, fmt.Errorf("invalid background image fit mode: %s", cfg.BackgroundImageFit)
		}
		bg = background.
			NewImageBackground(backgroundImage).
			WithScaleMode(fit).
			WithPaddingDetailed(cfg.PadVert, cfg.PadHoriz, cfg.PadVert, cfg.PadHoriz)
	} else if cfg.GradientType != "" {
		stops, err := ParseGradientStops(cfg.GradientStops)
		if err != nil {
			return nil, fmt.Errorf("invalid gradient stops: %v", err)
		}

		var gradient background.GradientType
		switch cfg.GradientType {
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
			return nil, fmt.Errorf("invalid gradient type: %s", cfg.GradientType)
		}

		bg = background.NewGradientBackground(gradient, stops...).
			WithAngle(cfg.GradientAngle).
			WithCenter(cfg.GradientCenterX, cfg.GradientCenterY).
			WithIntensity(cfg.GradientIntensity).
			WithCenter(cfg.GradientCenterX, cfg.GradientCenterY).
			WithPaddingDetailed(cfg.PadVert, cfg.PadHoriz, cfg.PadVert, cfg.PadHoriz)
	} else if cfg.BackgroundColor != "" {
		// Parse background color
		var bgColor color.Color
		if cfg.BackgroundColor == "transparent" {
			bgColor = color.Transparent
		} else {
			bgColor, err = ParseHexColor(cfg.BackgroundColor)
		}

		if err != nil {
			return nil, fmt.Errorf("invalid background color: %v", err)
		}

		bg = background.NewColorBackground().
			WithColor(bgColor).
			WithPaddingDetailed(cfg.PadVert, cfg.PadHoriz, cfg.PadVert, cfg.PadHoriz)
	}

	if bg != nil {
		// Configure shadow if enabled
		if cfg.ShadowBlurRadius > 0 {
			shadowCol, err := ParseHexColor(cfg.ShadowColor)
			if err != nil {
				return nil, fmt.Errorf("invalid shadow color: %v", err)
			}
			bg = bg.WithShadow(background.NewShadow().
				WithBlur(cfg.ShadowBlurRadius).
				WithOffset(cfg.ShadowOffsetX, cfg.ShadowOffsetY).
				WithColor(shadowCol).
				WithSpread(cfg.ShadowSpread))
		}

		// Configure corner radius
		if cfg.CornerRadius > 0 {
			bg = bg.WithCornerRadius(cfg.CornerRadius)
		}

		// Set background
		canvas.WithBackground(bg)
	}

	return canvas, nil
}

// renderAndSave renders the canvas to an image and saves it according to the configuration
func renderAndSave(canvas *render.Canvas, cfg *config.Config, echo bool) error {
	// Render to image
	img, err := canvas.RenderToImage()
	if err != nil {
		return err
	}

	// Handle clipboard output
	if cfg.ToClipboard {
		imgBytes, err := ImageToBytes(img)
		if err != nil {
			return err
		}

		if err := clipboard.WriteAll(string(imgBytes)); err != nil {
			return fmt.Errorf("failed to copy image to clipboard: %v", err)
		}

		if echo {
			config.LogMessage(config.Styles.SuccessBox, "COPIED", "to clipboard")
		}
	}

	// Handle stdout output
	if cfg.ToStdout {
		imgBytes, err := ImageToBytes(img)
		if err != nil {
			return err
		}

		if _, err := os.Stdout.Write(imgBytes); err != nil {
			return fmt.Errorf("failed to write image to stdout: %v", err)
		}

		if echo {
			config.LogMessage(config.Styles.SuccessBox, "WROTE", "to stdout")
		}
	}

	// Handle file output
	if cfg.OutputFile != "" {
		outputFile := NewFilenameFunc(cfg.OutputFile, cfg)()
		resolvedFilename, err := SaveImageToFile(img, outputFile)
		if err != nil {
			return fmt.Errorf("failed to save image: %v", err)
		}

		if echo && resolvedFilename != "" {
			config.LogMessage(config.Styles.SuccessBox, "WROTE", resolvedFilename)
		}
	}

	return nil
}
