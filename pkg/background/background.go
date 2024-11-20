package background

import (
	"image"
	"image/color"
)

// Background represents any type that can be used as a background
type Background interface {
	// Render applies the background to the given content image
	// It returns a new image with the background applied
	Render(content image.Image) image.Image

	// SetCornerRadius sets the corner radius for the background
	SetCornerRadius(radius float64) Background

	// SetShadow sets the shadow configuration for the background
	SetShadow(shadow Shadow) Background
}

var (
	// DarkColor is the default dark mode background color
	DarkColor = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	// LightColor is the default light mode background color
	LightColor = color.RGBA{R: 240, G: 240, B: 240, A: 255}
)

// Padding represents padding values for a background
type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// NewPadding creates a new Padding with equal values on all sides
func NewPadding(value int) Padding {
	return Padding{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}

// NewPaddingHV creates a new Padding with equal horizontal and vertical values
func NewPaddingHV(horizontal, vertical int) Padding {
	return Padding{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}

// ToPoint converts the padding to an image.Point for compatibility
// This uses the horizontal (Left) and vertical (Top) values
func (p Padding) ToPoint() image.Point {
	return image.Point{X: p.Left, Y: p.Top}
}
