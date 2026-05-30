// Package widgets provides composable TUI widgets built on top of the
// github.com/losinggeneration/tui library.
package widgets

import (
	"github.com/losinggeneration/tui"
	"github.com/losinggeneration/tui/geom"
	"github.com/losinggeneration/tui/style"
	"github.com/losinggeneration/tui/text"
)

// TextRun is a segment of text with uniform styling.
type TextRun struct {
	Text  string
	Style style.Style
}

// StyledText renders consecutive text segments with independent styles.
// Segments are painted left-to-right, advancing by rune width.
func StyledText(d tui.Drawer, x, y int, runs []TextRun) {
	for _, run := range runs {
		d.DrawText(geom.Point{X: x, Y: y}, run.Text, run.Style)
		x += text.Width(run.Text)
	}
}

// StyledRunText renders a string where each rune's style is computed by
// styleFn. The function receives the rune index and the rune itself,
// returning the style to use for that cell.
//
// This is the building block for rainbow text, mood text, and any
// per-character coloring pattern.
func StyledRunText(d tui.Drawer, x, y int, s string, styleFn func(idx int, r rune) style.Style) {
	cd, ok := tui.CellDrawerOf(d)
	if !ok {
		// Fallback: draw entire string with the first rune's style
		d.DrawText(geom.Point{X: x, Y: y}, s, style.Style{})
		return
	}
	for i, r := range s {
		st := styleFn(i, r)
		w := tui.RuneWidth(r)
		if w <= 0 {
			continue
		}
		// Only set the left cell; setCellAt handles the
		// continuation cell automatically (WideCont flag).
		cd.SetCell(x, y, r, st)
		x += w
	}
}
