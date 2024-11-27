package code

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

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
	Lines            []Line      // The lines of code with their tokens
	BackgroundColor  color.Color // Background color for the code block
	GutterColor      color.Color // Color for the gutter (line numbers background)
	LineNumberColor  color.Color // Color for line numbers
	HighlightColor   color.Color // Color for highlighted lines
	HighlightedLines []int       // Lines that should be highlighted
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
func Highlight(code string, opts *CodeStyle) (*HighlightedCode, error) {
	if opts == nil {
		return nil, fmt.Errorf("no options provided")
	}

	// Use Chroma for syntax highlighting
	var lexer chroma.Lexer
	if opts.Language != "" {
		lexer = lexers.Get(opts.Language)
		if lexer == nil {
			return nil, fmt.Errorf("no lexer found for language: %s", opts.Language)
		}
	} else {
		lexer = lexers.Analyse(code)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}

	// Get the style
	style := styles.Get(opts.Theme)
	if style == nil {
		style = styles.Fallback
	}

	// Create a custom formatter
	backgroundColor := getBackgroundColor(style)
	gutterColor := getGutterColor(style)
	lineNumberColor := getLineNumberColor(style)
	highlightColor := getHighlightColor(style)

	formatter := &customFormatter{
		highlightedLines: make(map[int]bool),
		tabWidth:         opts.TabWidth,
		Result: &HighlightedCode{
			BackgroundColor: backgroundColor,
			GutterColor:     gutterColor,
			LineNumberColor: lineNumberColor,
			HighlightColor:  highlightColor,
		},
	}

	// Set up highlighted lines
	if len(opts.LineHighlightRanges) > 0 {
		ranges := opts.LineHighlightRanges
		for _, rangePair := range ranges {
			// Convert 1-based line numbers to 0-based for internal use
			start := rangePair.Start - 1
			end := rangePair.End - 1
			for i := start; i <= end; i++ {
				formatter.highlightedLines[i] = true
				formatter.Result.HighlightedLines = append(formatter.Result.HighlightedLines, i+1)
			}
		}
	}

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, fmt.Errorf("error tokenizing code: %v", err)
	}

	// Format the tokens
	err = formatter.Format(iterator.Tokens(), style)
	if err != nil {
		return nil, fmt.Errorf("error formatting tokens: %v", err)
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
