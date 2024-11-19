package main

import (
	"fmt"
	"image/color"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/render"
	"github.com/watzon/goshot/pkg/syntax"
)

func main() {
	// Sample code to highlight
	sampleCode := `package syntax

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// Style represents a syntax highlighting style
type Style struct {
	Name  string
	Style *chroma.Style
}

// Token represents a syntax-highlighted token
type Token struct {
	Text   string
	Color  color.Color
	Bold   bool
	Italic bool
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
						SetPaddingValue(40),
				).
				SetSyntaxOptions(&syntax.HighlightOptions{
					Style:        "dracula",
					TabWidth:     4,
					ShowLineNums: true,
				}).
				SetRenderConfig(
					syntax.DefaultConfig().
						SetShowLineNumbers(true).
						SetStartLineNumber(3).
						SetEndLineNumber(12),
				),
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
						SetPaddingValue(40),
				).
				SetSyntaxOptions(&syntax.HighlightOptions{
					Style:        "catppuccin-mocha",
					TabWidth:     4,
					ShowLineNums: false,
				}).
				SetRenderConfig(syntax.DefaultConfig().SetShowLineNumbers(false)),
		},
		{
			name: "catppuccin-latte",
			canvas: render.NewCanvas().
				SetChrome(chrome.NewWindows11Chrome(
					chrome.WithTitle("My App"),
					chrome.WithTitleBar(false),
					chrome.WithCornerRadius(10),
				)).
				SetSyntaxOptions(&syntax.HighlightOptions{
					Style:        "catppuccin-latte",
					TabWidth:     4,
					ShowLineNums: true,
				}).
				SetRenderConfig(syntax.DefaultConfig().SetShowLineNumbers(true)),
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
						SetPaddingValue(40),
				).
				SetSyntaxOptions(&syntax.HighlightOptions{
					Style:        "gruvbox",
					TabWidth:     4,
					ShowLineNums: false,
				}).
				SetRenderConfig(syntax.DefaultConfig().SetShowLineNumbers(false)),
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
