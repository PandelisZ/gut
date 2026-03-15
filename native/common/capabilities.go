package common

import "sort"

type Capability string

const (
	CapabilityMouseMove       Capability = "mouse.move"
	CapabilityMouseDrag       Capability = "mouse.drag"
	CapabilityMousePosition   Capability = "mouse.position"
	CapabilityMouseClick      Capability = "mouse.click"
	CapabilityMouseToggle     Capability = "mouse.toggle"
	CapabilityMouseScroll     Capability = "mouse.scroll"
	CapabilityMouseDelay      Capability = "mouse.delay"
	CapabilityKeyboardTap     Capability = "keyboard.tap"
	CapabilityKeyboardToggle  Capability = "keyboard.toggle"
	CapabilityKeyboardType    Capability = "keyboard.type"
	CapabilityKeyboardDelay   Capability = "keyboard.delay"
	CapabilityScreenSize      Capability = "screen.size"
	CapabilityScreenHighlight Capability = "screen.highlight"
	CapabilityScreenCapture   Capability = "screen.capture"
	CapabilityWindowList      Capability = "window.list"
	CapabilityWindowActive    Capability = "window.active"
	CapabilityWindowRect      Capability = "window.rect"
	CapabilityWindowTitle     Capability = "window.title"
	CapabilityWindowFocus     Capability = "window.focus"
	CapabilityWindowMove      Capability = "window.move"
	CapabilityWindowResize    Capability = "window.resize"
	CapabilityWindowMinimize  Capability = "window.minimize"
	CapabilityWindowRestore   Capability = "window.restore"
	CapabilityX11DisplayGet   Capability = "x11.display.get"
	CapabilityX11DisplaySet   Capability = "x11.display.set"
)

type Availability string

const (
	AvailabilityAvailable   Availability = "available"
	AvailabilityUnavailable Availability = "unavailable"
	AvailabilityStubbed     Availability = "stubbed"
	AvailabilityUnsupported Availability = "unsupported"
)

type CapabilityStatus struct {
	Capability   Capability
	Availability Availability
	Reason       string
}

type CapabilitySet map[Capability]CapabilityStatus

func NewCapabilitySet(statuses ...CapabilityStatus) CapabilitySet {
	set := make(CapabilitySet, len(statuses))
	for _, status := range statuses {
		set[status.Capability] = status
	}
	return set
}

func (s CapabilitySet) Status(capability Capability) CapabilityStatus {
	if status, ok := s[capability]; ok {
		return status
	}
	return CapabilityStatus{
		Capability:   capability,
		Availability: AvailabilityUnsupported,
		Reason:       "capability not declared",
	}
}

func (s CapabilitySet) Supports(capability Capability) bool {
	return s.Status(capability).Availability == AvailabilityAvailable
}

func (s CapabilitySet) List() []CapabilityStatus {
	statuses := make([]CapabilityStatus, 0, len(s))
	for _, status := range s {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Capability < statuses[j].Capability
	})
	return statuses
}
