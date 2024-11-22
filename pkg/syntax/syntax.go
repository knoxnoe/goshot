package syntax

import (
	"fmt"
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

// Token represents a syntax highlighted token with its style information
type Token struct {
	Text   string      // The text content
	Color  color.Color // The color to render the token in
	Bold   bool        // Whether to render the token in bold
	Italic bool        // Whether to render the token in italic
}

// Line represents a single line of highlighted code
type Line struct {
	Tokens    []Token // The tokens in this line
	Highlight bool    // Whether this line should be highlighted
}

// HighlightedCode represents syntax highlighted code ready for rendering
type HighlightedCode struct {
	Lines           []Line      // The lines of code with their tokens
	BackgroundColor color.Color // Background color for the code block
	GutterColor     color.Color // Color for the gutter (line numbers background)
	LineNumberColor color.Color // Color for line numbers
	HighlightColor  color.Color // Color for highlighted lines

	HighlightedLines []int // Lines that should be highlighted
}

// HighlightOptions contains options for syntax highlighting
type HighlightOptions struct {
	Style            string // Name of the chroma style to use
	Language         string // Optional: Language to use for highlighting (e.g., "go", "python")
	TabWidth         int    // Number of spaces per tab
	ShowLineNums     bool   // Whether to show line numbers
	HighlightedLines []int  // Lines that should be highlighted
}

// DefaultOptions returns the default highlight options
func DefaultOptions() *HighlightOptions {
	return &HighlightOptions{
		Style:            "dracula",
		Language:         "", // Empty means auto-detect
		TabWidth:         4,
		ShowLineNums:     true,
		HighlightedLines: []int{},
	}
}

// GetAvailableStyles returns a list of all available syntax highlighting styles
func GetAvailableStyles() []string {
	return styles.Names()
}

// GetAvailableLanguages returns a list of all supported languages
func GetAvailableLanguages(aliases bool) []string {
	var langs []string
	langs = append(langs, lexers.Names(aliases)...)
	return langs
}

// GetLanguageInfo returns a map of language names to their aliases
func GetLanguageInfo() map[string][]string {
	info := make(map[string][]string)
	for _, l := range lexers.Names(false) {
		config := lexers.Get(l).Config()
		info[config.Name] = config.Aliases
	}
	return info
}

// GetLanguageByAlias returns the canonical language name for a given alias
func GetLanguageByAlias(alias string) string {
	alias = strings.ToLower(alias)
	for _, l := range lexers.Names(true) {
		config := lexers.Get(l).Config()
		if strings.EqualFold(config.Name, alias) {
			return config.Name
		}
		for _, a := range config.Aliases {
			if strings.EqualFold(a, alias) {
				return config.Name
			}
		}
	}
	return ""
}

// Highlight performs syntax highlighting on the given code
func Highlight(code string, opts *HighlightOptions) (*HighlightedCode, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Get lexer based on language or detect
	var lexer chroma.Lexer
	if opts.Language != "" {
		// Convert alias to canonical name if needed
		canonicalName := GetLanguageByAlias(opts.Language)
		if canonicalName != "" {
			lexer = lexers.Get(canonicalName)
		}

		// If not found by canonical name, try direct lookup
		if lexer == nil {
			lexer = lexers.Get(opts.Language)
		}
	}
	if lexer == nil {
		// Try to detect the language
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Configure the lexer
	lexer = chroma.Coalesce(lexer)

	// Get style
	style := styles.Get(opts.Style)
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	tokens := make([]chroma.Token, 0)
	for token := iterator(); token != chroma.EOF; token = iterator() {
		tokens = append(tokens, token)
	}

	// Format the code
	formatter := &customFormatter{
		tabWidth:         opts.TabWidth,
		highlightedLines: make(map[int]bool),
		Result: &HighlightedCode{
			BackgroundColor:  getBackgroundColor(style),
			GutterColor:      getGutterColor(style),
			LineNumberColor:  getLineNumberColor(style),
			HighlightColor:   getHighlightColor(style),
			HighlightedLines: opts.HighlightedLines,
		},
	}

	for _, line := range opts.HighlightedLines {
		formatter.highlightedLines[line] = true
	}

	err = formatter.Format(tokens, style)
	if err != nil {
		return nil, err
	}

	return formatter.Result, nil
}

// PrintTokens prints all tokens for debugging
func PrintTokens(code string) error {
	lexer := lexers.Get("go")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return err
	}

	fmt.Println("\nToken stream:")
	for token := iterator(); token != chroma.EOF; token = iterator() {
		fmt.Printf("Token{Type: %-20v, Value: %q}\n", token.Type, token.String())
	}
	return nil
}

// customFormatter implements the chroma.Formatter interface
type customFormatter struct {
	highlightedLines map[int]bool
	lineNumber       int
	tabWidth         int
	Result           *HighlightedCode
	currentLine      Line
	currentColumn    int // Track current column position for tab expansion
}

func expandTabs(text string, currentColumn, tabWidth int) (string, int) {
	if !strings.Contains(text, "\t") {
		return text, currentColumn + len(text)
	}

	var result strings.Builder
	col := currentColumn

	for _, ch := range text {
		if ch == '\t' {
			spaces := tabWidth - (col % tabWidth)
			result.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		} else {
			result.WriteRune(ch)
			col++
		}
	}

	return result.String(), col
}

func (f *customFormatter) createToken(text string, tokenType chroma.TokenType, style *chroma.Style) Token {
	entry := style.Get(tokenType)
	return Token{
		Text:   text,
		Color:  getColorFromChroma(entry.Colour),
		Bold:   entry.Bold == chroma.Yes,
		Italic: entry.Italic == chroma.Yes,
	}
}

func (f *customFormatter) addLine(line Line) {
	line.Highlight = f.highlightedLines[f.lineNumber]
	f.Result.Lines = append(f.Result.Lines, line)
	f.lineNumber++
	f.currentColumn = 0 // Reset column position for new line
}

func (f *customFormatter) addToken(text string, tokenType chroma.TokenType, style *chroma.Style) {
	if text == "" {
		return
	}

	// Expand tabs to spaces
	expandedText, newColumn := expandTabs(text, f.currentColumn, f.tabWidth)
	f.currentColumn = newColumn

	// Check if this token should be joined with the previous token
	if len(f.currentLine.Tokens) > 0 && shouldJoinTokens(tokenType) {
		lastToken := &f.currentLine.Tokens[len(f.currentLine.Tokens)-1]
		// Only join if the colors match
		if lastToken.Color == f.createToken(expandedText, tokenType, style).Color {
			lastToken.Text += expandedText
			return
		}
	}

	// Add the token with expanded text
	if expandedText != "" {
		f.currentLine.Tokens = append(f.currentLine.Tokens, f.createToken(expandedText, tokenType, style))
	}
}

func shouldJoinTokens(tokenType chroma.TokenType) bool {
	// Join punctuation tokens
	return tokenType == chroma.Punctuation ||
		strings.Contains(tokenType.String(), "Punctuation") ||
		strings.Contains(tokenType.String(), "Operator") ||
		strings.Contains(tokenType.String(), "Parenthesis") ||
		strings.Contains(tokenType.String(), "Bracket") ||
		strings.Contains(tokenType.String(), "Brace")
}

func (f *customFormatter) processNewlines(text string, tokenType chroma.TokenType, style *chroma.Style) (Line, bool) {
	if !strings.Contains(text, "\n") {
		return Line{}, false
	}

	parts := strings.Split(text, "\n")
	lastIndex := len(parts) - 1

	// Handle all but the last part (they end in newline)
	for i := 0; i < lastIndex; i++ {
		if parts[i] != "" {
			f.addToken(parts[i], tokenType, style)
		}
		f.addLine(f.currentLine)
		f.currentLine = Line{}
	}

	// The last part doesn't end in a newline, so return it as the current line
	var nextLine Line
	if parts[lastIndex] != "" {
		expandedText, col := expandTabs(parts[lastIndex], 0, f.tabWidth)
		if expandedText != "" {
			nextLine.Tokens = append(nextLine.Tokens, f.createToken(expandedText, tokenType, style))
			f.currentColumn = col
		}
	} else {
		f.currentColumn = 0
	}
	return nextLine, true
}

func (f *customFormatter) Format(tokens []chroma.Token, style *chroma.Style) error {
	f.Result.Lines = make([]Line, 0)
	f.currentLine = Line{}
	f.lineNumber = 1
	f.currentColumn = 0
	f.highlightedLines = make(map[int]bool)
	for _, line := range f.Result.HighlightedLines {
		f.highlightedLines[line] = true
	}

	for _, token := range tokens {
		text := token.String()

		// Handle newlines
		if newLine, handled := f.processNewlines(text, token.Type, style); handled {
			f.currentLine = newLine
			continue
		}

		// Add token (handles tabs internally)
		f.addToken(text, token.Type, style)
	}

	// Add any remaining line
	if len(f.currentLine.Tokens) > 0 {
		f.addLine(f.currentLine)
	}

	return nil
}

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

func getHighlightColor(style *chroma.Style) color.Color {
	// Use LineHighlight token type for line highlighting
	highlightColor := style.Get(chroma.LineHighlight)

	// If the style doesn't define LineHighlight, create a semi-transparent highlight
	if highlightColor.Background == 0 && highlightColor.Colour == 0 {
		baseColor := style.Get(chroma.Background)
		// Make the highlight slightly lighter/darker than the background
		if isLight(baseColor.Background) {
			return color.NRGBA{R: 0, G: 0, B: 0, A: 128} // More transparent (darker)
		} else {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 128} // More transparent (lighter)
		}
	}

	// Use the style's LineHighlight color
	if highlightColor.Background != 0 {
		return color.NRGBA{
			R: highlightColor.Background.Red(),
			G: highlightColor.Background.Green(),
			B: highlightColor.Background.Blue(),
			A: 128, // Semi-transparent (adjust this value between 0-255 to control opacity)
		}
	}

	// Fallback to foreground color if background is not set
	return color.NRGBA{
		R: highlightColor.Colour.Red(),
		G: highlightColor.Colour.Green(),
		B: highlightColor.Colour.Blue(),
		A: 128, // Semi-transparent (adjust this value between 0-255 to control opacity)
	}
}

func isLight(c chroma.Colour) bool {
	// Convert RGB values to 0-255 range
	r := float64(c.Red())
	g := float64(c.Green())
	b := float64(c.Blue())

	// Calculate perceived brightness (ITU-R BT.709)
	brightness := (r*0.299 + g*0.587 + b*0.114)

	// Consider the color light if brightness is greater than 128 (half of 255)
	return brightness > 128
}

func getColorFromChroma(c chroma.Colour) color.Color {
	return color.RGBA{
		R: c.Red(),
		G: c.Green(),
		B: c.Blue(),
		A: 255,
	}
}
