//go:build darwin && cgo

package libnutcore

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gut/native/common"
)

func TestDarwinPermissionSnapshotIncludesExpectedEntries(t *testing.T) {
	client := New(Options{})
	snapshot, err := client.GetPermissionSnapshot()
	if err != nil {
		t.Fatalf("unexpected permission snapshot error: %v", err)
	}
	if !snapshot.Accessibility.Supported {
		t.Fatal("expected Accessibility permission introspection to be supported")
	}
	if snapshot.Accessibility.Reason == "" {
		t.Fatal("expected Accessibility permission reason")
	}
	if snapshot.ScreenRecording.Reason == "" {
		t.Fatal("expected Screen Recording permission reason")
	}
}

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

func TestDarwinAccessibilityCapabilitiesReflectPermissionState(t *testing.T) {
	client := New(Options{})
	snapshot, err := client.GetPermissionSnapshot()
	if err != nil {
		t.Fatalf("unexpected permission snapshot error: %v", err)
	}
	capabilities := client.Capabilities()
	expected := common.AvailabilityPermissionBlocked
	if snapshot.Accessibility.Granted {
		expected = common.AvailabilityAvailable
	}
	for _, capability := range []common.Capability{
		common.CapabilityMouseMove,
		common.CapabilityKeyboardTap,
		common.CapabilityWindowFocus,
		common.CapabilityAXFocusedWindowMetadata,
		common.CapabilityAXFocusedElementMetadata,
		common.CapabilityAXElementAtPointMetadata,
		common.CapabilityAXFocusedWindowRaise,
		common.CapabilityAXFocusedElementAction,
		common.CapabilityAXElementActionAtPoint,
		common.CapabilityAXElementFocusAtPoint,
		common.CapabilityAXElementSearch,
		common.CapabilityAXElementFocusMatch,
		common.CapabilityAXElementActionMatch,
	} {
		if status := capabilities.Status(capability); status.Availability != expected {
			t.Fatalf("expected %s to be %s, got %s (%s)", capability, expected, status.Availability, status.Reason)
		}
	}

	permissionStatus := capabilities.Status(common.CapabilityPermissionReadiness)
	if permissionStatus.Availability != common.AvailabilityAvailable {
		t.Fatalf("expected permission readiness capability to be available, got %s", permissionStatus.Availability)
	}
}

func TestDarwinKeyTapRejectsModifiers(t *testing.T) {
	client := New(Options{})
	if !getPermissionSnapshotOrSkip(t, client).Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}
	if err := client.KeyTap("a", string(common.KeyModifierShift)); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinKeyToggleRejectsModifiers(t *testing.T) {
	client := New(Options{})
	if !getPermissionSnapshotOrSkip(t, client).Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}
	if err := client.KeyToggle("a", common.KeyStateDown, string(common.KeyModifierMeta)); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinKeyActionsRejectUnsafeSpecialKeys(t *testing.T) {
	client := New(Options{})
	if !getPermissionSnapshotOrSkip(t, client).Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	if err := client.KeyTap("audio_vol_up"); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected media key tap to return ErrCapabilityUnavailable, got %v", err)
	}
	if err := client.KeyToggle("printscreen", common.KeyStateDown); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected special key toggle to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinAccessibilityPermissionBlockedOperationsAreDeterministic(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is granted")
	}

	if _, err := client.GetFocusedWindow(); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected focused window metadata to return ErrPermissionDenied, got %v", err)
	}
	if err := client.RaiseFocusedWindow(); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected raiseFocusedWindow to return ErrPermissionDenied, got %v", err)
	}
	if _, err := client.GetFocusedElement(); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected focused element metadata to return ErrPermissionDenied, got %v", err)
	}
	if err := client.PerformFocusedElementAction(common.AXPress); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected performFocusedElementAction to return ErrPermissionDenied, got %v", err)
	}
	if _, err := client.GetElementAtPoint(common.Point{X: 0, Y: 0}); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected element-at-point metadata to return ErrPermissionDenied, got %v", err)
	}
	if err := client.PerformElementActionAtPoint(common.Point{X: 0, Y: 0}, common.AXPress); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected performElementActionAtPoint to return ErrPermissionDenied, got %v", err)
	}
	if err := client.FocusElementAtPoint(common.Point{X: 0, Y: 0}); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected focusElementAtPoint to return ErrPermissionDenied, got %v", err)
	}
	if _, err := client.SearchAXElements(common.AXElementSearchQuery{Scope: common.AXSearchScopeFocusedWindow, Limit: 1, MaxDepth: 0}); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected searchAXElements to return ErrPermissionDenied, got %v", err)
	}
	if err := client.FocusAXElement(common.AXElementRef{Scope: common.AXSearchScopeFocusedWindow}); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected focusAXElement to return ErrPermissionDenied, got %v", err)
	}
	if err := client.PerformAXElementAction(common.AXElementRef{Scope: common.AXSearchScopeFocusedWindow}, common.AXPress); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected performAXElementAction to return ErrPermissionDenied, got %v", err)
	}
	if err := client.MoveMouse(common.Point{X: 1, Y: 1}); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected moveMouse to return ErrPermissionDenied, got %v", err)
	}
	if err := client.KeyTap("a"); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected keyTap to return ErrPermissionDenied, got %v", err)
	}
	if _, err := client.FocusWindow(0); !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected focusWindow to return ErrPermissionDenied, got %v", err)
	}
}

func TestDarwinAccessibilityMetadataMethodsReturnStructuredDataWhenGranted(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	window, err := waitForFocusedWindow(t, client)
	if err != nil {
		t.Fatalf("unexpected focused window error: %v", err)
	}
	if hasFocusedWindowMetadata(window) {
		if !window.Focused {
			t.Fatalf("expected focused window metadata to report Focused=true when a focused AX window is returned: %+v", window)
		}
	} else {
		element, err := waitForFocusedElement(t, client)
		if err != nil {
			t.Fatalf("unexpected focused element error while checking zero focused window metadata: %v", err)
		}
		if hasFocusedElementMetadata(element) {
			t.Fatalf("expected focused window metadata when focused element metadata is available, got zero window metadata")
		}
		t.Log("no focused AX window or focused AX UI element was available on this machine; accepted deterministic zero metadata")
	}

	if err := client.PerformFocusedElementAction(common.AXAction("AXDefinitelyUnsupportedSyntheticAction")); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected unsupported focused-element action to return ErrCapabilityUnavailable, got %v", err)
	}

	element, err := waitForFocusedElement(t, client)
	if err != nil {
		t.Fatalf("unexpected focused element error: %v", err)
	}
	if !hasFocusedElementMetadata(element) && hasFocusedWindowMetadata(window) {
		t.Fatalf("expected focused element metadata when focused window metadata is available, got zero focused element metadata: %+v", window)
	}

	position, err := client.GetMousePosition()
	if err != nil {
		t.Fatalf("unexpected mouse position error: %v", err)
	}
	elementAtPoint, err := client.GetElementAtPoint(position)
	if err != nil {
		t.Fatalf("unexpected element-at-point error: %v", err)
	}
	if elementAtPoint.Role == "" && elementAtPoint.Title == "" && elementAtPoint.Description == "" && elementAtPoint.Value == "" {
		t.Log("no AX element was available at the current mouse position; accepted deterministic zero metadata")
	}

	if err := client.PerformElementActionAtPoint(position, common.AXAction("AXDefinitelyUnsupportedSyntheticAction")); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected unsupported element-at-point action to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinFocusElementAtPointKeepsFocusedLookupLive(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	element, err := waitForFocusedElement(t, client)
	if err != nil {
		t.Fatalf("unexpected focused element error: %v", err)
	}
	if !hasFocusedElementMetadata(element) {
		t.Skip("no focused AX UI element is available on this machine")
	}
	if !element.FrameKnown || element.Frame.Width <= 0 || element.Frame.Height <= 0 {
		t.Skip("focused AX UI element does not expose a usable frame")
	}

	center := common.Point{X: element.Frame.X + element.Frame.Width/2, Y: element.Frame.Y + element.Frame.Height/2}
	if err := client.FocusElementAtPoint(center); err != nil {
		if errors.Is(err, common.ErrCapabilityUnavailable) {
			t.Skip("focused AX UI element at its frame center does not expose a settable AXFocused attribute")
		}
		t.Fatalf("unexpected focusElementAtPoint error: %v", err)
	}

	refocused, err := waitForFocusedElement(t, client)
	if err != nil {
		t.Fatalf("unexpected focused element error after focusElementAtPoint: %v", err)
	}
	if !hasFocusedElementMetadata(refocused) {
		t.Fatal("expected focused element metadata after focusElementAtPoint succeeded")
	}

	if err := client.PerformFocusedElementAction(common.AXAction("AXDefinitelyUnsupportedSyntheticAction")); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected unsupported focused-element action after focusElementAtPoint to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinOffScreenPointOperationsReturnCapabilityUnavailable(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	position := common.Point{X: -1_000_000, Y: -1_000_000}

	if _, err := client.GetElementAtPoint(position); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected off-screen getElementAtPoint to return ErrCapabilityUnavailable, got %v", err)
	}
	if err := client.PerformElementActionAtPoint(position, common.AXPress); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected off-screen performElementActionAtPoint to return ErrCapabilityUnavailable, got %v", err)
	}
	if err := client.FocusElementAtPoint(position); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected off-screen focusElementAtPoint to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinSearchAXElementsReturnsStructuredMatchesOrEmpty(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	matches, err := client.SearchAXElements(common.AXElementSearchQuery{
		Scope:    common.AXSearchScopeFocusedWindow,
		Limit:    5,
		MaxDepth: 2,
	})
	if err != nil {
		t.Fatalf("unexpected searchAXElements error: %v", err)
	}
	for _, match := range matches {
		if match.Ref.Scope != common.AXSearchScopeFocusedWindow {
			t.Fatalf("unexpected match scope: %s", match.Ref.Scope)
		}
		if match.Depth < 0 {
			t.Fatalf("unexpected negative match depth: %d", match.Depth)
		}
		if match.ActionPointKnown && (match.ActionPoint.X == 0 && match.ActionPoint.Y == 0) && match.Metadata.FrameKnown && (match.Metadata.Frame.Width > 0 || match.Metadata.Frame.Height > 0) {
			t.Fatalf("expected structured action point when known, got %+v", match)
		}
		for _, index := range match.Ref.Path {
			if index < 0 {
				t.Fatalf("unexpected negative path index: %+v", match.Ref)
			}
		}
	}
}

func TestDarwinSearchAXElementsImpossibleRoleReturnsEmpty(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	matches, err := client.SearchAXElements(common.AXElementSearchQuery{
		Scope:    common.AXSearchScopeFrontmostApplication,
		Role:     "AXDefinitelyImpossibleSyntheticRole",
		Limit:    5,
		MaxDepth: 3,
	})
	if err != nil {
		t.Fatalf("unexpected searchAXElements error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected impossible role search to return an empty result slice, got %d matches", len(matches))
	}
}

func TestDarwinRefBasedFollowUpOperationsReturnCapabilityUnavailableWhenUnresolved(t *testing.T) {
	client := New(Options{})
	snapshot := getPermissionSnapshotOrSkip(t, client)
	if !snapshot.Accessibility.Granted {
		t.Skip("Accessibility permission is not granted")
	}

	ref := common.AXElementRef{
		Scope:        common.AXSearchScopeFocusedWindow,
		OwnerPID:     1,
		WindowHandle: 1,
		Path:         []int{999999},
	}

	if err := client.FocusAXElement(ref); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected unresolved focusAXElement to return ErrCapabilityUnavailable, got %v", err)
	}
	if err := client.PerformAXElementAction(ref, common.AXPress); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected unresolved performAXElementAction to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestDarwinSearchValidationRejectsInvalidQuery(t *testing.T) {
	client := New(Options{})
	if _, err := client.SearchAXElements(common.AXElementSearchQuery{Scope: common.AXSearchScope("bad"), Limit: 1, MaxDepth: 0}); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid scope to return ErrInvalidToken, got %v", err)
	}
	if _, err := client.SearchAXElements(common.AXElementSearchQuery{Scope: common.AXSearchScopeFocusedWindow, Limit: 0, MaxDepth: 0}); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected zero limit to return ErrInvalidToken, got %v", err)
	}
	if _, err := client.SearchAXElements(common.AXElementSearchQuery{Scope: common.AXSearchScopeFocusedWindow, Limit: 1, MaxDepth: -1}); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected negative max depth to return ErrInvalidToken, got %v", err)
	}
}

func TestDarwinRefValidationRejectsInvalidScope(t *testing.T) {
	client := New(Options{})
	if err := client.FocusAXElement(common.AXElementRef{Scope: common.AXSearchScope("bad")}); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid ref scope to return ErrInvalidToken, got %v", err)
	}
	if err := client.PerformAXElementAction(common.AXElementRef{Scope: common.AXSearchScopeFocusedWindow, Path: []int{-1}}, common.AXPress); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid ref path to return ErrInvalidToken, got %v", err)
	}
}

func TestDarwinCapabilityNotesDescribeSafetyModel(t *testing.T) {
	client := New(Options{})
	capabilities := client.Capabilities()
	snapshot := getPermissionSnapshotOrSkip(t, client)

	keyboardTap := capabilities.Status(common.CapabilityKeyboardTap)
	if snapshot.Accessibility.Granted {
		if keyboardTap.Availability != common.AvailabilityAvailable {
			t.Fatalf("expected keyboard tap to be available, got %s", keyboardTap.Availability)
		}
		if !strings.Contains(keyboardTap.Reason, "modifier chords") {
			t.Fatalf("unexpected keyboard tap reason: %q", keyboardTap.Reason)
		}
	} else if keyboardTap.Availability != common.AvailabilityPermissionBlocked {
		t.Fatalf("expected keyboard tap to be permission-blocked, got %s", keyboardTap.Availability)
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
	if !strings.Contains(joinedNotes, "permission readiness") || !strings.Contains(joinedNotes, "focused-window raise") || !strings.Contains(joinedNotes, "Screen capture is intentionally unavailable") {
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

func hasFocusedWindowMetadata(window common.FocusedWindowMetadata) bool {
	return window.Handle != 0 || window.Title != "" || window.Role != "" || window.Subrole != "" || window.RectKnown || window.OwnerPID != 0 || window.OwnerName != "" || window.BundleID != ""
}

func hasFocusedElementMetadata(element common.UIElementMetadata) bool {
	return element.Role != "" || element.Subrole != "" || element.Title != "" || element.Description != "" || element.Value != "" || element.FrameKnown || len(element.Actions) != 0
}

func waitForFocusedWindow(t *testing.T, client Client) (common.FocusedWindowMetadata, error) {
	t.Helper()
	var (
		window common.FocusedWindowMetadata
		err    error
	)
	for attempt := 0; attempt < 5; attempt++ {
		window, err = client.GetFocusedWindow()
		if err != nil || hasFocusedWindowMetadata(window) || attempt == 4 {
			return window, err
		}
		time.Sleep(30 * time.Millisecond)
	}
	return window, err
}

func waitForFocusedElement(t *testing.T, client Client) (common.UIElementMetadata, error) {
	t.Helper()
	var (
		element common.UIElementMetadata
		err     error
	)
	for attempt := 0; attempt < 5; attempt++ {
		element, err = client.GetFocusedElement()
		if err != nil || hasFocusedElementMetadata(element) || attempt == 4 {
			return element, err
		}
		time.Sleep(30 * time.Millisecond)
	}
	return element, err
}

func getPermissionSnapshotOrSkip(t *testing.T, client Client) common.PermissionSnapshot {
	t.Helper()
	snapshot, err := client.GetPermissionSnapshot()
	if err != nil {
		t.Fatalf("unexpected permission snapshot error: %v", err)
	}
	return snapshot
}
