package gutmcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/provider"
)

func TestParseKeyAndButton(t *testing.T) {
	key, err := parseKey("left_cmd")
	if err != nil {
		t.Fatalf("parseKey failed: %v", err)
	}
	if got := key.String(); got != "LeftCmd" {
		t.Fatalf("unexpected key: %s", got)
	}

	button, err := parseButton("RIGHT")
	if err != nil {
		t.Fatalf("parseButton failed: %v", err)
	}
	if got := button.String(); got != "right" {
		t.Fatalf("unexpected button: %s", got)
	}
}

func TestParseAXEnums(t *testing.T) {
	action, err := parseAXAction("ax_press")
	if err != nil || action != common.AXPress {
		t.Fatalf("unexpected action parse: action=%q err=%v", action, err)
	}

	scope, err := parseAXScope("frontmost_application")
	if err != nil || scope != common.AXSearchScopeFrontmostApplication {
		t.Fatalf("unexpected scope parse: scope=%q err=%v", scope, err)
	}
}

func TestActionableErrorFormatsImportantCases(t *testing.T) {
	err := actionableError("accessibility_search", common.PermissionDeniedOperation(
		"searchAXElements",
		"darwin",
		common.CapabilityAXElementSearch,
		"grant Accessibility access",
	))
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission guidance, got %q", err)
	}

	err = actionableError("window_list", provider.ErrMissingWindowProvider)
	if !strings.Contains(err.Error(), `provider "window"`) {
		t.Fatalf("expected missing provider detail, got %q", err)
	}

	if !errors.Is(actionableError("noop", context.Canceled), context.Canceled) {
		t.Fatalf("expected context cancellation to pass through")
	}
}

func TestColorInputToSharedAllowsExplicitZeroAlpha(t *testing.T) {
	alpha := 0
	color, err := (ColorInput{R: 10, G: 20, B: 30, A: &alpha}).toShared()
	if err != nil {
		t.Fatalf("toShared failed: %v", err)
	}
	if color.A != 0 {
		t.Fatalf("expected explicit alpha 0, got %d", color.A)
	}

	color, err = (ColorInput{R: 10, G: 20, B: 30}).toShared()
	if err != nil {
		t.Fatalf("toShared failed: %v", err)
	}
	if color.A != 255 {
		t.Fatalf("expected omitted alpha to default to 255, got %d", color.A)
	}
}
