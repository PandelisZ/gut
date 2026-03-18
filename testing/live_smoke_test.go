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
