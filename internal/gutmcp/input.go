package gutmcp

import (
	"context"
	"fmt"

	"github.com/PandelisZ/gut/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerInputTools(server *mcp.Server) {
	mcp.AddTool(server, mutatingTool(
		"mouse_action",
		"Mouse Action",
		"Move, click, press, release, scroll, or drag the mouse.",
	), s.mouseActionTool)

	mcp.AddTool(server, mutatingTool(
		"keyboard_action",
		"Keyboard Action",
		"Type text or tap, press, and release keyboard keys.",
	), s.keyboardActionTool)
}

func (s *Service) mouseActionTool(ctx context.Context, _ *mcp.CallToolRequest, input MouseActionInput) (*mcp.CallToolResult, MouseActionOutput, error) {
	if err := s.requireMutation("mouse_action"); err != nil {
		return nil, MouseActionOutput{}, actionableError("mouse_action", err)
	}

	output := MouseActionOutput{Action: input.Kind}
	switch normalizeEnum(input.Kind) {
	case "move":
		if input.Point == nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", fmt.Errorf("point is required for move"))
		}
		if err := s.nut.Mouse.SetPosition(ctx, input.Point.toShared()); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
	case "click":
		button, err := parseMouseButtonOrDefault(input.Button)
		if err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		if err := s.nut.Mouse.Click(ctx, button); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		output.Button = button.String()
	case "doubleclick":
		button, err := parseMouseButtonOrDefault(input.Button)
		if err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		if err := s.nut.Mouse.DoubleClick(ctx, button); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		output.Button = button.String()
	case "press":
		button, err := parseMouseButtonOrDefault(input.Button)
		if err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		if err := s.nut.Mouse.PressButton(ctx, button); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		output.Button = button.String()
	case "release":
		button, err := parseMouseButtonOrDefault(input.Button)
		if err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		if err := s.nut.Mouse.ReleaseButton(ctx, button); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		output.Button = button.String()
	case "scroll":
		if err := performScroll(ctx, s, input.Direction, input.Amount); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
	case "drag":
		path, err := mouseDragPath(ctx, s, input)
		if err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
		if err := s.nut.Mouse.Drag(ctx, path); err != nil {
			return nil, MouseActionOutput{}, actionableError("mouse_action", err)
		}
	default:
		return nil, MouseActionOutput{}, actionableError("mouse_action", fmt.Errorf("unknown mouse action kind %q", input.Kind))
	}

	if position, err := s.nut.Mouse.Position(ctx); err == nil {
		positionJSON := pointToJSON(position)
		output.Position = &positionJSON
	}
	return nil, output, nil
}

func (s *Service) keyboardActionTool(ctx context.Context, _ *mcp.CallToolRequest, input KeyboardActionInput) (*mcp.CallToolResult, KeyboardActionOutput, error) {
	if err := s.requireMutation("keyboard_action"); err != nil {
		return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
	}

	output := KeyboardActionOutput{Action: input.Kind}
	switch normalizeEnum(input.Kind) {
	case "type":
		if input.Text == "" {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", fmt.Errorf("text is required for type"))
		}
		if err := s.nut.Keyboard.TypeText(ctx, input.Text); err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		output.Text = input.Text
	case "tap":
		keys, names, err := parseKeyList(input.Keys)
		if err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		if err := s.nut.Keyboard.Tap(ctx, keys...); err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		output.Keys = names
	case "press":
		keys, names, err := parseKeyList(input.Keys)
		if err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		if err := s.nut.Keyboard.Press(ctx, keys...); err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		output.Keys = names
	case "release":
		keys, names, err := parseKeyList(input.Keys)
		if err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		if err := s.nut.Keyboard.Release(ctx, keys...); err != nil {
			return nil, KeyboardActionOutput{}, actionableError("keyboard_action", err)
		}
		output.Keys = names
	default:
		return nil, KeyboardActionOutput{}, actionableError("keyboard_action", fmt.Errorf("unknown keyboard action kind %q", input.Kind))
	}

	return nil, output, nil
}

func parseMouseButtonOrDefault(name string) (shared.Button, error) {
	if name == "" {
		return shared.ButtonLeft, nil
	}
	return parseButton(name)
}

func parseKeyList(names []string) ([]shared.Key, []string, error) {
	if len(names) == 0 {
		return nil, nil, fmt.Errorf("keys must not be empty")
	}
	keys := make([]shared.Key, 0, len(names))
	canonical := make([]string, 0, len(names))
	for _, name := range names {
		key, err := parseKey(name)
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		canonical = append(canonical, key.String())
	}
	return keys, canonical, nil
}

func performScroll(ctx context.Context, service *Service, direction string, amount int) error {
	if amount <= 0 {
		amount = 1
	}
	switch normalizeEnum(direction) {
	case "up":
		return service.nut.Mouse.ScrollUp(ctx, amount)
	case "down":
		return service.nut.Mouse.ScrollDown(ctx, amount)
	case "left":
		return service.nut.Mouse.ScrollLeft(ctx, amount)
	case "right":
		return service.nut.Mouse.ScrollRight(ctx, amount)
	default:
		return fmt.Errorf("unknown scroll direction %q", direction)
	}
}

func mouseDragPath(ctx context.Context, service *Service, input MouseActionInput) ([]shared.Point, error) {
	if len(input.Path) != 0 {
		path := make([]shared.Point, 0, len(input.Path))
		for _, point := range input.Path {
			path = append(path, point.toShared())
		}
		return path, nil
	}
	if input.Point == nil {
		return nil, fmt.Errorf("point or path is required for drag")
	}
	return service.nut.Mouse.StraightTo(ctx, input.Point.toShared())
}
