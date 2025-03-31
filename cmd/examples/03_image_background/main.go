package main

import (
	"image"
	_ "image/jpeg" // Register JPEG format
	"log"
	"os"

	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/content/code"
	"github.com/watzon/goshot/render"
)

func main() {
	// Load a background image
	file, err := os.Open("./cmd/examples/03_image_background/background.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	input := `// Example of a concurrent worker pool
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, j)
        time.Sleep(time.Second)
        results <- j * 2
    }
}`

	// Create an image background
	// Create the content renderer once to reuse
	content := code.DefaultRenderer(input).
		WithLanguage("go").
		WithTheme("dracula").
		WithTabWidth(4).
		WithLineNumbers(true)

	// Create image backgrounds with different blur effects
	noBlur := background.NewImageBackground(img).
		WithScaleMode(background.ImageScaleTile).
		WithOpacity(1.0).
		WithPadding(40).
		WithCornerRadius(10)

	gaussianBlur := background.NewImageBackground(img).
		WithScaleMode(background.ImageScaleTile).
		WithBlur(background.GaussianBlur, 3.0).
		WithOpacity(1.0).
		WithPadding(40).
		WithCornerRadius(10)

	pixelatedBlur := background.NewImageBackground(img).
		WithScaleMode(background.ImageScaleTile).
		WithBlur(background.PixelatedBlur, 8.0).
		WithOpacity(1.0).
		WithPadding(40).
		WithCornerRadius(10)

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
			WithChrome(
				chrome.NewMacChrome(
					chrome.MacStyleSequoia,
					chrome.WithTitle("Image Background - "+bg.name),
				)).
			WithBackground(bg.background).
			WithContent(content)

		err = canvas.SaveAsPNG("example_output/image_background_" + bg.name + ".png")
		if err != nil {
			log.Fatal(err)
		}
	}
}
