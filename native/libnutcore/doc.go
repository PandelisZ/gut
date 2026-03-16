// Package libnutcore exposes the Go bridge for the historical libnut-core native surface.
//
// The package keeps libnutcore as the single backend abstraction for native primitives. When cgo is
// enabled, a thin local C ABI shim under this package calls the vendored libnut-core internals
// directly without any Node/N-API runtime dependency. When cgo is disabled, supported platforms
// still compile and report deterministic unavailable capability reasons so provider wiring and
// higher-level API work can proceed cleanly.
//
// Known gaps:
//   - window minimize and restore remain unsupported because libnut-core does not expose those primitives
//   - macOS screen capture is intentionally disabled in the safety model until a safe implementation is ready
//   - macOS key tap/toggle only expose primitive no-modifier CGEvent paths; higher-level key chords must be composed above this layer
//   - the real cgo bridge path is present but is not exercised in the default CGO_ENABLED=0 test environment
package libnutcore
