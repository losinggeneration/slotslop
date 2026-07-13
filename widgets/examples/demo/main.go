// Package main demonstrates all slotslop/widgets components.
// Run with: go run ./examples/demo/main.go
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/losinggeneration/slotslop/widgets"
	"github.com/losinggeneration/rovel"
	"github.com/losinggeneration/rovel/geom"
	"github.com/losinggeneration/rovel/style"
	"github.com/losinggeneration/rovel/text"
)

func hsvColor(h, s, v float64) style.Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g = c, x
	case h < 120:
		r, g = x, c
	case h < 180:
		g, b = c, x
	case h < 240:
		g, b = x, c
	case h < 300:
		r, b = x, c
	default:
		r, b = c, x
	}
	u8 := func(v float64) uint8 { return uint8(math.Round((v + m) * 255)) }
	return style.ColorRGB(u8(r), u8(g), u8(b))
}

var state struct {
	phase   float64
	counter int
}

// ── Child views ───────────────────────────────────────────────────────────────

type counterContent struct{ id rovel.ID; rect geom.Rect }
func newCounterContent() *counterContent { return &counterContent{id: rovel.NewID()} }
func (c *counterContent) ID() rovel.ID      { return c.id }
func (c *counterContent) Rect() geom.Rect { return c.rect }
func (c *counterContent) MinSize() geom.Size { return geom.Size{W: 22, H: 5} }
func (c *counterContent) Layout(r geom.Rect)  { c.rect = r }
func (c *counterContent) Handle(rovel.Event, *rovel.Ctx) bool { return false }
func (c *counterContent) Focusable() bool { return false }
func (c *counterContent) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	st := ctx.Theme.Base
	d.DrawText(geom.Point{X: c.rect.X, Y: c.rect.Y + 1}, fmt.Sprintf("  counter: %d", state.counter), st.WithAttr(style.AttrBold))
	d.DrawText(geom.Point{X: c.rect.X, Y: c.rect.Y + 2}, "  press space to tick", st.WithAttr(style.AttrDim))
}

type emptyContent struct{ id rovel.ID; rect geom.Rect }
func newEmptyContent() *emptyContent { return &emptyContent{id: rovel.NewID()} }
func (e *emptyContent) ID() rovel.ID      { return e.id }
func (e *emptyContent) Rect() geom.Rect { return e.rect }
func (e *emptyContent) MinSize() geom.Size { return geom.Size{W: 18, H: 5} }
func (e *emptyContent) Layout(r geom.Rect)  { e.rect = r }
func (e *emptyContent) Handle(rovel.Event, *rovel.Ctx) bool { return false }
func (e *emptyContent) Focusable() bool { return false }
func (e *emptyContent) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	d.DrawText(geom.Point{X: e.rect.X + 2, Y: e.rect.Y + 2}, "no title text", ctx.Theme.Base.WithAttr(style.AttrDim))
}

// ── Main view ─────────────────────────────────────────────────────────────────

type contentView struct {
	id          rovel.ID
	rect        geom.Rect
	borderTitle *widgets.PhaseBorder
	borderPlain *widgets.PhaseBorder
}

func newContentView() *contentView {
	v := &contentView{id: rovel.NewID()}

	rainbowFn := func(part rovel.BoxPart, x, y int, r geom.Rect, phase, t float64) style.Style {
		hue := 360*t + phase*200
		return style.Style{FG: hsvColor(hue, 1, 1), Attr: style.AttrBold}
	}

	v.borderTitle = widgets.NewPhaseBorder(newCounterContent(), rainbowFn)
	v.borderTitle.SetTitle(" with title ")

	v.borderPlain = widgets.NewPhaseBorder(newEmptyContent(), rainbowFn)

	return v
}

func (v *contentView) ID() rovel.ID      { return v.id }
func (v *contentView) Rect() geom.Rect { return v.rect }
func (v *contentView) MinSize() geom.Size { return geom.Size{W: 60, H: 16} }
func (v *contentView) Layout(r geom.Rect) { v.rect = r }
func (v *contentView) Focusable() bool    { return false }
func (v *contentView) Children() []rovel.View { return nil }

func (v *contentView) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	r := v.rect
	cd, _ := rovel.CellDrawerOf(d)
	phase := state.phase

	v.borderTitle.SetPhase(phase)
	v.borderPlain.SetPhase(phase)

	// Row 0: rainbow title
	title := "🎰  Widget Demo  🎰"
	titleX := r.X + (r.W-text.Width(title))/2
	widgets.StyledRunText(d, titleX, r.Y, title, func(i int, ch rune) style.Style {
		return style.Style{FG: hsvColor(phase*120+float64(i)*16, 1, 1), Attr: style.AttrBold}
	})

	// Row 1: marquee bar
	for x := range r.W {
		hue := -phase*220 + float64(x)*9
		cd.SetCell(r.X+x, r.Y+1, '█', style.Style{FG: hsvColor(hue, 1, 1)})
	}

	// Row 3: labels
	dim := ctx.Theme.Base.WithAttr(style.AttrDim)
	d.DrawText(geom.Point{X: r.X + 2, Y: r.Y + 3}, "with title:", dim)
	d.DrawText(geom.Point{X: r.X + r.W/2 + 2, Y: r.Y + 3}, "no title (SlotSlop style):", dim)

	// Rows 4-10: PhaseBorder boxes side by side
	boxY := r.Y + 4
	boxH := 7
	leftX := r.X + 2
	rightX := r.X + r.W/2 + 2

	v.borderTitle.Layout(geom.Rect{X: leftX, Y: boxY, W: v.borderTitle.MinSize().W, H: boxH})
	v.borderTitle.Paint(d, ctx)

	v.borderPlain.Layout(geom.Rect{X: rightX, Y: boxY, W: v.borderPlain.MinSize().W, H: boxH})
	v.borderPlain.Paint(d, ctx)

	// Second marquee
	for x := range r.W {
		hue := -phase*220 + float64(x)*9 + 180
		cd.SetCell(r.X+x, r.Y+r.H-3, '█', style.Style{FG: hsvColor(hue, 1, 1)})
	}

	// Bottom status line
	boldSt := ctx.Theme.Base.WithAttr(style.AttrBold)
	widgets.StyledText(d, r.X+2, r.Y+r.H-1, []widgets.TextRun{
		{Text: "press ", Style: dim},
		{Text: "q", Style: boldSt},
		{Text: " to quit  ·  press ", Style: dim},
		{Text: "space", Style: boldSt},
		{Text: " to increment  ·  ticks: ", Style: dim},
		{Text: fmt.Sprintf("%d", state.counter), Style: boldSt},
	})
}

func (v *contentView) Handle(e rovel.Event, ctx *rovel.Ctx) bool {
	if ke, ok := e.(rovel.KeyEvent); ok {
		switch {
		case ke.Key == rovel.KeyRune && ke.Rune == 'q', ke.Key == rovel.KeyEsc:
			ctx.Quit()
			return true
		case ke.Key == rovel.KeyRune && ke.Rune == ' ':
			state.counter++
			ctx.InvalidateAll()
			return true
		}
	}
	return false
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	app, err := rovel.New(rovel.AppOpts{
		Theme: rovel.DefaultThemeModern(),
		Input: rovel.InputOpts{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}

	app.SetRoot(newContentView())

	if err := app.Enable(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable: %v\n", err)
		os.Exit(1)
	}
	defer app.Restore()

	ticker := widgets.NewTicker(app, 33*time.Millisecond, func(dt time.Duration, size geom.Size, ctx *rovel.UpdateCtx) {
		state.phase += dt.Seconds()
		ctx.InvalidateAll()
	})
	ticker.Start()
	defer ticker.Stop()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
	}
}
