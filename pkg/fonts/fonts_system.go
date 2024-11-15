package fonts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// systemFontPaths contains the default system font directories for different operating systems
var systemFontPaths = map[string][]string{
	"linux": {
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		"~/.local/share/fonts",
		"~/.fonts",
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

func getFont(name string, style *FontStyle) (*Font, error) {
	variants, err := getFontVariants(name)
	if err != nil {
		return nil, err
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("font %s not found", name)
	}

	// If no style specified, return regular if available, otherwise first variant
	if style == nil {
		for _, font := range variants {
			if font.Style.Weight == WeightRegular && !font.Style.Italic {
				return font, nil
			}
		}
		return variants[0], nil
	}

	// Find best matching style
	var bestMatch *Font
	var bestScore int
	for _, font := range variants {
		score := matchStyleScore(font.Style, *style)
		if score > bestScore {
			bestMatch = font
			bestScore = score
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("font %s with requested style not found", name)
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

func getFontVariants(name string) ([]*Font, error) {
	paths := systemFontPaths[runtime.GOOS]
	if paths == nil {
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Expand home directory if necessary
	if home, err := os.UserHomeDir(); err == nil {
		for i, path := range paths {
			if strings.HasPrefix(path, "~") {
				paths[i] = filepath.Join(home, path[2:])
			}
		}
	}

	var variants []*Font
	seen := make(map[string]bool)

	// Search for the font in system paths
	for _, dir := range paths {
		// Skip directories that don't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
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

			// Check if this font matches our search
			baseName := cleanFontName(filepath.Base(path[:len(path)-len(ext)]))
			if strings.EqualFold(baseName, name) {
				// Read the font data
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				// Create the font object
				font := &Font{
					Name:     name,
					FilePath: path,
					Data:     data,
					Style:    extractFontStyle(filepath.Base(path)),
				}

				// Only add if we haven't seen this exact file before
				if !seen[path] {
					seen[path] = true
					variants = append(variants, font)
				}
			}

			return nil
		})
		if err != nil {
			continue // Skip directories we can't access
		}
	}

	return variants, nil
}

// extractFontStyle analyzes a font filename to determine its style
func extractFontStyle(filename string) FontStyle {
	nameLower := strings.ToLower(filename)
	style := FontStyle{
		Weight: WeightRegular,
	}

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
	}

	// Check for style indicators
	style.Italic = strings.Contains(nameLower, "italic") || strings.Contains(nameLower, "oblique")
	style.Condensed = strings.Contains(nameLower, "condensed") || strings.Contains(nameLower, "narrow")
	style.Mono = strings.Contains(nameLower, "mono") || strings.Contains(nameLower, "console")

	return style
}

func listFonts() []string {
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
	}

	// Common separators
	separators := []string{"-", " ", "_"}

	// First, remove the file extension if present
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Build combined patterns
	var patterns []string

	// Add weight-only patterns
	for _, w := range weights {
		patterns = append(patterns, w)
	}

	// Add style-only patterns
	for _, s := range styles {
		patterns = append(patterns, s)
	}

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

func fontFS() fs.FS {
	// Return the first available system font directory
	paths := systemFontPaths[runtime.GOOS]
	if len(paths) > 0 {
		return systemFontFS{root: paths[0]}
	}
	return nil
}
