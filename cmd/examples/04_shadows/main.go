package main

import (
	"image/color"
	"log"
	"os"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	// Simple code example
	input := `func hello() string {
    return "Hello, World!"
}`

	// Create a canvas with a shadow
	canvas := render.NewCanvas().
		WithChrome(chrome.NewMacChrome(
			chrome.MacStyleSequoia,
			chrome.WithTitle("Shadow Example"))).
		WithBackground(
			background.NewColorBackground().
				WithColor(color.RGBA{R: 40, G: 42, B: 54, A: 255}).
				WithPadding(30).
				WithCornerRadius(8).
				WithShadow(
					background.NewShadow().
						WithOffset(0, 0).                                // Slightly downward offset
						WithBlur(30).                                    // Large blur for softness
						WithSpread(10).                                  // Large spread for more presence
						WithColor(color.RGBA{R: 0, G: 0, B: 0, A: 120}), // Slightly opaque
				),
		).
		WithContent(code.DefaultRenderer(input).
			WithLanguage("go").
			WithTheme("dracula").
			WithTabWidth(4).
			WithLineNumbers(true),
		)

	os.MkdirAll("example_output", 0755)
	err := canvas.SaveAsPNG("example_output/shadow.png")
	if err != nil {
		log.Fatal(err)
	}
}
