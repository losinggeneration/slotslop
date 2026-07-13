// Package main is the SlotSlop entry point.
//
// Run with: go run ./cmd/slotslop/main.go "fix my auth bug"
package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/losinggeneration/rovel"
	"github.com/losinggeneration/rovel/geom"
	"github.com/losinggeneration/rovel/style"
	"github.com/losinggeneration/rovel/text"

	"github.com/losinggeneration/slotslop/internal/data"
	"github.com/losinggeneration/slotslop/internal/engine"
	"github.com/losinggeneration/slotslop/internal/fx"
	"github.com/losinggeneration/slotslop/internal/outcome"
	"github.com/losinggeneration/slotslop/widgets"
)

// ── Layout constants ──────────────────────────────────────────────────────────

func calcMaxLabel() int {
	m := len("▸ HARNESS ◂")
	for _, h := range data.Harnesses {
		if len(h.Label) > m {
			m = len(h.Label)
		}
		for _, m2 := range h.Models {
			if len(m2.Label) > m {
				m = len(m2.Label)
			}
		}
	}
	for _, e := range data.Efforts {
		if len(string(e)) > m {
			m = len(string(e))
		}
	}
	return m
}

var maxLabel = calcMaxLabel()

var (
	contentW = maxLabel + 2    // inner content width of a reel
	reelW    = contentW + 4    // total reel box width (border + padding)
	gap      = 2               // horizontal gap between reels
	machineW = reelW*3 + gap*2 // total machine width
	leftPad  = 1               // left padding (matches TSX padding: 1)
)

var reelOffsets = [5]int{-2, -1, 0, 1, 2} // 5 visible rows; 0 = payline

// ── Color helpers ─────────────────────────────────────────────────────────────

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

func hexColor(hex string) style.Color {
	h := strings.TrimPrefix(hex, "#")
	rv, _ := strconv.ParseUint(h[0:2], 16, 8)
	gv, _ := strconv.ParseUint(h[2:4], 16, 8)
	bv, _ := strconv.ParseUint(h[4:6], 16, 8)
	return style.ColorRGB(uint8(rv), uint8(gv), uint8(bv))
}

// ── Theme with black background ──────────────────────────────────────────────

func blackTheme() rovel.Theme {
	t := rovel.DefaultThemeModern()
	t.Base.BG = style.ColorRGB(0, 0, 0)
	return t
}

// ── Payline centering ────────────────────────────────────────────────────────

func payline(label string, w int) string {
	r := []rune(label)
	if len(r) >= w {
		return string(r[:w])
	}
	total := w - len(r)
	left := total / 2
	return strings.Repeat(" ", left) + label + strings.Repeat(" ", total-left)
}

// ── State ─────────────────────────────────────────────────────────────────────

type slotState struct {
	eng        *engine.SlotEngine
	confetti   *fx.Confetti
	shocks     *fx.Shockwaves
	phase      float64
	acc        int64 // ms accumulator for engine ticks
	prevActive int
	prevDone   bool
	oc         *outcome.Outcome
	prompt     string

	// Set by Handle before Quit so main can finish after Restore
	finishCmd string
	finishRun bool
}

var S slotState

// ── View ──────────────────────────────────────────────────────────────────────

type appView struct {
	id   rovel.ID
	rect geom.Rect
}

func newAppView() *appView { return &appView{id: rovel.NewID()} }

func (v *appView) ID() rovel.ID           { return v.id }
func (v *appView) Rect() geom.Rect        { return v.rect }
func (v *appView) MinSize() geom.Size     { return geom.Size{W: machineW + 4, H: 24} }
func (v *appView) Layout(r geom.Rect)     { v.rect = r }
func (v *appView) Focusable() bool        { return false }
func (v *appView) Children() []rovel.View { return nil }

// Layout (matching the original TSX exactly):
//
//	Row  0: blank (top padding)
//	Row  1: Title — centered within machineW
//	Row  2: Task — centered within machineW
//	Row  3: blank spacer
//	Row  4: Top marquee (machineW wide)
//	Row  5: blank spacer
//	Row  6: Reel titles (▸ HARNESS ◂ etc.)
//	Row  7-13: Reel border boxes (7 rows: border + 5 labels)
//	Row 14: blank spacer
//	Row 15: Bottom marquee
//	Row 16: blank spacer
//	Row 17: Instructions / results

func (v *appView) Paint(dd rovel.Drawer, ctx *rovel.Ctx) {
	r := v.rect
	phase := S.phase
	cw := contentW
	base := ctx.Theme.Base
	bold := base.WithAttr(style.AttrBold)
	dim := base.WithAttr(style.AttrDim)
	cd, _ := rovel.CellDrawerOf(dd)

	// Everything is left-aligned with leftPad cells from the left edge
	// (matching the original TSX's padding: 1)
	mx := r.X + leftPad // machine X origin

	// ── Row 0: blank (top padding) ───────────────────────────
	row := r.Y

	// ── Row 1: Rainbow title — centered within MACHINE_W ──────
	row++
	title := "🎰   S L O T - S L O P   🎰"
	titleX := mx + (machineW-text.Width(title))/2
	widgets.StyledRunText(cd, titleX, row, title, func(i int, ch rune) style.Style {
		return style.Style{FG: hsvColor(phase*120+float64(i)*16, 1, 1), Attr: style.AttrBold}
	})

	// ── Row 2: Task — centered within MACHINE_W ───────────────
	row++
	promptStr := "task: " + S.prompt
	px := mx + (machineW-text.Width(promptStr))/2
	dd.DrawText(geom.Point{X: px, Y: row}, promptStr, dim)

	// ── Row 3: blank spacer ──────────────────────────────────
	row++

	// ── Row 4: Top marquee ──────────────────────────────────
	row++
	paintMarquee(cd, mx, row, machineW, phase, outcomeHue())

	// ── Row 5: blank spacer ──────────────────────────────────
	row++

	// ── Row 6: Reel titles ───────────────────────────────────
	row++
	reelTitleY := row

	// ── Row 7-12: Reel border boxes ───────────────────────────
	reelBoxY := row + 1
	for i := range S.eng.Cols {
		paintReel(cd, dd, ctx, i, mx, reelTitleY, reelBoxY, cw)
	}

	// ── Row 13: blank spacer ──────────────────────────────────
	row = reelBoxY + 7

	// ── Row 14: Bottom marquee ────────────────────────────────
	botMarqueeY := row + 1
	paintMarquee(cd, mx, botMarqueeY, machineW, phase+0.5, outcomeHue())

	// ── Row 15: blank spacer ──────────────────────────────────
	// ── Row 16+: FX + instructions ───────────────────────────

	// ── FX: Shimmer border on active reel ──────────────────────
	if !S.eng.Done {
		ax := mx + S.eng.Active*(reelW+gap)
		ar := geom.Rect{X: ax, Y: reelBoxY, W: reelW, H: 7}
		for i, c := range borderCells(ar) {
			if c.x >= r.X && c.x < r.X+r.W && c.y >= r.Y && c.y < r.Y+r.H {
				cd.SetCell(c.x, c.y, []rune(c.ch)[0],
					style.Style{FG: hsvColor(phase*260-float64(i)*12, 1, 1), Attr: style.AttrBold})
			}
		}
	}

	// ── FX: Shockwave rings ────────────────────────────────────
	for _, wv := range S.shocks.Waves {
		radius := (wv.T / fx.ShockDur) * fx.ShockMaxR
		bright := 1.0 - wv.T/fx.ShockDur
		ch := "·"
		if bright > 0.6 {
			ch = "●"
		} else if bright > 0.3 {
			ch = "•"
		}
		vv := 90 + bright*165
		col := style.ColorRGB(uint8(math.Round(vv*0.55)), uint8(math.Round(vv)), uint8(math.Round(vv)))
		st := style.Style{FG: col, Attr: style.AttrBold}
		steps := int(math.Max(18, math.Round(radius*9)))
		seen := make(map[[2]int]bool)
		for k := 0; k < steps; k++ {
			a := (float64(k) / float64(steps)) * math.Pi * 2
			cx := int(math.Round(wv.CX + math.Cos(a)*radius*2))
			cy := int(math.Round(wv.CY + math.Sin(a)*radius))
			key := [2]int{cx, cy}
			if seen[key] || cx < r.X || cx >= r.X+r.W || cy < r.Y || cy >= r.Y+r.H {
				continue
			}
			seen[key] = true
			cd.SetCell(cx, cy, []rune(ch)[0], st)
		}
	}

	// ── Confetti particles ─────────────────────────────────────
	for _, pt := range S.confetti.Particles {
		px := int(math.Round(pt.X))
		py := int(math.Round(pt.Y))
		if px < r.X || px >= r.X+r.W || py < r.Y || py >= r.Y+r.H {
			continue
		}
		cd.SetCell(px, py, []rune(pt.Char)[0],
			style.Style{FG: hexColor(pt.Color), Attr: style.AttrBold})
	}

	// ── Instructions / results ─────────────────────────────────
	resY := botMarqueeY + 2 // +1 blank spacer + 1
	cmd := S.eng.Command(S.prompt)
	instrX := mx

	if S.eng.Done && cmd != "" && S.oc != nil {
		o := S.oc

		if o.Vibe == outcome.VibeWin {
			widgets.StyledRunText(cd, instrX, resY, o.Title, func(i int, ch rune) style.Style {
				return style.Style{FG: hsvColor(phase*200+float64(i)*10, 1, 1), Attr: style.AttrBold}
			})
		} else {
			paintMoodText(cd, instrX, resY, o.Title, *o.Hue, phase)
		}
		resY++
		dd.DrawText(geom.Point{X: instrX, Y: resY}, o.Subtitle, dim)
		resY += 2 // blank line before command box

		// Command box with animated border
		var bHue float64
		if o.Vibe == outcome.VibeWin {
			bHue = phase * 200
		} else {
			bHue = *o.Hue + math.Sin(phase*4)*12
		}
		cmdW := text.Width(cmd) + 4
		dd.DrawBorder(geom.Rect{X: instrX, Y: resY, W: cmdW, H: 3}, rovel.BoxStyle{
			Glyphs: rovel.BoxGlyphsLight,
			Edges:  rovel.BoxEdgesAll,
			Style:  style.Style{FG: hsvColor(bHue, 1, 0.6)},
		})
		dd.DrawText(geom.Point{X: instrX + 2, Y: resY + 1}, cmd, bold)
		resY += 4

		widgets.StyledText(dd, instrX, resY, []widgets.TextRun{
			{Text: "⏎", Style: bold},
			{Text: " run it now    ·   ", Style: dim},
			{Text: "esc", Style: bold},
			{Text: " to leave", Style: dim},
		})
	} else {
		activeTitle := strings.ToLower(string(S.eng.Cols[S.eng.Active].Title))
		widgets.StyledText(dd, instrX, resY, []widgets.TextRun{
			{Text: "press ", Style: dim},
			{Text: "⏎ / space", Style: bold},
			{Text: " to stop the ", Style: dim},
			{Text: activeTitle, Style: bold},
			{Text: " reel     (q to quit)", Style: dim},
		})
	}
}

func (v *appView) Handle(e rovel.Event, ctx *rovel.Ctx) bool {
	ke, ok := e.(rovel.KeyEvent)
	if !ok {
		return false
	}

	isEnter := ke.Key == rovel.KeyEnter
	isSpace := ke.Key == rovel.KeyRune && ke.Rune == ' '

	switch {
	case ke.Key == rovel.KeyEsc || (ke.Key == rovel.KeyRune && ke.Rune == 'q'):
		S.finishCmd = S.eng.Command(S.prompt)
		S.finishRun = false
		ctx.Quit()

	case isEnter && S.eng.Done:
		S.finishCmd = S.eng.Command(S.prompt)
		S.finishRun = true
		ctx.Quit()

	case isEnter || isSpace:
		cx, cy := reelCenter(S.eng.Active, v.rect)
		S.shocks.Emit(cx, cy)
		S.eng.Press()
		ctx.InvalidateAll()
	}
	return true
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func outcomeHue() *float64 {
	if S.oc != nil {
		return S.oc.Hue
	}
	return nil
}

func reelCenter(active int, root geom.Rect) (float64, float64) {
	mx := root.X + leftPad
	rx := mx + active*(reelW+gap)
	cx := float64(rx + reelW/2)
	// reelBoxY = root.Y + 7, box is 7 tall, center = +3
	cy := float64(root.Y + 7 + 3)
	return cx, cy
}

// ── Sub-painters ──────────────────────────────────────────────────────────────

func paintMarquee(cd rovel.CellDrawer, x, y, width int, phase float64, hue *float64) {
	for i := range width {
		var st style.Style
		if hue == nil {
			st = style.Style{FG: hsvColor(-phase*220+float64(i)*9, 1, 1)}
		} else {
			s := math.Sin(float64(i)*0.35+phase*5) * 18
			v := 0.45 + 0.5*math.Abs(math.Sin(phase*3.5+float64(i)*0.25))
			st = style.Style{FG: hsvColor(*hue+s, 1, v)}
		}
		cd.SetCell(x+i, y, '█', st)
	}
}

func paintMoodText(cd rovel.CellDrawer, x, y int, txt string, hue, phase float64) {
	flash := math.Sin(phase*7) > 0.78
	widgets.StyledRunText(cd, x, y, txt, func(i int, ch rune) style.Style {
		if flash {
			return style.Style{FG: style.ColorRGB(0xff, 0xff, 0xff), Attr: style.AttrBold}
		}
		v := 0.55 + 0.45*math.Abs(math.Sin(phase*4+float64(i)*0.2))
		return style.Style{FG: hsvColor(hue, 1, v), Attr: style.AttrBold}
	})
}

func paintReel(cd rovel.CellDrawer, dd rovel.Drawer, ctx *rovel.Ctx, idx, mx, titleY, boxY, cw int) {
	col := &S.eng.Cols[idx]
	x := mx + idx*(reelW+gap)
	active := idx == S.eng.Active
	done := S.eng.Done
	stopped := col.State == engine.StateStopped
	L := len(col.Labels)
	if L == 0 {
		return
	}

	dim := ctx.Theme.Base.WithAttr(style.AttrDim)
	bold := ctx.Theme.Base.WithAttr(style.AttrBold)

	// Title row
	title := col.Title
	titleSt := dim
	if active && !done {
		titleSt = bold
		title = "▸ " + col.Title + " ◂"
	} else if stopped {
		title = "✔ " + col.Title
	}
	titleX := x + (reelW-text.Width(title))/2
	dd.DrawText(geom.Point{X: titleX, Y: titleY}, title, titleSt)

	// Border box (7 rows)
	bCol := style.ColorBasic(8)
	if stopped {
		bCol = style.ColorBasic(2)
	}
	dd.DrawBorder(geom.Rect{X: x, Y: boxY, W: reelW, H: 7}, rovel.BoxStyle{
		Glyphs: rovel.BoxGlyphsLight,
		Edges:  rovel.BoxEdgesAll,
		Style:  style.Style{FG: bCol},
	})

	// 5 label rows inside the border
	for ri, offset := range reelOffsets {
		li := (col.Idx + offset + L*8) % L
		label := col.Labels[li]
		ly := boxY + 1 + ri

		if offset == 0 {
			var bg style.Color
			switch col.State {
			case engine.StateStopped:
				bg = style.ColorBasic(2)
			case engine.StateStopping:
				bg = style.ColorBasic(3)
			default:
				bg = style.ColorBasic(7)
			}
			dd.DrawText(geom.Point{X: x + 2, Y: ly}, payline(label, cw), style.Style{
				FG:   style.ColorBasic(0),
				BG:   bg,
				Attr: style.AttrBold,
			})
		} else {
			dd.DrawText(geom.Point{X: x + 2, Y: ly}, payline(label, cw), dim)
		}
	}
}

// ── Perimeter cells (for shimmer border) ─────────────────────────────────────

type borderCell struct {
	x, y int
	ch   string
}

func borderCells(r geom.Rect) []borderCell {
	var out []borderCell
	for i := 0; i < r.W; i++ {
		ch := "─"
		if i == 0 {
			ch = "┌"
		} else if i == r.W-1 {
			ch = "┐"
		}
		out = append(out, borderCell{r.X + i, r.Y, ch})
	}
	for j := 1; j < r.H-1; j++ {
		out = append(out, borderCell{r.X + r.W - 1, r.Y + j, "│"})
	}
	for i := r.W - 1; i >= 0; i-- {
		ch := "─"
		if i == 0 {
			ch = "└"
		} else if i == r.W-1 {
			ch = "┘"
		}
		out = append(out, borderCell{r.X + i, r.Y + r.H - 1, ch})
	}
	for j := r.H - 2; j >= 1; j-- {
		out = append(out, borderCell{r.X, r.Y + j, "│"})
	}
	return out
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Printf("\n%s", helpText)
		return
	}

	S.prompt = strings.Join(args, " ")
	if S.prompt == "" {
		S.prompt = "do something cool"
	}

	app, err := rovel.New(rovel.AppOpts{
		Theme: blackTheme(),
		Input: rovel.InputOpts{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}

	S.eng = engine.NewSlotEngine()
	S.confetti = fx.NewConfetti()
	S.shocks = fx.NewShockwaves()

	root := newAppView()
	app.SetRoot(root)

	if err := app.Enable(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable: %v\n", err)
		os.Exit(1)
	}

	ticker := widgets.NewTicker(app, 33*time.Millisecond, func(dt time.Duration, size geom.Size, ctx *rovel.UpdateCtx) {
		sec := dt.Seconds()
		S.phase += sec

		// Advance engine at ~70ms cadence
		S.acc += dt.Milliseconds()
		for S.acc >= 70 {
			S.acc -= 70
			S.eng.Tick()
		}

		// Classify outcome when done
		if S.eng.Done && S.oc == nil {
			h, m, e := S.eng.SelectedHarness, S.eng.SelectedModel, S.eng.SelectedEffort
			if h != nil && m != nil && e != "" {
				o := outcome.Classify(h, m, e)
				S.oc = &o
			}
		}

		// Reel-lock confetti bursts
		if S.eng.Active > S.prevActive {
			cx, cy := reelCenter(S.prevActive, root.Rect())
			S.confetti.Burst(cx, cy-1, 42, &fx.ThemeParty)
			S.prevActive = S.eng.Active
		}

		// Final celebration
		if S.eng.Done && !S.prevDone {
			S.prevDone = true
			if S.oc != nil {
				cx, cy := reelCenter(2, root.Rect())
				switch S.oc.Vibe {
				case outcome.VibeWin:
					S.confetti.Burst(cx, cy-1, 70, &fx.ThemeParty)
					S.confetti.Rain(float64(size.W), 90, &fx.ThemeParty)
				case outcome.VibeSoClose:
					S.confetti.Burst(cx, cy-1, 55, &fx.ThemeAmber)
				case outcome.VibeLose:
					if S.oc.Kind == outcome.KindGemini {
						S.confetti.Rain(float64(size.W), 40, &fx.ThemeGloom)
					} else {
						S.confetti.Rain(float64(size.W), 120, &fx.ThemeBust)
						S.confetti.Burst(cx, cy, 40, &fx.ThemeBust)
					}
				}
			}
		}

		// Continuous rain after done
		if S.eng.Done && S.oc != nil {
			switch S.oc.Vibe {
			case outcome.VibeWin:
				S.confetti.Rain(float64(size.W), 2, &fx.ThemeParty)
			case outcome.VibeLose:
				if S.oc.Kind == outcome.KindGemini {
					S.confetti.Rain(float64(size.W), 1, &fx.ThemeGloom)
				} else {
					S.confetti.Rain(float64(size.W), 3, &fx.ThemeBust)
				}
			}
		}

		S.confetti.Update(sec, size.H)
		S.shocks.Update(sec)
		ctx.InvalidateAll()
	})
	ticker.Start()
	defer ticker.Stop()

	if err := app.Run(); err != nil {
		app.Restore()
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}

	// Restore terminal BEFORE printing any output
	app.Restore()
	finish(S.finishCmd, S.finishRun)
}

func finish(cmd string, run bool) {
	if run && cmd != "" {
		fmt.Printf("\n  ▶ launching:  \x1b[1m%s\x1b[0m\n\n", cmd)
		c := exec.Command("sh", "-c", cmd)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		c.Run()
		os.Exit(0)
	}
	if cmd != "" {
		fmt.Printf("\n  left without running:\n\n    \x1b[1m%s\x1b[0m\n\n", cmd)
	} else {
		fmt.Print("\n  bye 👋\n\n")
	}
	os.Exit(0)
}

var helpText = `  🎰 slotslop — spin for your harness / model / effort

  usage:
    slotslop "your task prompt here"

  press ⏎ or space to stop each reel, left to right.
  the harness you land on restricts which models can appear,
  and the model restricts which effort levels can appear.
  on the results screen, ⏎ runs the rolled command, esc leaves.
`
