package main

import (
	"image"
	"image/color"
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

	code := `func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, j)
        time.Sleep(time.Second)
        results <- j * 2
    }
}`

	// Create a canvas with an image background and shadow
	canvas := render.NewCanvas().
		SetChrome(chrome.NewMacOSChrome(chrome.WithTitle("Image Background Shadow"))).
		SetBackground(
			background.NewImageBackground(img).
				SetScaleMode(background.ImageScaleCover).
				SetBlurRadius(2).
				SetOpacity(0.8).
				SetPadding(40).
				SetCornerRadius(8).
				SetShadow(
					background.NewShadow().
						SetOffset(0, 3).                                // Slightly downward offset
						SetBlur(20).                                    // Large blur for softness
						SetSpread(8).                                   // Large spread for more presence
						SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 200}), // Slightly opaque
				),
		).
		SetCodeStyle(&render.CodeStyle{
			Language:        "go",
			Theme:           "dracula",
			TabWidth:        4,
			ShowLineNumbers: true,
		})

	// Render to file
	img, err = canvas.RenderToImage(code)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.SaveAsPNG(img, "output.png"); err != nil {
		log.Fatal(err)
	}
}
