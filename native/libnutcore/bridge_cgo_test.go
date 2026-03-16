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
