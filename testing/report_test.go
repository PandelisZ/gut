package guttesting_test

import (
	"reflect"
	"testing"

	"gut/native/common"
	"gut/native/libnutcore"
	guttesting "gut/testing"
)

func TestCapabilityReportSortsCapabilitiesAndGates(t *testing.T) {
	report := guttesting.Report{
		Backend: libnutcore.BackendInfo{
			Name:         libnutcore.BackendName,
			Platform:     "linux",
			BindingState: libnutcore.BindingStateUnavailable,
			Notes:        []string{"stub backend"},
		},
		Capabilities: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityWindowList, Availability: common.AvailabilityUnavailable, Reason: "window manager unavailable"},
			common.CapabilityStatus{Capability: common.CapabilityKeyboardTap, Availability: common.AvailabilityStubbed, Reason: "stubbed in test"},
			common.CapabilityStatus{Capability: common.CapabilityScreenSize, Availability: common.AvailabilityAvailable, Reason: "screen size available"},
		),
		RequiredCapabilities: []common.Capability{common.CapabilityWindowList, common.CapabilityScreenSize, common.CapabilityWindowList},
		LiveEnabled:          false,
		MutationEnabled:      false,
		MutableRequested:     true,
		Platform:             "linux",
		SupportedPlatform:    true,
		LinuxDisplaySession:  true,
		Reasons: []string{
			"set GUT_ENABLE_LIVE_TESTS=1 to enable live native integration tests",
			"set GUT_ENABLE_MUTATION_TESTS=1 to enable mutable live integration tests",
			"requires capability window.list (unavailable: window manager unavailable)",
		},
	}

	envReport := report.CapabilityReport()

	if envReport.Ready {
		t.Fatal("expected report to remain not ready")
	}

	if !reflect.DeepEqual(envReport.RequiredCapabilities, []common.Capability{common.CapabilityScreenSize, common.CapabilityWindowList}) {
		t.Fatalf("unexpected required capability ordering: %#v", envReport.RequiredCapabilities)
	}

	wantRequiredStatuses := []common.CapabilityStatus{
		{Capability: common.CapabilityScreenSize, Availability: common.AvailabilityAvailable, Reason: "screen size available"},
		{Capability: common.CapabilityWindowList, Availability: common.AvailabilityUnavailable, Reason: "window manager unavailable"},
	}
	if !reflect.DeepEqual(envReport.RequiredCapabilityStatuses, wantRequiredStatuses) {
		t.Fatalf("unexpected required capability statuses: %#v", envReport.RequiredCapabilityStatuses)
	}

	wantCapabilities := []common.CapabilityStatus{
		{Capability: common.CapabilityKeyboardTap, Availability: common.AvailabilityStubbed, Reason: "stubbed in test"},
		{Capability: common.CapabilityScreenSize, Availability: common.AvailabilityAvailable, Reason: "screen size available"},
		{Capability: common.CapabilityWindowList, Availability: common.AvailabilityUnavailable, Reason: "window manager unavailable"},
	}
	if !reflect.DeepEqual(envReport.Capabilities, wantCapabilities) {
		t.Fatalf("unexpected capability ordering: %#v", envReport.Capabilities)
	}

	wantGates := []guttesting.GateStatus{
		{Name: guttesting.EnvEnableLiveTests, Status: "blocked", Detail: "set GUT_ENABLE_LIVE_TESTS=1 to enable live native integration tests"},
		{Name: "supported_platform", Status: "ready", Detail: "platform \"linux\" is supported by libnut-core live tests"},
		{Name: "linux_display_session", Status: "ready", Detail: "DISPLAY or WAYLAND_DISPLAY is set"},
		{Name: guttesting.EnvEnableMutationTests, Status: "blocked", Detail: "set GUT_ENABLE_MUTATION_TESTS=1 to enable mutable live integration tests"},
	}
	if !reflect.DeepEqual(envReport.Gates, wantGates) {
		t.Fatalf("unexpected gates: %#v", envReport.Gates)
	}
}

func TestFormatReportTextIsDeterministic(t *testing.T) {
	report := guttesting.Report{
		Backend: libnutcore.BackendInfo{
			Name:         libnutcore.BackendName,
			Platform:     "windows",
			BindingState: libnutcore.BindingStateLinked,
		},
		Capabilities: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityMouseMove, Availability: common.AvailabilityAvailable, Reason: "available"},
		),
		RequiredCapabilities: []common.Capability{common.CapabilityMouseMove},
		LiveEnabled:          true,
		MutationEnabled:      false,
		MutableRequested:     false,
		Platform:             "windows",
		SupportedPlatform:    true,
	}

	got := guttesting.FormatReportText(report)
	want := `ready: true
platform: windows
backend: libnut-core/linked
backend_platform: windows
backend_notes:
  - none
env_gates:
  - GUT_ENABLE_LIVE_TESTS: ready (GUT_ENABLE_LIVE_TESTS is enabled)
  - supported_platform: ready (platform "windows" is supported by libnut-core live tests)
  - linux_display_session: not_applicable (only required on linux)
  - GUT_ENABLE_MUTATION_TESTS: not_applicable (mutable evaluation not requested)
reasons:
  - none
required_capabilities:
  - mouse.move: available (available)
capabilities:
  - mouse.move: available (available)`

	if got != want {
		t.Fatalf("unexpected text report:\n%s", got)
	}
}
