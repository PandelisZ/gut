package common

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented           = errors.New("native operation not implemented")
	ErrNativeBindingUnavailable = errors.New("native binding unavailable")
	ErrUnsupportedPlatform      = errors.New("native platform unsupported")
	ErrCapabilityUnavailable    = errors.New("native capability unavailable")
	ErrInvalidToken             = errors.New("invalid native token")
)

type OperationError struct {
	Operation  string
	Platform   string
	Capability Capability
	Detail     string
	Cause      error
}

func (e *OperationError) Error() string {
	message := e.Operation
	if message == "" {
		message = "native operation"
	}
	if e.Platform != "" {
		message += " on " + e.Platform
	}
	if e.Capability != "" {
		message += fmt.Sprintf(" [%s]", e.Capability)
	}
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	if e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *OperationError) Unwrap() error {
	return e.Cause
}

func UnavailableOperation(operation, platform string, capability Capability, detail string) error {
	return &OperationError{
		Operation:  operation,
		Platform:   platform,
		Capability: capability,
		Detail:     detail,
		Cause:      ErrNativeBindingUnavailable,
	}
}

func CapabilityUnavailable(operation, platform string, capability Capability, detail string) error {
	return &OperationError{
		Operation:  operation,
		Platform:   platform,
		Capability: capability,
		Detail:     detail,
		Cause:      ErrCapabilityUnavailable,
	}
}

func UnsupportedOperation(operation, platform string, capability Capability, detail string) error {
	return &OperationError{
		Operation:  operation,
		Platform:   platform,
		Capability: capability,
		Detail:     detail,
		Cause:      ErrUnsupportedPlatform,
	}
}
