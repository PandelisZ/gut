package gutmcp

import (
	"context"
	"os"

	guttesting "github.com/PandelisZ/gut/testing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerStatusTools(server *mcp.Server) {
	mcp.AddTool(server, readOnlyTool(
		"status",
		"Status",
		"Inspect gut MCP runtime status, capability availability, and desktop permission readiness.",
	), s.statusTool)
}

func (s *Service) statusTool(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, StatusOutput, error) {
	output := StatusOutput{
		ServerVersion:   Version,
		MutationAllowed: s.config.AllowMutation,
		Environment:     s.evaluateEnvironment(),
	}

	accessibility, err := s.registry.Accessibility()
	if err != nil {
		output.PermissionError = actionableError("status", err).Error()
		return nil, output, nil
	}

	snapshot, err := accessibility.GetPermissionSnapshot(ctx)
	if err != nil {
		output.PermissionError = actionableError("status", err).Error()
		return nil, output, nil
	}

	output.Permissions = permissionSnapshotToJSON(snapshot)
	return nil, output, nil
}

func (s *Service) evaluateEnvironment() guttesting.EnvironmentReport {
	lookupEnv := func(key string) (string, bool) {
		switch key {
		case guttesting.EnvEnableLiveTests:
			return "1", true
		case guttesting.EnvEnableMutationTests:
			if s.config.AllowMutation {
				return "1", true
			}
			return "", false
		default:
			return os.LookupEnv(key)
		}
	}

	return guttesting.Evaluate(guttesting.Options{
		Mutable:   s.config.AllowMutation,
		LookupEnv: lookupEnv,
	}).CapabilityReport()
}
