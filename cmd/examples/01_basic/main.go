package main

import (
	"image/color"
	"log"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	// Simple example showing basic usage with a solid color background
	code := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`

	canvas := render.NewCanvas().
		SetChrome(chrome.NewMacOSChrome(chrome.WithTitle("Basic Example"))).
		SetBackground(
			background.NewColorBackground().
				SetColor(color.RGBA{R: 20, G: 30, B: 40, A: 255}).
				SetPadding(40),
		).
		SetCodeStyle(&render.CodeStyle{
			Language:        "go",
			Theme:           "dracula",
			TabWidth:        4,
			ShowLineNumbers: true,
		})

	img, err := canvas.RenderToImage(code)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.SaveAsPNG(img, "basic.png"); err != nil {
		log.Fatal(err)
	}
}
