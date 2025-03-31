package fonts

import (
	"embed"
	"fmt"
	"image"
	"io"
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
	"golang.org/x/image/math/fixed"
)

//go:embed embedded
var embeddedFonts embed.FS

// fontCache caches loaded fonts to avoid re-parsing font files
var (
	fontCache   = make(map[string][]*Font)
	fontCacheMu sync.RWMutex
)

// FontStretch represents the width/stretch of a font
type FontStretch int

const (
	StretchUltraCondensed FontStretch = iota + 1
	StretchExtraCondensed
	StretchCondensed
	StretchSemiCondensed
	StretchNormal
	StretchSemiExpanded
	StretchExpanded
	StretchExtraExpanded
	StretchUltraExpanded
)

// FontStyle represents the style of a font
type FontStyle struct {
	Weight    FontWeight  // Font weight
	Stretch   FontStretch // Font width/stretch
	Italic    bool        // Whether the font is italic
	Underline bool        // Whether the font should be underlined
	Mono      bool        // Whether the font is monospaced
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

// Font represents a loaded font
type Font struct {
	Name        string         // Name of the font family
	Font        *opentype.Font // Parsed font
	FilePath    string         // Path to the font file on disk
	Filename    string         // Name of the font file
	IsMonospace bool           // Whether the font is monospaced
	Style       FontStyle      // Font style
	maxWidth    fixed.Int26_6  // Maximum glyph width (lazy loaded)
	maxWidthMu  sync.Once      // Ensures maxWidth is computed only once
}

// Face represents a font face with specific style and size
type Face struct {
	Font  *Font
	Style FontStyle
	Size  float64
	Face  font.Face
}

// GetFace returns a new Face with the specified style and size
func (f *Font) GetFace(size float64, style *FontStyle) (*Face, error) {
	if style == nil {
		style = &FontStyle{
			Weight:  WeightRegular,
			Stretch: StretchNormal,
		}
	}

	// Try to find a font variant that matches our style
	variants, err := GetFontVariants(f.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get font variants: %v", err)
	}

	// Find the best matching variant
	var bestVariant *Font
	bestScore := -1
	for _, variant := range variants {
		score := matchStyleScore(variant.Style, *style)
		if score > bestScore {
			bestScore = score
			bestVariant = variant
		}
	}

	face, err := opentype.NewFace(bestVariant.Font, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create face: %v", err)
	}

	return &Face{
		Font:  bestVariant,
		Style: bestVariant.Style,
		Size:  size,
		Face:  face,
	}, nil
}

// Close releases the resources used by the face
func (f *Face) Close() {
	if closer, ok := f.Face.(io.Closer); ok {
		closer.Close()
	}
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

func detectMonospace(ttf *truetype.Font) bool {
	// Get the advance width of a reference character (e.g., 'M')
	refIndex := ttf.Index('M')
	if refIndex == 0 {
		refIndex = ttf.Index('m')
	}
	if refIndex == 0 {
		return false
	}

	refWidth := ttf.HMetric(2048, refIndex).AdvanceWidth // Use a standard units-per-em value

	// Check a range of common characters
	for _, r := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		idx := ttf.Index(r)
		if idx == 0 {
			continue
		}

		width := ttf.HMetric(2048, idx).AdvanceWidth
		if width != refWidth {
			return false
		}
	}

	return true
}

// IsMono returns whether this font is monospaced
func (f *Font) IsMono() bool {
	return f.IsMonospace
}

// GetMaxWidth returns the maximum glyph width in the font
func (f *Font) GetMaxWidth() (fixed.Int26_6, error) {
	if f == nil || f.Font == nil {
		return 0, fmt.Errorf("invalid font")
	}

	f.maxWidthMu.Do(func() {
		ttf, err := f.ToTrueType()
		if err != nil {
			return
		}

		// Standard UPM (units per em) value
		const upm = 2048

		// Check common characters and some typically wide ones
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
			"WMm@★→←↔⇒⇐⇔▶◀△▽□◇○◎●◐◑∇∆∩∪∫∮∼≒≠≡≦≧≨≩⊂⊃⊆⊇⊕⊖⊗⊘⊙⊚⊛⊜⊝⊞⊟"

		var maxWidth fixed.Int26_6
		for _, r := range chars {
			idx := ttf.Index(r)
			if idx == 0 {
				continue
			}

			width := ttf.HMetric(upm, idx).AdvanceWidth
			if width > fixed.Int26_6(maxWidth) {
				maxWidth = fixed.Int26_6(width)
			}
		}

		f.maxWidth = maxWidth
	})

	return f.maxWidth, nil
}

// MeasureString returns the width of a string in pixels
func (f *Font) MeasureString(s string, size float64, style *FontStyle) (fixed.Int26_6, error) {
	face, err := f.GetFace(size, style)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	return font.MeasureString(face.Face, s), nil
}

// GetMonoFace returns a font face that will render with fixed-width characters
// size is the desired font size in points
// cellWidth is the desired cell width in pixels (if 0, uses the font's natural maximum width)
func (f *Font) GetMonoFace(size float64, cellWidth int) (*Face, error) {
	if f == nil || f.Font == nil {
		return nil, fmt.Errorf("invalid font")
	}

	// Create the base face first
	baseOpts := &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}

	baseFace, err := opentype.NewFace(f.Font, baseOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create base face: %v", err)
	}

	// If the font is already monospace and no specific cell width is requested,
	// return the base face as is
	if f.IsMonospace && cellWidth == 0 {
		return &Face{
			Font: f,
			Style: FontStyle{
				Weight:  WeightRegular,
				Stretch: StretchNormal,
			},
			Size: size,
			Face: baseFace,
		}, nil
	}

	// Create a wrapper face that forces fixed width
	maxWidth := cellWidth
	if maxWidth == 0 {
		// Use the font's natural maximum width
		naturalMax, err := f.GetMaxWidth()
		if err != nil {
			return nil, err
		}
		maxWidth = int(naturalMax * fixed.Int26_6(size) / 2048)
	}

	monoFace := &monospaceFace{
		Face:      baseFace,
		cellWidth: fixed.I(maxWidth),
	}

	return &Face{
		Font: f,
		Style: FontStyle{
			Weight:  WeightRegular,
			Stretch: StretchNormal,
		},
		Size: size,
		Face: monoFace,
	}, nil
}

// monospaceFace wraps a font.Face to make it render with fixed-width characters
type monospaceFace struct {
	Face      font.Face
	cellWidth fixed.Int26_6
}

func (m *monospaceFace) Close() error {
	if closer, ok := m.Face.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (m *monospaceFace) Glyph(dot fixed.Point26_6, r rune) (
	dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {

	// Get the original glyph
	dr, mask, maskp, advance, ok = m.Face.Glyph(dot, r)
	if !ok {
		return
	}

	// Center the glyph in its cell
	width := m.cellWidth
	shift := (width - advance) / 2
	dr = dr.Add(image.Point{X: fixed.Int26_6(shift).Round(), Y: 0})
	maskp = maskp.Add(image.Point{X: fixed.Int26_6(shift).Round(), Y: 0})

	// Return the fixed advance width
	return dr, mask, maskp, width, true
}

func (m *monospaceFace) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	bounds, advance, ok = m.Face.GlyphBounds(r)
	if !ok {
		return
	}

	// Center the bounds in the cell
	width := m.cellWidth
	shift := (width - advance) / 2
	bounds.Min.X += shift
	bounds.Max.X += shift

	return bounds, width, true
}

func (m *monospaceFace) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	_, ok = m.Face.GlyphAdvance(r)
	if !ok {
		return 0, false
	}
	return m.cellWidth, true
}

func (m *monospaceFace) Kern(r0, r1 rune) fixed.Int26_6 {
	// No kerning in monospace
	return 0
}

func (m *monospaceFace) Metrics() font.Metrics {
	return m.Face.Metrics()
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

// GetFont returns a font with the specified name and style
func GetFont(name string, style *FontStyle) (*Font, error) {
	if name == "" {
		return nil, fmt.Errorf("font name cannot be empty")
	}

	// Normalize font name
	name = cleanFontName(name)

	// Try to load all variants of the font
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
			score := matchStyleScore(*style, font.Style)
			if score > bestScore {
				selected = font
				bestScore = score
			}
		}
	}

	if selected != nil {
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

// GetFallback returns either JetBrainsMono or Inter as the fallback font
func GetFallback(variant FallbackVariant) (font *Font, err error) {
	switch variant {
	case FallbackMono:
		font, err = GetFont("JetBrainsMonoNerdFont", nil) // Let GetFont handle style selection
	case FallbackSans:
		font, err = GetFont("Inter", nil) // Let GetFont handle style selection
	default:
		return nil, fmt.Errorf("invalid fallback variant")
	}
	return
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

	// Weight matching (max 100 points)
	weightDiff := abs(int(a.Weight) - int(b.Weight))
	score += 100 - weightDiff*10 // Lose 10 points for each step away from desired weight

	// Stretch matching (max 50 points)
	stretchDiff := abs(int(a.Stretch) - int(b.Stretch))
	score += 50 - stretchDiff*5 // Lose 5 points for each step away from desired stretch

	// Exact italic match (50 points)
	if a.Italic == b.Italic {
		score += 50
	}

	// Exact underline match (50 points)
	if a.Underline == b.Underline {
		score += 50
	}

	// Exact mono match (25 points)
	if a.Mono == b.Mono {
		score += 25
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
		if cached == nil {
			return nil, fmt.Errorf("font %s not found", name)
		}
		return cached, nil
	}
	fontCacheMu.RUnlock()

	var variants []*Font

	// First check embedded fonts
	entries, err := embeddedFonts.ReadDir("embedded")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			cleanName := cleanFontName(entry.Name())
			if cleanName == name {
				data, err := embeddedFonts.ReadFile("embedded/" + entry.Name())
				if err != nil {
					continue
				}

				font, err := opentype.Parse(data)
				if err != nil {
					continue
				}

				// Convert to truetype to check if monospace
				ttf, err := truetype.Parse(data)
				if err != nil {
					continue
				}

				variant := &Font{
					Name:        cleanName,
					Font:        font,
					Filename:    entry.Name(),
					IsMonospace: detectMonospace(ttf),
					Style:       extractFontStyle(entry.Name()),
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
			if cleanName == name {
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				font, err := opentype.Parse(data)
				if err != nil {
					return nil
				}

				// Convert to truetype to check if monospace
				ttf, err := truetype.Parse(data)
				if err != nil {
					return nil
				}

				variant := &Font{
					Name:        cleanName,
					Font:        font,
					FilePath:    path,
					IsMonospace: detectMonospace(ttf),
					Style:       extractFontStyle(filepath.Base(path)),
				}
				variants = append(variants, variant)
			}

			return nil
		})

		if err != nil {
			log.Printf("Error walking font directory %s: %v", basePath, err)
		}
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("font %s not found", name)
	}

	// Cache the results
	fontCacheMu.Lock()
	fontCache[name] = variants
	fontCacheMu.Unlock()

	return variants, nil
}

// extractFontStyle analyzes a font filename to determine its style
func extractFontStyle(filename string) FontStyle {
	filename = strings.ToLower(filename)
	style := FontStyle{
		Weight:  WeightRegular,
		Stretch: StretchNormal,
	}

	// Common weight indicators in filenames
	switch {
	case strings.Contains(filename, "thin"):
		style.Weight = WeightThin
	case strings.Contains(filename, "extralight") || strings.Contains(filename, "extra-light") || strings.Contains(filename, "ultra-light") || strings.Contains(filename, "ultralight"):
		style.Weight = WeightExtraLight
	case strings.Contains(filename, "light"):
		style.Weight = WeightLight
	case strings.Contains(filename, "regular") || strings.Contains(filename, "normal"):
		style.Weight = WeightRegular
	case strings.Contains(filename, "medium"):
		style.Weight = WeightMedium
	case strings.Contains(filename, "semibold") || strings.Contains(filename, "semi-bold") || strings.Contains(filename, "demibold") || strings.Contains(filename, "demi-bold"):
		style.Weight = WeightSemiBold
	case strings.Contains(filename, "extrabold") || strings.Contains(filename, "extra-bold") || strings.Contains(filename, "ultra-bold") || strings.Contains(filename, "ultrabold"):
		style.Weight = WeightExtraBold
	case strings.Contains(filename, "black") || strings.Contains(filename, "heavy"):
		style.Weight = WeightBlack
	case strings.Contains(filename, "bold"):
		style.Weight = WeightBold
	}

	// Common stretch indicators
	switch {
	case strings.Contains(filename, "ultracondensed") || strings.Contains(filename, "ultra-condensed"):
		style.Stretch = StretchUltraCondensed
	case strings.Contains(filename, "extracondensed") || strings.Contains(filename, "extra-condensed"):
		style.Stretch = StretchExtraCondensed
	case strings.Contains(filename, "condensed"):
		style.Stretch = StretchCondensed
	case strings.Contains(filename, "semicondensed") || strings.Contains(filename, "semi-condensed"):
		style.Stretch = StretchSemiCondensed
	case strings.Contains(filename, "expanded"):
		style.Stretch = StretchExpanded
	case strings.Contains(filename, "extraexpanded") || strings.Contains(filename, "extra-expanded"):
		style.Stretch = StretchExtraExpanded
	case strings.Contains(filename, "ultraexpanded") || strings.Contains(filename, "ultra-expanded"):
		style.Stretch = StretchUltraExpanded
	}

	// Italic/Oblique detection
	style.Italic = strings.Contains(filename, "italic") || strings.Contains(filename, "oblique")

	// Underline detection
	style.Underline = strings.Contains(filename, "underline")

	// Mono detection
	style.Mono = strings.Contains(filename, "mono") || strings.Contains(filename, "console") || strings.Contains(filename, "typewriter")

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

	// Common font stretches
	stretches := []string{
		"UltraCondensed",
		"ExtraCondensed",
		"Condensed",
		"SemiCondensed",
		"Normal",
		"SemiExpanded",
		"Expanded",
		"ExtraExpanded",
		"UltraExpanded",
	}

	// Common font styles
	styles := []string{
		"Italic",
		"Oblique",
		"Mono",
		"Console",
		"Underline",
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

	// Add stretch-only patterns
	patterns = append(patterns, stretches...)

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

// ClearCache clears the font cache
func ClearCache() {
	fontCacheMu.Lock()
	fontCache = make(map[string][]*Font)
	fontCacheMu.Unlock()
}
