package gutmcp

import (
	"fmt"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
)

func (p PointInput) toShared() shared.Point {
	return shared.Point{X: p.X, Y: p.Y}
}

func (s SizeInput) toShared() shared.Size {
	return shared.Size{Width: s.Width, Height: s.Height}
}

func (r RegionInput) validate() error {
	if r.Width < 0 || r.Height < 0 {
		return fmt.Errorf("region width and height must be non-negative")
	}
	return nil
}

func (r RegionInput) toShared() shared.Region {
	return shared.Region{Left: r.Left, Top: r.Top, Width: r.Width, Height: r.Height}
}

func (c ColorInput) toShared() (shared.RGBA, error) {
	channels := []struct {
		name  string
		value int
	}{
		{name: "r", value: c.R},
		{name: "g", value: c.G},
		{name: "b", value: c.B},
		{name: "a", value: c.alpha()},
	}
	for _, channel := range channels {
		if channel.value < 0 || channel.value > 255 {
			return shared.RGBA{}, fmt.Errorf("%s must be between 0 and 255", channel.name)
		}
	}
	return shared.RGBA{
		R: uint8(c.R),
		G: uint8(c.G),
		B: uint8(c.B),
		A: uint8(c.alpha()),
	}, nil
}

func (c ColorInput) alpha() int {
	if c.A == nil {
		return 255
	}
	return *c.A
}

func pointToJSON(point shared.Point) JSONPoint {
	return JSONPoint{X: point.X, Y: point.Y}
}

func regionToJSON(region shared.Region) JSONRegion {
	return JSONRegion{
		Left:   region.Left,
		Top:    region.Top,
		Width:  region.Width,
		Height: region.Height,
	}
}

func colorToJSON(color shared.RGBA) JSONColor {
	return JSONColor{
		R:   color.R,
		G:   color.G,
		B:   color.B,
		A:   color.A,
		Hex: color.Hex(),
	}
}

func capabilityStatusesToJSON(statuses []common.CapabilityStatus) []JSONCapabilityStatus {
	result := make([]JSONCapabilityStatus, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, JSONCapabilityStatus{
			Capability:   string(status.Capability),
			Availability: string(status.Availability),
			Reason:       status.Reason,
		})
	}
	return result
}

func permissionSnapshotToJSON(snapshot common.PermissionSnapshot) *JSONPermissionSnapshot {
	return &JSONPermissionSnapshot{
		Accessibility: JSONPermissionStatus{
			Granted:   snapshot.Accessibility.Granted,
			Supported: snapshot.Accessibility.Supported,
			Reason:    snapshot.Accessibility.Reason,
		},
		ScreenRecording: JSONPermissionStatus{
			Granted:   snapshot.ScreenRecording.Granted,
			Supported: snapshot.ScreenRecording.Supported,
			Reason:    snapshot.ScreenRecording.Reason,
		},
	}
}

func windowToJSON(window JSONWindow) JSONWindow {
	return window
}

func sharedWindowElementToJSON(element shared.WindowElement) map[string]any {
	result := map[string]any{}
	if element.Type != nil {
		result["type"] = *element.Type
	}
	if element.Title != nil {
		result["title"] = *element.Title
	}
	if element.Value != nil {
		result["value"] = *element.Value
	}
	if element.IsFocused != nil {
		result["isFocused"] = *element.IsFocused
	}
	if element.SelectedText != nil {
		result["selectedText"] = *element.SelectedText
	}
	if element.IsEnabled != nil {
		result["isEnabled"] = *element.IsEnabled
	}
	if element.Role != nil {
		result["role"] = *element.Role
	}
	if element.SubRole != nil {
		result["subRole"] = *element.SubRole
	}
	if element.Region != nil {
		result["region"] = regionToJSON(*element.Region)
	}
	if len(element.Children) != 0 {
		children := make([]map[string]any, 0, len(element.Children))
		for _, child := range element.Children {
			children = append(children, sharedWindowElementToJSON(child))
		}
		result["children"] = children
	}
	return result
}

func focusedWindowToJSON(window common.FocusedWindowMetadata) *JSONFocusedWindowMetadata {
	result := &JSONFocusedWindowMetadata{
		Handle:    uint64(window.Handle),
		Title:     window.Title,
		Role:      window.Role,
		Subrole:   window.Subrole,
		Focused:   window.Focused,
		Main:      window.Main,
		Minimized: window.Minimized,
		OwnerPID:  window.OwnerPID,
		OwnerName: window.OwnerName,
		BundleID:  window.BundleID,
	}
	if window.RectKnown {
		rect := regionFromNative(window.Rect)
		region := regionToJSON(rect)
		result.Rect = &region
	}
	return result
}

func uiElementToJSON(element common.UIElementMetadata) *JSONUIElementMetadata {
	result := &JSONUIElementMetadata{
		Role:        element.Role,
		Subrole:     element.Subrole,
		Title:       element.Title,
		Description: element.Description,
		Value:       element.Value,
		Enabled:     element.Enabled,
		Focused:     element.Focused,
		Actions:     append([]string(nil), element.Actions...),
	}
	if element.FrameKnown {
		frame := regionFromNative(element.Frame)
		region := regionToJSON(frame)
		result.Frame = &region
	}
	return result
}

func axRefToJSON(ref common.AXElementRef) *JSONAXRef {
	return &JSONAXRef{
		Scope:        string(ref.Scope),
		OwnerPID:     ref.OwnerPID,
		WindowHandle: uint64(ref.WindowHandle),
		Path:         append([]int(nil), ref.Path...),
	}
}

func axMatchToJSON(match common.AXElementMatch) JSONAXMatch {
	result := JSONAXMatch{
		Ref:              *axRefToJSON(match.Ref),
		Metadata:         *uiElementToJSON(match.Metadata),
		Depth:            match.Depth,
		ActionPointKnown: match.ActionPointKnown,
	}
	if match.ActionPointKnown {
		point := pointToJSON(shared.Point{X: match.ActionPoint.X, Y: match.ActionPoint.Y})
		result.ActionPoint = &point
	}
	return result
}

func regionFromNative(rect common.Rect) shared.Region {
	return shared.Region{
		Left:   rect.X,
		Top:    rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	}
}

func axRefInputToNative(ref AXRefInput) (common.AXElementRef, error) {
	scope, err := parseAXScope(ref.Scope)
	if err != nil {
		return common.AXElementRef{}, err
	}
	return common.AXElementRef{
		Scope:        scope,
		OwnerPID:     ref.OwnerPID,
		WindowHandle: common.WindowHandle(ref.WindowHandle),
		Path:         append([]int(nil), ref.Path...),
	}, nil
}
