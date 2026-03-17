package libnutcore

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"gut/native/common"
)

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()
	if options.MouseDelay != 10*time.Millisecond {
		t.Fatalf("unexpected mouse delay: %s", options.MouseDelay)
	}
	if options.KeyboardDelay != 10*time.Millisecond {
		t.Fatalf("unexpected keyboard delay: %s", options.KeyboardDelay)
	}
}

func TestNewUsesDefaultsAndReportsPlatform(t *testing.T) {
	client := New(Options{})
	info := client.Info()
	wantPlatform := runtime.GOOS
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		wantPlatform = "unsupported"
	}

	if info.Name != BackendName {
		t.Fatalf("unexpected backend name: %s", info.Name)
	}
	if info.Platform != wantPlatform {
		t.Fatalf("unexpected platform: got %s want %s", info.Platform, wantPlatform)
	}
	wantBinding := BindingStateUnavailable
	if client.Capabilities().Status(common.CapabilityMouseMove).Availability == common.AvailabilityAvailable {
		wantBinding = BindingStateLinked
	}
	if info.BindingState != wantBinding {
		t.Fatalf("unexpected binding state: got %s want %s", info.BindingState, wantBinding)
	}
	if len(info.Notes) == 0 {
		t.Fatal("expected backend notes")
	}
}

func TestCapabilitiesReturnsCopy(t *testing.T) {
	client := New(Options{})
	capabilities := client.Capabilities()
	original := capabilities.Status(common.CapabilityMouseMove)
	capabilities[common.CapabilityMouseMove] = common.CapabilityStatus{
		Capability:   common.CapabilityMouseMove,
		Availability: common.AvailabilityAvailable,
	}

	again := client.Capabilities()
	if again[common.CapabilityMouseMove].Availability != original.Availability {
		t.Fatalf("expected original capability set to remain unchanged, got %s", again[common.CapabilityMouseMove].Availability)
	}
}

func TestUnavailableOperationIsStable(t *testing.T) {
	client := New(Options{})
	if client.Capabilities().Status(common.CapabilityMouseMove).Availability == common.AvailabilityAvailable {
		t.Skip("mouse move is available with the linked native backend")
	}

	err := client.MoveMouse(common.Point{X: 10, Y: 20})
	if !errors.Is(err, common.ErrNativeBindingUnavailable) {
		t.Fatalf("expected ErrNativeBindingUnavailable, got %v", err)
	}

	var opErr *common.OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}
	if opErr.Operation != "moveMouse" {
		t.Fatalf("unexpected operation name: %s", opErr.Operation)
	}
}

func TestWindowCapabilityGapsAreDeterministic(t *testing.T) {
	client := New(Options{})
	capabilities := client.Capabilities()
	for _, capability := range []common.Capability{common.CapabilityWindowMinimize, common.CapabilityWindowRestore} {
		status := capabilities.Status(capability)
		if status.Availability != common.AvailabilityUnsupported {
			t.Fatalf("expected %s to be unsupported, got %s", capability, status.Availability)
		}
	}

	if _, err := client.MinimizeWindow(1); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected minimize to return ErrCapabilityUnavailable, got %v", err)
	}
	if _, err := client.RestoreWindow(1); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected restore to return ErrCapabilityUnavailable, got %v", err)
	}
}

func TestAXInteractionCapabilitiesAreDeclared(t *testing.T) {
	linked := linkedCapabilities()
	unavailable := unavailableCapabilities("binding unavailable")

	for _, capability := range []common.Capability{
		common.CapabilityAXFocusedWindowRaise,
		common.CapabilityAXFocusedElementAction,
		common.CapabilityAXElementActionAtPoint,
		common.CapabilityAXElementFocusAtPoint,
	} {
		if status := linked.Status(capability); status.Availability != common.AvailabilityUnsupported {
			t.Fatalf("expected linked %s to be unsupported off-Darwin by default, got %s", capability, status.Availability)
		}
		if status := unavailable.Status(capability); status.Availability != common.AvailabilityUnavailable {
			t.Fatalf("expected unavailable %s to be unavailable, got %s", capability, status.Availability)
		}
	}
}

func TestXDisplayCapabilityMatchesPlatform(t *testing.T) {
	client := New(Options{XDisplayName: ":99"})
	capabilities := client.Capabilities()
	getStatus := capabilities.Status(common.CapabilityX11DisplayGet)
	setStatus := capabilities.Status(common.CapabilityX11DisplaySet)

	if runtime.GOOS == "linux" {
		if getStatus.Availability != common.AvailabilityAvailable || setStatus.Availability != common.AvailabilityAvailable {
			t.Fatalf("expected Linux X11 display configuration to be available, got %s/%s", getStatus.Availability, setStatus.Availability)
		}

		name, err := client.GetXDisplayName()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != ":99" {
			t.Fatalf("unexpected X display name: %q", name)
		}

		if err := client.SetXDisplayName(":100"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		name, err = client.GetXDisplayName()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != ":100" {
			t.Fatalf("unexpected X display name after set: %q", name)
		}
		return
	}

	if getStatus.Availability != common.AvailabilityUnsupported || setStatus.Availability != common.AvailabilityUnsupported {
		t.Fatalf("expected non-Linux X11 display configuration to be unsupported, got %s/%s", getStatus.Availability, setStatus.Availability)
	}
	if _, err := client.GetXDisplayName(); !errors.Is(err, common.ErrUnsupportedPlatform) {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
	if err := client.SetXDisplayName(":100"); !errors.Is(err, common.ErrUnsupportedPlatform) {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}
