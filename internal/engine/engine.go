// Package engine implements the pure slot-machine state machine.
// No rendering — the UI layer reads cols/done and calls Tick each frame
// and Press on keypress.
package engine

import (
	"github.com/losinggeneration/slotslop/internal/data"
)

type ColKind string

const (
	ColHarness ColKind = "harness"
	ColModel   ColKind = "model"
	ColEffort  ColKind = "effort"
)

type ColState string

const (
	StateSpin     ColState = "spin"
	StateStopping ColState = "stopping"
	StateStopped  ColState = "stopped"
)

type Column struct {
	Kind   ColKind
	Title  string
	Labels []string
	Idx    int
	State  ColState
	Wait   int
	Delay  int
}

const StopAt = 7 // higher = longer wind-down

type SlotEngine struct {
	Cols   [3]Column
	Active int
	Done   bool

	SelectedHarness *data.HarnessDef
	SelectedModel   *data.ModelDef
	SelectedEffort  data.Effort

	modelPool  []*data.ModelDef
	effortPool []data.Effort
}

func NewSlotEngine() *SlotEngine {
	e := &SlotEngine{
		effortPool: data.Efforts,
	}
	e.Cols[0] = mkCol(ColHarness, "HARNESS", harnessLabels())
	e.Cols[1] = mkCol(ColModel, "MODEL", data.AllModelLabels())
	e.Cols[2] = mkCol(ColEffort, "EFFORT", effortLabels())
	return e
}

func mkCol(kind ColKind, title string, labels []string) Column {
	return Column{
		Kind:   kind,
		Title:  title,
		Labels: labels,
		Idx:    data.RandIndex(len(labels)),
		State:  StateSpin,
	}
}

func harnessLabels() []string {
	out := make([]string, len(data.Harnesses))
	for i, h := range data.Harnesses {
		out[i] = h.Label
	}
	return out
}

func effortLabels() []string {
	out := make([]string, len(data.Efforts))
	for i, e := range data.Efforts {
		out[i] = string(e)
	}
	return out
}

// Tick advances the machine one animation frame.
func (e *SlotEngine) Tick() {
	for i := range e.Cols {
		e.step(i)
	}
}

func (e *SlotEngine) step(i int) {
	col := &e.Cols[i]
	if col.State == StateStopped {
		return
	}
	if col.State == StateSpin {
		col.Idx = (col.Idx + 1) % len(col.Labels)
		return
	}
	// stopping: step less and less often until it settles
	col.Wait--
	if col.Wait <= 0 {
		col.Idx = (col.Idx + 1) % len(col.Labels)
		col.Delay++
		col.Wait = col.Delay
		if col.Delay > StopAt {
			col.State = StateStopped
			e.onStop(i)
		}
	}
}

func (e *SlotEngine) onStop(i int) {
	col := &e.Cols[i]
	switch col.Kind {
	case ColHarness:
		e.SelectedHarness = data.Harnesses[col.Idx]
		e.modelPool = e.SelectedHarness.Models
		m := &e.Cols[1]
		m.Labels = make([]string, len(e.modelPool))
		for j, md := range e.modelPool {
			m.Labels[j] = md.Label
		}
		m.Idx = data.RandIndex(len(e.modelPool))
		e.Active = 1

	case ColModel:
		e.SelectedModel = e.modelPool[col.Idx]
		e.effortPool = e.SelectedModel.Efforts
		ef := &e.Cols[2]
		ef.Labels = make([]string, len(e.effortPool))
		for j, ev := range e.effortPool {
			ef.Labels[j] = string(ev)
		}
		ef.Idx = data.RandIndex(len(e.effortPool))
		e.Active = 2

	case ColEffort:
		e.SelectedEffort = e.effortPool[col.Idx]
		e.Done = true
	}
}

// Press stops the currently-active reel (called on keypress).
func (e *SlotEngine) Press() {
	col := &e.Cols[e.Active]
	if col.State == StateSpin {
		col.State = StateStopping
		col.Delay = 1
		col.Wait = 1
	}
}

// Command builds the final command string, or returns "" if not done.
func (e *SlotEngine) Command(prompt string) string {
	if !e.Done || e.SelectedHarness == nil || e.SelectedModel == nil || e.SelectedEffort == "" {
		return ""
	}
	return e.SelectedHarness.BuildCommand(e.SelectedModel, e.SelectedEffort, prompt)
}
