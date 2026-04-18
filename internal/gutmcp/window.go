package gutmcp

import (
	"context"
	"fmt"

	gutpkg "github.com/PandelisZ/gut"
	"github.com/PandelisZ/gut/shared"
	gutwindow "github.com/PandelisZ/gut/window"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerWindowTools(server *mcp.Server) {
	mcp.AddTool(server, readOnlyTool(
		"window_list",
		"Window List",
		"List the currently visible window handles, titles, and normalized regions.",
	), s.windowListTool)

	mcp.AddTool(server, readOnlyTool(
		"window_active",
		"Active Window",
		"Read the active window handle, title, and region.",
	), s.windowActiveTool)

	mcp.AddTool(server, readOnlyTool(
		"window_elements",
		"Window Elements",
		"Inspect the accessibility element tree for a window or the active window.",
	), s.windowElementsTool)

	mcp.AddTool(server, readOnlyTool(
		"window_find_elements",
		"Find Window Elements",
		"Find exact-match window elements by role, type, title, value, or selected text.",
	), s.windowFindElementsTool)

	mcp.AddTool(server, mutatingTool(
		"window_action",
		"Window Action",
		"Focus, move, resize, minimize, or restore a window or the active window.",
	), s.windowActionTool)
}

func (s *Service) windowListTool(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, WindowListOutput, error) {
	windows, err := gutpkg.GetWindows(ctx, s.registry)
	if err != nil {
		return nil, WindowListOutput{}, actionableError("window_list", err)
	}

	result := make([]JSONWindow, 0, len(windows))
	for _, item := range windows {
		summary, err := s.windowSummary(ctx, item)
		if err != nil {
			return nil, WindowListOutput{}, actionableError("window_list", err)
		}
		result = append(result, summary)
	}

	return nil, WindowListOutput{Windows: result}, nil
}

func (s *Service) windowActiveTool(ctx context.Context, _ *mcp.CallToolRequest, _ WindowSelectorInput) (*mcp.CallToolResult, WindowActiveOutput, error) {
	window, err := gutpkg.GetActiveWindow(ctx, s.registry)
	if err != nil {
		return nil, WindowActiveOutput{}, actionableError("window_active", err)
	}

	summary, err := s.windowSummary(ctx, window)
	if err != nil {
		return nil, WindowActiveOutput{}, actionableError("window_active", err)
	}
	return nil, WindowActiveOutput{Window: summary}, nil
}

func (s *Service) windowElementsTool(ctx context.Context, _ *mcp.CallToolRequest, input WindowElementsInput) (*mcp.CallToolResult, WindowElementsOutput, error) {
	window, err := s.resolveWindow(ctx, input.Handle)
	if err != nil {
		return nil, WindowElementsOutput{}, actionableError("window_elements", err)
	}
	if input.MaxElements < 0 {
		return nil, WindowElementsOutput{}, actionableError("window_elements", fmt.Errorf("maxElements must be non-negative"))
	}

	maxElements := input.MaxElements
	if maxElements <= 0 {
		maxElements = 200
	}
	root, err := window.GetElements(ctx, maxElements)
	if err != nil {
		return nil, WindowElementsOutput{}, actionableError("window_elements", err)
	}

	summary, err := s.windowSummary(ctx, window)
	if err != nil {
		return nil, WindowElementsOutput{}, actionableError("window_elements", err)
	}

	return nil, WindowElementsOutput{
		Window:      summary,
		MaxElements: maxElements,
		Root:        sharedWindowElementToJSON(root),
	}, nil
}

func (s *Service) windowFindElementsTool(ctx context.Context, _ *mcp.CallToolRequest, input WindowFindElementsInput) (*mcp.CallToolResult, WindowFindElementsOutput, error) {
	window, err := s.resolveWindow(ctx, input.Handle)
	if err != nil {
		return nil, WindowFindElementsOutput{}, actionableError("window_find_elements", err)
	}
	if input.Limit < 0 {
		return nil, WindowFindElementsOutput{}, actionableError("window_find_elements", fmt.Errorf("limit must be non-negative"))
	}

	query, err := buildWindowElementQuery(input)
	if err != nil {
		return nil, WindowFindElementsOutput{}, actionableError("window_find_elements", err)
	}

	matches, err := window.FindAll(ctx, query)
	if err != nil {
		return nil, WindowFindElementsOutput{}, actionableError("window_find_elements", err)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	items := make([]map[string]any, 0, min(limit, len(matches)))
	for idx, match := range matches {
		if idx >= limit {
			break
		}
		items = append(items, sharedWindowElementToJSON(match))
	}

	summary, err := s.windowSummary(ctx, window)
	if err != nil {
		return nil, WindowFindElementsOutput{}, actionableError("window_find_elements", err)
	}

	return nil, WindowFindElementsOutput{
		Window:       summary,
		TotalMatches: len(matches),
		Matches:      items,
	}, nil
}

func (s *Service) windowActionTool(ctx context.Context, _ *mcp.CallToolRequest, input WindowActionInput) (*mcp.CallToolResult, WindowActionOutput, error) {
	if err := s.requireMutation("window_action"); err != nil {
		return nil, WindowActionOutput{}, actionableError("window_action", err)
	}

	window, err := s.resolveWindow(ctx, input.Handle)
	if err != nil {
		return nil, WindowActionOutput{}, actionableError("window_action", err)
	}

	var changed bool
	switch normalizeEnum(input.Kind) {
	case "focus":
		changed, err = window.Focus(ctx)
	case "move":
		if input.Origin == nil {
			return nil, WindowActionOutput{}, actionableError("window_action", fmt.Errorf("origin is required for move"))
		}
		changed, err = window.Move(ctx, input.Origin.toShared())
	case "resize":
		if input.Size == nil {
			return nil, WindowActionOutput{}, actionableError("window_action", fmt.Errorf("size is required for resize"))
		}
		changed, err = window.Resize(ctx, input.Size.toShared())
	case "minimize":
		changed, err = window.Minimize(ctx)
	case "restore":
		changed, err = window.Restore(ctx)
	default:
		return nil, WindowActionOutput{}, actionableError("window_action", fmt.Errorf("unknown window action %q", input.Kind))
	}
	if err != nil {
		return nil, WindowActionOutput{}, actionableError("window_action", err)
	}

	summary, err := s.windowSummary(ctx, window)
	if err != nil {
		return nil, WindowActionOutput{}, actionableError("window_action", err)
	}

	return nil, WindowActionOutput{
		Action:  input.Kind,
		Changed: changed,
		Window:  summary,
	}, nil
}

func (s *Service) resolveWindow(ctx context.Context, handle uint64) (*gutwindow.Window, error) {
	if handle != 0 {
		return gutwindow.New(s.registry, shared.WindowHandle(handle)), nil
	}
	return gutpkg.GetActiveWindow(ctx, s.registry)
}

func (s *Service) windowSummary(ctx context.Context, window *gutwindow.Window) (JSONWindow, error) {
	title, err := window.Title(ctx)
	if err != nil {
		return JSONWindow{}, err
	}
	region, err := window.Region(ctx)
	if err != nil {
		return JSONWindow{}, err
	}
	regionJSON := regionToJSON(region)
	return JSONWindow{
		Handle: uint64(window.Handle),
		Title:  title,
		Region: &regionJSON,
	}, nil
}

func buildWindowElementQuery(input WindowFindElementsInput) (shared.WindowElementQuery, error) {
	description := shared.WindowElementDescription{
		Role: input.Role,
		Type: input.Type,
	}
	if input.Title != "" {
		description.Title = matcherPtr(shared.MatchString(input.Title))
	}
	if input.Value != "" {
		description.Value = matcherPtr(shared.MatchString(input.Value))
	}
	if input.SelectedText != "" {
		description.SelectedText = matcherPtr(shared.MatchString(input.SelectedText))
	}
	if !description.Valid() {
		return shared.WindowElementQuery{}, fmt.Errorf("at least one query field must be provided")
	}
	return shared.NewWindowElementQuery("window_find_elements", description), nil
}

func matcherPtr(matcher shared.StringMatcher) *shared.StringMatcher {
	return &matcher
}
