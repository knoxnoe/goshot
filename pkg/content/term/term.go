package term

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"strconv"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/watzon/goshot/pkg/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type TermStyle struct {
	Theme         string      // The chroma syntax theme to use
	Font          *fonts.Font // The font to use
	FontSize      float64     // The font size in points
	LineHeight    float64     // The line height multiplier
	PaddingLeft   int         // Padding between the code and the left edge
	PaddingRight  int         // Padding between the code and the right edge
	PaddingTop    int         // Padding between the code and the top edge
	PaddingBottom int         // Padding between the code and the bottom edge
	Width         int         // Terminal width in cells
	Height        int         // Terminal height in cells
	AutoSize      bool        // Whether to automatically size the output to the content
}

type TermRenderer struct {
	Output []byte
	Style  *TermStyle
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
	Cells     [][]Cell
	Width     int
	Height    int
	CursorX   int
	CursorY   int
	CurrAttrs Attributes
	CurrFg    color.Color
	CurrBg    color.Color
	Style     *chroma.Style // Theme colors from chroma
	MaxX      int           // For dynamic sizing
	MaxY      int           // For dynamic sizing
	DefaultFg color.Color
	DefaultBg color.Color
}

func NewRenderer(input []byte, style *TermStyle) *TermRenderer {
	return &TermRenderer{
		Output: input,
		Style:  style,
	}
}

func DefaultRenderer(input []byte) *TermRenderer {
	font, err := fonts.GetFallback(fonts.FallbackMono)
	if err != nil {
		panic(err)
	}

	return NewRenderer(input, &TermStyle{
		Theme:         "monokai",
		Font:          font,
		FontSize:      14,
		LineHeight:    1.2,
		PaddingLeft:   20,
		PaddingRight:  20,
		PaddingTop:    20,
		PaddingBottom: 20,
		Width:         0,
		Height:        0,
	})
}

func (r *TermRenderer) WithTheme(theme string) *TermRenderer {
	r.Style.Theme = theme
	return r
}

func (r *TermRenderer) WithFont(font *fonts.Font) *TermRenderer {
	r.Style.Font = font
	return r
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

func (t *Terminal) Reset() {
	t.CursorX = 0
	t.CursorY = 0
	t.CurrAttrs = Attributes{}
	t.CurrFg = t.DefaultFg
	t.CurrBg = t.DefaultBg
}

func (t *Terminal) Resize(width, height int) {
	newCells := make([][]Cell, height)
	for i := range newCells {
		newCells[i] = make([]Cell, width)
		// Initialize with default colors and empty runes
		for j := range newCells[i] {
			newCells[i][j] = Cell{
				Char:    ' ',
				FgColor: t.DefaultFg,
				BgColor: t.DefaultBg,
			}
		}
	}

	// Copy existing content
	for y := 0; y < min(len(t.Cells), height); y++ {
		for x := 0; x < min(len(t.Cells[y]), width); x++ {
			newCells[y][x] = t.Cells[y][x]
		}
	}

	t.Cells = newCells
	t.Width = width
	t.Height = height
}

func (t *Terminal) SetCell(x, y int, ch rune) {
	// Ignore any attempts to set cells with negative coordinates
	if x < 0 || y < 0 {
		return
	}

	// Ensure we have enough rows
	for len(t.Cells) <= y {
		if t.Height > 0 && y >= t.Height {
			return
		}
		t.Cells = append(t.Cells, make([]Cell, max(t.Width, x+1)))
	}

	// Ensure the row has enough columns
	if len(t.Cells[y]) <= x {
		if t.Width > 0 && x >= t.Width {
			return
		}
		newRow := make([]Cell, max(t.Width, x+1))
		copy(newRow, t.Cells[y])
		t.Cells[y] = newRow
	}

	t.Cells[y][x] = Cell{
		Char:    ch,
		FgColor: t.CurrFg,
		BgColor: t.CurrBg,
		Attrs:   t.CurrAttrs,
	}

	t.MaxX = max(t.MaxX, x+1)
	t.MaxY = max(t.MaxY, y+1)
}

func (t *Terminal) NewLine() {
	t.CursorX = 0
	t.CursorY++
	if t.Height > 0 && t.CursorY >= t.Height {
		// Scroll up
		copy(t.Cells[0:], t.Cells[1:])
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

func NewTerminal(style *TermStyle) *Terminal {
	// Get theme colors from chroma
	chromaStyle := styles.Get(style.Theme)
	if chromaStyle == nil {
		chromaStyle = styles.Fallback
	}

	// Get default colors from the theme
	defaultBg := chromaStyle.Get(chroma.Background).Background
	defaultFg := chromaStyle.Get(chroma.Text).Colour

	term := &Terminal{
		Style: chromaStyle,
		DefaultFg: color.RGBA{
			R: defaultFg.Red(),
			G: defaultFg.Green(),
			B: defaultFg.Blue(),
			A: 255,
		},
		DefaultBg: color.RGBA{
			R: defaultBg.Red(),
			G: defaultBg.Green(),
			B: defaultBg.Blue(),
			A: 255,
		},
	}

	if style.Width > 0 && style.Height > 0 {
		term.Resize(style.Width, style.Height)
	}

	term.Reset()
	return term
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

	term := NewTerminal(r.Style)
	in := r.Output

	fmt.Printf("Input bytes: %v\n", in)

	for len(in) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(in, state, p)
		if n == 0 {
			// If we can't decode the sequence, skip one byte
			fmt.Printf("Failed to decode sequence, skipping byte: %x\n", in[0])
			in = in[1:]
			continue
		}

		if width > 0 {
			// This is a character
			fmt.Printf("Character: %q (hex: %x) width: %d\n", seq, seq, width)
			term.SetCell(term.CursorX, term.CursorY, rune(seq[0]))
			fmt.Printf("Set cell at (%d,%d) to %q, MaxX: %d, MaxY: %d\n",
				term.CursorX, term.CursorY, rune(seq[0]), term.MaxX, term.MaxY)
			term.CursorX++
			if term.Width > 0 && term.CursorX >= term.Width {
				term.NewLine()
			}
		} else {
			// This is a control sequence
			prefix := getPrefix(seq)
			s := string(seq)
			fmt.Printf("Control sequence: %q prefix: %s\n", s, prefix)

			if s == "\n" {
				fmt.Printf("Newline detected! CursorY: %d -> %d\n", term.CursorY, term.CursorY+1)
				term.NewLine()
			} else if s == "\r" {
				fmt.Printf("Carriage return! CursorX: %d -> 0\n", term.CursorX)
				term.CursorX = 0
			} else {
				switch prefix {
				case "CSI":
					// Handle CSI sequences
					if strings.HasSuffix(s, "m") {
						// SGR (Select Graphic Rendition)
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "m")
						fmt.Printf("SGR params: %s\n", params)
						for i, param := range strings.Split(params, ";") {
							code, _ := strconv.Atoi(param)
							switch {
							case code == 0:
								term.CurrAttrs = Attributes{}
								term.CurrFg = term.DefaultFg
								term.CurrBg = term.DefaultBg
								fmt.Printf("Reset attributes\n")
							case code == 1:
								term.CurrAttrs.Bold = true
								fmt.Printf("Set bold\n")
							case code == 3:
								term.CurrAttrs.Italic = true
								fmt.Printf("Set italic\n")
							case code == 4:
								term.CurrAttrs.Underline = true
								fmt.Printf("Set underline\n")
							case code == 5:
								term.CurrAttrs.Blink = true
								fmt.Printf("Set blink\n")
							case code == 7:
								// Reverse video (swap fg and bg)
								term.CurrFg, term.CurrBg = term.CurrBg, term.CurrFg
								fmt.Printf("Reverse video\n")
							case code == 9:
								term.CurrAttrs.Strikethrough = true
								fmt.Printf("Set strikethrough\n")
							case code == 22:
								term.CurrAttrs.Bold = false
								fmt.Printf("Reset bold\n")
							case code == 23:
								term.CurrAttrs.Italic = false
								fmt.Printf("Reset italic\n")
							case code == 24:
								term.CurrAttrs.Underline = false
								fmt.Printf("Reset underline\n")
							case code == 25:
								term.CurrAttrs.Blink = false
								fmt.Printf("Reset blink\n")
							case code == 27:
								// Reset reverse video
								term.CurrFg, term.CurrBg = term.DefaultFg, term.DefaultBg
								fmt.Printf("Reset reverse video\n")
							case code == 29:
								term.CurrAttrs.Strikethrough = false
								fmt.Printf("Reset strikethrough\n")
							case code >= 30 && code <= 37:
								// Foreground color
								term.CurrFg = ansiColor(code-30, r.Style)
								fmt.Printf("Set fg color: %d\n", code-30)
							case code == 38:
								// 24-bit RGB foreground color
								paramSlice := strings.Split(params, ";")
								if i < len(paramSlice)-4 && paramSlice[i+1] == "2" {
									r, _ := strconv.Atoi(paramSlice[i+2])
									g, _ := strconv.Atoi(paramSlice[i+3])
									b, _ := strconv.Atoi(paramSlice[i+4])
									term.CurrFg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
									fmt.Printf("Set RGB fg color: %d,%d,%d\n", r, g, b)
									i += 4 // Skip the color components we just processed
									continue
								}
							case code == 39:
								// Default foreground color
								term.CurrFg = term.DefaultFg
								fmt.Printf("Reset fg color\n")
							case code >= 40 && code <= 47:
								// Background color
								term.CurrBg = ansiColor(code-40, r.Style)
								fmt.Printf("Set bg color: %d\n", code-40)
							case code == 48:
								// 24-bit RGB background color
								paramSlice := strings.Split(params, ";")
								if i < len(paramSlice)-4 && paramSlice[i+1] == "2" {
									r, _ := strconv.Atoi(paramSlice[i+2])
									g, _ := strconv.Atoi(paramSlice[i+3])
									b, _ := strconv.Atoi(paramSlice[i+4])
									term.CurrBg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
									fmt.Printf("Set RGB bg color: %d,%d,%d\n", r, g, b)
									i += 4 // Skip the color components we just processed
									continue
								}
							case code == 49:
								// Default background color
								term.CurrBg = term.DefaultBg
								fmt.Printf("Reset bg color\n")
							case code >= 90 && code <= 97:
								// Bright foreground color
								term.CurrFg = ansiBrightColor(code-90, r.Style)
								fmt.Printf("Set bright fg color: %d\n", code-90)
							case code >= 100 && code <= 107:
								// Bright background color
								term.CurrBg = ansiBrightColor(code-100, r.Style)
								fmt.Printf("Set bright bg color: %d\n", code-100)
							}
						}
					} else if strings.HasSuffix(s, "H") || strings.HasSuffix(s, "f") {
						// Cursor position
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "H")
						if params == "" {
							term.CursorX = 0
							term.CursorY = 0
							fmt.Printf("Reset cursor position to (0,0)\n")
						} else {
							parts := strings.Split(params, ";")
							if len(parts) == 2 {
								row, _ := strconv.Atoi(parts[0])
								col, _ := strconv.Atoi(parts[1])
								term.CursorY = max(0, row-1)
								term.CursorX = max(0, col-1)
								fmt.Printf("Set cursor position to (%d,%d)\n", term.CursorX, term.CursorY)
							}
						}
					} else if strings.HasSuffix(s, "A") {
						// Cursor up
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "A")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						term.CursorY = max(0, term.CursorY-n)
						fmt.Printf("Cursor up %d to Y=%d\n", n, term.CursorY)
					} else if strings.HasSuffix(s, "B") {
						// Cursor down
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "B")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						term.CursorY = min(term.Height-1, term.CursorY+n)
						fmt.Printf("Cursor down %d to Y=%d\n", n, term.CursorY)
					} else if strings.HasSuffix(s, "C") {
						// Cursor forward
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "C")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						term.CursorX = min(term.Width-1, term.CursorX+n)
						fmt.Printf("Cursor forward %d to X=%d\n", n, term.CursorX)
					} else if strings.HasSuffix(s, "D") {
						// Cursor backward
						n := 1
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "D")
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						term.CursorX = max(0, term.CursorX-n)
						fmt.Printf("Cursor backward %d to X=%d\n", n, term.CursorX)
					} else if strings.HasSuffix(s, "K") {
						// Erase in line
						params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "K")
						n := 0
						if params != "" {
							n, _ = strconv.Atoi(params)
						}
						switch n {
						case 0: // Clear from cursor to end of line
							for x := term.CursorX; x < len(term.Cells[term.CursorY]); x++ {
								term.Cells[term.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: term.DefaultFg,
									BgColor: term.DefaultBg,
								}
							}
							fmt.Printf("Clear from cursor to end of line\n")
						case 1: // Clear from cursor to start of line
							for x := 0; x <= term.CursorX; x++ {
								term.Cells[term.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: term.DefaultFg,
									BgColor: term.DefaultBg,
								}
							}
							fmt.Printf("Clear from cursor to start of line\n")
						case 2: // Clear entire line
							for x := range term.Cells[term.CursorY] {
								term.Cells[term.CursorY][x] = Cell{
									Char:    ' ',
									FgColor: term.DefaultFg,
									BgColor: term.DefaultBg,
								}
							}
							fmt.Printf("Clear entire line\n")
						}
					}
				case "OSC":
					// Just ignore OSC sequences for now
					fmt.Printf("Ignoring OSC sequence: %q\n", s)
				case "DCS":
					// Just ignore DCS sequences for now
					fmt.Printf("Ignoring DCS sequence: %q\n", s)
				}
			}
		}

		in = in[n:]
		state = newState
	}

	// Print final dimensions
	fmt.Printf("Final dimensions - MaxX: %d, MaxY: %d\n", term.MaxX, term.MaxY)
	fmt.Printf("Grid contents:\n")
	for y := 0; y < term.MaxY; y++ {
		fmt.Printf("Row %d: ", y)
		for x := 0; x < term.MaxX; x++ {
			if y < len(term.Cells) && x < len(term.Cells[y]) {
				fmt.Printf("%q ", term.Cells[y][x].Char)
			} else {
				fmt.Printf("  ")
			}
		}
		fmt.Printf("\n")
	}

	// Calculate final dimensions
	width := term.Width
	height := term.Height
	if width == 0 || r.Style.AutoSize {
		width = term.MaxX
	}
	if height == 0 || r.Style.AutoSize {
		height = term.MaxY
	}

	// Create the image
	bounds := image.Rect(0, 0,
		width*int(r.Style.FontSize)+r.Style.PaddingLeft+r.Style.PaddingRight,
		height*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop+r.Style.PaddingBottom)
	img := image.NewRGBA(bounds)

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{term.DefaultBg}, image.Point{}, draw.Src)

	// Create font face
	face, err := r.Style.Font.GetFace(r.Style.FontSize, &fonts.FontStyle{
		Weight:  fonts.WeightRegular,
		Stretch: fonts.StretchNormal,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}
	defer face.Close()

	// Draw cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if y >= len(term.Cells) || x >= len(term.Cells[y]) {
				continue
			}

			cell := term.Cells[y][x]
			if cell.Char == 0 || cell.Char == ' ' {
				continue
			}

			// Draw background if different from default
			if cell.BgColor != term.DefaultBg {
				cellRect := image.Rect(
					x*int(r.Style.FontSize)+r.Style.PaddingLeft,
					y*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop,
					(x+1)*int(r.Style.FontSize)+r.Style.PaddingLeft,
					(y+1)*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop,
				)
				draw.Draw(img, cellRect, &image.Uniform{cell.BgColor}, image.Point{}, draw.Src)
			}

			// Draw text
			point := fixed.Point26_6{
				X: fixed.Int26_6(x*int(r.Style.FontSize)+r.Style.PaddingLeft) << 6,
				Y: fixed.Int26_6(y*int(float64(r.Style.FontSize)*r.Style.LineHeight)+r.Style.PaddingTop+int(r.Style.FontSize)) << 6,
			}
			d := &font.Drawer{
				Dst:  img,
				Src:  &image.Uniform{cell.FgColor},
				Face: face.Face,
				Dot:  point,
			}
			d.DrawString(string(cell.Char))
		}
	}

	return img, nil
}

func ansiColor(code int, style *TermStyle) color.Color {
	// Map ANSI colors to chroma theme colors where possible
	chromaStyle := styles.Get(style.Theme)
	if chromaStyle == nil {
		chromaStyle = styles.Fallback
	}

	switch code {
	case 0: // Black
		c := chromaStyle.Get(chroma.Background).Background
		return chromaColorToRGBA(c)
	case 1: // Red
		c := chromaStyle.Get(chroma.Error).Colour
		return chromaColorToRGBA(c)
	case 2: // Green
		c := chromaStyle.Get(chroma.String).Colour
		return chromaColorToRGBA(c)
	case 3: // Yellow
		c := chromaStyle.Get(chroma.Keyword).Colour
		return chromaColorToRGBA(c)
	case 4: // Blue
		c := chromaStyle.Get(chroma.NameFunction).Colour
		return chromaColorToRGBA(c)
	case 5: // Magenta
		c := chromaStyle.Get(chroma.Operator).Colour
		return chromaColorToRGBA(c)
	case 6: // Cyan
		c := chromaStyle.Get(chroma.Name).Colour
		return chromaColorToRGBA(c)
	case 7: // White
		c := chromaStyle.Get(chroma.Text).Colour
		return chromaColorToRGBA(c)
	}

	// Fallback to standard ANSI colors
	return color.RGBA{
		R: []uint8{0, 205, 0, 205, 0, 205, 205, 229}[code],
		G: []uint8{0, 0, 205, 205, 0, 0, 205, 229}[code],
		B: []uint8{0, 0, 0, 0, 205, 205, 205, 229}[code],
		A: 255,
	}
}

func ansiBrightColor(code int, style *TermStyle) color.Color {
	// Get the chroma style
	chromaStyle := styles.Get(style.Theme)
	if chromaStyle == nil {
		chromaStyle = styles.Fallback
	}
	
	// Map to chroma theme colors with increased brightness
	baseColor := ansiColor(code, style)
	if rgba, ok := baseColor.(color.RGBA); ok {
		// Increase brightness by 20%
		return color.RGBA{
			R: uint8(min(255, int(float64(rgba.R)*1.2))),
			G: uint8(min(255, int(float64(rgba.G)*1.2))),
			B: uint8(min(255, int(float64(rgba.B)*1.2))),
			A: rgba.A,
		}
	}
	
	// Fallback to standard bright colors
	return color.RGBA{
		R: []uint8{127, 255, 0, 255, 0, 255, 0, 255}[code],
		G: []uint8{127, 0, 255, 255, 0, 0, 255, 255}[code],
		B: []uint8{127, 0, 0, 0, 255, 255, 255, 255}[code],
		A: 255,
	}
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

func chromaColorToRGBA(c chroma.Colour) color.RGBA {
	return color.RGBA{
		R: c.Red(),
		G: c.Green(),
		B: c.Blue(),
		A: 255,
	}
}
