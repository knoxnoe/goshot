package chrome

import (
	"image"
	"image/color"
)

// Theme represents a color theme for window chrome
type Theme struct {
	// Font name for the title
	TitleFont string
	// Background color of the title bar
	TitleBackground color.Color
	// Text color for the title
	TitleText color.Color
	// Color of window controls (close, minimize, maximize)
	ControlsColor color.Color
	// Background color for the content area
	ContentBackground color.Color
	// Text color for the content area
	TextColor color.Color
	// Dark text color for the content area
	DarkTextColor color.Color
}

// Chrome represents a window chrome renderer
type Chrome interface {
	// Render renders the window chrome around the given content
	// Returns the final image and any error that occurred
	Render(content image.Image) (image.Image, error)

	// MinimumSize returns the minimum size required for the chrome
	MinimumSize() (width, height int)

	// ContentInsets returns the insets for the content area
	// This is used to properly position the content within the chrome
	ContentInsets() (top, right, bottom, left int)

	// SetTheme sets the theme for the window chrome
	SetTheme(theme Theme) Chrome

	// SetDarkTheme sets the dark theme for the window chrome
	SetDarkTheme(theme Theme) Chrome

	// SetTitle sets the window title
	SetTitle(title string) Chrome

	// SetCornerRadius sets the corner radius for the window
	SetCornerRadius(radius float64) Chrome

	// SetDarkMode enables or disables dark mode
	SetDarkMode(darkMode bool) Chrome

	// SetTitleBar enables or disables the title bar
	SetTitleBar(enabled bool) Chrome

	// DefaultTheme returns the default light theme for the chrome implementation
	DefaultTheme() Theme

	// DarkTheme returns the default dark theme for the chrome implementation
	DarkTheme() Theme
}

// ChromeOption is a function that configures a Chrome implementation
type ChromeOption func(Chrome) Chrome

// WithTitle sets the window title
func WithTitle(title string) ChromeOption {
	return func(c Chrome) Chrome {
		if w, ok := c.(interface{ SetTitle(string) Chrome }); ok {
			return w.SetTitle(title)
		}
		return c
	}
}

// WithDarkMode enables dark mode for the window chrome
func WithDarkMode(enabled bool) ChromeOption {
	return func(c Chrome) Chrome {
		if w, ok := c.(interface{ SetDarkMode(bool) Chrome }); ok {
			return w.SetDarkMode(enabled)
		}
		return c
	}
}

// WithTitleBar enables or disables the title bar
func WithTitleBar(enabled bool) ChromeOption {
	return func(c Chrome) Chrome {
		if w, ok := c.(interface{ SetTitleBar(bool) Chrome }); ok {
			return w.SetTitleBar(enabled)
		}
		return c
	}
}

// WithCornerRadius sets the corner radius for the window
func WithCornerRadius(radius float64) ChromeOption {
	return func(c Chrome) Chrome {
		if w, ok := c.(interface{ SetCornerRadius(float64) Chrome }); ok {
			return w.SetCornerRadius(radius)
		}
		return c
	}
}
