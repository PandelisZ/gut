package gutmcp

import (
	"context"
	"strings"
	"testing"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
)

func TestAccessibilitySnapshotReturnsPartialErrorsWithoutFailingTool(t *testing.T) {
	deps := newTestDeps(t, false)
	deps.accessibility.focusedWindowErr = common.PermissionDeniedOperation(
		"getFocusedWindow",
		"darwin",
		common.CapabilityAXFocusedWindowMetadata,
		"grant Accessibility",
	)

	result, output, err := deps.service.accessibilitySnapshotTool(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected no custom result")
	}
	if output.Permissions == nil {
		t.Fatalf("expected permissions snapshot")
	}
	if !strings.Contains(output.FocusedWindowError, "permission denied") {
		t.Fatalf("expected focused window error guidance, got %q", output.FocusedWindowError)
	}
	if output.FocusedElement == nil {
		t.Fatalf("expected focused element to remain available")
	}
}

func TestWindowElementsDefaultsToActiveWindow(t *testing.T) {
	deps := newTestDeps(t, false)

	_, output, err := deps.service.windowElementsTool(context.Background(), nil, WindowElementsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Window.Handle != 2 {
		t.Fatalf("expected active window handle 2, got %d", output.Window.Handle)
	}
	if got, ok := output.Root["title"].(string); !ok || got != "Terminal" {
		t.Fatalf("unexpected root title: %#v", output.Root["title"])
	}
}

func TestWindowActionMutatesResolvedWindow(t *testing.T) {
	deps := newTestDeps(t, true)

	_, output, err := deps.service.windowActionTool(context.Background(), nil, WindowActionInput{
		Kind:   "move",
		Handle: 1,
		Origin: &PointInput{X: 100, Y: 120},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Changed {
		t.Fatalf("expected changed=true")
	}
	if deps.windows.lastMoveHandle != shared.WindowHandle(1) {
		t.Fatalf("unexpected move handle: %d", deps.windows.lastMoveHandle)
	}
	if deps.windows.lastMoveOrigin != (shared.Point{X: 100, Y: 120}) {
		t.Fatalf("unexpected move origin: %#v", deps.windows.lastMoveOrigin)
	}
}

func TestAccessibilitySearchValidatesWindowHandleScope(t *testing.T) {
	deps := newTestDeps(t, false)

	_, _, err := deps.service.accessibilitySearchTool(context.Background(), nil, AccessibilitySearchInput{
		Scope: "window_handle",
	})
	if err == nil || !strings.Contains(err.Error(), "windowHandle is required") {
		t.Fatalf("expected windowHandle validation error, got %v", err)
	}
}

func TestAccessibilityActionValidatesRefPath(t *testing.T) {
	deps := newTestDeps(t, true)

	_, _, err := deps.service.accessibilityActionTool(context.Background(), nil, AccessibilityActionInput{
		Kind: "focus_ref",
		Ref: &AXRefInput{
			Scope:        "window_handle",
			WindowHandle: 2,
			Path:         []int{-1},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "ref.path entries must be non-negative") {
		t.Fatalf("expected ref path validation error, got %v", err)
	}
}

func TestNegativeLimitsAreRejected(t *testing.T) {
	deps := newTestDeps(t, false)

	if _, _, err := deps.service.screenFindColorTool(context.Background(), nil, ScreenFindColorInput{
		Color: ColorInput{R: 255, G: 0, B: 0},
		Limit: -1,
	}); err == nil || !strings.Contains(err.Error(), "limit must be non-negative") {
		t.Fatalf("expected screen_find_color limit error, got %v", err)
	}

	if _, _, err := deps.service.windowFindElementsTool(context.Background(), nil, WindowFindElementsInput{
		Role:  "button",
		Limit: -1,
	}); err == nil || !strings.Contains(err.Error(), "limit must be non-negative") {
		t.Fatalf("expected window_find_elements limit error, got %v", err)
	}

	if _, _, err := deps.service.windowElementsTool(context.Background(), nil, WindowElementsInput{
		MaxElements: -1,
	}); err == nil || !strings.Contains(err.Error(), "maxElements must be non-negative") {
		t.Fatalf("expected window_elements maxElements error, got %v", err)
	}
}

func TestBackgroundWindowMouseActionUsesStrictVirtualPath(t *testing.T) {
	deps := newTestDeps(t, true)
	deps.accessibility.searchMatches = []common.AXElementMatch{{
		Ref: common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 2, Path: []int{0, 1}},
		Metadata: common.UIElementMetadata{
			Role:       "AXButton",
			Title:      "Run",
			Enabled:    true,
			FrameKnown: true,
			Frame:      common.Rect{X: 240, Y: 180, Width: 90, Height: 40},
			Actions:    []string{string(common.AXPress)},
		},
		Depth:            1,
		ActionPointKnown: true,
		ActionPoint:      common.Point{X: 285, Y: 200},
	}}

	_, output, err := deps.service.backgroundWindowMouseActionTool(context.Background(), nil, BackgroundWindowMouseActionInput{
		Handle: 2,
		Kind:   "click",
		Point:  &PointInput{X: 225, Y: 135},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Action != "click" || output.Ref == nil || output.ScreenPoint == nil {
		t.Fatalf("unexpected output payload: %#v", output)
	}
	if output.PerformedAction != string(common.AXPress) {
		t.Fatalf("unexpected performed action: %#v", output)
	}
	if deps.windows.lastFocus != 0 {
		t.Fatalf("expected strict background action to avoid FocusWindow, got %d", deps.windows.lastFocus)
	}
	if deps.accessibility.lastRefAction.action != common.AXPress {
		t.Fatalf("expected AXPress follow-up, got %#v", deps.accessibility.lastRefAction)
	}
}

func TestBackgroundWindowMouseActionRequiresPointOrRef(t *testing.T) {
	deps := newTestDeps(t, true)

	_, _, err := deps.service.backgroundWindowMouseActionTool(context.Background(), nil, BackgroundWindowMouseActionInput{
		Handle: 2,
		Kind:   "click",
	})
	if err == nil || !strings.Contains(err.Error(), "point or ref is required") {
		t.Fatalf("expected point/ref validation error, got %v", err)
	}
}
