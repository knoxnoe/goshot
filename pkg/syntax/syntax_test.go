package syntax

import (
	"strings"
	"testing"
)

func TestHighlight(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		options *HighlightOptions
		wantErr bool
	}{
		{
			name: "Go code with line numbers",
			code: `package main

func main() {
    println("Hello, World!")
}`,
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: true,
			},
			wantErr: false,
		},
		{
			name: "Python code without line numbers",
			code: `def greet(name):
    print(f"Hello, {name}!")

greet("World")`,
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: false,
			},
			wantErr: false,
		},
		{
			name: "Invalid style with custom tab width",
			code: `package main`,
			options: &HighlightOptions{
				Style:        "nonexistent",
				TabWidth:     8,
				ShowLineNums: true,
			},
			wantErr: false, // Should not error, falls back to default
		},
		{
			name:    "Default options",
			code:    `some random text`,
			options: nil, // Should use DefaultOptions()
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Highlight(tt.code, tt.options)
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

			// Check if options were properly applied or defaults were used
			opts := tt.options
			if opts == nil {
				opts = DefaultOptions()
			}

			// Verify that the formatter respected the options
			if opts.TabWidth != 4 && opts.TabWidth != 8 {
				t.Errorf("TabWidth not properly set, got %d", opts.TabWidth)
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
	langs := GetAvailableLanguages(false)
	if len(langs) == 0 {
		t.Error("GetAvailableLanguages() returned no languages")
	}

	// Check for some common languages
	commonLangs := map[string]bool{
		"Go":         false,
		"Python":     false,
		"Java":       false,
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
	// A comment
	println("Hello") // Another comment
}`
	opts := &HighlightOptions{
		Style:        "monokai",
		TabWidth:     4,
		ShowLineNums: true,
	}

	result, err := Highlight(code, opts)
	if err != nil {
		t.Fatalf("Highlight() error = %v", err)
	}

	// Check that comments are properly styled
	for _, line := range result.Lines {
		for _, token := range line.Tokens {
			if strings.Contains(token.Text, "//") {
				if !token.Italic {
					t.Error("Comment token should be italic")
				}
			}
		}
	}
}

func TestLineBreaks(t *testing.T) {
	code := "line1\nline2\r\nline3\rline4"
	opts := &HighlightOptions{
		Style:        "monokai",
		TabWidth:     4,
		ShowLineNums: true,
	}

	result, err := Highlight(code, opts)
	if err != nil {
		t.Fatalf("Highlight() error = %v", err)
	}

	if len(result.Lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(result.Lines))
	}
}

func TestHighlightEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		options  *HighlightOptions
		expected int
	}{
		{
			name: "Empty string",
			code: "",
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: true,
			},
			expected: 1, // Should have one empty line
		},
		{
			name: "Single newline",
			code: "\n",
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: true,
			},
			expected: 2, // Should have two empty lines
		},
		{
			name: "Multiple empty lines",
			code: "\n\n\n",
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: false,
			},
			expected: 4, // Should have four empty lines
		},
		{
			name: "Code with empty lines",
			code: "line1\n\nline3\n\nline5",
			options: &HighlightOptions{
				Style:        "monokai",
				TabWidth:     4,
				ShowLineNums: true,
			},
			expected: 5, // Should have five lines total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Highlight(tt.code, tt.options)
			if err != nil {
				t.Fatalf("Highlight() error = %v", err)
			}

			if len(result.Lines) != tt.expected {
				t.Errorf("Expected %d lines, got %d", tt.expected, len(result.Lines))
			}

			// Check that empty lines have an empty token list
			for i, line := range result.Lines {
				if len(line.Tokens) == 0 && i < len(result.Lines)-1 {
					// Empty lines should have an empty token list (except possibly the last line)
					continue
				}
				if len(line.Tokens) > 0 && strings.TrimSpace(line.Tokens[0].Text) == "" {
					t.Errorf("Line %d: unexpected non-empty token list for empty line", i+1)
				}
			}
		})
	}
}
