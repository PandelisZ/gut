package common

import (
	"errors"
	"testing"
)

func TestParseMouseButton(t *testing.T) {
	button, err := ParseMouseButton("left")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if button != MouseButtonLeft {
		t.Fatalf("unexpected button: %s", button)
	}

	_, err = ParseMouseButton("primary")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestNormalizeModifierToken(t *testing.T) {
	tests := []struct {
		platform string
		value    string
		want     KeyModifier
	}{
		{platform: "darwin", value: "cmd", want: KeyModifierMeta},
		{platform: "windows", value: "win", want: KeyModifierMeta},
		{platform: "linux", value: "right_alt", want: KeyModifierAlt},
		{platform: "linux", value: "right_control", want: KeyModifierControl},
	}

	for _, test := range tests {
		got, err := NormalizeModifierToken(test.platform, test.value)
		if err != nil {
			t.Fatalf("unexpected error for %s/%s: %v", test.platform, test.value, err)
		}
		if got != test.want {
			t.Fatalf("unexpected modifier for %s/%s: got %s want %s", test.platform, test.value, got, test.want)
		}
	}
}

func TestNormalizeModifierTokenRejectsPlatformSpecificAliases(t *testing.T) {
	_, err := NormalizeModifierToken("windows", "cmd")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
