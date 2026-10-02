package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tactics — настройки поведения автопилота. Значения по умолчанию
// совпадают с константами, которые раньше были зашиты в код.
type Tactics struct {
	RetreatHPPct    int  `json:"retreat_hp_pct"`    // отступать в город, если суммарное HP отряда ниже N%
	RetreatMinAlive int  `json:"retreat_min_alive"` // отступать, если живых бойцов не больше N
	FleeHPPct       int  `json:"flee_hp_pct"`       // пытаться сбежать из боя при HP отряда ниже N%
	PotionHPPct     int  `json:"potion_hp_pct"`     // пить лечебное зелье при HP героя не выше N%
	SkipTraps       bool `json:"skip_traps"`        // не вскрывать сундуки-ловушки
	SkipAltars      bool `json:"skip_altars"`       // не использовать кровавые алтари
	GambleEvents    bool `json:"gamble_events"`     // принимать рискованные сделки в событиях
}

const tacticsCount = 7

// DefaultTactics возвращает настройки «как в оригинале».
func DefaultTactics() Tactics {
	return Tactics{
		RetreatHPPct:    35,
		RetreatMinAlive: 2,
		FleeHPPct:       25,
		PotionHPPct:     40,
		SkipTraps:       false,
		SkipAltars:      false,
		GambleEvents:    true,
	}
}

type tacticRange struct {
	min, max, step int
	isBool         bool
	isPct          bool
	labelKey       string
	descKey        string
}

var tacticRanges = [tacticsCount]tacticRange{
	{min: 10, max: 70, step: 5, isPct: true, labelKey: "tactics.retreat_hp", descKey: "tactics.retreat_hp.desc"},
	{min: 0, max: 4, step: 1, labelKey: "tactics.retreat_alive", descKey: "tactics.retreat_alive.desc"},
	{min: 10, max: 50, step: 5, isPct: true, labelKey: "tactics.flee_hp", descKey: "tactics.flee_hp.desc"},
	{min: 10, max: 80, step: 5, isPct: true, labelKey: "tactics.potion_hp", descKey: "tactics.potion_hp.desc"},
	{min: 0, max: 1, step: 1, isBool: true, labelKey: "tactics.skip_traps", descKey: "tactics.skip_traps.desc"},
	{min: 0, max: 1, step: 1, isBool: true, labelKey: "tactics.skip_altars", descKey: "tactics.skip_altars.desc"},
	{min: 0, max: 1, step: 1, isBool: true, labelKey: "tactics.gamble", descKey: "tactics.gamble.desc"},
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// value возвращает значение i-й настройки как число (для bool — 0/1).
func (t Tactics) value(i int) int {
	switch i {
	case 0:
		return t.RetreatHPPct
	case 1:
		return t.RetreatMinAlive
	case 2:
		return t.FleeHPPct
	case 3:
		return t.PotionHPPct
	case 4:
		return boolToInt(t.SkipTraps)
	case 5:
		return boolToInt(t.SkipAltars)
	case 6:
		return boolToInt(t.GambleEvents)
	}
	return 0
}

// setValue записывает значение i-й настройки, ограничивая его допустимым диапазоном.
func (t *Tactics) setValue(i, v int) {
	if i < 0 || i >= tacticsCount {
		return
	}
	r := tacticRanges[i]
	v = clampInt(v, r.min, r.max)
	switch i {
	case 0:
		t.RetreatHPPct = v
	case 1:
		t.RetreatMinAlive = v
	case 2:
		t.FleeHPPct = v
	case 3:
		t.PotionHPPct = v
	case 4:
		t.SkipTraps = v == 1
	case 5:
		t.SkipAltars = v == 1
	case 6:
		t.GambleEvents = v == 1
	}
}

// adjust меняет i-ю настройку на шаг в сторону dir (+1/-1). Булевы значения переключаются.
func (t *Tactics) adjust(i, dir int) {
	if i < 0 || i >= tacticsCount {
		return
	}
	r := tacticRanges[i]
	if r.isBool {
		t.setValue(i, 1-t.value(i))
		return
	}
	t.setValue(i, t.value(i)+dir*r.step)
}

// Clamped возвращает копию, в которой все значения приведены к допустимым диапазонам.
func (t Tactics) Clamped() Tactics {
	for i := 0; i < tacticsCount; i++ {
		t.setValue(i, t.value(i))
	}
	return t
}

func (t Tactics) display(i int, lang Language) string {
	r := tacticRanges[i]
	v := t.value(i)
	switch {
	case r.isBool:
		if v == 1 {
			return healStyle.Render("[" + T(lang, "tactics.on") + "]")
		}
		return subtleStyle.Render("[" + T(lang, "tactics.off") + "]")
	case r.isPct:
		return fmt.Sprintf("◀ %d%% ▶", v)
	}
	return fmt.Sprintf("◀ %d ▶", v)
}

// handleTacticsKey обрабатывает клавиши экрана тактики. Второе значение — «клавиша обработана».
func (m *Model) handleTacticsKey(key string) (tea.Cmd, bool) {
	switch key {
	case "up", "k":
		m.TacticsSel = (m.TacticsSel + tacticsCount - 1) % tacticsCount
		return nil, true
	case "down", "j":
		m.TacticsSel = (m.TacticsSel + 1) % tacticsCount
		return nil, true
	case "left", "h":
		m.Tactics.adjust(m.TacticsSel, -1)
		return nil, true
	case "right", "enter", " ":
		m.Tactics.adjust(m.TacticsSel, +1)
		return nil, true
	case "t", "esc":
		m.State = StatePlaying
		return m.restartTicks(m.SpeedMs), true
	}
	return nil, false
}

func (m Model) renderTacticsScreen() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(T(m.Lang, "tactics.title") + " [T]")
	sb.WriteString(header + "\n")
	sb.WriteString(subtleStyle.Render(T(m.Lang, "tactics.subtitle")) + "\n\n")

	showDesc := m.TermHeight >= 26
	for i := 0; i < tacticsCount; i++ {
		label := T(m.Lang, tacticRanges[i].labelKey)
		line := fmt.Sprintf("%-34s %s", label, m.Tactics.display(i, m.Lang))
		if i == m.TacticsSel {
			sb.WriteString(accentStyle.Render("▶ ") + line + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
		if showDesc {
			sb.WriteString("    " + subtleStyle.Render(T(m.Lang, tacticRanges[i].descKey)) + "\n")
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(T(m.Lang, "tactics.controls")))

	box := statsBoxStyle.Width(max(40, m.TermWidth-4)).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}
