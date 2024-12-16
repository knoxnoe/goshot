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
	bg := background.NewImageBackground(img).
		WithScaleMode(background.ImageScaleTile).
		WithBlurRadius(0.5).
		WithOpacity(1.0).
		WithPadding(40).
		WithCornerRadius(10)

	canvas := render.NewCanvas().
		WithChrome(
			chrome.NewMacChrome(
				chrome.MacStyleSequoia,
				chrome.WithTitle("Image Background Example"),
			)).
		WithBackground(bg).
		WithContent(
			code.DefaultRenderer(input).
				WithLanguage("go").
				WithTheme("dracula").
				WithTabWidth(4).
				WithLineNumbers(true),
		)

	os.MkdirAll("example_output", 0755)
	err = canvas.SaveAsPNG("example_output/image_background.png")
	if err != nil {
		log.Fatal(err)
	}
}
