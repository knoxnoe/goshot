package chrome

import (
	"fmt"
	"image/color"
	"log"

	"github.com/fogleman/gg"
	"github.com/watzon/goshot/pkg/fonts"
)

// WindowStyle represents the style of window controls (macOS, Windows, etc)
type WindowStyle int

const (
	// MacOSStyle represents macOS-style window controls
	MacOSStyle WindowStyle = iota
	// Windows11Style represents Windows 11-style window controls
	Windows11Style
)

// DrawWindowBase draws the base window shape with rounded corners
func DrawWindowBase(dc *gg.Context, width int, height int, cornerRadius float64, titleBackground, contentBackground color.Color, titleBarHeight int) error {
	// Clear the background to transparent
	dc.Clear()

	// Draw the rounded rectangle for clipping
	dc.DrawRoundedRectangle(0, 0, float64(width), float64(height), cornerRadius)
	dc.Clip()

	// Draw content background
	dc.SetColor(contentBackground)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()

	// Draw title bar background
	dc.SetColor(titleBackground)
	dc.DrawRectangle(0, 0, float64(width), float64(titleBarHeight))
	dc.Fill()

	return nil
}

// DrawTitleText draws centered title text in the title bar
func DrawTitleText(dc *gg.Context, title string, width, titleBarHeight int, textColor color.Color, fontSize float64, fontName string) error {
	// Draw text centered horizontally and vertically in the title bar
	x := float64(width) / 2
	y := float64(titleBarHeight) / 2
	return drawTitleText(dc, title, x, y, fontSize, textColor, fontName)
}

func drawTitleText(ctx *gg.Context, text string, x, y float64, size float64, c color.Color, fontName string) error {
	var font *fonts.Font
	var err error

	// Try to load the requested font
	if fontName != "" {
		font, err = fonts.GetFont(fontName, nil)
		if err != nil {
			log.Printf("Failed to load requested font %s: %v", fontName, err)
		}
	}

	// If the requested font failed to load, use fallback
	if font == nil {
		font, err = fonts.GetFallback(fonts.FallbackSans)
		if err != nil {
			return fmt.Errorf("failed to load fallback font: %v", err)
		}
	}

	face, err := font.GetFontFace(size)
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}

	ctx.SetFontFace(face)
	ctx.SetColor(c)

	// Adjust Y position to account for font metrics and achieve true vertical centering
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent.Round())
	descent := float64(metrics.Descent.Round())
	adjustment := (ascent - descent) / 2
	y = y - adjustment/2

	// Draw the text centered at the specified position
	ctx.DrawStringAnchored(text, x, y, 0.5, 0.5)

	return nil
}

// DrawCross draws an X symbol for the close button
func DrawCross(dc *gg.Context, x, y, size float64, color color.Color) {
	dc.SetColor(color)
	dc.SetLineWidth(1.5)

	padding := size * 0.2
	x += padding
	y += padding
	size -= padding * 2

	// Draw X
	dc.MoveTo(x, y)
	dc.LineTo(x+size, y+size)
	dc.Stroke()

	dc.MoveTo(x+size, y)
	dc.LineTo(x, y+size)
	dc.Stroke()
}

// DrawSquare draws a square symbol for the maximize button
func DrawSquare(dc *gg.Context, x, y, size float64, color color.Color) {
	dc.SetColor(color)
	dc.SetLineWidth(1.5)

	padding := size * 0.2
	x += padding
	y += padding
	size -= padding * 2

	// Draw square using individual lines to avoid corner overlaps
	dc.MoveTo(x, y)
	dc.LineTo(x+size, y)
	dc.Stroke()

	dc.MoveTo(x+size, y)
	dc.LineTo(x+size, y+size)
	dc.Stroke()

	dc.MoveTo(x+size, y+size)
	dc.LineTo(x, y+size)
	dc.Stroke()

	dc.MoveTo(x, y+size)
	dc.LineTo(x, y)
	dc.Stroke()
}

// DrawLine draws a horizontal line for the minimize button
func DrawLine(dc *gg.Context, x, y, size float64, color color.Color) {
	dc.SetColor(color)
	dc.SetLineWidth(1.5)

	padding := size * 0.2
	x += padding
	size -= padding * 2

	// Draw horizontal line
	dc.MoveTo(x, y)
	dc.LineTo(x+size, y)
	dc.Stroke()
}
