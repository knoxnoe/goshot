package chrome

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

const (
	win11TitleBarHeight  = 40
	win11ControlIconSize = 14
	win11ControlSpacing  = 18
	win11CloseSpacing    = 20
	win11TitleFontSize   = 13
	win11ControlPadding  = 12
	win11CornerRadius    = 8.0
)

// Windows11Chrome implements the Chrome interface with Windows 11-style window decorations
type Windows11Chrome struct {
	title        string
	darkMode     bool
	cornerRadius float64
	theme        Theme
	darkTheme    Theme
	showTitleBar bool
}

// NewWindows11Chrome creates a new Windows 11-style window chrome
func NewWindows11Chrome(opts ...ChromeOption) *Windows11Chrome {
	chrome := &Windows11Chrome{
		theme: Theme{
			TitleFont:         "Segoe UI",
			TitleBackground:   color.RGBA{R: 241, G: 244, B: 249, A: 255},
			TitleText:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			ControlsColor:     color.RGBA{R: 32, G: 32, B: 32, A: 255},
			ContentBackground: color.RGBA{R: 25, G: 25, B: 25, A: 255},
			TextColor:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			DarkTextColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		darkTheme: Theme{
			TitleFont:         "Segoe UI",
			TitleBackground:   color.RGBA{R: 32, G: 32, B: 32, A: 255},
			TitleText:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ControlsColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
			ContentBackground: color.RGBA{R: 25, G: 25, B: 25, A: 255},
			TextColor:         color.RGBA{R: 0, G: 0, B: 0, A: 255},
			DarkTextColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		cornerRadius: win11CornerRadius,
		showTitleBar: true,
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

func (c *Windows11Chrome) SetTitleBar(enabled bool) Chrome {
	c.showTitleBar = enabled
	return c
}

// Render implements the Chrome interface
func (w *Windows11Chrome) Render(content image.Image) (image.Image, error) {
	width := content.Bounds().Dx()
	height := content.Bounds().Dy()
	titleBarHeight := win11TitleBarHeight

	if !w.showTitleBar {
		titleBarHeight = 0
	}

	// Create context for drawing
	dc := gg.NewContext(width, height+titleBarHeight)

	// Get the current theme
	theme := w.theme
	if w.darkMode {
		theme = w.darkTheme
	}

	// Draw window base (background and rounded corners)
	err := DrawWindowBase(dc, width, height+titleBarHeight, w.cornerRadius, theme.TitleBackground, theme.ContentBackground, titleBarHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to draw window base: %v", err)
	}

	// Draw content
	dc.DrawImage(content, 0, titleBarHeight)

	if w.showTitleBar {
		// Draw window controls (Windows 11 style)
		controlY := float64(titleBarHeight-win11ControlIconSize) / 2 // Center vertically

		// Close button (rightmost)
		closeX := float64(width) - win11ControlIconSize - win11ControlPadding
		DrawCross(dc, closeX, controlY, float64(win11ControlIconSize), theme.ControlsColor)

		// Maximize button (more space before close)
		maximizeX := closeX - win11ControlIconSize - float64(win11CloseSpacing)
		DrawSquare(dc, maximizeX, controlY, float64(win11ControlIconSize), theme.ControlsColor)

		// Minimize button (less space before maximize)
		minimizeX := maximizeX - win11ControlIconSize - float64(win11ControlSpacing)
		DrawLine(dc, minimizeX, controlY+float64(win11ControlIconSize)/2, float64(win11ControlIconSize), theme.ControlsColor)

		// Draw title text if provided (centered)
		if w.title != "" {
			textColor := theme.TextColor
			if w.darkMode {
				textColor = theme.DarkTextColor
			}
			err := DrawTitleText(dc, w.title, width, titleBarHeight, textColor, win11TitleFontSize, theme.TitleFont)
			if err != nil {
				return nil, fmt.Errorf("failed to draw title text: %v", err)
			}
		}
	}

	return dc.Image(), nil
}

// DefaultTheme implements the Chrome interface
func (w *Windows11Chrome) DefaultTheme() Theme {
	return w.theme
}

// DarkTheme implements the Chrome interface
func (w *Windows11Chrome) DarkTheme() Theme {
	return w.darkTheme
}

// MinimumSize implements the Chrome interface
func (w *Windows11Chrome) MinimumSize() (width, height int) {
	return 3*win11ControlIconSize + 100, // Minimum width to accommodate controls and title
		w.titleBarHeight()
}

// ContentInsets implements the Chrome interface
func (w *Windows11Chrome) ContentInsets() (top, right, bottom, left int) {
	return w.titleBarHeight(), 0, 0, 0
}

// titleBarHeight returns the actual title bar height based on whether it's shown
func (w *Windows11Chrome) titleBarHeight() int {
	if w.showTitleBar {
		return win11TitleBarHeight
	}
	return 0
}
