package gutmcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	capabilitiesResourceURI = "gut://reference/capabilities"
	keysResourceURI         = "gut://reference/keys"
)

func (s *Service) registerResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "gut-capabilities",
		Title:       "gut Capabilities Reference",
		Description: "Capability names, status semantics, and default-registry limitations for gut MCP.",
		MIMEType:    "text/markdown",
		URI:         capabilitiesResourceURI,
	}, s.capabilitiesResource)

	server.AddResource(&mcp.Resource{
		Name:        "gut-keys",
		Title:       "gut Keys Reference",
		Description: "Canonical key and mouse button names accepted by gut MCP tools.",
		MIMEType:    "text/markdown",
		URI:         keysResourceURI,
	}, s.keysResource)
}

func (s *Service) capabilitiesResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	_ = ctx
	if req.Params.URI != capabilitiesResourceURI {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	text := strings.TrimSpace(fmt.Sprintf(
		"# gut MCP Capabilities\n\n"+
			"Use the `status` tool before desktop automation. It reports the host capability matrix and permission readiness that the default `gut` registry sees at runtime.\n\n"+
			"## Default registry notes\n\n"+
			"- Keyboard, mouse, screen, window, accessibility, element inspection, clipboard, image read/write, and color finding are wired by default.\n"+
			"- OCR/text finding is not wired by default.\n"+
			"- Image template finding is not wired by default.\n"+
			"- Window title matching is available through the window finder.\n\n"+
			"## Capability names\n\n%s\n",
		bulletList(capabilityNames()),
	))

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      capabilitiesResourceURI,
			MIMEType: "text/markdown",
			Text:     text,
		}},
	}, nil
}

func (s *Service) keysResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	_ = ctx
	if req.Params.URI != keysResourceURI {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	text := strings.TrimSpace(fmt.Sprintf(`
# gut MCP Keys

Keyboard actions accept the canonical key names below. Matching is forgiving: case, spaces, dashes, and underscores are ignored on input.

## Mouse buttons

%s

## Keyboard keys

%s
`, bulletList(canonicalButtonNames()), bulletList(canonicalKeyNames())))

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      keysResourceURI,
			MIMEType: "text/markdown",
			Text:     text,
		}},
	}, nil
}

func capabilityNames() []string {
	names := make([]string, 0, len(allCapabilities()))
	for _, capability := range allCapabilities() {
		names = append(names, string(capability))
	}
	return names
}

func bulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
}
