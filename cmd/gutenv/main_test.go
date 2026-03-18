package main

import (
	"reflect"
	"testing"

	"github.com/PandelisZ/gut/native/common"
)

func TestParseCapabilityListSortsAndDeduplicates(t *testing.T) {
	got := parseCapabilityList(" window.list,screen.size,window.list ,, keyboard.tap ")
	want := []common.Capability{
		common.CapabilityKeyboardTap,
		common.CapabilityScreenSize,
		common.CapabilityWindowList,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected capability list: %#v", got)
	}
}

func TestParseArgsParsesMutableAndRequiredCapabilities(t *testing.T) {
	cfg, err := parseArgs([]string{"-format", "JSON", "-require", "window.list, screen.size", "-mutable"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.Format != "json" {
		t.Fatalf("unexpected format: %q", cfg.Format)
	}
	if !cfg.Mutable {
		t.Fatal("expected mutable flag to be enabled")
	}

	want := []common.Capability{common.CapabilityScreenSize, common.CapabilityWindowList}
	if !reflect.DeepEqual(cfg.RequiredCapabilities, want) {
		t.Fatalf("unexpected required capabilities: %#v", cfg.RequiredCapabilities)
	}
}

func TestParseArgsRejectsUnsupportedFormat(t *testing.T) {
	_, err := parseArgs([]string{"-format", "yaml"})
	if err == nil {
		t.Fatal("expected parseArgs to reject an unsupported format")
	}
}
