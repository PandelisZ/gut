# gut

`gut` is the Go desktop automation layer in this repo. It now ships with:

- the library APIs under `./`
- the capability report CLI at `./cmd/gutenv`
- a local stdio MCP server at `./cmd/gutmcp`
- a Codex plugin plus bundled skills at [`plugins/gut`](./plugins/gut)

## Toolchain requirement

`gut` now depends on the official Go MCP SDK, which currently requires Go `1.25`. The module `go` version was updated accordingly in [`go.mod`](./go.mod).

## Reproducible local commands

On Windows, the local `go.bat` wrapper keeps `GOTMPDIR` at `gut/.gotmp`, but it special-cases the default `go test ./testing` and `go test ./...` flows. For the `gut/testing` package it compiles `./testing` into the stable artifact path `gut/.artifacts/go-test/gut-testing.exe` and runs that binary directly, which avoids executing the blocked `gut/.gotmp/go-build...` test binary. If you already export `GOTMPDIR`, the wrapper leaves it unchanged.

Default unit/CI-style run:

```sh
go test ./...
```

Coverage artifact generation:

```sh
go test -coverprofile=.artifacts/coverage/coverage.out ./...
go tool cover -func=.artifacts/coverage/coverage.out > .artifacts/coverage/coverage.txt
```

Capability/environment report:

```sh
go run ./cmd/gutenv -format text
go run ./cmd/gutenv -format json
```

You can require specific capabilities in the report when checking a local/CI host:

```sh
go run ./cmd/gutenv -format text -require screen.size,window.list
go run ./cmd/gutenv -format json -require screen.size,window.list -mutable
```

Live integration/parity smoke entrypoint:

```sh
GUT_ENABLE_LIVE_TESTS=1 go test -v ./testing/...
```

## MCP server

Run the local MCP server over stdio:

```sh
go run ./cmd/gutmcp
```

Enable mutating tools:

```sh
GUT_MCP_ALLOW_MUTATION=1 go run ./cmd/gutmcp
```

Or:

```sh
go run ./cmd/gutmcp --allow-mutation
```

### Exposed tools

Read-only:

- `status`
- `screen_capture`
- `screen_color_at`
- `screen_find_color`
- `window_list`
- `window_active`
- `window_elements`
- `window_find_elements`
- `accessibility_snapshot`
- `accessibility_search`
- `clipboard_read`

Mutating:

- `window_action`
- `accessibility_action`
- `mouse_action`
- `keyboard_action`
- `clipboard_write`

### Static resources

- `gut://reference/capabilities`
- `gut://reference/keys`

## Codex plugin and skills

The repo-local Codex plugin lives at [`plugins/gut`](./plugins/gut). It bundles:

- the local MCP server configuration in [`plugins/gut/.mcp.json`](./plugins/gut/.mcp.json)
- a launcher script in [`plugins/gut/scripts/run-gutmcp.sh`](./plugins/gut/scripts/run-gutmcp.sh)
- four skills for inspection, targeting, and control

The repo-local marketplace entry is at [`.agents/plugins/marketplace.json`](./.agents/plugins/marketplace.json).

The bundled plugin configuration enables mutation by default via `GUT_MCP_ALLOW_MUTATION=1`, so agents should still use `status` and confirm intent before invoking mutating tools.

## Generic MCP client config

For a generic local MCP host, point it at the server directly:

```json
{
  "mcpServers": {
    "gut": {
      "command": "go",
      "args": ["run", "./cmd/gutmcp"],
      "env": {
        "GUT_MCP_ALLOW_MUTATION": "1"
      }
    }
  }
}
```

If the host does not launch from the repo root, use the plugin launcher script instead:

```json
{
  "mcpServers": {
    "gut": {
      "command": "bash",
      "args": ["/absolute/path/to/gut/plugins/gut/scripts/run-gutmcp.sh"]
    }
  }
}
```

On Windows, prefer the direct `go run ./cmd/gutmcp` configuration instead of the plugin launcher script, because the bundled launcher currently assumes `bash`.

## Platform and permission caveats

- v1 is optimized for macOS.
- Linux support depends on the active display session and capability availability.
- Window, screen, and accessibility behavior is capability-driven; use `status` before automation.
- macOS screen capture and accessibility flows may require Screen Recording and Accessibility permissions.

## Explicitly unsupported in v1

- OCR/text finding in the default registry
- image-template finding in the default registry
- MCP prompts
- packaged MCP bundles or registry publication

## What remains gated

The live `gut/testing` smoke suite stays gated by the existing harness rules:

- `GUT_ENABLE_LIVE_TESTS=1`
- a supported `libnut-core` live platform (`darwin`, `linux`, or `windows`)
- on Linux, `DISPLAY` or `WAYLAND_DISPLAY`
- every required capability reporting `available`

More invasive checks can also require:

- `GUT_ENABLE_MUTATION_TESTS=1`

`go run ./cmd/gutenv` uses the default `gut/testing` evaluation path and default `libnut-core` client so the report matches the same backend info, environment gates, readiness decision, reasons, and capability statuses used by the live test harness.
