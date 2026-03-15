package provider

import (
	"bytes"
	"context"
	"errors"
	stdlog "log"
	"testing"
	"time"

	gutlog "gut/log"
	"gut/shared"
)

type stubKeyboardProvider struct{}

func (s *stubKeyboardProvider) SetKeyboardDelay(time.Duration) {}

func (s *stubKeyboardProvider) Type(context.Context, string) error { return nil }

func (s *stubKeyboardProvider) Click(context.Context, ...shared.Key) error { return nil }

func (s *stubKeyboardProvider) PressKey(context.Context, ...shared.Key) error { return nil }

func (s *stubKeyboardProvider) ReleaseKey(context.Context, ...shared.Key) error { return nil }

func TestRegistryReturnsStableMissingProviderErrors(t *testing.T) {
	registry := NewRegistry()

	if _, err := registry.Keyboard(); !errors.Is(err, ErrMissingKeyboardProvider) {
		t.Fatalf("expected keyboard missing error, got %v", err)
	}
	if _, err := registry.Mouse(); !errors.Is(err, ErrMissingMouseProvider) {
		t.Fatalf("expected mouse missing error, got %v", err)
	}
	if _, err := registry.Clipboard(); !errors.Is(err, ErrMissingClipboardProvider) {
		t.Fatalf("expected clipboard missing error, got %v", err)
	}
}

func TestRegistryFallsBackToNoopLogger(t *testing.T) {
	registry := NewRegistry()
	logger := registry.Logger()
	if logger == nil {
		t.Fatal("expected noop logger fallback")
	}
	logger.Info("hello", gutlog.Fields{"scope": "test"})
	logger.Error(errors.New("boom"), nil)
}

func TestRegistryReturnsRegisteredLogger(t *testing.T) {
	registry := NewRegistry()
	var buf bytes.Buffer
	registered := gutlog.NewStdLogger(stdlog.New(&buf, "", 0))
	registry.RegisterLogger(registered)

	registry.Logger().Info("hello", nil)
	if got := buf.String(); got == "" {
		t.Fatal("expected registered logger to be used")
	}
}

func TestRegistryReturnsRegisteredProviders(t *testing.T) {
	registry := NewRegistry()
	keyboard := &stubKeyboardProvider{}
	registry.RegisterKeyboard(keyboard)

	resolved, err := registry.Keyboard()
	if err != nil {
		t.Fatalf("unexpected keyboard lookup error: %v", err)
	}
	if resolved != keyboard {
		t.Fatal("expected the registered keyboard provider to be returned")
	}
}
