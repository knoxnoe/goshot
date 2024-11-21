package render

import (
	"fmt"
	"image"

	"github.com/golang/freetype/truetype"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/syntax"
	"golang.org/x/image/font"
)

// LineRange represents a range of line numbers
type LineRange struct {
	Start int
	End   int
}

// CodeStyle represents all code styling options
type CodeStyle struct {
	// Syntax highlighting options
	Theme               string
	Language            string
	TabWidth            int
	ShowLineNumbers     bool
	LineNumberRange     LineRange
	LineHighlightRanges []LineRange

	// Rendering options
	FontSize          float64
	FontFamily        *fonts.Font
	LineHeight        float64
	PaddingLeft       int
	PaddingRight      int
	PaddingTop        int
	PaddingBottom     int
	MinWidth          int
	MaxWidth          int
	LineNumberPadding int
}

// NewCodeStyle creates a new CodeStyle with default values
func NewCodeStyle() *CodeStyle {
	return &CodeStyle{
		FontSize:      14,
		LineHeight:    1.5,
		PaddingLeft:   10,
		PaddingRight:  10,
		PaddingTop:    10,
		PaddingBottom: 10,
	}
}

// Canvas represents a rendering canvas with all necessary configuration
type Canvas struct {
	chrome     chrome.Chrome
	background background.Background
	codeStyle  *CodeStyle
}

// NewCanvas creates a new Canvas instance with default options
func NewCanvas() *Canvas {
	// Get default monospace font
	defaultFont, err := fonts.GetFallback(fonts.FallbackMono)
	if err != nil {
		// Fallback will be handled by the syntax package
		defaultFont = nil
	}

	return &Canvas{
		chrome:     chrome.NewBlankChrome(),
		background: nil, // No background by default
		codeStyle: &CodeStyle{
			Theme:               "dracula",
			Language:            "", // Empty means auto-detect
			TabWidth:            4,
			ShowLineNumbers:     true,
			LineNumberRange:     LineRange{},
			LineHighlightRanges: []LineRange{},
			FontSize:            14,
			FontFamily:          defaultFont,
			LineHeight:          1.5,
			PaddingLeft:         16,
			PaddingRight:        16,
			PaddingTop:          16,
			PaddingBottom:       16,
			MinWidth:            0,
			MaxWidth:            0,
			LineNumberPadding:   16,
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

// SetFont sets the font family to use for rendering
func (c *Canvas) SetFont(fontName string) error {
	font, err := fonts.GetFont(fontName, nil)
	if err != nil {
		return fmt.Errorf("error setting font: %v", err)
	}
	c.codeStyle.FontFamily = font
	return nil
}

// SetFontWithStyle sets the font family with specific style options
func (c *Canvas) SetFontWithStyle(fontName string, style *fonts.FontStyle) error {
	font, err := fonts.GetFont(fontName, style)
	if err != nil {
		return fmt.Errorf("error setting font: %v", err)
	}
	c.codeStyle.FontFamily = font
	return nil
}

// SetFontSize sets the font size in points
func (c *Canvas) SetFontSize(size float64) *Canvas {
	c.codeStyle.FontSize = size
	return c
}

// SetLineHeight sets the line height as a multiplier of font size
func (c *Canvas) SetLineHeight(height float64) *Canvas {
	c.codeStyle.LineHeight = height
	return c
}

// RenderToImage renders the code to an image
func (c *Canvas) RenderToImage(code string) (image.Image, error) {
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

	// Apply font settings
	if c.codeStyle.FontFamily == nil {
		fallback, err := fonts.GetFallback(fonts.FallbackMono)
		if err != nil {
			return nil, fmt.Errorf("error getting fallback font: %v", err)
		}
		c.codeStyle.FontFamily = fallback
	}

	ttf, err := c.codeStyle.FontFamily.ToTrueType()
	if err != nil {
		return nil, fmt.Errorf("failed to convert font: %v", err)
	}

	face := truetype.NewFace(ttf, &truetype.Options{
		Size:    c.codeStyle.FontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	renderConfig.SetFontFace(face, ttf, c.codeStyle.FontSize)

	if c.codeStyle.LineHeight > 0 {
		renderConfig.SetLineHeight(c.codeStyle.LineHeight)
	}
	if c.codeStyle.PaddingLeft > 0 {
		renderConfig.SetPaddingLeft(c.codeStyle.PaddingLeft)
	}
	if c.codeStyle.PaddingRight > 0 {
		renderConfig.SetPaddingRight(c.codeStyle.PaddingRight)
	}
	if c.codeStyle.PaddingTop > 0 {
		renderConfig.SetPaddingTop(c.codeStyle.PaddingTop)
	}
	if c.codeStyle.PaddingBottom > 0 {
		renderConfig.SetPaddingBottom(c.codeStyle.PaddingBottom)
	}
	if c.codeStyle.MinWidth > 0 {
		renderConfig.SetMinWidth(c.codeStyle.MinWidth)
	}
	if c.codeStyle.MaxWidth > 0 {
		renderConfig.SetMaxWidth(c.codeStyle.MaxWidth)
	}
	if c.codeStyle.LineNumberPadding > 0 {
		renderConfig.SetLineNumberPadding(c.codeStyle.LineNumberPadding)
	}
	if c.codeStyle.TabWidth > 0 {
		renderConfig.SetTabWidth(c.codeStyle.TabWidth)
	}

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
