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
)

// embeddedFS is used to store bundled fonts when built with bundle_fonts tag
var embeddedFS *embed.FS

// fontCache caches loaded fonts to avoid repeated disk access
var (
	fontCache      = make(map[string][]*Font)
	fontCacheMu    sync.RWMutex
	variantCache   = make(map[string]*Font)
	variantCacheMu sync.RWMutex
)

// FontStyle represents the style of a font
type FontStyle struct {
	Weight    FontWeight // Font weight
	Italic    bool       // Whether the font is italicS
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

	// Try the requested font first
	variants, err := GetFontVariants(name)
	if err == nil && len(variants) > 0 {
		log.Printf("Found font %s with %d variants", name, len(variants))
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
	}

	// If we're looking for a specific font and it wasn't found, return an error
	// before trying fallbacks
	if len(variants) == 0 {
		return nil, fmt.Errorf("font %s not found", name)
	}

	// Try fallback fonts in order
	fallbacks := []string{
		"Inter",
		"SF Pro",
		"Segoe UI",
		"NotoSans",
		"DejaVuSans",
		"Liberation Sans",
		"Arial",
		"Helvetica",
	}

	for _, fallback := range fallbacks {
		variants, err := GetFontVariants(fallback)
		if err == nil && len(variants) > 0 {
			log.Printf("Found fallback font %s with %d variants", fallback, len(variants))
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
		}
	}

	return nil, fmt.Errorf("no suitable font found for %s", name)
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
		return cached, nil
	}
	fontCacheMu.RUnlock()

	var variants []*Font

	// Search embedded fonts first
	if embeddedFS != nil {
		entries, err := fs.ReadDir(embeddedFS, ".")
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				cleanName := cleanFontName(entry.Name())
				if strings.EqualFold(cleanName, name) {
					data, err := fs.ReadFile(embeddedFS, entry.Name())
					if err != nil {
						continue
					}

					font := &Font{
						Name:     cleanName,
						Data:     data,
						Style:    extractFontStyle(entry.Name()),
						FilePath: entry.Name(),
					}
					variants = append(variants, font)
				}
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

				font := &Font{
					Name:     cleanName,
					Data:     data,
					Style:    extractFontStyle(info.Name()),
					FilePath: path,
				}
				variants = append(variants, font)
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
			fmt.Printf("Directory does not exist: %s\n", dir)
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

			fmt.Printf("Found font: %s (from %s)\n", name, path)

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

func FontFS() fs.FS {
	if embeddedFS != nil {
		return embeddedFS
	}

	// Return the first available system font directory
	paths := systemFontPaths[runtime.GOOS]
	if len(paths) > 0 {
		return systemFontFS{root: paths[0]}
	}
	return nil
}

// ToTrueType converts the Font to a truetype.Font
func (f *Font) ToTrueType() (*truetype.Font, error) {
	if f == nil || len(f.Data) == 0 {
		return nil, fmt.Errorf("invalid font or empty font data")
	}
	return truetype.Parse(f.Data)
}

// GetFontFace returns a font.Face with the specified size
func (f *Font) GetFontFace(size float64) (font.Face, error) {
	font, err := f.ToTrueType()
	if err != nil {
		return nil, err
	}
	return truetype.NewFace(font, &truetype.Options{
		Size: size,
	}), nil
}

// ClearCache clears the font cache
func ClearCache() {
	fontCacheMu.Lock()
	fontCache = make(map[string][]*Font)
	fontCacheMu.Unlock()
}
