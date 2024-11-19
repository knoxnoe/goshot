package chrome

import (
	"image"
	"image/color"
)

// Theme represents a color theme for window chrome
type Theme struct {
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

	// DefaultTheme returns the default light theme for the chrome implementation
	DefaultTheme() Theme

	// DarkTheme returns the default dark theme for the chrome implementation
	DarkTheme() Theme
}

// ChromeOption is a function that configures a Chrome implementation
type ChromeOption func(Chrome) Chrome
