package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/bmp"
)

// ExpandPath expands environment variables and tilde in the given path
func ExpandPath(path string) string {
	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand tilde to home directory
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	return path
}

// EscapeCommand escapes a terminal command for use in a filename
func EscapeCommand(command string) string {
	return strings.NewReplacer(
		"/", "_",
		" ", "_",
		":", "_",
		"\\", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	).Replace(command)
}

// SaveImageToFile saves the given image to a file with the specified configuration
func SaveImageToFile(img image.Image, outputFile string) (string, error) {
	if outputFile == "" {
		return "", nil
	}

	// Expand any environment variables and tilde in the path
	resolvedFilename := ExpandPath(outputFile)

	// Get the extension from the filename
	ext := strings.ToLower(filepath.Ext(resolvedFilename))
	if ext == "" {
		ext = ".png"
		resolvedFilename += ext
	}

	// Ensure the directory exists
	if dir := filepath.Dir(resolvedFilename); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Create the file
	f, err := os.Create(resolvedFilename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Save in the format matching the extension
	switch ext {
	case ".png":
		return resolvedFilename, png.Encode(f, img)
	case ".jpg", ".jpeg":
		return resolvedFilename, jpeg.Encode(f, img, nil)
	case ".bmp":
		return resolvedFilename, bmp.Encode(f, img)
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
}

// GetDefaultFilename generates a default filename if none is provided
func GetDefaultFilename(baseDir string) string {
	now := time.Now()
	return filepath.Join(
		baseDir,
		fmt.Sprintf("goshot_%s.png", now.Format("2006-01-02_15-04-05")),
	)
}

// ImageToBytes converts an image to bytes in PNG format
func ImageToBytes(img image.Image) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image to png: %v", err)
	}
	return buf.Bytes(), nil
}
