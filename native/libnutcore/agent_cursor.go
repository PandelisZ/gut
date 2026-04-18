package libnutcore

import (
	"time"

	"github.com/PandelisZ/gut/native/common"
)

type AgentCursorEventKind string

const (
	AgentCursorEventMove        AgentCursorEventKind = "move"
	AgentCursorEventClick       AgentCursorEventKind = "click"
	AgentCursorEventDoubleClick AgentCursorEventKind = "double_click"
	AgentCursorEventDragStart   AgentCursorEventKind = "drag_start"
	AgentCursorEventDragEnd     AgentCursorEventKind = "drag_end"
	AgentCursorEventScroll      AgentCursorEventKind = "scroll"
	AgentCursorEventMouseDown   AgentCursorEventKind = "mouse_down"
	AgentCursorEventMouseUp     AgentCursorEventKind = "mouse_up"
	AgentCursorEventHide        AgentCursorEventKind = "hide"
)

type AgentCursorDirection string

const (
	AgentCursorDirectionUp    AgentCursorDirection = "up"
	AgentCursorDirectionDown  AgentCursorDirection = "down"
	AgentCursorDirectionLeft  AgentCursorDirection = "left"
	AgentCursorDirectionRight AgentCursorDirection = "right"
)

type AgentCursorEvent struct {
	Kind      AgentCursorEventKind
	Position  common.Point
	Target    *common.Point
	Button    common.MouseButton
	Direction AgentCursorDirection
	Pressed   bool
	Duration  time.Duration
}
