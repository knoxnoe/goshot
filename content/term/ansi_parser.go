package term

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

type ANSIParser struct {
	terminal *Terminal
	parser   *ansi.Parser
	state    byte
}

func NewANSIParser(t *Terminal) *ANSIParser {
	return &ANSIParser{
		terminal: t,
		parser:   ansi.GetParser(),
	}
}

func (ap *ANSIParser) Parse(input []byte) {
	defer ansi.PutParser(ap.parser)

	for len(input) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(input, ap.state, ap.parser)
		if n == 0 {
			input = input[1:]
			continue
		}

		if width > 0 {
			r := []rune(string(seq))[0]
			ap.terminal.SetCell(ap.terminal.CursorX, ap.terminal.CursorY, r)
			ap.terminal.CursorX++
			if ap.terminal.Width > 0 && ap.terminal.CursorX >= ap.terminal.Width {
				ap.terminal.NewLine()
			}
		} else {
			ap.handleControlSequence(seq)
		}

		input = input[n:]
		ap.state = newState
	}
}

func (ap *ANSIParser) handleControlSequence(seq []byte) {
	prefix := getPrefix(seq)
	s := string(seq)

	if s == "\n" {
		ap.terminal.NewLine()
	} else if s == "\r" {
		ap.terminal.CursorX = 0
	} else {
		switch prefix {
		case "CSI":
			ap.handleCSISequence(s)
		case "OSC":
			// Ignore OSC sequences
		case "DCS":
			// Ignore DCS sequences
		}
	}
}

func (ap *ANSIParser) handleCSISequence(s string) {
	if strings.HasSuffix(s, "m") {
		ap.handleSGR(s)
	} else if strings.HasSuffix(s, "G") {
		ap.handleCHA(s)
	} else if strings.HasSuffix(s, "H") || strings.HasSuffix(s, "f") {
		ap.handleCUP(s)
	} else if strings.HasSuffix(s, "A") {
		ap.handleCUU(s)
	} else if strings.HasSuffix(s, "B") {
		ap.handleCUD(s)
	} else if strings.HasSuffix(s, "C") {
		ap.handleCUF(s)
	} else if strings.HasSuffix(s, "D") {
		ap.handleCUB(s)
	} else if strings.HasSuffix(s, "K") {
		ap.handleEL(s)
	}
}

func (ap *ANSIParser) handleSGR(s string) {
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "m")
	paramSlice := strings.Split(params, ";")
	for i := 0; i < len(paramSlice); i++ {
		code, _ := strconv.Atoi(paramSlice[i])
		switch {
		case code == 0:
			ap.terminal.CurrAttrs = Attributes{}
			ap.terminal.CurrFg = ap.terminal.DefaultFg
			ap.terminal.CurrBg = ap.terminal.DefaultBg
		case code == 1:
			ap.terminal.CurrAttrs.Bold = true
		case code == 3:
			ap.terminal.CurrAttrs.Italic = true
		case code == 4:
			ap.terminal.CurrAttrs.Underline = true
		case code == 5:
			ap.terminal.CurrAttrs.Blink = true
		case code == 7:
			ap.terminal.CurrFg, ap.terminal.CurrBg = ap.terminal.CurrBg, ap.terminal.CurrFg
		case code == 9:
			ap.terminal.CurrAttrs.Strikethrough = true
		case code == 22:
			ap.terminal.CurrAttrs.Bold = false
		case code == 23:
			ap.terminal.CurrAttrs.Italic = false
		case code == 24:
			ap.terminal.CurrAttrs.Underline = false
		case code == 25:
			ap.terminal.CurrAttrs.Blink = false
		case code == 27:
			ap.terminal.CurrFg, ap.terminal.CurrBg = ap.terminal.DefaultFg, ap.terminal.DefaultBg
		case code == 29:
			ap.terminal.CurrAttrs.Strikethrough = false
		case code >= 30 && code <= 37:
			ap.terminal.CurrFg = ansiColor(code-30, ap.terminal.Style)
		case code >= 40 && code <= 47:
			ap.terminal.CurrBg = ansiColor(code-40, ap.terminal.Style)
		case code >= 90 && code <= 97:
			ap.terminal.CurrFg = ansiBrightColor(code-90, ap.terminal.Style)
		case code >= 100 && code <= 107:
			ap.terminal.CurrBg = ansiBrightColor(code-100, ap.terminal.Style)
		case code == 38:
			if i+4 < len(paramSlice) && paramSlice[i+1] == "2" {
				r, _ := strconv.Atoi(paramSlice[i+2])
				g, _ := strconv.Atoi(paramSlice[i+3])
				b, _ := strconv.Atoi(paramSlice[i+4])
				ap.terminal.CurrFg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
				i += 4
			} else if i+2 < len(paramSlice) && paramSlice[i+1] == "5" {
				colorNum, _ := strconv.Atoi(paramSlice[i+2])
				ap.terminal.CurrFg = ap.get256Color(colorNum)
				i += 2
			}
		case code == 48:
			if i+4 < len(paramSlice) && paramSlice[i+1] == "2" {
				r, _ := strconv.Atoi(paramSlice[i+2])
				g, _ := strconv.Atoi(paramSlice[i+3])
				b, _ := strconv.Atoi(paramSlice[i+4])
				ap.terminal.CurrBg = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
				i += 4
			} else if i+2 < len(paramSlice) && paramSlice[i+1] == "5" {
				colorNum, _ := strconv.Atoi(paramSlice[i+2])
				ap.terminal.CurrBg = ap.get256Color(colorNum)
				i += 2
			}
		case code == 39:
			ap.terminal.CurrFg = ap.terminal.DefaultFg
		case code == 49:
			ap.terminal.CurrBg = ap.terminal.DefaultBg
		}
	}
}

func (ap *ANSIParser) handleCHA(s string) {
	n := 1
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "G")
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	n = max(1, n)
	ap.terminal.CursorX = min(ap.terminal.Width-ap.terminal.PaddingRight-1, ap.terminal.PaddingLeft+n-1)
}

func (ap *ANSIParser) handleCUP(s string) {
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
	ap.terminal.CursorY = min(ap.terminal.Height-1, max(1, row))
	ap.terminal.CursorX = min(ap.terminal.Width-ap.terminal.PaddingRight-1, max(ap.terminal.PaddingLeft, col-1+ap.terminal.PaddingLeft))
}

func (ap *ANSIParser) handleCUU(s string) {
	n := 1
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "A")
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	ap.terminal.CursorY = max(1, ap.terminal.CursorY-n)
}

func (ap *ANSIParser) handleCUD(s string) {
	n := 1
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "B")
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	ap.terminal.CursorY = min(ap.terminal.Height-1, ap.terminal.CursorY+n)
}

func (ap *ANSIParser) handleCUF(s string) {
	n := 1
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "C")
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	ap.terminal.CursorX = min(ap.terminal.Width-ap.terminal.PaddingRight-1, ap.terminal.CursorX+n)
}

func (ap *ANSIParser) handleCUB(s string) {
	n := 1
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "D")
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	ap.terminal.CursorX = max(ap.terminal.PaddingLeft, ap.terminal.CursorX-n)
}

func (ap *ANSIParser) handleEL(s string) {
	params := strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), "K")
	n := 0
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	switch n {
	case 0:
		for x := ap.terminal.CursorX; x < len(ap.terminal.Cells[ap.terminal.CursorY]); x++ {
			ap.terminal.Cells[ap.terminal.CursorY][x] = Cell{
				Char:    ' ',
				FgColor: ap.terminal.DefaultFg,
				BgColor: ap.terminal.DefaultBg,
			}
		}
	case 1:
		for x := 0; x <= ap.terminal.CursorX; x++ {
			ap.terminal.Cells[ap.terminal.CursorY][x] = Cell{
				Char:    ' ',
				FgColor: ap.terminal.DefaultFg,
				BgColor: ap.terminal.DefaultBg,
			}
		}
	case 2:
		for x := range ap.terminal.Cells[ap.terminal.CursorY] {
			ap.terminal.Cells[ap.terminal.CursorY][x] = Cell{
				Char:    ' ',
				FgColor: ap.terminal.DefaultFg,
				BgColor: ap.terminal.DefaultBg,
			}
		}
	}
}

func (ap *ANSIParser) get256Color(colorNum int) color.Color {
	if colorNum < 8 {
		return ansiColor(colorNum, ap.terminal.Style)
	} else if colorNum < 16 {
		return ansiBrightColor(colorNum-8, ap.terminal.Style)
	} else if colorNum < 232 {
		colorNum -= 16
		b := colorNum % 6
		colorNum /= 6
		g := colorNum % 6
		r := colorNum / 6
		return color.RGBA{
			uint8(r * 42),
			uint8(g * 42),
			uint8(b * 42),
			255,
		}
	} else {
		gray := uint8((colorNum-232)*10 + 8)
		return color.RGBA{gray, gray, gray, 255}
	}
}
