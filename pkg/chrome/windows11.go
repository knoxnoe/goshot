package chrome

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	win11TitleBarHeight  = 40
	win11ControlSize     = 52
	win11ControlIconSize = 16
	win11ControlSpacing  = 2
	win11CloseSpacing    = 4
	win11TitleFontSize   = 13
	win11ControlPadding  = 0
	win11CornerRadius    = 8.0
)

// Windows11Chrome implements the Chrome interface with Windows 11-style window decorations
type Windows11Chrome struct {
	theme        Theme
	darkTheme    Theme
	cornerRadius float64
	title        string
	darkMode     bool
}

// NewWindows11Chrome creates a new Windows 11-style window chrome
func NewWindows11Chrome(opts ...ChromeOption) *Windows11Chrome {
	chrome := &Windows11Chrome{
		theme: Theme{
			TitleBackground:   color.White,
			TitleText:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			ControlsColor:     color.RGBA{R: 32, G: 32, B: 32, A: 255},
			ContentBackground: color.White,
			TextColor:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			DarkTextColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		darkTheme: Theme{
			TitleBackground:   color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TitleText:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground: color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TextColor:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			DarkTextColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		cornerRadius: win11CornerRadius,
	}

	for _, opt := range opts {
		chrome = opt(chrome).(*Windows11Chrome)
	}

	return chrome
}

func (c *Windows11Chrome) SetTheme(theme Theme) Chrome {
	c.theme = theme
	return c
}

func (c *Windows11Chrome) SetDarkTheme(theme Theme) Chrome {
	c.darkTheme = theme
	return c
}

func (c *Windows11Chrome) SetTitle(title string) Chrome {
	c.title = title
	return c
}

func (c *Windows11Chrome) SetCornerRadius(radius float64) Chrome {
	c.cornerRadius = radius
	return c
}

func (c *Windows11Chrome) SetDarkMode(darkMode bool) Chrome {
	c.darkMode = darkMode
	return c
}

// Render implements the Chrome interface
func (c *Windows11Chrome) Render(content image.Image) (image.Image, error) {
	theme := c.theme
	if c.darkMode {
		theme = c.darkTheme
	}

	bounds := content.Bounds()
	width := bounds.Dx()
	height := bounds.Dy() + win11TitleBarHeight

	dc := gg.NewContext(width, height)

	// Draw window base with rounded corners
	if err := DrawWindowBase(dc, width, height, c.cornerRadius, theme.TitleBackground, theme.ContentBackground, win11TitleBarHeight); err != nil {
		return nil, fmt.Errorf("failed to draw window base: %v", err)
	}

	// Draw content
	dc.DrawImage(content, 0, win11TitleBarHeight)

	// Draw window controls (Windows 11 style)
	controlY := float64(win11TitleBarHeight-win11ControlIconSize)/2 - 2 // Center vertically with slight upward adjustment

	// Close button (rightmost)
	closeX := float64(width) - win11ControlSize
	DrawWindowControl(dc, closeX, 0, float64(win11ControlSize), color.Transparent)
	DrawCross(dc, closeX+float64(win11ControlSize-win11ControlIconSize)/2, controlY, float64(win11ControlIconSize), theme.ControlsColor)

	// Maximize button (more space before close)
	maximizeX := closeX - win11ControlSize - float64(win11CloseSpacing)
	DrawWindowControl(dc, maximizeX, 0, float64(win11ControlSize), color.Transparent)
	DrawSquare(dc, maximizeX+float64(win11ControlSize-win11ControlIconSize)/2, controlY, float64(win11ControlIconSize), theme.ControlsColor)

	// Minimize button (less space before maximize)
	minimizeX := maximizeX - win11ControlSize - float64(win11ControlSpacing)
	DrawWindowControl(dc, minimizeX, 0, float64(win11ControlSize), color.Transparent)
	DrawLine(dc, minimizeX+float64(win11ControlSize-win11ControlIconSize)/2, controlY+float64(win11ControlIconSize)/2, float64(win11ControlIconSize), theme.ControlsColor)

	// Draw title text if provided (centered)
	if c.title != "" {
		textColor := c.theme.TextColor
		if c.darkMode {
			textColor = c.theme.DarkTextColor
		}
		if err := DrawTitleText(dc, c.title, width, win11TitleBarHeight, textColor, 14); err != nil {
			return nil, err
		}
	}

	return dc.Image(), nil
}

func (w *Windows11Chrome) DefaultTheme() Theme {
	return w.theme
}

func (w *Windows11Chrome) DarkTheme() Theme {
	return w.darkTheme
}

// MinimumSize implements the Chrome interface
func (w *Windows11Chrome) MinimumSize() (width, height int) {
	return 3*win11ControlSize + 100, // Minimum width to accommodate controls and title
		win11TitleBarHeight
}

// ContentInsets implements the Chrome interface
func (w *Windows11Chrome) ContentInsets() (top, right, bottom, left int) {
	return win11TitleBarHeight, 0, 0, 0
}
