package guttesting

import (
	"fmt"
	"strings"

	"gut/native/common"
	"gut/native/libnutcore"
)

type GateStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type EnvironmentReport struct {
	Ready                      bool                      `json:"ready"`
	Platform                   string                    `json:"platform"`
	SupportedPlatform          bool                      `json:"supported_platform"`
	LiveEnabled                bool                      `json:"live_enabled"`
	MutationEnabled            bool                      `json:"mutation_enabled"`
	MutableRequested           bool                      `json:"mutable_requested"`
	LinuxDisplaySession        bool                      `json:"linux_display_session"`
	Backend                    libnutcore.BackendInfo    `json:"backend"`
	Gates                      []GateStatus              `json:"gates"`
	Reasons                    []string                  `json:"reasons"`
	RequiredCapabilities       []common.Capability       `json:"required_capabilities"`
	RequiredCapabilityStatuses []common.CapabilityStatus `json:"required_capability_statuses"`
	Capabilities               []common.CapabilityStatus `json:"capabilities"`
}

func (r Report) CapabilityStatuses() []common.CapabilityStatus {
	if r.Capabilities == nil {
		return nil
	}
	return r.Capabilities.List()
}

func (r Report) RequiredCapabilityStatuses() []common.CapabilityStatus {
	requiredCapabilities := cloneRequiredCapabilities(r.RequiredCapabilities)
	if len(requiredCapabilities) == 0 {
		return nil
	}

	statuses := make([]common.CapabilityStatus, 0, len(requiredCapabilities))
	for _, capability := range requiredCapabilities {
		statuses = append(statuses, r.Capabilities.Status(capability))
	}
	return statuses
}

func (r Report) GateStatuses() []GateStatus {
	gates := []GateStatus{
		{
			Name:   EnvEnableLiveTests,
			Status: gateStatus(r.LiveEnabled),
			Detail: liveGateDetail(r.LiveEnabled),
		},
		{
			Name:   "supported_platform",
			Status: gateStatus(r.SupportedPlatform),
			Detail: supportedPlatformDetail(r.Platform, r.SupportedPlatform),
		},
		{
			Name:   "linux_display_session",
			Status: linuxDisplayGateStatus(r.Platform, r.LinuxDisplaySession),
			Detail: linuxDisplayGateDetail(r.Platform, r.LinuxDisplaySession),
		},
		{
			Name:   EnvEnableMutationTests,
			Status: mutationGateStatus(r.MutableRequested, r.MutationEnabled),
			Detail: mutationGateDetail(r.MutableRequested, r.MutationEnabled),
		},
	}
	return gates
}

func (r Report) CapabilityReport() EnvironmentReport {
	return EnvironmentReport{
		Ready:                      r.Ready(),
		Platform:                   r.Platform,
		SupportedPlatform:          r.SupportedPlatform,
		LiveEnabled:                r.LiveEnabled,
		MutationEnabled:            r.MutationEnabled,
		MutableRequested:           r.MutableRequested,
		LinuxDisplaySession:        r.LinuxDisplaySession,
		Backend:                    cloneBackendInfo(r.Backend),
		Gates:                      r.GateStatuses(),
		Reasons:                    cloneReasons(r.Reasons),
		RequiredCapabilities:       cloneRequiredCapabilities(r.RequiredCapabilities),
		RequiredCapabilityStatuses: r.RequiredCapabilityStatuses(),
		Capabilities:               r.CapabilityStatuses(),
	}
}

func FormatReportText(report Report) string {
	return report.CapabilityReport().Text()
}

func (r EnvironmentReport) Text() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "ready: %t\n", r.Ready)
	fmt.Fprintf(&builder, "platform: %s\n", reportValue(r.Platform, "unknown"))
	fmt.Fprintf(&builder, "backend: %s\n", backendLabel(r.Backend))
	fmt.Fprintf(&builder, "backend_platform: %s\n", reportValue(r.Backend.Platform, "unknown"))
	writeStringList(&builder, "backend_notes", r.Backend.Notes)
	writeGateStatuses(&builder, "env_gates", r.Gates)
	writeStringList(&builder, "reasons", r.Reasons)
	writeCapabilityStatuses(&builder, "required_capabilities", r.RequiredCapabilityStatuses)
	writeCapabilityStatuses(&builder, "capabilities", r.Capabilities)

	return strings.TrimRight(builder.String(), "\n")
}

func gateStatus(ready bool) string {
	if ready {
		return "ready"
	}
	return "blocked"
}

func liveGateDetail(enabled bool) string {
	if enabled {
		return fmt.Sprintf("%s is enabled", EnvEnableLiveTests)
	}
	return fmt.Sprintf("set %s=1 to enable live native integration tests", EnvEnableLiveTests)
}

func supportedPlatformDetail(platform string, supported bool) string {
	if supported {
		return fmt.Sprintf("platform %q is supported by libnut-core live tests", platform)
	}
	return fmt.Sprintf("platform %q is not supported by libnut-core live tests", platform)
}

func linuxDisplayGateStatus(platform string, ready bool) string {
	if platform != "linux" {
		return "not_applicable"
	}
	return gateStatus(ready)
}

func linuxDisplayGateDetail(platform string, ready bool) string {
	if platform != "linux" {
		return "only required on linux"
	}
	if ready {
		return "DISPLAY or WAYLAND_DISPLAY is set"
	}
	return "linux live tests require DISPLAY or WAYLAND_DISPLAY to be set"
}

func mutationGateStatus(requested bool, enabled bool) string {
	if !requested {
		return "not_applicable"
	}
	return gateStatus(enabled)
}

func mutationGateDetail(requested bool, enabled bool) string {
	if !requested {
		return "mutable evaluation not requested"
	}
	if enabled {
		return fmt.Sprintf("%s is enabled", EnvEnableMutationTests)
	}
	return fmt.Sprintf("set %s=1 to enable mutable live integration tests", EnvEnableMutationTests)
}

func backendLabel(info libnutcore.BackendInfo) string {
	if info.Name == "" && info.BindingState == "" {
		return "unknown"
	}
	if info.Name == "" {
		return string(info.BindingState)
	}
	if info.BindingState == "" {
		return info.Name
	}
	return fmt.Sprintf("%s/%s", info.Name, info.BindingState)
}

func reportValue(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func writeStringList(builder *strings.Builder, label string, values []string) {
	fmt.Fprintf(builder, "%s:\n", label)
	if len(values) == 0 {
		builder.WriteString("  - none\n")
		return
	}
	for _, value := range values {
		fmt.Fprintf(builder, "  - %s\n", value)
	}
}

func writeGateStatuses(builder *strings.Builder, label string, gates []GateStatus) {
	fmt.Fprintf(builder, "%s:\n", label)
	if len(gates) == 0 {
		builder.WriteString("  - none\n")
		return
	}
	for _, gate := range gates {
		fmt.Fprintf(builder, "  - %s: %s (%s)\n", gate.Name, gate.Status, gate.Detail)
	}
}

func writeCapabilityStatuses(builder *strings.Builder, label string, statuses []common.CapabilityStatus) {
	fmt.Fprintf(builder, "%s:\n", label)
	if len(statuses) == 0 {
		builder.WriteString("  - none\n")
		return
	}
	for _, status := range statuses {
		if status.Reason == "" {
			fmt.Fprintf(builder, "  - %s: %s\n", status.Capability, status.Availability)
			continue
		}
		fmt.Fprintf(builder, "  - %s: %s (%s)\n", status.Capability, status.Availability, status.Reason)
	}
}

func cloneBackendInfo(info libnutcore.BackendInfo) libnutcore.BackendInfo {
	info.Notes = cloneReasons(info.Notes)
	return info
}

func cloneReasons(reasons []string) []string {
	if len(reasons) == 0 {
		return nil
	}
	clone := make([]string, len(reasons))
	copy(clone, reasons)
	return clone
}
