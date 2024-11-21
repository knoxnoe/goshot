package chrome

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	blankDefaultCornerRadius = 6.0
)

// BlankChrome implements the Chrome interface with minimal decoration
type BlankChrome struct {
	theme        Theme
	cornerRadius float64
}

// NewBlankChrome creates a new blank window chrome
func NewBlankChrome(opts ...ChromeOption) *BlankChrome {
	c := &BlankChrome{
		cornerRadius: blankDefaultCornerRadius,
		theme: Theme{
			Type:    "blank",
			Variant: ThemeVariantLight,
			Name:    "blank",
			Properties: ThemeProperties{
				ContentBackground: color.White,
				CornerRadius:      blankDefaultCornerRadius,
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *BlankChrome) SetTheme(theme Theme) Chrome {
	c.theme = theme
	return c
}

func (c *BlankChrome) SetThemeByName(name string, variant ThemeVariant) Chrome {
	// Blank chrome only has one theme
	return c
}

func (c *BlankChrome) GetCurrentThemeName() string {
	return "blank"
}

func (c *BlankChrome) GetCurrentVariant() ThemeVariant {
	return c.theme.Variant
}

func (c *BlankChrome) SetVariant(variant ThemeVariant) Chrome {
	c.theme.Variant = variant
	return c
}

func (c *BlankChrome) CurrentTheme() Theme {
	return c.theme
}

func (c *BlankChrome) SetTitle(_ string) Chrome {
	// Blank chrome doesn't support titles
	return c
}

func (c *BlankChrome) SetCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *BlankChrome) SetTitleBar(_ bool) Chrome {
	// Blank chrome doesn't support title bars
	return c
}

func (c *BlankChrome) Render(content image.Image) (image.Image, error) {
	width := content.Bounds().Dx()
	height := content.Bounds().Dy()

	// Create context for drawing
	dc := gg.NewContext(width, height)

	// Draw the base window with rounded corners
	fmt.Printf("Blank chrome: width=%d, height=%d, cornerRadius=%f\n", width, height, c.cornerRadius)
	if err := DrawWindowBase(dc, width, height, c.cornerRadius,
		c.theme.Properties.ContentBackground,
		c.theme.Properties.ContentBackground,
		0); err != nil {
		return nil, err
	}

	// Draw content
	dc.DrawImage(content, 0, 0)

	return dc.Image(), nil
}

func (c *BlankChrome) MinimumSize() (width, height int) {
	return 100, 100 // Minimal reasonable size
}

func (c *BlankChrome) ContentInsets() (top, right, bottom, left int) {
	return 0, 0, 0, 0 // No insets in blank chrome
}
