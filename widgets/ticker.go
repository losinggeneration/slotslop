package widgets

import (
	"sync"
	"time"

	"github.com/losinggeneration/tui"
	"github.com/losinggeneration/tui/geom"
)

// TickFn is the callback invoked on each tick. It receives the elapsed
// time since the last tick, a painter-style context for invalidation,
// and the current terminal size.
//
// Unlike a raw Post callback, TickFn runs at a fixed cadence and
// receives timing information — making it suitable for animations,
// physics updates, and game loops.
type TickFn func(dt time.Duration, size geom.Size, ctx *tui.UpdateCtx)

// Ticker provides periodic tick callbacks on the app's event loop.
//
// It wraps the common pattern of time.NewTicker + App.Post into a
// convenient Start/Stop lifecycle. Ticks are serialized with keyboard
// and mouse events — no mutex needed for state access inside TickFn.
//
// Ticker is the Go TUI equivalent of Qt's QTimer or TurboVision's
// evTimer. It does not require any changes to the core tui library.
//
// Usage:
//
//	ticker := widgets.NewTicker(app, 33*time.Millisecond, func(dt time.Duration, size geom.Size, ctx *tui.UpdateCtx) {
//	    state.Advance(dt.Seconds())
//	    ctx.InvalidateAll()
//	})
//	ticker.Start()
//	defer ticker.Stop()
type Ticker struct {
	app      *tui.App
	interval time.Duration
	fn       TickFn
	stopCh   chan struct{}
	once     sync.Once
}

// NewTicker creates a periodic ticker that invokes fn on the app loop.
// The ticker does not start until Start() is called.
func NewTicker(app *tui.App, interval time.Duration, fn TickFn) *Ticker {
	return &Ticker{
		app:      app,
		interval: interval,
		fn:       fn,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the tick loop in a background goroutine.
// It is safe to call Start only once. Subsequent calls are no-ops.
func (t *Ticker) Start() {
	t.once.Do(func() {
		go t.run()
	})
}

// Stop terminates the tick loop. It is safe to call multiple times.
func (t *Ticker) Stop() {
	select {
	case t.stopCh <- struct{}{}:
	default:
	}
}

func (t *Ticker) run() {
	tc := time.NewTicker(t.interval)
	defer tc.Stop()

	last := time.Now()
	for {
		select {
		case <-t.stopCh:
			return
		case now := <-tc.C:
			dt := now.Sub(last)
			last = now

			size := t.app.Size()

			_ = t.app.Post(func(ctx *tui.UpdateCtx) {
				t.fn(dt, size, ctx)
			})
		}
	}
}
