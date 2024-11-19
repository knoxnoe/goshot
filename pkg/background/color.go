package background

import (
	"image"
	"image/color"
	"image/draw"
)

// ColorBackground represents a solid color background
type ColorBackground struct {
	color   color.Color
	padding Padding
}

// NewColorBackground creates a new ColorBackground with the given color
func NewColorBackground() ColorBackground {
	return ColorBackground{
		color:   LightColor,
		padding: NewPadding(20),
	}
}

// SetColor sets the background color
func (bg ColorBackground) SetColor(c color.Color) ColorBackground {
	bg.color = c
	return bg
}

// SetPadding sets the padding for the background
func (bg ColorBackground) SetPadding(p Padding) ColorBackground {
	bg.padding = p
	return bg
}

// SetPaddingValue sets equal padding for all sides
func (bg ColorBackground) SetPaddingValue(value int) ColorBackground {
	bg.padding = NewPadding(value)
	return bg
}

// SetPaddingHV sets equal horizontal and vertical padding
func (bg ColorBackground) SetPaddingHV(horizontal, vertical int) ColorBackground {
	bg.padding = NewPaddingHV(horizontal, vertical)
	return bg
}

// Render applies the background to the given content image
// It returns a new image with the background applied and the content centered
func (bg ColorBackground) Render(content image.Image) image.Image {
	bounds := content.Bounds()

	// Calculate new dimensions including padding
	width := bounds.Dx() + bg.padding.Left + bg.padding.Right
	height := bounds.Dy() + bg.padding.Top + bg.padding.Bottom

	// Create new image with background color
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with background color
	draw.Draw(img, img.Bounds(), &image.Uniform{bg.color}, image.Point{}, draw.Src)

	// Draw the content in the center (accounting for padding)
	contentRect := image.Rect(
		bg.padding.Left,
		bg.padding.Top,
		width-bg.padding.Right,
		height-bg.padding.Bottom,
	)
	draw.Draw(img, contentRect, content, bounds.Min, draw.Over)

	return img
}
