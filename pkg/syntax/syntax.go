package syntax

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// Style represents a syntax highlighting style
type Style struct {
	Name  string
	Style *chroma.Style
}

// Token represents a syntax-highlighted token
type Token struct {
	Text   string
	Color  color.Color
	Bold   bool
	Italic bool
}

// Line represents a line of syntax-highlighted tokens
type Line struct {
	Tokens []Token
}

// HighlightedCode represents syntax-highlighted code
type HighlightedCode struct {
	Lines []Line
}

// GetAvailableStyles returns a list of all available syntax highlighting styles
func GetAvailableStyles() []string {
	return styles.Names()
}

// GetAvailableLanguages returns a list of all supported languages
func GetAvailableLanguages() []string {
	var langs []string
	for _, l := range lexers.Registry.Lexers {
		langs = append(langs, l.Config().Name)
	}
	return langs
}

// Highlight performs syntax highlighting on the given code
func Highlight(code, language, styleName string) (*HighlightedCode, error) {
	// Get lexer for the language
	l := lexers.Get(language)
	if l == nil {
		l = lexers.Analyse(code)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Get the style
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	// Create a custom formatter that builds our HighlightedCode structure
	formatter := &customFormatter{}

	// Create an iterator for the tokens
	iterator, err := l.Tokenise(nil, code)
	if err != nil {
		return nil, fmt.Errorf("error tokenizing code: %v", err)
	}

	// Format the tokens
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return nil, fmt.Errorf("error formatting code: %v", err)
	}

	return formatter.Result, nil
}

// customFormatter implements chroma.Formatter interface
type customFormatter struct {
	Result *HighlightedCode
}

func (f *customFormatter) Format(w *bytes.Buffer, style *chroma.Style, iterator chroma.Iterator) error {
	f.Result = &HighlightedCode{}
	currentLine := Line{}

	for token := iterator(); token != chroma.EOF; token = iterator() {
		entry := style.Get(token.Type)
		tokenColor := entry.Colour
		c := color.RGBA{
			R: uint8(tokenColor.Red()),
			G: uint8(tokenColor.Green()),
			B: uint8(tokenColor.Blue()),
			A: 255, // Chroma colors are always fully opaque
		}

		// Split token text by newlines
		parts := bytes.Split([]byte(token.Value), []byte("\n"))
		for i, part := range parts {
			if len(part) > 0 {
				currentLine.Tokens = append(currentLine.Tokens, Token{
					Text:   string(part),
					Color:  c,
					Bold:   entry.Bold == chroma.Yes,
					Italic: entry.Italic == chroma.Yes,
				})
			}

			// If this isn't the last part, we've hit a newline
			if i < len(parts)-1 {
				f.Result.Lines = append(f.Result.Lines, currentLine)
				currentLine = Line{}
			}
		}
	}

	// Add the last line if it has any tokens
	if len(currentLine.Tokens) > 0 {
		f.Result.Lines = append(f.Result.Lines, currentLine)
	}

	return nil
}

func (f *customFormatter) TabWidth() int { return 8 }
