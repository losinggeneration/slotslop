// Package outcome classifies a finished spin as jackpot, near-miss, or loss.
package outcome

import (
	"fmt"

	"github.com/losinggeneration/slotslop/internal/data"
)

type Vibe string

const (
	VibeWin     Vibe = "win"
	VibeLose    Vibe = "lose"
	VibeSoClose Vibe = "soclose"
)

type OutcomeKind string

const (
	KindWin       OutcomeKind = "win"
	KindSoClose   OutcomeKind = "soclose"
	KindBust      OutcomeKind = "bust"
	KindLowEffort OutcomeKind = "loweffort"
	KindGemini    OutcomeKind = "gemini"
)

type Outcome struct {
	Kind     OutcomeKind
	Vibe     Vibe
	Title    string    // big banner
	Subtitle string    // smaller contextual line
	Hue      *float64  // mood hue in degrees; nil means rainbow (win)
}

func floatPtr(v float64) *float64 { return &v }

var greatModels = map[string]bool{"opus-4.8": true, "gpt-5.5": true}
var dumbModels = map[string]bool{
	"sonnet-4.6": true, "haiku-4.6": true, "gpt-5.4-mini": true, "gpt-5.3-codex-spark": true,
}
var goodHarnesses = map[string]bool{"cursor": true, "claude-code": true, "codex": true}

func Classify(h *data.HarnessDef, m *data.ModelDef, e data.Effort) Outcome {
	lowEffort := e == data.EffortNoReasoning || e == data.EffortLow

	// 1. Anything Gemini — tough luck.
	if m.Provider == data.ProviderGoogle {
		return Outcome{
			Kind:     KindGemini,
			Vibe:     VibeLose,
			Hue:      floatPtr(265),
			Title:    "🫠  T O U G H   L U C K  : (",
			Subtitle: fmt.Sprintf("you rolled %s. tough luck :(   —   press ⏎ to run it anyway", m.Label),
		}
	}

	// 2. A dumb model busts regardless of everything else.
	if dumbModels[m.ID] {
		return Outcome{
			Kind:     KindBust,
			Vibe:     VibeLose,
			Hue:      floatPtr(0),
			Title:    "💀  B U S T  💀",
			Subtitle: fmt.Sprintf("%s?! that thing couldn't ship a semicolon. catastrophic.   —   ⏎ anyway", m.Label),
		}
	}

	if lowEffort {
		// 3. Good model AND good harness, but you fumbled the effort.
		if greatModels[m.ID] && goodHarnesses[h.ID] {
			return Outcome{
				Kind:     KindSoClose,
				Vibe:     VibeSoClose,
				Hue:      floatPtr(38),
				Title:    "😫  S O   C L O S E !",
				Subtitle: fmt.Sprintf("%s on %s… and then %s effort?! so close.   —   ⏎ to run it", m.Label, h.Label, string(e)),
			}
		}
		// 4. Generic low-effort flop.
		return Outcome{
			Kind:     KindLowEffort,
			Vibe:     VibeLose,
			Hue:      floatPtr(14),
			Title:    "📉  L O W   E F F O R T",
			Subtitle: fmt.Sprintf("%s reasoning on %s? you get what you pay for.   —   ⏎ anyway", string(e), m.Label),
		}
	}

	// 5. Decent model + real effort — jackpot.
	return Outcome{
		Kind:     KindWin,
		Vibe:     VibeWin,
		Hue:      nil,
		Title:    "🎉  J A C K P O T !  🎉",
		Subtitle: fmt.Sprintf("%s on %s @ %s effort — now go ship it 🚀   —   ⏎ to run", m.Label, h.Label, string(e)),
	}
}
