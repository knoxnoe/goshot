package term

import (
	"fmt"
	"image"
	"image/draw"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/watzon/goshot/content"
	"github.com/watzon/goshot/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Ensure TermRenderer implements content.Content
var _ content.Content = (*TermRenderer)(nil)

func NewRenderer(input []byte, style *TermStyle) *TermRenderer {
	// Get the theme once during renderer creation
	theme := GetTheme(style.Theme)
	if theme == nil {
		theme = GetTheme("Dracula") // Default to Dracula theme
	}

	return &TermRenderer{
		Output: input,
		Style:  style,
		theme:  theme,
	}
}

func DefaultRenderer(input []byte) *TermRenderer {
	font, err := fonts.GetFallback(fonts.FallbackMono)
	if err != nil {
		panic(err)
	}

	return NewRenderer(input, &TermStyle{
		Theme:         "Dracula",
		Font:          font,
		FontSize:      14,
		LineHeight:    1.25,
		PaddingLeft:   1,
		PaddingRight:  1,
		PaddingTop:    1,
		PaddingBottom: 1,
		Width:         120,
		Height:        40,
		AutoSize:      false,
		ShowPrompt:    false,
		PromptFunc:    func(command string) string { return fmt.Sprintf("❯ %s", command) },
	})
}

func (r *TermRenderer) WithTheme(theme string) *TermRenderer {
	r.Style.Theme = theme
	// Update the actual theme instance
	newTheme := GetTheme(theme)
	if newTheme == nil {
		newTheme = GetTheme("Dracula") // Default to Dracula theme
	}
	r.theme = newTheme
	return r
}

func (r *TermRenderer) WithFont(font *fonts.Font) *TermRenderer {
	r.Style.Font = font
	return r
}

func (r *TermRenderer) WithFontName(name string, style *fonts.FontStyle) *TermRenderer {
	font, err := fonts.GetFont(name, style)
	if err != nil {
		panic(err)
	}
	return r.WithFont(font)
}

func (r *TermRenderer) WithFontSize(size float64) *TermRenderer {
	r.Style.FontSize = size
	return r
}

func (r *TermRenderer) WithLineHeight(height float64) *TermRenderer {
	r.Style.LineHeight = height
	return r
}

func (r *TermRenderer) WithPadding(left, right, top, bottom int) *TermRenderer {
	r.Style.PaddingLeft = left
	r.Style.PaddingRight = right
	r.Style.PaddingTop = top
	r.Style.PaddingBottom = bottom
	return r
}

func (r *TermRenderer) WithWidth(width int) *TermRenderer {
	r.Style.Width = width
	return r
}

func (r *TermRenderer) WithHeight(height int) *TermRenderer {
	r.Style.Height = height
	return r
}

func (r *TermRenderer) WithTerminalSize(fd uintptr) *TermRenderer {
	width, height, _ := term.GetSize(fd)
	r.Style.Width = width
	r.Style.Height = height
	return r
}

func (r *TermRenderer) WithAutoSize() *TermRenderer {
	if r.Style == nil {
		r.Style = &TermStyle{}
	}
	r.Style.AutoSize = true
	return r
}

func (r *TermRenderer) WithShowPrompt() *TermRenderer {
	if r.Style == nil {
		r.Style = &TermStyle{}
	}
	r.Style.ShowPrompt = true
	return r
}

func (r *TermRenderer) WithPromptFunc(promptFunc func(command string) string) *TermRenderer {
	if r.Style == nil {
		r.Style = &TermStyle{}
	}
	r.Style.PromptFunc = promptFunc
	return r
}

func (r *TermRenderer) WithArgs(args []string) *TermRenderer {
	if r.Style == nil {
		r.Style = &TermStyle{}
	}
	r.Style.Args = args
	return r
}

// Render implements the content.Content interface
func (r *TermRenderer) Render() (image.Image, error) {
	// Create a new terminal with the current style
	t := NewTerminal(r.Style, r.theme)
	in := r.Output

	// Add prompt if needed
	if r.Style.ShowPrompt && r.Style.PromptFunc != nil && len(r.Style.Args) > 0 {
		// Join args into a command string
		cmd := strings.Join(r.Style.Args, " ")
		// Generate prompt text
		promptText := r.Style.PromptFunc(cmd)
		// Convert to bytes and prepend to input
		promptBytes := []byte(promptText + "\n")
		in = append(promptBytes, in...)
	}

	// Use ANSIParser to handle ANSI sequences
	parser := NewANSIParser(t)
	parser.Parse(in)

	// Calculate final dimensions
	width := t.Width
	height := t.Height
	if width == 0 || r.Style.AutoSize {
		width = t.MaxX + t.PaddingRight // Add right padding
	}
	if height == 0 || r.Style.AutoSize {
		height = t.MaxY + t.PaddingBottom // Add bottom padding
	}

	// Create font face using the base font's style
	face, err := r.Style.Font.GetFace(r.Style.FontSize, &r.Style.Font.Style)
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}
	defer face.Close()

	// Create a map to cache font faces for different styles
	fontFaces := make(map[Attributes]*fonts.Face)
	defer func() {
		// Close all font faces when done
		for _, f := range fontFaces {
			f.Close()
		}
	}()

	// Helper function to get or create a font face for a style
	getFontFace := func(attrs Attributes) (*fonts.Face, error) {
		if face, ok := fontFaces[attrs]; ok {
			return face, nil
		}

		// Determine font style based on attributes
		style := &fonts.FontStyle{
			Weight:  fonts.WeightRegular,
			Stretch: fonts.StretchNormal,
		}

		// Set font weight
		if attrs.Bold {
			style.Weight = fonts.WeightBold
		}

		// Set font style
		if attrs.Italic {
			style.Italic = true
		}

		face, err := r.Style.Font.GetFace(r.Style.FontSize, style)
		if err != nil {
			// Try fallback if the exact style is not available
			style.Weight = fonts.WeightRegular
			style.Italic = false
			face, err = r.Style.Font.GetFace(r.Style.FontSize, style)
			if err != nil {
				return nil, fmt.Errorf("failed to create font face: %v", err)
			}
		}
		fontFaces[attrs] = face
		return face, nil
	}

	// Measure the character width using the font metrics
	charWidthI26, _ := face.Face.GlyphAdvance('M')
	charWidth := charWidthI26.Round() // Convert to int

	// Create the image with correct dimensions based on character width
	bounds := image.Rect(0, 0,
		width*charWidth+r.Style.PaddingLeft+r.Style.PaddingRight+width*r.Style.CellSpacing,
		height*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop+r.Style.PaddingBottom)
	img := image.NewRGBA(bounds)

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{t.DefaultBg}, image.Point{}, draw.Src)

	// Draw cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if y >= len(t.Cells) || x >= len(t.Cells[y]) {
				continue
			}

			cell := t.Cells[y][x]

			// Draw background if different from default
			if cell.BgColor != t.DefaultBg {
				cellRect := image.Rect(
					x*charWidth+r.Style.PaddingLeft+r.Style.CellSpacing*x,
					y*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop,
					(x+1)*charWidth+r.Style.PaddingLeft+r.Style.CellSpacing*x,
					(y+1)*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop,
				)
				draw.Draw(img, cellRect, &image.Uniform{cell.BgColor}, image.Point{}, draw.Src)
			}

			if cell.Char == 0 || cell.Char == ' ' {
				continue
			}

			// Draw the character
			point := fixed.Point26_6{
				X: fixed.Int26_6(x*charWidth+r.Style.PaddingLeft+r.Style.CellSpacing*x) << 6,
				Y: fixed.Int26_6(y*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop+int(r.Style.FontSize)) << 6,
			}

			// Get the appropriate font face for this cell's attributes
			cellFace, err := getFontFace(cell.Attrs)
			if err != nil {
				return nil, fmt.Errorf("failed to get font face for cell at (%d,%d): %v", x, y, err)
			}

			d := &font.Drawer{
				Dst:  img,
				Src:  &image.Uniform{cell.FgColor},
				Face: cellFace.Face,
				Dot:  point,
			}
			d.DrawString(string(cell.Char))
		}
	}

	return img, nil
}
