package guttesting_test

import (
	"context"
	"testing"

	"github.com/PandelisZ/gut"
	"github.com/PandelisZ/gut/native/common"
	guttesting "github.com/PandelisZ/gut/testing"
)

func TestLiveReadOnlyScreenSizeViaDefaultRegistry(t *testing.T) {
	report := guttesting.Require(t, guttesting.Options{
		RequiredCapabilities: []common.Capability{common.CapabilityScreenSize},
	})

	registry := gut.NewDefaultRegistry()
	screen := gut.NewScreen(registry)

	width, err := screen.Width(context.Background())
	if err != nil {
		t.Fatalf("screen width failed: %v", err)
	}
	height, err := screen.Height(context.Background())
	if err != nil {
		t.Fatalf("screen height failed: %v", err)
	}
	if width <= 0 || height <= 0 {
		t.Fatalf("expected positive screen size, got %dx%d", width, height)
	}

	nativeSize, err := report.Client.GetScreenSize()
	if err != nil {
		t.Fatalf("native screen size failed: %v", err)
	}
	if nativeSize.Width != width || nativeSize.Height != height {
		t.Fatalf("default registry screen size %dx%d did not match native client %dx%d", width, height, nativeSize.Width, nativeSize.Height)
	}
}

func TestLiveReadOnlyWindowEnumerationViaDefaultRegistry(t *testing.T) {
	guttesting.Require(t, guttesting.Options{
		RequiredCapabilities: []common.Capability{common.CapabilityWindowList},
	})

	windows, err := gut.GetWindows(context.Background(), gut.NewDefaultRegistry())
	if err != nil {
		t.Fatalf("window enumeration failed: %v", err)
	}
	if len(windows) == 0 {
		t.Log("window enumeration returned no windows")
		return
	}

	title, err := windows[0].Title(context.Background())
	if err != nil {
		t.Fatalf("window title lookup failed: %v", err)
	}
	t.Logf("enumerated %d window(s); first title length=%d", len(windows), len(title))
}

func TestLiveReadOnlyBackgroundWindowHandleAXSearchViaNativeClient(t *testing.T) {
	report := guttesting.Require(t, guttesting.Options{
		RequiredCapabilities: []common.Capability{
			common.CapabilityWindowList,
			common.CapabilityWindowActive,
			common.CapabilityAXElementSearch,
		},
	})

	active, err := report.Client.GetActiveWindow()
	if err != nil {
		t.Fatalf("active window lookup failed: %v", err)
	}

	windows, err := report.Client.GetWindows()
	if err != nil {
		t.Fatalf("window enumeration failed: %v", err)
	}

	var target common.WindowHandle
	for _, handle := range windows {
		if handle != active {
			target = handle
			break
		}
	}
	if target == 0 {
		t.Skip("no non-active window was available for background AX search verification")
	}

	matches, err := report.Client.SearchAXElements(common.AXElementSearchQuery{
		Scope:        common.AXSearchScopeWindowHandle,
		WindowHandle: target,
		Limit:        10,
		MaxDepth:     2,
	})
	if err != nil {
		t.Fatalf("background window-handle AX search failed for handle %d: %v", target, err)
	}

	for _, match := range matches {
		if match.Ref.Scope != common.AXSearchScopeWindowHandle {
			t.Fatalf("unexpected AX ref scope: %+v", match.Ref)
		}
		if match.Ref.WindowHandle != target {
			t.Fatalf("unexpected AX ref window handle: got %d want %d", match.Ref.WindowHandle, target)
		}
	}

	title, _ := report.Client.GetWindowTitle(target)
	t.Logf("background window-handle AX search succeeded for handle=%d title=%q matches=%d active=%d", target, title, len(matches), active)
}
