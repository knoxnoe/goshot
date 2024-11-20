package background

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// ColorBackground represents a solid color background
type ColorBackground struct {
	color        color.Color
	padding      Padding
	cornerRadius float64
	shadow       Shadow
}

// NewColorBackground creates a new ColorBackground with the given color
func NewColorBackground() ColorBackground {
	return ColorBackground{
		color:        LightColor,
		padding:      NewPadding(20),
		cornerRadius: 0,
		shadow:       nil,
	}
}

// SetColor sets the background color
func (bg ColorBackground) SetColor(c color.Color) ColorBackground {
	bg.color = c
	return bg
}

// SetPadding sets equal padding for all sides
func (bg ColorBackground) SetPadding(value int) ColorBackground {
	bg.padding = NewPadding(value)
	return bg
}

// SetPaddingDetailed sets detailed padding for each side
func (bg ColorBackground) SetPaddingDetailed(top, right, bottom, left int) ColorBackground {
	bg.padding = Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return bg
}

// SetCornerRadius sets the corner radius for the background
func (bg ColorBackground) SetCornerRadius(radius float64) Background {
	bg.cornerRadius = radius
	return bg
}

// SetShadow sets the shadow configuration for the background
func (bg ColorBackground) SetShadow(shadow Shadow) Background {
	bg.shadow = shadow
	return bg
}

// drawRoundedRect draws a rounded rectangle on the destination image
func drawRoundedRect(dst draw.Image, r image.Rectangle, col color.Color, radius float64) {
	// Create a mask image for the rounded corners
	mask := image.NewAlpha(r)

	// Calculate center and radius of corner circles
	corners := []image.Point{
		{r.Min.X + int(radius), r.Min.Y + int(radius)},         // Top-left
		{r.Max.X - int(radius) - 1, r.Min.Y + int(radius)},     // Top-right
		{r.Min.X + int(radius), r.Max.Y - int(radius) - 1},     // Bottom-left
		{r.Max.X - int(radius) - 1, r.Max.Y - int(radius) - 1}, // Bottom-right
	}

	// Fill the mask
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			// Check if point is in a corner region
			var alpha uint8 = 255
			px := float64(x)
			py := float64(y)

			// For each corner
			for _, c := range corners {
				dx := px - float64(c.X)
				dy := py - float64(c.Y)
				dist := math.Sqrt(dx*dx + dy*dy)

				// If point is outside the circle
				if dist <= radius {
					// Point is inside the corner radius
					alpha = 255
				} else if (x < c.X && y < c.Y && c == corners[0]) || // Top-left
					(x > c.X && y < c.Y && c == corners[1]) || // Top-right
					(x < c.X && y > c.Y && c == corners[2]) || // Bottom-left
					(x > c.X && y > c.Y && c == corners[3]) { // Bottom-right
					alpha = 0
				}
			}

			mask.Set(x, y, color.Alpha{A: alpha})
		}
	}

	// Draw using the mask
	draw.DrawMask(dst, r, image.NewUniform(col), image.Point{}, mask, r.Min, draw.Over)
}

// Render applies the background to the given content image
// It returns a new image with the background applied and the content centered
func (bg ColorBackground) Render(content image.Image) image.Image {
	bounds := content.Bounds()

	// If shadow is configured, apply it to the content first
	if bg.shadow != nil {
		// Set the shadow's corner radius to match the background
		bg.shadow.(*shadowImpl).cornerRadius = bg.cornerRadius
		content = bg.shadow.Apply(content)
		bounds = content.Bounds() // Update bounds to include shadow
	}

	// Calculate new dimensions including padding
	width := bounds.Dx() + bg.padding.Left + bg.padding.Right
	height := bounds.Dy() + bg.padding.Top + bg.padding.Bottom

	// Create new image with transparent background
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Draw rounded rectangle background
	if bg.cornerRadius > 0 {
		drawRoundedRect(img, img.Bounds(), bg.color, bg.cornerRadius)
	} else {
		draw.Draw(img, img.Bounds(), &image.Uniform{bg.color}, image.Point{}, draw.Src)
	}

	// Draw the content with shadow in the center (accounting for padding)
	contentRect := image.Rect(
		bg.padding.Left,
		bg.padding.Top,
		width-bg.padding.Right,
		height-bg.padding.Bottom,
	)
	draw.Draw(img, contentRect, content, bounds.Min, draw.Over)

	return img
}
