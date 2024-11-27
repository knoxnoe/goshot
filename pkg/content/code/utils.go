package code

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"golang.org/x/image/font"
)

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

func validateLineRanges(lines []Line, ranges []LineRange) error {
	if len(ranges) > 0 {
		for _, lr := range ranges {
			if lr.Start > len(lines) {
				return fmt.Errorf("start line number %d is out of bounds (max: %d)", lr.Start, len(lines))
			}
			if lr.End > len(lines) {
				return fmt.Errorf("end line number %d is out of bounds (max: %d)", lr.End, len(lines))
			}
		}
	}

	return nil
}

func validateLineHighlightRanges(lines []Line, lineRanges []LineRange, highlightRanges []LineRange) error {
	if len(highlightRanges) > 0 {
		lr := lineRanges
		if len(lr) == 0 {
			lr = append(lr, LineRange{Start: 1, End: len(lines)})
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
