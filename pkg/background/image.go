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
	blurRadius   float64
	opacity      float64
	padding      Padding
	cornerRadius float64
}

// NewImageBackground creates a new ImageBackground
func NewImageBackground(img image.Image) ImageBackground {
	return ImageBackground{
		image:        img,
		scaleMode:    ImageScaleFit,
		blurRadius:   0,
		opacity:      1.0,
		padding:      NewPadding(20),
		cornerRadius: 0,
	}
}

// SetScaleMode sets the scaling mode for the image
func (bg ImageBackground) SetScaleMode(mode ImageScaleMode) ImageBackground {
	bg.scaleMode = mode
	return bg
}

// SetBlurRadius sets the blur radius for the background image
func (bg ImageBackground) SetBlurRadius(radius float64) ImageBackground {
	bg.blurRadius = radius
	return bg
}

// SetOpacity sets the opacity of the background image (0.0 - 1.0)
func (bg ImageBackground) SetOpacity(opacity float64) ImageBackground {
	bg.opacity = math.Max(0, math.Min(1, opacity))
	return bg
}

// SetPadding sets equal padding for all sides
func (bg ImageBackground) SetPadding(value int) ImageBackground {
	bg.padding = NewPadding(value)
	return bg
}

// SetPaddingDetailed sets detailed padding for each side
func (bg ImageBackground) SetPaddingDetailed(top, right, bottom, left int) ImageBackground {
	bg.padding = Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return bg
}

// SetCornerRadius sets the corner radius for the background
func (bg ImageBackground) SetCornerRadius(radius float64) Background {
	bg.cornerRadius = radius
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

// Render applies the image background to the given content image
func (bg ImageBackground) Render(content image.Image) image.Image {
	bounds := content.Bounds()
	width := bounds.Dx() + bg.padding.Left + bg.padding.Right
	height := bounds.Dy() + bg.padding.Top + bg.padding.Bottom

	// Create a new RGBA image for the background
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Scale the background image
	scaled := bg.scaleImage(width, height)

	// Apply blur if requested
	if bg.blurRadius > 0 {
		scaled = imaging.Blur(scaled, bg.blurRadius)
	}

	// Create a mask for rounded corners if needed
	var mask *image.Alpha
	if bg.cornerRadius > 0 {
		mask = image.NewAlpha(dst.Bounds())
		drawRoundedRect(mask, dst.Bounds(), color.Alpha{A: 255}, bg.cornerRadius)
	}

	// Convert scaled image to RGBA if it isn't already
	scaledRGBA := image.NewRGBA(scaled.Bounds())
	draw.Draw(scaledRGBA, scaledRGBA.Bounds(), scaled, scaled.Bounds().Min, draw.Src)

	// Apply opacity if needed
	if bg.opacity < 1.0 {
		for y := 0; y < scaledRGBA.Bounds().Dy(); y++ {
			for x := 0; x < scaledRGBA.Bounds().Dx(); x++ {
				c := scaledRGBA.RGBAAt(x, y)
				// For JPEG images, ensure full opacity before applying the opacity factor
				if _, ok := bg.image.(*image.YCbCr); ok {
					c.A = 255
				}
				c.A = uint8(float64(c.A) * bg.opacity)
				scaledRGBA.Set(x, y, c)
			}
		}
	} else if _, ok := bg.image.(*image.YCbCr); ok {
		// For JPEG images with no opacity, ensure full opacity
		for y := 0; y < scaledRGBA.Bounds().Dy(); y++ {
			for x := 0; x < scaledRGBA.Bounds().Dx(); x++ {
				c := scaledRGBA.RGBAAt(x, y)
				c.A = 255
				scaledRGBA.Set(x, y, c)
			}
		}
	}

	// Draw the background image
	if mask != nil {
		draw.DrawMask(dst, dst.Bounds(), scaledRGBA, scaledRGBA.Bounds().Min, mask, mask.Bounds().Min, draw.Over)
	} else {
		draw.Draw(dst, dst.Bounds(), scaledRGBA, scaledRGBA.Bounds().Min, draw.Over)
	}

	// Draw the content centered on the background
	contentPos := image.Point{
		X: bg.padding.Left,
		Y: bg.padding.Top,
	}
	contentRect := image.Rectangle{
		Min: contentPos,
		Max: image.Point{
			X: contentPos.X + bounds.Dx(),
			Y: contentPos.Y + bounds.Dy(),
		},
	}
	draw.Draw(dst, contentRect, content, bounds.Min, draw.Over)

	return dst
}
