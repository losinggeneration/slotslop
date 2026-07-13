package widgets

import (
	"github.com/losinggeneration/rovel"
	"github.com/losinggeneration/rovel/geom"
)

// Centered is a layout wrapper that centers its child horizontally within
// the allocated rect. The child receives a rect that is its MinSize wide,
// positioned at the horizontal center of the parent.
//
// Vertically, the child fills the full height of the parent.
type Centered struct {
	id    rovel.ID
	child rovel.View
	rect  geom.Rect
}

// NewCentered wraps child so it is horizontally centered within its parent.
func NewCentered(child rovel.View) *Centered {
	return &Centered{
		id:    rovel.NewID(),
		child: child,
	}
}

func (c *Centered) ID() rovel.ID      { return c.id }
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

func (c *Centered) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	d.WithClip(c.child.Rect(), func(d rovel.Drawer) {
		c.child.Paint(d, ctx)
	})
}

func (c *Centered) Handle(e rovel.Event, ctx *rovel.Ctx) bool {
	return c.child.Handle(e, ctx)
}

func (c *Centered) Focusable() bool {
	if f, ok := c.child.(interface{ Focusable() bool }); ok {
		return f.Focusable()
	}
	return false
}

func (c *Centered) Children() []rovel.View {
	return []rovel.View{c.child}
}
