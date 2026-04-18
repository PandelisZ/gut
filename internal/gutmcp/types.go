package gutmcp

import guttesting "github.com/PandelisZ/gut/testing"

type PointInput struct {
	X int `json:"x" jsonschema:"x coordinate"`
	Y int `json:"y" jsonschema:"y coordinate"`
}

type SizeInput struct {
	Width  int `json:"width" jsonschema:"width in pixels"`
	Height int `json:"height" jsonschema:"height in pixels"`
}

type RegionInput struct {
	Left   int `json:"left" jsonschema:"left origin in pixels"`
	Top    int `json:"top" jsonschema:"top origin in pixels"`
	Width  int `json:"width" jsonschema:"region width in pixels"`
	Height int `json:"height" jsonschema:"region height in pixels"`
}

type ColorInput struct {
	R int  `json:"r" jsonschema:"red channel 0-255"`
	G int  `json:"g" jsonschema:"green channel 0-255"`
	B int  `json:"b" jsonschema:"blue channel 0-255"`
	A *int `json:"a,omitempty" jsonschema:"alpha channel 0-255; defaults to 255 when omitted"`
}

type AXRefInput struct {
	Scope        string `json:"scope" jsonschema:"focused_window, frontmost_application, or window_handle"`
	OwnerPID     int    `json:"ownerPid,omitempty" jsonschema:"owning process ID when available"`
	WindowHandle uint64 `json:"windowHandle,omitempty" jsonschema:"window handle for window_handle scope"`
	Path         []int  `json:"path,omitempty" jsonschema:"tree path inside the accessibility hierarchy"`
}

type JSONPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type JSONSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type JSONRegion struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type JSONPixelDensity struct {
	ScaleX float64 `json:"scaleX"`
	ScaleY float64 `json:"scaleY"`
}

type JSONColor struct {
	R   uint8  `json:"r"`
	G   uint8  `json:"g"`
	B   uint8  `json:"b"`
	A   uint8  `json:"a"`
	Hex string `json:"hex"`
}

type JSONCapabilityStatus struct {
	Capability   string `json:"capability"`
	Availability string `json:"availability"`
	Reason       string `json:"reason,omitempty"`
}

type JSONPermissionStatus struct {
	Granted   bool   `json:"granted"`
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

type JSONPermissionSnapshot struct {
	Accessibility   JSONPermissionStatus `json:"accessibility"`
	ScreenRecording JSONPermissionStatus `json:"screenRecording"`
}

type JSONWindow struct {
	Handle uint64      `json:"handle"`
	Title  string      `json:"title"`
	Region *JSONRegion `json:"region,omitempty"`
}

type JSONUIElementMetadata struct {
	Role        string      `json:"role,omitempty"`
	Subrole     string      `json:"subrole,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       string      `json:"value,omitempty"`
	Enabled     bool        `json:"enabled"`
	Focused     bool        `json:"focused"`
	Frame       *JSONRegion `json:"frame,omitempty"`
	Actions     []string    `json:"actions,omitempty"`
}

type JSONFocusedWindowMetadata struct {
	Handle    uint64      `json:"handle"`
	Title     string      `json:"title,omitempty"`
	Role      string      `json:"role,omitempty"`
	Subrole   string      `json:"subrole,omitempty"`
	Rect      *JSONRegion `json:"rect,omitempty"`
	Focused   bool        `json:"focused"`
	Main      bool        `json:"main"`
	Minimized bool        `json:"minimized"`
	OwnerPID  int         `json:"ownerPid,omitempty"`
	OwnerName string      `json:"ownerName,omitempty"`
	BundleID  string      `json:"bundleId,omitempty"`
}

type JSONAXRef struct {
	Scope        string `json:"scope"`
	OwnerPID     int    `json:"ownerPid,omitempty"`
	WindowHandle uint64 `json:"windowHandle,omitempty"`
	Path         []int  `json:"path,omitempty"`
}

type JSONAXMatch struct {
	Ref              JSONAXRef             `json:"ref"`
	Metadata         JSONUIElementMetadata `json:"metadata"`
	Depth            int                   `json:"depth"`
	ActionPoint      *JSONPoint            `json:"actionPoint,omitempty"`
	ActionPointKnown bool                  `json:"actionPointKnown"`
}

type StatusOutput struct {
	ServerVersion   string                       `json:"serverVersion"`
	MutationAllowed bool                         `json:"mutationAllowed"`
	Environment     guttesting.EnvironmentReport `json:"environment"`
	Permissions     *JSONPermissionSnapshot      `json:"permissions,omitempty"`
	PermissionError string                       `json:"permissionError,omitempty"`
}

type ScreenCaptureInput struct {
	Region *RegionInput `json:"region,omitempty" jsonschema:"optional logical capture region"`
}

type ScreenCaptureOutput struct {
	Format         string           `json:"format"`
	CapturedRegion JSONRegion       `json:"capturedRegion"`
	PixelSize      JSONSize         `json:"pixelSize"`
	PixelDensity   JSONPixelDensity `json:"pixelDensity"`
}

type ScreenColorAtInput struct {
	Point PointInput `json:"point"`
}

type ScreenColorAtOutput struct {
	Point JSONPoint `json:"point"`
	Color JSONColor `json:"color"`
}

type ScreenFindColorInput struct {
	Color        ColorInput   `json:"color"`
	SearchRegion *RegionInput `json:"searchRegion,omitempty"`
	Limit        int          `json:"limit,omitempty" jsonschema:"maximum number of points to return; defaults to 20"`
}

type ScreenFindColorOutput struct {
	Color        JSONColor   `json:"color"`
	SearchRegion *JSONRegion `json:"searchRegion,omitempty"`
	TotalMatches int         `json:"totalMatches"`
	Matches      []JSONPoint `json:"matches"`
}

type WindowListOutput struct {
	Windows []JSONWindow `json:"windows"`
}

type WindowSelectorInput struct {
	Handle uint64 `json:"handle,omitempty" jsonschema:"window handle; defaults to the active window when omitted"`
}

type WindowActiveOutput struct {
	Window JSONWindow `json:"window"`
}

type WindowElementsInput struct {
	Handle      uint64 `json:"handle,omitempty" jsonschema:"window handle; defaults to the active window when omitted"`
	MaxElements int    `json:"maxElements,omitempty" jsonschema:"maximum number of AX elements to inspect; defaults to 200"`
}

type WindowElementsOutput struct {
	Window      JSONWindow     `json:"window"`
	MaxElements int            `json:"maxElements"`
	Root        map[string]any `json:"root"`
}

type WindowFindElementsInput struct {
	Handle       uint64 `json:"handle,omitempty" jsonschema:"window handle; defaults to the active window when omitted"`
	Role         string `json:"role,omitempty" jsonschema:"exact role match"`
	Type         string `json:"type,omitempty" jsonschema:"exact type match"`
	Title        string `json:"title,omitempty" jsonschema:"exact title match"`
	Value        string `json:"value,omitempty" jsonschema:"exact value match"`
	SelectedText string `json:"selectedText,omitempty" jsonschema:"exact selected text match"`
	Limit        int    `json:"limit,omitempty" jsonschema:"maximum matches to return; defaults to 20"`
}

type WindowFindElementsOutput struct {
	Window       JSONWindow       `json:"window"`
	TotalMatches int              `json:"totalMatches"`
	Matches      []map[string]any `json:"matches"`
}

type WindowActionInput struct {
	Handle uint64      `json:"handle,omitempty" jsonschema:"window handle; defaults to the active window when omitted"`
	Kind   string      `json:"kind" jsonschema:"focus, move, resize, minimize, or restore"`
	Origin *PointInput `json:"origin,omitempty" jsonschema:"required for move"`
	Size   *SizeInput  `json:"size,omitempty" jsonschema:"required for resize"`
}

type WindowActionOutput struct {
	Action  string     `json:"action"`
	Changed bool       `json:"changed"`
	Window  JSONWindow `json:"window"`
}

type AccessibilitySnapshotOutput struct {
	Capabilities        []JSONCapabilityStatus     `json:"capabilities"`
	Permissions         *JSONPermissionSnapshot    `json:"permissions,omitempty"`
	PermissionsError    string                     `json:"permissionsError,omitempty"`
	FocusedWindow       *JSONFocusedWindowMetadata `json:"focusedWindow,omitempty"`
	FocusedWindowError  string                     `json:"focusedWindowError,omitempty"`
	FocusedElement      *JSONUIElementMetadata     `json:"focusedElement,omitempty"`
	FocusedElementError string                     `json:"focusedElementError,omitempty"`
}

type AccessibilitySearchInput struct {
	Scope               string `json:"scope" jsonschema:"focused_window, frontmost_application, or window_handle"`
	WindowHandle        uint64 `json:"windowHandle,omitempty" jsonschema:"window handle when scope is window_handle"`
	Role                string `json:"role,omitempty" jsonschema:"role filter"`
	Subrole             string `json:"subrole,omitempty" jsonschema:"subrole filter"`
	TitleContains       string `json:"titleContains,omitempty" jsonschema:"substring filter on title"`
	ValueContains       string `json:"valueContains,omitempty" jsonschema:"substring filter on value"`
	DescriptionContains string `json:"descriptionContains,omitempty" jsonschema:"substring filter on description"`
	Action              string `json:"action,omitempty" jsonschema:"required action name filter"`
	Enabled             *bool  `json:"enabled,omitempty"`
	Focused             *bool  `json:"focused,omitempty"`
	Limit               int    `json:"limit,omitempty" jsonschema:"maximum results; defaults to 20"`
	MaxDepth            int    `json:"maxDepth,omitempty" jsonschema:"maximum traversal depth; defaults to 20"`
}

type AccessibilitySearchOutput struct {
	Query   AccessibilitySearchInput `json:"query"`
	Matches []JSONAXMatch            `json:"matches"`
}

type AccessibilityActionInput struct {
	Kind   string      `json:"kind" jsonschema:"raise_focused_window, perform_focused_element_action, perform_element_action_at_point, focus_element_at_point, focus_ref, or perform_ref_action"`
	Action string      `json:"action,omitempty" jsonschema:"AXPress, AXRaise, AXShowMenu, AXConfirm, or AXPick when required"`
	Point  *PointInput `json:"point,omitempty" jsonschema:"required for point-targeted actions"`
	Ref    *AXRefInput `json:"ref,omitempty" jsonschema:"required for ref-targeted actions"`
}

type AccessibilityActionOutput struct {
	Action string     `json:"action"`
	Target string     `json:"target"`
	Ref    *JSONAXRef `json:"ref,omitempty"`
	Point  *JSONPoint `json:"point,omitempty"`
}

type MouseActionInput struct {
	Kind      string       `json:"kind" jsonschema:"move, click, double_click, press, release, scroll, or drag"`
	Button    string       `json:"button,omitempty" jsonschema:"left, middle, or right"`
	Point     *PointInput  `json:"point,omitempty" jsonschema:"target point for move or drag"`
	Path      []PointInput `json:"path,omitempty" jsonschema:"explicit drag path"`
	Direction string       `json:"direction,omitempty" jsonschema:"up, down, left, or right for scroll"`
	Amount    int          `json:"amount,omitempty" jsonschema:"scroll amount; defaults to 1"`
}

type MouseActionOutput struct {
	Action   string     `json:"action"`
	Button   string     `json:"button,omitempty"`
	Position *JSONPoint `json:"position,omitempty"`
}

type KeyboardActionInput struct {
	Kind string   `json:"kind" jsonschema:"type, tap, press, or release"`
	Text string   `json:"text,omitempty" jsonschema:"text to type when kind is type"`
	Keys []string `json:"keys,omitempty" jsonschema:"key names for tap, press, or release"`
}

type KeyboardActionOutput struct {
	Action string   `json:"action"`
	Text   string   `json:"text,omitempty"`
	Keys   []string `json:"keys,omitempty"`
}

type ClipboardReadOutput struct {
	HasText bool   `json:"hasText"`
	Text    string `json:"text"`
}

type ClipboardWriteInput struct {
	Text string `json:"text" jsonschema:"clipboard text to write"`
}

type ClipboardWriteOutput struct {
	Length int `json:"length"`
}
