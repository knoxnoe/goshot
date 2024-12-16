package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/x/xpty"
	"golang.org/x/term"
)

// ExecuteCommand executes a command and returns its output
func ExecuteCommand(ctx context.Context, args []string) ([]byte, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
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

	// Create a pipe for stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := pty.Start(cmd); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	var errorOut bytes.Buffer
	go func() {
		_, _ = io.Copy(&out, pty)
	}()

	// Read stderr
	go func() {
		_, _ = io.Copy(&errorOut, stderrPipe)
	}()

	if err := xpty.WaitProcess(ctx, cmd); err != nil {
		// Return stderr and the error
		return errorOut.Bytes(), fmt.Errorf("%s %v", errorOut.String(), err)
	}
	return out.Bytes(), nil
}

// GetTerminalSize returns the current terminal dimensions
func GetTerminalSize() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

// GetDefaultTerminalSize returns default terminal dimensions if actual size cannot be determined
func GetDefaultTerminalSize() (width, height int) {
	return 80, 24
}
