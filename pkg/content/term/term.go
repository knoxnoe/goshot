package term

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"strconv"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/watzon/goshot/pkg/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type TermStyle struct {
	Args          []string                    // Command and arguments
	Theme         string                      // The terminal theme to use
	Font          *fonts.Font                 // The font to use
	FontSize      float64                     // The font size in points
	LineHeight    float64                     // The line height multiplier
	PaddingLeft   int                         // Padding between the code and the left edge
	PaddingRight  int                         // Padding between the code and the right edge
	PaddingTop    int                         // Padding between the code and the top edge
	PaddingBottom int                         // Padding between the code and the bottom edge
	Width         int                         // Terminal width in cells
	Height        int                         // Terminal height in cells
	AutoSize      bool                        // Whether to automatically size the output to the content
	CellSpacing   int                         // Additional horizontal spacing between cells
	ShowPrompt    bool                        // Whether to show a prompt
	PromptFunc    func(command string) string // Template function that returns the prompt text
}

type TermRenderer struct {
	Output []byte
	Style  *TermStyle
	theme  *Theme // Store theme here
}

type Attributes struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Blink         bool
}

type Cell struct {
	Char    rune
	FgColor color.Color
	BgColor color.Color
	Attrs   Attributes
	IsWide  bool // For handling wide characters
}

type Terminal struct {
	Cells         [][]Cell
	Width         int
	Height        int
	CursorX       int
	CursorY       int
	CurrAttrs     Attributes
	CurrFg        color.Color
	CurrBg        color.Color
	Style         *Theme // Theme colors from theme
	MaxX          int    // For dynamic sizing
	MaxY          int    // For dynamic sizing
	DefaultFg     color.Color
	DefaultBg     color.Color
	AutoSize      bool // Whether to automatically size the terminal
	PaddingLeft   int
	PaddingRight  int
	PaddingTop    int
	PaddingBottom int
}

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

func (t *Terminal) Reset() {
	t.CursorX = t.PaddingLeft
	t.CursorY = t.PaddingTop
	t.CurrAttrs = Attributes{}
	t.CurrFg = t.DefaultFg
	t.CurrBg = t.DefaultBg
}

func (t *Terminal) Resize(width, height int) {
	// Add padding to dimensions
	totalWidth := width + t.PaddingLeft + t.PaddingRight
	totalHeight := height + t.PaddingTop + t.PaddingBottom

	newCells := make([][]Cell, totalHeight)
	for i := range newCells {
		newCells[i] = make([]Cell, totalWidth)
		// Initialize with default colors and empty runes
		for j := range newCells[i] {
			newCells[i][j] = Cell{
				Char:    ' ',
				FgColor: t.DefaultFg,
				BgColor: t.DefaultBg,
			}
		}
	}

	// Copy existing content, accounting for padding
	for y := 0; y < min(len(t.Cells), totalHeight); y++ {
		for x := 0; x < min(len(t.Cells[y]), totalWidth); x++ {
			newCells[y][x] = t.Cells[y][x]
		}
	}

	t.Cells = newCells
	t.Width = totalWidth
	t.Height = totalHeight
}

func (t *Terminal) SetCell(x, y int, ch rune) {
	// Ignore any attempts to set cells with negative coordinates
	// or before padding
	if x < t.PaddingLeft || y < t.PaddingTop {
		return
	}

	// If auto-sizing is enabled, track the maximum dimensions regardless of Width/Height
	if t.AutoSize {
		t.MaxX = max(t.MaxX, x+1)
		t.MaxY = max(t.MaxY, y+1)
		// Resize if needed
		if y >= len(t.Cells) || x >= len(t.Cells[0]) {
			t.Resize(max(t.Width, x+1), max(t.Height, y+1))
		}
	} else {
		// If not auto-sizing, respect the terminal boundaries including padding
		if x >= t.Width || y >= t.Height {
			return
		}
	}

	// Ensure we have enough rows
	for len(t.Cells) <= y {
		t.Cells = append(t.Cells, make([]Cell, t.Width))
	}

	// Ensure the row has enough columns
	if len(t.Cells[y]) <= x {
		newRow := make([]Cell, t.Width)
		copy(newRow, t.Cells[y])
		t.Cells[y] = newRow
	}

	t.Cells[y][x] = Cell{
		Char:    ch,
		FgColor: t.CurrFg,
		BgColor: t.CurrBg,
		Attrs:   t.CurrAttrs,
	}
}

func (t *Terminal) NewLine() {
	t.CursorX = t.PaddingLeft
	t.CursorY++
	// If we're at the top padding, skip to the content area
	if t.CursorY < t.PaddingTop {
		t.CursorY = t.PaddingTop
	}
	if t.Height > 0 && t.CursorY >= t.Height {
		// Scroll up, preserving padding area
		copy(t.Cells[t.PaddingTop:], t.Cells[t.PaddingTop+1:])
		t.CursorY = t.Height - 1
		// Clear the new line
		for x := range t.Cells[t.CursorY] {
			t.Cells[t.CursorY][x] = Cell{
				Char:    ' ',
				FgColor: t.DefaultFg,
				BgColor: t.DefaultBg,
			}
		}
	}
}

func NewTerminal(style *TermStyle, theme *Theme) *Terminal {
	t := &Terminal{
		Width:         style.Width,
		Height:        style.Height,
		AutoSize:      style.AutoSize,
		PaddingLeft:   style.PaddingLeft,
		PaddingRight:  style.PaddingRight,
		PaddingTop:    style.PaddingTop,
		PaddingBottom: style.PaddingBottom,
		DefaultFg:     theme.GetForeground(),
		DefaultBg:     theme.GetBackground(),
		CurrFg:        theme.GetForeground(),
		CurrBg:        theme.GetBackground(),
		Style:         theme,
		// Initialize cursor position at the start of the content area
		CursorX: style.PaddingLeft,
		CursorY: style.PaddingTop,
	}

	// Initialize cells
	t.Resize(style.Width, style.Height)
	return t
}

func getPrefix(seq []byte) string {
	switch {
	case ansi.HasCsiPrefix(seq):
		return "CSI"
	case ansi.HasOscPrefix(seq):
		return "OSC"
	case ansi.HasDcsPrefix(seq):
		return "DCS"
	case ansi.HasApcPrefix(seq):
		return "APC"
	default:
		return ""
	}
}

func (r *TermRenderer) Render() (image.Image, error) {
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)

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

	for len(in) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(in, state, p)
		if n == 0 {
			// If we can't decode the sequence, skip one byte
			in = in[1:]
			continue
		}

		if width > 0 {
			// This is a character
			// Convert the entire sequence to a rune
			r := []rune(string(seq))[0]
			t.SetCell(t.CursorX, t.CursorY, r)
			t.CursorX++
			if t.Width > 0 && t.CursorX >= t.Width {
				t.NewLine()
			}
		} else {
			// This is a control sequence
			prefix := getPrefix(seq)
			s := string(seq)

			if s == "\n" {
				t.NewLine()
			} else if s == "\r" {
				t.CursorX = 0
			} else {
				switch prefix {
				case "CSI":
					// Handle CSI sequences
					if strings.HasSuffix(s, "m") {
						// SGR (Select Graphic Rendition)
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "m")
						paramSlice := strings.Split(params, ";")
						for i := 0; i < len(paramSlice); i++ {
							code, _ := strconv.Atoi(paramSlice[i])
							switch {
							case code == 0:
								t.CurrAttrs = Attributes{}
								t.CurrFg = t.DefaultFg
								t.CurrBg = t.DefaultBg
							case code == 1:
								t.CurrAttrs.Bold = true
							case code == 3:
								t.CurrAttrs.Italic = true
							case code == 4:
								t.CurrAttrs.Underline = true
							case code == 5:
								t.CurrAttrs.Blink = true
							case code == 7:
								// Reverse video (swap fg and bg)
								t.CurrFg, t.CurrBg = t.CurrBg, t.CurrFg
							case code == 9:
								t.CurrAttrs.Strikethrough = true
							case code == 22:
								t.CurrAttrs.Bold = false
							case code == 23:
								t.CurrAttrs.Italic = false
							case code == 24:
								t.CurrAttrs.Underline = false
							case code == 25:
								t.CurrAttrs.Blink = false
							case code == 27:
								// Reset reverse video
								t.CurrFg, t.CurrBg = t.DefaultFg, t.DefaultBg
							case code == 29:
								t.CurrAttrs.Strikethrough = false
							case code >= 30 && code <= 37:
								// Standard foreground colors (30-37)
								t.CurrFg = ansiColor(code-30, r.theme)
							case code >= 40 && code <= 47:
								// Standard background colors (40-47)
								t.CurrBg = ansiColor(code-40, r.theme)
							case code >= 90 && code <= 97:
								// Bright foreground colors (90-97)
								t.CurrFg = ansiBrightColor(code-90, r.theme)
							case code >= 100 && code <= 107:
								// Bright background colors (100-107)
								t.CurrBg = ansiBrightColor(code-100, r.theme)
							case code == 38:
								// Extended foreground color
								if i+4 < len(paramSlice) && paramSlice[i+1] == "2" {
									// RGB (38;2;r;g;b)
									r, _ := strconv.Atoi(paramSlice[i+2])
									g, _ := strconv.Atoi(paramSlice[i+3])
									b, _ := strconv.Atoi(paramSlice[i+4])
									t.CurrFg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
									i += 4
								} else if i+2 < len(paramSlice) && paramSlice[i+1] == "5" {
									// 256 color (38;5;n)
									if i+2 < len(paramSlice) {
										colorNum, _ := strconv.Atoi(paramSlice[i+2])
										if colorNum < 8 {
											// Standard colors (0-7)
											t.CurrFg = ansiColor(colorNum, r.theme)
										} else if colorNum < 16 {
											// Bright colors (8-15)
											t.CurrFg = ansiBrightColor(colorNum-8, r.theme)
										} else if colorNum < 232 {
											// 216 color cube (16-231): 6x6x6
											colorNum -= 16
											b := colorNum % 6
											colorNum /= 6
											g := colorNum % 6
											r := colorNum / 6
											t.CurrFg = color.RGBA{
												uint8(r * 42),
												uint8(g * 42),
												uint8(b * 42),
												255,
											}
										} else {
											// Grayscale (232-255): 24 shades
											gray := uint8((colorNum-232)*10 + 8)
											t.CurrFg = color.RGBA{gray, gray, gray, 255}
										}
										i += 2
									}
								}
							case code == 48:
								// Extended background color
								if i+4 < len(paramSlice) && paramSlice[i+1] == "2" {
									// RGB (48;2;r;g;b)
									r, _ := strconv.Atoi(paramSlice[i+2])
									g, _ := strconv.Atoi(paramSlice[i+3])
									b, _ := strconv.Atoi(paramSlice[i+4])
									t.CurrBg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
									i += 4
								} else if i+2 < len(paramSlice) && paramSlice[i+1] == "5" {
									// 256 color (48;5;n)
									if i+2 < len(paramSlice) {
										colorNum, _ := strconv.Atoi(paramSlice[i+2])
										if colorNum < 8 {
											// Standard colors (0-7)
											t.CurrBg = ansiColor(colorNum, r.theme)
										} else if colorNum < 16 {
											// Bright colors (8-15)
											t.CurrBg = ansiBrightColor(colorNum-8, r.theme)
										} else if colorNum < 232 {
											// 216 color cube (16-231): 6x6x6
											colorNum -= 16
											b := colorNum % 6
											colorNum /= 6
											g := colorNum % 6
											r := colorNum / 6
											t.CurrBg = color.RGBA{
												uint8(r * 42),
												uint8(g * 42),
												uint8(b * 42),
												255,
											}
										} else {
											// Grayscale (232-255): 24 shades
											gray := uint8((colorNum-232)*10 + 8)
											t.CurrBg = color.RGBA{gray, gray, gray, 255}
										}
										i += 2
									}
								}
							case code == 39:
								// Default foreground color
								t.CurrFg = t.DefaultFg
							case code == 49:
								// Default background color
								t.CurrBg = t.DefaultBg
							}
						}
					} else if strings.HasSuffix(s, "G") {
						// Cursor Horizontal Absolute (CHA)
						// Move cursor to specific column
						n := 1 // Default to column 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "G")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						// Convert from 1-based to 0-based indexing
						n = max(1, n) // Ensure n is at least 1
						t.CursorX = min(t.Width-t.PaddingRight-1, t.PaddingLeft+n-1)
					} else if strings.HasSuffix(s, "H") || strings.HasSuffix(s, "f") {
						// Cursor position
						row, col := 1, 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), string(s[len(s)-1]))
						if params != "" {
							parts := strings.Split(params, ";")
							if len(parts) >= 1 {
								row, _ = strconv.Atoi(parts[0])
							}
							if len(parts) >= 2 {
								col, _ = strconv.Atoi(parts[1])
							}
						}
						// Adjust for padding and 1-based indexing
						t.CursorY = min(t.Height-1, max(1, row))
						t.CursorX = min(t.Width-t.PaddingRight-1, max(t.PaddingLeft, col-1+t.PaddingLeft))
					} else if strings.HasSuffix(s, "A") {
						// Cursor up
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "A")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						t.CursorY = max(1, t.CursorY-n) // Keep minimum at 1 to preserve first row
					} else if strings.HasSuffix(s, "B") {
						// Cursor down
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "B")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						t.CursorY = min(t.Height-1, t.CursorY+n)
					} else if strings.HasSuffix(s, "C") {
						// Cursor forward
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "C")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						t.CursorX = min(t.Width-t.PaddingRight-1, t.CursorX+n)
					} else if strings.HasSuffix(s, "D") {
						// Cursor backward
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "D")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						t.CursorX = max(t.PaddingLeft, t.CursorX-n)
					} else if strings.HasSuffix(s, "K") {
						// Erase in line
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "K")
						n := 0
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						switch n {
						case 0: // Clear from cursor to end of line
							for x := t.CursorX; x < len(t.Cells[t.CursorY]); x++ {
								t.Cells[t.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: t.DefaultFg,
									BgColor: t.DefaultBg,
								}
							}
						case 1: // Clear from cursor to start of line
							for x := 0; x <= t.CursorX; x++ {
								t.Cells[t.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: t.DefaultFg,
									BgColor: t.DefaultBg,
								}
							}
						case 2: // Clear entire line
							for x := range t.Cells[t.CursorY] {
								t.Cells[t.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: t.DefaultFg,
									BgColor: t.DefaultBg,
								}
							}
						}
					}
				case "OSC":
					// Just ignore OSC sequences for now
				case "DCS":
					// Just ignore DCS sequences for now
				}
			}
		}

		in = in[n:]
		state = newState
	}

	// Print final dimensions

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

func ansiColor(code int, theme *Theme) color.Color {
	return theme.GetColor(code)
}

func ansiBrightColor(code int, theme *Theme) color.Color {
	return theme.GetColor(code + 8) // Bright colors start at index 8
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
