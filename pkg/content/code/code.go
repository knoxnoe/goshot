package code

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"

	"github.com/watzon/goshot/pkg/content"
	"github.com/watzon/goshot/pkg/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type CodeStyle struct {
	Theme               string              // The chroma syntax theme to use
	Language            string              // The language to highlight
	Font                *fonts.Font         // The font to use
	FontSize            float64             // The font size in points
	LineHeight          float64             // The line height multiplier
	PaddingLeft         int                 // Padding between the code and the left edge
	PaddingRight        int                 // Padding between the code and the right edge
	PaddingTop          int                 // Padding between the code and the top edge
	PaddingBottom       int                 // Padding between the code and the bottom edge
	LineNumberPadding   int                 // Padding between line numbers and code
	TabWidth            int                 // Width of tab characters in spaces
	MinWidth            int                 // Minimum width in pixels (0 means no minimum)
	MaxWidth            int                 // Maximum width in pixels (0 means no limit)
	ShowLineNumbers     bool                // Whether to show line numbers
	LineRanges          []content.LineRange // Ranges of lines to render
	LineHighlightRanges []content.LineRange // Ranges of lines to highlight
}

type CodeRenderer struct {
	Code  string
	Style *CodeStyle
}

func NewRenderer(input string, style *CodeStyle) *CodeRenderer {
	return &CodeRenderer{
		Code:  input,
		Style: style,
	}
}

func DefaultRenderer(input string) *CodeRenderer {
	font, err := fonts.GetFallback(fonts.FallbackMono)
	if err != nil {
		panic(err)
	}

	return NewRenderer(input, &CodeStyle{
		Theme:             "monokai",
		Language:          "go",
		Font:              font,
		FontSize:          12,
		LineHeight:        1.2,
		PaddingLeft:       10,
		PaddingRight:      10,
		PaddingTop:        10,
		PaddingBottom:     10,
		LineNumberPadding: 10,
		TabWidth:          4,
		MinWidth:          300,
		MaxWidth:          900,
		ShowLineNumbers:   true,
	})
}

func (r *CodeRenderer) WithTheme(theme string) *CodeRenderer {
	r.Style.Theme = theme
	return r
}

func (r *CodeRenderer) WithLanguage(language string) *CodeRenderer {
	r.Style.Language = language
	return r
}

func (r *CodeRenderer) WithFontSize(size float64) *CodeRenderer {
	r.Style.FontSize = size
	return r
}

func (r *CodeRenderer) WithLineHeight(height float64) *CodeRenderer {
	r.Style.LineHeight = height
	return r
}

func (r *CodeRenderer) WithPadding(left, right, top, bottom int) *CodeRenderer {
	r.Style.PaddingLeft = left
	r.Style.PaddingRight = right
	r.Style.PaddingTop = top
	r.Style.PaddingBottom = bottom
	return r
}

func (r *CodeRenderer) WithLineNumberPadding(padding int) *CodeRenderer {
	r.Style.LineNumberPadding = padding
	return r
}

func (r *CodeRenderer) WithTabWidth(width int) *CodeRenderer {
	r.Style.TabWidth = width
	return r
}

func (r *CodeRenderer) WithMinWidth(width int) *CodeRenderer {
	r.Style.MinWidth = width
	return r
}

func (r *CodeRenderer) WithMaxWidth(width int) *CodeRenderer {
	r.Style.MaxWidth = width
	return r
}

func (r *CodeRenderer) WithLineNumbers(show bool) *CodeRenderer {
	r.Style.ShowLineNumbers = show
	return r
}

func (r *CodeRenderer) WithFont(font *fonts.Font) *CodeRenderer {
	r.Style.Font = font
	return r
}

func (r *CodeRenderer) WithFontName(name string, style *fonts.FontStyle) *CodeRenderer {
	font, err := fonts.GetFont(name, style)
	if err != nil {
		panic(err)
	}
	return r.WithFont(font)
}

func (r *CodeRenderer) WithStyle(style *CodeStyle) *CodeRenderer {
	r.Style = style
	return r
}

func (r *CodeRenderer) WithLineRange(start, end int) *CodeRenderer {
	r.Style.LineRanges = append(r.Style.LineRanges, content.LineRange{Start: start, End: end})
	return r
}

func (r *CodeRenderer) WithLineHighlightRange(start, end int) *CodeRenderer {
	r.Style.LineHighlightRanges = append(r.Style.LineHighlightRanges, content.LineRange{Start: start, End: end})
	return r
}

func drawText(img *image.RGBA, face font.Face, text string, x, y int, col color.Color) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)
}

func (r *CodeRenderer) Render() (image.Image, error) {
	config := r.Style
	h, err := Highlight(r.Code, r.Style)
	if err != nil {
		return nil, err
	}

	// Get the font face
	face, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightRegular,
		Stretch: fonts.StretchNormal,
	})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	// Get lines
	lines := h.Lines

	// Validate that the requested line ranges are within bounds
	err = validateLineRanges(lines, config.LineRanges)
	if err != nil {
		return nil, err
	}

	// Validate that the requested line highlight ranges are within the bounds of the set line ranges
	err = validateLineHighlightRanges(lines, config.LineRanges, config.LineHighlightRanges)
	if err != nil {
		return nil, err
	}

	// Create ellipsis token with theme color
	ellipsisToken := Token{
		Text:  "...",
		Color: h.LineNumberColor,
	}

	// Filter lines based on ranges and add ellipses
	var filteredLines []Line
	var lineNumberMap []int // Map filtered line indices to original line numbers
	if len(config.LineRanges) > 0 {
		// Add ellipsis at start if first range doesn't start at 1
		if config.LineRanges[0].Start > 1 {
			filteredLines = append(filteredLines, Line{
				Tokens: []Token{ellipsisToken},
			})
			lineNumberMap = append(lineNumberMap, 0) // 0 indicates ellipsis line
		}

		// Process each range
		for i, lr := range config.LineRanges {
			// Convert to 0-based indices
			start := lr.Start - 1
			end := lr.End - 1

			// Add lines in this range
			for j := start; j <= end; j++ {
				filteredLines = append(filteredLines, lines[j])
				lineNumberMap = append(lineNumberMap, j+1) // Store 1-based line number
			}

			// Add ellipsis between ranges
			if i < len(config.LineRanges)-1 && lr.End+1 < config.LineRanges[i+1].Start {
				filteredLines = append(filteredLines, Line{
					Tokens: []Token{ellipsisToken},
				})
				lineNumberMap = append(lineNumberMap, 0) // 0 indicates ellipsis line
			}
		}

		// Add ellipsis at end if last range doesn't end at the last line
		if config.LineRanges[len(config.LineRanges)-1].End < len(lines) {
			filteredLines = append(filteredLines, Line{
				Tokens: []Token{ellipsisToken},
			})
			lineNumberMap = append(lineNumberMap, 0) // 0 indicates ellipsis line
		}

		lines = filteredLines
	} else {
		// If no ranges specified, create 1:1 mapping
		lineNumberMap = make([]int, len(lines))
		for i := range lines {
			lineNumberMap[i] = i + 1
		}
	}

	// Calculate line number width if needed
	lineNumberOffset := 0
	if config.ShowLineNumbers {
		// Calculate width needed for the largest line number
		var maxLineNumber int
		if len(config.LineRanges) > 0 {
			// Use the last line number from the last range
			maxLineNumber = config.LineRanges[len(config.LineRanges)-1].End
		} else {
			maxLineNumber = len(lines)
		}
		lineCountStr := strconv.Itoa(maxLineNumber)
		maxDigits := len(lineCountStr)

		// Calculate base width for digits
		lnw, err := config.Font.MeasureString(strings.Repeat("9", maxDigits), config.FontSize, &fonts.FontStyle{
			Weight:  fonts.WeightRegular,
			Stretch: fonts.StretchNormal,
		})
		if err != nil {
			return nil, err
		}

		// Round to nearest pixel
		lineNumberWidth := lnw.Round()

		// Only add padding to the right side of line numbers
		lineNumberOffset = lineNumberWidth + config.LineNumberPadding
	}

	// Calculate max text width (total width minus padding and line numbers)
	maxTextWidth := config.MaxWidth - config.PaddingLeft - config.PaddingRight - lineNumberOffset

	// Calculate initial dimensions
	metrics := face.Face.Metrics()
	lineHeight := int(float64(metrics.Height.Round()) * config.LineHeight)
	maxLineWidth := 0

	// First measure ellipsis width if we have ranges
	ellipsisWidth := 0
	if len(config.LineRanges) > 0 {
		ellipsisWidth = font.MeasureString(face.Face, "...").Round()
		if ellipsisWidth > maxLineWidth {
			maxLineWidth = ellipsisWidth
		}
	}

	// Wrap lines and calculate max width
	wrappedLines := [][]Token{}
	lineToWrappedMap := make([]int, 0) // Maps wrapped line index to original line index
	for i, line := range lines {
		var wrapped [][]Token
		if len(line.Tokens) > 0 {
			wrapped = wrapTokens(line.Tokens, face.Face, maxTextWidth, 0)
		} else {
			// For empty lines, add an empty token list
			wrapped = [][]Token{{}}
		}
		// Add mapping for each wrapped line back to original line
		for range wrapped {
			lineToWrappedMap = append(lineToWrappedMap, i)
		}
		wrappedLines = append(wrappedLines, wrapped...)

		// Calculate max line width
		for _, wline := range wrapped {
			lineWidth := 0
			currentColumn := 0
			for _, token := range wline {
				// Handle tab expansion for width calculation
				if strings.Contains(token.Text, "\t") {
					expandedText, newColumn := expandTabs(token.Text, currentColumn, config.TabWidth)
					lineWidth += font.MeasureString(face.Face, expandedText).Round()
					currentColumn = newColumn
				} else {
					lineWidth += font.MeasureString(face.Face, token.Text).Round()
					currentColumn += len(token.Text)
				}
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
	totalHeight := (lineHeight * len(wrappedLines)) + (config.PaddingTop + config.PaddingBottom)
	if len(config.LineRanges) > 0 {
		// compensate for the ellipsis lines added between ranges and at the beginning and end
		// if the given line ranges don't cover the entire code
		totalHeight += (len(config.LineRanges) - 1)
	}

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
	currentY := config.PaddingTop
	for i := range wrappedLines {
		originalLineIdx := lineToWrappedMap[i]
		if lines[originalLineIdx].Highlight {

			// Calculate highlight rectangle based on whether line numbers are shown
			var highlightRect image.Rectangle
			if config.ShowLineNumbers {
				// With line numbers, start after the line number area
				highlightRect = image.Rect(
					config.PaddingLeft+lineNumberOffset,
					currentY,
					codeWidth+config.PaddingLeft+lineNumberOffset,
					currentY+lineHeight,
				)
			} else {
				// Without line numbers, extend to both edges
				highlightRect = image.Rect(
					0,
					currentY,
					totalWidth,
					currentY+lineHeight,
				)
			}

			uniform := image.NewUniform(h.HighlightColor)
			draw.Draw(img, highlightRect, uniform, image.Point{}, draw.Over)
		}
		currentY += lineHeight
	}

	// Draw line numbers and text
	currentY = config.PaddingTop
	for i, tokens := range wrappedLines {
		originalLineIdx := lineToWrappedMap[i]
		isFirstWrappedLine := i == 0 || lineToWrappedMap[i-1] != originalLineIdx

		// For ellipsis lines, we don't increment the line number
		if len(tokens) == 1 && tokens[0].Text == "..." {
			// Don't show anything in the line number area for ellipsis lines
			if config.ShowLineNumbers {
				// Leave gutter empty for ellipsis lines
			}
		} else {
			if config.ShowLineNumbers && isFirstWrappedLine {
				// For regular lines, show the actual line number from the map
				lineNumberStr := fmt.Sprintf("%d", lineNumberMap[originalLineIdx])
				lineNumberWidth, _ := config.Font.MeasureString(lineNumberStr, config.FontSize, &fonts.FontStyle{
					Weight:  fonts.WeightRegular,
					Stretch: fonts.StretchNormal,
				})
				drawText(img, face.Face, lineNumberStr, config.PaddingLeft+lineNumberOffset-lineNumberWidth.Round()-config.LineNumberPadding, currentY+metrics.Ascent.Round(), h.LineNumberColor)
			}
		}

		// Draw tokens
		x := config.PaddingLeft + lineNumberOffset
		currentColumn := 0
		for _, token := range tokens {
			// Handle tab expansion for drawing
			if strings.Contains(token.Text, "\t") {
				expandedText, newColumn := expandTabs(token.Text, currentColumn, config.TabWidth)
				drawText(img, face.Face, expandedText, x, currentY+metrics.Ascent.Round(), token.Color)
				x += font.MeasureString(face.Face, expandedText).Round()
				currentColumn = newColumn
			} else {
				drawText(img, face.Face, token.Text, x, currentY+metrics.Ascent.Round(), token.Color)
				x += font.MeasureString(face.Face, token.Text).Round()
				currentColumn += len(token.Text)
			}
		}
		currentY += lineHeight
	}

	return img, nil
}
