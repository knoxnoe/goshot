package render

import (
	"fmt"
	"image"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/syntax"
)

// LineRange represents a range of line numbers
type LineRange struct {
	Start int
	End   int
}

// CodeStyle represents all code styling options
type CodeStyle struct {
	Theme               string
	Language            string
	TabWidth            int
	ShowLineNumbers     bool
	LineNumberRange     LineRange
	LineHighlightRanges []LineRange // Ranges of lines to highlight
}

// Canvas represents a rendering canvas with all necessary configuration
type Canvas struct {
	chrome     chrome.Chrome
	background background.Background
	codeStyle  *CodeStyle
}

// NewCanvas creates a new Canvas instance with default options
func NewCanvas() *Canvas {
	return &Canvas{
		chrome:     chrome.NewWindows11Chrome(),
		background: nil, // No background by default
		codeStyle: &CodeStyle{
			Theme:               "dracula",
			Language:            "", // Empty means auto-detect
			TabWidth:            4,
			ShowLineNumbers:     true,
			LineNumberRange:     LineRange{},
			LineHighlightRanges: []LineRange{},
		},
	}
}

// SetChrome sets the chrome renderer
func (c *Canvas) SetChrome(chrome chrome.Chrome) *Canvas {
	c.chrome = chrome
	return c
}

// SetBackground sets the background renderer
func (c *Canvas) SetBackground(bg background.Background) *Canvas {
	c.background = bg
	return c
}

// SetCodeStyle sets the code styling options
func (c *Canvas) SetCodeStyle(style *CodeStyle) *Canvas {
	c.codeStyle = style
	return c
}

// RenderCode renders the code to an image
func (c *Canvas) RenderCode(code string) (image.Image, error) {
	// Get highlighted code
	highlightOpts := &syntax.HighlightOptions{
		Style:            c.codeStyle.Theme,
		Language:         c.codeStyle.Language,
		TabWidth:         c.codeStyle.TabWidth,
		ShowLineNums:     c.codeStyle.ShowLineNumbers,
		HighlightedLines: flattenHighlightRanges(c.codeStyle.LineHighlightRanges),
	}

	highlighted, err := syntax.Highlight(code, highlightOpts)
	if err != nil {
		return nil, fmt.Errorf("error highlighting code: %v", err)
	}

	// Create render config using highlighted code's colors
	renderConfig := syntax.DefaultConfig().
		SetShowLineNumbers(c.codeStyle.ShowLineNumbers).
		SetLineHighlightColor(highlighted.HighlightColor).
		SetLineNumberColor(highlighted.LineNumberColor).
		SetLineNumberBg(highlighted.GutterColor)

	// Set line number range if specified
	if c.codeStyle.LineNumberRange.Start > 0 || c.codeStyle.LineNumberRange.End > 0 {
		renderConfig.StartLineNumber = c.codeStyle.LineNumberRange.Start
		renderConfig.EndLineNumber = c.codeStyle.LineNumberRange.End
	}

	// Create the image
	img, err := highlighted.RenderToImage(renderConfig)
	if err != nil {
		return nil, fmt.Errorf("error rendering code: %v", err)
	}

	// Apply chrome if set
	if c.chrome != nil {
		img, err = c.chrome.Render(img)
		if err != nil {
			return nil, fmt.Errorf("error rendering chrome: %v", err)
		}
	}

	// Apply background if set
	if c.background != nil {
		img = c.background.Render(img)
	}

	return img, nil
}

// flattenHighlightRanges converts a slice of LineRanges into a slice of line numbers
func flattenHighlightRanges(ranges []LineRange) []int {
	if len(ranges) == 0 {
		return nil
	}

	// First, count how many lines we need
	count := 0
	for _, r := range ranges {
		if r.End < r.Start {
			continue // Skip invalid ranges
		}
		count += r.End - r.Start + 1
	}

	// Create and fill the slice
	lines := make([]int, 0, count)
	for _, r := range ranges {
		if r.End < r.Start {
			continue
		}
		for line := r.Start; line <= r.End; line++ {
			lines = append(lines, line)
		}
	}

	return lines
}
