package chrome

import (
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

func (c *BlankChrome) WithTheme(theme Theme) Chrome {
	c.theme = theme
	return c
}

func (c *BlankChrome) WithThemeByName(name string, variant ThemeVariant) Chrome {
	// Blank chrome only has one theme
	return c
}

func (c *BlankChrome) GetCurrentThemeName() string {
	return "blank"
}

func (c *BlankChrome) GetCurrentVariant() ThemeVariant {
	return c.theme.Variant
}

func (c *BlankChrome) WithVariant(variant ThemeVariant) Chrome {
	c.theme.Variant = variant
	return c
}

func (c *BlankChrome) CurrentTheme() Theme {
	return c.theme
}

func (c *BlankChrome) WithTitle(_ string) Chrome {
	// Blank chrome doesn't support titles
	return c
}

func (c *BlankChrome) WithCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *BlankChrome) WithTitleBar(_ bool) Chrome {
	// Blank chrome doesn't support title bars
	return c
}

func (c *BlankChrome) Render(content image.Image) (image.Image, error) {
	content, width, height := contentOrBlank(c, content)

	// Create context for drawing
	dc := gg.NewContext(width, height)

	// Draw the base window with rounded corners
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
