//go:generate go run gen_theme_map.go
package term

import (
	"embed"
	"image/color"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed themes/*.yml
var themeFS embed.FS

// Theme represents a terminal color theme
type Theme struct {
	Name    string `yaml:"name"`
	Author  string `yaml:"author"`
	Variant string `yaml:"variant"`

	// Standard colors (0-7)
	Color01 string `yaml:"color_01"` // Black
	Color02 string `yaml:"color_02"` // Red
	Color03 string `yaml:"color_03"` // Green
	Color04 string `yaml:"color_04"` // Yellow
	Color05 string `yaml:"color_05"` // Blue
	Color06 string `yaml:"color_06"` // Magenta
	Color07 string `yaml:"color_07"` // Cyan
	Color08 string `yaml:"color_08"` // White

	// Bright colors (8-15)
	Color09 string `yaml:"color_09"` // Bright Black
	Color10 string `yaml:"color_10"` // Bright Red
	Color11 string `yaml:"color_11"` // Bright Green
	Color12 string `yaml:"color_12"` // Bright Yellow
	Color13 string `yaml:"color_13"` // Bright Blue
	Color14 string `yaml:"color_14"` // Bright Magenta
	Color15 string `yaml:"color_15"` // Bright Cyan
	Color16 string `yaml:"color_16"` // Bright White

	// Special colors
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`
	Cursor     string `yaml:"cursor"`
}

var (
	themes    = make(map[string]*Theme)
	themeOnce sync.Once
)

// loadThemes loads all themes from the embedded filesystem
func loadThemes() {
	// Load theme on demand when requested
}

// normalizeThemeName converts a theme name to lowercase kebab case
func normalizeThemeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	// Remove any non-alphanumeric characters except hyphens
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return -1
	}, name)
	// Replace multiple consecutive hyphens with a single hyphen
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	// Trim hyphens from start and end
	return strings.Trim(name, "-")
}

// GetTheme returns a theme by name
func GetTheme(name string) *Theme {
	themeOnce.Do(loadThemes)
	name = normalizeThemeName(name)

	// Check if theme is already loaded
	if theme := themes[name]; theme != nil {
		return theme
	}

	// If not loaded, check if it exists in our file map
	if filename, ok := themeFileMap[name]; ok {
		// Load the theme from the file
		data, err := themeFS.ReadFile(filepath.Join("themes", filename))
		if err != nil {
			return nil
		}

		var theme Theme
		if err := yaml.Unmarshal(data, &theme); err != nil {
			return nil
		}

		// Cache the theme for future use
		themes[name] = &theme
		return &theme
	}

	return nil
}

// ListThemes returns a list of all available theme names
func ListThemes() []string {
	names := make([]string, 0, len(themeFileMap))
	for name := range themeFileMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// parseHexColor parses a hex color string into a color.Color
func parseHexColor(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 6 {
		r, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b, _ := strconv.ParseUint(hex[4:6], 16, 8)
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}
	return color.Black
}

// GetColor returns the color for the given index (0-15)
func (t *Theme) GetColor(index int) color.Color {
	switch index {
	case 0:
		return parseHexColor(t.Color01)
	case 1:
		return parseHexColor(t.Color02)
	case 2:
		return parseHexColor(t.Color03)
	case 3:
		return parseHexColor(t.Color04)
	case 4:
		return parseHexColor(t.Color05)
	case 5:
		return parseHexColor(t.Color06)
	case 6:
		return parseHexColor(t.Color07)
	case 7:
		return parseHexColor(t.Color08)
	case 8:
		return parseHexColor(t.Color09)
	case 9:
		return parseHexColor(t.Color10)
	case 10:
		return parseHexColor(t.Color11)
	case 11:
		return parseHexColor(t.Color12)
	case 12:
		return parseHexColor(t.Color13)
	case 13:
		return parseHexColor(t.Color14)
	case 14:
		return parseHexColor(t.Color15)
	case 15:
		return parseHexColor(t.Color16)
	default:
		return color.Black
	}
}

// GetBackground returns the theme's background color
func (t *Theme) GetBackground() color.Color {
	if t.Background != "" {
		return parseHexColor(t.Background)
	}
	return parseHexColor(t.Color01) // Fall back to black
}

// GetForeground returns the theme's foreground color
func (t *Theme) GetForeground() color.Color {
	if t.Foreground != "" {
		return parseHexColor(t.Foreground)
	}
	return parseHexColor(t.Color08) // Fall back to white
}

// GetCursor returns the theme's cursor color
func (t *Theme) GetCursor() color.Color {
	if t.Cursor != "" {
		return parseHexColor(t.Cursor)
	}
	return parseHexColor(t.Color08) // Fall back to white
}
