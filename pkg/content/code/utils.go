package code

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/watzon/goshot/pkg/content"
	"golang.org/x/image/font"
)

func getBackgroundColor(style *chroma.Style) color.Color {
	bgColor := style.Get(chroma.Background)
	if bgColor.Background != 0 {
		return color.RGBA{
			R: bgColor.Background.Red(),
			G: bgColor.Background.Green(),
			B: bgColor.Background.Blue(),
			A: 255,
		}
	}
	if bgColor.Colour != 0 {
		return color.RGBA{
			R: bgColor.Colour.Red(),
			G: bgColor.Colour.Green(),
			B: bgColor.Colour.Blue(),
			A: 255,
		}
	}
	// Default to white background if none is provided
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func getGutterColor(style *chroma.Style) color.Color {
	lineTableColor := style.Get(chroma.LineTable)
	if lineTableColor.Background == 0 {
		// If no gutter color is provided, make it slightly different from the background
		bg := getBackgroundColor(style)
		bgRGBA := bg.(color.RGBA)
		if isLight(chroma.Colour(uint32(bgRGBA.R)<<16 | uint32(bgRGBA.G)<<8 | uint32(bgRGBA.B))) {
			// For light backgrounds, make gutter slightly darker
			return color.RGBA{
				R: maxu8(0, bgRGBA.R-20),
				G: maxu8(0, bgRGBA.G-20),
				B: maxu8(0, bgRGBA.B-20),
				A: 255,
			}
		} else {
			// For dark backgrounds, make gutter slightly lighter
			return color.RGBA{
				R: minu8(255, bgRGBA.R+20),
				G: minu8(255, bgRGBA.G+20),
				B: minu8(255, bgRGBA.B+20),
				A: 255,
			}
		}
	}
	return color.RGBA{
		R: lineTableColor.Background.Red(),
		G: lineTableColor.Background.Green(),
		B: lineTableColor.Background.Blue(),
		A: 255,
	}
}

func getLineNumberColor(style *chroma.Style) color.Color {
	lineNumColor := style.Get(chroma.LineNumbers)
	if lineNumColor.Colour == 0 {
		// If no line number color is provided, make it a muted version of the text color
		textColor := style.Get(chroma.Text)
		if textColor.Colour == 0 {
			// If no text color either, base it on background
			bg := getBackgroundColor(style)
			bgRGBA := bg.(color.RGBA)
			if isLight(chroma.Colour(uint32(bgRGBA.R)<<16 | uint32(bgRGBA.G)<<8 | uint32(bgRGBA.B))) {
				return color.RGBA{R: 110, G: 110, B: 110, A: 255}
			} else {
				return color.RGBA{R: 145, G: 145, B: 145, A: 255}
			}
		}
		return color.RGBA{
			R: textColor.Colour.Red(),
			G: textColor.Colour.Green(),
			B: textColor.Colour.Blue(),
			A: 180, // Make it slightly transparent
		}
	}
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

func getColorFromChroma(style *chroma.Style, c chroma.Colour) color.Color {
	if c != 0 {
		return color.RGBA{
			R: c.Red(),
			G: c.Green(),
			B: c.Blue(),
			A: 255,
		}
	}

	// If no specific color is provided, use the theme's "Other" color as the default text color
	otherColor := style.Get(chroma.Other)
	if otherColor.Colour != 0 {
		return color.RGBA{
			R: otherColor.Colour.Red(),
			G: otherColor.Colour.Green(),
			B: otherColor.Colour.Blue(),
			A: 255,
		}
	}

	// If no "Other" color is available, base the color on the background
	bg := getBackgroundColor(style)
	bgRGBA := bg.(color.RGBA)
	if isLight(chroma.Colour(uint32(bgRGBA.R)<<16 | uint32(bgRGBA.G)<<8 | uint32(bgRGBA.B))) {
		// For light backgrounds, use a darker gray
		return color.RGBA{R: 74, G: 74, B: 74, A: 255}
	} else {
		// For dark backgrounds, use a light gray
		return color.RGBA{R: 204, G: 204, B: 204, A: 255}
	}
}

// Helper functions for color manipulation
func maxu8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func minu8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

// wrapTokens splits tokens into multiple lines if they exceed maxWidth
func wrapTokens(tokens []Token, face font.Face, maxWidth, startX int) [][]Token {
	if maxWidth <= 0 {
		return [][]Token{tokens}
	}

	var result [][]Token
	var currentLine []Token
	currentWidth := startX

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tokenWidth := font.MeasureString(face, token.Text).Round()

		// Check if this token would exceed max width
		if currentWidth+tokenWidth > maxWidth {
			if len(currentLine) > 0 {
				// Before starting a new line, check if splitting the current token would
				// allow the next token to fit on the same line
				if i+1 < len(tokens) {
					nextToken := tokens[i+1]
					nextTokenWidth := font.MeasureString(face, nextToken.Text).Round()

					// If the next token starts with a space and would fit after splitting
					if strings.HasPrefix(nextToken.Text, " ") &&
						currentWidth+nextTokenWidth <= maxWidth {
						// Keep the current line and continue to next token
						currentLine = append(currentLine, token)
						currentWidth += tokenWidth
						continue
					}
				}

				// Otherwise start a new line
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
		}

		currentLine = append(currentLine, token)
		currentWidth += tokenWidth
	}

	if len(currentLine) > 0 {
		result = append(result, currentLine)
	}

	return result
}

// findMaxFittingWords finds the maximum number of words that can fit within maxWidth
func findMaxFittingWords(text string, face font.Face, startWidth, maxWidth int) (int, int) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0, 0
	}

	currentWidth := startWidth
	lastPos := 0
	for i, word := range words {
		wordWidth := font.MeasureString(face, word).Round()
		if i > 0 {
			wordWidth += font.MeasureString(face, " ").Round() // Add space before word
		}
		if currentWidth+wordWidth > maxWidth {
			return i, lastPos
		}
		currentWidth += wordWidth
		lastPos += len(word)
		if i < len(words)-1 {
			lastPos++ // Account for space
		}
	}
	return len(words), len(text)
}

// splitToken splits a single token into multiple tokens that fit within maxWidth
func splitToken(token Token, face font.Face, maxWidth int) []Token {
	var result []Token
	text := token.Text
	for len(text) > 0 {
		// Try to fit as many complete words as possible
		numWords, endPos := findMaxFittingWords(text, face, 0, maxWidth)

		if numWords > 0 {
			// We can fit at least one word
			result = append(result, Token{
				Text:   strings.TrimRight(text[:endPos], " "),
				Color:  token.Color,
				Bold:   token.Bold,
				Italic: token.Italic,
			})
			text = strings.TrimLeft(text[endPos:], " ")
			continue
		}

		// If we can't fit even one word, we need to split the first word
		firstSpace := strings.IndexAny(text, " \t\n")
		if firstSpace == -1 {
			firstSpace = len(text)
		}
		word := text[:firstSpace]

		// Binary search for the maximum characters that fit
		low, high := 1, len(word)
		for low < high {
			mid := (low + high + 1) / 2
			width := font.MeasureString(face, word[:mid]).Round()
			if width <= maxWidth {
				low = mid
			} else {
				high = mid - 1
			}
		}

		if low > 0 {
			// Only split if we can fit at least one character
			result = append(result, Token{
				Text:   word[:low],
				Color:  token.Color,
				Bold:   token.Bold,
				Italic: token.Italic,
			})
			text = word[low:] + text[firstSpace:]
		} else {
			// Emergency fallback: take at least one character
			result = append(result, Token{
				Text:   text[:1],
				Color:  token.Color,
				Bold:   token.Bold,
				Italic: token.Italic,
			})
			text = text[1:]
		}
	}

	return result
}

func validateLineRanges(lines []Line, ranges []content.LineRange) error {
	if len(ranges) > 0 {
		for i := range ranges {
			// If start is <= 1, set it to 1
			if ranges[i].Start <= 1 {
				ranges[i].Start = 1
			}

			// If end is <= 1, set it to the total number of lines
			if ranges[i].End <= 1 {
				ranges[i].End = len(lines)
			}

			// Validate bounds after adjustment
			if ranges[i].Start > len(lines) {
				return fmt.Errorf("start line number %d is out of bounds (max: %d)", ranges[i].Start, len(lines))
			}
			if ranges[i].End > len(lines) {
				return fmt.Errorf("end line number %d is out of bounds (max: %d)", ranges[i].End, len(lines))
			}
		}
	}

	return nil
}

func validateLineHighlightRanges(lines []Line, lineRanges []content.LineRange, highlightRanges []content.LineRange) error {
	if len(highlightRanges) > 0 {
		lr := lineRanges
		if len(lr) == 0 {
			lr = append(lr, content.LineRange{Start: 1, End: len(lines)})
		}

		for _, lhr := range highlightRanges {
			if lhr.Start > len(lines) {
				return fmt.Errorf("start line number %d is out of bounds (max: %d)", lhr.Start, len(lines))
			}
			if lhr.End > len(lines) {
				return fmt.Errorf("end line number %d is out of bounds (max: %d)", lhr.End, len(lines))
			}
			if lhr.Start < lr[0].Start {
				return fmt.Errorf("start line number %d is out of bounds (min: %d)", lhr.Start, lr[0].Start)
			}
			if lhr.End > lr[len(lr)-1].End {
				return fmt.Errorf("end line number %d is out of bounds (max: %d)", lhr.End, lr[len(lr)-1].End)
			}
		}
	}

	return nil
}

func expandTabs(text string, currentColumn, tabWidth int) (string, int) {
	if !strings.Contains(text, "\t") {
		return text, currentColumn + len(text)
	}

	var result strings.Builder
	col := currentColumn

	// Default to 4 spaces per tab if tabWidth is 0 or not set
	if tabWidth <= 0 {
		tabWidth = 4
	}

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
