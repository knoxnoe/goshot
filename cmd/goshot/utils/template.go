package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/watzon/goshot/cmd/goshot/config"
)

// TemplateData holds data that can be used in templates
type TemplateData struct {
	// System information
	User    string
	Host    string
	Path    string
	Command string

	// File information (from input file)
	Filename string // Full filename with extension (or "stdin" for clipboard/stdin input)
	FileBase string // Filename without extension (or "goshot" for clipboard/stdin input)
	FileExt  string // File extension with dot (or "" for clipboard/stdin input)
	FileDir  string // Directory containing the file (or cwd for clipboard/stdin input)

	// Other data
	Config   *config.Config
	DateTime time.Time
}

// NewTemplateData creates a new TemplateData instance with the given command and config
func NewTemplateData(command string, cfg *config.Config) (*TemplateData, error) {
	data := &TemplateData{
		Command:  command,
		DateTime: time.Now(),
		Config:   cfg,
		FileBase: "goshot", // Default filename base
	}

	// Get system information
	if usr, err := user.Current(); err == nil {
		data.User = usr.Username
	}
	if host, err := os.Hostname(); err == nil {
		data.Host = host
	}
	if cwd, err := os.Getwd(); err == nil {
		data.Path = cwd
		data.FileDir = cwd // Default FileDir to cwd
	}

	// Only process file information if we have a config
	if cfg != nil {
		// Set file information based on input type
		switch {
		case cfg.Input != "":
			// Using a file input
			data.Filename = filepath.Base(cfg.Input)
			data.FileBase = strings.TrimSuffix(filepath.Base(cfg.Input), filepath.Ext(cfg.Input))
			data.FileExt = filepath.Ext(cfg.Input)
			data.FileDir = filepath.Dir(cfg.Input)
		case cfg.FromClipboard:
			// Using clipboard input
			data.Filename = "clipboard"
			data.FileBase = "goshot"
			data.FileExt = ""
		case command != "" || len(cfg.Args) > 0:
			// Using command input
			if command != "" {
				data.Filename = EscapeCommand(command)
			} else {
				data.Filename = EscapeCommand(strings.Join(cfg.Args, " "))
			}
			data.FileBase = "goshot"
			data.FileExt = ""
		default:
			// Using stdin or other input
			data.Filename = "stdin"
			data.FileBase = "goshot"
			data.FileExt = ""
		}
	}

	return data, nil
}

// NewPromptFunc creates a function that generates a prompt string using the given template
func NewPromptFunc(tmpl string, cfg *config.Config) func(command string) string {
	return func(command string) string {
		t, err := template.New("prompt").Parse(tmpl)
		if err != nil {
			return tmpl // Return raw template on error
		}

		data, err := NewTemplateData(command, cfg)
		if err != nil {
			return tmpl
		}

		var buf strings.Builder
		if err := t.Execute(&buf, data); err != nil {
			return tmpl
		}

		return buf.String()
	}
}

// NewFilenameFunc creates a function that generates a filename using the given template
func NewFilenameFunc(tmpl string, cfg *config.Config) func() string {
	return func() string {
		// Create template with custom delimiters to avoid conflicts with shell
		t := template.New("filename").Delims("{{", "}}")

		// Add template functions
		t = t.Funcs(template.FuncMap{
			"formatDate": func(format string) string {
				return time.Now().Format(format)
			},
		})

		// Parse template
		t, err := t.Parse(tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid filename template '%s': %v\n", tmpl, err)
			return GetDefaultFilename(filepath.Dir(cfg.OutputFile))
		}

		// Create template data
		data, err := NewTemplateData("", cfg)
		if err != nil {
			return GetDefaultFilename(filepath.Dir(cfg.OutputFile))
		}

		// Execute template
		var buf strings.Builder
		if err := t.Execute(&buf, data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to execute filename template: %v\n", err)
			return GetDefaultFilename(filepath.Dir(cfg.OutputFile))
		}

		// Expand environment variables and tilde in the result
		result := ExpandPath(buf.String())

		// Ensure the directory exists
		dir := filepath.Dir(result)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create directory '%s': %v\n", dir, err)
			return GetDefaultFilename(filepath.Dir(cfg.OutputFile))
		}

		return result
	}
}
