package main

import (
	"fmt"
	"image/color"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	// Sample code to highlight
	sampleCode := `package syntax

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

// RenderConfig holds configuration for rendering highlighted code to an image
type RenderConfig struct {
	FontSize   float64
	LineHeight float64
	PaddingX   int
	PaddingY   int
	FontFamily *truetype.Font
	Background image.Image
	TabWidth   int // Width of tab characters in spaces
	MinWidth   int // Minimum width in pixels (0 means no minimum)
	MaxWidth   int // Maximum width in pixels (0 means no limit)

	// Line number settings
	ShowLineNumbers   bool
	LineNumberColor   color.Color
	LineNumberPadding int         // Padding on either side of line numbers in pixels
	LineNumberBg      color.Color // Background color for line numbers
	StartLineNumber   int         // Line number to start from
	EndLineNumber     int         // Line number to end at

	// Line highlighting settings
	LineHighlightColor color.Color // Color for highlighted lines
}

// DefaultConfig returns a default rendering configuration
func DefaultConfig() *RenderConfig {
	f, _ := truetype.Parse(gomono.TTF)
	return &RenderConfig{
		FontSize:   14,
		LineHeight: 1.5,
		PaddingX:   10,
		PaddingY:   10,
		FontFamily: f,
		TabWidth:   4,    // Default 4 spaces per tab
		MinWidth:   200,  // Minimum width of 200px
		MaxWidth:   1460, // Maximum width for 120 characters

		// Line number defaults
		ShowLineNumbers:   true,
		LineNumberColor:   color.RGBA{R: 128, G: 128, B: 128, A: 255}, // Gray color
		LineNumberPadding: 10,
		LineNumberBg:      color.RGBA{R: 245, G: 245, B: 245, A: 255}, // Light gray background
		StartLineNumber:   1,
		EndLineNumber:     0,

		// Line highlighting defaults
		LineHighlightColor: color.RGBA{R: 68, G: 68, B: 68, A: 40}, // Semi-transparent dark color
	}
}`

	// Sample configurations
	samples := []struct {
		name   string
		canvas *render.Canvas
	}{
		{
			name: "dracula",
			canvas: render.NewCanvas().
				SetChrome(chrome.NewWindows11Chrome(chrome.WithTitle("My App"))).
				SetBackground(
					background.NewColorBackground().
						SetColor(color.RGBA{R: 25, G: 25, B: 25, A: 255}).
						SetPadding(40),
				).
				SetCodeStyle(&render.CodeStyle{
					Language:            "go",
					Theme:               "dracula",
					TabWidth:            4,
					ShowLineNumbers:     true,
					LineHighlightRanges: []render.LineRange{{Start: 18, End: 26}},
				}),
		},
		{
			name: "catppuccin-mocha",
			canvas: render.NewCanvas().
				SetChrome(chrome.NewWindows11Chrome(
					chrome.WithTitle("My App"),
					chrome.WithDarkMode(true),
				)).
				SetBackground(
					background.NewColorBackground().
						SetColor(color.RGBA{R: 120, G: 120, B: 120, A: 255}).
						SetPadding(40),
				).
				SetCodeStyle(&render.CodeStyle{
					Language:            "go",
					Theme:               "catppuccin-mocha",
					TabWidth:            4,
					ShowLineNumbers:     false,
					LineHighlightRanges: []render.LineRange{{Start: 18, End: 26}},
				}),
		},
		{
			name: "catppuccin-latte",
			canvas: render.NewCanvas().
				SetChrome(chrome.NewWindows11Chrome(
					chrome.WithTitle("My App"),
					chrome.WithTitleBar(false),
					chrome.WithCornerRadius(10),
				)).
				SetCodeStyle(&render.CodeStyle{
					Language:            "go",
					Theme:               "catppuccin-latte",
					TabWidth:            4,
					ShowLineNumbers:     true,
					LineHighlightRanges: []render.LineRange{{Start: 18, End: 26}},
				}),
		},
		{
			name: "gruvbox",
			canvas: render.NewCanvas().
				SetChrome(chrome.NewWindows11Chrome(
					chrome.WithTitle("My App"),
					chrome.WithDarkMode(true),
					chrome.WithTitleBar(false),
				)).
				SetBackground(
					background.NewColorBackground().
						SetColor(color.RGBA{R: 70, G: 70, B: 70, A: 255}).
						SetPadding(40),
				).
				SetCodeStyle(&render.CodeStyle{
					Language:            "go",
					Theme:               "gruvbox",
					TabWidth:            4,
					ShowLineNumbers:     false,
					LineHighlightRanges: []render.LineRange{{Start: 18, End: 26}},
				}),
		},
	}

	// Process each sample
	for _, sample := range samples {
		fmt.Printf("Processing %s style...\n", sample.name)

		// Render the code to an image
		result, err := sample.canvas.RenderCode(sampleCode)
		if err != nil {
			fmt.Printf("Error rendering code: %v\n", err)
			continue
		}

		// Save the image
		err = render.SaveAsPNG(result, fmt.Sprintf("output-%s.png", sample.name))
		if err != nil {
			fmt.Printf("Error saving image: %v\n", err)
			continue
		}
	}
}
