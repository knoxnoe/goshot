package main

import (
	"image/color"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/render"
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
		"glacier-ice",
		"modern-term",
		"neo-noir",
		"obsidian-deep",
		"papercut",
		"parchment-scroll",
		"secret-garden",
	}

	// Template for the wiki page that lists every available color scheme
	md_template := `# Available Color Schemes

Goshot includes a ton of color schemes, including all of those available in [Chroma](https://github.com/alecthomas/chroma) and several that are not. Some of these are specifically designed to be color blind friendly, making Goshot a great choice for accessibility.

## All Schemes

{{ range .Schemes }}
- [{{ . }}](#{{ . }})
{{ end }}

## Color Blindness Friendly Schemes

{{ range .CbSchemes }}
- [{{ . }}](#{{ . }})
{{ end }}
 
## Examples

{{ range .Schemes }}
### {{ . }}{{ if contains $.CbSchemes . }} (colorblind friendly){{ end }}
![{{ . }}](./{{ . }}.png)

{{ end }}`

	// Simple example showing basic usage with a solid color background
	input := `package main

import (
    "fmt"
    "strings"
)

// This is a very long comment that should wrap to multiple lines when the width is constrained. We'll make it even longer to ensure it wraps at least once or twice when rendered.
func main() {
    message := "Hello, " + strings.Repeat("World! ", 10)
    fmt.Println(message)
}`

	color_schemes := code.GetAvailableStyles()
	sort.Strings(color_schemes)

	canvas := render.NewCanvas().
		WithChrome(chrome.NewMacChrome(
			chrome.MacStyleSequoia,
			chrome.WithTitle("Basic Example"))).
		WithBackground(
			background.NewColorBackground().
				WithColor(color.RGBA{R: 20, G: 30, B: 40, A: 255}).
				WithPadding(40),
		)

	content := code.DefaultRenderer(input).
		WithLanguage("go").
		WithTabWidth(4).
		WithLineNumbers(true)

	for _, scheme := range color_schemes {
		content = content.WithTheme(scheme)

		canvas = canvas.WithContent(content)

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

	err = os.WriteFile("example_output/index.md", []byte(buf.String()), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
