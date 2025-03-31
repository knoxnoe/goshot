package chrome

import (
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

const (
	gnomeDefaultTitleBarHeight = 32
	gnomeDefaultControlSize    = 16
	gnomeDefaultControlSpacing = 8
	gnomeDefaultTitleFontSize  = 14
	gnomeDefaultControlPadding = 8
	gnomeDefaultCornerRadius   = 8.0
	adwaitaTitleBarHeight      = 45
	adwaitaControlSize         = 20
	adwaitaControlSpacing      = 16
	adwaitaLeftPadding         = 8
	adwaitaRightPadding        = 16
	adwaitaTitleFontSize       = 14
	adwaitaCornerRadius        = 12
)

// GNOMEStyle represents different GNOME UI styles
type GNOMEStyle string

const (
	GNOMEStyleAdwaita GNOMEStyle = "adwaita" // Adwaita style
	GNOMEStyleBreeze  GNOMEStyle = "breeze"  // Breeze style
)

// GNOMEChrome implements the Chrome interface with GNOME-style window decorations
type GNOMEChrome struct {
	theme        Theme
	cornerRadius float64
	title        string
	themeName    string
	variant      ThemeVariant
	titleBar     bool
	style        GNOMEStyle
}

func init() {
	// Register Adwaita theme
	registerAdwaitaTheme()
	// Register Breeze theme
	registerBreezeTheme()
}

func registerAdwaitaTheme() {
	lightTheme := Theme{
		Type:    ThemeTypeGNOME,
		Variant: ThemeVariantLight,
		Name:    "adwaita",
		Properties: ThemeProperties{
			TitleFont:          "Cantarell",
			TitleFontSize:      adwaitaTitleFontSize,
			TitleBackground:    color.RGBA{R: 242, G: 242, B: 242, A: 255},
			TitleText:          color.RGBA{R: 40, G: 40, B: 40, A: 255},
			ControlsColor:      color.RGBA{R: 40, G: 40, B: 40, A: 255},
			ContentBackground:  color.White,
			TextColor:          color.RGBA{R: 40, G: 40, B: 40, A: 255},
			AccentColor:        color.RGBA{R: 53, G: 132, B: 228, A: 255},
			BorderColor:        color.RGBA{R: 220, G: 220, B: 220, A: 255},
			InactiveTitleBg:    color.RGBA{R: 247, G: 247, B: 247, A: 255},
			InactiveTitleText:  color.RGBA{R: 120, G: 120, B: 120, A: 255},
			ButtonHoverColor:   color.RGBA{R: 230, G: 230, B: 230, A: 255},
			ButtonPressedColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			CornerRadius:       adwaitaCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": GNOMEStyleAdwaita,
			},
		},
	}

	darkTheme := Theme{
		Type:    ThemeTypeGNOME,
		Variant: ThemeVariantDark,
		Name:    "adwaita",
		Properties: ThemeProperties{
			TitleFont:          "Cantarell",
			TitleFontSize:      adwaitaTitleFontSize,
			TitleBackground:    color.RGBA{R: 36, G: 36, B: 36, A: 255},
			TitleText:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground:  color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TextColor:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			AccentColor:        color.RGBA{R: 53, G: 132, B: 228, A: 255},
			BorderColor:        color.RGBA{R: 22, G: 22, B: 22, A: 255},
			InactiveTitleBg:    color.RGBA{R: 42, G: 42, B: 42, A: 255},
			InactiveTitleText:  color.RGBA{R: 180, G: 180, B: 180, A: 255},
			ButtonHoverColor:   color.RGBA{R: 50, G: 50, B: 50, A: 255},
			ButtonPressedColor: color.RGBA{R: 60, G: 60, B: 60, A: 255},
			CornerRadius:       adwaitaCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": GNOMEStyleAdwaita,
			},
		},
	}

	DefaultRegistry.RegisterTheme(ThemeTypeGNOME, "adwaita", ThemeVariantLight, lightTheme)
	DefaultRegistry.RegisterTheme(ThemeTypeGNOME, "adwaita", ThemeVariantDark, darkTheme)
}

func registerBreezeTheme() {
	lightTheme := Theme{
		Type:    ThemeTypeGNOME,
		Variant: ThemeVariantLight,
		Name:    "breeze",
		Properties: ThemeProperties{
			TitleFont:          "Cantarell",
			TitleFontSize:      gnomeDefaultTitleFontSize,
			TitleBackground:    color.RGBA{R: 205, G: 209, B: 214, A: 255},
			TitleText:          color.RGBA{R: 109, G: 113, B: 120, A: 255},
			ControlsColor:      color.RGBA{R: 40, G: 40, B: 40, A: 255},
			ContentBackground:  color.White,
			TextColor:          color.RGBA{R: 40, G: 40, B: 40, A: 255},
			AccentColor:        color.RGBA{R: 53, G: 132, B: 228, A: 255},
			BorderColor:        color.RGBA{R: 220, G: 220, B: 220, A: 255},
			InactiveTitleBg:    color.RGBA{R: 247, G: 247, B: 247, A: 255},
			InactiveTitleText:  color.RGBA{R: 120, G: 120, B: 120, A: 255},
			ButtonHoverColor:   color.RGBA{R: 230, G: 230, B: 230, A: 255},
			ButtonPressedColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			CornerRadius:       adwaitaCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": GNOMEStyleBreeze,
			},
		},
	}

	darkTheme := Theme{
		Type:    ThemeTypeGNOME,
		Variant: ThemeVariantDark,
		Name:    "breeze",
		Properties: ThemeProperties{
			TitleFont:          "Cantarell",
			TitleFontSize:      gnomeDefaultTitleFontSize,
			TitleBackground:    color.RGBA{R: 68, G: 82, B: 91, A: 255},
			TitleText:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground:  color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TextColor:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			AccentColor:        color.RGBA{R: 53, G: 132, B: 228, A: 255},
			BorderColor:        color.RGBA{R: 22, G: 22, B: 22, A: 255},
			InactiveTitleBg:    color.RGBA{R: 42, G: 42, B: 42, A: 255},
			InactiveTitleText:  color.RGBA{R: 180, G: 180, B: 180, A: 255},
			ButtonHoverColor:   color.RGBA{R: 50, G: 50, B: 50, A: 255},
			ButtonPressedColor: color.RGBA{R: 60, G: 60, B: 60, A: 255},
			CornerRadius:       adwaitaCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": GNOMEStyleBreeze,
			},
		},
	}

	DefaultRegistry.RegisterTheme(ThemeTypeGNOME, "breeze", ThemeVariantLight, lightTheme)
	DefaultRegistry.RegisterTheme(ThemeTypeGNOME, "breeze", ThemeVariantDark, darkTheme)
}

// NewGNOMEChrome creates a new GNOME-style window chrome
func NewGNOMEChrome(style GNOMEStyle, opts ...ChromeOption) *GNOMEChrome {
	chrome := &GNOMEChrome{
		cornerRadius: gnomeDefaultCornerRadius,
		title:        "",
		titleBar:     true,
		themeName:    "adwaita",
		variant:      ThemeVariantLight,
		style:        style,
	}

	// Set initial theme
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeGNOME, "adwaita", ThemeVariantLight); ok {
		chrome.theme = theme
	}

	// Apply options
	for _, opt := range opts {
		chrome = opt(chrome).(*GNOMEChrome)
	}

	return chrome
}

func (c *GNOMEChrome) WithTheme(theme Theme) Chrome {
	c.theme = theme
	c.themeName = theme.Name
	c.variant = theme.Variant
	if style, ok := theme.Properties.CustomProperties["style"].(GNOMEStyle); ok {
		c.style = style
	}
	return c
}

func (c *GNOMEChrome) WithThemeByName(name string, variant ThemeVariant) Chrome {
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeGNOME, name, variant); ok {
		c.themeName = name
		c.variant = variant
		c.theme = theme
		if style, ok := theme.Properties.CustomProperties["style"].(GNOMEStyle); ok {
			c.style = style
		}
	}
	return c
}

func (c *GNOMEChrome) GetCurrentThemeName() string {
	return c.themeName
}

func (c *GNOMEChrome) GetCurrentVariant() ThemeVariant {
	return c.variant
}

func (c *GNOMEChrome) WithVariant(variant ThemeVariant) Chrome {
	return c.WithThemeByName(c.themeName, variant)
}

func (c *GNOMEChrome) CurrentTheme() Theme {
	return c.theme
}

func (c *GNOMEChrome) WithTitle(title string) Chrome {
	c.title = title
	return c
}

func (c *GNOMEChrome) WithCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *GNOMEChrome) WithTitleBar(enabled bool) Chrome {
	c.titleBar = enabled
	return c
}

// Render implements the Chrome interface
func (c *GNOMEChrome) Render(content image.Image) (image.Image, error) {
	content, width, height := contentOrBlank(c, content)

	titleBarHeight := c.titleBarHeight()

	// Create context for drawing
	dc := gg.NewContext(width, height+titleBarHeight)

	// Draw the base window with rounded corners
	if err := DrawWindowBase(dc, width, height+titleBarHeight, c.cornerRadius,
		c.theme.Properties.TitleBackground,
		c.theme.Properties.TitleBackground,
		titleBarHeight); err != nil {
		return nil, err
	}

	if c.titleBar {
		// Draw title text if enabled
		if c.title != "" {
			DrawTitleText(dc, c.title, width, titleBarHeight, c.theme.Properties.TitleText, c.theme.Properties.TitleFontSize, c.theme.Properties.TitleFont)
		}

		// Draw window controls based on style
		c.renderWindowControls(dc, width, titleBarHeight)
	}

	// Draw content
	dc.DrawImage(content, 0, titleBarHeight)

	return dc.Image(), nil
}

func (c *GNOMEChrome) renderWindowControls(dc *gg.Context, width, titleBarHeight int) {
	switch c.style {
	case GNOMEStyleAdwaita:
		c.renderAdwaitaControls(dc, width, titleBarHeight)
	case GNOMEStyleBreeze:
		c.renderBreezeControls(dc, width, titleBarHeight)
	default:
		c.renderAdwaitaControls(dc, width, titleBarHeight)
	}
}

func (c *GNOMEChrome) renderAdwaitaControls(dc *gg.Context, width, titleBarHeight int) {
	controlY := float64(titleBarHeight-adwaitaControlSize) / 2
	closeX := float64(width - adwaitaRightPadding - adwaitaControlSize)
	maximizeX := closeX - float64(adwaitaControlSize) - float64(adwaitaControlSpacing)
	minimizeX := maximizeX - float64(adwaitaControlSize) - float64(adwaitaControlSpacing)

	buttonSize := float64(adwaitaControlSize)
	iconSize := buttonSize * 0.4
	strokeWidth := 2.0

	dc.SetLineWidth(strokeWidth)
	dc.SetColor(c.theme.Properties.ControlsColor)

	// Close icon (X) - slightly smaller
	closeIconSize := buttonSize * 0.45 // Reduced from 0.5 (half of button) to 0.45
	closeOffset := (buttonSize - closeIconSize) / 2
	dc.MoveTo(closeX+closeOffset, controlY+closeOffset)
	dc.LineTo(closeX+closeOffset+closeIconSize, controlY+closeOffset+closeIconSize)
	dc.MoveTo(closeX+closeOffset, controlY+closeOffset+closeIconSize)
	dc.LineTo(closeX+closeOffset+closeIconSize, controlY+closeOffset)
	dc.Stroke()

	// Maximize icon (rectangle)
	dc.DrawRectangle(maximizeX+buttonSize*0.3, controlY+buttonSize*0.3, iconSize, iconSize)
	dc.Stroke()

	// Minimize icon (line) - aligned with bottom of maximize
	minimizeY := controlY + buttonSize*0.65 // Aligned with bottom of maximize icon
	dc.MoveTo(minimizeX+buttonSize*0.25, minimizeY)
	dc.LineTo(minimizeX+buttonSize*0.75, minimizeY)
	dc.Stroke()
}

func (c *GNOMEChrome) renderBreezeControls(dc *gg.Context, width, titleBarHeight int) {
	controlY := float64(titleBarHeight-gnomeDefaultControlSize) / 2
	closeX := float64(width) - float64(gnomeDefaultControlSize) - float64(gnomeDefaultControlPadding)
	minimizeX := closeX - float64(gnomeDefaultControlSize) - float64(gnomeDefaultControlSpacing)

	// Draw controls
	dc.SetColor(c.theme.Properties.ControlsColor)

	// Close button circle with X
	dc.DrawCircle(closeX+float64(gnomeDefaultControlSize)/2, controlY+float64(gnomeDefaultControlSize)/2, float64(gnomeDefaultControlSize)/2)
	dc.Fill()

	// Draw X inside close button (rotated 45 degrees)
	xPadding := float64(gnomeDefaultControlSize) / 4
	centerX := closeX + float64(gnomeDefaultControlSize)/2
	centerY := controlY + float64(gnomeDefaultControlSize)/2
	xSize := float64(gnomeDefaultControlSize)/2 - xPadding

	// Draw rotated X with title background color
	dc.SetColor(c.theme.Properties.TitleBackground)
	dc.DrawLine(
		centerX-xSize/math.Sqrt2,
		centerY-xSize/math.Sqrt2,
		centerX+xSize/math.Sqrt2,
		centerY+xSize/math.Sqrt2,
	)
	dc.DrawLine(
		centerX-xSize/math.Sqrt2,
		centerY+xSize/math.Sqrt2,
		centerX+xSize/math.Sqrt2,
		centerY-xSize/math.Sqrt2,
	)
	dc.Stroke()

	// Reset color for minimize button
	dc.SetColor(c.theme.Properties.ControlsColor)

	// Minimize button (downward caret)
	caretSize := float64(gnomeDefaultControlSize) / 2
	centerX = minimizeX + float64(gnomeDefaultControlSize)/2
	centerY = controlY + float64(gnomeDefaultControlSize)/2

	dc.DrawLine(centerX-caretSize/2, centerY-caretSize/4,
		centerX, centerY+caretSize/4)
	dc.DrawLine(centerX, centerY+caretSize/4,
		centerX+caretSize/2, centerY-caretSize/4)
	dc.Stroke()
}

func (c *GNOMEChrome) MinimumSize() (width, height int) {
	return 100, gnomeDefaultTitleBarHeight // Minimum size required for controls
}

func (c *GNOMEChrome) ContentInsets() (top, right, bottom, left int) {
	return c.titleBarHeight(), 0, 0, 0
}

func (c *GNOMEChrome) titleBarHeight() int {
	if c.titleBar {
		switch c.style {
		case GNOMEStyleAdwaita:
			return adwaitaTitleBarHeight
		default:
			return gnomeDefaultTitleBarHeight
		}
	}
	return 0
}
