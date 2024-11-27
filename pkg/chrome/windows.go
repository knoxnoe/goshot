package chrome

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	winDefaultTitleBarHeight = 36
	winDefaultControlSize    = 25
	winDefaultControlSpacing = 2
	winDefaultCloseSpacing   = 0
	winDefaultTitleFontSize  = 14
	winDefaultControlPadding = 12
	winDefaultCornerRadius   = 8.0
	winDefaultButtonWidth    = 48
)

// WindowsStyle represents different Windows UI styles
type WindowsStyle string

const (
	WindowsStyleWin11 WindowsStyle = "windows11"
	WindowsStyleWin10 WindowsStyle = "windows10"
	WindowsStyleWin8  WindowsStyle = "windows8"
	WindowsStyleWinXP WindowsStyle = "windowsxp"
)

// WindowsChrome implements the Chrome interface with Windows-style window decorations
type WindowsChrome struct {
	theme        Theme
	cornerRadius float64
	title        string
	themeName    string
	variant      ThemeVariant
	titleBar     bool
	style        WindowsStyle
}

func init() {
	// Register Windows 11 themes
	registerWindows11Themes()
	// Register Windows 10 themes
	registerWindows10Themes()
	// Register Windows 8 themes
	registerWindows8Themes()
	// Register Windows XP themes
	registerWindowsXPThemes()
}

func registerWindows11Themes() {
	lightTheme := Theme{
		Type:    ThemeTypeWindows,
		Variant: ThemeVariantLight,
		Name:    "windows11",
		Properties: ThemeProperties{
			TitleFont:          "Inter",
			TitleBackground:    color.RGBA{R: 243, G: 243, B: 243, A: 255},
			TitleText:          color.RGBA{R: 0, G: 0, B: 0, A: 255},
			ControlsColor:      color.RGBA{R: 0, G: 0, B: 0, A: 255},
			ContentBackground:  color.White,
			TextColor:          color.RGBA{R: 0, G: 0, B: 0, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 120, B: 212, A: 255},
			BorderColor:        color.RGBA{R: 229, G: 229, B: 229, A: 255},
			InactiveTitleBg:    color.RGBA{R: 249, G: 249, B: 249, A: 255},
			InactiveTitleText:  color.RGBA{R: 128, G: 128, B: 128, A: 255},
			ButtonHoverColor:   color.RGBA{R: 229, G: 229, B: 229, A: 255},
			ButtonPressedColor: color.RGBA{R: 204, G: 204, B: 204, A: 255},
			CornerRadius:       winDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": WindowsStyleWin11,
			},
		},
	}

	darkTheme := Theme{
		Type:    ThemeTypeWindows,
		Variant: ThemeVariantDark,
		Name:    "windows11",
		Properties: ThemeProperties{
			TitleFont:          "Inter",
			TitleBackground:    color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TitleText:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground:  color.RGBA{R: 28, G: 28, B: 28, A: 255},
			TextColor:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			AccentColor:        color.RGBA{R: 0, G: 120, B: 212, A: 255},
			BorderColor:        color.RGBA{R: 51, G: 51, B: 51, A: 255},
			InactiveTitleBg:    color.RGBA{R: 38, G: 38, B: 38, A: 255},
			InactiveTitleText:  color.RGBA{R: 128, G: 128, B: 128, A: 255},
			ButtonHoverColor:   color.RGBA{R: 51, G: 51, B: 51, A: 255},
			ButtonPressedColor: color.RGBA{R: 68, G: 68, B: 68, A: 255},
			CornerRadius:       winDefaultCornerRadius,
			BorderWidth:        1.0,
			CustomProperties: map[string]any{
				"style": WindowsStyleWin11,
			},
		},
	}

	DefaultRegistry.RegisterTheme(ThemeTypeWindows, "windows11", ThemeVariantLight, lightTheme)
	DefaultRegistry.RegisterTheme(ThemeTypeWindows, "windows11", ThemeVariantDark, darkTheme)
}

func registerWindows10Themes() {
	// TODO: Implement Windows 10 themes
}

func registerWindows8Themes() {
	// TODO: Implement Windows 8 themes
}

func registerWindowsXPThemes() {
	// TODO: Implement Windows XP themes
}

// NewWindowsChrome creates a new Windows-style window chrome
func NewWindowsChrome(style WindowsStyle, opts ...ChromeOption) *WindowsChrome {
	chrome := &WindowsChrome{
		cornerRadius: winDefaultCornerRadius,
		title:        "Screenshot",
		titleBar:     true,
		themeName:    string(style),
		variant:      ThemeVariantLight,
		style:        style,
	}

	// Set initial theme
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeWindows, string(style), ThemeVariantLight); ok {
		chrome.theme = theme
	}

	// Apply options
	for _, opt := range opts {
		chrome = opt(chrome).(*WindowsChrome)
	}

	return chrome
}

func (c *WindowsChrome) WithTheme(theme Theme) Chrome {
	c.theme = theme
	c.themeName = theme.Name
	c.variant = theme.Variant
	if style, ok := theme.Properties.CustomProperties["style"].(WindowsStyle); ok {
		c.style = style
	}
	return c
}

func (c *WindowsChrome) WithThemeByName(name string, variant ThemeVariant) Chrome {
	if theme, ok := DefaultRegistry.GetTheme(ThemeTypeWindows, name, variant); ok {
		c.themeName = name
		c.variant = variant
		c.theme = theme
		if style, ok := theme.Properties.CustomProperties["style"].(WindowsStyle); ok {
			c.style = style
		}
	}
	return c
}

func (c *WindowsChrome) GetCurrentThemeName() string {
	return c.themeName
}

func (c *WindowsChrome) GetCurrentVariant() ThemeVariant {
	return c.variant
}

func (c *WindowsChrome) WithVariant(variant ThemeVariant) Chrome {
	return c.WithThemeByName(c.themeName, variant)
}

func (c *WindowsChrome) CurrentTheme() Theme {
	return c.theme
}

func (c *WindowsChrome) WithTitle(title string) Chrome {
	c.title = title
	return c
}

func (c *WindowsChrome) WithCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *WindowsChrome) WithTitleBar(enabled bool) Chrome {
	c.titleBar = enabled
	return c
}

// Render implements the Chrome interface
func (c *WindowsChrome) Render(content image.Image) (image.Image, error) {
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
		// Draw title text
		if c.title != "" {
			DrawTitleText(dc, c.title, width, titleBarHeight, c.theme.Properties.TitleText, winDefaultTitleFontSize, c.theme.Properties.TitleFont)
		}

		// Draw window controls based on style
		c.renderWindowControls(dc, width, titleBarHeight)
	}

	// Draw content
	dc.DrawImage(content, 0, titleBarHeight)

	return dc.Image(), nil
}

func (c *WindowsChrome) renderWindowControls(dc *gg.Context, width, titleBarHeight int) {
	switch c.style {
	case WindowsStyleWin11:
		c.renderWindows11Controls(dc, width, titleBarHeight)
	case WindowsStyleWin10:
		c.renderWindows10Controls(dc, width, titleBarHeight)
	case WindowsStyleWin8:
		c.renderWindows8Controls(dc, width, titleBarHeight)
	case WindowsStyleWinXP:
		c.renderWindowsXPControls(dc, width, titleBarHeight)
	}
}

func (c *WindowsChrome) renderWindows11Controls(dc *gg.Context, width, titleBarHeight int) {
	controlY := float64(titleBarHeight-winDefaultControlSize) / 2
	buttonWidth := float64(winDefaultButtonWidth)

	// Calculate button positions from right edge
	closeX := float64(width) - buttonWidth
	maximizeX := closeX - buttonWidth
	minimizeX := maximizeX - buttonWidth

	buttonSize := float64(winDefaultControlSize)
	iconSize := buttonSize * 0.45
	strokeWidth := 1.25

	dc.SetLineWidth(strokeWidth)

	// Close button (red background in hover state)
	if c.variant == ThemeVariantDark {
		dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	} else {
		dc.SetColor(c.theme.Properties.ControlsColor)
	}

	// Center the icons within their button areas
	closeIconX := closeX + (buttonWidth-buttonSize)/2
	maximizeIconX := maximizeX + (buttonWidth-buttonSize)/2
	minimizeIconX := minimizeX + (buttonWidth-buttonSize)/2

	// Close icon (X)
	closeOffset := buttonSize * 0.28
	dc.MoveTo(closeIconX+closeOffset, controlY+closeOffset)
	dc.LineTo(closeIconX+buttonSize-closeOffset, controlY+buttonSize-closeOffset)
	dc.MoveTo(closeIconX+closeOffset, controlY+buttonSize-closeOffset)
	dc.LineTo(closeIconX+buttonSize-closeOffset, controlY+closeOffset)
	dc.Stroke()

	// Maximize icon (rectangle)
	rectOffset := (buttonSize - iconSize) / 2
	dc.DrawRectangle(maximizeIconX+rectOffset, controlY+rectOffset, iconSize, iconSize)
	dc.Stroke()

	// Minimize icon (line)
	lineY := controlY + buttonSize*0.65
	dc.MoveTo(minimizeIconX+buttonSize*0.25, lineY)
	dc.LineTo(minimizeIconX+buttonSize*0.75, lineY)
	dc.Stroke()
}

func (c *WindowsChrome) renderWindows10Controls(dc *gg.Context, width, titleBarHeight int) {
	// TODO: Implement Windows 10 controls
}

func (c *WindowsChrome) renderWindows8Controls(dc *gg.Context, width, titleBarHeight int) {
	// TODO: Implement Windows 8 controls
}

func (c *WindowsChrome) renderWindowsXPControls(dc *gg.Context, width, titleBarHeight int) {
	// TODO: Implement Windows XP controls
}

func (c *WindowsChrome) MinimumSize() (width, height int) {
	return 100, winDefaultTitleBarHeight // Minimum size required for controls
}

func (c *WindowsChrome) ContentInsets() (top, right, bottom, left int) {
	return c.titleBarHeight(), 0, 0, 0
}

func (c *WindowsChrome) titleBarHeight() int {
	if c.titleBar {
		return winDefaultTitleBarHeight
	}
	return 0
}
