package syntax

import (
	"image/color"
	"testing"
)

func TestHighlight(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		language  string
		styleName string
		wantErr   bool
	}{
		{
			name: "Go code",
			code: `package main

func main() {
    println("Hello, World!")
}`,
			language:  "go",
			styleName: "monokai",
			wantErr:   false,
		},
		{
			name: "Python code",
			code: `def greet(name):
    print(f"Hello, {name}!")

greet("World")`,
			language:  "python",
			styleName: "monokai",
			wantErr:   false,
		},
		{
			name: "Invalid language",
			code: `some random text`,
			language:  "nonexistent",
			styleName: "monokai",
			wantErr:   false, // Should not error, falls back to text
		},
		{
			name: "Invalid style",
			code: `package main`,
			language:  "go",
			styleName: "nonexistent",
			wantErr:   false, // Should not error, falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Highlight(tt.code, tt.language, tt.styleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Highlight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result == nil {
				t.Error("Highlight() returned nil result")
				return
			}
			if len(result.Lines) == 0 {
				t.Error("Highlight() returned no lines")
			}
		})
	}
}

func TestGetAvailableStyles(t *testing.T) {
	styles := GetAvailableStyles()
	if len(styles) == 0 {
		t.Error("GetAvailableStyles() returned no styles")
	}

	// Check for some common styles
	commonStyles := map[string]bool{
		"monokai": false,
		"github":  false,
		"vs":      false,
		"dracula": false,
	}

	for _, style := range styles {
		if _, ok := commonStyles[style]; ok {
			commonStyles[style] = true
		}
	}

	for style, found := range commonStyles {
		if !found {
			t.Errorf("Common style %q not found in available styles", style)
		}
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	langs := GetAvailableLanguages()
	if len(langs) == 0 {
		t.Error("GetAvailableLanguages() returned no languages")
	}

	// Check for some common languages
	commonLangs := map[string]bool{
		"Go":     false,
		"Python": false,
		"Java":   false,
		"JavaScript": false,
	}

	for _, lang := range langs {
		if _, ok := commonLangs[lang]; ok {
			commonLangs[lang] = true
		}
	}

	for lang, found := range commonLangs {
		if !found {
			t.Errorf("Common language %q not found in available languages", lang)
		}
	}
}

func TestTokenProperties(t *testing.T) {
	code := `package main

func main() {
    println("Hello, World!")
}`
	result, err := Highlight(code, "go", "monokai")
	if err != nil {
		t.Fatalf("Highlight() failed: %v", err)
	}

	// Test that tokens have reasonable properties
	for _, line := range result.Lines {
		for _, token := range line.Tokens {
			// Text should not be empty
			if token.Text == "" {
				t.Error("Token has empty text")
			}

			// Color should not be transparent
			if c, ok := token.Color.(color.RGBA); ok {
				if c.A == 0 {
					t.Error("Token has transparent color")
				}
			}
		}
	}
}

func TestLineBreaks(t *testing.T) {
	code := "line1\nline2\nline3"
	result, err := Highlight(code, "text", "monokai")
	if err != nil {
		t.Fatalf("Highlight() failed: %v", err)
	}

	if len(result.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(result.Lines))
	}
}
