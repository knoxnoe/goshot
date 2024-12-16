package main

import (
	"image/color"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/content/code"
	"github.com/watzon/goshot/render"
)

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func main() {
	// List of color blind friendly color schemes
	cb_schemes := []string{
		"blueprint",
		"glacier",
		"terminal",
		"neonoir",
		"obsidian",
		"papercut",
		"parchment",
		"garden",
	}

	// Template for the wiki page that lists every available color scheme
	md_template := `# Available Color Schemes

Goshot includes a ton of color schemes, including all of those available in [Chroma](https://github.com/alecthomas/chroma) and several that are not. Some of these are specifically designed to be color blind friendly, making Goshot a great choice for accessibility.

## Table of Contents
- Color Blindness Friendly Schemes
{{ range .CbSchemes }}    - [{{ . }}](#{{ . }}-colorblind-friendly)
{{ end }}
- Other Schemes
{{ range .Schemes }}{{ if not (contains $.CbSchemes .) }}    - [{{ . }}](#{{ . }})
{{ end }}{{ end }}

## Examples

{{ range .Schemes }}
### {{ . }}{{ if contains $.CbSchemes . }} (colorblind friendly){{ end }}
![{{ . }}](./assets/images/examples/{{ . }}.png)

{{ end }}`

	// Simple example showing basic usage with a solid color background
	input := `package main

import (
    "fmt"
    "strings"
    "time"
)

// ColorName represents a named color in our system. We support a variety of standard colors that can be used
// throughout the application. Each color can be customized by the user through their configuration file.
type ColorName string

const (
    // These are our predefined colors. Each one maps to a specific RGB or HSL value in the theme configuration.
    ColorPrimary   ColorName = "primary"   // Main application color, used for highlights and important UI elements
    ColorSecondary ColorName = "secondary" // Used for less prominent elements, providing visual hierarchy
    ColorAccent    ColorName = "accent"    // Bright, attention-grabbing color for calls-to-action and highlights
    ColorNeutral   ColorName = "neutral"   // Balanced color used for general UI elements and backgrounds
)

// formatColor applies ANSI color codes to create visual styling in the terminal. It supports both foreground
// and background colors, as well as additional styling like bold, italic, and underline effects.
func formatColor(text string, colorName ColorName) string {
    prefix := strings.Repeat("=", 20)
    suffix := strings.Repeat("=", 20)
    timestamp := time.Now().Format("15:04:05")

    return fmt.Sprintf("%s [%s] %s: %s %s", prefix, timestamp, colorName, text, suffix)
}

func main() {
    // Demonstrate different color formatting with timestamps and decorative elements
    colors := []ColorName{ColorPrimary, ColorSecondary, ColorAccent, ColorNeutral}
    
    for _, color := range colors {
        message := formatColor("This is a sample message", color)
        fmt.Println(message)
    }
}`

	color_schemes := code.GetAvailableStyles()
	sort.Strings(color_schemes)

	for _, scheme := range color_schemes {
		canvas := render.NewCanvas().
			WithChrome(chrome.NewMacChrome(
				chrome.MacStyleSequoia,
				chrome.WithTitle(scheme+" example"))).
			WithBackground(
				background.NewColorBackground().
					WithColor(color.RGBA{R: 20, G: 30, B: 40, A: 255}).
					WithPadding(40),
			).
			WithContent(code.DefaultRenderer(input).
				WithLanguage("go").
				WithTheme(scheme).
				WithTabWidth(4).
				WithLineNumbers(true))

		os.MkdirAll("example_output", 0755)
		err := canvas.SaveAsPNG("example_output/" + scheme + ".png")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Apply the template and export the markdown file
	tmpl := template.New("md").Funcs(template.FuncMap{
		"contains": contains,
	})
	tmpl, err := tmpl.Parse(md_template)
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		CbSchemes []string
		Schemes   []string
	}{
		CbSchemes: cb_schemes,
		Schemes:   color_schemes,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("example_output/Color-Schemes.md", []byte(buf.String()), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
