package chrome

import (
	"image/color"
	"math"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/assert"
	"github.com/watzon/goshot/pkg/fonts"
)

// Helper function to check if a color matches expected RGB values
func colorMatches(t *testing.T, c color.Color, minR, minG, minB uint32, msg string) bool {
	r, g, b, _ := c.RGBA()
	return assert.True(t, r >= minR && g >= minG && b >= minB, msg)
}

func TestDrawTitleText(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		width     int
		height    int
		fontSize  float64
		textColor color.Color
		setup     func()
		validate  func(*testing.T, *gg.Context)
	}{
		{
			name:      "Basic title text",
			title:     "Test Window",
			width:     200,
			height:    30,
			fontSize:  12,
			textColor: color.Black,
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
				for x := centerX - 5; x <= centerX + 5; x++ {
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
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				assert.Equal(t, 100, bounds.Dx(), "Image width should match")
				assert.Equal(t, 20, bounds.Dy(), "Image height should match")
			},
		},
		{
			name:      "Colored text",
			title:     "Colored Window",
			width:     150,
			height:    25,
			fontSize:  11,
			textColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			setup:     func() { fonts.ClearCache() },
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()
				bounds := img.Bounds()
				centerX := bounds.Dx() / 2
				centerY := bounds.Dy() / 2

				// Check multiple pixels around the center for red text
				found := false
				for x := centerX - 5; x <= centerX + 5; x++ {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx := gg.NewContext(tt.width, tt.height)
			ctx.SetRGB(1, 1, 1)
			ctx.Clear()

			err := DrawTitleText(ctx, tt.title, tt.width, tt.height, tt.textColor, tt.fontSize)
			assert.NoError(t, err, "DrawTitleText should not return error")

			tt.validate(t, ctx)
		})
	}
}

func TestDrawWindowControl(t *testing.T) {
	tests := []struct {
		name     string
		x, y     float64
		diameter float64
		color    color.Color
		validate func(*testing.T, *gg.Context)
	}{
		{
			name:     "Red close button",
			x:        15,
			y:        15,
			diameter: 12,
			color:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()

				// Sample multiple points near the center of the button
				found := false
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x, y := 15+float64(dx), 15+float64(dy)
						c := img.At(int(x+6), int(y+6))
						r, g, b, _ := c.RGBA()
						if r > 32768 && g < 16384 && b < 16384 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find red pixels near button center")

				// Check edge pixels
				found = false
				for theta := 0.0; theta < 6.28; theta += 0.5 {
					x := 21 + 6*math.Cos(theta)
					y := 21 + 6*math.Sin(theta)
					c := img.At(int(x), int(y))
					r, g, b, _ := c.RGBA()
					if r > 32768 && g < 16384 && b < 16384 {
						found = true
						break
					}
				}
				assert.True(t, found, "Should find red pixels on button edge")
			},
		},
		{
			name:     "Yellow minimize button",
			x:        45,
			y:        15,
			diameter: 12,
			color:    color.RGBA{R: 255, G: 255, B: 0, A: 255},
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()

				// Sample multiple points near the center
				found := false
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x, y := 45+float64(dx), 15+float64(dy)
						c := img.At(int(x+6), int(y+6))
						r, g, b, _ := c.RGBA()
						if r > 32768 && g > 32768 && b < 16384 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find yellow pixels near button center")
			},
		},
		{
			name:     "Green maximize button",
			x:        75,
			y:        15,
			diameter: 12,
			color:    color.RGBA{R: 0, G: 255, B: 0, A: 255},
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()

				// Sample multiple points near the center
				found := false
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x, y := 75+float64(dx), 15+float64(dy)
						c := img.At(int(x+6), int(y+6))
						r, g, b, _ := c.RGBA()
						if r < 16384 && g > 32768 && b < 16384 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find green pixels near button center")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := gg.NewContext(100, 30)
			ctx.SetRGB(1, 1, 1)
			ctx.Clear()

			DrawWindowControl(ctx, tt.x, tt.y, tt.diameter, tt.color)

			tt.validate(t, ctx)
		})
	}
}

func TestDrawWindowControls(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		style    WindowStyle
		validate func(*testing.T, *gg.Context)
	}{
		{
			name:   "macOS style controls",
			width:  200,
			height: 30,
			style:  MacOSStyle,
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()

				// Check for red close button
				found := false
				for dx := -2; dx <= 2; dx++ {
					for dy := -2; dy <= 2; dy++ {
						x, y := 15+dx, 15+dy
						c := img.At(x, y)
						r, g, b, _ := c.RGBA()
						if r > 32768 && g < 32768 && b < 32768 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find red close button")

				// Check for yellow minimize button
				found = false
				for dx := -2; dx <= 2; dx++ {
					for dy := -2; dy <= 2; dy++ {
						x, y := 35+dx, 15+dy
						c := img.At(x, y)
						r, g, b, _ := c.RGBA()
						if r > 32768 && g > 32768 && b < 16384 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find yellow minimize button")

				// Check for green maximize button
				found = false
				for dx := -2; dx <= 2; dx++ {
					for dy := -2; dy <= 2; dy++ {
						x, y := 55+dx, 15+dy
						c := img.At(x, y)
						r, g, b, _ := c.RGBA()
						if r < 16384 && g > 32768 && b < 16384 {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Should find green maximize button")
			},
		},
		{
			name:   "Windows 11 style controls",
			width:  200,
			height: 30,
			style:  Windows11Style,
			validate: func(t *testing.T, ctx *gg.Context) {
				img := ctx.Image()

				// Check for buttons by looking for non-white pixels
				checkButton := func(x, y int) bool {
					for dx := -2; dx <= 2; dx++ {
						for dy := -2; dy <= 2; dy++ {
							c := img.At(x+dx, y+dy)
							_, _, _, a := c.RGBA()
							if a > 0 {
								return true
							}
						}
					}
					return false
				}

				assert.True(t, checkButton(185, 15), "Close button should be visible")
				assert.True(t, checkButton(165, 15), "Maximize button should be visible")
				assert.True(t, checkButton(145, 15), "Minimize button should be visible")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := gg.NewContext(tt.width, tt.height)
			ctx.SetRGB(1, 1, 1)
			ctx.Clear()

			DrawWindowControls(ctx, tt.width, tt.height, tt.style)

			tt.validate(t, ctx)
		})
	}
}
