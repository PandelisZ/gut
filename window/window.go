package window

import (
	"context"
	"fmt"
	"time"

	"gut/provider"
	"gut/shared"
	"gut/util"
)

type Window struct {
	registry *provider.Registry
	Handle   shared.WindowHandle
}

func New(registry *provider.Registry, handle shared.WindowHandle) *Window {
	if registry == nil {
		registry = provider.NewRegistry()
	}

	return &Window{registry: registry, Handle: handle}
}

func GetWindows(ctx context.Context, registry *provider.Registry) ([]*Window, error) {
	windows, err := registry.Window()
	if err != nil {
		return nil, err
	}

	handles, err := windows.GetWindows(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Window, 0, len(handles))
	for _, handle := range handles {
		result = append(result, New(registry, handle))
	}
	return result, nil
}

func GetActiveWindow(ctx context.Context, registry *provider.Registry) (*Window, error) {
	windows, err := registry.Window()
	if err != nil {
		return nil, err
	}

	handle, err := windows.GetActiveWindow(ctx)
	if err != nil {
		return nil, err
	}
	return New(registry, handle), nil
}

func (w *Window) Title(ctx context.Context) (string, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return "", err
	}
	return windows.GetWindowTitle(ctx, w.Handle)
}

func (w *Window) Region(ctx context.Context) (shared.Region, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return shared.Region{}, err
	}
	rawRegion, err := windows.GetWindowRegion(ctx, w.Handle)
	if err != nil {
		return shared.Region{}, err
	}

	screens, err := w.registry.Screen()
	if err != nil {
		return shared.Region{}, err
	}
	bounds, err := screens.ScreenSize(ctx)
	if err != nil {
		return shared.Region{}, err
	}

	return normalizeRegion(rawRegion, bounds), nil
}

func (w *Window) Move(ctx context.Context, origin shared.Point) (bool, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return false, err
	}
	return windows.MoveWindow(ctx, w.Handle, origin)
}

func (w *Window) Resize(ctx context.Context, size shared.Size) (bool, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return false, err
	}
	return windows.ResizeWindow(ctx, w.Handle, size)
}

func (w *Window) Focus(ctx context.Context) (bool, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return false, err
	}
	return windows.FocusWindow(ctx, w.Handle)
}

func (w *Window) Minimize(ctx context.Context) (bool, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return false, err
	}
	return windows.MinimizeWindow(ctx, w.Handle)
}

func (w *Window) Restore(ctx context.Context) (bool, error) {
	windows, err := w.registry.Window()
	if err != nil {
		return false, err
	}
	return windows.RestoreWindow(ctx, w.Handle)
}

func (w *Window) GetElements(ctx context.Context, maxElements int) (shared.WindowElement, error) {
	inspection, err := w.registry.ElementInspection()
	if err != nil {
		return shared.WindowElement{}, err
	}
	return inspection.GetElements(ctx, w.Handle, maxElements)
}

func (w *Window) Find(ctx context.Context, query shared.WindowElementQuery) (shared.WindowElement, error) {
	inspection, err := w.registry.ElementInspection()
	if err != nil {
		return shared.WindowElement{}, err
	}
	return inspection.FindElement(ctx, w.Handle, query.By.Description)
}

func (w *Window) FindAll(ctx context.Context, query shared.WindowElementQuery) ([]shared.WindowElement, error) {
	inspection, err := w.registry.ElementInspection()
	if err != nil {
		return nil, err
	}
	return inspection.FindElements(ctx, w.Handle, query.By.Description)
}

func (w *Window) WaitFor(ctx context.Context, query shared.WindowElementQuery, timeout time.Duration, interval time.Duration) (shared.WindowElement, error) {
	if timeout <= 0 {
		return w.Find(ctx, query)
	}

	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return shared.WindowElement{}, err
		}

		element, err := w.Find(ctx, query)
		if err == nil {
			return element, nil
		}
		lastErr = err

		if time.Now().After(deadline) {
			return shared.WindowElement{}, fmt.Errorf("timed out waiting for window element query %q after %s: %w", query.ID, timeout, lastErr)
		}
		if err := util.SleepContext(ctx, interval); err != nil {
			return shared.WindowElement{}, err
		}
	}
}

func normalizeRegion(region shared.Region, bounds shared.Region) shared.Region {
	left := region.Left
	top := region.Top
	width := region.Width
	height := region.Height

	if left < bounds.Left {
		width -= bounds.Left - left
		left = bounds.Left
	}
	if top < bounds.Top {
		height -= bounds.Top - top
		top = bounds.Top
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	maxRight := bounds.Left + bounds.Width
	maxBottom := bounds.Top + bounds.Height
	if left+width > maxRight {
		width = maxRight - left
	}
	if top+height > maxBottom {
		height = maxBottom - top
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return shared.Region{Left: left, Top: top, Width: width, Height: height}
}
