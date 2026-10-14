package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// ВРАТА ПОДЗЕМЬЯ: ДВА ВИДА ВХОДА
// ============================================================
//
//  1. «Продолжить спуск» — на самый глубокий достигнутый этаж.
//  2. «Пройденный этаж» — на один из этажей выше: он заново заселяется небольшими
//     стаями монстров, подходящих этому этажу, и служит для набора опыта новичками.
//
// Автопилот выбирает вход сам (см. planDeparture), игрок — в карточке Врат.

const (
	gateModeMain   = 0
	gateModeFloors = 1
)

// GateFloorInfo — сведения о пройденном этаже для списка выбора.
type GateFloorInfo struct {
	Floor    int
	Biome    BiomeType
	RecLevel int
	Packs    int // живых стай, известных на этаже
	Explored int // % исследованного
}

// GateServiceState — карточка Врат подземелья.
type GateServiceState struct {
	Mode          int
	Cursor        int // 0 — продолжить спуск, 1 — пройденный этаж
	FloorCursor   int // выбранная строка списка этажей
	Floors        []GateFloorInfo
	Plan          departurePlan // рекомендация автопилота
	ActionSummary string
}

// floorStateExplored — процент исследованной (непроходимые клетки не считаются) части сохранённого этажа.
func floorStateExplored(fs *FloorState) int {
	if fs == nil {
		return 0
	}
	open, seen := 0, 0
	for y := range fs.Grid {
		for x, t := range fs.Grid[y] {
			if t == TileWall {
				continue
			}
			open++
			if y < len(fs.Explored) && x < len(fs.Explored[y]) && fs.Explored[y][x] {
				seen++
			}
		}
	}
	if open == 0 {
		return 0
	}
	return seen * 100 / open
}

func (m *Model) initGateScreen() {
	st := GateServiceState{Plan: m.planDeparture()}
	for _, f := range m.trainingFloors() {
		info := GateFloorInfo{Floor: f, Biome: getBiome(f).Name, RecLevel: recommendedLevel(f)}
		if fs, ok := m.VisitedFloors[f]; ok && fs != nil {
			for _, p := range fs.Packs {
				if p.LivingCount() > 0 {
					info.Packs++
				}
			}
			info.Explored = floorStateExplored(fs)
		}
		st.Floors = append(st.Floors, info)
	}

	switch {
	case st.Plan.Training:
		for i, fi := range st.Floors {
			if fi.Floor == st.Plan.Floor {
				st.FloorCursor = i
			}
		}
		st.Cursor = 1
	case len(st.Floors) > 0:
		st.FloorCursor = len(st.Floors) - 1
	}
	if !m.ManualMode {
		st.ActionSummary = T(m.Lang, st.Plan.Reason, st.Plan.Args...)
	}
	m.GateState = st
}

// departFromTown выводит отряд из города на этаж плана.
func (m *Model) departFromTown(plan departurePlan) {
	m.syncMaxFloor()
	for _, h := range m.Party {
		h.ReviveCooldown = 0
	}
	m.TownHistory = []string{}
	m.InTown = false
	m.TownWalking, m.TownVisit, m.TownLeaving = false, false, false
	m.PathHistory = []Point{}
	m.LoopDetectCount = 0

	m.TripStartMaxFloor = m.MaxFloor
	m.Floor = clampInt(plan.Floor, 1, max(1, m.MaxFloor))
	m.Training = plan.Training || (m.Floor < m.MaxFloor)
	m.initDungeonForFloor(m.Floor)

	if m.Training {
		m.repopulateTrainingFloor()
		m.saveCurrentFloorState()
		m.addLog(accentStyle.Render(T(m.Lang, "gate.log.training", m.Floor)))
		return
	}

	// Если отряд вернулся на максимальный этаж после отступления, а живых стай там не осталось — дозаселяем
	living := 0
	for _, p := range m.Packs {
		if p != nil && p.LivingCount() > 0 {
			living++
		}
	}
	if living == 0 {
		m.repopulateTrainingFloor()
		m.saveCurrentFloorState()
	}

	m.addLog(accentStyle.Render(T(m.Lang, "town.log.depart")))
}

// stepGateAuto — автопилот у Врат: выполняет решение planDeparture.
func (m *Model) stepGateAuto() {
	m.departFromTown(m.planDeparture())
}

// gateConfirm — подтверждение выбора в ручном режиме.
func (m *Model) gateConfirm() {
	st := &m.GateState
	switch st.Mode {
	case gateModeMain:
		if st.Cursor == 0 {
			m.departFromTown(departurePlan{Floor: max(1, m.MaxFloor)})
			return
		}
		if len(st.Floors) > 0 {
			st.Mode = gateModeFloors
		}
	case gateModeFloors:
		if st.FloorCursor >= 0 && st.FloorCursor < len(st.Floors) {
			m.departFromTown(departurePlan{Floor: st.Floors[st.FloorCursor].Floor, Training: true})
		}
	}
}

// handleGateKey обрабатывает клавиши карточки Врат. Второе значение — «клавиша обработана».
func (m *Model) handleGateKey(key string) bool {
	st := &m.GateState
	switch key {
	case "up", "k":
		if st.Mode == gateModeMain {
			st.Cursor = (st.Cursor + 1) % 2
		} else if n := len(st.Floors); n > 0 {
			st.FloorCursor = (st.FloorCursor + n - 1) % n
		}
		return true
	case "down", "j":
		if st.Mode == gateModeMain {
			st.Cursor = (st.Cursor + 1) % 2
		} else if n := len(st.Floors); n > 0 {
			st.FloorCursor = (st.FloorCursor + 1) % n
		}
		return true
	case "1":
		if st.Mode == gateModeMain {
			st.Cursor = 0
			m.gateConfirm()
			return true
		}
	case "2":
		if st.Mode == gateModeMain {
			st.Cursor = 1
			m.gateConfirm()
			return true
		}
	case "enter", " ":
		m.gateConfirm()
		return true
	case "esc":
		if st.Mode == gateModeFloors {
			st.Mode = gateModeMain
			return true
		}
	}
	return false
}

// weakMark — пометка бойца, отстающего от рекомендуемого уровня.
func (m Model) weakMark(h *Hero) string {
	for _, w := range m.weakHeroes() {
		if w == h {
			return dangerStyle.Render(" ⚠")
		}
	}
	return ""
}

func (m Model) renderGateScreen(viewW, viewH int) string {
	st := m.GateState
	innerW := max(34, viewW)
	var sb strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🕳 " + T(m.Lang, "town.gate"))
	sb.WriteString(fmt.Sprintf("%s  |  %s\n", title, subtleStyle.Render(T(m.Lang, "gate.deepest", max(1, m.MaxFloor)))))

	if m.ManualMode {
		key := "gate.controls"
		if st.Mode == gateModeFloors {
			key = "gate.controls_floors"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(T(m.Lang, key), innerW)) + "\n")
	} else {
		sb.WriteString(m.renderAutoBanner(st.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	if st.Mode == gateModeMain {
		sb.WriteString(m.gateOption(0, st.Cursor == 0, true,
			T(m.Lang, "gate.opt.continue", max(1, m.MaxFloor), recommendedLevel(max(1, m.MaxFloor)))) + "\n")
		sb.WriteString(m.gateOption(1, st.Cursor == 1, len(st.Floors) > 0, T(m.Lang, "gate.opt.training")) + "\n")
		if len(st.Floors) == 0 {
			sb.WriteString(subtleStyle.Render("    "+T(m.Lang, "gate.no_floors")) + "\n")
		}
	} else {
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "gate.pick_floor")) + "\n")
		visible := max(3, viewH-12)
		from := clampInt(st.FloorCursor-visible/2, 0, max(0, len(st.Floors)-visible))
		for i := from; i < len(st.Floors) && i < from+visible; i++ {
			fi := st.Floors[i]
			marker := "  "
			style := subtleStyle
			if i == st.FloorCursor {
				marker = accentStyle.Render("▶ ")
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
			}
			line := T(m.Lang, "gate.floor_row", fi.Floor, T(m.Lang, "biome."+string(fi.Biome)), fi.RecLevel, fi.Packs, fi.Explored)
			sb.WriteString(marker + style.Render(shortenItemName(line, innerW-4)) + "\n")
		}
	}

	// Готовность отряда: уровень каждого бойца против рекомендуемого для самого глубокого этажа.
	rec := recommendedLevel(max(1, m.MaxFloor))
	sb.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "gate.party_ready", rec)) + "\n")
	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %s %s%s\n", padRight(shortenItemName(h.FullName(m.Lang), 22), 22),
			subtleStyle.Render(fmt.Sprintf("%s.%d", T(m.Lang, "ui.level_short"), h.Level)), m.weakMark(h)))
	}
	if st.Plan.Training && m.ManualMode {
		sb.WriteString("\n" + healStyle.Render("✓ "+T(m.Lang, st.Plan.Reason, st.Plan.Args...)) + "\n")
	}
	return padTownLines(sb.String(), viewW, viewH)
}

func (m Model) gateOption(idx int, selected, enabled bool, text string) string {
	marker := "  "
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	if selected {
		marker = accentStyle.Render("▶ ")
		style = style.Bold(true).Underline(true)
	}
	if !enabled {
		style = subtleStyle
	}
	return fmt.Sprintf("%s[%d] %s", marker, idx+1, style.Render(text))
}