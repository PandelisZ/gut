package keyboard

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gut/provider"
	"gut/shared"
)

var keyboardTestBinarySalt = "gut-keyboard-test-salt-2"

type fakeKeyboardProvider struct {
	delay        time.Duration
	typed        []string
	clicks       [][]shared.Key
	pressed      [][]shared.Key
	released     [][]shared.Key
	typeCalls    int
	clickCalls   int
	pressCalls   int
	releaseCalls int
}

func (f *fakeKeyboardProvider) SetKeyboardDelay(delay time.Duration) {
	f.delay = delay
}

func (f *fakeKeyboardProvider) Type(_ context.Context, input string) error {
	f.typeCalls++
	f.typed = append(f.typed, input)
	return nil
}

func (f *fakeKeyboardProvider) Click(_ context.Context, keys ...shared.Key) error {
	f.clickCalls++
	f.clicks = append(f.clicks, append([]shared.Key(nil), keys...))
	return nil
}

func (f *fakeKeyboardProvider) PressKey(_ context.Context, keys ...shared.Key) error {
	f.pressCalls++
	f.pressed = append(f.pressed, append([]shared.Key(nil), keys...))
	return nil
}

func (f *fakeKeyboardProvider) ReleaseKey(_ context.Context, keys ...shared.Key) error {
	f.releaseCalls++
	f.released = append(f.released, append([]shared.Key(nil), keys...))
	return nil
}

func TestNewSetsDefaultProviderDelay(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeKeyboardProvider{}
	registry.RegisterKeyboard(fake)

	New(registry)

	if fake.delay != 300*time.Millisecond {
		t.Fatalf("expected default delay to be configured, got %v", fake.delay)
	}
}

func TestTypeTypesStringsRuneByRune(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeKeyboardProvider{}
	registry.RegisterKeyboard(fake)
	keyboard := New(registry)
	keyboard.SetAutoDelay(0)

	if err := keyboard.Type(context.Background(), "hello", "world"); err != nil {
		t.Fatalf("unexpected type error: %v", err)
	}

	expected := []string{"h", "e", "l", "l", "o", " ", "w", "o", "r", "l", "d"}
	if !reflect.DeepEqual(fake.typed, expected) {
		t.Fatalf("unexpected typed input: got %v want %v", fake.typed, expected)
	}
	if fake.clickCalls != 0 {
		t.Fatalf("expected no click calls, got %d", fake.clickCalls)
	}
}

func TestTypeDelegatesKeyCombo(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeKeyboardProvider{}
	registry.RegisterKeyboard(fake)
	keyboard := New(registry)
	keyboard.SetAutoDelay(0)

	combo := []shared.Key{shared.KeyLeftControl, shared.KeyC}
	args := []any{combo[0], combo[1]}
	if err := keyboard.Type(context.Background(), args...); err != nil {
		t.Fatalf("unexpected type error: %v", err)
	}

	if fake.clickCalls != 1 {
		t.Fatalf("expected one click call, got %d", fake.clickCalls)
	}
	if !reflect.DeepEqual(fake.clicks[0], combo) {
		t.Fatalf("unexpected click combo: got %v want %v", fake.clicks[0], combo)
	}
	if fake.typeCalls != 0 {
		t.Fatalf("expected no type calls, got %d", fake.typeCalls)
	}
}

func TestTypeRejectsMixedInput(t *testing.T) {
	keyboard := New(provider.NewRegistry())
	keyboard.SetAutoDelay(0)

	err := keyboard.Type(context.Background(), "hello", shared.KeyEnter)
	if err == nil {
		t.Fatal("expected mixed input to fail")
	}
}

func TestPressAndReleaseDelegate(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeKeyboardProvider{}
	registry.RegisterKeyboard(fake)
	keyboard := New(registry)
	keyboard.SetAutoDelay(0)

	keys := []shared.Key{shared.KeyLeftShift, shared.KeyA}
	if err := keyboard.Press(context.Background(), keys...); err != nil {
		t.Fatalf("unexpected press error: %v", err)
	}
	if err := keyboard.Release(context.Background(), keys...); err != nil {
		t.Fatalf("unexpected release error: %v", err)
	}

	if fake.pressCalls != 1 {
		t.Fatalf("expected one press call, got %d", fake.pressCalls)
	}
	if fake.releaseCalls != 1 {
		t.Fatalf("expected one release call, got %d", fake.releaseCalls)
	}
	if !reflect.DeepEqual(fake.pressed[0], keys) {
		t.Fatalf("unexpected pressed keys: got %v want %v", fake.pressed[0], keys)
	}
	if !reflect.DeepEqual(fake.released[0], keys) {
		t.Fatalf("unexpected released keys: got %v want %v", fake.released[0], keys)
	}
}

func TestKeyboardHighLevelSmoke(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeKeyboardProvider{}
	registry.RegisterKeyboard(fake)
	keyboard := New(registry)
	keyboard.SetAutoDelay(0)

	if err := keyboard.TypeText(context.Background(), "go", "test"); err != nil {
		t.Fatalf("unexpected TypeText error: %v", err)
	}
	if err := keyboard.Tap(context.Background(), shared.KeyEnter); err != nil {
		t.Fatalf("unexpected Tap error: %v", err)
	}
	if err := keyboard.Press(context.Background(), shared.KeyLeftControl); err != nil {
		t.Fatalf("unexpected Press error: %v", err)
	}
	if err := keyboard.Release(context.Background(), shared.KeyLeftControl); err != nil {
		t.Fatalf("unexpected Release error: %v", err)
	}

	if len(fake.typed) == 0 || fake.clickCalls == 0 || fake.pressCalls == 0 || fake.releaseCalls == 0 {
		t.Fatalf("expected all high-level operations to reach the provider, got typed=%v clickCalls=%d pressCalls=%d releaseCalls=%d", fake.typed, fake.clickCalls, fake.pressCalls, fake.releaseCalls)
	}
}
