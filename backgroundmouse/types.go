package backgroundmouse

import (
	"time"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
)

type ActionKind string

const (
	ActionMove        ActionKind = "move"
	ActionClick       ActionKind = "click"
	ActionDoubleClick ActionKind = "double_click"
	ActionFocus       ActionKind = "focus"
	ActionRightClick  ActionKind = "right_click"
	ActionShowMenu    ActionKind = "show_menu"
)

type Config struct {
	MaxWindowElements int
	SnapDistance      int
	DoubleClickGap    time.Duration
}

type SnapshotElement struct {
	Ref                   common.AXElementRef
	Metadata              common.UIElementMetadata
	ActionPoint           shared.Point
	ActionPointKnown      bool
	Depth                 int
	BackgroundSafeActions []ActionKind
}

type WindowSnapshot struct {
	WindowHandle shared.WindowHandle
	WindowRegion shared.Region
	Elements     []SnapshotElement
}

type PointResolution struct {
	RequestedPoint shared.Point
	ScreenPoint    shared.Point
	Snapped        bool
	MatchedElement SnapshotElement
	MatchedRef     common.AXElementRef
	MatchedActions []ActionKind
}

type ActionRequest struct {
	WindowHandle shared.WindowHandle
	Kind         ActionKind
	Point        *shared.Point
	Ref          *common.AXElementRef
}

type SnapshotActionRequest struct {
	Kind  ActionKind
	Point *shared.Point
	Ref   *common.AXElementRef
}

type ActionResult struct {
	Kind            ActionKind
	RequestedPoint  *shared.Point
	ScreenPoint     shared.Point
	Snapped         bool
	MatchedElement  SnapshotElement
	MatchedRef      common.AXElementRef
	MatchedActions  []ActionKind
	PerformedAction common.AXAction
}
