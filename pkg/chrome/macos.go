package chrome

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	macOSTitleBarHeight = 28
	macOSControlSize    = 12
	macOSControlSpacing = 8
	macOSLeftPadding    = 10
	macOSRightPadding   = 8
	macOSTitleFontSize  = 13
	defaultCornerRadius = 9
)

// MacOSChrome implements the Chrome interface with macOS-style window decorations
type MacOSChrome struct {
	theme        Theme
	darkTheme    Theme
	cornerRadius float64
	title        string
	darkMode     bool
	titleBar     bool
}

// NewMacOSChrome creates a new macOS-style window chrome
func NewMacOSChrome(opts ...ChromeOption) *MacOSChrome {
	chrome := &MacOSChrome{
		theme: Theme{
			TitleFont:         "SF Pro",
			TitleBackground:   color.RGBA{R: 236, G: 236, B: 236, A: 255},
			TitleText:         color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ControlsColor:     color.RGBA{R: 76, G: 76, B: 76, A: 255},
			ContentBackground: color.White,
			TextColor:         color.RGBA{R: 76, G: 76, B: 76, A: 255},
			DarkTextColor:     color.RGBA{R: 236, G: 236, B: 236, A: 255},
		},
		darkTheme: Theme{
			TitleFont:         "SF Pro",
			TitleBackground:   color.RGBA{R: 48, G: 48, B: 48, A: 255},
			TitleText:         color.RGBA{R: 236, G: 236, B: 236, A: 255},
			ControlsColor:     color.RGBA{R: 236, G: 236, B: 236, A: 255},
			ContentBackground: color.RGBA{R: 28, G: 28, B: 28, A: 255},
			TextColor:         color.RGBA{R: 76, G: 76, B: 76, A: 255},
			DarkTextColor:     color.RGBA{R: 236, G: 236, B: 236, A: 255},
		},
		cornerRadius: defaultCornerRadius,
		title:        "Screenshot",
		darkMode:     false,
		titleBar:     true,
	}

	for _, opt := range opts {
		chrome = opt(chrome).(*MacOSChrome)
	}

	return chrome
}

func (c *MacOSChrome) SetTheme(theme Theme) Chrome {
	c.theme = theme
	return c
}

func (c *MacOSChrome) SetDarkTheme(theme Theme) Chrome {
	c.darkTheme = theme
	return c
}

func (c *MacOSChrome) SetTitle(title string) Chrome {
	c.title = title
	return c
}

func (c *MacOSChrome) SetCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *MacOSChrome) SetDarkMode(darkMode bool) Chrome {
	c.darkMode = darkMode
	return c
}

func (c *MacOSChrome) SetTitleBar(show bool) Chrome {
	c.titleBar = show
	return c
}

// Render implements the Chrome interface
func (c *MacOSChrome) Render(content image.Image) (image.Image, error) {
	if !c.titleBar {
		return content, nil
	}

	theme := c.theme
	if c.darkMode {
		theme = c.darkTheme
	}

	bounds := content.Bounds()
	width := bounds.Dx()
	height := bounds.Dy() + macOSTitleBarHeight

	dc := gg.NewContext(width, height)

	// Draw window base with rounded corners
	if err := DrawWindowBase(dc, width, height, c.cornerRadius, theme.TitleBackground, theme.ContentBackground, macOSTitleBarHeight); err != nil {
		return nil, fmt.Errorf("failed to draw window base: %v", err)
	}

	// Draw content
	dc.DrawImage(content, 0, macOSTitleBarHeight)

	// Draw window controls (macOS style)
	controlY := float64(macOSTitleBarHeight-macOSControlSize) / 2

	// Close button (leftmost)
	closeX := float64(macOSLeftPadding)
	DrawMacOSWindowControl(dc, closeX, controlY, float64(macOSControlSize), color.RGBA{R: 255, G: 95, B: 87, A: 255})

	// Minimize button
	minimizeX := closeX + float64(macOSControlSize) + float64(macOSControlSpacing)
	DrawMacOSWindowControl(dc, minimizeX, controlY, float64(macOSControlSize), color.RGBA{R: 255, G: 189, B: 46, A: 255})

	// Maximize button
	maximizeX := minimizeX + float64(macOSControlSize) + float64(macOSControlSpacing)
	DrawMacOSWindowControl(dc, maximizeX, controlY, float64(macOSControlSize), color.RGBA{R: 39, G: 201, B: 63, A: 255})

	// Draw title text if provided (centered)
	if c.title != "" {
		textColor := c.theme.TextColor
		if c.darkMode {
			textColor = c.theme.DarkTextColor
		}
		if err := DrawTitleText(dc, c.title, int(width), int(macOSTitleBarHeight), textColor, macOSTitleFontSize, c.theme.TitleFont); err != nil {
			return nil, err
		}
	}

	return dc.Image(), nil
}

func (m *MacOSChrome) DefaultTheme() Theme {
	return m.theme
}

func (m *MacOSChrome) DarkTheme() Theme {
	return m.darkTheme
}

// MinimumSize implements the Chrome interface
func (m *MacOSChrome) MinimumSize() (width, height int) {
	return macOSLeftPadding + 3*macOSControlSize + 2*macOSControlSpacing + macOSRightPadding + 100,
		macOSTitleBarHeight
}

// ContentInsets implements the Chrome interface
func (m *MacOSChrome) ContentInsets() (top, right, bottom, left int) {
	return macOSTitleBarHeight, macOSRightPadding, 0, macOSLeftPadding
}

// DrawMacOSWindowControl draws a circular window control button
func DrawMacOSWindowControl(dc *gg.Context, x, y, size float64, color color.Color) {
	dc.SetColor(color)
	dc.DrawCircle(x+size/2, y+size/2, size/2)
	dc.Fill()
}
