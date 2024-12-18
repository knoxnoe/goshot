package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration options for goshot
type Config struct {
	// Interactive mode
	Interactive bool

	// Input/Output options
	Input         string
	Args          []string
	OutputFile    string
	ToClipboard   bool
	FromClipboard bool
	ToStdout      bool

	// Appearance
	WindowChrome       string
	ChromeThemeName    string
	LightMode          bool
	Theme              string
	Language           string
	Font               string
	LineHeight         float64
	BackgroundColor    string
	BackgroundImage    string
	BackgroundImageFit string
	BackgroundBlur     float64
	BackgroundBlurType string
	NoLineNumbers      bool
	CornerRadius       float64
	NoWindowControls   bool
	WindowTitle        string
	WindowCornerRadius float64
	LineRanges         []string
	HighlightLines     []string

	// Gradient options
	GradientType      string
	GradientStops     []string
	GradientAngle     float64
	GradientCenterX   float64
	GradientCenterY   float64
	GradientIntensity float64

	// Padding and layout
	TabWidth          int
	StartLine         int
	EndLine           int
	LinePadding       int
	PadHoriz          int
	PadVert           int
	CodePadTop        int
	CodePadBottom     int
	CodePadLeft       int
	CodePadRight      int
	LineNumberPadding int
	MinWidth          int
	MaxWidth          int

	// Shadow options
	ShadowBlurRadius float64
	ShadowColor      string
	ShadowSpread     float64
	ShadowOffsetX    float64
	ShadowOffsetY    float64

	// Exec specific
	CellWidth      int
	CellHeight     int
	AutoSize       bool
	CellPadLeft    int
	CellPadRight   int
	CellPadTop     int
	CellPadBottom  int
	CellSpacing    int
	ShowPrompt     bool
	AutoTitle      bool
	PromptTemplate string

	// Redaction settings
	RedactionEnabled    bool
	RedactionStyle      string
	RedactionBlurRadius float64
	RedactionPatterns   []string
	RedactionAreas      []string
}

// Default configuration instance
var Default Config

func getConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error getting home directory: %v", err), "\n")
			return ""
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configHome, "goshot")
}

func saveDefaultConfig() error {
	configDir := getConfigDir()
	if configDir == "" {
		return fmt.Errorf("could not determine config directory")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Skip if config file already exists
	if _, err := os.Stat(configFile); err == nil {
		return nil
	}

	// Get all settings from viper
	settings := viper.AllSettings()

	// Marshal to YAML
	yamlData, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the file
	if err := os.WriteFile(configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// bindConfig binds viper values to the Default config struct
func bindConfig() {
	Default.OutputFile = viper.GetString("io.output_file")
	Default.ToClipboard = viper.GetBool("io.copy_to_clipboard")

	// Appearance
	Default.WindowChrome = viper.GetString("appearance.window_chrome")
	Default.ChromeThemeName = viper.GetString("appearance.chrome_theme")
	Default.LightMode = viper.GetBool("appearance.light_mode")
	Default.Theme = viper.GetString("appearance.theme")
	Default.Font = viper.GetString("appearance.font")
	Default.LineHeight = viper.GetFloat64("appearance.line_height")
	Default.BackgroundColor = viper.GetString("appearance.background.color")
	Default.BackgroundImage = viper.GetString("appearance.background.image.source")
	Default.BackgroundImageFit = viper.GetString("appearance.background.image_fit")
	Default.BackgroundBlur = viper.GetFloat64("appearance.background.blur.radius")
	Default.BackgroundBlurType = viper.GetString("appearance.background.blur.type")
	Default.NoLineNumbers = !viper.GetBool("appearance.line_numbers")
	Default.CornerRadius = viper.GetFloat64("appearance.corner_radius")
	Default.NoWindowControls = !viper.GetBool("appearance.window.controls")
	Default.WindowTitle = viper.GetString("appearance.window.title")
	Default.WindowCornerRadius = viper.GetFloat64("appearance.window.corner_radius")
	Default.LineRanges = viper.GetStringSlice("appearance.lines.ranges")
	Default.HighlightLines = viper.GetStringSlice("appearance.lines.highlight")

	// Gradient
	Default.GradientType = viper.GetString("appearance.background.gradient.type")
	Default.GradientStops = viper.GetStringSlice("appearance.background.gradient.stops")
	Default.GradientAngle = viper.GetFloat64("appearance.background.gradient.angle")
	Default.GradientCenterX = viper.GetFloat64("appearance.background.gradient.center.x")
	Default.GradientCenterY = viper.GetFloat64("appearance.background.gradient.center.y")
	Default.GradientIntensity = viper.GetFloat64("appearance.background.gradient.intensity")

	// Shadow
	Default.ShadowBlurRadius = viper.GetFloat64("appearance.shadow.blur_radius")
	Default.ShadowColor = viper.GetString("appearance.shadow.color")
	Default.ShadowSpread = viper.GetFloat64("appearance.shadow.spread")
	Default.ShadowOffsetX = viper.GetFloat64("appearance.shadow.offset.x")
	Default.ShadowOffsetY = viper.GetFloat64("appearance.shadow.offset.y")

	// Layout
	Default.LinePadding = viper.GetInt("appearance.layout.line_pad")
	Default.PadHoriz = viper.GetInt("appearance.layout.padding.horizontal")
	Default.PadVert = viper.GetInt("appearance.layout.padding.vertical")
	Default.CodePadTop = viper.GetInt("appearance.layout.code_padding.top")
	Default.CodePadBottom = viper.GetInt("appearance.layout.code_padding.bottom")
	Default.CodePadLeft = viper.GetInt("appearance.layout.code_padding.left")
	Default.CodePadRight = viper.GetInt("appearance.layout.code_padding.right")
	Default.LineNumberPadding = viper.GetInt("appearance.layout.line_number_padding")
	Default.MinWidth = viper.GetInt("appearance.layout.width.min")
	Default.MaxWidth = viper.GetInt("appearance.layout.width.max")
	Default.TabWidth = viper.GetInt("appearance.layout.tab_width")

	// Terminal
	Default.CellWidth = viper.GetInt("terminal.width")
	Default.CellHeight = viper.GetInt("terminal.height")
	Default.AutoSize = viper.GetBool("terminal.auto_size")
	Default.CellPadLeft = viper.GetInt("terminal.padding.left")
	Default.CellPadRight = viper.GetInt("terminal.padding.right")
	Default.CellPadTop = viper.GetInt("terminal.padding.top")
	Default.CellPadBottom = viper.GetInt("terminal.padding.bottom")
	Default.CellSpacing = viper.GetInt("terminal.cell_spacing")
	Default.ShowPrompt = viper.GetBool("terminal.show_prompt")
	Default.PromptTemplate = viper.GetString("terminal.prompt_template")

	// Redaction
	Default.RedactionEnabled = viper.GetBool("redaction.enabled")
	Default.RedactionStyle = viper.GetString("redaction.style")
	Default.RedactionBlurRadius = viper.GetFloat64("redaction.blur_radius")
	Default.RedactionPatterns = viper.GetStringSlice("redaction.patterns")
	Default.RedactionAreas = viper.GetStringSlice("redaction.areas")
}

// Initialize sets up the configuration with defaults and loads from config file
func Initialize() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set config paths
	configDir := getConfigDir()
	if configDir != "" {
		viper.AddConfigPath(configDir)
	}
	viper.AddConfigPath(".")

	// Input/Output options
	viper.SetDefault("io.output_file", "output.png")
	viper.SetDefault("io.copy_to_clipboard", false)

	// Appearance
	viper.SetDefault("appearance.window_chrome", "mac")
	viper.SetDefault("appearance.chrome_theme", "")
	viper.SetDefault("appearance.light_mode", false)
	viper.SetDefault("appearance.theme", "ayu-dark")
	viper.SetDefault("appearance.font", "JetBrainsMonoNerdFont")
	viper.SetDefault("appearance.line_height", 1.0)
	viper.SetDefault("appearance.background.color", "#ABB8C3")
	viper.SetDefault("appearance.background.image.source", "")
	viper.SetDefault("appearance.background.image_fit", "cover")
	viper.SetDefault("appearance.background.blur.radius", 0.0)
	viper.SetDefault("appearance.background.blur.type", "gaussian")
	viper.SetDefault("appearance.background.gradient.type", "")
	viper.SetDefault("appearance.background.gradient.stops", []string{"#232323;0", "#383838;100"})
	viper.SetDefault("appearance.background.gradient.angle", 45.0)
	viper.SetDefault("appearance.background.gradient.center.x", 0.5)
	viper.SetDefault("appearance.background.gradient.center.y", 0.5)
	viper.SetDefault("appearance.background.gradient.intensity", 5.0)
	viper.SetDefault("appearance.line_numbers", true)
	viper.SetDefault("appearance.corner_radius", 10.0)
	viper.SetDefault("appearance.window.controls", true)
	viper.SetDefault("appearance.window.title", "")
	viper.SetDefault("appearance.window.corner_radius", 10.0)
	viper.SetDefault("appearance.lines.ranges", []string{})
	viper.SetDefault("appearance.lines.highlight", []string{})
	viper.SetDefault("appearance.shadow.blur_radius", 0.0)
	viper.SetDefault("appearance.shadow.color", "#00000033")
	viper.SetDefault("appearance.shadow.spread", 0.0)
	viper.SetDefault("appearance.shadow.offset.x", 0.0)
	viper.SetDefault("appearance.shadow.offset.y", 0.0)
	viper.SetDefault("appearance.layout.line_pad", 2)
	viper.SetDefault("appearance.layout.padding.horizontal", 100)
	viper.SetDefault("appearance.layout.padding.vertical", 80)
	viper.SetDefault("appearance.layout.code_padding.top", 10)
	viper.SetDefault("appearance.layout.code_padding.bottom", 10)
	viper.SetDefault("appearance.layout.code_padding.left", 10)
	viper.SetDefault("appearance.layout.code_padding.right", 10)
	viper.SetDefault("appearance.layout.line_number_padding", 10)
	viper.SetDefault("appearance.layout.width.min", 0)
	viper.SetDefault("appearance.layout.width.max", 0)
	viper.SetDefault("appearance.layout.tab_width", 4)

	// Terminal options
	viper.SetDefault("terminal.width", 120)
	viper.SetDefault("terminal.height", 40)
	viper.SetDefault("terminal.auto_size", false)
	viper.SetDefault("terminal.padding.left", 1)
	viper.SetDefault("terminal.padding.right", 1)
	viper.SetDefault("terminal.padding.top", 1)
	viper.SetDefault("terminal.padding.bottom", 1)
	viper.SetDefault("terminal.cell_spacing", 0)
	viper.SetDefault("terminal.show_prompt", false)
	viper.SetDefault("terminal.prompt_template", "\x1b[1;35m❯ \x1b[0;32m[command]\x1b[0m\n")

	// Redaction options
	viper.SetDefault("redaction.enabled", false)
	viper.SetDefault("redaction.style", "block")
	viper.SetDefault("redaction.blur_radius", 5.0)
	viper.SetDefault("redaction.patterns", []string{})
	viper.SetDefault("redaction.areas", []string{})

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; save default config
			if err := saveDefaultConfig(); err != nil {
				fmt.Fprintf(os.Stderr, fmt.Sprintf("Error saving default config: %v", err), "\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error reading config file: %v", err), "\n")
		}
	}

	// Set up environment variable support
	viper.SetEnvPrefix("GOSHOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind config values to Default struct
	bindConfig()
}
