package clipboard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/PandelisZ/gut/provider"
)

var ErrBackendUnavailable = errors.New("clipboard backend unavailable")

type commandExecutor func(ctx context.Context, name string, args []string, input string) (string, error)

type commandBackend struct {
	copyName  string
	copyArgs  []string
	pasteName string
	pasteArgs []string
}

type SystemProvider struct {
	backend *commandBackend
	exec    commandExecutor
}

func NewSystemProvider() *SystemProvider {
	return newSystemProvider(runtime.GOOS, currentEnv(), exec.LookPath, runCommand)
}

func newSystemProvider(goos string, env map[string]string, lookPath func(string) (string, error), executor commandExecutor) *SystemProvider {
	if executor == nil {
		executor = runCommand
	}
	return &SystemProvider{
		backend: selectBackend(goos, env, lookPath),
		exec:    executor,
	}
}

func (p *SystemProvider) HasText(ctx context.Context) (bool, error) {
	text, err := p.Paste(ctx)
	if err != nil {
		return false, err
	}
	return len(text) > 0, nil
}

func (p *SystemProvider) Clear(ctx context.Context) (bool, error) {
	if err := p.Copy(ctx, ""); err != nil {
		return false, err
	}
	return true, nil
}

func (p *SystemProvider) Copy(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.backend == nil {
		return ErrBackendUnavailable
	}
	_, err := p.exec(ctx, p.backend.copyName, p.backend.copyArgs, text)
	return err
}

func (p *SystemProvider) Paste(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if p.backend == nil {
		return "", ErrBackendUnavailable
	}
	return p.exec(ctx, p.backend.pasteName, p.backend.pasteArgs, "")
}

func selectBackend(goos string, env map[string]string, lookPath func(string) (string, error)) *commandBackend {
	hasCommand := func(name string) bool {
		if lookPath == nil {
			return false
		}
		_, err := lookPath(name)
		return err == nil
	}

	switch goos {
	case "darwin":
		if hasCommand("pbcopy") && hasCommand("pbpaste") {
			return &commandBackend{
				copyName:  "pbcopy",
				pasteName: "pbpaste",
			}
		}
	case "linux":
		wlBackend := &commandBackend{
			copyName:  "wl-copy",
			pasteName: "wl-paste",
			pasteArgs: []string{"--no-newline"},
		}
		xclipBackend := &commandBackend{
			copyName:  "xclip",
			copyArgs:  []string{"-selection", "clipboard"},
			pasteName: "xclip",
			pasteArgs: []string{"-selection", "clipboard", "-o"},
		}
		xselBackend := &commandBackend{
			copyName:  "xsel",
			copyArgs:  []string{"--clipboard", "--input"},
			pasteName: "xsel",
			pasteArgs: []string{"--clipboard", "--output"},
		}

		candidates := []*commandBackend{xclipBackend, xselBackend, wlBackend}
		if suggestsWayland(env) {
			candidates = []*commandBackend{wlBackend, xclipBackend, xselBackend}
		}
		for _, backend := range candidates {
			if hasCommand(backend.copyName) && hasCommand(backend.pasteName) {
				return backend
			}
		}
	case "windows":
		copyArgs := []string{"-NoProfile", "-NonInteractive", "-Command", "Set-Clipboard -Value ([Console]::In.ReadToEnd())"}
		pasteArgs := []string{"-NoProfile", "-NonInteractive", "-Command", "$text = Get-Clipboard -Raw; if ($null -ne $text) { [Console]::Out.Write($text) }"}
		for _, shell := range []string{"powershell", "pwsh"} {
			if hasCommand(shell) {
				return &commandBackend{
					copyName:  shell,
					copyArgs:  copyArgs,
					pasteName: shell,
					pasteArgs: pasteArgs,
				}
			}
		}
	}
	return nil
}

func suggestsWayland(env map[string]string) bool {
	if env == nil {
		return false
	}
	if env["WAYLAND_DISPLAY"] != "" {
		return true
	}
	return strings.EqualFold(env["XDG_SESSION_TYPE"], "wayland")
}

func currentEnv() map[string]string {
	env := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		env[key] = value
	}
	return env
}

func runCommand(ctx context.Context, name string, args []string, input string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", fmt.Errorf("clipboard command %q failed: %w", name, err)
		}
		return "", fmt.Errorf("clipboard command %q failed: %s: %w", name, message, err)
	}
	return stdout.String(), nil
}

var _ provider.ClipboardProvider = (*SystemProvider)(nil)
