// Package data defines harnesses, models, efforts, and command builders.
package data

import (
	"fmt"
	"math/rand"
	"strings"
)

type Effort string

const (
	EffortNoReasoning Effort = "no-reasoning"
	EffortLow         Effort = "low"
	EffortMedium      Effort = "medium"
	EffortHigh        Effort = "high"
)

var Efforts = []Effort{EffortNoReasoning, EffortLow, EffortMedium, EffortHigh}

type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
	ProviderGoogle    Provider = "google"
)

type ModelDef struct {
	ID       string
	Label    string
	Provider Provider
	Slug     string
	Efforts  []Effort
}

type HarnessDef struct {
	ID           string
	Label        string
	Models       []*ModelDef
	BuildCommand func(model *ModelDef, effort Effort, prompt string) string
}

func HasReasoning(e Effort) bool {
	return e != EffortNoReasoning
}

// ShellQuote single-quotes a prompt for the shell, escaping embedded single quotes.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ── Models ──────────────────────────────────────────────────────────────────

var ModelSonnet = &ModelDef{ID: "sonnet-4.6", Label: "Sonnet 4.6", Provider: ProviderAnthropic, Slug: "claude-sonnet-4-6", Efforts: []Effort{EffortNoReasoning, EffortLow, EffortMedium, EffortHigh}}
var ModelHaiku = &ModelDef{ID: "haiku-4.6", Label: "Haiku 4.6", Provider: ProviderAnthropic, Slug: "claude-haiku-4-6", Efforts: []Effort{EffortNoReasoning, EffortLow, EffortMedium, EffortHigh}}
var ModelOpus = &ModelDef{ID: "opus-4.8", Label: "Opus 4.8", Provider: ProviderAnthropic, Slug: "claude-opus-4-8", Efforts: []Effort{EffortNoReasoning, EffortLow, EffortMedium, EffortHigh}}
var ModelGPT55 = &ModelDef{ID: "gpt-5.5", Label: "GPT-5.5", Provider: ProviderOpenAI, Slug: "gpt-5.5", Efforts: []Effort{EffortLow, EffortMedium, EffortHigh}}
var ModelGPT54 = &ModelDef{ID: "gpt-5.4", Label: "GPT-5.4", Provider: ProviderOpenAI, Slug: "gpt-5.4", Efforts: []Effort{EffortLow, EffortMedium, EffortHigh}}
var ModelGPT54Mini = &ModelDef{ID: "gpt-5.4-mini", Label: "GPT-5.4-mini", Provider: ProviderOpenAI, Slug: "gpt-5.4-mini", Efforts: []Effort{EffortNoReasoning, EffortLow, EffortMedium}}
var ModelCodexSpark = &ModelDef{ID: "gpt-5.3-codex-spark", Label: "GPT-5.3-codex-spark", Provider: ProviderOpenAI, Slug: "gpt-5.3-codex-spark", Efforts: []Effort{EffortLow, EffortMedium, EffortHigh}}
var ModelGeminiPro = &ModelDef{ID: "gemini-3.1-pro", Label: "Gemini 3.1 Pro", Provider: ProviderGoogle, Slug: "gemini-3.1-pro", Efforts: []Effort{EffortLow, EffortMedium, EffortHigh}}
var ModelGeminiFlash = &ModelDef{ID: "gemini-3.5-flash", Label: "Gemini 3.5 Flash", Provider: ProviderGoogle, Slug: "gemini-3.5-flash", Efforts: []Effort{EffortNoReasoning, EffortLow, EffortMedium, EffortHigh}}

// ── Harnesses ────────────────────────────────────────────────────────────────

var Harnesses = []*HarnessDef{
	{
		ID:    "claude-code",
		Label: "Claude Code",
		Models: []*ModelDef{ModelSonnet, ModelHaiku, ModelOpus},
		BuildCommand: func(m *ModelDef, e Effort, p string) string {
			effort := ""
			if HasReasoning(e) {
				effort = fmt.Sprintf(" --effort %s", string(e))
			}
			return fmt.Sprintf("claude --model %s%s %s", m.ID, effort, ShellQuote(p))
		},
	},
	{
		ID:    "codex",
		Label: "Codex",
		Models: []*ModelDef{ModelGPT55, ModelGPT54, ModelGPT54Mini, ModelCodexSpark},
		BuildCommand: func(m *ModelDef, e Effort, p string) string {
			effort := ""
			if HasReasoning(e) {
				effort = fmt.Sprintf(` -c model_reasoning_effort="%s"`, string(e))
			}
			return fmt.Sprintf("codex -m %s%s %s", m.ID, effort, ShellQuote(p))
		},
	},
	{
		ID:    "opencode",
		Label: "OpenCode",
		Models: []*ModelDef{
			ModelSonnet, ModelHaiku, ModelOpus,
			ModelGPT55, ModelGPT54, ModelGPT54Mini, ModelCodexSpark,
			ModelGeminiPro, ModelGeminiFlash,
		},
		BuildCommand: func(m *ModelDef, e Effort, p string) string {
			variant := ""
			if HasReasoning(e) {
				variant = fmt.Sprintf(" --variant %s", string(e))
			}
			return fmt.Sprintf("opencode run -m %s/%s%s %s", string(m.Provider), m.Slug, variant, ShellQuote(p))
		},
	},
	{
		ID:     "pi",
		Label:  "Pi",
		Models: []*ModelDef{ModelSonnet, ModelHaiku, ModelOpus, ModelGPT55, ModelGPT54, ModelGPT54Mini},
		BuildCommand: func(m *ModelDef, _ Effort, p string) string {
			return fmt.Sprintf("pi --model %s %s", m.ID, ShellQuote(p))
		},
	},
	{
		ID:     "antigravity",
		Label:  "Antigravity CLI",
		Models: []*ModelDef{ModelGeminiPro, ModelGeminiFlash},
		BuildCommand: func(m *ModelDef, _ Effort, p string) string {
			return fmt.Sprintf("agy -m %s %s", m.ID, ShellQuote(p))
		},
	},
	{
		ID:     "cursor",
		Label:  "Cursor CLI",
		Models: []*ModelDef{ModelSonnet, ModelOpus, ModelGPT55, ModelGPT54, ModelGeminiPro},
		BuildCommand: func(m *ModelDef, _ Effort, p string) string {
			return fmt.Sprintf("cursor-agent --model %s %s", m.ID, ShellQuote(p))
		},
	},
}

// AllModelLabels returns deduplicated model labels across all harnesses.
func AllModelLabels() []string {
	seen := make(map[string]bool)
	var out []string
	for _, h := range Harnesses {
		for _, m := range h.Models {
			if !seen[m.ID] {
				seen[m.ID] = true
				out = append(out, m.Label)
			}
		}
	}
	return out
}

// RandIndex returns a random index in [0, n).
func RandIndex(n int) int {
	return rand.Intn(n)
}
