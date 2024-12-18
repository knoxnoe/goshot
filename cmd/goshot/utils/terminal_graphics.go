package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TerminalGraphicsProtocol represents different terminal graphics protocols
type TerminalGraphicsProtocol int

const (
	ProtocolNone TerminalGraphicsProtocol = iota
	ProtocolKitty
	ProtocolITerm
	ProtocolSixel
)

// ProtocolDetectionResult contains information about the detected protocol
type ProtocolDetectionResult struct {
	Protocol TerminalGraphicsProtocol
	Details  string
}

var ForceANSIFallback bool

// DetectGraphicsProtocol detects the available terminal graphics protocol
func DetectGraphicsProtocol() ProtocolDetectionResult {
	var result ProtocolDetectionResult
	result.Protocol = ProtocolNone

	// Check if we're running in a terminal
	if !isTerminal() {
		result.Details = "Not running in a terminal"
		return result
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	result.Details = fmt.Sprintf("Terminal type: %s\n", term)

	// Check for Kitty protocol support
	switch term {
	case "xterm-kitty", "xterm-ghostty":
		result.Details += fmt.Sprintf("Detected %s terminal (supports Kitty graphics protocol)\n", term)
		supported, details := isKittyProtocolSupported()
		result.Details += details
		if supported {
			result.Protocol = ProtocolKitty
			return result
		} else {
			result.Details += "WARNING: Terminal should support Kitty protocol but detection failed\n"
		}
	default:
		result.Details += "Terminal does not indicate Kitty protocol support\n"
	}

	// Check for other environment hints
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		result.Details += "Kitty window ID detected but protocol not supported\n"
	}

	// Add terminal size information
	if width, height, err := getTerminalSize(); err == nil {
		result.Details += fmt.Sprintf("Terminal size: %dx%d\n", width, height)
	}

	return result
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// getTerminalSize attempts to get the terminal dimensions
func getTerminalSize() (width, height int, err error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(string(out))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected stty output format")
	}

	height, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	width, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

// isKittyProtocolSupported checks if the terminal actually supports the Kitty graphics protocol
func isKittyProtocolSupported() (bool, string) {
	var details strings.Builder
	details.WriteString("Checking Kitty graphics protocol support:\n")

	// Send a query to the terminal
	// According to protocol docs, we should send a minimal query
	queryCmd := "\x1b_Gi=31,s=1,v=1,a=q;AAAA\x1b\\"
	details.WriteString(fmt.Sprintf("Sending query: %q\n", queryCmd))

	// Send query and immediately request device attributes
	// This helps ensure we get a response
	fmt.Print(queryCmd)
	fmt.Print("\x1b[c")

	// Try to read the response with a timeout
	ch := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		// Read in a loop to handle multiple response parts
		var response []byte
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errCh <- fmt.Errorf("error reading response: %v", err)
				return
			}
			response = append(response, buf[:n]...)
			// Look for command terminator
			if bytes.Contains(response, []byte("\x1b\\")) {
				ch <- response
				return
			}
		}
	}()

	// Wait for response with timeout
	select {
	case response := <-ch:
		details.WriteString(fmt.Sprintf("Got response (%d bytes): %q\n", len(response), response))

		// Parse the response
		details.WriteString(fmt.Sprintf("Got response (%d bytes): %q\n", len(response), response))
		details.WriteString(fmt.Sprintf("Raw response: %q\n", response))

		// According to protocol docs, we need to look for a response in the format:
		// <ESC>_Gi=31;OK<ESC>\
		if bytes.Contains(response, []byte("OK")) {
			details.WriteString("Received valid Kitty graphics protocol response\n")
			return true, details.String()
		}

		// Check for known error responses
		if bytes.Contains(response, []byte("ENOTSUP")) {
			details.WriteString("Terminal returned ENOTSUP (protocol not supported)\n")
			return false, details.String()
		}

		// If we got a response but no OK and no error, log it as unexpected
		details.WriteString(fmt.Sprintf("Unexpected response format: %q\n", response))
		if bytes.Contains(response, []byte("\x1b_G")) {
			details.WriteString("WARNING: Terminal should support Kitty protocol but detection failed\n")
		}
		return false, details.String()

	case err := <-errCh:
		details.WriteString(fmt.Sprintf("Error: %v\n", err))
		return false, details.String()

	case <-time.After(500 * time.Millisecond):
		details.WriteString("Timed out waiting for response (increased timeout to 500ms)\n")
		return false, details.String()
	}
}

// RenderImageToTerminal renders an image using the best available terminal graphics protocol
func RenderImageToTerminal(img image.Image) (string, error) {
	detection := DetectGraphicsProtocol()
	var result strings.Builder

	// Start with debug info in a muted color
	result.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Terminal Detection:\n%s\n", detection.Details)))

	// Add image size info
	result.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Image size: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())))

	// Add mode info if in fallback
	if ForceANSIFallback {
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Forcing ANSI fallback mode (press F5 to toggle)\n"))
	}

	// Add a separator
	result.WriteString("\n")

	// Render the image
	var output string
	var err error
	if !ForceANSIFallback && detection.Protocol == ProtocolKitty {
		output, err = renderKittyGraphics(img)
		if err != nil {
			return "", fmt.Errorf("kitty graphics error: %v", err)
		}
	} else {
		output, err = renderANSIFallback(img)
		if err != nil {
			return "", fmt.Errorf("ANSI fallback error: %v", err)
		}
	}

	// Add the rendered output
	result.WriteString(output)

	// Add a separator after the image
	result.WriteString("\n")
	return result.String(), nil
}

// renderKittyGraphics renders an image using the Kitty graphics protocol
func renderKittyGraphics(img image.Image) (string, error) {
	// Convert image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	// Encode the raw pixels
	encoded := base64.StdEncoding.EncodeToString(rgba.Pix)

	var result strings.Builder

	// First add debug info
	result.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Terminal Detection:\n")))
	result.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Terminal type: xterm-ghostty\n")))
	result.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Image size: %dx%d\n\n", bounds.Dx(), bounds.Dy())))

	// Position cursor for image (after debug info)
	result.WriteString("\x1b[?25l") // Hide cursor
	result.WriteString("\x1b[H")    // Move to top
	result.WriteString("\x1b[55C")  // Move to preview section
	result.WriteString("\x1b[20B")  // Move down past debug info

	// Place the image
	result.WriteString("\x1b_G") // Start graphics command
	result.WriteString(fmt.Sprintf(
		"a=T,f=32,s=%d,v=%d,q=2", // q=2 suppresses responses
		bounds.Dx(),
		bounds.Dy(),
	))
	result.WriteString(";")
	result.WriteString(encoded)
	result.WriteString("\x1b\\")

	// Reset terminal state
	result.WriteString("\x1b[?25h") // Show cursor

	return result.String(), nil
}

// renderANSIFallback renders an image using ANSI escape codes as a fallback
func renderANSIFallback(img image.Image) (string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Scale down the image to fit in terminal
	const maxWidth = 80
	const maxHeight = 24
	scaleX := float64(width) / float64(maxWidth)
	scaleY := float64(height) / float64(maxHeight)
	scale := scaleX
	if scaleY > scale {
		scale = scaleY
	}
	if scale < 1 {
		scale = 1
	}

	scaledWidth := int(float64(width) / scale)
	scaledHeight := int(float64(height) / scale)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Terminal preview (%dx%d):\n", scaledWidth, scaledHeight))

	// Use block characters to represent pixels
	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			origX := int(float64(x) * scale)
			origY := int(float64(y) * scale)
			r, g, b, _ := img.At(bounds.Min.X+origX, bounds.Min.Y+origY).RGBA()
			// Convert to 8-bit color
			r8 := r >> 8
			g8 := g >> 8
			b8 := b >> 8
			// Use grayscale for simplicity
			gray := (r8*30 + g8*59 + b8*11) / 100
			if gray > 200 {
				result.WriteString("█")
			} else if gray > 150 {
				result.WriteString("▓")
			} else if gray > 100 {
				result.WriteString("▒")
			} else if gray > 50 {
				result.WriteString("░")
			} else {
				result.WriteString(" ")
			}
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}
