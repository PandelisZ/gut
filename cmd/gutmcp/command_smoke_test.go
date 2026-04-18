package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPluginCommandTransportSmoke(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("plugin launcher uses bash; use direct go run configuration on Windows")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	pluginRoot := filepath.Join(repoRoot, "plugins", "gut")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.Command("bash", "./scripts/run-gutmcp.sh")
	cmd.Dir = pluginRoot

	client := mcp.NewClient(&mcp.Implementation{Name: "cmd-gutmcp-test", Version: "v0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{
		Command:           cmd,
		TerminateDuration: 2 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "status"})
	if err != nil {
		t.Fatalf("status call failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("status returned tool error: %#v", result)
	}

	resource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "gut://reference/keys"})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(resource.Contents) != 1 || resource.Contents[0].Text == "" {
		t.Fatalf("unexpected resource contents: %#v", resource.Contents)
	}
}
