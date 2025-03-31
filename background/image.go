package background

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/disintegration/imaging"
)

// ImageScaleMode determines how the image is scaled to fit the background
type ImageScaleMode int

const (
	// ImageScaleFit scales the image to fit within the bounds while maintaining aspect ratio
	ImageScaleFit ImageScaleMode = iota
	// ImageScaleFill scales the image to fill the bounds while maintaining aspect ratio
	ImageScaleFill
	// ImageScaleCover scales the image to cover the entire area while maintaining aspect ratio (like CSS background-size: cover)
	ImageScaleCover
	// ImageScaleStretch stretches the image to exactly fit the bounds
	ImageScaleStretch
	// ImageScaleTile repeats the image to fill the bounds
	ImageScaleTile
)

// ImageBackground represents an image background
type ImageBackground struct {
	image        image.Image
	scaleMode    ImageScaleMode
	blur         *BlurConfig
	opacity      float64
	padding      Padding
	cornerRadius float64
	shadow       Shadow
}

// NewImageBackground creates a new ImageBackground
func NewImageBackground(img image.Image) ImageBackground {
	return ImageBackground{
		image:        img,
		scaleMode:    ImageScaleFit,
		blur:         nil,
		opacity:      1.0,
		padding:      NewPadding(20),
		cornerRadius: 0,
		shadow:       nil,
	}
}

// NewImageBackgroundFromFile creates a new ImageBackground from a file path
func NewImageBackgroundFromFile(path string) (ImageBackground, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return ImageBackground{}, err
	}
	return NewImageBackground(img), nil
}

// WithScaleMode sets the scaling mode for the image
func (bg ImageBackground) WithScaleMode(mode ImageScaleMode) ImageBackground {
	bg.scaleMode = mode
	return bg
}

// WithScaleModeString sets the scaling mode for the image from a string
func (bg ImageBackground) WithScaleModeString(mode string) ImageBackground {
	switch mode {
	case "fit":
		bg.scaleMode = ImageScaleFit
	case "fill":
		bg.scaleMode = ImageScaleFill
	case "cover":
		bg.scaleMode = ImageScaleCover
	case "stretch":
		bg.scaleMode = ImageScaleStretch
	case "tile":
		bg.scaleMode = ImageScaleTile
	}
	return bg
}

// WithBlur sets the blur configuration for the background image
func (bg ImageBackground) WithBlur(blurType BlurType, radius float64) ImageBackground {
	bg.blur = &BlurConfig{
		Type:   blurType,
		Radius: radius,
	}
	return bg
}

// WithOpacity sets the opacity of the background image (0.0 - 1.0)
func (bg ImageBackground) WithOpacity(opacity float64) ImageBackground {
	bg.opacity = math.Max(0, math.Min(1, opacity))
	return bg
}

// WithPadding sets equal padding for all sides
func (bg ImageBackground) WithPadding(value int) ImageBackground {
	bg.padding = NewPadding(value)
	return bg
}

// WithPaddingDetailed sets detailed padding for each side
func (bg ImageBackground) WithPaddingDetailed(top, right, bottom, left int) ImageBackground {
	bg.padding = Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return bg
}

// WithCornerRadius sets the corner radius for the background
func (bg ImageBackground) WithCornerRadius(radius float64) Background {
	bg.cornerRadius = radius
	return bg
}

// WithShadow sets the shadow configuration for the background
func (bg ImageBackground) WithShadow(shadow Shadow) Background {
	bg.shadow = shadow
	return bg
}

// scaleImage scales the image according to the scale mode
func (bg ImageBackground) scaleImage(width, height int) image.Image {
	bounds := bg.image.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	switch bg.scaleMode {
	case ImageScaleStretch:
		return imaging.Resize(bg.image, width, height, imaging.Lanczos)

	case ImageScaleFill, ImageScaleCover:
		srcRatio := float64(srcWidth) / float64(srcHeight)
		dstRatio := float64(width) / float64(height)

		var newWidth, newHeight int
		if srcRatio > dstRatio {
			newHeight = height
			newWidth = int(float64(height) * srcRatio)
		} else {
			newWidth = width
			newHeight = int(float64(width) / srcRatio)
		}
		scaled := imaging.Resize(bg.image, newWidth, newHeight, imaging.Lanczos)

		// For ImageScaleCover, we center crop the image
		if bg.scaleMode == ImageScaleCover {
			return imaging.CropCenter(scaled, width, height)
		}

		// For ImageScaleFill, we anchor to the top-left
		return imaging.Crop(scaled, image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{width, height},
		})

	case ImageScaleFit:
		srcRatio := float64(srcWidth) / float64(srcHeight)
		dstRatio := float64(width) / float64(height)

		var newWidth, newHeight int
		if srcRatio > dstRatio {
			newWidth = width
			newHeight = int(float64(width) / srcRatio)
		} else {
			newHeight = height
			newWidth = int(float64(height) * srcRatio)
		}
		scaled := imaging.Resize(bg.image, newWidth, newHeight, imaging.Lanczos)
		// Create a new image with the target size and center the scaled image
		centered := image.NewRGBA(image.Rect(0, 0, width, height))
		scaledBounds := scaled.Bounds()
		centerX := (width - scaledBounds.Dx()) / 2
		centerY := (height - scaledBounds.Dy()) / 2
		draw.Draw(centered, scaledBounds.Add(image.Point{centerX, centerY}), scaled, scaledBounds.Min, draw.Over)
		return centered

	case ImageScaleTile:
		// Create a new image to hold the tiled pattern
		tiled := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y += srcHeight {
			for x := 0; x < width; x += srcWidth {
				r := image.Rectangle{
					Min: image.Point{x, y},
					Max: image.Point{x + srcWidth, y + srcHeight},
				}
				draw.Draw(tiled, r, bg.image, bounds.Min, draw.Over)
			}
		}
		return tiled
	}

	return bg.image
}

// applyOpacity applies an opacity value to an RGBA image
func applyOpacity(img *image.RGBA, opacity float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			c.A = uint8(float64(c.A) * opacity)
			result.Set(x, y, c)
		}
	}

	return result
}

// Render applies the image background to the given content image
func (bg ImageBackground) Render(content image.Image) (image.Image, error) {
	if content == nil {
		width := bg.padding.Left + bg.padding.Right
		height := bg.padding.Top + bg.padding.Bottom
		content = image.NewRGBA(image.Rect(0, 0, width, height))
	}

	// Create a new image for the content with shadow
	contentWithShadow := content
	if bg.shadow != nil {
		contentWithShadow = bg.shadow.Apply(content)
	}

	// Calculate total size including padding and shadow bounds
	shadowBounds := contentWithShadow.Bounds()
	width := shadowBounds.Dx() + bg.padding.Left + bg.padding.Right
	height := shadowBounds.Dy() + bg.padding.Top + bg.padding.Bottom

	// Scale and process the background image
	scaledImg := bg.scaleImage(width, height)

	if bg.blur != nil {
		// Convert to NRGBA for imaging operations
		nrgba := imaging.Clone(scaledImg)

		switch bg.blur.Type {
		case GaussianBlur:
			nrgba = imaging.Blur(nrgba, bg.blur.Radius)
		case PixelatedBlur:
			// Create pixelated effect by scaling down and back up
			w := nrgba.Bounds().Dx()
			h := nrgba.Bounds().Dy()
			// Scale factor based on radius (larger radius = more pixelation)
			factor := math.Max(1, bg.blur.Radius)
			smallW := int(float64(w) / factor)
			smallH := int(float64(h) / factor)
			if smallW < 1 {
				smallW = 1
			}
			if smallH < 1 {
				smallH = 1
			}
			// Scale down
			small := imaging.Resize(nrgba, smallW, smallH, imaging.Box)
			// Scale back up
			nrgba = imaging.Resize(small, w, h, imaging.NearestNeighbor)
		}

		// Convert back to original format
		result := image.NewRGBA(scaledImg.Bounds())
		draw.Draw(result, result.Bounds(), nrgba, image.Point{}, draw.Src)
		scaledImg = result
	}

	// Create the final image
	result := image.NewRGBA(image.Rect(0, 0, width, height))

	// Draw the scaled background image
	draw.Draw(result, result.Bounds(), scaledImg, image.Point{}, draw.Over)

	// Apply opacity if needed
	if bg.opacity < 1.0 {
		result = applyOpacity(result, bg.opacity)
	}

	// Apply rounded corners if needed
	if bg.cornerRadius > 0 {
		mask := image.NewRGBA(result.Bounds())
		drawRoundedRect(mask, result.Bounds(), color.White, bg.cornerRadius)

		final := image.NewRGBA(result.Bounds())
		draw.DrawMask(final, result.Bounds(), result, image.Point{}, mask, image.Point{}, draw.Over)
		result = final
	}

	// Draw the content (with shadow) centered on the background
	contentPos := image.Point{
		X: bg.padding.Left - shadowBounds.Min.X,
		Y: bg.padding.Top - shadowBounds.Min.Y,
	}
	draw.Draw(result, shadowBounds.Add(contentPos), contentWithShadow, shadowBounds.Min, draw.Over)

	return result, nil
}
