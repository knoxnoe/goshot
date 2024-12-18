package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/watzon/goshot/cmd/goshot/config"
	"github.com/watzon/goshot/cmd/goshot/utils"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	// Form styles
	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 0)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	// Preview styles
	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 0)

	// Section styles
	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			MarginTop(1)
)

type formInput struct {
	textinput textinput.Model
	label     string
	section   string
}

type model struct {
	inputs     []formInput
	viewport   viewport.Model
	preview    string
	previewImg []byte
	err        error
	width      int
	height     int
	focused    int
	cfg        config.Config
}

func newInput(placeholder, value, label, section string) formInput {
	i := textinput.New()
	i.Placeholder = placeholder
	if value != "" {
		i.SetValue(value)
	}
	return formInput{
		textinput: i,
		label:     label,
		section:   section,
	}
}

func initialModel() model {
	var inputs []formInput

	// Input/Output section
	inputs = append(inputs, newInput("", "", "Input File", "Input/Output"))
	inputs = append(inputs, newInput("output.png", "", "Output File", "Input/Output"))
	inputs = append(inputs, newInput("", "", "Background Image", "Input/Output"))

	// Appearance section
	inputs = append(inputs, newInput("mac", "mac", "Window Chrome", "Appearance"))
	inputs = append(inputs, newInput("", "", "Chrome Theme", "Appearance"))
	inputs = append(inputs, newInput("ayu-dark", "ayu-dark", "Theme", "Appearance"))
	inputs = append(inputs, newInput("JetBrainsMonoNerdFont", "JetBrainsMonoNerdFont", "Font", "Appearance"))
	inputs = append(inputs, newInput("1.0", "1.0", "Line Height", "Appearance"))
	inputs = append(inputs, newInput("#ABB8C3", "#ABB8C3", "Background Color", "Appearance"))
	inputs = append(inputs, newInput("cover", "cover", "Background Image Fit", "Appearance"))
	inputs = append(inputs, newInput("0.0", "0.0", "Background Blur", "Appearance"))
	inputs = append(inputs, newInput("gaussian", "gaussian", "Background Blur Type", "Appearance"))
	inputs = append(inputs, newInput("10.0", "10.0", "Corner Radius", "Appearance"))
	inputs = append(inputs, newInput("", "", "Window Title", "Appearance"))

	// Gradient section
	inputs = append(inputs, newInput("", "", "Gradient Type", "Gradient"))
	inputs = append(inputs, newInput("#232323;0,#383838;100", "", "Gradient Stops", "Gradient"))
	inputs = append(inputs, newInput("45.0", "45.0", "Gradient Angle", "Gradient"))
	inputs = append(inputs, newInput("0.5", "0.5", "Gradient Center X", "Gradient"))
	inputs = append(inputs, newInput("0.5", "0.5", "Gradient Center Y", "Gradient"))
	inputs = append(inputs, newInput("5.0", "5.0", "Gradient Intensity", "Gradient"))

	// Layout section
	inputs = append(inputs, newInput("2", "2", "Line Padding", "Layout"))
	inputs = append(inputs, newInput("100", "100", "Horizontal Padding", "Layout"))
	inputs = append(inputs, newInput("80", "80", "Vertical Padding", "Layout"))
	inputs = append(inputs, newInput("4", "4", "Tab Width", "Layout"))
	inputs = append(inputs, newInput("0", "0", "Min Width", "Layout"))
	inputs = append(inputs, newInput("0", "0", "Max Width", "Layout"))

	// Shadow section
	inputs = append(inputs, newInput("0.0", "0.0", "Shadow Blur Radius", "Shadow"))
	inputs = append(inputs, newInput("#00000033", "#00000033", "Shadow Color", "Shadow"))
	inputs = append(inputs, newInput("0.0", "0.0", "Shadow Spread", "Shadow"))
	inputs = append(inputs, newInput("0.0", "0.0", "Shadow Offset X", "Shadow"))
	inputs = append(inputs, newInput("0.0", "0.0", "Shadow Offset Y", "Shadow"))

	// Focus first input
	inputs[0].textinput.Focus()

	return model{
		inputs:  inputs,
		focused: 0,
		cfg:     config.Default,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "f5":
			// Toggle between Kitty graphics and ANSI fallback
			utils.ForceANSIFallback = !utils.ForceANSIFallback
			m.updatePreview()
			return m, nil
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focused--
				if m.focused < 0 {
					m.focused = len(m.inputs) - 1
				}
			} else {
				m.focused++
				if m.focused >= len(m.inputs) {
					m.focused = 0
				}
			}

			for i := 0; i < len(m.inputs); i++ {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].textinput.Focus())
				} else {
					m.inputs[i].textinput.Blur()
				}
			}

			return m, tea.Batch(cmds...)
		case "enter":
			// Update preview when enter is pressed
			m.updatePreview()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// Calculate widths based on screen size
		// Give more space to the preview area
		formWidth := m.width / 3
		if formWidth < 50 {
			formWidth = 50 // Minimum form width
		}
		previewWidth := m.width - formWidth - 4

		// Update styles with new dimensions
		formStyle = formStyle.Width(formWidth)
		previewStyle = previewStyle.
			Width(previewWidth).
			Height(m.height - 2) // Account for margins

		if m.viewport.Width == 0 {
			m.viewport = viewport.New(previewWidth-2, m.height-4)
		} else {
			m.viewport.Width = previewWidth - 2
			m.viewport.Height = m.height - 4
		}

		// Re-render preview with new dimensions
		m.updatePreview()
	}

	// Handle character input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i].textinput, cmds[i] = m.inputs[i].textinput.Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *model) updatePreview() {
	// Update config from inputs
	m.cfg.Input = m.inputs[0].textinput.Value()
	m.cfg.OutputFile = m.inputs[1].textinput.Value()
	m.cfg.BackgroundImage = m.inputs[2].textinput.Value()
	m.cfg.WindowChrome = m.inputs[3].textinput.Value()
	m.cfg.ChromeThemeName = m.inputs[4].textinput.Value()
	m.cfg.Theme = m.inputs[5].textinput.Value()
	m.cfg.Font = m.inputs[6].textinput.Value()
	m.cfg.LineHeight = utils.ParseFloat64(m.inputs[7].textinput.Value(), 1.0)
	m.cfg.BackgroundColor = m.inputs[8].textinput.Value()
	m.cfg.BackgroundImageFit = m.inputs[9].textinput.Value()
	m.cfg.BackgroundBlur = utils.ParseFloat64(m.inputs[10].textinput.Value(), 0.0)
	m.cfg.BackgroundBlurType = m.inputs[11].textinput.Value()
	m.cfg.CornerRadius = utils.ParseFloat64(m.inputs[12].textinput.Value(), 10.0)
	m.cfg.WindowTitle = m.inputs[13].textinput.Value()

	// Gradient settings
	m.cfg.GradientType = m.inputs[14].textinput.Value()
	m.cfg.GradientStops = strings.Split(m.inputs[15].textinput.Value(), ",")
	m.cfg.GradientAngle = utils.ParseFloat64(m.inputs[16].textinput.Value(), 45.0)
	m.cfg.GradientCenterX = utils.ParseFloat64(m.inputs[17].textinput.Value(), 0.5)
	m.cfg.GradientCenterY = utils.ParseFloat64(m.inputs[18].textinput.Value(), 0.5)
	m.cfg.GradientIntensity = utils.ParseFloat64(m.inputs[19].textinput.Value(), 5.0)

	// Layout settings
	m.cfg.LinePadding = utils.ParseInt(m.inputs[20].textinput.Value(), 2)
	m.cfg.PadHoriz = utils.ParseInt(m.inputs[21].textinput.Value(), 100)
	m.cfg.PadVert = utils.ParseInt(m.inputs[22].textinput.Value(), 80)
	m.cfg.TabWidth = utils.ParseInt(m.inputs[23].textinput.Value(), 4)
	m.cfg.MinWidth = utils.ParseInt(m.inputs[24].textinput.Value(), 0)
	m.cfg.MaxWidth = utils.ParseInt(m.inputs[25].textinput.Value(), 0)

	// Shadow settings
	m.cfg.ShadowBlurRadius = utils.ParseFloat64(m.inputs[26].textinput.Value(), 0.0)
	m.cfg.ShadowColor = m.inputs[27].textinput.Value()
	m.cfg.ShadowSpread = utils.ParseFloat64(m.inputs[28].textinput.Value(), 0.0)
	m.cfg.ShadowOffsetX = utils.ParseFloat64(m.inputs[29].textinput.Value(), 0.0)
	m.cfg.ShadowOffsetY = utils.ParseFloat64(m.inputs[30].textinput.Value(), 0.0)

	// Try to render preview
	if m.cfg.Input == "" {
		m.preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Enter an input file path to see the preview")
		return
	}

	// Read input file
	content, err := os.ReadFile(m.cfg.Input)
	if err != nil {
		m.preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Render(fmt.Sprintf("Error reading file: %v", err))
		return
	}

	// Create temporary file for preview
	tmpFile := fmt.Sprintf("preview_%d.png", time.Now().UnixNano())
	m.cfg.OutputFile = tmpFile
	defer os.Remove(tmpFile) // Ensure cleanup happens

	// Render the code to an image
	if err := utils.RenderCode(&m.cfg, false, string(content)); err != nil {
		m.preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Render(fmt.Sprintf("Error rendering code: %v", err))
		return
	}

	// Wait a moment for the file to be written
	time.Sleep(100 * time.Millisecond)

	// Read the generated preview image
	img, err := utils.LoadImage(tmpFile)
	if err != nil {
		m.preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Render(fmt.Sprintf("Error loading preview: %v", err))
		return
	}

	// Render the image using terminal graphics
	preview, err := utils.RenderImageToTerminal(img)
	if err != nil {
		m.preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Render(fmt.Sprintf("Error rendering preview: %v", err))
		return
	}

	m.preview = preview
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}

	// Group inputs by section
	sections := make(map[string][]string)
	for _, input := range m.inputs {
		label := input.label + ":"
		formattedInput := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(20).Render(label),
			input.textinput.View(),
		)
		sections[input.section] = append(sections[input.section], formattedInput)
	}

	// Build form content
	var formContent strings.Builder
	for _, section := range []string{"Input/Output", "Appearance", "Gradient", "Layout", "Shadow"} {
		if inputs, ok := sections[section]; ok {
			formContent.WriteString(sectionStyle.Render(section))
			formContent.WriteString("\n")
			formContent.WriteString(strings.Join(inputs, "\n"))
			formContent.WriteString("\n")
		}
	}

	// Create preview content
	previewContent := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Preview"),
		"",
		m.preview,
	)

	// Add help text at the bottom of the form
	formContent.WriteString("\n")
	formContent.WriteString(lipgloss.NewStyle().
		Faint(true).
		MarginTop(1).
		Render("Press F5 to toggle between Kitty graphics and ANSI mode"))

	// Layout form and preview side by side
	return docStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			formStyle.Render(formContent.String()),
			previewStyle.Render(previewContent),
		),
	)
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive mode",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
