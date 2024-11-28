package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/xpty"
	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	content_term "github.com/watzon/goshot/pkg/content/term"
	"github.com/watzon/goshot/pkg/fonts"
	"github.com/watzon/goshot/pkg/render"
)

var dirStyle = lipgloss.NewStyle().
	Foreground(lipgloss.ANSIColor(39)).Bold(true)

var promptStyle = lipgloss.NewStyle().
	Foreground(lipgloss.ANSIColor(198)).Bold(true)

var commandStyle = lipgloss.NewStyle().
	Foreground(lipgloss.ANSIColor(40)).Bold(true)

func promptFunc(command string) string {
	return fmt.Sprintf(
		"%s\n%s %s\n",
		dirStyle.Render("~/goshot/cmd/examples/06_term"),
		promptStyle.Render("❯ "),
		commandStyle.Render(command))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get the command and args from the environment
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("No command specified")
		os.Exit(1)
	}

	out, err := executeCommand(ctx, args)
	if err != nil {
		panic(err)
	}

	canvas := render.NewCanvas().
		WithChrome(chrome.NewGNOMEChrome(chrome.GNOMEStyleAdwaita).
			WithVariant(chrome.ThemeVariantDark).
			WithTitle(strings.Join(args, " "))).
		WithBackground(background.NewColorBackground().
			WithPadding(80).
			WithColor(color.RGBA{R: 60, G: 56, B: 54, A: 255})).
		WithContent(content_term.DefaultRenderer(out).
			WithArgs(args).
			WithTheme("ubuntu").
			WithWidth(300).WithHeight(200).
			WithAutoSize().
			// WithShowPrompt().
			WithPromptFunc(promptFunc).
			WithFontName("JetBrainsMonoNerdFont", &fonts.FontStyle{
				Weight: fonts.WeightRegular,
				Mono:   true,
			}),
		)

	os.MkdirAll("example_output", 0755)
	err = canvas.SaveAsPNG("example_output/term.png")
	if err != nil {
		log.Fatal(err)
	}
}

func executeCommand(ctx context.Context, args []string) ([]byte, error) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width = 80
		height = 24
	}

	pty, err := xpty.NewPty(width, height)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = pty.Close()
	}()

	cmd := exec.Command(args[0], args[1:]...)
	if err := pty.Start(cmd); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	var errorOut bytes.Buffer
	go func() {
		_, _ = io.Copy(&out, pty)
		errorOut.Write(out.Bytes())
	}()

	if err := xpty.WaitProcess(ctx, cmd); err != nil {
		return errorOut.Bytes(), err //nolint: wrapcheck
	}
	return out.Bytes(), nil
}
