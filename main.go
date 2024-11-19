package main

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/syntax"
)

func main() {
	// Sample code to highlight
	sampleCode := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`

	// Sample configurations
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
		// Highlight the code
		highlighted, err := syntax.Highlight(sampleCode, "go", "monokai")
		if err != nil {
			fmt.Printf("Error highlighting code: %v\n", err)
			continue
		}

		// Create render config with theme-appropriate colors
		config := syntax.DefaultConfig()
		if sample.darkMode {
			config.LineNumberColor = color.RGBA{R: 128, G: 128, B: 128, A: 255}
			config.LineNumberBg = color.RGBA{R: 40, G: 40, B: 40, A: 255}
		} else {
			config.LineNumberColor = color.RGBA{R: 128, G: 128, B: 128, A: 255}
			config.LineNumberBg = color.RGBA{R: 245, G: 245, B: 245, A: 255}
		}

		// Create an image from the highlighted code
		content, err := highlighted.RenderToImage(config)
		if err != nil {
			fmt.Printf("Error rendering code: %v\n", err)
			continue
		}

		// Render the chrome
		bg := background.NewColorBackground()
		// Use a lilac color for all themes
		lilacColor := color.RGBA{R: 230, G: 220, B: 255, A: 255}
		bg.SetColor(lilacColor)

		result, err := sample.chrome.Render(content)
		if err != nil {
			fmt.Printf("Error rendering chrome: %v\n", err)
			continue
		}
		result = bg.Apply(result)

		// Save the result
		outDir := "samples"
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			continue
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("%s.png", sample.name))
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			continue
		}
		defer f.Close()

		if err := png.Encode(f, result); err != nil {
			fmt.Printf("Error encoding PNG: %v\n", err)
		}
	}
}
