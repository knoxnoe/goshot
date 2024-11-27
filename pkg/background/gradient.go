package background

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// GradientType represents the type of gradient
type GradientType int

const (
	// LinearGradient represents a linear gradient
	LinearGradient GradientType = iota
	// RadialGradient represents a radial gradient
	RadialGradient
	// AngularGradient represents an angular/conic gradient
	AngularGradient
	// DiamondGradient represents a diamond-shaped gradient
	DiamondGradient
	// SpiralGradient represents a spiral-shaped gradient
	SpiralGradient
	// SquareGradient represents a square-shaped gradient
	SquareGradient
	// StarGradient represents a star-shaped gradient
	StarGradient
)

// GradientStop represents a color stop in a gradient
type GradientStop struct {
	Color    color.Color
	Position float64 // Position between 0 and 1
}

// GradientBackground represents a gradient background
type GradientBackground struct {
	gradientType GradientType
	stops        []GradientStop
	angle        float64 // Angle in degrees for linear/angular gradients
	centerX      float64 // Center X position for radial/angular gradients (0-1)
	centerY      float64 // Center Y position for radial/angular gradients (0-1)
	intensity    float64 // Intensity modifier for special gradients (spiral tightness, star points)
	padding      Padding
	cornerRadius float64
	shadow       Shadow
}

// NewGradientBackground creates a new GradientBackground
func NewGradientBackground(gradientType GradientType, stops ...GradientStop) GradientBackground {
	return GradientBackground{
		gradientType: gradientType,
		stops:        stops,
		angle:        0,
		centerX:      0.5, // Default to center
		centerY:      0.5,
		intensity:    5.0, // Default intensity
		padding:      NewPadding(20),
		cornerRadius: 0,
		shadow:       nil,
	}
}

// WithAngle sets the angle for linear gradients (in degrees)
func (bg GradientBackground) WithAngle(angle float64) GradientBackground {
	bg.angle = angle
	return bg
}

// WithCenter sets the center point for radial and angular gradients
func (bg GradientBackground) WithCenter(x, y float64) GradientBackground {
	bg.centerX = x
	bg.centerY = y
	return bg
}

// WithIntensity sets the intensity modifier for special gradients
func (bg GradientBackground) WithIntensity(intensity float64) GradientBackground {
	bg.intensity = intensity
	return bg
}

// WithPadding sets equal padding for all sides
func (bg GradientBackground) WithPadding(value int) GradientBackground {
	bg.padding = NewPadding(value)
	return bg
}

// WithPaddingDetailed sets detailed padding for each side
func (bg GradientBackground) WithPaddingDetailed(top, right, bottom, left int) GradientBackground {
	bg.padding = Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return bg
}

// WithCornerRadius sets the corner radius for the background
func (bg GradientBackground) WithCornerRadius(radius float64) Background {
	bg.cornerRadius = radius
	return bg
}

// WithShadow sets the shadow configuration for the background
func (bg GradientBackground) WithShadow(shadow Shadow) Background {
	bg.shadow = shadow
	return bg
}

// interpolateColor interpolates between two colors based on t (0 to 1)
func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8(uint32((float64(r1)*(1-t) + float64(r2)*t)) >> 8)
	g := uint8(uint32((float64(g1)*(1-t) + float64(g2)*t)) >> 8)
	b := uint8(uint32((float64(b1)*(1-t) + float64(b2)*t)) >> 8)
	a := uint8(uint32((float64(a1)*(1-t) + float64(a2)*t)) >> 8)

	return color.RGBA{r, g, b, a}
}

// getColorAt returns the color at a specific position in the gradient
func (bg GradientBackground) getColorAt(pos float64) color.Color {
	if len(bg.stops) == 0 {
		return color.Black
	}
	if len(bg.stops) == 1 {
		return bg.stops[0].Color
	}

	// Find the two stops we're between
	var stop1, stop2 GradientStop
	for i := 0; i < len(bg.stops)-1; i++ {
		if pos >= bg.stops[i].Position && pos <= bg.stops[i+1].Position {
			stop1 = bg.stops[i]
			stop2 = bg.stops[i+1]
			break
		}
	}

	// Calculate the interpolation factor
	t := (pos - stop1.Position) / (stop2.Position - stop1.Position)
	return interpolateColor(stop1.Color, stop2.Color, t)
}

// Render applies the gradient background to the given content image
func (bg GradientBackground) Render(content image.Image) (image.Image, error) {
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

	// Create a new RGBA image for the gradient
	gradientImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create a mask for rounded corners if needed
	var mask *image.Alpha
	if bg.cornerRadius > 0 {
		mask = image.NewAlpha(gradientImg.Bounds())
		drawRoundedRect(mask, gradientImg.Bounds(), color.Alpha{A: 255}, bg.cornerRadius)
	}

	// Calculate center coordinates in pixels
	centerX := float64(width) * bg.centerX
	centerY := float64(height) * bg.centerY

	// Draw the gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var pos float64
			switch bg.gradientType {
			case LinearGradient:
				// Convert angle to radians and calculate position
				angleRad := bg.angle * math.Pi / 180
				// Project point onto gradient line
				pos = (float64(x)*math.Cos(angleRad) + float64(y)*math.Sin(angleRad)) /
					(float64(width)*math.Abs(math.Cos(angleRad)) + float64(height)*math.Abs(math.Sin(angleRad)))
			case RadialGradient:
				// Calculate distance from center
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				// Calculate maximum possible distance from center to any corner
				maxDist := math.Sqrt(math.Max(centerX*centerX, math.Pow(float64(width)-centerX, 2)) +
					math.Max(centerY*centerY, math.Pow(float64(height)-centerY, 2)))
				pos = math.Sqrt(dx*dx+dy*dy) / maxDist
			case AngularGradient:
				// Calculate angle from center
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				angle := math.Atan2(dy, dx)
				// Convert to degrees and normalize to 0-1 range
				pos = math.Mod(((angle/math.Pi+1)/2 + bg.angle/360), 1)
			case DiamondGradient:
				// Calculate Manhattan distance from center
				dx := math.Abs(float64(x) - centerX)
				dy := math.Abs(float64(y) - centerY)
				// Normalize by the maximum possible Manhattan distance
				maxDist := math.Max(centerX, float64(width)-centerX) + math.Max(centerY, float64(height)-centerY)
				pos = (dx + dy) / maxDist
			case SpiralGradient:
				// Calculate polar coordinates
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				angle := math.Atan2(dy, dx)
				dist := math.Sqrt(dx*dx + dy*dy)
				maxDist := math.Sqrt(centerX*centerX + centerY*centerY)
				// Combine angle and distance for spiral effect
				pos = math.Mod(((angle/math.Pi+1)/2 + dist*bg.intensity/maxDist + bg.angle/360), 1)
			case SquareGradient:
				// Calculate distance using max norm (L∞)
				dx := math.Abs(float64(x) - centerX)
				dy := math.Abs(float64(y) - centerY)
				maxDist := math.Max(centerX, float64(width)-centerX)
				pos = math.Max(dx, dy) / maxDist
			case StarGradient:
				// Calculate polar coordinates
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				angle := math.Atan2(dy, dx)
				dist := math.Sqrt(dx*dx + dy*dy)
				maxDist := math.Sqrt(centerX*centerX + centerY*centerY)
				// Create star pattern using sine wave
				points := bg.intensity // Number of star points
				starFactor := math.Abs(math.Sin(angle*points + bg.angle*math.Pi/180))
				pos = (dist/maxDist + starFactor*0.5) / 1.5
			}

			// Clamp position between 0 and 1
			pos = math.Max(0, math.Min(1, pos))
			c := bg.getColorAt(pos)

			if bg.cornerRadius > 0 {
				// Apply the rounded corner mask
				_, _, _, a := mask.At(x, y).RGBA()
				if a > 0 {
					gradientImg.Set(x, y, c)
				}
			} else {
				gradientImg.Set(x, y, c)
			}
		}
	}

	// Draw the content (with shadow) centered on the background
	contentPos := image.Point{
		X: bg.padding.Left - shadowBounds.Min.X,
		Y: bg.padding.Top - shadowBounds.Min.Y,
	}
	draw.Draw(gradientImg, shadowBounds.Add(contentPos), contentWithShadow, shadowBounds.Min, draw.Over)

	return gradientImg, nil
}
