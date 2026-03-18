package guttesting

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	stdtesting "testing"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/native/libnutcore"
)

const (
	EnvEnableLiveTests     = "GUT_ENABLE_LIVE_TESTS"
	EnvEnableMutationTests = "GUT_ENABLE_MUTATION_TESTS"
)

type Options struct {
	Mutable              bool
	RequiredCapabilities []common.Capability
	Client               libnutcore.Client
	CapabilitySet        common.CapabilitySet
	GOOS                 string
	LookupEnv            func(string) (string, bool)
}

type Report struct {
	Client               libnutcore.Client
	Backend              libnutcore.BackendInfo
	Capabilities         common.CapabilitySet
	RequiredCapabilities []common.Capability
	LiveEnabled          bool
	MutationEnabled      bool
	MutableRequested     bool
	Platform             string
	SupportedPlatform    bool
	LinuxDisplaySession  bool
	Reasons              []string
}

func Evaluate(options Options) Report {
	lookupEnv := options.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	platform := options.GOOS
	if platform == "" {
		platform = runtime.GOOS
	}

	report := Report{
		Platform:             platform,
		SupportedPlatform:    isSupportedPlatform(platform),
		LiveEnabled:          envEnabled(lookupEnv, EnvEnableLiveTests),
		MutationEnabled:      envEnabled(lookupEnv, EnvEnableMutationTests),
		MutableRequested:     options.Mutable,
		RequiredCapabilities: cloneRequiredCapabilities(options.RequiredCapabilities),
	}

	if !report.LiveEnabled {
		report.Reasons = append(report.Reasons, fmt.Sprintf("set %s=1 to enable live native integration tests", EnvEnableLiveTests))
	}
	if !report.SupportedPlatform {
		report.Reasons = append(report.Reasons, fmt.Sprintf("platform %q is not supported by libnut-core live tests", platform))
	}
	if platform == "linux" {
		report.LinuxDisplaySession = hasLinuxDisplaySession(lookupEnv)
		if !report.LinuxDisplaySession {
			report.Reasons = append(report.Reasons, "linux live tests require DISPLAY or WAYLAND_DISPLAY to be set")
		}
	}
	if options.Mutable && !report.MutationEnabled {
		report.Reasons = append(report.Reasons, fmt.Sprintf("set %s=1 to enable mutable live integration tests", EnvEnableMutationTests))
	}

	client := options.Client
	if client == nil {
		client = libnutcore.New(libnutcore.DefaultOptions())
	}
	report.Client = client
	report.Backend = client.Info()

	capabilities := cloneCapabilities(options.CapabilitySet)
	if capabilities == nil {
		capabilities = client.Capabilities()
	}
	report.Capabilities = capabilities

	for _, capability := range report.RequiredCapabilities {
		status := capabilities.Status(capability)
		if status.Availability == common.AvailabilityAvailable {
			continue
		}
		report.Reasons = append(report.Reasons, fmt.Sprintf("requires capability %s (%s: %s)", capability, status.Availability, status.Reason))
	}

	return report
}

func Require(t *stdtesting.T, options Options) Report {
	t.Helper()

	report := Evaluate(options)
	if !report.Ready() {
		t.Skip(report.SkipMessage())
	}
	return report
}

func (r Report) Ready() bool {
	return len(r.Reasons) == 0
}

func (r Report) SkipMessage() string {
	parts := make([]string, 0, len(r.Reasons)+1)
	parts = append(parts, "live integration test skipped")
	if r.Backend.Name != "" {
		parts = append(parts, fmt.Sprintf("backend=%s/%s", r.Backend.Name, r.Backend.BindingState))
	}
	if len(r.Reasons) != 0 {
		parts = append(parts, strings.Join(r.Reasons, "; "))
	}
	return strings.Join(parts, ": ")
}

func isSupportedPlatform(platform string) bool {
	switch platform {
	case "darwin", "linux", "windows":
		return true
	default:
		return false
	}
}

func hasLinuxDisplaySession(lookupEnv func(string) (string, bool)) bool {
	for _, key := range []string{"DISPLAY", "WAYLAND_DISPLAY"} {
		value, ok := lookupEnv(key)
		if ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func envEnabled(lookupEnv func(string) (string, bool), key string) bool {
	value, ok := lookupEnv(key)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func cloneRequiredCapabilities(capabilities []common.Capability) []common.Capability {
	if len(capabilities) == 0 {
		return nil
	}

	set := make(map[common.Capability]struct{}, len(capabilities))
	clone := make([]common.Capability, 0, len(capabilities))
	for _, capability := range capabilities {
		if _, ok := set[capability]; ok {
			continue
		}
		set[capability] = struct{}{}
		clone = append(clone, capability)
	}
	sort.Slice(clone, func(i, j int) bool {
		return clone[i] < clone[j]
	})
	return clone
}

func cloneCapabilities(capabilities common.CapabilitySet) common.CapabilitySet {
	if capabilities == nil {
		return nil
	}
	clone := make(common.CapabilitySet, len(capabilities))
	for capability, status := range capabilities {
		clone[capability] = status
	}
	return clone
}
