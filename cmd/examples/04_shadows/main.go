package main

import (
	"image/color"
	"log"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	// Simple code example
	code := `func hello() string {
    return "Hello, World!"
}`

	// Create a canvas with a shadow
	canvas := render.NewCanvas().
		SetChrome(chrome.NewMacChrome(
			chrome.MacStyleSequoia,
			chrome.WithTitle("Shadow Example"))).
		SetBackground(
			background.NewColorBackground().
				SetColor(color.RGBA{R: 40, G: 42, B: 54, A: 255}).
				SetPadding(30).
				SetCornerRadius(8).
				SetShadow(
					background.NewShadow().
						SetOffset(0, 0).                                // Slightly downward offset
						SetBlur(30).                                    // Large blur for softness
						SetSpread(10).                                  // Large spread for more presence
						SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 120}), // Slightly opaque
				),
		).
		SetCodeStyle(&render.CodeStyle{
			Language:        "go",
			Theme:           "dracula",
			TabWidth:        4,
			ShowLineNumbers: true,
		})

	// Render to file
	img, err := canvas.RenderToImage(code)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.SaveAsPNG(img, "shadow.png"); err != nil {
		log.Fatal(err)
	}
}
