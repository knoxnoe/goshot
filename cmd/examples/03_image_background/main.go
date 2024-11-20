package main

import (
	"image"
	_ "image/jpeg" // Register JPEG format
	"log"
	"os"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/render"
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

	code := `// Example of a concurrent worker pool
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, j)
        time.Sleep(time.Second)
        results <- j * 2
    }
}`

	// Create an image background
	bg := background.NewImageBackground(img).
		SetScaleMode(background.ImageScaleTile).
		SetBlurRadius(0.5).
		SetOpacity(1.0).
		SetPadding(40).
		SetCornerRadius(10)

	canvas := render.NewCanvas().
		SetChrome(
			chrome.NewMacOSChrome(
				chrome.WithTitle("Image Background Example"),
				chrome.WithDarkMode(true),
			)).
		SetBackground(bg).
		SetCodeStyle(&render.CodeStyle{
			Language:        "go",
			Theme:           "dracula",
			TabWidth:        4,
			ShowLineNumbers: true,
		})

	img, err = canvas.RenderToImage(code)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.SaveAsPNG(img, "image_background.png"); err != nil {
		log.Fatal(err)
	}
}
