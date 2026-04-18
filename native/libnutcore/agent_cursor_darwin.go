//go:build darwin && cgo

package libnutcore

/*
#include "bridge_shim.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/PandelisZ/gut/native/common"
)

func ShowAgentCursor(event AgentCursorEvent) error {
	kind := strings.TrimSpace(string(event.Kind))
	if kind == "" {
		return fmt.Errorf("%w: showAgentCursor [%s]", common.ErrInvalidToken, common.CapabilityScreenHighlight)
	}

	cKind := C.CString(kind)
	defer C.free(unsafe.Pointer(cKind))

	cButton := cStringOrNil(string(event.Button))
	if cButton != nil {
		defer C.free(unsafe.Pointer(cButton))
	}

	cDirection := cStringOrNil(string(event.Direction))
	if cDirection != nil {
		defer C.free(unsafe.Pointer(cDirection))
	}

	var hasTarget C.int
	var targetX C.int64_t
	var targetY C.int64_t
	if event.Target != nil {
		hasTarget = 1
		targetX = C.int64_t(event.Target.X)
		targetY = C.int64_t(event.Target.Y)
	}

	return bridgeStatusError("showAgentCursor", common.CapabilityScreenHighlight, int(C.gut_show_agent_cursor(
		cKind,
		C.int64_t(event.Position.X),
		C.int64_t(event.Position.Y),
		hasTarget,
		targetX,
		targetY,
		cButton,
		cDirection,
		boolToCInt(event.Pressed),
		C.int64_t(event.Duration.Milliseconds()),
	)))
}

func HideAgentCursor() error {
	return bridgeStatusError("hideAgentCursor", common.CapabilityScreenHighlight, int(C.gut_hide_agent_cursor()))
}

func cStringOrNil(value string) *C.char {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return C.CString(value)
}

func boolToCInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}
