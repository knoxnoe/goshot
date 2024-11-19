package background

import (
	"image"
	"image/color"
	"image/draw"
)

var (
	// DarkColor is the default dark mode background color
	DarkColor = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	// LightColor is the default light mode background color
	LightColor = color.RGBA{R: 240, G: 240, B: 240, A: 255}
)

// ColorBackground represents a solid color background
type ColorBackground struct {
	color   color.Color
	padding image.Point
}

// NewColorBackground creates a new ColorBackground with the given color
func NewColorBackground() *ColorBackground {
	return &ColorBackground{
		color:   LightColor,
		padding: image.Point{X: 20, Y: 20},
	}
}

// SetColor sets the background color
func (bg *ColorBackground) SetColor(c color.Color) *ColorBackground {
	bg.color = c
	return bg
}

// SetPadding sets the padding for the background
// horizontal and vertical are in pixels
func (bg *ColorBackground) SetPadding(horizontal, vertical int) *ColorBackground {
	bg.padding = image.Point{X: horizontal, Y: vertical}
	return bg
}

// Apply applies the background to the given content image
// It returns a new image with the background applied and the content centered
func (bg *ColorBackground) Apply(content image.Image) image.Image {
	bounds := content.Bounds()

	// Calculate new dimensions including padding
	width := bounds.Dx() + (bg.padding.X * 2)  // horizontal padding on both sides
	height := bounds.Dy() + (bg.padding.Y * 2) // vertical padding on top and bottom

	// Create new image with background color
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with background color
	draw.Draw(img, img.Bounds(), &image.Uniform{bg.color}, image.Point{}, draw.Src)

	// Draw the content in the center (accounting for padding)
	contentRect := image.Rect(
		bg.padding.X,
		bg.padding.Y,
		width-bg.padding.X,
		height-bg.padding.Y,
	)
	draw.Draw(img, contentRect, content, bounds.Min, draw.Over)

	return img
}
