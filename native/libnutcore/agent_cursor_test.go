package libnutcore

import (
	"testing"
	"time"

	"github.com/PandelisZ/gut/native/common"
)

func TestAgentCursorAPIsAreStable(t *testing.T) {
	target := common.Point{X: 24, Y: 32}
	if err := ShowAgentCursor(AgentCursorEvent{
		Kind:     AgentCursorEventMove,
		Position: common.Point{X: 12, Y: 16},
		Target:   &target,
		Button:   common.MouseButtonLeft,
		Duration: 10 * time.Millisecond,
	}); err != nil {
		t.Fatalf("ShowAgentCursor returned error: %v", err)
	}
	if err := HideAgentCursor(); err != nil {
		t.Fatalf("HideAgentCursor returned error: %v", err)
	}
}
