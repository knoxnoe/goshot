package chrome

import (
	"image"
	"image/color"
)

// ThemeType represents the type of window chrome theme
type ThemeType string

const (
	ThemeTypeWindows ThemeType = "windows"
	ThemeTypeMac     ThemeType = "mac"
	ThemeTypeGNOME   ThemeType = "gnome"
)

// ThemeVariant represents a variant of a theme (e.g., light or dark)
type ThemeVariant string

const (
	ThemeVariantLight ThemeVariant = "light"
	ThemeVariantDark  ThemeVariant = "dark"
)

// ThemeProperties contains the visual properties of a theme
type ThemeProperties struct {
	// Basic properties
	TitleFont         string
	TitleText         color.Color
	TitleFontSize     float64
	TitleBackground   color.Color
	ControlsColor     color.Color
	ContentBackground color.Color
	TextColor         color.Color

	// Extended properties for different window managers
	AccentColor        color.Color    // Primary accent color
	BorderColor        color.Color    // Window border color
	InactiveTitleBg    color.Color    // Title bar background when window is inactive
	InactiveTitleText  color.Color    // Title text color when window is inactive
	ButtonHoverColor   color.Color    // Button hover state color
	ButtonPressedColor color.Color    // Button pressed state color
	CornerRadius       float64        // Window corner radius
	BorderWidth        float64        // Window border width
	CustomProperties   map[string]any // Additional theme-specific properties
}

// Theme represents a complete window chrome theme
type Theme struct {
	Type       ThemeType
	Variant    ThemeVariant
	Name       string
	Properties ThemeProperties
}

// Chrome defines the interface for window chrome implementations
type Chrome interface {
	SetTitle(title string) Chrome
	SetCornerRadius(radius float64) Chrome
	SetTitleBar(enabled bool) Chrome
	SetTheme(theme Theme) Chrome
	SetThemeByName(name string, variant ThemeVariant) Chrome
	SetVariant(variant ThemeVariant) Chrome
	GetCurrentThemeName() string
	GetCurrentVariant() ThemeVariant
	CurrentTheme() Theme
	Render(content image.Image) (image.Image, error)
	MinimumSize() (width, height int)
	ContentInsets() (top, right, bottom, left int)
}

// ChromeOption is a function that modifies a Chrome instance
type ChromeOption func(Chrome) Chrome

// WithTitle sets the window title
func WithTitle(title string) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetTitle(title)
	}
}

// WithCornerRadius sets the corner radius
func WithCornerRadius(radius float64) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetCornerRadius(radius)
	}
}

// WithTitleBar enables or disables the title bar
func WithTitleBar(enabled bool) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetTitleBar(enabled)
	}
}

// WithTheme sets the theme
func WithTheme(theme Theme) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetTheme(theme)
	}
}

// WithThemeByName sets the theme by name and variant
func WithThemeByName(name string, variant ThemeVariant) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetThemeByName(name, variant)
	}
}

// WithVariant sets the theme variant
func WithVariant(variant ThemeVariant) ChromeOption {
	return func(c Chrome) Chrome {
		return c.SetVariant(variant)
	}
}

// ThemeRegistry maintains a registry of available themes
type ThemeRegistry struct {
	themes map[ThemeType]map[string]map[ThemeVariant]Theme
}

// NewThemeRegistry creates a new theme registry
func NewThemeRegistry() *ThemeRegistry {
	return &ThemeRegistry{
		themes: make(map[ThemeType]map[string]map[ThemeVariant]Theme),
	}
}

// RegisterTheme registers a theme with the registry
func (r *ThemeRegistry) RegisterTheme(themeType ThemeType, name string, variant ThemeVariant, theme Theme) {
	if r.themes[themeType] == nil {
		r.themes[themeType] = make(map[string]map[ThemeVariant]Theme)
	}
	if r.themes[themeType][name] == nil {
		r.themes[themeType][name] = make(map[ThemeVariant]Theme)
	}
	r.themes[themeType][name][variant] = theme
}

// GetTheme retrieves a theme from the registry
func (r *ThemeRegistry) GetTheme(themeType ThemeType, name string, variant ThemeVariant) (Theme, bool) {
	if variants, ok := r.themes[themeType][name]; ok {
		if theme, ok := variants[variant]; ok {
			return theme, true
		}
	}
	return Theme{}, false
}

// GetThemeNames returns all registered theme names for a given type
func (r *ThemeRegistry) GetThemeNames(themeType ThemeType) []string {
	if themes, ok := r.themes[themeType]; ok {
		names := make([]string, 0, len(themes))
		for name := range themes {
			names = append(names, name)
		}
		return names
	}
	return nil
}

// DefaultRegistry is the default theme registry
var DefaultRegistry = NewThemeRegistry()
