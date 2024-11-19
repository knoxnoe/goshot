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
func DrawTitleText(dc *gg.Context, title string, width, titleBarHeight int, textColor color.Color, fontSize float64) error {
	// Draw text centered horizontally and vertically in the title bar
	x := float64(width) / 2
	y := float64(titleBarHeight) / 2
	return drawTitleText(dc, title, x, y, fontSize, textColor)
}

func drawTitleText(ctx *gg.Context, text string, x, y float64, size float64, c color.Color) error {
	log.Printf("Drawing title text: %q at size %.1f", text, size)

	// Try system fonts in order
	fontNames := []string{
		"Inter",
		"SF Pro",
		"Segoe UI",
		"NotoSans",
		"DejaVuSans",
		"Liberation Sans",
	}

	var font *fonts.Font
	var err error

	for _, name := range fontNames {
		font, err = fonts.GetFont(name, nil)
		if err == nil {
			log.Printf("Using font: %s", name)
			break
		}
		log.Printf("Failed to load %s font: %v", name, err)
	}

	if font == nil {
		return fmt.Errorf("failed to load any system fonts: %v", err)
	}

	face, err := font.GetFontFace(size)
	if err != nil {
		return err
	}

	ctx.SetFontFace(face)
	ctx.SetColor(c)

	// Get text dimensions
	w, h := ctx.MeasureString(text)

	// Draw the text centered at the specified position
	ctx.DrawStringAnchored(text, x, y, 0.5, 0.5)
	log.Printf("Drew text at %.1f,%.1f (width: %.1f, height: %.1f)", x, y, w, h)

	return nil
}

// DrawWindowControl draws a circular window control button
func DrawWindowControl(dc *gg.Context, x, y, size float64, color color.Color) {
	dc.SetColor(color)
	dc.DrawCircle(x+size/2, y+size/2, size/2)
	dc.Fill()
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

// DrawWindowControls draws the window control buttons based on the style
func DrawWindowControls(dc *gg.Context, width, height int, style WindowStyle) {
	switch style {
	case MacOSStyle:
		// macOS style - left aligned "traffic light" buttons
		buttonSize := 12.0
		spacing := 8.0
		x := spacing
		y := float64(height)/2 - buttonSize/2

		// Close button (red)
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{R: 255, G: 95, B: 87, A: 255})
		x += buttonSize + spacing

		// Minimize button (yellow)
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{R: 255, G: 189, B: 46, A: 255})
		x += buttonSize + spacing

		// Maximize button (green)
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{R: 39, G: 201, B: 63, A: 255})

	case Windows11Style:
		// Windows 11 style - right aligned buttons
		buttonSize := 14.0
		spacing := 6.0
		y := float64(height)/2 - buttonSize/2

		// Close button (rightmost)
		x := float64(width) - buttonSize - spacing
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{A: 255})
		DrawCross(dc, x, y, buttonSize, color.RGBA{A: 255})

		// Maximize button
		x -= buttonSize + spacing
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{A: 255})
		DrawSquare(dc, x, y, buttonSize, color.RGBA{A: 255})

		// Minimize button
		x -= buttonSize + spacing
		DrawWindowControl(dc, x, y, buttonSize, color.RGBA{A: 255})
		DrawLine(dc, x, y+buttonSize/2, buttonSize, color.RGBA{A: 255})
	}
}
