//go:build cgo

package libnutcore

import (
	"errors"
	"testing"

	"github.com/PandelisZ/gut/native/common"
)

func TestBridgeStatusErrorCapabilityUnavailable(t *testing.T) {
	err := bridgeStatusError("captureScreen", common.CapabilityScreenCapture, 4)
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}

	var opErr *common.OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}
	if opErr.Capability != common.CapabilityScreenCapture {
		t.Fatalf("unexpected capability: %s", opErr.Capability)
	}
}

func TestBridgeStatusErrorPermissionDenied(t *testing.T) {
	err := bridgeStatusError("focusWindow", common.CapabilityWindowFocus, 5)
	if !errors.Is(err, common.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestBridgeStatusErrorNoElementAtPoint(t *testing.T) {
	err := bridgeStatusError("getElementAtPoint", common.CapabilityAXElementAtPointMetadata, 6)
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}

	var opErr *common.OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}
	if opErr.Detail != "no AX element is available at the requested point" {
		t.Fatalf("unexpected detail: %q", opErr.Detail)
	}
}

func TestValidateAXActionRejectsEmpty(t *testing.T) {
	err := validateAXAction("", "performFocusedElementAction", common.CapabilityAXFocusedElementAction)
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAXElementSearchQueryRejectsInvalidScope(t *testing.T) {
	err := validateAXElementSearchQuery(common.AXElementSearchQuery{Scope: common.AXSearchScope("nope"), Limit: 1, MaxDepth: 0})
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAXElementSearchQueryRejectsNonPositiveLimit(t *testing.T) {
	err := validateAXElementSearchQuery(common.AXElementSearchQuery{Scope: common.AXSearchScopeFocusedWindow, Limit: 0, MaxDepth: 0})
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAXElementSearchQueryRejectsNegativeMaxDepth(t *testing.T) {
	err := validateAXElementSearchQuery(common.AXElementSearchQuery{Scope: common.AXSearchScopeFocusedWindow, Limit: 1, MaxDepth: -1})
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAXElementRefRejectsInvalidScope(t *testing.T) {
	err := validateAXElementRef(common.AXElementRef{Scope: common.AXSearchScope("nope")}, "focusAXElement", common.CapabilityAXElementFocusMatch)
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAXElementRefRejectsNegativePathIndex(t *testing.T) {
	err := validateAXElementRef(common.AXElementRef{Scope: common.AXSearchScopeFocusedWindow, Path: []int{-1}}, "focusAXElement", common.CapabilityAXElementFocusMatch)
	if !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestBoolPointerToTristate(t *testing.T) {
	if got := boolPointerToTristate(nil); got != -1 {
		t.Fatalf("expected nil bool to map to -1, got %d", got)
	}
	value := false
	if got := boolPointerToTristate(&value); got != 0 {
		t.Fatalf("expected false bool to map to 0, got %d", got)
	}
	value = true
	if got := boolPointerToTristate(&value); got != 1 {
		t.Fatalf("expected true bool to map to 1, got %d", got)
	}
}

func TestBridgeAXInteractionErrorCapabilityUnavailable(t *testing.T) {
	err := bridgeAXInteractionError("focusElementAtPoint", common.CapabilityAXElementFocusAtPoint, 4, "no AX element is available at the requested point or it does not expose a settable AXFocused attribute")
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}

	var opErr *common.OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OperationError, got %T", err)
	}
	if opErr.Capability != common.CapabilityAXElementFocusAtPoint {
		t.Fatalf("unexpected capability: %s", opErr.Capability)
	}
}

func TestBridgeAXInteractionErrorNoElementAtPoint(t *testing.T) {
	err := bridgeAXInteractionError("performElementActionAtPoint", common.CapabilityAXElementActionAtPoint, 6, "no AX element is available at the requested point or it does not support the requested AX action")
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
}
