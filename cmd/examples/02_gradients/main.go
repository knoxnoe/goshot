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

	// Create a gradient background
	bg := background.NewGradientBackground(
		background.LinearGradient,
		background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
		background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
	).WithAngle(45).WithPadding(40)

	canvas := render.NewCanvas().
		WithChrome(chrome.NewMacChrome(
			chrome.MacStyleSequoia,
			chrome.WithTitle("Gradient Example"))).
		WithBackground(bg).
		WithContent(code.DefaultRenderer(input).
			WithLanguage("go").
			WithTheme("dracula").
			WithTabWidth(4).
			WithLineNumbers(true),
		)

	os.MkdirAll("example_output", 0755)
	err := canvas.SaveAsPNG("example_output/gradients.png")
	if err != nil {
		log.Fatal(err)
	}
}
