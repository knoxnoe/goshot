package syntax

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/math/fixed"
)

// RenderConfig holds configuration for rendering highlighted code to an image
type RenderConfig struct {
	FontSize      float64
	LineHeight    float64
	PaddingLeft   int
	PaddingRight  int
	PaddingTop    int
	PaddingBottom int
	FontFace      font.Face
	Font          *truetype.Font
	Background    image.Image
	TabWidth      int // Width of tab characters in spaces
	MinWidth      int // Minimum width in pixels (0 means no minimum)
	MaxWidth      int // Maximum width in pixels (0 means no limit)

	// Line number settings
	ShowLineNumbers   bool
	LineNumberColor   color.Color
	LineNumberPadding int         // Padding on either side of line numbers in pixels
	LineNumberBg      color.Color // Background color for line numbers
	StartLineNumber   int         // Line number to start from
	EndLineNumber     int         // Line number to end at

	// Line highlighting settings
	LineHighlightColor color.Color // Color for highlighted lines
}

// DefaultConfig returns a default rendering configuration
func DefaultConfig() *RenderConfig {
	f, _ := truetype.Parse(gomono.TTF)
	size := 14.0
	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
		DPI:  72,
	})
	return &RenderConfig{
		LineHeight:    1.5,
		PaddingLeft:   10,
		PaddingRight:  10,
		PaddingTop:    10,
		PaddingBottom: 10,
		FontFace:      face,
		Font:          f,
		FontSize:      size,
		TabWidth:      4,    // Default 4 spaces per tab
		MinWidth:      200,  // Minimum width of 200px
		MaxWidth:      1460, // Maximum width for 120 characters

		// Line number defaults
		ShowLineNumbers:   true,
		LineNumberColor:   color.RGBA{R: 128, G: 128, B: 128, A: 255}, // Gray color
		LineNumberPadding: 10,
		LineNumberBg:      color.RGBA{R: 245, G: 245, B: 245, A: 255}, // Light gray background
		StartLineNumber:   1,
		EndLineNumber:     0,

		// Line highlighting defaults
		LineHighlightColor: color.RGBA{R: 68, G: 68, B: 68, A: 40}, // Semi-transparent dark color
	}
}

// Getters and setters for RenderConfig
func (c *RenderConfig) GetLineHeight() float64 { return c.LineHeight }
func (c *RenderConfig) SetLineHeight(height float64) *RenderConfig {
	c.LineHeight = height
	return c
}

func (c *RenderConfig) GetPaddingLeft() int { return c.PaddingLeft }
func (c *RenderConfig) SetPaddingLeft(padding int) *RenderConfig {
	c.PaddingLeft = padding
	return c
}

func (c *RenderConfig) GetPaddingRight() int { return c.PaddingRight }
func (c *RenderConfig) SetPaddingRight(padding int) *RenderConfig {
	c.PaddingRight = padding
	return c
}

func (c *RenderConfig) GetPaddingTop() int { return c.PaddingTop }
func (c *RenderConfig) SetPaddingTop(padding int) *RenderConfig {
	c.PaddingTop = padding
	return c
}

func (c *RenderConfig) GetPaddingBottom() int { return c.PaddingBottom }
func (c *RenderConfig) SetPaddingBottom(padding int) *RenderConfig {
	c.PaddingBottom = padding
	return c
}

// SetFontFace sets both the font face and underlying TrueType font
func (c *RenderConfig) SetFontFace(face font.Face, f *truetype.Font, size float64) *RenderConfig {
	c.FontFace = face
	c.Font = f
	c.FontSize = size
	return c
}

// GetFontFace returns the current font face
func (c *RenderConfig) GetFontFace() font.Face {
	return c.FontFace
}

func (c *RenderConfig) GetBackground() image.Image { return c.Background }
func (c *RenderConfig) SetBackground(bg image.Image) *RenderConfig {
	c.Background = bg
	return c
}

func (c *RenderConfig) GetTabWidth() int { return c.TabWidth }
func (c *RenderConfig) SetTabWidth(width int) *RenderConfig {
	c.TabWidth = width
	return c
}

func (c *RenderConfig) GetMinWidth() int { return c.MinWidth }
func (c *RenderConfig) SetMinWidth(width int) *RenderConfig {
	c.MinWidth = width
	// Ensure MaxWidth is not less than MinWidth if both are set
	if c.MaxWidth > 0 && c.MaxWidth < width {
		c.MaxWidth = width
	}
	return c
}

func (c *RenderConfig) GetMaxWidth() int { return c.MaxWidth }
func (c *RenderConfig) SetMaxWidth(width int) *RenderConfig {
	// Ensure MaxWidth is not less than MinWidth if both are set
	if width > 0 && c.MinWidth > 0 && width < c.MinWidth {
		width = c.MinWidth
	}
	c.MaxWidth = width
	return c
}

// Line number settings
func (c *RenderConfig) GetShowLineNumbers() bool { return c.ShowLineNumbers }
func (c *RenderConfig) SetShowLineNumbers(show bool) *RenderConfig {
	c.ShowLineNumbers = show
	return c
}

func (c *RenderConfig) GetLineNumberColor() color.Color { return c.LineNumberColor }
func (c *RenderConfig) SetLineNumberColor(col color.Color) *RenderConfig {
	c.LineNumberColor = col
	return c
}

func (c *RenderConfig) GetLineNumberPadding() int { return c.LineNumberPadding }
func (c *RenderConfig) SetLineNumberPadding(padding int) *RenderConfig {
	c.LineNumberPadding = padding
	return c
}

func (c *RenderConfig) GetLineNumberBg() color.Color { return c.LineNumberBg }
func (c *RenderConfig) SetLineNumberBg(bg color.Color) *RenderConfig {
	c.LineNumberBg = bg
	return c
}

func (c *RenderConfig) GetStartLineNumber() int { return c.StartLineNumber }
func (c *RenderConfig) SetStartLineNumber(line int) *RenderConfig {
	c.StartLineNumber = line
	return c
}

func (c *RenderConfig) GetEndLineNumber() int { return c.EndLineNumber }
func (c *RenderConfig) SetEndLineNumber(line int) *RenderConfig {
	c.EndLineNumber = line
	return c
}

func (c *RenderConfig) GetLineHighlightColor() color.Color { return c.LineHighlightColor }
func (c *RenderConfig) SetLineHighlightColor(col color.Color) *RenderConfig {
	c.LineHighlightColor = col
	return c
}

// WithFont is a convenience method to set the font face from TTF data
func (c *RenderConfig) WithFont(ttfData []byte) (*RenderConfig, error) {
	font, err := truetype.Parse(ttfData)
	if err != nil {
		return c, fmt.Errorf("failed to parse font: %v", err)
	}
	return c.SetFontFace(truetype.NewFace(font, &truetype.Options{
		Size: 14,
		DPI:  72,
	}), font, 14.0), nil
}

// Clone creates a deep copy of the RenderConfig
func (c *RenderConfig) Clone() *RenderConfig {
	clone := *c // shallow copy

	// Deep copy any pointer or interface fields
	if c.FontFace != nil {
		clone.FontFace = c.FontFace // Font is immutable, so pointer copy is safe
	}
	if c.Font != nil {
		clone.Font = c.Font // Font is immutable, so pointer copy is safe
	}
	if c.Background != nil {
		switch bg := c.Background.(type) {
		case *image.RGBA:
			newBg := *bg // copy the struct
			clone.Background = &newBg
		case *image.Uniform:
			newBg := *bg // copy the struct
			clone.Background = &newBg
		default:
			// For other image types, create a new RGBA image
			bounds := bg.Bounds()
			newBg := image.NewRGBA(bounds)
			draw.Draw(newBg, bounds, bg, bounds.Min, draw.Src)
			clone.Background = newBg
		}
	}

	return &clone
}

// GetMonospaceWidth calculates the width needed for a given number of characters
func (c *RenderConfig) GetMonospaceWidth(charCount int) int {
	// Measure a single character (using 'M' as reference)
	charWidth := font.MeasureString(c.FontFace, "M").Round()
	return (charWidth * charCount) + (c.PaddingLeft + c.PaddingRight)
}

// wrapTokens splits tokens into multiple lines if they exceed maxWidth
func wrapTokens(tokens []Token, face font.Face, maxWidth, startX int) [][]Token {
	if maxWidth <= 0 {
		return [][]Token{tokens}
	}

	var result [][]Token
	var currentLine []Token
	currentWidth := startX

	for _, token := range tokens {
		tokenWidth := font.MeasureString(face, token.Text).Round()

		// If this token would exceed max width, start a new line
		if currentWidth+tokenWidth > maxWidth && len(currentLine) > 0 {
			result = append(result, currentLine)
			currentLine = nil
			currentWidth = startX
		}

		// If a single token is wider than maxWidth, split it
		if tokenWidth > maxWidth-startX {
			parts := splitToken(token, face, maxWidth-startX)
			for _, part := range parts {
				if len(currentLine) > 0 {
					result = append(result, currentLine)
					currentLine = nil
				}
				result = append(result, []Token{part})
			}
			currentWidth = startX
			continue
		}

		currentLine = append(currentLine, token)
		currentWidth += tokenWidth
	}

	if len(currentLine) > 0 {
		result = append(result, currentLine)
	}

	return result
}

// splitToken splits a single token into multiple tokens that fit within maxWidth
func splitToken(token Token, face font.Face, maxWidth int) []Token {
	var result []Token
	text := token.Text
	for len(text) > 0 {
		i := len(text)
		width := font.MeasureString(face, text[:i]).Round()

		// Binary search for the maximum substring that fits
		for width > maxWidth {
			i = i / 2
			width = font.MeasureString(face, text[:i]).Round()
		}

		// Try to break at word boundary if possible
		if i < len(text) {
			if spaceIdx := strings.LastIndex(text[:i], " "); spaceIdx > 0 {
				i = spaceIdx + 1
			}
		}

		// Create new token with the split text
		result = append(result, Token{
			Text:   text[:i],
			Color:  token.Color,
			Bold:   token.Bold,
			Italic: token.Italic,
		})

		text = text[i:]
		// Trim leading spaces from remaining text
		text = strings.TrimLeft(text, " ")
	}

	return result
}

// RenderToImage converts highlighted code to an image
func (h *HighlightedCode) RenderToImage(config *RenderConfig) (image.Image, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Calculate dimensions
	wrappedLines := [][]Token{}
	maxLineWidth := 0

	// Get lines
	lines := h.Lines

	// Validate that the requested line range is within bounds
	if config.StartLineNumber > len(lines) {
		return nil, fmt.Errorf("start line number %d is out of bounds (max: %d)", config.StartLineNumber, len(lines))
	}
	if config.EndLineNumber > len(lines) {
		return nil, fmt.Errorf("end line number %d is out of bounds (max: %d)", config.EndLineNumber, len(lines))
	}

	if config.EndLineNumber > 0 {
		startLineNumber := config.StartLineNumber
		if startLineNumber < 1 {
			startLineNumber = 1
		}

		// Create a new slice to hold our lines
		newLines := make([]Line, 0)

		// Add ellipses line before if we're not starting from the beginning
		if startLineNumber > 1 {
			newLines = append(newLines, Line{Tokens: []Token{{Text: "...", Color: config.LineNumberColor}}})
		}

		// Add the actual code lines
		newLines = append(newLines, lines[startLineNumber-1:config.EndLineNumber]...)

		// Add ellipses line after if we're not ending at the last line
		if config.EndLineNumber < len(lines) {
			newLines = append(newLines, Line{Tokens: []Token{{Text: "...", Color: config.LineNumberColor}}})
		}

		lines = newLines
	}

	// Calculate line number width if needed
	lineNumberOffset := 0
	if config.ShowLineNumbers {
		// Calculate width needed for the largest line number
		lineCount := len(lines)
		maxDigits := len(fmt.Sprintf("%d", lineCount))

		// Add one extra digit width for better scaling at larger font sizes
		maxDigits++

		// Calculate base width for digits
		lineNumberWidth := font.MeasureString(config.FontFace, strings.Repeat("0", maxDigits)).Round()

		// Scale padding with font size
		scaledPadding := int(float64(config.LineNumberPadding) * (float64(config.FontFace.Metrics().Height.Round()) / 14.0))

		if scaledPadding < config.LineNumberPadding {
			scaledPadding = config.LineNumberPadding
		}

		lineNumberOffset = lineNumberWidth + (scaledPadding * 2)
	}

	// Calculate max text width (total width minus padding and line numbers)
	maxTextWidth := 0
	if config.MaxWidth > 0 {
		maxTextWidth = config.MaxWidth - (config.PaddingLeft + config.PaddingRight)
		if config.ShowLineNumbers {
			maxTextWidth -= lineNumberOffset
		}
	}

	// Wrap lines and calculate max width
	for _, line := range lines {
		var wrapped [][]Token
		if len(line.Tokens) > 0 {
			wrapped = wrapTokens(line.Tokens, config.FontFace, maxTextWidth, 0)
		} else {
			// For empty lines, add an empty token list
			wrapped = [][]Token{{}}
		}
		wrappedLines = append(wrappedLines, wrapped...)

		// Calculate max line width
		for _, wline := range wrapped {
			lineWidth := 0
			for _, token := range wline {
				lineWidth += font.MeasureString(config.FontFace, token.Text).Round()
			}
			if lineWidth > maxLineWidth {
				maxLineWidth = lineWidth
			}
		}
	}

	// Calculate final image dimensions
	codeWidth := maxLineWidth + (config.PaddingLeft + config.PaddingRight)

	// Apply min/max width constraints
	if config.MinWidth > 0 && codeWidth < config.MinWidth {
		codeWidth = config.MinWidth
	}
	if config.MaxWidth > 0 && codeWidth > config.MaxWidth {
		codeWidth = config.MaxWidth
	}

	totalWidth := codeWidth
	if config.ShowLineNumbers {
		totalWidth += lineNumberOffset
	}

	// Calculate total height
	metrics := config.FontFace.Metrics()
	lineHeight := int(float64(metrics.Height.Round()) * config.LineHeight)
	totalHeight := (lineHeight * len(wrappedLines)) + (config.PaddingTop + config.PaddingBottom)

	// Create the image
	img := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	// Fill background with theme color
	bgColor := h.BackgroundColor
	if bgColor == nil {
		bgColor = color.White
	}
	for y := 0; y < totalHeight; y++ {
		for x := 0; x < totalWidth; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw line highlights
	for i, line := range lines {
		if line.Highlight {
			y := config.PaddingTop + (i * lineHeight)
			highlightRect := image.Rect(
				lineNumberOffset,
				y,
				totalWidth,
				y+lineHeight,
			)
			uniform := image.NewUniform(config.LineHighlightColor)
			draw.Draw(img, highlightRect, uniform, image.Point{}, draw.Over)
		}
	}

	// Draw line numbers if enabled
	if config.ShowLineNumbers {
		// Draw gutter background
		gutterBg := h.GutterColor
		if gutterBg == nil {
			gutterBg = config.LineNumberBg
		}
		if gutterBg != nil {
			for y := 0; y < totalHeight; y++ {
				for x := 0; x < lineNumberOffset; x++ {
					img.Set(x, y, gutterBg)
				}
			}
		}

		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFont(config.Font)
		c.SetClip(img.Bounds())
		c.SetDst(img)

		// Use theme's line number color if available
		lineNumColor := h.LineNumberColor
		if lineNumColor == nil {
			lineNumColor = config.LineNumberColor
		}
		c.SetSrc(image.NewUniform(lineNumColor))

		for i := range wrappedLines {
			startLine := config.StartLineNumber
			if startLine < 1 {
				startLine = 1
			}
			lineNum := fmt.Sprintf("%d", startLine+i)
			// Calculate position for right-aligned number
			lineNumWidth := font.MeasureString(config.FontFace, lineNum).Round()
			x := lineNumberOffset - config.LineNumberPadding - lineNumWidth
			y := config.PaddingTop + ((i + 1) * lineHeight) - (metrics.Descent.Round() * 2)
			pt := freetype.Pt(x, y)
			c.DrawString(lineNum, pt)
		}
	}

	// Create context for drawing text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(config.Font)
	c.SetClip(img.Bounds())
	c.SetDst(img)

	// Draw each line
	for i, line := range wrappedLines {
		// Calculate baseline Y position
		y := config.PaddingTop + ((i + 1) * lineHeight) - (metrics.Descent.Round() * 2)

		// Draw code
		x := config.PaddingLeft
		if config.ShowLineNumbers {
			x += lineNumberOffset
		}

		// Check if we're using a monospace font
		isMono := isMonospace(config.FontFace)

		// Draw each token
		for _, token := range line {
			c.SetSrc(image.NewUniform(token.Color))
			pt := freetype.Pt(x, y)

			if isMono {
				// For monospace fonts, use fixed character spacing
				charWidth := font.MeasureString(config.FontFace, "M").Round()
				for _, ch := range token.Text {
					c.DrawString(string(ch), pt)
					pt.X += fixed.Int26_6(charWidth << 6)
				}
				x += charWidth * len([]rune(token.Text))
			} else {
				// For proportional fonts, use natural spacing
				c.DrawString(token.Text, pt)
				x += font.MeasureString(config.FontFace, token.Text).Round()
			}
		}
	}

	return img, nil
}

// isMonospace checks if the font is monospace by sampling character widths
func isMonospace(fontFace font.Face) bool {
	isMonospace := true
	samples := []rune{'M', 'i', '.', ' ', 'W'}
	width := font.MeasureString(fontFace, string(samples[0])).Round()
	for _, ch := range samples[1:] {
		if font.MeasureString(fontFace, string(ch)).Round() != width {
			isMonospace = false
			break
		}
	}
	return isMonospace
}
