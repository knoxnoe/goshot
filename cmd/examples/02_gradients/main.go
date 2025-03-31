package main

import (
	"image/color"
	"log"
	"os"

	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/content/code"
	"github.com/watzon/goshot/render"
)

func main() {
	input := `// Example function demonstrating error handling
func processItem(item string) error {
    if item == "" {
        return errors.New("item cannot be empty")
    }
    
    result, err := process(item)
    if err != nil {
        return fmt.Errorf("failed to process item: %w", err)
    }
    
    return nil
}`

	// Create gradient backgrounds with different blur effects
	gaussianBlur := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(45).WithPadding(40).WithBlur(background.GaussianBlur, 3)

	pixelatedBlur := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(45).WithPadding(40).WithBlur(background.PixelatedBlur, 8)

	noBlur := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(45).WithPadding(40)

	// Create the content renderer once to reuse
	content := code.DefaultRenderer(input).
		WithLanguage("go").
		WithTheme("dracula").
		WithTabWidth(4).
		WithLineNumbers(true)

	// Create directory for output
	os.MkdirAll("example_output", 0755)

	// Render each variant
	backgrounds := []struct {
		name       string
		background background.Background
	}{
		{"no_blur", noBlur},
		{"gaussian_blur", gaussianBlur},
		{"pixelated_blur", pixelatedBlur},
	}

	for _, bg := range backgrounds {
		canvas := render.NewCanvas().
			WithChrome(chrome.NewMacChrome(
				chrome.MacStyleSequoia,
				chrome.WithTitle("Gradient Example - "+bg.name))).
			WithBackground(bg.background).
			WithContent(content)

		err := canvas.SaveAsPNG("example_output/gradients_" + bg.name + ".png")
		if err != nil {
			log.Fatal(err)
		}
	}
}
