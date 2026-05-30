// Package fx implements pure-state visual effects: confetti particles and
// expanding-ring shockwaves. No rendering — the UI layer reads state and draws.
package fx

import (
	"math"
	"math/rand"
)

// ── HSV helpers ───────────────────────────────────────────────────────────────

func HsvToHex(h, s, v float64) string {
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
	toU8 := func(v float64) uint8 { return uint8(math.Round((v + m) * 255)) }
	return fmtHex(toU8(r), toU8(g), toU8(b))
}

func fmtHex(r, g, b uint8) string {
	return fmtByte(r) + fmtByte(g) + fmtByte(b)
}

func fmtByte(v uint8) string {
	return [...]string{
		"00", "01", "02", "03", "04", "05", "06", "07", "08", "09",
		"0a", "0b", "0c", "0d", "0e", "0f", "10", "11", "12", "13",
		"14", "15", "16", "17", "18", "19", "1a", "1b", "1c", "1d",
		"1e", "1f", "20", "21", "22", "23", "24", "25", "26", "27",
		"28", "29", "2a", "2b", "2c", "2d", "2e", "2f", "30", "31",
		"32", "33", "34", "35", "36", "37", "38", "39", "3a", "3b",
		"3c", "3d", "3e", "3f", "40", "41", "42", "43", "44", "45",
		"46", "47", "48", "49", "4a", "4b", "4c", "4d", "4e", "4f",
		"50", "51", "52", "53", "54", "55", "56", "57", "58", "59",
		"5a", "5b", "5c", "5d", "5e", "5f", "60", "61", "62", "63",
		"64", "65", "66", "67", "68", "69", "6a", "6b", "6c", "6d",
		"6e", "6f", "70", "71", "72", "73", "74", "75", "76", "77",
		"78", "79", "7a", "7b", "7c", "7d", "7e", "7f", "80", "81",
		"82", "83", "84", "85", "86", "87", "88", "89", "8a", "8b",
		"8c", "8d", "8e", "8f", "90", "91", "92", "93", "94", "95",
		"96", "97", "98", "99", "9a", "9b", "9c", "9d", "9e", "9f",
		"a0", "a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9",
		"aa", "ab", "ac", "ad", "ae", "af", "b0", "b1", "b2", "b3",
		"b4", "b5", "b6", "b7", "b8", "b9", "ba", "bb", "bc", "bd",
		"be", "bf", "c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7",
		"c8", "c9", "ca", "cb", "cc", "cd", "ce", "cf", "d0", "d1",
		"d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "da", "db",
		"dc", "dd", "de", "df", "e0", "e1", "e2", "e3", "e4", "e5",
		"e6", "e7", "e8", "e9", "ea", "eb", "ec", "ed", "ee", "ef",
		"f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9",
		"fa", "fb", "fc", "fd", "fe", "ff",
	}[v]
}

// ── Shockwave ─────────────────────────────────────────────────────────────────

const ShockDur = 0.5 // seconds to fully expand + fade
const ShockMaxR = 9  // max radius in rows (X stretched ~2×)

type Wave struct {
	ID int
	CX float64
	CY float64
	T  float64 // seconds since emit
}

type Shockwaves struct {
	Waves []Wave
	nextID int
}

func NewShockwaves() *Shockwaves { return &Shockwaves{} }

func (s *Shockwaves) Emit(cx, cy float64) {
	s.Waves = append(s.Waves, Wave{ID: s.nextID, CX: cx, CY: cy, T: 0})
	s.nextID++
}

func (s *Shockwaves) Update(dtSec float64) {
	for i := range s.Waves {
		s.Waves[i].T += dtSec
	}
	j := 0
	for _, w := range s.Waves {
		if w.T < ShockDur {
			s.Waves[j] = w
			j++
		}
	}
	s.Waves = s.Waves[:j]
}

// ── Confetti ──────────────────────────────────────────────────────────────────

const Gravity = 42.0 // cells / s²

type Particle struct {
	ID    int
	X     float64
	Y     float64
	VX    float64
	VY    float64
	Char  string
	Color string
	Life  float64
	Spin  float64
}

type Theme struct {
	Glyphs []string
	Colors []string
}

var ThemeParty = Theme{
	Glyphs: []string{"★", "✦", "✧", "●", "◆", "▲", "▼", "✺", "❉", "♦", "♥", "•", "*"},
	Colors: []string{"#ff2d95", "#00e5ff", "#ffe600", "#7cff00", "#ff7a00", "#b026ff", "#ff0040", "#19ffd0", "#ffffff"},
}
var ThemeBust = Theme{
	Glyphs: []string{"✗", "✘", "×", "▼", "☓", "■", "↓"},
	Colors: []string{"#e74c3c", "#c0392b", "#7f1d1d", "#922b21", "#555555", "#ff3b3b"},
}
var ThemeAmber = Theme{
	Glyphs: []string{"✦", "★", "✧", "·", "•"},
	Colors: []string{"#f39c12", "#e67e22", "#d35400", "#f1c40f", "#888888"},
}
var ThemeGloom = Theme{
	Glyphs: []string{"·", "˙", ":", ".", "‚", "⋅"},
	Colors: []string{"#6272a4", "#46506b", "#5b6b8c", "#444455", "#384057"},
}

type Confetti struct {
	Particles []Particle
	nextID    int
	cap       int
}

func NewConfetti(cap ...int) *Confetti {
	c := 220
	if len(cap) > 0 {
		c = cap[0]
	}
	return &Confetti{cap: c}
}

func (c *Confetti) add(p Particle) {
	if len(c.Particles) >= c.cap && len(c.Particles) > 0 {
		c.Particles = c.Particles[1:]
	}
	p.ID = c.nextID
	c.nextID++
	c.Particles = append(c.Particles, p)
}

func rnd(a, b float64) float64 { return a + rand.Float64()*(b-a) }
func pickStr(arr []string) string { return arr[rand.Intn(len(arr))] }

// Burst creates a popper-style outward explosion from (x, y), biased upward.
func (c *Confetti) Burst(x, y float64, n int, theme *Theme) {
	if theme == nil {
		theme = &ThemeParty
	}
	for i := 0; i < n; i++ {
		a := rnd(-math.Pi*0.92, -math.Pi*0.08) // upper hemisphere
		speed := rnd(12, 30)
		c.add(Particle{
			X:     x,
			Y:     y,
			VX:    math.Cos(a) * speed,
			VY:    math.Sin(a) * speed,
			Char:  pickStr(theme.Glyphs),
			Color: pickStr(theme.Colors),
			Life:  rnd(1.1, 2.0),
			Spin:  rnd(0, 10),
		})
	}
}

// Rain creates particles falling from above the top edge.
func (c *Confetti) Rain(width float64, n int, theme *Theme) {
	if theme == nil {
		theme = &ThemeParty
	}
	for i := 0; i < n; i++ {
		c.add(Particle{
			X:     rnd(0, width),
			Y:     rnd(-4, 0),
			VX:    rnd(-4, 4),
			VY:    rnd(8, 16),
			Char:  pickStr(theme.Glyphs),
			Color: pickStr(theme.Colors),
			Life:  rnd(2.5, 4),
			Spin:  rnd(0, 10),
		})
	}
}

func (c *Confetti) Update(dtSec float64, height int) {
	next := c.Particles[:0]
	for _, p := range c.Particles {
		p.VY += Gravity * dtSec
		p.X += p.VX * dtSec
		p.Y += p.VY * dtSec
		p.Life -= dtSec
		p.Spin += dtSec * 8
		if p.Life > 0 && p.Y < float64(height)+1 {
			next = append(next, p)
		}
	}
	c.Particles = next
}
