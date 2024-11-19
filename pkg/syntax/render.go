package syntax

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

// RenderConfig holds configuration for rendering highlighted code to an image
type RenderConfig struct {
	FontSize   float64
	LineHeight float64
	PaddingX   int
	PaddingY   int
	FontFamily *truetype.Font
	Background image.Image

	// Line number settings
	ShowLineNumbers bool
	LineNumberColor color.Color
	LineNumberWidth int         // Width of the line number area in pixels
	LineNumberBg    color.Color // Background color for line numbers
}

// DefaultConfig returns a default rendering configuration
func DefaultConfig() *RenderConfig {
	f, _ := truetype.Parse(goregular.TTF)
	return &RenderConfig{
		FontSize:   14,
		LineHeight: 1.5,
		PaddingX:   20,
		PaddingY:   20,
		FontFamily: f,

		// Line number defaults
		ShowLineNumbers: true,
		LineNumberColor: color.RGBA{R: 128, G: 128, B: 128, A: 255}, // Gray color
		LineNumberWidth: 50,
		LineNumberBg:    color.RGBA{R: 245, G: 245, B: 245, A: 255}, // Light gray background
	}
}

// RenderToImage converts highlighted code to an image
func (h *HighlightedCode) RenderToImage(config *RenderConfig) (image.Image, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Calculate image dimensions
	maxLineWidth := 0
	for _, line := range h.Lines {
		lineWidth := 0
		for _, token := range line.Tokens {
			lineWidth += len(token.Text)
		}
		if lineWidth > maxLineWidth {
			maxLineWidth = lineWidth
		}
	}

	// Estimate image dimensions based on font metrics
	face := truetype.NewFace(config.FontFamily, &truetype.Options{
		Size: config.FontSize,
		DPI:  72,
	})
	defer face.Close()

	// Calculate dimensions
	charWidth := font.MeasureString(face, "M").Round()
	lineHeight := int(config.FontSize * config.LineHeight)
	codeWidth := (maxLineWidth * charWidth) + (config.PaddingX * 2)

	// Add line number width if enabled
	totalWidth := codeWidth
	lineNumberOffset := 0
	if config.ShowLineNumbers {
		totalWidth += config.LineNumberWidth
		lineNumberOffset = config.LineNumberWidth
	}

	imgHeight := (len(h.Lines) * lineHeight) + (config.PaddingY * 2)

	// Create the image
	img := image.NewRGBA(image.Rect(0, 0, totalWidth, imgHeight))

	// Draw line number background if enabled
	if config.ShowLineNumbers {
		lineNumBg := image.NewUniform(config.LineNumberBg)
		draw.Draw(img, image.Rect(0, 0, config.LineNumberWidth, imgHeight), lineNumBg, image.Point{}, draw.Src)
	}

	// Create context for drawing text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(config.FontFamily)
	c.SetFontSize(config.FontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)

	// Draw each line
	for i, line := range h.Lines {
		y := config.PaddingY + ((i + 1) * lineHeight)

		// Draw line number if enabled
		if config.ShowLineNumbers {
			c.SetSrc(image.NewUniform(config.LineNumberColor))
			lineNum := fmt.Sprintf("%3d │", i+1)
			pt := freetype.Pt(5, y)
			c.DrawString(lineNum, pt)
		}

		// Draw code
		x := lineNumberOffset + config.PaddingX
		for _, token := range line.Tokens {
			c.SetSrc(image.NewUniform(token.Color))
			pt := freetype.Pt(x, y)
			c.DrawString(token.Text, pt)
			x += font.MeasureString(face, token.Text).Round()
		}
	}

	return img, nil
}
