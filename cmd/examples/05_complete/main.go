package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/content/code"
	"github.com/watzon/goshot/fonts"
	"github.com/watzon/goshot/render"
)

func main() {
	// Simple code example
	input := `package main

import "(
	"fmt"
	"strings"
)"

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
		WithColor(color.RGBA{R: 40, G: 42, B: 54, A: 255}).
		WithPadding(30).
		WithCornerRadius(8).
		WithShadow(
			background.NewShadow().
				WithOffset(0, 0).                                // Slightly downward offset
				WithBlur(30).                                    // Large blur for softness
				WithSpread(10).                                  // Large spread for more presence
				WithColor(color.RGBA{R: 0, G: 0, B: 0, A: 120}), // Slightly opaque
		)

	linearGradientBackground := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40)

	radialGradientBackground := background.NewGradientBackground(
		background.RadialGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithCenter(0.5, 0.5).WithPadding(40)

	angularGradientBackground := background.NewGradientBackground(
		background.AngularGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40).WithCenter(0.5, 0.5)

	diamondGradientBackground := background.NewGradientBackground(
		background.DiamondGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40).WithCenter(0.5, 0.5)

	spiralGradientBackground := background.NewGradientBackground(
		background.SpiralGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40).WithCenter(0.5, 0.5).WithIntensity(1.5)

	squareGradientBackground := background.NewGradientBackground(
		background.SquareGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40).WithCenter(0.5, 0.5)

	starGradientBackground := background.NewGradientBackground(
		background.StarGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(90).WithPadding(40).WithCenter(0.5, 0.5).WithIntensity(1.5)

	imageBackground, err := background.NewImageBackgroundFromFile("cmd/examples/03_image_background/background.png")
	if err != nil {
		log.Fatal(err)
	}
	imageBackground = imageBackground.
		WithScaleMode(background.ImageScaleTile).
		WithBlur(background.GaussianBlur, 0.5)

	// Set up the code style
	content := code.DefaultRenderer(input).
		WithLanguage("go").
		WithTheme("dracula").
		WithTabWidth(4).
		WithLineNumbers(true).
		WithLineRange(1, 17).
		WithLineHighlightRange(8, 10).
		WithFontName("JetBrainsMonoNerdFont", &fonts.FontStyle{
			Weight:  fonts.WeightRegular,
			Stretch: fonts.StretchNormal,
		}).
		WithFontSize(16).
		WithLineHeight(1.2).
		WithPadding(10, 30, 30, 30).
		WithMaxWidth(800).
		WithLineNumberPadding(20)

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

	os.MkdirAll("example_output", 0755)
	for i, chrome := range chromes {
		for j, background := range backgrounds {
			canvas := render.NewCanvas().
				WithChrome(chrome).
				WithBackground(background).
				WithContent(content)
			err := canvas.SaveAsPNG(fmt.Sprintf("example_output/output_%d_%d.png", i, j))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
