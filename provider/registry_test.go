package provider

import (
	"bytes"
	"context"
	"errors"
	stdlog "log"
	"testing"
	"time"

	gutlog "gut/log"
	"gut/native/common"
	"gut/shared"
)

type stubAccessibilityProvider struct{}

func (s *stubAccessibilityProvider) GetPermissionSnapshot(context.Context) (common.PermissionSnapshot, error) {
	return common.PermissionSnapshot{}, nil
}

func (s *stubAccessibilityProvider) GetFocusedWindow(context.Context) (common.FocusedWindowMetadata, error) {
	return common.FocusedWindowMetadata{}, nil
}

func (s *stubAccessibilityProvider) GetFocusedElement(context.Context) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, nil
}

func (s *stubAccessibilityProvider) GetElementAtPoint(context.Context, shared.Point) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, nil
}

func (s *stubAccessibilityProvider) RaiseFocusedWindow(context.Context) error { return nil }

func (s *stubAccessibilityProvider) PerformFocusedElementAction(context.Context, common.AXAction) error {
	return nil
}

func (s *stubAccessibilityProvider) PerformElementActionAtPoint(context.Context, shared.Point, common.AXAction) error {
	return nil
}

func (s *stubAccessibilityProvider) FocusElementAtPoint(context.Context, shared.Point) error {
	return nil
}

func (s *stubAccessibilityProvider) SearchAXElements(context.Context, common.AXElementSearchQuery) ([]common.AXElementMatch, error) {
	return nil, nil
}

func (s *stubAccessibilityProvider) FocusAXElement(context.Context, common.AXElementRef) error {
	return nil
}

func (s *stubAccessibilityProvider) PerformAXElementAction(context.Context, common.AXElementRef, common.AXAction) error {
	return nil
}

func (s *stubAccessibilityProvider) Capabilities() common.CapabilitySet { return nil }

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
	if _, err := registry.Accessibility(); !errors.Is(err, ErrMissingAccessibilityProvider) {
		t.Fatalf("expected accessibility missing error, got %v", err)
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
	accessibility := &stubAccessibilityProvider{}
	registry.RegisterKeyboard(keyboard)
	registry.RegisterAccessibility(accessibility)

	resolvedKeyboard, err := registry.Keyboard()
	if err != nil {
		t.Fatalf("unexpected keyboard lookup error: %v", err)
	}
	if resolvedKeyboard != keyboard {
		t.Fatal("expected the registered keyboard provider to be returned")
	}

	resolvedAccessibility, err := registry.Accessibility()
	if err != nil {
		t.Fatalf("unexpected accessibility lookup error: %v", err)
	}
	if resolvedAccessibility != accessibility {
		t.Fatal("expected the registered accessibility provider to be returned")
	}
}
