package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/xpty"
	content_term "github.com/watzon/goshot/pkg/content/term"
	"github.com/watzon/goshot/pkg/render"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please provide a command to execute")
		os.Exit(1)
	}

	out, err := executeComamand(ctx, args)
	if err != nil {
		panic(err)
	}

	canvas := render.NewCanvas().
		WithContent(content_term.DefaultRenderer(out).
			WithTheme("dracula").
			// Set initial size, but the actual output will be cropped to content
			WithWidth(300).
			WithHeight(300).
			// Only render the area that's actually used
			WithAutoSize(),
		)

	os.MkdirAll("example_output", 0755)
	err = canvas.SaveAsPNG("example_output/term.png")
	if err != nil {
		log.Fatal(err)
	}
}

func executeComamand(ctx context.Context, args []string) ([]byte, error) {
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

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint: gosec
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
