package chrome

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	macDefaultTitleBarHeight = 30
	macDefaultControlSize    = 14
	macDefaultControlSpacing = 8
	macDefaultTitleFontSize  = 13
	macDefaultControlPadding = 8
	macDefaultCornerRadius   = 6.0
)

// MacStyle represents different macOS UI styles
type MacStyle string

const (
	MacStyleSequoia      MacStyle = "sequoia"      // macOS 15
	MacStyleSonoma       MacStyle = "sonoma"       // macOS 14
	MacStyleVentura      MacStyle = "ventura"      // macOS 13
	MacStyleMonterey     MacStyle = "monterey"     // macOS 12
	MacStyleBigSur       MacStyle = "bigsur"       // macOS 11
	MacStyleCatalina     MacStyle = "catalina"     // macOS 10.15
	MacStyleMojave       MacStyle = "mojave"       // macOS 10.14
	MacStyleHighSierra   MacStyle = "highsierra"   // macOS 10.13
	MacStyleSierra       MacStyle = "sierra"       // macOS 10.12
	MacStyleElCapitan    MacStyle = "elcapitan"    // OS X 10.11
	MacStyleYosemite     MacStyle = "yosemite"     // OS X 10.10
	MacStyleMavericks    MacStyle = "mavericks"    // OS X 10.9
	MacStyleMountainLion MacStyle = "mountainlion" // OS X 10.8
	MacStyleLion         MacStyle = "lion"         // OS X 10.7
	MacStyleSnowLeopard  MacStyle = "snowleopard"  // Mac OS X 10.6
)

// MacChrome implements the Chrome interface with macOS-style window decorations
type MacChrome struct {
	theme        Theme
	cornerRadius float64
	title        string
	themeName    string
	variant      ThemeVariant
	titleBar     bool
	style        MacStyle
}

func init() {
	// Register modern macOS themes (Sequoia)
	registerMacSequoiaThemes()
	// Register modern macOS themes (Ventura)
	registerMacVenturaThemes()
	// Register Big Sur themes
	registerMacBigSurThemes()
	// Register Catalina themes
	registerMacCatalinaThemes()
	// Register older macOS themes
	registerMacLegacyThemes()
}

func registerMacSequoiaThemes() {
	// Sequoia theme (default)
	lightTheme := Theme{
		Type:    ThemeTypeMac,
		Variant: ThemeVariantLight,
		Name:    "sequoia",
		Properties: ThemeProperties{
			TitleFont:          "Inter",
			TitleBackground:    color.RGBA{R: 236, G: 236, B: 236, A: 255},
			TitleText:          color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ControlsColor:      color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ContentBackground:  color.White,
			TextColor:          color.RGBA{R: 76, G: 76, B: 76, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 122, B: 255, A: 255},
			BorderColor:        color.RGBA{R: 220, G: 220, B: 220, A: 255},
			InactiveTitleBg:    color.RGBA{R: 246, G: 246, B: 246, A: 255},
			InactiveTitleText:  color.RGBA{R: 161, G: 161, B: 161, A: 255},
			ButtonHoverColor:   color.RGBA{R: 230, G: 230, B: 230, A: 255},
			ButtonPressedColor: color.RGBA{R: 210, G: 210, B: 210, A: 255},
			CornerRadius:       macDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": MacStyleSequoia,
			},
		},
	}

	darkTheme := Theme{
		Type:    ThemeTypeMac,
		Variant: ThemeVariantDark,
		Name:    "sequoia",
		Properties: ThemeProperties{
			TitleFont:          "Inter",
			TitleBackground:    color.RGBA{R: 36, G: 36, B: 36, A: 255},
			TitleText:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground:  color.RGBA{R: 28, G: 28, B: 28, A: 255},
			TextColor:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 122, B: 255, A: 255},
			BorderColor:        color.RGBA{R: 48, G: 48, B: 48, A: 255},
			InactiveTitleBg:    color.RGBA{R: 40, G: 40, B: 40, A: 255},
			InactiveTitleText:  color.RGBA{R: 128, G: 128, B: 128, A: 255},
			ButtonHoverColor:   color.RGBA{R: 48, G: 48, B: 48, A: 255},
			ButtonPressedColor: color.RGBA{R: 56, G: 56, B: 56, A: 255},
			CornerRadius:       macDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": MacStyleSequoia,
			},
		},
	}

	DefaultRegistry.RegisterTheme(ThemeTypeMac, "sequoia", ThemeVariantLight, lightTheme)
	DefaultRegistry.RegisterTheme(ThemeTypeMac, "sequoia", ThemeVariantDark, darkTheme)
}

func registerMacVenturaThemes() {
	lightTheme := Theme{
		Type:    ThemeTypeMac,
		Variant: ThemeVariantLight,
		Name:    "ventura",
		Properties: ThemeProperties{
			TitleFont:          "",
			TitleBackground:    color.RGBA{R: 236, G: 236, B: 236, A: 255},
			TitleText:          color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ControlsColor:      color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ContentBackground:  color.White,
			TextColor:          color.RGBA{R: 76, G: 76, B: 76, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 122, B: 255, A: 255},
			BorderColor:        color.RGBA{R: 200, G: 200, B: 200, A: 255},
			InactiveTitleBg:    color.RGBA{R: 246, G: 246, B: 246, A: 255},
			InactiveTitleText:  color.RGBA{R: 161, G: 161, B: 161, A: 255},
			ButtonHoverColor:   color.RGBA{R: 96, G: 96, B: 96, A: 255},
			ButtonPressedColor: color.RGBA{R: 56, G: 56, B: 56, A: 255},
			CornerRadius:       macDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": MacStyleVentura,
			},
		},
	}

	darkTheme := Theme{
		Type:    ThemeTypeMac,
		Variant: ThemeVariantDark,
		Name:    "ventura",
		Properties: ThemeProperties{
			TitleFont:          "",
			TitleBackground:    color.RGBA{R: 48, G: 48, B: 48, A: 255},
			TitleText:          color.RGBA{R: 236, G: 236, B: 236, A: 255},
			ControlsColor:      color.RGBA{R: 236, G: 236, B: 236, A: 255},
			ContentBackground:  color.RGBA{R: 28, G: 28, B: 28, A: 255},
			TextColor:          color.RGBA{R: 236, G: 236, B: 236, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 122, B: 255, A: 255},
			BorderColor:        color.RGBA{R: 60, G: 60, B: 60, A: 255},
			InactiveTitleBg:    color.RGBA{R: 38, G: 38, B: 38, A: 255},
			InactiveTitleText:  color.RGBA{R: 128, G: 128, B: 128, A: 255},
			ButtonHoverColor:   color.RGBA{R: 246, G: 246, B: 246, A: 255},
			ButtonPressedColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			CornerRadius:       macDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": MacStyleVentura,
			},
		},
	}

	DefaultRegistry.RegisterTheme(ThemeTypeMac, "ventura", ThemeVariantLight, lightTheme)
	DefaultRegistry.RegisterTheme(ThemeTypeMac, "ventura", ThemeVariantDark, darkTheme)
}

func registerMacBigSurThemes() {
	// TODO: Implement Big Sur themes with rounded corners and translucent materials
}

func registerMacCatalinaThemes() {
	// TODO: Implement Catalina themes with more traditional appearance
}

func registerMacLegacyThemes() {
	// TODO: Implement legacy themes (Snow Leopard, Lion, etc.)
}

// NewMacChrome creates a new macOS-style window chrome
func NewMacChrome(style MacStyle, opts ...ChromeOption) *MacChrome {
	chrome := &MacChrome{
		cornerRadius: macDefaultCornerRadius,
		title:        "Screenshot",
		titleBar:     true,
		themeName:    string(style),
		variant:      ThemeVariantLight,
		style:        style,
	}

	// Set initial theme
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeMac, string(style), ThemeVariantLight); ok {
		chrome.theme = theme
	}

	// Apply options
	for _, opt := range opts {
		chrome = opt(chrome).(*MacChrome)
	}

	return chrome
}

func (c *MacChrome) SetTheme(theme Theme) Chrome {
	c.theme = theme
	c.themeName = theme.Name
	c.variant = theme.Variant
	if style, ok := theme.Properties.CustomProperties["style"].(MacStyle); ok {
		c.style = style
	}
	return c
}

func (c *MacChrome) SetThemeByName(name string, variant ThemeVariant) Chrome {
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeMac, name, variant); ok {
		c.themeName = name
		c.variant = variant
		c.theme = theme
		if style, ok := theme.Properties.CustomProperties["style"].(MacStyle); ok {
			c.style = style
		}
	}
	return c
}

func (c *MacChrome) GetCurrentThemeName() string {
	return c.themeName
}

func (c *MacChrome) GetCurrentVariant() ThemeVariant {
	return c.variant
}

func (c *MacChrome) SetVariant(variant ThemeVariant) Chrome {
	return c.SetThemeByName(c.themeName, variant)
}

func (c *MacChrome) CurrentTheme() Theme {
	return c.theme
}

func (c *MacChrome) SetTitle(title string) Chrome {
	c.title = title
	return c
}

func (c *MacChrome) SetCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *MacChrome) SetTitleBar(enabled bool) Chrome {
	c.titleBar = enabled
	return c
}

// Render implements the Chrome interface
func (c *MacChrome) Render(content image.Image) (image.Image, error) {
	width := content.Bounds().Dx()
	height := content.Bounds().Dy()
	titleBarHeight := macDefaultTitleBarHeight

	if !c.titleBar {
		titleBarHeight = 0
	}

	// Create context for drawing
	dc := gg.NewContext(width, height+titleBarHeight)

	// Draw title bar background with corner radius
	if c.titleBar {
		dc.SetColor(c.theme.Properties.TitleBackground)
		if c.cornerRadius > 0 {
			dc.DrawRoundedRectangle(0, 0, float64(width), float64(height+titleBarHeight), c.cornerRadius)
		} else {
			dc.DrawRectangle(0, 0, float64(width), float64(height+titleBarHeight))
		}
		dc.Fill()

		// Draw window controls
		c.renderWindowControls(dc, titleBarHeight)

		// Draw title text if enabled
		if c.title != "" {
			DrawTitleText(dc, c.title, width, titleBarHeight, c.theme.Properties.TitleText, macDefaultTitleFontSize, c.theme.Properties.TitleFont)
		}
	}

	// Draw content
	dc.DrawImage(content, 0, titleBarHeight)

	return dc.Image(), nil
}

func (c *MacChrome) renderWindowControls(dc *gg.Context, titleBarHeight int) {
	switch c.style {
	case MacStyleSequoia, MacStyleSonoma, MacStyleVentura, MacStyleMonterey, MacStyleBigSur:
		c.renderModernControls(dc, titleBarHeight)
	case MacStyleCatalina, MacStyleMojave:
		c.renderFlatControls(dc, titleBarHeight)
	default:
		c.renderLegacyControls(dc, titleBarHeight)
	}
}

func (c *MacChrome) renderModernControls(dc *gg.Context, titleBarHeight int) {
	controlY := float64(titleBarHeight-macDefaultControlSize) / 2
	closeX := float64(macDefaultControlPadding)
	minimizeX := closeX + float64(macDefaultControlSize) + float64(macDefaultControlSpacing)
	maximizeX := minimizeX + float64(macDefaultControlSize) + float64(macDefaultControlSpacing)
	buttonSize := float64(macDefaultControlSize)

	// Close button (red)
	dc.SetColor(color.RGBA{R: 255, G: 95, B: 87, A: 255})
	dc.DrawCircle(closeX+buttonSize/2, controlY+buttonSize/2, buttonSize/2)
	dc.Fill()

	// Minimize button (yellow)
	dc.SetColor(color.RGBA{R: 255, G: 189, B: 46, A: 255})
	dc.DrawCircle(minimizeX+buttonSize/2, controlY+buttonSize/2, buttonSize/2)
	dc.Fill()

	// Maximize button (green)
	dc.SetColor(color.RGBA{R: 39, G: 201, B: 63, A: 255})
	dc.DrawCircle(maximizeX+buttonSize/2, controlY+buttonSize/2, buttonSize/2)
	dc.Fill()
}

func (c *MacChrome) renderFlatControls(dc *gg.Context, titleBarHeight int) {
	// TODO: Implement flat style controls (Catalina/Mojave)
}

func (c *MacChrome) renderLegacyControls(dc *gg.Context, titleBarHeight int) {
	// TODO: Implement legacy style controls (pre-Mojave)
}

func (c *MacChrome) MinimumSize() (width, height int) {
	return 100, macDefaultTitleBarHeight // Minimum size required for controls
}

func (c *MacChrome) ContentInsets() (top, right, bottom, left int) {
	return c.titleBarHeight(), 0, 0, 0
}

func (c *MacChrome) titleBarHeight() int {
	if c.titleBar {
		return macDefaultTitleBarHeight
	}
	return 0
}
