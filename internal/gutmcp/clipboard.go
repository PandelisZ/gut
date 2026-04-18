package gutmcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerClipboardTools(server *mcp.Server) {
	mcp.AddTool(server, readOnlyTool(
		"clipboard_read",
		"Clipboard Read",
		"Read the current clipboard text content.",
	), s.clipboardReadTool)

	mcp.AddTool(server, mutatingTool(
		"clipboard_write",
		"Clipboard Write",
		"Replace the current clipboard text content.",
	), s.clipboardWriteTool)
}

func (s *Service) clipboardReadTool(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ClipboardReadOutput, error) {
	clipboard, err := s.registry.Clipboard()
	if err != nil {
		return nil, ClipboardReadOutput{}, actionableError("clipboard_read", err)
	}

	text, err := clipboard.Paste(ctx)
	if err != nil {
		return nil, ClipboardReadOutput{}, actionableError("clipboard_read", err)
	}
	return nil, ClipboardReadOutput{
		HasText: len(text) != 0,
		Text:    text,
	}, nil
}

func (s *Service) clipboardWriteTool(ctx context.Context, _ *mcp.CallToolRequest, input ClipboardWriteInput) (*mcp.CallToolResult, ClipboardWriteOutput, error) {
	if err := s.requireMutation("clipboard_write"); err != nil {
		return nil, ClipboardWriteOutput{}, actionableError("clipboard_write", err)
	}

	clipboard, err := s.registry.Clipboard()
	if err != nil {
		return nil, ClipboardWriteOutput{}, actionableError("clipboard_write", err)
	}
	if err := clipboard.Copy(ctx, input.Text); err != nil {
		return nil, ClipboardWriteOutput{}, actionableError("clipboard_write", err)
	}
	return nil, ClipboardWriteOutput{Length: len(input.Text)}, nil
}
