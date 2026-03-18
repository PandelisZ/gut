package keyboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	gutlog "github.com/PandelisZ/gut/log"
	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
	"github.com/PandelisZ/gut/util"
)

const defaultAutoDelay = 300 * time.Millisecond

type Config struct {
	AutoDelay time.Duration
}

type Keyboard struct {
	registry *provider.Registry
	config   Config
}

func New(registry *provider.Registry) *Keyboard {
	if registry == nil {
		registry = provider.NewRegistry()
	}

	keyboard := &Keyboard{
		registry: registry,
		config: Config{
			AutoDelay: defaultAutoDelay,
		},
	}

	if p, err := registry.Keyboard(); err == nil {
		p.SetKeyboardDelay(keyboard.config.AutoDelay)
	}

	return keyboard
}

func (k *Keyboard) SetAutoDelay(delay time.Duration) {
	k.config.AutoDelay = delay
	if p, err := k.provider(); err == nil {
		p.SetKeyboardDelay(delay)
	}
}

func (k *Keyboard) Type(ctx context.Context, input ...any) error {
	if len(input) == 0 {
		return nil
	}

	stringsInput, ok := toStrings(input)
	if ok {
		return k.typeText(ctx, strings.Join(stringsInput, " "))
	}

	keys, ok := toKeys(input)
	if ok {
		return k.tap(ctx, keys...)
	}

	return fmt.Errorf("keyboard type expects either only strings or only shared.Key values")
}

func (k *Keyboard) TypeText(ctx context.Context, parts ...string) error {
	return k.typeText(ctx, strings.Join(parts, " "))
}

func (k *Keyboard) Tap(ctx context.Context, keys ...shared.Key) error {
	return k.tap(ctx, keys...)
}

func (k *Keyboard) Press(ctx context.Context, keys ...shared.Key) error {
	if len(keys) == 0 {
		return nil
	}
	if err := util.SleepContext(ctx, k.config.AutoDelay); err != nil {
		return err
	}
	p, err := k.provider()
	if err != nil {
		return err
	}
	return p.PressKey(ctx, keys...)
}

func (k *Keyboard) Release(ctx context.Context, keys ...shared.Key) error {
	if len(keys) == 0 {
		return nil
	}
	if err := util.SleepContext(ctx, k.config.AutoDelay); err != nil {
		return err
	}
	p, err := k.provider()
	if err != nil {
		return err
	}
	return p.ReleaseKey(ctx, keys...)
}

func (k *Keyboard) typeText(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}
	p, err := k.provider()
	if err != nil {
		return err
	}
	for _, r := range text {
		if err := util.SleepContext(ctx, k.config.AutoDelay); err != nil {
			return err
		}
		if err := p.Type(ctx, string(r)); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keyboard) tap(ctx context.Context, keys ...shared.Key) error {
	if len(keys) == 0 {
		return nil
	}
	if err := util.SleepContext(ctx, k.config.AutoDelay); err != nil {
		return err
	}
	p, err := k.provider()
	if err != nil {
		return err
	}
	return p.Click(ctx, keys...)
}

func (k *Keyboard) provider() (provider.KeyboardProvider, error) {
	p, err := k.registry.Keyboard()
	if err != nil {
		k.logger().Error(err, gutlog.Fields{"component": "keyboard"})
		return nil, err
	}
	return p, nil
}

func (k *Keyboard) logger() gutlog.Logger {
	if k.registry == nil {
		return gutlog.Nop()
	}
	return k.registry.Logger()
}

func toStrings(input []any) ([]string, bool) {
	parts := make([]string, 0, len(input))
	for _, item := range input {
		part, ok := item.(string)
		if !ok {
			return nil, false
		}
		parts = append(parts, part)
	}
	return parts, true
}

func toKeys(input []any) ([]shared.Key, bool) {
	keys := make([]shared.Key, 0, len(input))
	for _, item := range input {
		key, ok := item.(shared.Key)
		if !ok {
			return nil, false
		}
		keys = append(keys, key)
	}
	return keys, true
}
