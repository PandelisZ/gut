//go:build cgo

package libnutcore

import (
	"errors"
	"testing"

	"gut/native/common"
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
