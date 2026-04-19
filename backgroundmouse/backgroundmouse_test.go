package backgroundmouse

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
)

type fakeWindowProvider struct {
	regions    map[shared.WindowHandle]shared.Region
	focusCalls []shared.WindowHandle
}

func (f *fakeWindowProvider) GetWindows(context.Context) ([]shared.WindowHandle, error) {
	return nil, nil
}

func (f *fakeWindowProvider) GetActiveWindow(context.Context) (shared.WindowHandle, error) {
	return 0, nil
}

func (f *fakeWindowProvider) GetWindowTitle(context.Context, shared.WindowHandle) (string, error) {
	return "", nil
}

func (f *fakeWindowProvider) GetWindowRegion(_ context.Context, handle shared.WindowHandle) (shared.Region, error) {
	return f.regions[handle], nil
}

func (f *fakeWindowProvider) FocusWindow(_ context.Context, handle shared.WindowHandle) (bool, error) {
	f.focusCalls = append(f.focusCalls, handle)
	return true, nil
}

func (f *fakeWindowProvider) MoveWindow(context.Context, shared.WindowHandle, shared.Point) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) ResizeWindow(context.Context, shared.WindowHandle, shared.Size) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) MinimizeWindow(context.Context, shared.WindowHandle) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) RestoreWindow(context.Context, shared.WindowHandle) (bool, error) {
	return true, nil
}

type performedAXActionCall struct {
	ref    common.AXElementRef
	action common.AXAction
}

type fakeAccessibilityProvider struct {
	searchMatches  []common.AXElementMatch
	focusAXCalls   []common.AXElementRef
	performAXCalls []performedAXActionCall
	searchErr      error
	focusAXErr     error
	performAXErr   error
}

func (f *fakeAccessibilityProvider) GetPermissionSnapshot(context.Context) (common.PermissionSnapshot, error) {
	return common.PermissionSnapshot{}, nil
}

func (f *fakeAccessibilityProvider) GetFocusedWindow(context.Context) (common.FocusedWindowMetadata, error) {
	return common.FocusedWindowMetadata{}, nil
}

func (f *fakeAccessibilityProvider) GetFocusedElement(context.Context) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, nil
}

func (f *fakeAccessibilityProvider) GetElementAtPoint(context.Context, shared.Point) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, nil
}

func (f *fakeAccessibilityProvider) RaiseFocusedWindow(context.Context) error {
	return nil
}

func (f *fakeAccessibilityProvider) PerformFocusedElementAction(context.Context, common.AXAction) error {
	return nil
}

func (f *fakeAccessibilityProvider) PerformElementActionAtPoint(context.Context, shared.Point, common.AXAction) error {
	return nil
}

func (f *fakeAccessibilityProvider) FocusElementAtPoint(context.Context, shared.Point) error {
	return nil
}

func (f *fakeAccessibilityProvider) SearchAXElements(context.Context, common.AXElementSearchQuery) ([]common.AXElementMatch, error) {
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return append([]common.AXElementMatch(nil), f.searchMatches...), nil
}

func (f *fakeAccessibilityProvider) FocusAXElement(_ context.Context, ref common.AXElementRef) error {
	f.focusAXCalls = append(f.focusAXCalls, ref)
	return f.focusAXErr
}

func (f *fakeAccessibilityProvider) PerformAXElementAction(_ context.Context, ref common.AXElementRef, action common.AXAction) error {
	f.performAXCalls = append(f.performAXCalls, performedAXActionCall{ref: ref, action: action})
	return f.performAXErr
}

func (f *fakeAccessibilityProvider) Capabilities() common.CapabilitySet {
	return nil
}

func TestSnapshotWindowCarriesActionPointAndBackgroundSafeActions(t *testing.T) {
	restore := forceDarwinPlatform()
	defer restore()

	registry := provider.NewRegistry()
	registry.RegisterWindow(&fakeWindowProvider{
		regions: map[shared.WindowHandle]shared.Region{
			7: {Left: 300, Top: 120, Width: 800, Height: 600},
		},
	})
	registry.RegisterAccessibility(&fakeAccessibilityProvider{
		searchMatches: []common.AXElementMatch{{
			Ref: common.AXElementRef{
				Scope:        common.AXSearchScopeWindowHandle,
				OwnerPID:     77,
				WindowHandle: 7,
				Path:         []int{0},
			},
			Metadata: common.UIElementMetadata{
				Role:       "AXButton",
				Title:      "Send",
				FrameKnown: true,
				Frame:      common.Rect{X: 1000, Y: 700, Width: 32, Height: 24},
				Actions:    []string{string(common.AXPress), string(common.AXShowMenu)},
			},
			Depth:            2,
			ActionPoint:      common.Point{X: 1016, Y: 712},
			ActionPointKnown: true,
		}},
	})

	background := New(registry)
	snapshot, err := background.SnapshotWindow(context.Background(), 7)
	if err != nil {
		t.Fatalf("SnapshotWindow returned error: %v", err)
	}
	if snapshot.WindowRegion != (shared.Region{Left: 300, Top: 120, Width: 800, Height: 600}) {
		t.Fatalf("unexpected window region: %#v", snapshot.WindowRegion)
	}
	if len(snapshot.Elements) != 1 {
		t.Fatalf("expected 1 snapshot element, got %d", len(snapshot.Elements))
	}
	if snapshot.Elements[0].ActionPoint != (shared.Point{X: 1016, Y: 712}) || !snapshot.Elements[0].ActionPointKnown {
		t.Fatalf("unexpected action point metadata: %#v", snapshot.Elements[0])
	}
	wantActions := []ActionKind{ActionClick, ActionDoubleClick, ActionFocus, ActionRightClick, ActionShowMenu}
	if !reflect.DeepEqual(snapshot.Elements[0].BackgroundSafeActions, wantActions) {
		t.Fatalf("unexpected background-safe actions: got %#v want %#v", snapshot.Elements[0].BackgroundSafeActions, wantActions)
	}
}

func TestResolveInSnapshotPrefersDeepestContainingElement(t *testing.T) {
	restore := forceDarwinPlatform()
	defer restore()

	background := New(nil)
	snapshot := WindowSnapshot{
		WindowHandle: 9,
		WindowRegion: shared.Region{Left: 100, Top: 200, Width: 640, Height: 480},
		Elements: []SnapshotElement{
			{
				Ref:                   common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 9, Path: []int{0}},
				Metadata:              common.UIElementMetadata{FrameKnown: true, Frame: common.Rect{X: 110, Y: 210, Width: 140, Height: 140}},
				Depth:                 1,
				BackgroundSafeActions: []ActionKind{ActionClick},
			},
			{
				Ref:                   common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 9, Path: []int{0, 0}},
				Metadata:              common.UIElementMetadata{FrameKnown: true, Frame: common.Rect{X: 118, Y: 220, Width: 40, Height: 40}},
				Depth:                 3,
				BackgroundSafeActions: []ActionKind{ActionClick},
			},
		},
	}

	result, err := background.ResolveInSnapshot(snapshot, shared.Point{X: 24, Y: 28})
	if err != nil {
		t.Fatalf("ResolveInSnapshot returned error: %v", err)
	}
	if result.Snapped {
		t.Fatalf("expected containing resolution to avoid snapping, got %#v", result)
	}
	if !reflect.DeepEqual(result.MatchedRef.Path, []int{0, 0}) {
		t.Fatalf("expected deepest child to win, got %#v", result.MatchedRef)
	}
	if result.ScreenPoint != (shared.Point{X: 124, Y: 228}) {
		t.Fatalf("unexpected translated point: %#v", result.ScreenPoint)
	}
}

func TestResolveInSnapshotSnapsOnlyWithinDistanceThreshold(t *testing.T) {
	restore := forceDarwinPlatform()
	defer restore()

	background := New(nil)
	snapshot := WindowSnapshot{
		WindowHandle: 3,
		WindowRegion: shared.Region{Left: 100, Top: 200, Width: 300, Height: 200},
		Elements: []SnapshotElement{
			{
				Ref:                   common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 3, Path: []int{0}},
				ActionPoint:           shared.Point{X: 140, Y: 210},
				ActionPointKnown:      true,
				BackgroundSafeActions: []ActionKind{ActionClick},
				Depth:                 1,
			},
		},
	}

	snapped, err := background.ResolveInSnapshot(snapshot, shared.Point{X: 10, Y: 10})
	if err != nil {
		t.Fatalf("ResolveInSnapshot returned error: %v", err)
	}
	if !snapped.Snapped || snapped.ScreenPoint != (shared.Point{X: 140, Y: 210}) {
		t.Fatalf("expected actionable snap to action point, got %#v", snapped)
	}

	if _, err := background.ResolveInSnapshot(snapshot, shared.Point{X: 7, Y: 10}); !errors.Is(err, ErrUnresolved) {
		t.Fatalf("expected point outside snap threshold to fail with ErrUnresolved, got %v", err)
	}
}

func TestPerformInSnapshotMapsActionsWithoutFocusingWindow(t *testing.T) {
	restore := forceDarwinPlatform()
	defer restore()

	window := &fakeWindowProvider{
		regions: map[shared.WindowHandle]shared.Region{
			7: {Left: 300, Top: 120, Width: 800, Height: 600},
		},
	}
	accessibility := &fakeAccessibilityProvider{}
	registry := provider.NewRegistry()
	registry.RegisterWindow(window)
	registry.RegisterAccessibility(accessibility)
	background := New(registry)

	snapshot := WindowSnapshot{
		WindowHandle: 7,
		WindowRegion: shared.Region{Left: 300, Top: 120, Width: 800, Height: 600},
		Elements: []SnapshotElement{{
			Ref: common.AXElementRef{
				Scope:        common.AXSearchScopeWindowHandle,
				OwnerPID:     77,
				WindowHandle: 7,
				Path:         []int{0},
			},
			Metadata: common.UIElementMetadata{
				FrameKnown: true,
				Frame:      common.Rect{X: 1000, Y: 700, Width: 32, Height: 24},
				Actions:    []string{string(common.AXPress), string(common.AXShowMenu)},
			},
			ActionPoint:           shared.Point{X: 1016, Y: 712},
			ActionPointKnown:      true,
			Depth:                 2,
			BackgroundSafeActions: []ActionKind{ActionClick, ActionDoubleClick, ActionFocus, ActionRightClick, ActionShowMenu},
		}},
	}

	clickResult, err := background.PerformInSnapshot(context.Background(), snapshot, SnapshotActionRequest{
		Kind:  ActionClick,
		Point: &shared.Point{X: 710, Y: 590},
	})
	if err != nil {
		t.Fatalf("PerformInSnapshot(click) returned error: %v", err)
	}
	if clickResult.PerformedAction != common.AXPress || len(accessibility.performAXCalls) != 1 {
		t.Fatalf("unexpected click mapping: result=%#v calls=%#v", clickResult, accessibility.performAXCalls)
	}
	if len(window.focusCalls) != 0 {
		t.Fatalf("expected strict background click to avoid FocusWindow, got %+v", window.focusCalls)
	}

	doubleResult, err := background.PerformInSnapshot(context.Background(), snapshot, SnapshotActionRequest{
		Kind: ActionDoubleClick,
		Ref:  &snapshot.Elements[0].Ref,
	})
	if err != nil {
		t.Fatalf("PerformInSnapshot(double_click) returned error: %v", err)
	}
	if doubleResult.PerformedAction != common.AXPress || len(accessibility.performAXCalls) != 3 {
		t.Fatalf("expected double click to emit two AXPress calls, got result=%#v calls=%#v", doubleResult, accessibility.performAXCalls)
	}

	focusResult, err := background.PerformInSnapshot(context.Background(), snapshot, SnapshotActionRequest{
		Kind: ActionFocus,
		Ref:  &snapshot.Elements[0].Ref,
	})
	if err != nil {
		t.Fatalf("PerformInSnapshot(focus) returned error: %v", err)
	}
	if focusResult.PerformedAction != common.AXAction("AXFocused") || len(accessibility.focusAXCalls) != 1 {
		t.Fatalf("unexpected focus mapping: result=%#v calls=%#v", focusResult, accessibility.focusAXCalls)
	}

	menuResult, err := background.PerformInSnapshot(context.Background(), snapshot, SnapshotActionRequest{
		Kind: ActionRightClick,
		Ref:  &snapshot.Elements[0].Ref,
	})
	if err != nil {
		t.Fatalf("PerformInSnapshot(right_click) returned error: %v", err)
	}
	if menuResult.PerformedAction != common.AXShowMenu || len(accessibility.performAXCalls) != 4 {
		t.Fatalf("unexpected show-menu mapping: result=%#v calls=%#v", menuResult, accessibility.performAXCalls)
	}
}

func TestPerformInSnapshotRejectsUnsupportedStrictBackgroundAction(t *testing.T) {
	restore := forceDarwinPlatform()
	defer restore()

	registry := provider.NewRegistry()
	registry.RegisterWindow(&fakeWindowProvider{})
	registry.RegisterAccessibility(&fakeAccessibilityProvider{})
	background := New(registry)

	ref := common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 7, Path: []int{0}}
	snapshot := WindowSnapshot{
		WindowHandle: 7,
		Elements: []SnapshotElement{{
			Ref:                   ref,
			BackgroundSafeActions: []ActionKind{ActionClick, ActionFocus},
		}},
	}

	_, err := background.PerformInSnapshot(context.Background(), snapshot, SnapshotActionRequest{
		Kind: ActionShowMenu,
		Ref:  &ref,
	})
	if !errors.Is(err, ErrActionUnsupported) {
		t.Fatalf("expected unsupported background action, got %v", err)
	}
}

func forceDarwinPlatform() func() {
	previous := currentGOOS
	currentGOOS = "darwin"
	return func() {
		currentGOOS = previous
	}
}
