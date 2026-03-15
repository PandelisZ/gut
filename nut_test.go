package gut

import (
	"testing"
)

func TestNewWiresComponentsAndPreservesRegistry(t *testing.T) {
	registry := NewRegistry()
	nut := New(registry)
	if nut == nil {
		t.Fatal("expected Nut instance")
	}
	if nut.Registry != registry {
		t.Fatalf("expected provided registry to be preserved, got %#v want %#v", nut.Registry, registry)
	}
	if nut.Keyboard == nil || nut.Mouse == nil || nut.Screen == nil || nut.Assert == nil {
		t.Fatalf("expected all high-level components to be wired, got keyboard=%#v mouse=%#v screen=%#v assert=%#v", nut.Keyboard, nut.Mouse, nut.Screen, nut.Assert)
	}

	defaulted := New(nil)
	if defaulted == nil || defaulted.Registry == nil || defaulted.Keyboard == nil || defaulted.Mouse == nil || defaulted.Screen == nil || defaulted.Assert == nil {
		t.Fatalf("expected New(nil) to create a fully wired Nut, got %#v", defaulted)
	}
}

func TestDefaultConstructorsProvideDeterministicCoreProviders(t *testing.T) {
	registry := NewDefaultRegistry()
	if registry == nil {
		t.Fatal("expected default registry")
	}
	if _, err := registry.Keyboard(); err != nil {
		t.Fatalf("expected keyboard provider, got %v", err)
	}
	if _, err := registry.Mouse(); err != nil {
		t.Fatalf("expected mouse provider, got %v", err)
	}
	if _, err := registry.Screen(); err != nil {
		t.Fatalf("expected screen provider, got %v", err)
	}
	if _, err := registry.Window(); err != nil {
		t.Fatalf("expected window provider, got %v", err)
	}

	nut := NewDefault()
	if nut == nil || nut.Registry == nil || nut.Keyboard == nil || nut.Mouse == nil || nut.Screen == nil || nut.Assert == nil {
		t.Fatalf("expected NewDefault to return a fully wired Nut, got %#v", nut)
	}
}
