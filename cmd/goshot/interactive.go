package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"regexp"
	"strconv"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
	"github.com/watzon/goshot/pkg/syntax"
)

// InteractiveConfig holds all the configuration options for the interactive mode
type InteractiveConfig struct {
	// Input/Output
	Input         string
	Output        string
	ToClipboard   bool
	FromClipboard bool

	// Appearance
	WindowChrome    string
	DarkMode        bool
	Theme           string
	Language        string
	Font            string
	BackgroundColor string
	BackgroundImage string
	BackgroundFit   string
	ShowLineNumbers bool
	CornerRadius    float64
	WindowControls  bool
	WindowTitle     string

	// Padding and layout
	TabWidth     int
	StartLine    int
	EndLine      int
	LinePadding  int
	PadHoriz     int
	PadVert      int
	CodePadVert  int
	CodePadHoriz int

	// Shadow
	ShadowBlur    float64
	ShadowColor   string
	ShadowOffsetX float64
	ShadowOffsetY float64

	// Highlighting
	HighlightLines string
}

type model struct {
	config   *InteractiveConfig
	form     *huh.Form
	viewport viewport.Model
	err      error
	quitting bool
	width    int
	height   int
	// Temporary string values for numeric inputs
	cornerRadiusStr  string
	tabWidthStr      string
	padHorizStr      string
	padVertStr       string
	codePadHorizStr  string
	codePadVertStr   string
	shadowBlurStr    string
	shadowOffsetXStr string
	shadowOffsetYStr string
}

func initialModel(config *InteractiveConfig) model {
	return model{
		config:           config,
		cornerRadiusStr:  fmt.Sprintf("%.1f", config.CornerRadius),
		tabWidthStr:      fmt.Sprintf("%d", config.TabWidth),
		padHorizStr:      fmt.Sprintf("%d", config.PadHoriz),
		padVertStr:       fmt.Sprintf("%d", config.PadVert),
		codePadHorizStr:  fmt.Sprintf("%d", config.CodePadHoriz),
		codePadVertStr:   fmt.Sprintf("%d", config.CodePadVert),
		shadowBlurStr:    fmt.Sprintf("%.1f", config.ShadowBlur),
		shadowOffsetXStr: fmt.Sprintf("%.1f", config.ShadowOffsetX),
		shadowOffsetYStr: fmt.Sprintf("%.1f", config.ShadowOffsetY),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.form != nil {
			m.form.WithWidth(m.width)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+s":
			if err := m.config.Save(); err != nil {
				m.err = err
				return m, nil
			}
			return m, tea.Quit
		}
	}

	if m.form != nil {
		var cmd tea.Cmd
		newForm, cmd := m.form.Update(msg)
		if f, ok := newForm.(*huh.Form); ok {
			m.form = f
		}
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.form != nil {
		return m.form.View()
	}

	return ""
}

// StartInteractiveMode starts the interactive TUI mode
func StartInteractiveMode(defaultConfig *InteractiveConfig) (*InteractiveConfig, error) {
	// Get available syntax styles
	styles := syntax.GetAvailableStyles()

	// Create initial model with string conversions
	m := initialModel(defaultConfig)

	// Create form with multiple groups
	form := huh.NewForm(
		// Input/Output Group
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Input File").
				Description("Choose a code file to screenshot").
				Value(&defaultConfig.Input).
				Picking(true).
				Height(10),

			huh.NewInput().
				Title("Output").
				Description("Output location for image").
				Value(&defaultConfig.Output),

			huh.NewConfirm().
				Title("Copy to Clipboard").
				Value(&defaultConfig.ToClipboard),

			huh.NewConfirm().
				Title("Read from Clipboard").
				Value(&defaultConfig.FromClipboard),
		).Title("Input/Output Settings"),

		// Window Settings Group
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Window Chrome").
				Description("Choose window style").
				Options(
					huh.NewOption("macOS", "macos"),
					huh.NewOption("Windows 11", "windows11"),
				).
				Value(&defaultConfig.WindowChrome),

			huh.NewConfirm().
				Title("Dark Mode").
				Value(&defaultConfig.DarkMode),

			huh.NewConfirm().
				Title("Window Controls").
				Value(&defaultConfig.WindowControls),

			huh.NewInput().
				Title("Window Title").
				Value(&defaultConfig.WindowTitle),

			huh.NewInput().
				Title("Corner Radius").
				Value(&m.cornerRadiusStr).
				Validate(validateFloat),

			huh.NewSelect[string]().
				Title("Theme").
				Description("Syntax highlighting theme").
				Options(huh.NewOptions(styles...)...).
				Value(&defaultConfig.Theme).
				Height(10),

			huh.NewSelect[string]().
				Title("Language").
				Description("Programming language").
				Options(huh.NewOptions(syntax.GetAvailableLanguages(false)...)...).
				Value(&defaultConfig.Language).
				Height(10),

			huh.NewSelect[string]().
				Title("Font").
				Description("Code font").
				Options(huh.NewOptions(fonts.ListFonts()...)...).
				Value(&defaultConfig.Font).
				Height(10),

			huh.NewConfirm().
				Title("Show Line Numbers").
				Value(&defaultConfig.ShowLineNumbers),

			huh.NewInput().
				Title("Background Color").
				Description("Hex color code (e.g. #FF0000)").
				Value(&defaultConfig.BackgroundColor).
				Validate(validateColor),

			huh.NewFilePicker().
				Title("Background Image").
				Description("Choose background image").
				Value(&defaultConfig.BackgroundImage),

			huh.NewInput().
				Title("Tab Width").
				Value(&m.tabWidthStr).
				Validate(validateInteger),

			huh.NewInput().
				Title("Horizontal Padding").
				Value(&m.padHorizStr).
				Validate(validateInteger),

			huh.NewInput().
				Title("Vertical Padding").
				Value(&m.padVertStr).
				Validate(validateInteger),

			huh.NewInput().
				Title("Code Horizontal Padding").
				Value(&m.codePadHorizStr).
				Validate(validateInteger),

			huh.NewInput().
				Title("Code Vertical Padding").
				Value(&m.codePadVertStr).
				Validate(validateInteger),

			huh.NewInput().
				Title("Shadow Blur").
				Value(&m.shadowBlurStr).
				Validate(validateFloat),

			huh.NewInput().
				Title("Shadow Color").
				Description("Hex color code (e.g. #000000)").
				Value(&defaultConfig.ShadowColor).
				Validate(validateColor),

			huh.NewInput().
				Title("Shadow X Offset").
				Value(&m.shadowOffsetXStr).
				Validate(validateFloat),

			huh.NewInput().
				Title("Shadow Y Offset").
				Value(&m.shadowOffsetYStr).
				Validate(validateFloat),
		).Title("Settings").WithHeight(33),
	).WithTheme(setupTheme()).WithWidth(40)

	m.form = form

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel.(model).quitting {
		return nil, fmt.Errorf("cancelled")
	}

	// Convert string values back to their proper types
	fm := finalModel.(model)
	if v, err := strconv.ParseFloat(fm.cornerRadiusStr, 64); err == nil {
		fm.config.CornerRadius = v
	}
	if v, err := strconv.Atoi(fm.tabWidthStr); err == nil {
		fm.config.TabWidth = v
	}
	if v, err := strconv.Atoi(fm.padHorizStr); err == nil {
		fm.config.PadHoriz = v
	}
	if v, err := strconv.Atoi(fm.padVertStr); err == nil {
		fm.config.PadVert = v
	}
	if v, err := strconv.Atoi(fm.codePadHorizStr); err == nil {
		fm.config.CodePadHoriz = v
	}
	if v, err := strconv.Atoi(fm.codePadVertStr); err == nil {
		fm.config.CodePadVert = v
	}
	if v, err := strconv.ParseFloat(fm.shadowBlurStr, 64); err == nil {
		fm.config.ShadowBlur = v
	}
	if v, err := strconv.ParseFloat(fm.shadowOffsetXStr, 64); err == nil {
		fm.config.ShadowOffsetX = v
	}
	if v, err := strconv.ParseFloat(fm.shadowOffsetYStr, 64); err == nil {
		fm.config.ShadowOffsetY = v
	}

	return fm.config, nil
}

func (c *InteractiveConfig) Save() error {
	err := c.WriteImage()
	if err != nil {
		return err
	}

	if c.ToClipboard {
		fmt.Printf("Screenshot saved to clipboard\n")
		return nil
	}

	if c.Output != "" {
		fmt.Printf("Screenshot saved to %s\n", c.Output)
		return nil
	}

	return nil
}

// WriteImage saves the current configuation as an image
func (c *InteractiveConfig) WriteImage() error {
	var err error
	var input string

	// Handle input source
	if c.FromClipboard {
		input, err = clipboard.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read from clipboard: %w", err)
		}
	} else {
		if c.Input == "" {
			return errors.New("no input provided")
		}
		input = c.Input
	}

	// Create renderer with the configuration
	r := render.NewCanvas()

	// Configure window chrome
	themeVariant := chrome.ThemeVariantLight
	if c.DarkMode {
		themeVariant = chrome.ThemeVariantDark
	}

	var window chrome.Chrome
	switch c.WindowChrome {
	case "macos":
		window = chrome.NewMacChrome(chrome.MacStyleSequoia)
	case "windows11":
		window = chrome.NewWindowsChrome(chrome.WindowsStyleWin11)
	case "gnome":
		window = chrome.NewGNOMEChrome(chrome.GNOMEStyleBreeze)
	}

	window.SetTitle(c.WindowTitle)
	window.SetCornerRadius(c.CornerRadius)
	window.SetTitleBar(c.WindowControls)
	window.SetThemeByName(c.Theme, themeVariant)

	r.SetChrome(window)

	// Configure background
	if c.BackgroundImage != "" {
		bg, err := background.NewImageBackgroundFromFile(c.BackgroundImage)
		if err != nil {
			return fmt.Errorf("failed to load background image: %w", err)
		}
		r.SetBackground(bg)
	} else if c.BackgroundColor != "" {
		color, err := parseHexColor(c.BackgroundColor)
		if err != nil {
			return fmt.Errorf("invalid background color: %w", err)
		}
		bg := background.NewColorBackground().
			SetColor(color).
			SetPadding(c.PadVert).
			SetCornerRadius(c.CornerRadius)

		if c.ShadowBlur > 0 {
			shadowColor, err := parseHexColor(c.ShadowColor)
			if err != nil {
				return fmt.Errorf("invalid shadow color: %w", err)
			}
			shadow := background.NewShadow().
				SetOffset(c.ShadowOffsetX, c.ShadowOffsetY).
				SetBlur(c.ShadowBlur).
				SetColor(shadowColor)
			bg.SetShadow(shadow)
		}

		r.SetBackground(bg)
	}

	// Configure syntax highlighting
	codeStyle := &render.CodeStyle{
		Theme:           c.Theme,
		Language:        c.Language,
		TabWidth:        c.TabWidth,
		ShowLineNumbers: c.ShowLineNumbers,
		PaddingX:        c.CodePadHoriz,
		PaddingY:        c.CodePadVert,
	}

	if c.Font != "" {
		font, err := fonts.GetFont(c.Font, &fonts.FontStyle{})
		if err != nil {
			return fmt.Errorf("failed to load font: %w", err)
		}
		codeStyle.FontFamily = font
	}

	r.SetCodeStyle(codeStyle)

	// Handle output
	img, err := r.RenderToImage(input)
	if err != nil {
		return fmt.Errorf("failed to render image: %w", err)
	}

	if c.ToClipboard {
		var pngData bytes.Buffer
		err = png.Encode(&pngData, img)
		if err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}

		err = clipboard.WriteAll(pngData.String())
		if err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
	}

	if c.Output != "" {
		err = render.SaveAsPNG(img, c.Output)
		if err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}
	}

	return nil
}

func setupTheme() *huh.Theme {
	green := lipgloss.Color("#03BF87")
	theme := huh.ThemeCharm()
	theme.FieldSeparator = lipgloss.NewStyle()
	theme.Blurred.TextInput.Text = theme.Blurred.TextInput.Text.Foreground(lipgloss.Color("243"))
	theme.Blurred.BlurredButton = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).PaddingRight(1)
	theme.Blurred.FocusedButton = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).PaddingRight(1)
	theme.Focused.BlurredButton = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).PaddingRight(1)
	theme.Focused.FocusedButton = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).PaddingRight(1)
	theme.Focused.NoteTitle = theme.Focused.NoteTitle.Margin(1, 0)
	theme.Blurred.NoteTitle = theme.Blurred.NoteTitle.Margin(1, 0)
	theme.Blurred.Description = theme.Blurred.Description.Foreground(lipgloss.Color("0"))
	theme.Focused.Description = theme.Focused.Description.Foreground(lipgloss.Color("7"))
	theme.Blurred.Title = theme.Blurred.Title.Width(18).Foreground(lipgloss.Color("7"))
	theme.Focused.Title = theme.Focused.Title.Width(18).Foreground(green).Bold(true)
	theme.Blurred.SelectedOption = theme.Blurred.SelectedOption.Foreground(lipgloss.Color("243"))
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(green)
	theme.Focused.Base.BorderForeground(green)
	return theme
}

var colorRegex = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")

func validateColor(s string) error {
	if s == "" {
		return nil
	}
	if !colorRegex.MatchString(s) {
		return errors.New("invalid color format")
	}
	return nil
}

func validateInteger(s string) error {
	if s == "" {
		return nil
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("must be an integer")
	}
	return nil
}

func validateFloat(s string) error {
	if s == "" {
		return nil
	}
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return errors.New("must be a number")
	}
	return nil
}
