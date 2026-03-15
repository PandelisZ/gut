package gut

import (
	"context"

	"gut/assert"
	"gut/keyboard"
	"gut/mouse"
	"gut/native/libnutcore"
	"gut/provider"
	"gut/provider/defaults"
	"gut/screen"
	"gut/shared"
	"gut/window"
)

type Keyboard = keyboard.Keyboard

type KeyboardConfig = keyboard.Config

type Mouse = mouse.Mouse

type MouseConfig = mouse.Config

type Screen = screen.Screen

type ScreenConfig = screen.Config

type ScreenFindOptions = screen.FindOptions

type Window = window.Window

type Assert = assert.Assert

type Nut struct {
	Registry *provider.Registry
	Keyboard *keyboard.Keyboard
	Mouse    *mouse.Mouse
	Screen   *screen.Screen
	Assert   *assert.Assert
}

func New(registry *provider.Registry) *Nut {
	registry = registryOrNew(registry)
	scr := NewScreen(registry)
	return &Nut{
		Registry: registry,
		Keyboard: NewKeyboard(registry),
		Mouse:    NewMouse(registry),
		Screen:   scr,
		Assert:   NewAssert(scr),
	}
}

func NewRegistry() *provider.Registry {
	return provider.NewRegistry()
}

func NewDefaultRegistry() *provider.Registry {
	registry := provider.NewRegistry()
	client := libnutcore.New(libnutcore.DefaultOptions())
	provider.RegisterLibnutcoreProviders(registry, client)
	defaults.RegisterAuxiliary(registry)
	return registry
}

func NewDefault() *Nut {
	return New(NewDefaultRegistry())
}

func NewKeyboard(registry *provider.Registry) *keyboard.Keyboard {
	return keyboard.New(registryOrNew(registry))
}

func NewMouse(registry *provider.Registry) *mouse.Mouse {
	return mouse.New(registryOrNew(registry))
}

func NewScreen(registry *provider.Registry) *screen.Screen {
	return screen.New(registryOrNew(registry))
}

func NewAssert(scr *screen.Screen) *assert.Assert {
	return assert.New(scr)
}

func GetWindows(ctx context.Context, registry *provider.Registry) ([]*window.Window, error) {
	return window.GetWindows(ctx, registryOrNew(registry))
}

func GetActiveWindow(ctx context.Context, registry *provider.Registry) (*window.Window, error) {
	return window.GetActiveWindow(ctx, registryOrNew(registry))
}

func SingleWord(word string) shared.TextQuery {
	return shared.NewWordQuery(word, word)
}

func TextLine(line string) shared.TextQuery {
	return shared.NewLineQuery(line, line)
}

func WindowWithTitle(title string) shared.WindowQuery {
	return shared.NewWindowQuery(title, shared.MatchString(title))
}

func PixelWithColor(color shared.RGBA) shared.ColorQuery {
	return shared.NewColorQuery(color.Hex(), color)
}

func registryOrNew(registry *provider.Registry) *provider.Registry {
	if registry != nil {
		return registry
	}
	return provider.NewRegistry()
}
