package chrome

import (
	"image/color"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/assert"
	"github.com/watzon/goshot/pkg/fonts"
)

func TestDrawTitleText(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		width     int
		height    int
		fontSize  float64
		textColor color.Color
		fontName  string
		setup     func()
		validate  func(*testing.T, *gg.Context)
	}{
		{
			name:      "Basic title text with fallback font",
			title:     "Test Window",
			width:     200,
			height:    30,
			fontSize:  12,
			textColor: color.Black,
			fontName:  "", // Empty string to test fallback
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				assert.Equal(t, 200, bounds.Dx(), "Image width should match")
				assert.Equal(t, 30, bounds.Dy(), "Image height should match")

				// Check multiple pixels around the center for black text
				centerX := bounds.Dx() / 2
				centerY := bounds.Dy() / 2
				found := false
				for x := centerX - 5; x <= centerX+5; x++ {
					c := img.At(x, centerY)
					r, g, b, _ := c.RGBA()
					if r < 65535 && g < 65535 && b < 65535 {
						found = true
						break
					}
				}
				assert.True(t, found, "Should find some black text pixels near center")
			},
		},
		{
			name:      "Empty title",
			title:     "",
			width:     100,
			height:    20,
			fontSize:  10,
			textColor: color.Black,
			fontName:  "", // Empty string to test fallback
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				assert.Equal(t, 100, bounds.Dx(), "Image width should match")
				assert.Equal(t, 20, bounds.Dy(), "Image height should match")
			},
		},
		{
			name:      "Colored text with SF Pro font",
			title:     "Colored Window",
			width:     150,
			height:    25,
			fontSize:  11,
			textColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			fontName:  "SF Pro", // macOS default font
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				centerX := bounds.Dx() / 2
				centerY := bounds.Dy() / 2

				// Check multiple pixels around the center for red text
				found := false
				for x := centerX - 5; x <= centerX+5; x++ {
					c := img.At(x, centerY)
					r, g, b, _ := c.RGBA()
					if r > 32768 && g < 16384 && b < 16384 {
						found = true
						break
					}
				}
				assert.True(t, found, "Should find some red text pixels near center")
			},
		},
		{
			name:      "Text with Segoe UI font",
			title:     "Windows Style",
			width:     150,
			height:    25,
			fontSize:  11,
			textColor: color.Black,
			fontName:  "Segoe UI", // Windows default font
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				centerX := bounds.Dx() / 2
				centerY := bounds.Dy() / 2

				// Check multiple pixels around the center for black text
				found := false
				for x := centerX - 5; x <= centerX+5; x++ {
					c := img.At(x, centerY)
					r, g, b, _ := c.RGBA()
					if r < 32768 && g < 32768 && b < 32768 {
						found = true
						break
					}
				}
				assert.True(t, found, "Should find some black text pixels near center")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx := gg.NewContext(tt.width, tt.height)
			ctx.SetRGB(1, 1, 1)
			ctx.Clear()

			err := DrawTitleText(ctx, tt.title, tt.width, tt.height, tt.textColor, tt.fontSize, tt.fontName)
			assert.NoError(t, err, "DrawTitleText should not return error")

			tt.validate(t, ctx)
		})
	}
}
