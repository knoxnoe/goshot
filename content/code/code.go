package code

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/watzon/goshot/content"
	"github.com/watzon/goshot/fonts"
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
	RedactionConfig     *RedactionConfig    // Redaction configuration
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
		LineHeight:        1.0,
		PaddingLeft:       10,
		PaddingRight:      10,
		PaddingTop:        10,
		PaddingBottom:     10,
		LineNumberPadding: 10,
		TabWidth:          4,
		MinWidth:          300,
		MaxWidth:          900,
		ShowLineNumbers:   true,
		RedactionConfig:   NewRedactionConfig(),
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

func (r *CodeRenderer) WithRedactionEnabled(enabled bool) *CodeRenderer {
	if r.Style.RedactionConfig == nil {
		r.Style.RedactionConfig = NewRedactionConfig()
	}
	r.Style.RedactionConfig.Enabled = enabled
	return r
}

func (r *CodeRenderer) WithRedactionBlurRadius(radius float64) *CodeRenderer {
	if r.Style.RedactionConfig == nil {
		r.Style.RedactionConfig = NewRedactionConfig()
	}
	r.Style.RedactionConfig.BlurRadius = radius
	return r
}

func (r *CodeRenderer) WithRedactionPattern(pattern string, name string) *CodeRenderer {
	if r.Style.RedactionConfig == nil {
		r.Style.RedactionConfig = NewRedactionConfig()
	}
	compiled, err := regexp.Compile(pattern)
	if err == nil {
		r.Style.RedactionConfig.Patterns = append(r.Style.RedactionConfig.Patterns, RedactionPattern{
			Pattern: compiled,
			Name:    name,
		})
	} else {
		log.Printf("Failed to compile redaction pattern %q: %v", pattern, err)
	}
	return r
}

func (r *CodeRenderer) WithManualRedaction(x, y, width, height int) *CodeRenderer {
	if r.Style.RedactionConfig == nil {
		r.Style.RedactionConfig = NewRedactionConfig()
	}
	r.Style.RedactionConfig.AddManualRedaction(x, y, width, height)
	return r
}

func (r *CodeRenderer) WithRedactionStyle(style RedactionStyle) *CodeRenderer {
	if r.Style.RedactionConfig == nil {
		r.Style.RedactionConfig = NewRedactionConfig()
	}
	r.Style.RedactionConfig.Style = style
	return r
}

// getLineText concatenates all tokens in a line into a single string
func getLineText(line Line) string {
	var text strings.Builder
	for _, token := range line.Tokens {
		text.WriteString(token.Text)
	}
	return text.String()
}

func drawText(img *image.RGBA, face font.Face, text string, x, y int, col color.Color, token Token) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)

	// Draw underline if needed
	if token.Underline {
		metrics := face.Metrics()
		underlineY := y + metrics.Descent.Round()/2
		width := font.MeasureString(face, text).Round()

		// Draw a line 1px thick
		for dy := 0; dy < 1; dy++ {
			for dx := 0; dx < width; dx++ {
				img.Set(x+dx, underlineY+dy, col)
			}
		}
	}
}

func (r *CodeRenderer) Render() (image.Image, error) {
	config := r.Style
	h, err := Highlight(r.Code, r.Style)
	if err != nil {
		return nil, err
	}

	// Get the font face for each style combination we need
	regularFace, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightRegular,
		Stretch: fonts.StretchNormal,
	})
	if err != nil {
		return nil, err
	}
	defer regularFace.Close()

	boldFace, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightBold,
		Stretch: fonts.StretchNormal,
	})
	if err != nil {
		return nil, err
	}
	defer boldFace.Close()

	italicFace, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightRegular,
		Stretch: fonts.StretchNormal,
		Italic:  true,
	})
	if err != nil {
		return nil, err
	}
	defer italicFace.Close()

	boldItalicFace, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightBold,
		Stretch: fonts.StretchNormal,
		Italic:  true,
	})
	if err != nil {
		return nil, err
	}
	defer boldItalicFace.Close()

	// Function to get the appropriate face based on token style
	getFaceForToken := func(token Token) font.Face {
		if token.Bold && token.Italic && !token.NoItalic {
			return boldItalicFace.Face
		} else if token.Bold {
			return boldFace.Face
		} else if token.Italic && !token.NoItalic {
			return italicFace.Face
		}
		return regularFace.Face
	}

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

	// Create ellipsis token with comment color
	ellipsisToken := Token{
		Text:   "...",
		Color:  h.CommentColor,
		Italic: true, // Comments are typically italic
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
			// For the first ellipsis, show the line number that comes before the first range
			lineNumberMap = append(lineNumberMap, config.LineRanges[0].Start-1)
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
				// For ellipsis between ranges, show the line number that comes after the previous range
				lineNumberMap = append(lineNumberMap, lr.End+1)
			}
		}

		// Add ellipsis at end if last range doesn't end at the last line
		if config.LineRanges[len(config.LineRanges)-1].End < len(lines) {
			filteredLines = append(filteredLines, Line{
				Tokens: []Token{ellipsisToken},
			})
			// For the last ellipsis, show the line number that comes after the last range
			lineNumberMap = append(lineNumberMap, config.LineRanges[len(config.LineRanges)-1].End+1)
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
	metrics := regularFace.Face.Metrics()
	lineHeight := int(float64(metrics.Height.Round()) * config.LineHeight)
	maxLineWidth := 0

	// First measure ellipsis width if we have ranges
	ellipsisWidth := 0
	if len(config.LineRanges) > 0 {
		ellipsisWidth = font.MeasureString(regularFace.Face, "...").Round()
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
			wrapped = wrapTokens(line.Tokens, regularFace.Face, maxTextWidth, 0)
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
					lineWidth += font.MeasureString(regularFace.Face, expandedText).Round()
					currentColumn = newColumn
				} else {
					lineWidth += font.MeasureString(regularFace.Face, token.Text).Round()
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

	// Find redaction ranges if redaction is enabled
	var lineRedactionRanges map[int][]RedactionRange
	if r.Style.RedactionConfig != nil && r.Style.RedactionConfig.Enabled {
		lineRedactionRanges = make(map[int][]RedactionRange)

		// First find redaction ranges in the entire text
		var fullText strings.Builder
		lineStarts := make([]int, len(lines))
		currentPos := 0

		for i, line := range lines {
			lineStarts[i] = currentPos
			text := getLineText(line)
			fullText.WriteString(text)
			fullText.WriteString("\n")
			currentPos = fullText.Len()
		}

		// Find ranges in the full text
		ranges := FindRedactionRanges(r.Style.RedactionConfig, fullText.String())

		// Map the ranges back to individual lines
		for _, r := range ranges {
			// Find which line(s) this range belongs to
			for i := 0; i < len(lines); i++ {
				lineStart := lineStarts[i]
				lineEnd := lineStarts[i] + len(getLineText(lines[i]))

				// Check if this range overlaps with the current line
				if r.StartIndex <= lineEnd && r.EndIndex > lineStart {
					// Calculate the portion of the range that falls within this line
					startInLine := max(0, r.StartIndex-lineStart)
					endInLine := min(lineEnd-lineStart, r.EndIndex-lineStart)

					lineRange := RedactionRange{
						StartIndex: startInLine,
						EndIndex:   endInLine,
						Pattern:    r.Pattern,
					}
					lineRedactionRanges[i] = append(lineRedactionRanges[i], lineRange)
				}
			}
		}
	}

	// Create a separate image for blur-style redactions
	var blurImg *image.RGBA
	if r.Style.RedactionConfig != nil && r.Style.RedactionConfig.Style == RedactionStyleBlur {
		blurImg = image.NewRGBA(img.Bounds())
		draw.Draw(blurImg, blurImg.Bounds(), img, image.Point{}, draw.Src)
	}

	// Track areas to blur
	type blurArea struct {
		startX, startY int
		width          int
	}
	var currentBlurArea *blurArea
	var blurAreas []blurArea

	// Track character offsets for wrapped lines
	type wrappedLineInfo struct {
		originalLineIdx int
		startOffset     int // Character offset where this wrapped line starts in the original line
	}
	wrappedLineOffsets := make([]wrappedLineInfo, len(wrappedLines))
	currentOffset := 0
	for i, tokens := range wrappedLines {
		originalLineIdx := lineToWrappedMap[i]

		// If this is the first wrapped line for this original line, reset the offset
		if i == 0 || lineToWrappedMap[i-1] != originalLineIdx {
			currentOffset = 0
		}

		wrappedLineOffsets[i] = wrappedLineInfo{
			originalLineIdx: originalLineIdx,
			startOffset:     currentOffset,
		}

		// Calculate the length of this wrapped line for the next offset
		lineLength := 0
		for _, token := range tokens {
			if strings.Contains(token.Text, "\t") {
				expandedText, _ := expandTabs(token.Text, lineLength, config.TabWidth)
				lineLength += len(expandedText)
			} else {
				lineLength += len(token.Text)
			}
		}
		currentOffset += lineLength
	}

	// Draw line numbers and text
	currentY = config.PaddingTop

	for i, tokens := range wrappedLines {
		// Draw line numbers if enabled
		if config.ShowLineNumbers {
			originalLineIdx := lineToWrappedMap[i]
			lineNumber := lineNumberMap[originalLineIdx]
			lineNumberStr := strconv.Itoa(lineNumber)
			lineNumberWidth := font.MeasureString(regularFace.Face, lineNumberStr)

			// Get the font face for line numbers
			face, err := config.Font.GetFace(config.FontSize, &fonts.FontStyle{
				Weight:  fonts.WeightRegular,
				Stretch: fonts.StretchNormal,
			})
			if err != nil {
				return nil, err
			}
			defer face.Close()

			// Draw the line number
			drawText(img, regularFace.Face, lineNumberStr, config.PaddingLeft+lineNumberOffset-lineNumberWidth.Round()-config.LineNumberPadding, currentY+metrics.Ascent.Round(), h.LineNumberColor, Token{Text: lineNumberStr})
		}

		// Draw tokens
		x := config.PaddingLeft + lineNumberOffset
		currentColumn := wrappedLineOffsets[i].startOffset
		originalLineIdx := wrappedLineOffsets[i].originalLineIdx
		redactionRanges := lineRedactionRanges[originalLineIdx]

		for _, token := range tokens {
			// Handle tab expansion for drawing
			if strings.Contains(token.Text, "\t") {
				expandedText, newColumn := expandTabs(token.Text, currentColumn, config.TabWidth)
				// Draw expanded text character by character
				charX := x
				for j, ch := range expandedText {
					shouldRedact := false
					if len(redactionRanges) > 0 {
						shouldRedact = ShouldRedact(currentColumn+j, redactionRanges)
					}

					if shouldRedact {
						if r.Style.RedactionConfig.Style == RedactionStyleBlock {
							// Draw a block character
							drawText(img, getFaceForToken(token), "█", charX, currentY+metrics.Ascent.Round(), token.Color, token)
						} else {
							// For blur style, track the area to blur
							if currentBlurArea == nil {
								currentBlurArea = &blurArea{
									startX: charX,
									startY: currentY,
									width:  0,
								}
							}
							// Still draw the actual character but we'll blur it later
							drawText(blurImg, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						}
					} else {
						// If we were tracking a blur area, finish it
						if currentBlurArea != nil {
							blurAreas = append(blurAreas, *currentBlurArea)
							currentBlurArea = nil
						}
						// Draw the character normally
						if r.Style.RedactionConfig.Style == RedactionStyleBlur {
							drawText(blurImg, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						} else {
							drawText(img, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						}
					}

					charWidth := font.MeasureString(getFaceForToken(token), string(ch)).Round()
					if currentBlurArea != nil {
						currentBlurArea.width += charWidth
					}
					charX += charWidth
				}
				x = charX
				currentColumn = newColumn
			} else {
				// Draw regular text character by character
				charX := x
				for j, ch := range token.Text {
					shouldRedact := false
					if len(redactionRanges) > 0 {
						shouldRedact = ShouldRedact(currentColumn+j, redactionRanges)
					}

					if shouldRedact {
						if r.Style.RedactionConfig.Style == RedactionStyleBlock {
							// Draw a block character
							drawText(img, getFaceForToken(token), "█", charX, currentY+metrics.Ascent.Round(), token.Color, token)
						} else {
							// For blur style, track the area to blur
							if currentBlurArea == nil {
								currentBlurArea = &blurArea{
									startX: charX,
									startY: currentY,
									width:  0,
								}
							}
							// Still draw the actual character but we'll blur it later
							drawText(blurImg, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						}
					} else {
						// If we were tracking a blur area, finish it
						if currentBlurArea != nil {
							blurAreas = append(blurAreas, *currentBlurArea)
							currentBlurArea = nil
						}
						// Draw the character normally
						if r.Style.RedactionConfig.Style == RedactionStyleBlur {
							drawText(blurImg, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						} else {
							drawText(img, getFaceForToken(token), string(ch), charX, currentY+metrics.Ascent.Round(), token.Color, token)
						}
					}

					charWidth := font.MeasureString(getFaceForToken(token), string(ch)).Round()
					if currentBlurArea != nil {
						currentBlurArea.width += charWidth
					}
					charX += charWidth
				}
				x = charX
				currentColumn += len(token.Text)
			}
		}

		// If we have an unfinished blur area at the end of the line, add it
		if currentBlurArea != nil {
			blurAreas = append(blurAreas, *currentBlurArea)
			currentBlurArea = nil
		}

		currentY += lineHeight
	}

	// Apply blur effect to collected areas if using blur style
	if r.Style.RedactionConfig != nil && r.Style.RedactionConfig.Style == RedactionStyleBlur {
		metrics := regularFace.Face.Metrics()
		lineHeight := metrics.Height.Round()

		for _, area := range blurAreas {
			redactArea(blurImg, area.startX, area.startY, area.width, lineHeight, r.Style.RedactionConfig.BlurRadius)
		}

		// Apply any manual redactions
		for _, area := range r.Style.RedactionConfig.ManualRedactions {
			redactArea(blurImg, area.X, area.Y, area.Width, area.Height, r.Style.RedactionConfig.BlurRadius)
		}

		img = blurImg
	} else if r.Style.RedactionConfig != nil {
		// Apply manual redactions with block style
		for _, area := range r.Style.RedactionConfig.ManualRedactions {
			// Fill the area with block characters
			blockChar := "█"
			blockWidth := font.MeasureString(regularFace.Face, blockChar).Round()
			numBlocks := area.Width / blockWidth

			for y := area.Y; y < area.Y+area.Height; y += lineHeight {
				for i := 0; i < numBlocks; i++ {
					drawText(img, regularFace.Face, blockChar, area.X+(i*blockWidth), y+metrics.Ascent.Round(), color.Black, Token{Text: blockChar})
				}
			}
		}
	}

	return img, nil
}
