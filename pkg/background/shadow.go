package background

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Shadow represents a shadow configuration that can be applied to a background
type Shadow interface {
	// SetOffset sets the X and Y offset of the shadow
	SetOffset(x, y float64) Shadow

	// SetBlur sets the blur radius of the shadow
	SetBlur(radius float64) Shadow

	// SetSpread sets the spread radius of the shadow
	SetSpread(radius float64) Shadow

	// SetColor sets the color of the shadow
	SetColor(c color.Color) Shadow

	// SetCornerRadius sets the corner radius of the shadow
	SetCornerRadius(radius float64) Shadow

	// Apply applies the shadow effect to the given image
	Apply(img image.Image) image.Image
}

// shadowImpl is the implementation of the Shadow interface
type shadowImpl struct {
	offsetX      float64
	offsetY      float64
	blur         float64
	spread       float64
	color        color.Color
	cornerRadius float64 // Added to match content's corner radius
}

// NewShadow creates a new shadow with default values
func NewShadow() Shadow {
	return &shadowImpl{
		offsetX:      5,
		offsetY:      5,
		blur:         10,
		spread:       0,
		color:        color.RGBA{0, 0, 0, 128},
		cornerRadius: 0,
	}
}

func (s *shadowImpl) SetOffset(x, y float64) Shadow {
	s.offsetX = x
	s.offsetY = y
	return s
}

func (s *shadowImpl) SetBlur(radius float64) Shadow {
	s.blur = radius
	return s
}

func (s *shadowImpl) SetSpread(radius float64) Shadow {
	s.spread = radius
	return s
}

func (s *shadowImpl) SetColor(c color.Color) Shadow {
	s.color = c
	return s
}

func (s *shadowImpl) SetCornerRadius(radius float64) Shadow {
	s.cornerRadius = radius
	return s
}

func (s *shadowImpl) Apply(img image.Image) image.Image {
	bounds := img.Bounds()

	// Calculate the expanded bounds to accommodate shadow and offset
	maxOffset := math.Max(math.Abs(s.offsetX), math.Abs(s.offsetY))
	expandBy := int(math.Ceil(s.blur + s.spread + maxOffset))

	// Create new bounds that can accommodate the shadow in any direction
	newBounds := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: bounds.Dx() + (expandBy * 2), Y: bounds.Dy() + (expandBy * 2)},
	}

	// Create a new RGBA image for the final result
	shadowImg := image.NewRGBA(newBounds)

	// Create the shadow mask with same dimensions
	shadowMask := image.NewRGBA(newBounds)

	// Content is always centered in the new bounds
	contentX := expandBy
	contentY := expandBy

	// Calculate shadow bounds relative to content position, including spread
	shadowBounds := image.Rectangle{
		Min: image.Point{
			X: contentX + int(s.offsetX) - int(s.spread),
			Y: contentY + int(s.offsetY) - int(s.spread),
		},
		Max: image.Point{
			X: contentX + bounds.Dx() + int(s.offsetX) + int(s.spread),
			Y: contentY + bounds.Dy() + int(s.offsetY) + int(s.spread),
		},
	}

	// Draw the rounded rectangle for the shadow with adjusted corner radius for spread
	cornerRadius := s.cornerRadius
	if s.spread > 0 {
		cornerRadius += s.spread
	}
	drawRoundedRect(shadowMask, shadowBounds, s.color, cornerRadius)

	// Apply gaussian blur to the shadow mask
	blurredShadow := applyGaussianBlur(shadowMask, s.blur)

	// Draw the blurred shadow
	draw.Draw(shadowImg, newBounds, blurredShadow, image.Point{}, draw.Over)

	// Draw the content centered
	contentBounds := image.Rectangle{
		Min: image.Point{X: contentX, Y: contentY},
		Max: image.Point{X: contentX + bounds.Dx(), Y: contentY + bounds.Dy()},
	}
	draw.Draw(shadowImg, contentBounds, img, bounds.Min, draw.Over)

	return shadowImg
}

// applyGaussianBlur applies a gaussian blur effect to the input image
func applyGaussianBlur(img *image.RGBA, radius float64) *image.RGBA {
	if radius <= 0 {
		return img
	}

	// Calculate kernel size (must be odd)
	size := int(math.Ceil(radius * 6))
	if size%2 == 0 {
		size++
	}

	// Create gaussian kernel
	kernel := makeGaussianKernel(radius, size)

	// Apply horizontal blur
	horizontal := horizontalBlur(img, kernel, size)

	// Apply vertical blur
	return verticalBlur(horizontal, kernel, size)
}

// makeGaussianKernel creates a 1D gaussian kernel
func makeGaussianKernel(radius float64, size int) []float64 {
	kernel := make([]float64, size)
	sigma := radius / 2
	twoSigmaSquare := 2 * sigma * sigma
	sum := 0.0

	mid := size / 2
	for i := 0; i < size; i++ {
		x := float64(i - mid)
		kernel[i] = math.Exp(-(x * x) / (twoSigmaSquare))
		sum += kernel[i]
	}

	// Normalize kernel
	for i := range kernel {
		kernel[i] /= sum
	}

	return kernel
}

// horizontalBlur applies the gaussian kernel horizontally
func horizontalBlur(img *image.RGBA, kernel []float64, size int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	mid := size / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a float64

			for k := 0; k < size; k++ {
				kx := x + k - mid
				if kx < bounds.Min.X {
					kx = bounds.Min.X
				}
				if kx >= bounds.Max.X {
					kx = bounds.Max.X - 1
				}

				c := img.RGBAAt(kx, y)
				weight := kernel[k]
				r += float64(c.R) * weight
				g += float64(c.G) * weight
				b += float64(c.B) * weight
				a += float64(c.A) * weight
			}

			result.Set(x, y, color.RGBA{
				R: uint8(math.Min(math.Max(r, 0), 255)),
				G: uint8(math.Min(math.Max(g, 0), 255)),
				B: uint8(math.Min(math.Max(b, 0), 255)),
				A: uint8(math.Min(math.Max(a, 0), 255)),
			})
		}
	}

	return result
}

// verticalBlur applies the gaussian kernel vertically
func verticalBlur(img *image.RGBA, kernel []float64, size int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	mid := size / 2

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			var r, g, b, a float64

			for k := 0; k < size; k++ {
				ky := y + k - mid
				if ky < bounds.Min.Y {
					ky = bounds.Min.Y
				}
				if ky >= bounds.Max.Y {
					ky = bounds.Max.Y - 1
				}

				c := img.RGBAAt(x, ky)
				weight := kernel[k]
				r += float64(c.R) * weight
				g += float64(c.G) * weight
				b += float64(c.B) * weight
				a += float64(c.A) * weight
			}

			result.Set(x, y, color.RGBA{
				R: uint8(math.Min(math.Max(r, 0), 255)),
				G: uint8(math.Min(math.Max(g, 0), 255)),
				B: uint8(math.Min(math.Max(b, 0), 255)),
				A: uint8(math.Min(math.Max(a, 0), 255)),
			})
		}
	}

	return result
}
