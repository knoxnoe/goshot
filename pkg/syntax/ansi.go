package syntax

import (
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/styles"
)

// ANSI color codes
const (
	ansiReset     = 0
	ansiBold      = 1
	ansiItalic    = 3
	ansiUnderline = 4
	ansiFgBlack   = 30
	ansiFgRed     = 31
	ansiFgGreen   = 32
	ansiFgYellow  = 33
	ansiFgBlue    = 34
	ansiFgMagenta = 35
	ansiFgCyan    = 36
	ansiFgWhite   = 37
	ansiBgBlack   = 40
	ansiBgRed     = 41
	ansiBgGreen   = 42
	ansiBgYellow  = 43
	ansiBgBlue    = 44
	ansiBgMagenta = 45
	ansiBgCyan    = 46
	ansiBgWhite   = 47
	// Bright foreground colors
	ansiFgBrightBlack   = 90
	ansiFgBrightRed     = 91
	ansiFgBrightGreen   = 92
	ansiFgBrightYellow  = 93
	ansiFgBrightBlue    = 94
	ansiFgBrightMagenta = 95
	ansiFgBrightCyan    = 96
	ansiFgBrightWhite   = 97
	// Bright background colors
	ansiBgBrightBlack   = 100
	ansiBgBrightRed     = 101
	ansiBgBrightGreen   = 102
	ansiBgBrightYellow  = 103
	ansiBgBrightBlue    = 104
	ansiBgBrightMagenta = 105
	ansiBgBrightCyan    = 106
	ansiBgBrightWhite   = 107
)

// Regular expression to match ANSI escape sequences
var (
	ansiRegex = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

	// Combined regex for all non-color sequences
	nonColorRegex = regexp.MustCompile(
		`\x1b[PD](?:[^\x1b]|\x1b[^\\])*\x1b\\|` + // DCS
			`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|` + // OSC
			`\x1b\[[0-9;]*[ABCDEFGHJKSTfnsu]|` + // Cursor
			`\x1b\[[0-9;]*[hli]`) // Mode

	// Color caches
	basicColorCache = make(map[int]map[bool]color.Color)
	color256Cache   = make(map[int]map[bool]color.Color)
	rgbColorCache   = make(map[uint32]color.Color)
)

// cacheKey creates a unique key for RGB colors
func rgbCacheKey(r, g, b uint8) uint32 {
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

// getCachedBasicColor returns a cached basic ANSI color
func getCachedBasicColor(index int, isLightTheme bool) color.Color {
	if cache, ok := basicColorCache[index]; ok {
		if c, ok := cache[isLightTheme]; ok {
			return c
		}
	}
	return nil
}

// getCached256Color returns a cached 256-color
func getCached256Color(index int, isLightTheme bool) color.Color {
	if cache, ok := color256Cache[index]; ok {
		if c, ok := cache[isLightTheme]; ok {
			return c
		}
	}
	return nil
}

// getCachedRGBColor returns a cached RGB color
func getCachedRGBColor(r, g, b uint8) color.Color {
	if c, ok := rgbColorCache[rgbCacheKey(r, g, b)]; ok {
		return c
	}
	return nil
}

// cacheBasicColor stores a basic ANSI color in the cache
func cacheBasicColor(index int, isLightTheme bool, c color.Color) {
	if basicColorCache[index] == nil {
		basicColorCache[index] = make(map[bool]color.Color)
	}
	basicColorCache[index][isLightTheme] = c
}

// cache256Color stores a 256-color in the cache
func cache256Color(index int, isLightTheme bool, c color.Color) {
	if color256Cache[index] == nil {
		color256Cache[index] = make(map[bool]color.Color)
	}
	color256Cache[index][isLightTheme] = c
}

// cacheRGBColor stores an RGB color in the cache
func cacheRGBColor(r, g, b uint8, c color.Color) {
	rgbColorCache[rgbCacheKey(r, g, b)] = c
}

// Starship-like prompt colors
var (
	promptArrowColorLight = color.RGBA{R: 214, G: 0, B: 143, A: 255}   // Light theme pink
	promptArrowColorDark  = color.RGBA{R: 255, G: 121, B: 198, A: 255} // Dark theme pink
	promptCmdColorLight   = color.RGBA{R: 34, G: 197, B: 94, A: 255}   // Light theme green
	promptCmdColorDark    = color.RGBA{R: 80, G: 250, B: 123, A: 255}  // Dark theme green
)

// ParseANSI parses text with ANSI escape sequences and returns HighlightedCode
func ParseANSI(text string, opts *HighlightOptions) (*HighlightedCode, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Get Chroma style for theme colors
	style := styles.Get(opts.Style)
	if style == nil {
		style = styles.Fallback
	}

	result := &HighlightedCode{
		Lines:            make([]Line, 0),
		BackgroundColor:  getBackgroundColor(style),
		GutterColor:      getGutterColor(style),
		LineNumberColor:  getLineNumberColor(style),
		HighlightColor:   getHighlightColor(style),
		HighlightedLines: opts.HighlightedLines,
	}

	// Split text into lines
	lines := strings.Split(text, "\n")
	currentState := newANSIState()

	// Use theme background color for default background
	currentState.bgColor = result.BackgroundColor
	currentState.isLightTheme = isLightColor(result.BackgroundColor)

	// Set default text color based on theme
	if currentState.isLightTheme {
		currentState.fgColor = color.Black
	} else {
		currentState.fgColor = color.White
	}

	// Add prompt if requested
	if opts.ShowPrompt {
		// Create minimal prompt line
		var arrowColor, cmdColor color.Color
		if currentState.isLightTheme {
			arrowColor = promptArrowColorLight
			cmdColor = promptCmdColorLight
		} else {
			arrowColor = promptArrowColorDark
			cmdColor = promptCmdColorDark
		}

		promptLine := Line{
			Tokens: []Token{
				{Text: "❯ ", Color: arrowColor},
				{Text: opts.PromptCommand, Color: cmdColor},
			},
		}
		result.Lines = append(result.Lines, promptLine)
	}

	for lineNum, lineText := range lines {
		lineText = filterNonColorEscapes(lineText)
		line := parseLine(lineText, currentState)
		result.Lines = append(result.Lines, line)
		if opts.HighlightedLines != nil {
			for _, hl := range opts.HighlightedLines {
				if hl == lineNum+1 {
					result.Lines[len(result.Lines)-1].Highlight = true
					break
				}
			}
		}
	}

	return result, nil
}

// ansiState keeps track of the current ANSI formatting state
type ansiState struct {
	fgColor      color.Color
	bgColor      color.Color
	bold         bool
	italic       bool
	underline    bool
	isLightTheme bool
}

func newANSIState() *ansiState {
	return &ansiState{
		fgColor:      color.Black,
		bgColor:      color.White,
		bold:         false,
		italic:       false,
		underline:    false,
		isLightTheme: true,
	}
}

// Clone creates a copy of the current state
func (s *ansiState) Clone() *ansiState {
	return &ansiState{
		fgColor:      s.fgColor,
		bgColor:      s.bgColor,
		bold:         s.bold,
		italic:       s.italic,
		underline:    s.underline,
		isLightTheme: s.isLightTheme,
	}
}

// Reset resets the state to defaults
func (s *ansiState) Reset() {
	s.fgColor = color.Black
	s.bgColor = color.White
	s.bold = false
	s.italic = false
	s.underline = false
}

// parseLine parses a single line of text with ANSI escape sequences
func parseLine(text string, state *ansiState) Line {
	var tokens []Token
	currentState := state.Clone()
	lastIndex := 0

	// Find all ANSI escape sequences in the text
	matches := ansiRegex.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// Add text before the escape sequence
		if match[0] > lastIndex {
			tokens = append(tokens, Token{
				Text:   text[lastIndex:match[0]],
				Color:  currentState.fgColor,
				Bold:   currentState.bold,
				Italic: currentState.italic,
			})
		}

		// Parse and apply the escape sequence
		if match[2] < match[3] { // If we have parameters
			paramStr := text[match[2]:match[3]]
			params := parseParams(paramStr)
			currentState.Apply(params)
		} else {
			currentState.Reset()
		}

		lastIndex = match[1]
	}

	// Add remaining text
	if lastIndex < len(text) {
		tokens = append(tokens, Token{
			Text:   text[lastIndex:],
			Color:  currentState.fgColor,
			Bold:   currentState.bold,
			Italic: currentState.italic,
		})
	}

	return Line{Tokens: tokens}
}

// parseParams converts ANSI parameter string to slice of integers
func parseParams(s string) []int {
	if s == "" {
		return []int{0}
	}

	parts := strings.Split(s, ";")
	params := make([]int, len(parts))

	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		params[i] = n
	}

	return params
}

// basicANSIColor returns a color.Color for a basic ANSI color index (0-7)
func basicANSIColor(index int, isLightTheme bool) color.Color {
	if c := getCachedBasicColor(index, isLightTheme); c != nil {
		return c
	}

	var c color.Color
	if isLightTheme {
		switch index {
		case 0:
			c = color.Black // Black
		case 1:
			c = color.RGBA{R: 205, G: 0, B: 0, A: 255} // Red
		case 2:
			c = color.RGBA{R: 0, G: 205, B: 0, A: 255} // Green
		case 3:
			c = color.RGBA{R: 205, G: 205, B: 0, A: 255} // Yellow
		case 4:
			c = color.RGBA{R: 48, G: 48, B: 238, A: 255} // Blue (more muted)
		case 5:
			c = color.RGBA{R: 205, G: 0, B: 205, A: 255} // Magenta
		case 6:
			c = color.RGBA{R: 0, G: 205, B: 205, A: 255} // Cyan
		case 7:
			c = color.RGBA{R: 229, G: 229, B: 229, A: 255} // White
		default:
			c = color.Black
		}
	} else {
		switch index {
		case 0:
			c = color.RGBA{R: 98, G: 114, B: 164, A: 255} // Black (lighter blue-gray)
		case 1:
			c = color.RGBA{R: 255, G: 85, B: 85, A: 255} // Red
		case 2:
			c = color.RGBA{R: 80, G: 250, B: 123, A: 255} // Green
		case 3:
			c = color.RGBA{R: 255, G: 255, B: 85, A: 255} // Yellow
		case 4:
			c = color.RGBA{R: 98, G: 114, B: 164, A: 255} // Blue (blue-gray)
		case 5:
			c = color.RGBA{R: 255, G: 121, B: 198, A: 255} // Magenta
		case 6:
			c = color.RGBA{R: 139, G: 233, B: 253, A: 255} // Cyan
		case 7:
			c = color.White // White
		default:
			c = color.RGBA{R: 98, G: 114, B: 164, A: 255}
		}
	}

	cacheBasicColor(index, isLightTheme, c)
	return c
}

// color256 returns a color.Color for an 8-bit color index (0-255)
func color256(index int, isLightTheme bool) color.Color {
	if c := getCached256Color(index, isLightTheme); c != nil {
		return c
	}

	var c color.Color

	// Basic 16 colors (0-15)
	if index < 16 {
		if index < 8 {
			c = basicANSIColor(index, isLightTheme)
		} else {
			// Bright versions of the basic colors (8-15)
			c = basicANSIColor(index-8, !isLightTheme) // Use opposite theme colors for bright variants
		}
		cache256Color(index, isLightTheme, c)
		return c
	}

	// Special case for index 103 (blue-gray in dark theme)
	if index == 103 && !isLightTheme {
		c = color.RGBA{R: 98, G: 114, B: 164, A: 255}
		cache256Color(index, isLightTheme, c)
		return c
	}

	// 216 colors (16-231): 6×6×6 cube
	if index < 232 {
		index -= 16
		r := uint8(((index / 36) % 6) * 51)
		g := uint8(((index / 6) % 6) * 51)
		b := uint8((index % 6) * 51)
		c = color.RGBA{R: r, G: g, B: b, A: 255}
		cache256Color(index+16, isLightTheme, c)
		return c
	}

	// Grayscale (232-255): 24 shades
	index -= 232
	v := uint8(8 + index*10)
	if !isLightTheme {
		// For dark themes, make grays slightly blue-tinted
		c = color.RGBA{R: v - 20, G: v - 10, B: v, A: 255}
	} else {
		c = color.RGBA{R: v, G: v, B: v, A: 255}
	}
	cache256Color(index+232, isLightTheme, c)
	return c
}

// Apply applies ANSI parameters to the current state
func (s *ansiState) Apply(params []int) {
	if len(params) == 0 {
		s.Reset()
		return
	}

	for i := 0; i < len(params); i++ {
		param := params[i]
		switch {
		case param == ansiReset:
			s.Reset()
		case param == ansiBold:
			s.bold = true
		case param == ansiItalic:
			s.italic = true
		case param == ansiUnderline:
			s.underline = true
		case param >= ansiFgBlack && param <= ansiFgWhite:
			s.fgColor = basicANSIColor(param-ansiFgBlack, s.isLightTheme)
		case param >= ansiBgBlack && param <= ansiBgWhite:
			s.bgColor = basicANSIColor(param-ansiBgBlack, s.isLightTheme)
		case param >= ansiFgBrightBlack && param <= ansiFgBrightWhite:
			s.fgColor = basicANSIColor(param-ansiFgBrightBlack, !s.isLightTheme)
		case param >= ansiBgBrightBlack && param <= ansiBgBrightWhite:
			s.bgColor = basicANSIColor(param-ansiBgBrightBlack, !s.isLightTheme)
		case param == 38: // 8-bit or 24-bit foreground color
			if i+2 < len(params) && params[i+1] == 5 { // 8-bit color
				s.fgColor = color256(params[i+2], s.isLightTheme)
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 { // 24-bit color
				r, g, b := uint8(params[i+2]), uint8(params[i+3]), uint8(params[i+4])
				if c := getCachedRGBColor(r, g, b); c != nil {
					s.fgColor = c
				} else {
					c = color.RGBA{R: r, G: g, B: b, A: 255}
					cacheRGBColor(r, g, b, c)
					s.fgColor = c
				}
				i += 4
			}
		case param == 48: // 8-bit or 24-bit background color
			if i+2 < len(params) && params[i+1] == 5 { // 8-bit color
				s.bgColor = color256(params[i+2], s.isLightTheme)
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 { // 24-bit color
				r, g, b := uint8(params[i+2]), uint8(params[i+3]), uint8(params[i+4])
				if c := getCachedRGBColor(r, g, b); c != nil {
					s.bgColor = c
				} else {
					c = color.RGBA{R: r, G: g, B: b, A: 255}
					cacheRGBColor(r, g, b, c)
					s.bgColor = c
				}
				i += 4
			}
		}
	}
}

// isLight determines if a color is light based on its perceived brightness
func isLightColor(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	// Convert from 0-65535 to 0-255 range
	r = r >> 8
	g = g >> 8
	b = b >> 8
	// Calculate perceived brightness using the formula from W3C
	// Perceived brightness = (R * 299 + G * 587 + B * 114) / 1000
	brightness := (299*r + 587*g + 114*b) / 1000
	return brightness > 128
}

// filterNonColorEscapes removes any ANSI escape sequences that aren't related to colors
func filterNonColorEscapes(text string) string {
	return nonColorRegex.ReplaceAllString(text, "")
}
