package fonts

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed embedded
var embeddedFonts embed.FS

// fontCache caches loaded fonts to avoid repeated disk access
var (
	fontCache      = make(map[string][]*Font)
	fontCacheMu    sync.RWMutex
	variantCache   = make(map[string]*Font)
	variantCacheMu sync.RWMutex
)

// FontStyle represents the style of a font
type FontStyle struct {
	Weight     FontWeight         // Font weight
	Italic     bool               // Whether the font is italic
	Condensed  bool               // Whether the font is condensed
	Mono       bool               // Whether the font is monospaced
	Variations map[string]float32 // Variable font variations
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

// Common variable font axis tags
const (
	AxisWeight      = "wght" // Weight axis
	AxisWidth       = "wdth" // Width axis
	AxisSlant       = "slnt" // Slant axis
	AxisItalic      = "ital" // Italic axis
	AxisOpticalSize = "opsz" // Optical size axis
	AxisTexture     = "TXTR" // Texture healing
	AxisLigatures   = "liga" // Ligatures
)

// Monaspace specific ranges
const (
	MonaspaceWeightMin = 200
	MonaspaceWeightMax = 800
	MonaspaceWidthMin  = 100
	MonaspaceWidthMax  = 125
	MonaspaceSlantMin  = -11
	MonaspaceSlantMax  = 1
)

// Font represents a loaded font with its metadata
type Font struct {
	Name     string         // Name of the font family
	Font     *opentype.Font // Parsed font
	Style    FontStyle      // Style information for this font variant
	FilePath string         // Path to the font file on disk
	Filename string         // Name of the font file
}

// ToTrueType converts the opentype.Font to a truetype.Font
func (f *Font) ToTrueType() (*truetype.Font, error) {
	if f == nil || f.Font == nil {
		return nil, fmt.Errorf("invalid font")
	}

	// Get the raw font data
	var data []byte
	var err error

	if f.FilePath != "" {
		data, err = os.ReadFile(f.FilePath)
	} else if f.Filename != "" {
		data, err = embeddedFonts.ReadFile("embedded/" + f.Filename)
	} else {
		return nil, fmt.Errorf("no font data available")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read font data: %v", err)
	}

	// Parse as truetype font
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse truetype font: %v", err)
	}

	return ttf, nil
}

// systemFontPaths contains the default system font directories for different operating systems
var systemFontPaths = map[string][]string{
	"linux": {
		"~/.fonts",
		"~/.local/share/fonts",
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	},
	"darwin": {
		"/System/Library/Fonts",
		"/Library/Fonts",
		"~/Library/Fonts",
	},
	"windows": {
		"C:\\Windows\\Fonts",
	},
}

type systemFontFS struct {
	root string
}

func (sfs systemFontFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(sfs.root, name))
}

func GetFont(name string, style *FontStyle) (*Font, error) {
	if name == "" {
		return nil, fmt.Errorf("font name cannot be empty")
	}

	cacheKey := name
	if style != nil {
		cacheKey = fmt.Sprintf("%s-%d-%v-%v-%v", name, style.Weight, style.Italic, style.Condensed, style.Mono)
	}

	// Check variant cache first
	variantCacheMu.RLock()
	if cached, ok := variantCache[cacheKey]; ok {
		variantCacheMu.RUnlock()
		return cached, nil
	}
	variantCacheMu.RUnlock()

	// Try to load the requested font
	variants, err := GetFontVariants(name)
	if err != nil || len(variants) == 0 {
		return nil, fmt.Errorf("font %s not found", name)
	}

	var selected *Font

	if style == nil {
		// Find regular variant
		for _, font := range variants {
			if font.Style.Weight == WeightRegular && !font.Style.Italic {
				selected = font
				break
			}
		}
		if selected == nil {
			selected = variants[0]
		}
	} else {
		// Find best matching style
		var bestScore int
		for _, font := range variants {
			score := matchStyleScore(font.Style, *style)
			if score > bestScore {
				selected = font
				bestScore = score
			}
		}
	}

	if selected != nil {
		// Cache the result
		variantCacheMu.Lock()
		variantCache[cacheKey] = selected
		variantCacheMu.Unlock()
		return selected, nil
	}

	return nil, fmt.Errorf("no suitable variant found for font %s", name)
}

// FallbackVariant represents the type of fallback font to use
type FallbackVariant string

const (
	FallbackSans FallbackVariant = "sans"
	FallbackMono FallbackVariant = "mono"
)

// GetFallback returns either Monaspace Argon or Neon as the fallback font
func GetFallback(variant FallbackVariant) (*Font, error) {
	var filename string
	var fontName string
	var variations map[string]float32

	switch variant {
	case FallbackMono:
		filename = "MonaspaceNeon-Regular.ttf"
		fontName = "Monaspace Neon"
		variations = map[string]float32{
			AxisWeight:    400, // Regular weight
			AxisWidth:     100, // Normal width
			AxisSlant:     0,   // No slant
			AxisTexture:   0,   // No texture healing
			AxisLigatures: 0,   // Disable ligatures
		}
	default: // FallbackSans
		filename = "MonaspaceArgon-Regular.ttf"
		fontName = "Monaspace Argon"
		variations = map[string]float32{
			AxisWeight:    400, // Regular weight
			AxisWidth:     100, // Normal width
			AxisSlant:     0,   // No slant
			AxisTexture:   0,   // No texture healing
			AxisLigatures: 0,   // Disable ligatures
		}
	}

	data, err := embeddedFonts.ReadFile("embedded/" + filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", fontName, err)
	}

	font, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %v", fontName, err)
	}

	return &Font{
		Name: fontName,
		Font: font,
		Style: FontStyle{
			Weight:     WeightRegular,
			Mono:       variant == FallbackMono,
			Variations: variations,
		},
		Filename: filename,
	}, nil
}

// IsFontAvailable is a super fast check to see if the given font is available on the system
func IsFontAvailable(name string) bool {
	_, err := GetFont(name, nil)
	return err == nil
}

// matchStyleScore returns a score indicating how well two font styles match
// Higher score means better match
func matchStyleScore(a, b FontStyle) int {
	score := 0

	// Weight match (closer weights = higher score)
	weightDiff := abs(int(a.Weight) - int(b.Weight))
	score += 100 - (weightDiff * 10)

	// Exact style matches
	if a.Italic == b.Italic {
		score += 50
	}
	if a.Condensed == b.Condensed {
		score += 50
	}
	if a.Mono == b.Mono {
		score += 50
	}

	return score
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GetFontVariants(name string) ([]*Font, error) {
	fontCacheMu.RLock()
	if cached, ok := fontCache[name]; ok {
		fontCacheMu.RUnlock()
		if len(cached) == 0 {
			return nil, fmt.Errorf("font %s not found", name)
		}
		return cached, nil
	}
	fontCacheMu.RUnlock()

	var variants []*Font

	// Search embedded fonts first
	entries, err := embeddedFonts.ReadDir("embedded")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			cleanName := cleanFontName(entry.Name())
			if strings.EqualFold(cleanName, name) {
				data, err := embeddedFonts.ReadFile(filepath.Join("embedded", entry.Name()))
				if err != nil {
					continue
				}

				font, err := opentype.Parse(data)
				if err != nil {
					continue
				}

				fontStyle := extractFontStyle(entry.Name())
				variant := &Font{
					Name:     cleanName,
					Font:     font,
					Style:    fontStyle,
					Filename: entry.Name(),
				}
				variants = append(variants, variant)
			}
		}
	}

	// Search system fonts
	osType := runtime.GOOS
	paths, ok := systemFontPaths[osType]
	if !ok {
		return nil, fmt.Errorf("unsupported OS: %s", osType)
	}

	for _, basePath := range paths {
		// Expand home directory if needed
		if strings.HasPrefix(basePath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			basePath = filepath.Join(home, basePath[2:])
		}

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip this file/directory
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".otf" {
				return nil
			}

			cleanName := cleanFontName(info.Name())
			if strings.EqualFold(cleanName, name) {
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				font, err := opentype.Parse(data)
				if err != nil {
					return nil
				}

				fontStyle := extractFontStyle(info.Name())
				variant := &Font{
					Name:     cleanName,
					Font:     font,
					Style:    fontStyle,
					FilePath: path,
				}
				variants = append(variants, variant)
			}

			return nil
		})

		if err != nil {
			log.Printf("Error walking font directory %s: %v", basePath, err)
		}
	}

	// Cache the results
	fontCacheMu.Lock()
	fontCache[name] = variants
	fontCacheMu.Unlock()

	if len(variants) == 0 {
		return nil, fmt.Errorf("font %s not found", name)
	}

	return variants, nil
}

// extractFontStyle analyzes a font filename to determine its style
func extractFontStyle(filename string) FontStyle {
	nameLower := strings.ToLower(filename)
	style := FontStyle{}

	// Check for weight indicators
	switch {
	case strings.Contains(nameLower, "thin"):
		style.Weight = WeightThin
	case strings.Contains(nameLower, "extralight"):
		style.Weight = WeightExtraLight
	case strings.Contains(nameLower, "light"):
		style.Weight = WeightLight
	case strings.Contains(nameLower, "medium"):
		style.Weight = WeightMedium
	case strings.Contains(nameLower, "semibold"), strings.Contains(nameLower, "demibold"):
		style.Weight = WeightSemiBold
	case strings.Contains(nameLower, "bold"):
		if strings.Contains(nameLower, "extrabold") || strings.Contains(nameLower, "ultrabold") {
			style.Weight = WeightExtraBold
		} else {
			style.Weight = WeightBold
		}
	case strings.Contains(nameLower, "black"), strings.Contains(nameLower, "heavy"):
		style.Weight = WeightBlack
	default:
		// Only set regular weight if we don't find any other weight indicators
		// and this is not an italic font
		if !strings.Contains(nameLower, "italic") && !strings.Contains(nameLower, "oblique") {
			style.Weight = WeightRegular
		}
	}

	// Check for style indicators
	style.Italic = strings.Contains(nameLower, "italic") || strings.Contains(nameLower, "oblique")
	style.Condensed = strings.Contains(nameLower, "condensed") || strings.Contains(nameLower, "narrow")
	style.Mono = strings.Contains(nameLower, "mono") || strings.Contains(nameLower, "console")

	return style
}

func ListFonts() []string {
	var fonts []string
	seen := make(map[string]bool)

	// List embedded fonts first
	if entries, err := embeddedFonts.ReadDir("embedded"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := cleanFontName(entry.Name())
			if !seen[name] {
				seen[name] = true
				fonts = append(fonts, name)
			}
		}
	}

	paths := systemFontPaths[runtime.GOOS]
	if len(paths) == 0 {
		fmt.Printf("No font paths found for OS %s\n", runtime.GOOS)
		return fonts
	}

	// Expand home directory if necessary
	if home, err := os.UserHomeDir(); err == nil {
		for i, path := range paths {
			if strings.HasPrefix(path, "~") {
				paths[i] = filepath.Join(home, path[2:])
			}
		}
	}

	// Walk through each font directory
	for _, dir := range paths {
		// Skip directories that don't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %s: %v\n", path, err)
				return nil // Skip files we can't access
			}

			// Skip directories
			if d.IsDir() {
				return nil
			}

			// Check if it's a font file
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".otf" {
				return nil
			}

			// Get the base name without extension
			name := filepath.Base(path)
			name = name[:len(name)-len(ext)]

			// Clean up common suffixes
			name = cleanFontName(name)

			// Add to list if we haven't seen it
			if !seen[name] {
				seen[name] = true
				fonts = append(fonts, name)
			}

			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory %s: %v\n", dir, err)
			continue // Skip directories we can't access
		}
	}

	return fonts
}

// cleanFontName removes common suffixes and normalizes the font name
func cleanFontName(name string) string {
	// Common font weights
	weights := []string{
		"Thin",
		"ExtraLight",
		"Light",
		"Regular",
		"Medium",
		"SemiBold",
		"Bold",
		"ExtraBold",
		"Black",
		"Heavy",
	}

	// Common font styles
	styles := []string{
		"Italic",
		"Oblique",
		"Condensed",
		"Narrow",
		"Wide",
		"Mono",
		"Text",
		"Display",
		"Book",
		"Normal",
		"UI",
	}

	// Common separators
	separators := []string{"-", " ", "_"}

	// First, remove the file extension if present
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Special handling for Noto fonts
	if strings.HasPrefix(name, "Noto") {
		// Extract the main font name (e.g., "NotoSans" from "NotoSans-Regular")
		parts := strings.Split(name, "-")
		if len(parts) > 1 {
			// Keep just the family name part
			name = parts[0]
		}

		// Handle UI variants
		name = strings.TrimSuffix(name, "UI")
	}

	// Build combined patterns
	var patterns []string

	// Add weight-only patterns
	patterns = append(patterns, weights...)

	// Add style-only patterns
	patterns = append(patterns, styles...)

	// Add weight+style combinations
	for _, w := range weights {
		for _, s := range styles {
			patterns = append(patterns, w+s, s+w)
		}
	}

	// Add all patterns with different separators
	var allPatterns []string
	for _, p := range patterns {
		// Add pattern as-is
		allPatterns = append(allPatterns, p)

		// Add with separators
		for _, sep := range separators {
			allPatterns = append(allPatterns,
				sep+p,     // -Regular
				p+sep,     // Regular-
				sep+p+sep, // -Regular-
			)
		}
	}

	// Sort patterns by length (longest first) to ensure we remove the most specific patterns first
	sort.Slice(allPatterns, func(i, j int) bool {
		return len(allPatterns[i]) > len(allPatterns[j])
	})

	// Remove version numbers (e.g., "Arial1", "TimesNewRoman2")
	name = strings.TrimRight(name, "0123456789")

	// Remove all matched patterns
	original := name
	for {
		prevName := name
		for _, pattern := range allPatterns {
			// Case insensitive removal from both start and end
			patternLower := strings.ToLower(pattern)
			nameLower := strings.ToLower(name)

			// Remove from end
			if strings.HasSuffix(nameLower, patternLower) {
				name = name[:len(name)-len(pattern)]
			}

			// Remove from start
			if strings.HasPrefix(nameLower, patternLower) {
				name = name[len(pattern):]
			}
		}

		// If no more patterns were removed, break
		if prevName == name {
			break
		}
	}

	// Clean up any remaining separators from the edges
	name = strings.Trim(name, "-_ ")

	// If we removed everything, return the original
	if name == "" {
		return original
	}

	return name
}

// FontFS returns a filesystem containing all available fonts
func FontFS() fs.FS {
	if runtime.GOOS == "windows" {
		return &systemFontFS{root: filepath.Join(os.Getenv("WINDIR"), "Fonts")}
	}

	// Create a multi-FS that combines system fonts with embedded fonts
	return &multiFS{
		filesystems: []fs.FS{
			&systemFontFS{root: systemFontPaths[runtime.GOOS][0]},
			&embeddedFonts,
		},
	}
}

// multiFS implements fs.FS and combines multiple filesystems
type multiFS struct {
	filesystems []fs.FS
}

func (m *multiFS) Open(name string) (fs.File, error) {
	for _, fs := range m.filesystems {
		if f, err := fs.Open(name); err == nil {
			return f, nil
		}
	}
	return nil, os.ErrNotExist
}

// GetFontFace returns a font.Face with the specified size
func (f *Font) GetFontFace(size float64) (font.Face, error) {
	// TODO: The current x/image/font/opentype package does not support variable font features.
	// We need to implement our own font rendering system that can properly handle OpenType
	// variable fonts and their variations (weight, width, slant, etc). This would involve:
	// 1. Parsing the OpenType font tables (fvar, gvar, etc.)
	// 2. Implementing variation interpolation
	// 3. Creating a custom font.Face implementation
	if f == nil || f.Font == nil {
		fallback, err := GetFallback(FallbackMono)
		if err != nil {
			return nil, err
		}
		return fallback.GetFontFace(size)
	}

	// Create face options
	opts := &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	}

	face, err := opentype.NewFace(f.Font, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}

	return face, nil
}

// ClearCache clears the font cache
func ClearCache() {
	fontCacheMu.Lock()
	fontCache = make(map[string][]*Font)
	fontCacheMu.Unlock()
}
