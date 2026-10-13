package main

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// ОТРИСОВКА КАРТЫ ГОРОДА
// ============================================================

// townGlyph — один символ карты вместе со стилем; соседние одинаковые символы
// склеиваются в один вызов lipgloss, чтобы кадр рисовался быстро.
type townGlyph struct {
	r      rune
	fg, bg uint8
	bold   bool
}

var townStyleCache = map[[3]uint8]lipgloss.Style{}

func townStyle(g townGlyph) lipgloss.Style {
	key := [3]uint8{g.fg, g.bg, 0}
	if g.bold {
		key[2] = 1
	}
	if st, ok := townStyleCache[key]; ok {
		return st
	}
	st := lipgloss.NewStyle()
	if g.fg != 0 {
		st = st.Foreground(lipgloss.Color(strconv.Itoa(int(g.fg))))
	}
	if g.bg != 0 {
		st = st.Background(lipgloss.Color(strconv.Itoa(int(g.bg))))
	}
	if g.bold {
		st = st.Bold(true)
	}
	townStyleCache[key] = st
	return st
}

// townSignKeys — ключи словаря с подписями зданий на вывесках.
var townSignKeys = map[TownPhase]string{
	TownPhaseSellLoot:   "town.sign.market",
	TownPhaseMagistrate: "town.sign.magistrate",
	TownPhaseChurch:     "town.sign.church",
	TownPhaseTavern:     "town.sign.tavern",
	TownPhaseGuild:      "town.sign.guild",
	TownPhaseSmithy:     "town.sign.smithy",
	TownPhaseTannery:    "town.sign.tannery",
	TownPhaseAlchemist:  "town.sign.alchemist",
	TownPhaseDepart:     "town.sign.gate",
}

// townBuildingType — понятный тип здания, разделённый на 2 строки для читаемости.
// В словаре строки разделены символом перевода строки.
func (m Model) townBuildingType(b TownBuilding) (string, string) {
	key, ok := townSignKeys[b.Phase]
	if !ok {
		return "", ""
	}
	first, second, _ := strings.Cut(T(m.Lang, key), "\n")
	return first, second
}

// townPlaque — подпись на нижней табличке: уровень заведения или этаж у Врат.
func (m Model) townPlaque(b TownBuilding) string {
	level := 0
	switch b.Phase {
	case TownPhaseChurch:
		level = m.Legacy.ChurchLevel
	case TownPhaseTavern:
		level = m.Legacy.TavernLevel
	case TownPhaseSmithy:
		level = m.Legacy.SmithyLevel
	case TownPhaseTannery:
		level = m.Legacy.TanneryLevel
	case TownPhaseDepart:
		return T(m.Lang, "town.gate_floor", max(1, m.MaxFloor))
	}
	if level <= 0 {
		return ""
	}
	return T(m.Lang, "ui.level_short") + "." + strconv.Itoa(level)
}

// centerRunes центрирует текст в ширине w, обрезая длинные подписи многоточием.
func centerRunes(s string, w int) []rune {
	r := []rune(s)
	if len(r) > w {
		r = append(r[:w-1], '…')
	}
	pad := w - len(r)
	left := pad / 2
	out := make([]rune, 0, w)
	for i := 0; i < left; i++ {
		out = append(out, ' ')
	}
	out = append(out, r...)
	for len(out) < w {
		out = append(out, ' ')
	}
	return out
}

// renderTownMap рисует окно карты города размером viewW×viewH с центром на отряде.
func (m Model) renderTownMap(viewW, viewH int) string {
	t := m.Town
	if t == nil || viewW <= 0 || viewH <= 0 {
		return ""
	}
	vw, vh := min(viewW, t.W), min(viewH, t.H)
	startX := clampInt(m.TownPos.X-vw/2, 0, t.W-vw)
	startY := clampInt(m.TownPos.Y-vh/2, 0, t.H-vh)

	// Тексты вывесок: 3 слота (тип строка 1, тип строка 2, уровень/этаж)
	texts := make([][3][]rune, len(t.Buildings))
	for i, b := range t.Buildings {
		t1, t2 := m.townBuildingType(b)
		texts[i][0] = centerRunes(t1, townSignLen)
		texts[i][1] = centerRunes(t2, townSignLen)
		texts[i][2] = centerRunes(m.townPlaque(b), townSignLen)
	}
	target := -1 // автопилот подсвечивает здание, к которому идёт отряд
	if !m.ManualMode {
		target = t.buildingFor(m.TownPhase)
	}

	var sb strings.Builder
	for y := 0; y < viewH; y++ {
		if y > 0 {
			sb.WriteString("\n")
		}
		if y >= vh {
			sb.WriteString(strings.Repeat(" ", viewW))
			continue
		}
		row := make([]townGlyph, 0, viewW)
		for x := 0; x < vw; x++ {
			row = append(row, m.townGlyphAt(t, texts, target, Point{startX + x, startY + y}))
		}
		for x := vw; x < viewW; x++ {
			row = append(row, townGlyph{r: ' '})
		}

		// Склеиваем одинаковые по стилю символы.
		for i := 0; i < len(row); {
			j := i
			var run strings.Builder
			for j < len(row) && row[j].fg == row[i].fg && row[j].bg == row[i].bg && row[j].bold == row[i].bold {
				run.WriteRune(row[j].r)
				j++
			}
			if row[i].fg == 0 && row[i].bg == 0 && !row[i].bold {
				sb.WriteString(run.String())
			} else {
				sb.WriteString(townStyle(row[i]).Render(run.String()))
			}
			i = j
		}
	}
	return sb.String()
}

func (m Model) townGlyphAt(t *TownMap, texts [][3][]rune, target int, p Point) townGlyph {
	if p == m.TownPos {
		return townGlyph{r: '@', fg: 226, bold: true}
	}
	c := t.Cells[p.Y][p.X]
	g := townGlyph{r: c.R, fg: c.Color, bold: c.Bold}

	if c.Slot != 0 && int(c.Bld) >= 0 && int(c.Bld) < len(texts) {
		line := texts[c.Bld][c.Slot-1]
		g.r = ' '
		if int(c.SlotX) < len(line) {
			g.r = line[c.SlotX]
		}
		switch {
		case (c.Slot == 1 || c.Slot == 2) && int(c.Bld) == target:
			g.fg, g.bg, g.bold = 16, 220, true // цель автопилота подсвечена
		case c.Slot == 1 || c.Slot == 2:
			g.fg, g.bg, g.bold = 255, 238, true
		default: // c.Slot == 3 (уровень/этаж)
			g.fg, g.bg = 244, 238
		}
	}
	return g
}