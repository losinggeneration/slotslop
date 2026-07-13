package widgets

import (
	"github.com/losinggeneration/rovel"
	"github.com/losinggeneration/rovel/geom"
	"github.com/losinggeneration/rovel/style"
)

// PhaseBorderStyleFn computes the style for a single border cell.
//
//   - part: which segment of the border (top, right, bottom, left, corner)
//   - x, y: absolute screen coordinates of this cell
//   - r: the full border rect
//   - phase: monotonically increasing animation value (caller-managed)
//   - t: normalized position around the perimeter, in [0, 1).
//     t=0 is the top-left corner; t increases clockwise.
//     Because t is normalized, using 360*t as a hue produces a seamless
//     rainbow loop with no visible seam.
//
// A minimal rainbow border:
//
//	widgets.NewPhaseBorder(child, func(part rovel.BoxPart, x, y int, r geom.Rect, phase, t float64) style.Style {
//	    return style.Style{FG: hsvColor(360*t + phase*200, 1, 1)}
//	})
type PhaseBorderStyleFn func(part rovel.BoxPart, x, y int, r geom.Rect, phase, t float64) style.Style

// PhaseBorder is a border wrapper with time-animated styling.
// It wraps a child view in a 1-cell border whose style is computed per-cell
// by a StyleFn that receives the current animation phase and a normalized
// position around the perimeter.
//
// The phase value is NOT managed by PhaseBorder — the owning view
// updates it via SetPhase() during tick handling.
//
// If SetTitle is called, the title text is rendered on the top edge using
// the same StyleFn, so the gradient flows through the title seamlessly.
// To omit the title (e.g. for SlotSlop-style animated borders), simply
// don't call SetTitle.
type PhaseBorder struct {
	id      rovel.ID
	child   rovel.View
	rect    geom.Rect
	title   string
	phase   float64
	styleFn PhaseBorderStyleFn
}

// NewPhaseBorder wraps child in an animated border. styleFn is called for
// every border cell during Paint with the current phase and normalized position.
func NewPhaseBorder(child rovel.View, styleFn PhaseBorderStyleFn) *PhaseBorder {
	return &PhaseBorder{
		id:      rovel.NewID(),
		child:   child,
		styleFn: styleFn,
	}
}

// SetTitle sets an optional title rendered on the top edge.
// Title cells participate in the same gradient as border cells.
// Don't call this for plain animated borders with no text.
func (b *PhaseBorder) SetTitle(title string) { b.title = title }

// SetPhase updates the animation phase. Call this from your tick handler
// before invalidation so the next paint uses the new phase.
func (b *PhaseBorder) SetPhase(phase float64) { b.phase = phase }

// Phase returns the current animation phase.
func (b *PhaseBorder) Phase() float64 { return b.phase }

func (b *PhaseBorder) ID() rovel.ID      { return b.id }
func (b *PhaseBorder) Rect() geom.Rect { return b.rect }

func (b *PhaseBorder) MinSize() geom.Size {
	childMin := b.child.MinSize()
	return geom.Size{W: childMin.W + 2, H: childMin.H + 2}
}

func (b *PhaseBorder) Layout(r geom.Rect) {
	b.rect = r
	inner := geom.Rect{
		X: r.X + 1,
		Y: r.Y + 1,
		W: r.W - 2,
		H: r.H - 2,
	}
	b.child.Layout(inner)
}

func (b *PhaseBorder) Paint(d rovel.Drawer, ctx *rovel.Ctx) {
	r := b.rect
	if r.W < 2 || r.H < 2 {
		return
	}

	total := perimeterLen(r)

	d.WithClip(r, func(d rovel.Drawer) {
		cd, ok := rovel.CellDrawerOf(d)
		if !ok {
			return
		}
		borderCells := computeBorderCells(r)
		for _, bc := range borderCells {
			idx := perimeterIndex(bc.part, bc.x, bc.y, r)
			t := float64(idx) / float64(total)
			st := b.styleFn(bc.part, bc.x, bc.y, r, b.phase, t)
			cd.SetCell(bc.x, bc.y, bc.ch, st)
		}

		// Draw title on top edge if present.
		if b.title != "" && r.W > 4 {
			titleX := r.X + 2
			maxW := r.W - 4
			t := b.title
			if len(t) > maxW {
				t = t[:maxW]
			}
			for i, ch := range t {
				idx := perimeterIndex(rovel.BoxPartTop, titleX+i, r.Y, r)
				nt := float64(idx) / float64(total)
				st := b.styleFn(rovel.BoxPartTop, titleX+i, r.Y, r, b.phase, nt)
				cd.SetCell(titleX+i, r.Y, ch, st)
			}
		}
	})

	inner := geom.Rect{
		X: r.X + 1,
		Y: r.Y + 1,
		W: r.W - 2,
		H: r.H - 2,
	}
	d.WithClip(inner, func(d rovel.Drawer) {
		b.child.Paint(d, ctx)
	})
}

func (b *PhaseBorder) Handle(e rovel.Event, ctx *rovel.Ctx) bool {
	return b.child.Handle(e, ctx)
}

func (b *PhaseBorder) Focusable() bool { return false }

func (b *PhaseBorder) Children() []rovel.View {
	return []rovel.View{b.child}
}

type cellInfo struct {
	part rovel.BoxPart
	x, y int
	ch   rune
}

func computeBorderCells(r geom.Rect) []cellInfo {
	var out []cellInfo
	// Top edge L→R
	for i := 0; i < r.W; i++ {
		ch := '─'
		part := rovel.BoxPartTop
		if i == 0 {
			ch, part = '┌', rovel.BoxPartCornerTL
		} else if i == r.W-1 {
			ch, part = '┐', rovel.BoxPartCornerTR
		}
		out = append(out, cellInfo{part, r.X + i, r.Y, ch})
	}
	// Right edge T→B
	for j := 1; j < r.H-1; j++ {
		out = append(out, cellInfo{rovel.BoxPartRight, r.X + r.W - 1, r.Y + j, '│'})
	}
	// Bottom edge R→L
	for i := r.W - 1; i >= 0; i-- {
		ch := '─'
		part := rovel.BoxPartBottom
		if i == 0 {
			ch, part = '└', rovel.BoxPartCornerBL
		} else if i == r.W-1 {
			ch, part = '┘', rovel.BoxPartCornerBR
		}
		out = append(out, cellInfo{part, r.X + i, r.Y + r.H - 1, ch})
	}
	// Left edge B→T
	for j := r.H - 2; j >= 1; j-- {
		out = append(out, cellInfo{rovel.BoxPartLeft, r.X, r.Y + j, '│'})
	}
	return out
}

// perimeterLen returns the total number of cells on the border of r.
// Equals 2*W + 2*H - 4 (corners are shared between edges).
func perimeterLen(r geom.Rect) int {
	if r.W <= 0 || r.H <= 0 {
		return 0
	}
	return 2*r.W + 2*r.H - 4
}

// perimeterIndex returns the clockwise index of a border cell starting from
// the top-left corner.
//
// Clockwise enumeration:
//
//	top    (L→R):  0 .. W-1
//	right  (T→B):  W .. W+H-3          (H-2 cells, excluding corners)
//	bottom (R→L):  W+H-2 .. 2W+H-3    (W cells, includes both corners)
//	left   (B→T):  2W+H-2 .. 2W+2H-5  (H-2 cells, excluding corners)
func perimeterIndex(part rovel.BoxPart, x, y int, r geom.Rect) int {
	switch {
	case y == r.Y:
		return x - r.X
	case y == r.Y+r.H-1:
		return (r.W + r.H - 2) + (r.W - 1) - (x - r.X)
	case x == r.X+r.W-1:
		return r.W + (y - r.Y - 1)
	case x == r.X:
		return (2*r.W + r.H - 2) + (r.Y + r.H - 2 - y)
	}
	return 0
}
