package window

import (
	"context"
	"errors"
	"testing"
	"time"

	"gut/provider"
	"gut/shared"
)

type fakeWindowProvider struct {
	windows        []shared.WindowHandle
	active         shared.WindowHandle
	titles         map[shared.WindowHandle]string
	regions        map[shared.WindowHandle]shared.Region
	focusResult    bool
	moveResult     bool
	resizeResult   bool
	minimizeResult bool
	restoreResult  bool
}

func (f *fakeWindowProvider) GetWindows(context.Context) ([]shared.WindowHandle, error) {
	return append([]shared.WindowHandle(nil), f.windows...), nil
}

func (f *fakeWindowProvider) GetActiveWindow(context.Context) (shared.WindowHandle, error) {
	return f.active, nil
}

func (f *fakeWindowProvider) GetWindowTitle(_ context.Context, handle shared.WindowHandle) (string, error) {
	return f.titles[handle], nil
}

func (f *fakeWindowProvider) GetWindowRegion(_ context.Context, handle shared.WindowHandle) (shared.Region, error) {
	return f.regions[handle], nil
}

func (f *fakeWindowProvider) FocusWindow(context.Context, shared.WindowHandle) (bool, error) {
	return f.focusResult, nil
}

func (f *fakeWindowProvider) MoveWindow(context.Context, shared.WindowHandle, shared.Point) (bool, error) {
	return f.moveResult, nil
}

func (f *fakeWindowProvider) ResizeWindow(context.Context, shared.WindowHandle, shared.Size) (bool, error) {
	return f.resizeResult, nil
}

func (f *fakeWindowProvider) MinimizeWindow(context.Context, shared.WindowHandle) (bool, error) {
	return f.minimizeResult, nil
}

func (f *fakeWindowProvider) RestoreWindow(context.Context, shared.WindowHandle) (bool, error) {
	return f.restoreResult, nil
}

type fakeScreenProvider struct {
	size shared.Region
}

func (f *fakeScreenProvider) GrabScreen(context.Context) (shared.Image, error) {
	return shared.Image{}, nil
}

func (f *fakeScreenProvider) GrabScreenRegion(context.Context, shared.Region) (shared.Image, error) {
	return shared.Image{}, nil
}

func (f *fakeScreenProvider) HighlightScreenRegion(context.Context, shared.Region, time.Duration, float64) error {
	return nil
}

func (f *fakeScreenProvider) ScreenWidth(context.Context) (int, error) {
	return f.size.Width, nil
}

func (f *fakeScreenProvider) ScreenHeight(context.Context) (int, error) {
	return f.size.Height, nil
}

func (f *fakeScreenProvider) ScreenSize(context.Context) (shared.Region, error) {
	return f.size, nil
}

type fakeElementInspection struct {
	root          shared.WindowElement
	findSequence  []windowElementResult
	findAllResult []shared.WindowElement
	findErr       error
	findCalls     int
}

type windowElementResult struct {
	element shared.WindowElement
	err     error
}

func (f *fakeElementInspection) GetElements(context.Context, shared.WindowHandle, int) (shared.WindowElement, error) {
	return f.root, nil
}

func (f *fakeElementInspection) FindElement(context.Context, shared.WindowHandle, shared.WindowElementDescription) (shared.WindowElement, error) {
	f.findCalls++
	if f.findCalls <= len(f.findSequence) {
		result := f.findSequence[f.findCalls-1]
		return result.element, result.err
	}
	if f.findErr != nil {
		return shared.WindowElement{}, f.findErr
	}
	return shared.WindowElement{}, errors.New("element not found")
}

func (f *fakeElementInspection) FindElements(context.Context, shared.WindowHandle, shared.WindowElementDescription) ([]shared.WindowElement, error) {
	return append([]shared.WindowElement(nil), f.findAllResult...), nil
}

func TestWindowRegionNormalizesAgainstScreenBounds(t *testing.T) {
	registry := provider.NewRegistry()
	registry.RegisterWindow(&fakeWindowProvider{
		regions: map[shared.WindowHandle]shared.Region{
			7: {Left: -10, Top: -5, Width: 40, Height: 30},
			8: {Left: -10, Top: 3, Width: 5, Height: 4},
		},
	})
	registry.RegisterScreen(&fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 25, Height: 20}})

	region, err := New(registry, 7).Region(context.Background())
	if err != nil {
		t.Fatalf("unexpected region error: %v", err)
	}
	if region != (shared.Region{Left: 0, Top: 0, Width: 25, Height: 20}) {
		t.Fatalf("unexpected normalized region: %#v", region)
	}

	zeroWidth, err := New(registry, 8).Region(context.Background())
	if err != nil {
		t.Fatalf("unexpected zero-width region error: %v", err)
	}
	if zeroWidth != (shared.Region{Left: 0, Top: 3, Width: 0, Height: 4}) {
		t.Fatalf("unexpected zero-width normalization: %#v", zeroWidth)
	}
}

func TestGetWindowsAndGetActiveWindowWrapHandles(t *testing.T) {
	registry := provider.NewRegistry()
	registry.RegisterWindow(&fakeWindowProvider{
		windows: []shared.WindowHandle{3, 4},
		active:  4,
	})

	windows, err := GetWindows(context.Background(), registry)
	if err != nil {
		t.Fatalf("unexpected get windows error: %v", err)
	}
	if len(windows) != 2 || windows[0].Handle != 3 || windows[1].Handle != 4 {
		t.Fatalf("unexpected wrapped windows: %#v", windows)
	}

	active, err := GetActiveWindow(context.Background(), registry)
	if err != nil {
		t.Fatalf("unexpected active window error: %v", err)
	}
	if active.Handle != 4 {
		t.Fatalf("unexpected active handle: %d", active.Handle)
	}
}

func TestWindowElementInspectionUsesRegisteredProvider(t *testing.T) {
	registry := provider.NewRegistry()
	registry.RegisterElementInspection(&fakeElementInspection{
		root:          shared.WindowElement{Children: []shared.WindowElement{{}}},
		findSequence:  []windowElementResult{{element: shared.WindowElement{Role: stringPtr("button")}}},
		findAllResult: []shared.WindowElement{{Role: stringPtr("button")}, {Role: stringPtr("button")}},
	})
	window := New(registry, 1)
	query := shared.NewWindowElementQuery("button", shared.WindowElementDescription{Role: "button"})

	root, err := window.GetElements(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected get elements error: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("unexpected root element: %#v", root)
	}

	element, err := window.Find(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected find error: %v", err)
	}
	if element.Role == nil || *element.Role != "button" {
		t.Fatalf("unexpected element: %#v", element)
	}

	elements, err := window.FindAll(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected find all error: %v", err)
	}
	if len(elements) != 2 {
		t.Fatalf("unexpected elements: %#v", elements)
	}
}

func TestWindowElementInspectionReturnsStableMissingProviderError(t *testing.T) {
	_, err := New(provider.NewRegistry(), 1).GetElements(context.Background(), 1)
	if !errors.Is(err, provider.ErrMissingElementInspectionProvider) {
		t.Fatalf("expected missing element inspection error, got %v", err)
	}
}

func TestWindowNewWithNilRegistryReturnsStableMissingProviderError(t *testing.T) {
	_, err := New(nil, 1).Title(context.Background())
	if !errors.Is(err, provider.ErrMissingWindowProvider) {
		t.Fatalf("expected missing window provider error, got %v", err)
	}
}

func TestWindowWaitForRetriesUntilElementIsFound(t *testing.T) {
	registry := provider.NewRegistry()
	inspection := &fakeElementInspection{
		findSequence: []windowElementResult{
			{err: errors.New("element not found")},
			{element: shared.WindowElement{Title: stringPtr("Ready")}},
		},
	}
	registry.RegisterElementInspection(inspection)
	window := New(registry, 1)

	element, err := window.WaitFor(
		context.Background(),
		shared.NewWindowElementQuery("ready", shared.WindowElementDescription{Title: matcherPtr(shared.MatchString("Ready"))}),
		50*time.Millisecond,
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
	if element.Title == nil || *element.Title != "Ready" {
		t.Fatalf("unexpected waited element: %#v", element)
	}
}

func stringPtr(value string) *string {
	return &value
}

func matcherPtr(value shared.StringMatcher) *shared.StringMatcher {
	return &value
}
