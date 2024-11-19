package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
)

func main() {
	// List available fonts
	fmt.Println("Available fonts:")
	for _, font := range fonts.ListFonts() {
		fmt.Printf("- %s\n", font)
	}

	// Create a sample content image
	content := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill content with a light gray color
	for y := 0; y < content.Bounds().Dy(); y++ {
		for x := 0; x < content.Bounds().Dx(); x++ {
			content.Set(x, y, color.RGBA{R: 240, G: 240, B: 240, A: 255})
		}
	}

	// Define sample configurations
	samples := []struct {
		name     string
		chrome   chrome.Chrome
		darkMode bool
	}{
		{
			name:   "macOS Light",
			chrome: chrome.NewMacOSChrome().SetTitle("My App"),
		},
		{
			name:     "macOS Dark",
			chrome:   chrome.NewMacOSChrome().SetTitle("My App").SetDarkMode(true),
			darkMode: true,
		},
		{
			name:   "Windows 11 Light",
			chrome: chrome.NewWindows11Chrome().SetTitle("My App"),
		},
		{
			name:     "Windows 11 Dark",
			chrome:   chrome.NewWindows11Chrome().SetTitle("My App").SetDarkMode(true),
			darkMode: true,
		},
	}

	// Generate samples
	for _, sample := range samples {
		// Render the chrome
		result, err := sample.chrome.Render(content)
		if err != nil {
			log.Fatalf("Failed to render chrome: %v", err)
		}

		// Save the result
		outputFile := fmt.Sprintf("chrome_%s.png", strings.ToLower(strings.ReplaceAll(sample.name, " ", "_")))
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer f.Close()

		if err := png.Encode(f, result); err != nil {
			log.Fatalf("Failed to encode PNG: %v", err)
		}

		fmt.Printf("Generated %s\n", outputFile)
	}
}
