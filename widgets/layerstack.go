package widgets

import (
	"github.com/losinggeneration/rovel"
	"github.com/losinggeneration/rovel/geom"
)

// LayerStack is a compositing container that renders children as
// z-ordered layers. All children receive the same full rect — layers
// overlap rather than partitioning space.
//
// Layers paint bottom-to-top (index 0 is the background). Events dispatch
// top-to-bottom (the topmost layer gets first chance to handle).
//
// Use LayerStack for:
//   - Rendering overlays that are always present (effects, decorations)
//   - Z-ordered non-interactive layers (particles, grid overlays)
//   - Any situation where multiple views paint into the same region
//
// For interactive modal overlays (dialogs, menus), use App.ShowOverlay instead.
type LayerStack struct {
	id     rovel.ID
	rect   geom.Rect
	layers []rovel.View
}

// NewLayerStack creates a compositing stack. Layers are painted in order:
// layers[0] is the background, layers[len-1] is the foreground.
func NewLayerStack(layers ...rovel.View) *LayerStack {
	return &LayerStack{
		id:     rovel.NewID(),
		layers: layers,
	}
}

// SetLayers replaces the layer list.
func (s *LayerStack) SetLayers(layers []rovel.View) {
	s.layers = layers
}

func (s *LayerStack) ID() rovel.ID      { return s.id }
func (s *LayerStack) Rect() geom.Rect { return s.rect }

func (s *LayerStack) MinSize() geom.Size {
	var sz geom.Size
	for _, l := range s.layers {
		ms := l.MinSize()
		if ms.W > sz.W {
			sz.W = ms.W
		}
		if ms.H > sz.H {
			sz.H = ms.H
		}
	}
	return sz
}

func (s *LayerStack) Layout(r geom.Rect) {
	s.rect = r
	for _, l := range s.layers {
		l.Layout(r)
	}
}

func (s *LayerStack) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	// Paint bottom-to-top — later layers overwrite earlier cells.
	for _, l := range s.layers {
		d.WithClip(s.rect, func(d rovel.Drawer) {
			l.Paint(d, ctx)
		})
	}
}

func (s *LayerStack) Handle(e rovel.Event, ctx *rovel.Ctx) bool {
	// Dispatch top-to-bottom — topmost layer gets first crack.
	for i := len(s.layers) - 1; i >= 0; i-- {
		if s.layers[i].Handle(e, ctx) {
			return true
		}
	}
	return false
}

func (s *LayerStack) Focusable() bool { return false }

func (s *LayerStack) Children() []rovel.View {
	return s.layers
}
