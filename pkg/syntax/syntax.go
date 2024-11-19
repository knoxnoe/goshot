package syntax

import (
	"bytes"
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
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
	Lines           []Line
	BackgroundColor color.Color
	LineNumberColor color.Color
	GutterColor     color.Color
}

// HighlightOptions contains options for syntax highlighting
type HighlightOptions struct {
	Style        string // Name of the chroma style to use
	TabWidth     int    // Number of spaces per tab
	ShowLineNums bool   // Whether to show line numbers
}

// DefaultOptions returns the default highlight options
func DefaultOptions() *HighlightOptions {
	return &HighlightOptions{
		Style:        "dracula",
		TabWidth:     4,
		ShowLineNums: true,
	}
}

// GetAvailableStyles returns a list of all available syntax highlighting styles
func GetAvailableStyles() []string {
	return styles.Names()
}

// GetAvailableLanguages returns a list of all supported languages
func GetAvailableLanguages(aliases bool) []string {
	var langs []string
	langs = append(langs, lexers.Names(false)...)
	return langs
}

// Highlight performs syntax highlighting on the given code
func Highlight(code string, opts *HighlightOptions) (*HighlightedCode, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Detect language
	lexer := lexers.Analyse(code)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get style
	style := styles.Get(opts.Style)
	if style == nil {
		style = styles.Fallback
	}

	// Format the code
	formatter := &customFormatter{
		tabWidth:     opts.TabWidth,
		showLineNums: opts.ShowLineNums,
		Result: &HighlightedCode{
			BackgroundColor: getBackgroundColor(style),
			GutterColor:     getGutterColor(style),
			LineNumberColor: getLineNumberColor(style),
		},
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return nil, err
	}

	return formatter.Result, nil
}

// customFormatter implements the chroma.Formatter interface
type customFormatter struct {
	tabWidth     int
	showLineNums bool
	Result       *HighlightedCode
}

func (f *customFormatter) Format(w *bytes.Buffer, style *chroma.Style, iterator chroma.Iterator) error {
	currentLine := Line{}
	var lastToken chroma.Token

	// Initialize with an empty line for empty input
	foundTokens := false

	for token := iterator(); token != chroma.EOF; token = iterator() {
		foundTokens = true
		lastToken = token
		entry := style.Get(token.Type)
		tokenColor := entry.Colour
		c := color.RGBA{
			R: tokenColor.Red(),
			G: tokenColor.Green(),
			B: tokenColor.Blue(),
			A: 255, // Chroma colors are always fully opaque
		}

		// Replace tabs with spaces
		text := token.Value
		if strings.Contains(text, "\t") {
			var b strings.Builder
			column := 0
			for _, ch := range text {
				if ch == '\t' {
					spaces := f.tabWidth - (column % f.tabWidth)
					b.WriteString(strings.Repeat(" ", spaces))
					column += spaces
				} else {
					b.WriteRune(ch)
					if ch == '\n' {
						column = 0
					} else {
						column++
					}
				}
			}
			text = b.String()
		}

		// Split token text by newlines
		parts := bytes.Split([]byte(text), []byte("\n"))
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
			// Always add the line, even if it's empty
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

	// Handle trailing newlines by checking if the last token was a newline
	if lastToken.Type == chroma.Text && len(lastToken.Value) > 0 && lastToken.Value[len(lastToken.Value)-1] == '\n' {
		f.Result.Lines = append(f.Result.Lines, Line{})
	}

	// If no tokens were found, add an empty line
	if !foundTokens {
		f.Result.Lines = append(f.Result.Lines, Line{})
	}

	return nil
}

func (f *customFormatter) TabWidth() int { return f.tabWidth }

func getBackgroundColor(style *chroma.Style) color.Color {
	bgColor := style.Get(chroma.Background)
	return color.RGBA{
		R: bgColor.Background.Red(),
		G: bgColor.Background.Green(),
		B: bgColor.Background.Blue(),
		A: 255,
	}
}

func getGutterColor(style *chroma.Style) color.Color {
	lineTableColor := style.Get(chroma.LineTable)
	return color.RGBA{
		R: lineTableColor.Background.Red(),
		G: lineTableColor.Background.Green(),
		B: lineTableColor.Background.Blue(),
		A: 255,
	}
}

func getLineNumberColor(style *chroma.Style) color.Color {
	lineNumColor := style.Get(chroma.LineNumbers)
	return color.RGBA{
		R: lineNumColor.Colour.Red(),
		G: lineNumColor.Colour.Green(),
		B: lineNumColor.Colour.Blue(),
		A: 255,
	}
}
