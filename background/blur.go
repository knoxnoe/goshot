package background

// BlurType represents the type of blur algorithm to use
type BlurType int

const (
	// GaussianBlur represents a gaussian blur algorithm
	GaussianBlur BlurType = iota
	// PixelatedBlur represents a pixelated blur effect
	PixelatedBlur
)

// BlurConfig represents the configuration for blur effects
type BlurConfig struct {
	Type   BlurType
	Radius float64
}
