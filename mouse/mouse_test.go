package mouse

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gut/provider"
	"gut/shared"
)

type fakeMouseProvider struct {
	delay         time.Duration
	position      shared.Point
	positions     []shared.Point
	clicks        []shared.Button
	doubleClicks  []shared.Button
	presses       []shared.Button
	releases      []shared.Button
	scrollUp      []int
	scrollDown    []int
	scrollLeft    []int
	scrollRight   []int
	events        []string
}

func (f *fakeMouseProvider) SetMouseDelay(delay time.Duration) {
	f.delay = delay
}

func (f *fakeMouseProvider) SetMousePosition(_ context.Context, point shared.Point) error {
	f.position = point
	f.positions = append(f.positions, point)
	f.events = append(f.events, "move")
	return nil
}

func (f *fakeMouseProvider) CurrentMousePosition(context.Context) (shared.Point, error) {
	return f.position, nil
}

func (f *fakeMouseProvider) Click(_ context.Context, button shared.Button) error {
	f.clicks = append(f.clicks, button)
	f.events = append(f.events, "click")
	return nil
}

func (f *fakeMouseProvider) DoubleClick(_ context.Context, button shared.Button) error {
	f.doubleClicks = append(f.doubleClicks, button)
	f.events = append(f.events, "double-click")
	return nil
}

func (f *fakeMouseProvider) ScrollUp(_ context.Context, amount int) error {
	f.scrollUp = append(f.scrollUp, amount)
	f.events = append(f.events, "scroll-up")
	return nil
}

func (f *fakeMouseProvider) ScrollDown(_ context.Context, amount int) error {
	f.scrollDown = append(f.scrollDown, amount)
	f.events = append(f.events, "scroll-down")
	return nil
}

func (f *fakeMouseProvider) ScrollLeft(_ context.Context, amount int) error {
	f.scrollLeft = append(f.scrollLeft, amount)
	f.events = append(f.events, "scroll-left")
	return nil
}

func (f *fakeMouseProvider) ScrollRight(_ context.Context, amount int) error {
	f.scrollRight = append(f.scrollRight, amount)
	f.events = append(f.events, "scroll-right")
	return nil
}

func (f *fakeMouseProvider) PressButton(_ context.Context, button shared.Button) error {
	f.presses = append(f.presses, button)
	f.events = append(f.events, "press")
	return nil
}

func (f *fakeMouseProvider) ReleaseButton(_ context.Context, button shared.Button) error {
	f.releases = append(f.releases, button)
	f.events = append(f.events, "release")
	return nil
}

func TestNewSetsProviderDelayToZero(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeMouseProvider{}
	registry.RegisterMouse(fake)

	New(registry)

	if fake.delay != 0 {
		t.Fatalf("expected provider mouse delay to be zero, got %v", fake.delay)
	}
}

func TestMovementTimingHelpers(t *testing.T) {
	if got := CalculateStepDuration(4); got != 250*time.Millisecond {
		t.Fatalf("unexpected step duration: got %v want %v", got, 250*time.Millisecond)
	}

	timesteps := CalculateMovementTimesteps(4, 4, Linear)
	expected := []time.Duration{
		250 * time.Millisecond,
		250 * time.Millisecond,
		250 * time.Millisecond,
		250 * time.Millisecond,
	}
	if !reflect.DeepEqual(timesteps, expected) {
		t.Fatalf("unexpected movement timesteps: got %v want %v", timesteps, expected)
	}
}

func TestStraightToUsesBresenhamLine(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeMouseProvider{position: shared.Point{X: 0, Y: 0}}
	registry.RegisterMouse(fake)
	mouse := New(registry)

	path, err := mouse.StraightTo(context.Background(), shared.Point{X: 3, Y: 2})
	if err != nil {
		t.Fatalf("unexpected StraightTo error: %v", err)
	}

	expected := []shared.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 2}}
	if !reflect.DeepEqual(path, expected) {
		t.Fatalf("unexpected path: got %v want %v", path, expected)
	}
}

func TestMoveFollowsPath(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeMouseProvider{position: shared.Point{X: 1, Y: 1}}
	registry.RegisterMouse(fake)
	mouse := New(registry)
	mouse.SetSpeed(0)

	path := []shared.Point{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}}
	if err := mouse.Move(context.Background(), path); err != nil {
		t.Fatalf("unexpected Move error: %v", err)
	}

	if !reflect.DeepEqual(fake.positions, path) {
		t.Fatalf("unexpected moved positions: got %v want %v", fake.positions, path)
	}
}

func TestClickDoubleClickScrollAndDrag(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeMouseProvider{position: shared.Point{X: 0, Y: 0}}
	registry.RegisterMouse(fake)
	mouse := New(registry)
	mouse.SetAutoDelay(0)
	mouse.SetSpeed(0)

	if err := mouse.LeftClick(context.Background()); err != nil {
		t.Fatalf("unexpected LeftClick error: %v", err)
	}
	if err := mouse.RightClick(context.Background()); err != nil {
		t.Fatalf("unexpected RightClick error: %v", err)
	}
	if err := mouse.DoubleClick(context.Background(), shared.ButtonMiddle); err != nil {
		t.Fatalf("unexpected DoubleClick error: %v", err)
	}
	if err := mouse.ScrollUp(context.Background(), 2); err != nil {
		t.Fatalf("unexpected ScrollUp error: %v", err)
	}
	if err := mouse.ScrollDown(context.Background(), 3); err != nil {
		t.Fatalf("unexpected ScrollDown error: %v", err)
	}
	if err := mouse.ScrollLeft(context.Background(), 4); err != nil {
		t.Fatalf("unexpected ScrollLeft error: %v", err)
	}
	if err := mouse.ScrollRight(context.Background(), 5); err != nil {
		t.Fatalf("unexpected ScrollRight error: %v", err)
	}

	dragPath := []shared.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}}
	if err := mouse.Drag(context.Background(), dragPath); err != nil {
		t.Fatalf("unexpected Drag error: %v", err)
	}

	if !reflect.DeepEqual(fake.clicks, []shared.Button{shared.ButtonLeft, shared.ButtonRight}) {
		t.Fatalf("unexpected clicks: got %v", fake.clicks)
	}
	if !reflect.DeepEqual(fake.doubleClicks, []shared.Button{shared.ButtonMiddle}) {
		t.Fatalf("unexpected double clicks: got %v", fake.doubleClicks)
	}
	if !reflect.DeepEqual(fake.scrollUp, []int{2}) || !reflect.DeepEqual(fake.scrollDown, []int{3}) || !reflect.DeepEqual(fake.scrollLeft, []int{4}) || !reflect.DeepEqual(fake.scrollRight, []int{5}) {
		t.Fatalf("unexpected scroll calls: up=%v down=%v left=%v right=%v", fake.scrollUp, fake.scrollDown, fake.scrollLeft, fake.scrollRight)
	}
	if !reflect.DeepEqual(fake.presses, []shared.Button{shared.ButtonLeft}) {
		t.Fatalf("unexpected drag press events: got %v", fake.presses)
	}
	if !reflect.DeepEqual(fake.releases, []shared.Button{shared.ButtonLeft}) {
		t.Fatalf("unexpected drag release events: got %v", fake.releases)
	}

	expectedMoves := append([]shared.Point(nil), dragPath...)
	if !reflect.DeepEqual(fake.positions, expectedMoves) {
		t.Fatalf("unexpected drag path positions: got %v want %v", fake.positions, expectedMoves)
	}
}

func TestMouseHighLevelSmoke(t *testing.T) {
	registry := provider.NewRegistry()
	fake := &fakeMouseProvider{position: shared.Point{X: 2, Y: 2}}
	registry.RegisterMouse(fake)
	mouse := New(registry)
	mouse.SetAutoDelay(0)
	mouse.SetSpeed(0)

	if err := mouse.SetPosition(context.Background(), shared.Point{X: 4, Y: 5}); err != nil {
		t.Fatalf("unexpected SetPosition error: %v", err)
	}
	position, err := mouse.Position(context.Background())
	if err != nil {
		t.Fatalf("unexpected Position error: %v", err)
	}
	if position != (shared.Point{X: 4, Y: 5}) {
		t.Fatalf("unexpected current position: got %v", position)
	}

	path, err := mouse.Right(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected Right error: %v", err)
	}
	if err := mouse.Move(context.Background(), path); err != nil {
		t.Fatalf("unexpected Move error: %v", err)
	}
	if err := mouse.Click(context.Background(), shared.ButtonLeft); err != nil {
		t.Fatalf("unexpected Click error: %v", err)
	}

	if len(fake.positions) == 0 || len(fake.clicks) == 0 {
		t.Fatalf("expected provider calls from high-level mouse API, got positions=%v clicks=%v", fake.positions, fake.clicks)
	}
}
