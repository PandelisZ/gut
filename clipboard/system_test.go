package clipboard

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type recordedCommand struct {
	name  string
	args  []string
	input string
}

func TestNewSystemProviderSelectsWaylandBackendWhenSuggested(t *testing.T) {
	provider := newSystemProvider(
		"linux",
		map[string]string{"WAYLAND_DISPLAY": "wayland-0"},
		func(name string) (string, error) {
			switch name {
			case "wl-copy", "wl-paste", "xclip", "xsel":
				return "/bin/" + name, nil
			default:
				return "", errors.New("missing")
			}
		},
		func(context.Context, string, []string, string) (string, error) {
			return "", nil
		},
	)

	if provider.backend == nil {
		t.Fatal("expected a backend to be selected")
	}
	if provider.backend.copyName != "wl-copy" {
		t.Fatalf("expected wl-copy backend, got %q", provider.backend.copyName)
	}
	if provider.backend.pasteName != "wl-paste" {
		t.Fatalf("expected wl-paste backend, got %q", provider.backend.pasteName)
	}
	if got := provider.backend.pasteArgs; !reflect.DeepEqual(got, []string{"--no-newline"}) {
		t.Fatalf("unexpected wl-paste args: %#v", got)
	}
}

func TestNewSystemProviderFallsBackWhenWaylandBackendMissing(t *testing.T) {
	provider := newSystemProvider(
		"linux",
		map[string]string{"XDG_SESSION_TYPE": "wayland"},
		func(name string) (string, error) {
			switch name {
			case "xclip":
				return "/bin/xclip", nil
			default:
				return "", errors.New("missing")
			}
		},
		func(context.Context, string, []string, string) (string, error) {
			return "", nil
		},
	)

	if provider.backend == nil {
		t.Fatal("expected a backend to be selected")
	}
	if provider.backend.copyName != "xclip" {
		t.Fatalf("expected xclip backend, got %q", provider.backend.copyName)
	}
}

func TestSystemProviderReturnsStableUnavailableError(t *testing.T) {
	provider := newSystemProvider(
		"plan9",
		nil,
		func(string) (string, error) { return "", errors.New("missing") },
		func(context.Context, string, []string, string) (string, error) {
			return "", nil
		},
	)

	if err := provider.Copy(context.Background(), "hello"); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("expected unavailable error from copy, got %v", err)
	}
	if _, err := provider.Paste(context.Background()); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("expected unavailable error from paste, got %v", err)
	}
	if _, err := provider.HasText(context.Background()); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("expected unavailable error from has text, got %v", err)
	}
	if _, err := provider.Clear(context.Background()); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("expected unavailable error from clear, got %v", err)
	}
}

func TestSystemProviderUsesInjectedExecutorForCopyPasteAndClear(t *testing.T) {
	var commands []recordedCommand
	provider := newSystemProvider(
		"darwin",
		nil,
		func(name string) (string, error) { return "/bin/" + name, nil },
		func(ctx context.Context, name string, args []string, input string) (string, error) {
			commands = append(commands, recordedCommand{name: name, args: append([]string(nil), args...), input: input})
			switch name {
			case "pbcopy":
				return "", nil
			case "pbpaste":
				return "hello", nil
			default:
				return "", errors.New("unexpected command")
			}
		},
	)

	if err := provider.Copy(context.Background(), "hello"); err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	text, err := provider.Paste(context.Background())
	if err != nil {
		t.Fatalf("paste failed: %v", err)
	}
	if text != "hello" {
		t.Fatalf("expected pasted text, got %q", text)
	}
	hasText, err := provider.HasText(context.Background())
	if err != nil {
		t.Fatalf("has text failed: %v", err)
	}
	if !hasText {
		t.Fatal("expected clipboard to report text present")
	}
	cleared, err := provider.Clear(context.Background())
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if !cleared {
		t.Fatal("expected clear to return true")
	}

	if len(commands) != 4 {
		t.Fatalf("expected 4 command invocations, got %d", len(commands))
	}
	if commands[0].name != "pbcopy" || commands[0].input != "hello" {
		t.Fatalf("unexpected copy command: %#v", commands[0])
	}
	if commands[1].name != "pbpaste" {
		t.Fatalf("unexpected paste command: %#v", commands[1])
	}
	if commands[2].name != "pbpaste" {
		t.Fatalf("unexpected has text command: %#v", commands[2])
	}
	if commands[3].name != "pbcopy" || commands[3].input != "" {
		t.Fatalf("unexpected clear command: %#v", commands[3])
	}
}

func TestSystemProviderPropagatesExecutorErrors(t *testing.T) {
	expected := errors.New("boom")
	provider := newSystemProvider(
		"darwin",
		nil,
		func(name string) (string, error) { return "/bin/" + name, nil },
		func(context.Context, string, []string, string) (string, error) {
			return "", expected
		},
	)

	if err := provider.Copy(context.Background(), "hello"); !errors.Is(err, expected) {
		t.Fatalf("expected wrapped executor error, got %v", err)
	}
}

func TestSystemProviderHonorsContextCancellationBeforeExecution(t *testing.T) {
	provider := newSystemProvider(
		"darwin",
		nil,
		func(name string) (string, error) { return "/bin/" + name, nil },
		func(context.Context, string, []string, string) (string, error) {
			t.Fatal("executor should not have been called")
			return "", nil
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := provider.Copy(ctx, "hello"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error from copy, got %v", err)
	}
	if _, err := provider.Paste(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error from paste, got %v", err)
	}
}
