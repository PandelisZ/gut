package common

import (
	"errors"
	"testing"
)

func TestOperationErrorWrapsCause(t *testing.T) {
	err := UnavailableOperation("mouse.move", "windows", CapabilityMouseMove, "binding unavailable")
	if !errors.Is(err, ErrNativeBindingUnavailable) {
		t.Fatalf("expected ErrNativeBindingUnavailable, got %v", err)
	}

	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}

	want := "mouse.move on windows [mouse.move]: binding unavailable: native binding unavailable"
	if opErr.Error() != want {
		t.Fatalf("unexpected error string: got %q want %q", opErr.Error(), want)
	}
}

func TestCapabilityUnavailableWrapsCause(t *testing.T) {
	err := CapabilityUnavailable("window.minimize", "linux", CapabilityWindowMinimize, "libnut-core does not expose minimize")
	if !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestPermissionDeniedOperationWrapsCause(t *testing.T) {
	err := PermissionDeniedOperation("focusWindow", "darwin", CapabilityWindowFocus, "Accessibility permission is not granted")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}

	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}
	if opErr.Capability != CapabilityWindowFocus {
		t.Fatalf("unexpected capability: %s", opErr.Capability)
	}
}

func TestUnsupportedOperationWrapsCause(t *testing.T) {
	err := UnsupportedOperation("x11.display.get", "windows", CapabilityX11DisplayGet, "Linux X11 only")
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}
