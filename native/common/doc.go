// Package common contains shared native automation concepts used by provider-specific backends.
//
// Packages under gut/native expose deterministic DTOs, capability reporting, and operation errors so
// higher layers can compile and branch cleanly when the native bridge is either available via cgo,
// unavailable in the current build, or unsupported on the current platform.
package common
