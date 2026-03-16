//go:build darwin && cgo

package libnutcore

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gut/native/common"
)

func TestDarwinCaptureScreenReportsCapabilityUnavailable(t *testing.T) {
	client := New(Options{})
	status := client.Capabilities().Status(common.CapabilityScreenCapture)
	if status.Availability != common.AvailabilityUnsupported {
		t.Fatalf("expected screen capture to be unsupported, got %s", status.Availability)
	}
	if !strings.Contains(status.Reason, "intentionally disabled") {
		t.Fatalf("unexpected screen capture reason: %q", status.Reason)
	}

	_, err := client.CaptureScreen(nil)
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinKeyTapRejectsModifiers(t *testing.T) {
	client := New(Options{})
	if err := client.KeyTap("a", string(common.KeyModifierShift)); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinKeyToggleRejectsModifiers(t *testing.T) {
	client := New(Options{})
	if err := client.KeyToggle("a", common.KeyStateDown, string(common.KeyModifierMeta)); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinKeyActionsRejectUnsafeSpecialKeys(t *testing.T) {
	client := New(Options{})

	if err := client.KeyTap("audio_vol_up"); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected media key tap to return ErrCapabilityUnavailable, got %v", err)
	}
	if err := client.KeyToggle("printscreen", common.KeyStateDown); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected special key toggle to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinCapabilityNotesDescribeSafetyModel(t *testing.T) {
	client := New(Options{})
	capabilities := client.Capabilities()

	keyboardTap := capabilities.Status(common.CapabilityKeyboardTap)
	if keyboardTap.Availability != common.AvailabilityAvailable {
		t.Fatalf("expected keyboard tap to remain available, got %s", keyboardTap.Availability)
	}
	if !strings.Contains(keyboardTap.Reason, "modifier chords") {
		t.Fatalf("unexpected keyboard tap reason: %q", keyboardTap.Reason)
	}

	highlight := capabilities.Status(common.CapabilityScreenHighlight)
	if highlight.Availability != common.AvailabilityAvailable {
		t.Fatalf("expected highlight to remain available, got %s", highlight.Availability)
	}
	if !strings.Contains(highlight.Reason, "AppKit main thread") {
		t.Fatalf("unexpected highlight reason: %q", highlight.Reason)
	}

	info := client.Info()
	joinedNotes := strings.Join(info.Notes, " ")
	if !strings.Contains(joinedNotes, "AppKit") || !strings.Contains(joinedNotes, "Screen capture is intentionally unavailable") {
		t.Fatalf("unexpected backend notes: %q", joinedNotes)
	}
}

func TestDarwinHighlightReturnsFromBackgroundGoroutine(t *testing.T) {
	client := New(Options{})
	done := make(chan error, 1)

	go func() {
		done <- client.Highlight(common.Rect{X: 0, Y: 0, Width: 1, Height: 1}, 10*time.Millisecond, 0)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected highlight error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("highlight from background goroutine did not return")
	}
}
