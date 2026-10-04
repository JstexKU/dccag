package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type townDisplayMode int

const (
	townDisplayWide townDisplayMode = iota
	townDisplayMedium
	townDisplayCompact
)

func townDisplayModeFor(width, height int) townDisplayMode {
	switch {
	case width >= 72 && height >= 12:
		return townDisplayWide
	case width >= 48 && height >= 9:
		return townDisplayMedium
	default:
		return townDisplayCompact
	}
}

func (m Model) renderTownHub(viewW, viewH int) string {
	mode := townDisplayModeFor(viewW, viewH)
	switch mode {
	case townDisplayWide:
		return m.renderTownHubWide(viewW, viewH)
	case townDisplayMedium:
		return m.renderTownHubMedium(viewW, viewH)
	default:
		return m.renderTownHubCompact(viewW, viewH)
	}
}

type townAction struct {
	phase TownPhase
	icon  string
	name  string
	level int
}

func (m Model) townActions() []townAction {
	return []townAction{
		{TownPhaseSellLoot, "⚖", T(m.Lang, "town.market"), 0},
		{TownPhaseMagistrate, "🏛", T(m.Lang, "town.magistrate"), 0},
		{TownPhaseChurch, "⛪", T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel},
		{TownPhaseTavern, "🍻", T(m.Lang, m.TownEst.TavernKey), m.Legacy.TavernLevel},
		{TownPhaseGuild, "⚔", T(m.Lang, m.TownEst.GuildKey), 0},
		{TownPhaseSmithy, "⚒️️", T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel},
		{TownPhaseTannery, "🎒", T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel},
		{TownPhaseAlchemist, "🧪", T(m.Lang, m.TownEst.AlchemistKey), 0},
	}
}

func (m Model) renderTownAction(action townAction, width int, detail bool) string {
	active := m.TownPhase == action.phase
	name := shortenItemName(action.name, max(3, width-7))

	label := fmt.Sprintf("%s %s", action.icon, name)
	if detail && action.level > 0 {
		label = fmt.Sprintf("%s %s.%d", action.icon, name, action.level)
	}

	label = padRightTruncate(label, max(1, width-2))
	if active {
		return lipgloss.NewStyle().
			Width(width).
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("236")).
			Bold(true).
			Render("► " + label)
	}
	return lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color("244")).
		Render("  " + label)
}

func (m Model) renderTownHistory(width, maxLines int) []string {
	if maxLines <= 0 {
		return nil
	}
	lines := []string{
		subtleStyle.Render(padRightTruncate("── "+T(m.Lang, "town.management")+" ──", width)),
	}
	for _, act := range m.TownHistory {
		if len(lines) >= maxLines {
			break
		}
		lines = append(lines, padRightTruncate("• "+shortenItemName(act, max(1, width-2)), width))
	}
	return lines
}

func (m Model) renderTownHubWide(viewW, viewH int) string {
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	actionW := max(16, (viewW-8)/3)

	var sb strings.Builder
	sb.WriteString(header.Render(padRightTruncate(T(m.Lang, "town.hub_title"), viewW)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(
		fmt.Sprintf("BAG %d/%d   |   SMITHY %d   TANNERY %d   CHURCH %d   TAVERN %d",
			len(m.Bag), m.currentBagCapacity(),
			m.Legacy.SmithyLevel, m.Legacy.TanneryLevel,
			m.Legacy.ChurchLevel, m.Legacy.TavernLevel), viewW,
	)) + "\n\n")

	actions := m.townActions()
	for row := 0; row < 3; row++ {
		var cells []string
		for col := 0; col < 3; col++ {
			i := row*3 + col
			if i >= len(actions) {
				break
			}
			cells = append(cells, m.renderTownAction(actions[i], actionW, true))
		}
		sb.WriteString(strings.Join(cells, "  ") + "\n")
	}

	historyLines := m.renderTownHistory(viewW, max(0, viewH-7))
	if len(historyLines) > 0 {
		sb.WriteString("\n")
		sb.WriteString(strings.Join(historyLines, "\n"))
	}
	return padTownLines(sb.String(), viewW, viewH)
}

func (m Model) renderTownHubMedium(viewW, viewH int) string {
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	actionW := max(18, (viewW-4)/2)

	var sb strings.Builder
	sb.WriteString(header.Render(padRightTruncate(T(m.Lang, "town.hub_title"), viewW)) + "\n\n")

	actions := m.townActions()
	for row := 0; row < 4; row++ {
		left := row * 2
		right := left + 1
		sb.WriteString(m.renderTownAction(actions[left], actionW, true))
		if right < len(actions) {
			sb.WriteString("  " + m.renderTownAction(actions[right], actionW, true))
		}
		sb.WriteString("\n")
	}

	historyLines := m.renderTownHistory(viewW, max(0, viewH-6))
	if len(historyLines) > 0 {
		sb.WriteString("\n" + strings.Join(historyLines, "\n"))
	}
	return padTownLines(sb.String(), viewW, viewH)
}

func (m Model) renderTownHubCompact(viewW, viewH int) string {
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	var sb strings.Builder
	sb.WriteString(header.Render(padRightTruncate(T(m.Lang, "town.hub_title"), viewW)) + "\n")

	for _, action := range m.townActions() {
		name := shortenItemName(action.name, max(3, viewW-5))
		level := ""
		if action.level > 0 {
			level = fmt.Sprintf(" %d", action.level)
		}
		active := " "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		if m.TownPhase == action.phase {
			active = "►"
			style = style.Foreground(lipgloss.Color("226")).Bold(true)
		}
		line := fmt.Sprintf("%s %s%s", active, action.icon, name+level)
		sb.WriteString(style.Render(padRightTruncate(line, viewW)) + "\n")
	}

	historyLines := m.renderTownHistory(viewW, max(0, viewH-10))
	if len(historyLines) > 0 {
		sb.WriteString(strings.Join(historyLines, "\n"))
	}
	return padTownLines(sb.String(), viewW, viewH)
}

func padTownLines(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	out := make([]string, height)
	for i := range out {
		if i < len(lines) && i < height {
			out[i] = padRightTruncate(lines[i], width)
		} else {
			out[i] = strings.Repeat(" ", width)
		}
	}
	return strings.Join(out, "\n")
}
