package common

import "testing"

func TestCapabilitySetStatusAndSupports(t *testing.T) {
	set := NewCapabilitySet(
		CapabilityStatus{Capability: CapabilityMouseMove, Availability: AvailabilityUnavailable, Reason: "binding unavailable"},
		CapabilityStatus{Capability: CapabilityKeyboardDelay, Availability: AvailabilityAvailable, Reason: "local configuration"},
	)

	if set.Supports(CapabilityMouseMove) {
		t.Fatal("expected unavailable capability to report unsupported")
	}
	if !set.Supports(CapabilityKeyboardDelay) {
		t.Fatal("expected available capability to report supported")
	}

	missing := set.Status(CapabilityScreenCapture)
	if missing.Availability != AvailabilityUnsupported {
		t.Fatalf("expected missing capability to be unsupported, got %s", missing.Availability)
	}
}

func TestCapabilitySetPermissionBlockedIsNotSupported(t *testing.T) {
	set := NewCapabilitySet(
		CapabilityStatus{Capability: CapabilityAXFocusedElementMetadata, Availability: AvailabilityPermissionBlocked, Reason: "Accessibility permission is not granted"},
	)

	if set.Supports(CapabilityAXFocusedElementMetadata) {
		t.Fatal("expected permission-blocked capability to report unsupported")
	}
}

func TestCapabilitySetListIsSorted(t *testing.T) {
	set := NewCapabilitySet(
		CapabilityStatus{Capability: CapabilityWindowResize, Availability: AvailabilityUnavailable},
		CapabilityStatus{Capability: CapabilityKeyboardTap, Availability: AvailabilityUnavailable},
		CapabilityStatus{Capability: CapabilityMouseMove, Availability: AvailabilityUnavailable},
	)

	list := set.List()
	if len(list) != 3 {
		t.Fatalf("unexpected list length: %d", len(list))
	}
	if list[0].Capability != CapabilityKeyboardTap || list[1].Capability != CapabilityMouseMove || list[2].Capability != CapabilityWindowResize {
		t.Fatalf("capabilities are not sorted: %#v", list)
	}
}
