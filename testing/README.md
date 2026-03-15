# gut/testing live integration harness

This directory contains the opt-in integration harness for the Go `gut` rewrite.

## Environment gates

Live/native smoke tests are skipped unless all of the following are true:

- `GUT_ENABLE_LIVE_TESTS=1`
- the current platform is supported by `libnut-core` (`darwin`, `linux`, or `windows`)
- on Linux, either `DISPLAY` or `WAYLAND_DISPLAY` is set
- every capability required by the test is reported as `available`

More invasive tests can use an additional mutable gate:

- `GUT_ENABLE_MUTATION_TESTS=1`

The harness reports skip reasons with the backend name/binding state and the exact unmet gate or missing capability.

## Running locally

On Windows, the local `go.bat` wrapper keeps `GOTMPDIR` at `gut/.gotmp`, but it special-cases `go test ./testing` and `go test ./...`. The wrapper compiles `./testing` into `gut/.artifacts/go-test/gut-testing.exe` and runs that stable binary directly so the harness does not execute from the blocked `gut/.gotmp` test path. An existing `GOTMPDIR` still takes precedence.

Default behavior is to compile and skip live checks:

```sh
go test ./testing
```

Inspect the current backend/capability state without running tests:

```sh
go run ./cmd/gutenv -format text
go run ./cmd/gutenv -format json
```

Enable the read-only live smoke suite:

```sh
GUT_ENABLE_LIVE_TESTS=1 go test -v ./testing/...
```

Linux usually also needs a desktop session, for example:

```sh
DISPLAY=:0 GUT_ENABLE_LIVE_TESTS=1 go test -v ./testing/...
```

A mutable run can require both env vars:

```sh
GUT_ENABLE_LIVE_TESTS=1 GUT_ENABLE_MUTATION_TESTS=1 go test -v ./testing/...
```

## What is covered

- harness evaluation and skip-message mechanics without native access
- reusable capability/environment reporting for local and CI inspection
- read-only live smoke checks through the real `gut` default registry path
- capability-gated scenarios such as screen size and window enumeration
