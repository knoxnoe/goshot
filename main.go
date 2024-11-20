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
					background.NewGradientBackground(
						background.DiamondGradient,
						// Pink purple and yellow
						background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 255, A: 255}, Position: 0},
						background.GradientStop{Color: color.RGBA{R: 128, G: 0, B: 128, A: 255}, Position: 0.1},
						background.GradientStop{Color: color.RGBA{R: 255, G: 255, B: 0, A: 255}, Position: 0.2},
						background.GradientStop{Color: color.RGBA{R: 128, G: 128, B: 0, A: 255}, Position: 0.3},
						background.GradientStop{Color: color.RGBA{R: 0, G: 255, B: 255, A: 255}, Position: 0.4},
						background.GradientStop{Color: color.RGBA{R: 0, G: 128, B: 128, A: 255}, Position: 0.5},
						background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Position: 0.6},
						background.GradientStop{Color: color.RGBA{R: 128, G: 0, B: 0, A: 255}, Position: 0.7},
						background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 255, A: 255}, Position: 0.8},
						background.GradientStop{Color: color.RGBA{R: 128, G: 0, B: 128, A: 255}, Position: 0.9},
						background.GradientStop{Color: color.RGBA{R: 255, G: 255, B: 0, A: 255}, Position: 1},
					).SetCenter(0.5, 0.5).SetPadding(100),
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
		result, err := sample.canvas.RenderToImage(sampleCode)
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
