package guttesting_test

import (
	"strings"
	"testing"

	"github.com/PandelisZ/gut/native/common"
	guttesting "github.com/PandelisZ/gut/testing"
)

func TestEvaluateReportsLiveGateWhenDisabled(t *testing.T) {
	report := guttesting.Evaluate(guttesting.Options{
		GOOS: "linux",
		LookupEnv: func(key string) (string, bool) {
			switch key {
			case "DISPLAY":
				return ":99", true
			default:
				return "", false
			}
		},
		CapabilitySet: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityScreenSize, Availability: common.AvailabilityAvailable, Reason: "available in test"},
		),
		RequiredCapabilities: []common.Capability{common.CapabilityScreenSize},
	})

	if report.Ready() {
		t.Fatal("expected live gate to block readiness when GUT_ENABLE_LIVE_TESTS is unset")
	}
	if !strings.Contains(report.SkipMessage(), guttesting.EnvEnableLiveTests) {
		t.Fatalf("expected skip message to mention %s, got %q", guttesting.EnvEnableLiveTests, report.SkipMessage())
	}
}

func TestEvaluateReportsCapabilityGate(t *testing.T) {
	report := guttesting.Evaluate(guttesting.Options{
		GOOS: "linux",
		LookupEnv: func(key string) (string, bool) {
			switch key {
			case guttesting.EnvEnableLiveTests:
				return "1", true
			case "DISPLAY":
				return ":99", true
			default:
				return "", false
			}
		},
		CapabilitySet: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityScreenSize, Availability: common.AvailabilityUnavailable, Reason: "native binding is not linked"},
		),
		RequiredCapabilities: []common.Capability{common.CapabilityScreenSize},
	})

	if report.Ready() {
		t.Fatal("expected unavailable capability to block readiness")
	}
	message := report.SkipMessage()
	if !strings.Contains(message, string(common.CapabilityScreenSize)) {
		t.Fatalf("expected skip message to mention the missing capability, got %q", message)
	}
	if !strings.Contains(message, "native binding is not linked") {
		t.Fatalf("expected skip message to include the capability reason, got %q", message)
	}
}

func TestEvaluateReportsMutableGate(t *testing.T) {
	report := guttesting.Evaluate(guttesting.Options{
		GOOS:    "windows",
		Mutable: true,
		LookupEnv: func(key string) (string, bool) {
			if key == guttesting.EnvEnableLiveTests {
				return "1", true
			}
			return "", false
		},
		CapabilitySet: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityMouseMove, Availability: common.AvailabilityAvailable, Reason: "available in test"},
		),
		RequiredCapabilities: []common.Capability{common.CapabilityMouseMove},
	})

	if report.Ready() {
		t.Fatal("expected mutable gate to block readiness when GUT_ENABLE_MUTATION_TESTS is unset")
	}
	if !strings.Contains(report.SkipMessage(), guttesting.EnvEnableMutationTests) {
		t.Fatalf("expected skip message to mention %s, got %q", guttesting.EnvEnableMutationTests, report.SkipMessage())
	}
}
