package gutmcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerAccessibilityTools(server *mcp.Server) {
	mcp.AddTool(server, readOnlyTool(
		"accessibility_snapshot",
		"Accessibility Snapshot",
		"Read desktop accessibility capabilities, permissions, the focused window, and the focused element.",
	), s.accessibilitySnapshotTool)

	mcp.AddTool(server, readOnlyTool(
		"accessibility_search",
		"Accessibility Search",
		"Search accessibility elements by scope, role, text filters, and other metadata.",
	), s.accessibilitySearchTool)

	mcp.AddTool(server, mutatingTool(
		"accessibility_action",
		"Accessibility Action",
		"Raise the focused window or perform accessibility actions against the focused element, a point, or a searched element ref.",
	), s.accessibilityActionTool)
}

func (s *Service) accessibilitySnapshotTool(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, AccessibilitySnapshotOutput, error) {
	output := AccessibilitySnapshotOutput{}

	accessibility, err := s.registry.Accessibility()
	if err != nil {
		message := actionableError("accessibility_snapshot", err).Error()
		output.PermissionsError = message
		output.FocusedWindowError = message
		output.FocusedElementError = message
		return nil, output, nil
	}

	output.Capabilities = capabilityStatusesToJSON(accessibility.Capabilities().List())

	if snapshot, err := accessibility.GetPermissionSnapshot(ctx); err != nil {
		output.PermissionsError = actionableError("accessibility_snapshot", err).Error()
	} else {
		output.Permissions = permissionSnapshotToJSON(snapshot)
	}

	if window, err := accessibility.GetFocusedWindow(ctx); err != nil {
		output.FocusedWindowError = actionableError("accessibility_snapshot", err).Error()
	} else {
		output.FocusedWindow = focusedWindowToJSON(window)
	}

	if element, err := accessibility.GetFocusedElement(ctx); err != nil {
		output.FocusedElementError = actionableError("accessibility_snapshot", err).Error()
	} else {
		output.FocusedElement = uiElementToJSON(element)
	}

	return nil, output, nil
}

func (s *Service) accessibilitySearchTool(ctx context.Context, _ *mcp.CallToolRequest, input AccessibilitySearchInput) (*mcp.CallToolResult, AccessibilitySearchOutput, error) {
	accessibility, err := s.registry.Accessibility()
	if err != nil {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", err)
	}

	if input.Limit < 0 {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", fmt.Errorf("limit must be non-negative"))
	}
	if input.MaxDepth < 0 {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", fmt.Errorf("maxDepth must be non-negative"))
	}

	scope, err := parseAXScope(input.Scope)
	if err != nil {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", err)
	}
	if scope == common.AXSearchScopeWindowHandle && input.WindowHandle == 0 {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", fmt.Errorf("windowHandle is required when scope is window_handle"))
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	maxDepth := input.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 20
	}

	query := common.AXElementSearchQuery{
		Scope:               scope,
		WindowHandle:        common.WindowHandle(input.WindowHandle),
		Role:                input.Role,
		Subrole:             input.Subrole,
		TitleContains:       input.TitleContains,
		ValueContains:       input.ValueContains,
		DescriptionContains: input.DescriptionContains,
		Action:              input.Action,
		Enabled:             input.Enabled,
		Focused:             input.Focused,
		Limit:               limit,
		MaxDepth:            maxDepth,
	}

	matches, err := accessibility.SearchAXElements(ctx, query)
	if err != nil {
		return nil, AccessibilitySearchOutput{}, actionableError("accessibility_search", err)
	}

	items := make([]JSONAXMatch, 0, len(matches))
	for _, match := range matches {
		items = append(items, axMatchToJSON(match))
	}

	output := input
	output.Scope = string(scope)
	output.Limit = limit
	output.MaxDepth = maxDepth

	return nil, AccessibilitySearchOutput{
		Query:   output,
		Matches: items,
	}, nil
}

func (s *Service) accessibilityActionTool(ctx context.Context, _ *mcp.CallToolRequest, input AccessibilityActionInput) (*mcp.CallToolResult, AccessibilityActionOutput, error) {
	if err := s.requireMutation("accessibility_action"); err != nil {
		return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", err)
	}

	accessibility, err := s.registry.Accessibility()
	if err != nil {
		return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", err)
	}

	actionName := normalizeEnum(input.Kind)
	output := AccessibilityActionOutput{Action: input.Kind}

	switch actionName {
	case "raisefocusedwindow":
		err = accessibility.RaiseFocusedWindow(ctx)
		output.Target = "focused_window"
	case "performfocusedelementaction":
		action, parseErr := parseRequiredAXAction(input.Action)
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		err = accessibility.PerformFocusedElementAction(ctx, action)
		output.Target = "focused_element"
	case "performelementactionatpoint":
		action, parseErr := parseRequiredAXAction(input.Action)
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		point, parseErr := requirePoint(input.Point, "point is required for perform_element_action_at_point")
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		err = accessibility.PerformElementActionAtPoint(ctx, point, action)
		output.Target = "point"
		pointJSON := pointToJSON(point)
		output.Point = &pointJSON
	case "focuselementatpoint":
		point, parseErr := requirePoint(input.Point, "point is required for focus_element_at_point")
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		err = accessibility.FocusElementAtPoint(ctx, point)
		output.Target = "point"
		pointJSON := pointToJSON(point)
		output.Point = &pointJSON
	case "focusref":
		ref, parseErr := requireAXRef(input.Ref, "ref is required for focus_ref")
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		err = accessibility.FocusAXElement(ctx, ref)
		output.Target = "ref"
		output.Ref = axRefToJSON(ref)
	case "performrefaction":
		action, parseErr := parseRequiredAXAction(input.Action)
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		ref, parseErr := requireAXRef(input.Ref, "ref is required for perform_ref_action")
		if parseErr != nil {
			return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", parseErr)
		}
		err = accessibility.PerformAXElementAction(ctx, ref, action)
		output.Target = "ref"
		output.Ref = axRefToJSON(ref)
	default:
		return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", fmt.Errorf("unknown accessibility action kind %q", input.Kind))
	}

	if err != nil {
		return nil, AccessibilityActionOutput{}, actionableError("accessibility_action", err)
	}
	return nil, output, nil
}

func parseRequiredAXAction(value string) (common.AXAction, error) {
	if value == "" {
		return "", fmt.Errorf("action is required")
	}
	return parseAXAction(value)
}

func requirePoint(input *PointInput, message string) (shared.Point, error) {
	if input == nil {
		return shared.Point{}, errors.New(message)
	}
	return input.toShared(), nil
}

func requireAXRef(input *AXRefInput, message string) (common.AXElementRef, error) {
	if input == nil {
		return common.AXElementRef{}, errors.New(message)
	}
	ref, err := axRefInputToNative(*input)
	if err != nil {
		return common.AXElementRef{}, err
	}
	if ref.Scope == common.AXSearchScopeWindowHandle && ref.WindowHandle == 0 {
		return common.AXElementRef{}, fmt.Errorf("windowHandle is required when ref.scope is window_handle")
	}
	for _, index := range ref.Path {
		if index < 0 {
			return common.AXElementRef{}, fmt.Errorf("ref.path entries must be non-negative")
		}
	}
	return ref, nil
}
