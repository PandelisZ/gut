package mouse

import (
	"context"
	"time"

	gutlog "gut/log"
	"gut/provider"
	"gut/shared"
	"gut/util"
)

const (
	defaultAutoDelay = 100 * time.Millisecond
	defaultSpeed     = 1000
)

type Config struct {
	AutoDelay time.Duration
	Speed     int
}

type EasingFunction func(progress float64) float64

type Mouse struct {
	registry *provider.Registry
	config   Config
}

func Linear(progress float64) float64 {
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func CalculateStepDuration(speed int) time.Duration {
	if speed <= 0 {
		return 0
	}
	return time.Duration(float64(time.Second) / float64(speed))
}

func CalculateMovementTimesteps(steps int, speed int, easing EasingFunction) []time.Duration {
	if steps <= 0 {
		return nil
	}
	if speed <= 0 {
		return make([]time.Duration, steps)
	}
	if easing == nil {
		easing = Linear
	}

	totalDuration := time.Duration(float64(time.Second) * float64(steps) / float64(speed))
	timesteps := make([]time.Duration, 0, steps)
	previous := time.Duration(0)
	for idx := 0; idx < steps; idx++ {
		progress := float64(idx+1) / float64(steps)
		current := time.Duration(float64(totalDuration) * Linear(easing(progress)))
		delay := current - previous
		if delay < 0 {
			delay = 0
		}
		timesteps = append(timesteps, delay)
		previous = current
	}
	return timesteps
}

func New(registry *provider.Registry) *Mouse {
	if registry == nil {
		registry = provider.NewRegistry()
	}

	mouse := &Mouse{
		registry: registry,
		config: Config{
			AutoDelay: defaultAutoDelay,
			Speed:     defaultSpeed,
		},
	}

	if p, err := registry.Mouse(); err == nil {
		p.SetMouseDelay(0)
	}

	return mouse
}

func (m *Mouse) SetAutoDelay(delay time.Duration) {
	m.config.AutoDelay = delay
}

func (m *Mouse) SetSpeed(speed int) {
	m.config.Speed = speed
}

func (m *Mouse) SetPosition(ctx context.Context, target shared.Point) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return p.SetMousePosition(ctx, target)
}

func (m *Mouse) Position(ctx context.Context) (shared.Point, error) {
	p, err := m.provider()
	if err != nil {
		return shared.Point{}, err
	}
	return p.CurrentMousePosition(ctx)
}

func (m *Mouse) Move(ctx context.Context, path []shared.Point, easing ...EasingFunction) error {
	if len(path) == 0 {
		return nil
	}

	p, err := m.provider()
	if err != nil {
		return err
	}

	movementType := Linear
	if len(easing) > 0 && easing[0] != nil {
		movementType = easing[0]
	}

	if err := p.SetMousePosition(ctx, path[0]); err != nil {
		return err
	}
	if len(path) == 1 {
		return nil
	}

	timesteps := CalculateMovementTimesteps(len(path)-1, m.config.Speed, movementType)
	for idx := 1; idx < len(path); idx++ {
		if err := util.SleepContext(ctx, timesteps[idx-1]); err != nil {
			return err
		}
		if err := p.SetMousePosition(ctx, path[idx]); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mouse) LeftClick(ctx context.Context) error {
	return m.Click(ctx, shared.ButtonLeft)
}

func (m *Mouse) RightClick(ctx context.Context) error {
	return m.Click(ctx, shared.ButtonRight)
}

func (m *Mouse) Click(ctx context.Context, button shared.Button) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return p.Click(ctx, button)
}

func (m *Mouse) DoubleClick(ctx context.Context, button shared.Button) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return p.DoubleClick(ctx, button)
}

func (m *Mouse) PressButton(ctx context.Context, button shared.Button) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return p.PressButton(ctx, button)
}

func (m *Mouse) ReleaseButton(ctx context.Context, button shared.Button) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return p.ReleaseButton(ctx, button)
}

func (m *Mouse) ScrollUp(ctx context.Context, amount int) error {
	return m.scroll(ctx, amount, func(ctx context.Context, p provider.MouseProvider, amount int) error {
		return p.ScrollUp(ctx, amount)
	})
}

func (m *Mouse) ScrollDown(ctx context.Context, amount int) error {
	return m.scroll(ctx, amount, func(ctx context.Context, p provider.MouseProvider, amount int) error {
		return p.ScrollDown(ctx, amount)
	})
}

func (m *Mouse) ScrollLeft(ctx context.Context, amount int) error {
	return m.scroll(ctx, amount, func(ctx context.Context, p provider.MouseProvider, amount int) error {
		return p.ScrollLeft(ctx, amount)
	})
}

func (m *Mouse) ScrollRight(ctx context.Context, amount int) error {
	return m.scroll(ctx, amount, func(ctx context.Context, p provider.MouseProvider, amount int) error {
		return p.ScrollRight(ctx, amount)
	})
}

func (m *Mouse) Drag(ctx context.Context, path []shared.Point) error {
	if len(path) == 0 {
		return nil
	}
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	if err := p.PressButton(ctx, shared.ButtonLeft); err != nil {
		return err
	}
	if err := m.Move(ctx, path); err != nil {
		return err
	}
	return p.ReleaseButton(ctx, shared.ButtonLeft)
}

func (m *Mouse) StraightTo(ctx context.Context, target shared.Point) ([]shared.Point, error) {
	position, err := m.Position(ctx)
	if err != nil {
		return nil, err
	}
	return line(position, target), nil
}

func (m *Mouse) Up(ctx context.Context, px int) ([]shared.Point, error) {
	return m.relativeLine(ctx, 0, -px)
}

func (m *Mouse) Down(ctx context.Context, px int) ([]shared.Point, error) {
	return m.relativeLine(ctx, 0, px)
}

func (m *Mouse) Left(ctx context.Context, px int) ([]shared.Point, error) {
	return m.relativeLine(ctx, -px, 0)
}

func (m *Mouse) Right(ctx context.Context, px int) ([]shared.Point, error) {
	return m.relativeLine(ctx, px, 0)
}

func (m *Mouse) scroll(ctx context.Context, amount int, fn func(context.Context, provider.MouseProvider, int) error) error {
	if err := util.SleepContext(ctx, m.config.AutoDelay); err != nil {
		return err
	}
	p, err := m.provider()
	if err != nil {
		return err
	}
	return fn(ctx, p, amount)
}

func (m *Mouse) relativeLine(ctx context.Context, dx int, dy int) ([]shared.Point, error) {
	position, err := m.Position(ctx)
	if err != nil {
		return nil, err
	}
	return line(position, shared.Point{X: position.X + dx, Y: position.Y + dy}), nil
}

func (m *Mouse) provider() (provider.MouseProvider, error) {
	p, err := m.registry.Mouse()
	if err != nil {
		m.logger().Error(err, gutlog.Fields{"component": "mouse"})
		return nil, err
	}
	return p, nil
}

func (m *Mouse) logger() gutlog.Logger {
	if m.registry == nil {
		return gutlog.Nop()
	}
	return m.registry.Logger()
}

func line(origin shared.Point, target shared.Point) []shared.Point {
	x0, y0 := origin.X, origin.Y
	x1, y1 := target.X, target.Y
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := step(x0, x1)
	sy := step(y0, y1)
	err := dx - dy

	points := make([]shared.Point, 0, max(dx, dy)+1)
	for {
		points = append(points, shared.Point{X: x0, Y: y0})
		if x0 == x1 && y0 == y1 {
			return points
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func step(from int, to int) int {
	switch {
	case from < to:
		return 1
	case from > to:
		return -1
	default:
		return 0
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
