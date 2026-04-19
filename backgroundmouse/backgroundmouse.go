package backgroundmouse

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"time"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
	"github.com/PandelisZ/gut/util"
)

const (
	defaultMaxWindowElements = 400
	defaultSnapDistance      = 32
	defaultDoubleClickGap    = 90 * time.Millisecond
)

var currentGOOS = runtime.GOOS

type BackgroundMouse struct {
	registry *provider.Registry
	config   Config
}

func New(registry *provider.Registry) *BackgroundMouse {
	if registry == nil {
		registry = provider.NewRegistry()
	}
	return &BackgroundMouse{
		registry: registry,
		config: Config{
			MaxWindowElements: defaultMaxWindowElements,
			SnapDistance:      defaultSnapDistance,
			DoubleClickGap:    defaultDoubleClickGap,
		},
	}
}

func (m *BackgroundMouse) SnapshotWindow(ctx context.Context, handle shared.WindowHandle) (WindowSnapshot, error) {
	if err := supportedPlatform(); err != nil {
		return WindowSnapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return WindowSnapshot{}, err
	}
	if handle == 0 {
		return WindowSnapshot{}, fmt.Errorf("window handle is required")
	}

	windowProvider, err := m.registry.Window()
	if err != nil {
		return WindowSnapshot{}, err
	}
	accessibility, err := m.registry.Accessibility()
	if err != nil {
		return WindowSnapshot{}, err
	}

	region, err := windowProvider.GetWindowRegion(ctx, handle)
	if err != nil {
		return WindowSnapshot{}, err
	}

	query := common.AXElementSearchQuery{
		Scope:        common.AXSearchScopeWindowHandle,
		WindowHandle: common.WindowHandle(handle),
		Limit:        m.maxWindowElements(),
		MaxDepth:     m.maxWindowElements(),
	}
	matches, err := accessibility.SearchAXElements(ctx, query)
	if err != nil {
		return WindowSnapshot{}, err
	}

	elements := make([]SnapshotElement, 0, len(matches))
	for _, match := range matches {
		elements = append(elements, SnapshotElement{
			Ref:                   match.Ref,
			Metadata:              match.Metadata,
			ActionPoint:           shared.Point{X: match.ActionPoint.X, Y: match.ActionPoint.Y},
			ActionPointKnown:      match.ActionPointKnown,
			Depth:                 match.Depth,
			BackgroundSafeActions: backgroundSafeActionsForMatch(match.Ref, match.Metadata.Actions),
		})
	}

	return WindowSnapshot{
		WindowHandle: handle,
		WindowRegion: region,
		Elements:     elements,
	}, nil
}

func (m *BackgroundMouse) Resolve(ctx context.Context, handle shared.WindowHandle, point shared.Point) (PointResolution, error) {
	if err := ctx.Err(); err != nil {
		return PointResolution{}, err
	}
	snapshot, err := m.SnapshotWindow(ctx, handle)
	if err != nil {
		return PointResolution{}, err
	}
	return m.ResolveInSnapshot(snapshot, point)
}

func (m *BackgroundMouse) ResolveInSnapshot(snapshot WindowSnapshot, point shared.Point) (PointResolution, error) {
	if err := supportedPlatform(); err != nil {
		return PointResolution{}, err
	}

	screenPoint := translatePoint(snapshot.WindowRegion, point)
	if candidate, ok := bestContainingElement(snapshot.Elements, screenPoint); ok {
		return PointResolution{
			RequestedPoint: point,
			ScreenPoint:    screenPoint,
			Snapped:        false,
			MatchedElement: candidate,
			MatchedRef:     candidate.Ref,
			MatchedActions: append([]ActionKind(nil), candidate.BackgroundSafeActions...),
		}, nil
	}

	candidate, snappedPoint, ok := nearestActionablePoint(snapshot.Elements, screenPoint, m.snapDistance())
	if !ok {
		return PointResolution{}, fmt.Errorf("%w: no background-safe element matched point %s in window %d", ErrUnresolved, point, snapshot.WindowHandle)
	}

	return PointResolution{
		RequestedPoint: point,
		ScreenPoint:    snappedPoint,
		Snapped:        true,
		MatchedElement: candidate,
		MatchedRef:     candidate.Ref,
		MatchedActions: append([]ActionKind(nil), candidate.BackgroundSafeActions...),
	}, nil
}

func (m *BackgroundMouse) Perform(ctx context.Context, req ActionRequest) (ActionResult, error) {
	if err := ctx.Err(); err != nil {
		return ActionResult{}, err
	}
	snapshot, err := m.SnapshotWindow(ctx, req.WindowHandle)
	if err != nil {
		return ActionResult{}, err
	}
	return m.PerformInSnapshot(ctx, snapshot, SnapshotActionRequest{
		Kind:  req.Kind,
		Point: req.Point,
		Ref:   req.Ref,
	})
}

func (m *BackgroundMouse) PerformInSnapshot(ctx context.Context, snapshot WindowSnapshot, req SnapshotActionRequest) (ActionResult, error) {
	if err := supportedPlatform(); err != nil {
		return ActionResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return ActionResult{}, err
	}
	if req.Point == nil && req.Ref == nil {
		return ActionResult{}, fmt.Errorf("point or ref is required")
	}

	accessibility, err := m.registry.Accessibility()
	if err != nil {
		return ActionResult{}, err
	}

	var (
		resolution *PointResolution
		element    SnapshotElement
	)

	if req.Point != nil {
		resolved, err := m.ResolveInSnapshot(snapshot, *req.Point)
		if err != nil {
			return ActionResult{}, err
		}
		resolution = &resolved
		element = resolved.MatchedElement
	} else {
		resolvedElement, err := findElementByRef(snapshot.Elements, *req.Ref)
		if err != nil {
			return ActionResult{}, err
		}
		element = resolvedElement
	}

	performedAction, err := actionToAXAction(req.Kind, element)
	if err != nil {
		return ActionResult{}, err
	}

	switch req.Kind {
	case ActionFocus:
		err = accessibility.FocusAXElement(ctx, element.Ref)
	case ActionDoubleClick:
		err = accessibility.PerformAXElementAction(ctx, element.Ref, performedAction)
		if err == nil {
			err = util.SleepContext(ctx, m.doubleClickGap())
		}
		if err == nil {
			err = accessibility.PerformAXElementAction(ctx, element.Ref, performedAction)
		}
	default:
		err = accessibility.PerformAXElementAction(ctx, element.Ref, performedAction)
	}
	if err != nil {
		return ActionResult{}, err
	}

	result := ActionResult{
		Kind:            req.Kind,
		MatchedElement:  element,
		MatchedRef:      element.Ref,
		MatchedActions:  append([]ActionKind(nil), element.BackgroundSafeActions...),
		PerformedAction: performedAction,
	}
	if req.Point != nil {
		requested := *req.Point
		result.RequestedPoint = &requested
		result.ScreenPoint = resolution.ScreenPoint
		result.Snapped = resolution.Snapped
		return result, nil
	}
	if element.ActionPointKnown {
		result.ScreenPoint = element.ActionPoint
	}
	return result, nil
}

func actionToAXAction(kind ActionKind, element SnapshotElement) (common.AXAction, error) {
	if !slices.Contains(element.BackgroundSafeActions, kind) {
		return "", fmt.Errorf("%w: %s is not available for ref %+v", ErrActionUnsupported, kind, element.Ref)
	}

	switch kind {
	case ActionClick, ActionDoubleClick:
		if action, ok := primaryClickAXAction(element.Metadata.Actions); ok {
			return action, nil
		}
	case ActionFocus:
		return common.AXAction("AXFocused"), nil
	case ActionRightClick, ActionShowMenu:
		return common.AXShowMenu, nil
	}
	return "", fmt.Errorf("%w: unsupported action %s", ErrActionUnsupported, kind)
}

func primaryClickAXAction(actions []string) (common.AXAction, bool) {
	for _, candidate := range []common.AXAction{common.AXPress, common.AXConfirm, common.AXPick} {
		if hasAXAction(actions, candidate) {
			return candidate, true
		}
	}
	return "", false
}

func backgroundSafeActionsForMatch(ref common.AXElementRef, actions []string) []ActionKind {
	result := make([]ActionKind, 0, 5)
	if _, ok := primaryClickAXAction(actions); ok {
		result = append(result, ActionClick, ActionDoubleClick)
	}
	if stringsHasRef(ref) {
		result = append(result, ActionFocus)
	}
	if hasAXAction(actions, common.AXShowMenu) {
		result = append(result, ActionRightClick, ActionShowMenu)
	}
	return slices.Compact(result)
}

func stringsHasRef(ref common.AXElementRef) bool {
	return ref.Scope != ""
}

func hasAXAction(actions []string, action common.AXAction) bool {
	for _, candidate := range actions {
		if candidate == string(action) {
			return true
		}
	}
	return false
}

func findElementByRef(elements []SnapshotElement, ref common.AXElementRef) (SnapshotElement, error) {
	for _, element := range elements {
		if refsEqual(element.Ref, ref) {
			return element, nil
		}
	}
	return SnapshotElement{}, fmt.Errorf("%w: ref %+v was not found in the cached window snapshot", ErrUnresolved, ref)
}

func refsEqual(left, right common.AXElementRef) bool {
	return left.Scope == right.Scope &&
		left.OwnerPID == right.OwnerPID &&
		left.WindowHandle == right.WindowHandle &&
		slices.Equal(left.Path, right.Path)
}

func bestContainingElement(elements []SnapshotElement, point shared.Point) (SnapshotElement, bool) {
	bestIndex := -1
	bestDepth := -1
	bestArea := 0
	for index, element := range elements {
		if !element.Metadata.FrameKnown {
			continue
		}
		frame := regionFromRect(element.Metadata.Frame)
		if !regionContains(frame, point) {
			continue
		}
		area := frame.Area()
		if bestIndex == -1 ||
			element.Depth > bestDepth ||
			(element.Depth == bestDepth && area < bestArea) {
			bestIndex = index
			bestDepth = element.Depth
			bestArea = area
		}
	}
	if bestIndex == -1 {
		return SnapshotElement{}, false
	}
	return elements[bestIndex], true
}

func nearestActionablePoint(elements []SnapshotElement, point shared.Point, maxDistance int) (SnapshotElement, shared.Point, bool) {
	bestIndex := -1
	bestDistance := 0
	bestDepth := -1
	bestArea := 0
	maxDistanceSquared := maxDistance * maxDistance
	for index, element := range elements {
		if !element.ActionPointKnown || len(element.BackgroundSafeActions) == 0 {
			continue
		}
		dx := element.ActionPoint.X - point.X
		dy := element.ActionPoint.Y - point.Y
		distanceSquared := dx*dx + dy*dy
		if distanceSquared > maxDistanceSquared {
			continue
		}
		area := 0
		if element.Metadata.FrameKnown {
			area = regionFromRect(element.Metadata.Frame).Area()
		}
		if bestIndex == -1 ||
			distanceSquared < bestDistance ||
			(distanceSquared == bestDistance && element.Depth > bestDepth) ||
			(distanceSquared == bestDistance && element.Depth == bestDepth && area < bestArea) {
			bestIndex = index
			bestDistance = distanceSquared
			bestDepth = element.Depth
			bestArea = area
		}
	}
	if bestIndex == -1 {
		return SnapshotElement{}, shared.Point{}, false
	}
	return elements[bestIndex], elements[bestIndex].ActionPoint, true
}

func translatePoint(window shared.Region, point shared.Point) shared.Point {
	return shared.Point{X: window.Left + point.X, Y: window.Top + point.Y}
}

func regionContains(region shared.Region, point shared.Point) bool {
	return point.X >= region.Left &&
		point.Y >= region.Top &&
		point.X < region.Left+region.Width &&
		point.Y < region.Top+region.Height
}

func regionFromRect(rect common.Rect) shared.Region {
	return shared.Region{
		Left:   rect.X,
		Top:    rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	}
}

func supportedPlatform() error {
	if currentGOOS != "darwin" {
		return ErrUnsupportedPlatform
	}
	return nil
}

func (m *BackgroundMouse) maxWindowElements() int {
	if m.config.MaxWindowElements > 0 {
		return m.config.MaxWindowElements
	}
	return defaultMaxWindowElements
}

func (m *BackgroundMouse) snapDistance() int {
	if m.config.SnapDistance > 0 {
		return m.config.SnapDistance
	}
	return defaultSnapDistance
}

func (m *BackgroundMouse) doubleClickGap() time.Duration {
	if m.config.DoubleClickGap > 0 {
		return m.config.DoubleClickGap
	}
	return defaultDoubleClickGap
}
