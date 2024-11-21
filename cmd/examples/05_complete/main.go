package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	// Simple code example
	code := `package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", strings.TrimSpace(name))
}

func main() {
	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		fmt.Println(greet(name))
	}
}`

	macOsChrome := chrome.NewMacChrome(
		chrome.MacStyleSequoia,
		chrome.WithTitle("Goshot Example"),
		chrome.WithVariant(chrome.ThemeVariantLight),
		chrome.WithTitleBar(true),
		chrome.WithCornerRadius(9))

	macOsChromeDark := chrome.NewMacChrome(
		chrome.MacStyleSequoia,
		chrome.WithTitle("Goshot Example"),
		chrome.WithVariant(chrome.ThemeVariantDark),
		chrome.WithTitleBar(true),
		chrome.WithCornerRadius(9))

	windows11Chrome := chrome.NewWindowsChrome(
		chrome.WindowsStyleWin11,
		chrome.WithTitle("Goshot Example"),
		chrome.WithVariant(chrome.ThemeVariantLight),
		chrome.WithTitleBar(true),
		chrome.WithCornerRadius(8))

	windows11ChromeDark := chrome.NewWindowsChrome(
		chrome.WindowsStyleWin11,
		chrome.WithTitle("Goshot Example"),
		chrome.WithVariant(chrome.ThemeVariantDark),
		chrome.WithTitleBar(true),
		chrome.WithCornerRadius(8))

	colorBackground := background.NewColorBackground().
		SetColor(color.RGBA{R: 40, G: 42, B: 54, A: 255}).
		SetPadding(30).
		SetCornerRadius(8).
		SetShadow(
			background.NewShadow().
				SetOffset(0, 0).                                // Slightly downward offset
				SetBlur(30).                                    // Large blur for softness
				SetSpread(10).                                  // Large spread for more presence
				SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 120}), // Slightly opaque
		)

	linearGradientBackground := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40)

	radialGradientBackground := background.NewGradientBackground(
		background.RadialGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetCenter(0.5, 0.5).SetPadding(40)

	angularGradientBackground := background.NewGradientBackground(
		background.AngularGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40).SetCenter(0.5, 0.5)

	diamondGradientBackground := background.NewGradientBackground(
		background.DiamondGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40).SetCenter(0.5, 0.5)

	spiralGradientBackground := background.NewGradientBackground(
		background.SpiralGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40).SetCenter(0.5, 0.5).SetIntensity(1.5)

	squareGradientBackground := background.NewGradientBackground(
		background.SquareGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40).SetCenter(0.5, 0.5)

	starGradientBackground := background.NewGradientBackground(
		background.StarGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).SetAngle(90).SetPadding(40).SetCenter(0.5, 0.5).SetIntensity(1.5)

	imageBackground, err := background.NewImageBackgroundFromFile("cmd/examples/03_image_background/background.png")
	if err != nil {
		log.Fatal(err)
	}
	imageBackground = imageBackground.
		SetScaleMode(background.ImageScaleTile).
		SetBlurRadius(0.5)

	codeStyle := &render.CodeStyle{
		Language:        "go",
		Theme:           "dracula",
		TabWidth:        4,
		ShowLineNumbers: true,
		LineNumberRange: render.LineRange{
			Start: 1,
			End:   17,
		},
		LineHighlightRanges: []render.LineRange{
			{
				Start: 8,
				End:   10,
			},
		},
		FontSize: 16,
		FontFamily: &fonts.Font{
			Name: "JetBrains Mono",
			Style: fonts.FontStyle{
				Weight:    fonts.WeightRegular,
				Italic:    false,
				Condensed: false,
				Mono:      true,
			},
		},
		LineHeight:        1.5,
		PaddingLeft:       10,
		PaddingRight:      30,
		PaddingTop:        30,
		PaddingBottom:     30,
		MinWidth:          400,
		MaxWidth:          800,
		LineNumberPadding: 20,
	}

	chromes := []chrome.Chrome{macOsChrome, windows11Chrome, macOsChromeDark, windows11ChromeDark}
	backgrounds := []background.Background{
		colorBackground,
		linearGradientBackground,
		radialGradientBackground,
		angularGradientBackground,
		diamondGradientBackground,
		spiralGradientBackground,
		squareGradientBackground,
		starGradientBackground,
		imageBackground,
	}

	for i, chrome := range chromes {
		for j, background := range backgrounds {
			canvas := render.NewCanvas().
				SetChrome(chrome).
				SetBackground(background).
				SetCodeStyle(codeStyle)
			img, err := canvas.RenderToImage(code)
			if err != nil {
				log.Fatal(err)
			}
			if err := render.SaveAsPNG(img, fmt.Sprintf("output_%d_%d.png", i, j)); err != nil {
				log.Fatal(err)
			}
		}
	}
}
