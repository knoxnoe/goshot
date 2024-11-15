// Package fonts provides font loading functionality for goshot
package fonts

import (
	"io/fs"
)

// FontStyle represents the style of a font
type FontStyle struct {
	Weight    FontWeight // Font weight
	Italic    bool       // Whether the font is italic
	Condensed bool       // Whether the font is condensed
	Mono      bool       // Whether the font is monospaced
}

// FontWeight represents the weight of a font
type FontWeight int

// Font weight constants represent standard font weights from thin to heavy
const (
	WeightThin       FontWeight = iota + 1 // Thinnest font weight
	WeightExtraLight                       // Extra light font weight
	WeightLight                            // Light font weight
	WeightRegular                          // Regular (normal) font weight
	WeightMedium                           // Medium font weight
	WeightSemiBold                         // Semi-bold font weight
	WeightBold                             // Bold font weight
	WeightExtraBold                        // Extra bold font weight
	WeightBlack                            // Black font weight
	WeightHeavy                            // Heaviest font weight
)

// Font represents a loaded font with its metadata
type Font struct {
	Name     string    // Name of the font family
	Data     []byte    // Raw font data
	Style    FontStyle // Style information for this font variant
	FilePath string    // Path to the font file on disk
}

// GetFont returns a font by name and style.
// When style is nil, returns the regular style if available.
func GetFont(name string, style *FontStyle) (*Font, error) {
	return getFont(name, style)
}

// GetFontVariants returns all available variants of a font
func GetFontVariants(name string) ([]*Font, error) {
	return getFontVariants(name)
}

// ListFonts returns a list of available font names.
// The implementation differs based on build tags.
func ListFonts() []string {
	return listFonts()
}

// FontFS returns an fs.FS interface for accessing fonts.
// When built with -tags bundled, it returns an embedded fs.
// Otherwise, it returns a directory-based fs pointing to system fonts.
func FontFS() fs.FS {
	return fontFS()
}
