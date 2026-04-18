package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	gutpkg "github.com/PandelisZ/gut"
	"github.com/PandelisZ/gut/internal/gutmcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, stderr io.Writer, args []string) error {
	allowMutationDefault := boolEnv(os.Getenv("GUT_MCP_ALLOW_MUTATION"))

	flags := flag.NewFlagSet("gutmcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	allowMutation := flags.Bool("allow-mutation", allowMutationDefault, "allow mutating desktop tools")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	_ = stdout
	_ = stderr

	service := gutmcp.NewService(gutpkg.NewDefaultRegistry(), gutmcp.Config{
		AllowMutation: *allowMutation,
	})
	return service.NewServer().Run(ctx, &mcp.StdioTransport{})
}

func boolEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
