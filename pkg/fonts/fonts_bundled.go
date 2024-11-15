//go:build bundled

package fonts

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
)

//go:embed fonts/*
var embeddedFonts embed.FS

// bundledFonts maps font names to their style variants
var bundledFonts = map[string][]struct {
	filename string
	style    FontStyle
}{
	"JetBrainsMono": {
		{"JetBrainsMono-Regular.ttf", FontStyle{Weight: WeightRegular}},
		{"JetBrainsMono-Bold.ttf", FontStyle{Weight: WeightBold}},
		{"JetBrainsMono-Italic.ttf", FontStyle{Weight: WeightRegular, Italic: true}},
		{"JetBrainsMono-BoldItalic.ttf", FontStyle{Weight: WeightBold, Italic: true}},
	},
	"FiraCode": {
		{"FiraCode-Regular.ttf", FontStyle{Weight: WeightRegular, Mono: true}},
		{"FiraCode-Bold.ttf", FontStyle{Weight: WeightBold, Mono: true}},
	},
}

type FontStyle struct {
	Weight int
	Italic bool
	Mono   bool
}

func matchStyleScore(a, b FontStyle) int {
	score := 0
	if a.Weight == b.Weight {
		score++
	}
	if a.Italic == b.Italic {
		score++
	}
	if a.Mono == b.Mono {
		score++
	}
	return score
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

func getFontVariants(name string) ([]*Font, error) {
	variants, ok := bundledFonts[name]
	if !ok {
		return nil, fmt.Errorf("font %s not found", name)
	}

	var fonts []*Font
	for _, variant := range variants {
		data, err := embeddedFonts.ReadFile(filepath.Join("fonts", variant.filename))
		if err != nil {
			continue
		}

		fonts = append(fonts, &Font{
			Name:     name,
			Data:     data,
			Style:    variant.style,
			FilePath: variant.filename,
		})
	}

	return fonts, nil
}

func listFonts() []string {
	var fonts []string
	for name := range bundledFonts {
		fonts = append(fonts, name)
	}
	return fonts
}

func fontFS() fs.FS {
	return embeddedFonts
}

type Font struct {
	Name     string
	Data     []byte
	Style    FontStyle
	FilePath string
}

const (
	WeightRegular = 400
	WeightBold    = 700
)
