package gutmcp

import (
	"fmt"

	gutpkg "github.com/PandelisZ/gut"
	"github.com/PandelisZ/gut/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const Version = "0.1.0"

type Config struct {
	AllowMutation bool
}

type Service struct {
	registry *provider.Registry
	nut      *gutpkg.Nut
	config   Config
}

func NewService(registry *provider.Registry, config Config) *Service {
	if registry == nil {
		registry = gutpkg.NewDefaultRegistry()
	}
	return &Service{
		registry: registry,
		nut:      gutpkg.New(registry),
		config:   config,
	}
}

func (s *Service) NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gut",
		Version: Version,
	}, nil)

	s.registerTools(server)
	s.registerResources(server)
	return server
}

func (s *Service) registerTools(server *mcp.Server) {
	s.registerStatusTools(server)
	s.registerScreenTools(server)
	s.registerWindowTools(server)
	s.registerAccessibilityTools(server)
	s.registerInputTools(server)
	s.registerClipboardTools(server)
}

func readOnlyTool(name string, title string, description string) *mcp.Tool {
	openWorld := false
	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			Title:         title,
			ReadOnlyHint:  true,
			OpenWorldHint: &openWorld,
		},
	}
}

func mutatingTool(name string, title string, description string) *mcp.Tool {
	openWorld := false
	destructive := true
	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			Title:           title,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		},
	}
}

func (s *Service) requireMutation(action string) error {
	if s.config.AllowMutation {
		return nil
	}
	return fmt.Errorf("%s requires mutation to be enabled; set GUT_MCP_ALLOW_MUTATION=1 or launch gutmcp with --allow-mutation", action)
}
