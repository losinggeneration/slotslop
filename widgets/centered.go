package widgets

import (
	"github.com/losinggeneration/tui"
	"github.com/losinggeneration/tui/geom"
)

// Centered is a layout wrapper that centers its child horizontally within
// the allocated rect. The child receives a rect that is its MinSize wide,
// positioned at the horizontal center of the parent.
//
// Vertically, the child fills the full height of the parent.
type Centered struct {
	id    tui.ID
	child tui.View
	rect  geom.Rect
}

// NewCentered wraps child so it is horizontally centered within its parent.
func NewCentered(child tui.View) *Centered {
	return &Centered{
		id:    tui.NewID(),
		child: child,
	}
}

func (c *Centered) ID() tui.ID      { return c.id }
func (c *Centered) Rect() geom.Rect { return c.rect }

func (c *Centered) MinSize() geom.Size {
	return c.child.MinSize()
}

func (c *Centered) Layout(r geom.Rect) {
	c.rect = r
	childMin := c.child.MinSize()
	childW := childMin.W
	if childW > r.W {
		childW = r.W
	}
	c.child.Layout(geom.Rect{
		X: r.X + (r.W-childW)/2,
		Y: r.Y,
		W: childW,
		H: r.H,
	})
}

func (c *Centered) Paint(d tui.Drawer, ctx *tui.Ctx) {
	d.WithClip(c.child.Rect(), func(d tui.Drawer) {
		c.child.Paint(d, ctx)
	})
}

func (c *Centered) Handle(e tui.Event, ctx *tui.Ctx) bool {
	return c.child.Handle(e, ctx)
}

func (c *Centered) Focusable() bool {
	if f, ok := c.child.(interface{ Focusable() bool }); ok {
		return f.Focusable()
	}
	return false
}

func (c *Centered) Children() []tui.View {
	return []tui.View{c.child}
}
