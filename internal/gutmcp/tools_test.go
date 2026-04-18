package gutmcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerListsToolsResourcesAndBlocksMutationWhenDisabled(t *testing.T) {
	deps := newTestDeps(t, false)
	server := deps.service.NewServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.1.0"}, nil)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), serverTransport, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}
	session, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(tools.Tools) != 16 {
		t.Fatalf("unexpected tool count: got %d", len(tools.Tools))
	}

	resources, err := session.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(resources.Resources) != 2 {
		t.Fatalf("unexpected resource count: got %d", len(resources.Resources))
	}

	status, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "status"})
	if err != nil {
		t.Fatalf("status tool failed: %v", err)
	}
	if status.IsError {
		t.Fatalf("status returned tool error: %#v", status)
	}

	windowAction, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "window_action",
		Arguments: map[string]any{"kind": "focus"},
	})
	if err != nil {
		t.Fatalf("window_action call failed: %v", err)
	}
	if !windowAction.IsError {
		t.Fatalf("expected mutation to be blocked")
	}

	resource, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: keysResourceURI})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(resource.Contents) != 1 || resource.Contents[0].Text == "" {
		t.Fatalf("unexpected resource contents: %#v", resource.Contents)
	}
}

func TestToolAnnotationsExposeReadOnlyAndMutatingHints(t *testing.T) {
	deps := newTestDeps(t, true)
	server := deps.service.NewServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.1.0"}, nil)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), serverTransport, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}
	session, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	toolByName := make(map[string]*mcp.Tool, len(tools.Tools))
	for _, tool := range tools.Tools {
		toolByName[tool.Name] = tool
	}

	if !toolByName["status"].Annotations.ReadOnlyHint {
		t.Fatalf("expected status to be marked read-only")
	}
	if toolByName["window_action"].Annotations.DestructiveHint == nil || !*toolByName["window_action"].Annotations.DestructiveHint {
		t.Fatalf("expected window_action destructive hint")
	}
}
