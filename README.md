# gut

This module contains the Go rewrite and its local automation/testing entrypoints.

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

## What remains gated

The live `gut/testing` smoke suite stays gated by the existing harness rules:

- `GUT_ENABLE_LIVE_TESTS=1`
- a supported `libnut-core` live platform (`darwin`, `linux`, or `windows`)
- on Linux, `DISPLAY` or `WAYLAND_DISPLAY`
- every required capability reporting `available`

More invasive checks can also require:

- `GUT_ENABLE_MUTATION_TESTS=1`

`go run ./cmd/gutenv` uses the default `gut/testing` evaluation path and default `libnut-core` client so the report matches the same backend info, environment gates, readiness decision, reasons, and capability statuses used by the live test harness.
