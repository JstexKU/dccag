package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderAutoBanner(summary string, width int) string {
	if summary == "" {
		summary = "Нет доступных действий или средств. Отряд осматривает квартал..."
	}
	tag := lipgloss.NewStyle().Background(lipgloss.Color("35")).Foreground(lipgloss.Color("16")).Bold(true).Render(" ⚙️ АВТОПИЛОТ ")
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render(" " + summary)
	return lipgloss.NewStyle().Width(width).Render(tag + text)
}

// 1. РЫНОК
func (m Model) renderMarketScreen(state MarketServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	halfW := (innerW - 3) / 2

	var sb strings.Builder
	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⚖ " + T(m.Lang, "town.market"))
	goldStr := goldStyle.Render(fmt.Sprintf("%dG", m.Gold))
	bagStr := subtleStyle.Render(fmt.Sprintf("%s: %d/%d", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity()))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s  |  🎒 %s\n", titleStr, goldStr, bagStr))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	var leftLines []string
	leftHeader := lipgloss.NewStyle().Bold(true).Render("📦 " + T(m.Lang, "ui.bag"))
	if state.Section == MarketSectionLoot {
		leftHeader = accentStyle.Render("▶ 📦 " + T(m.Lang, "ui.bag"))
	}
	leftLines = append(leftLines, leftHeader, subtleStyle.Render(strings.Repeat("─", halfW)))

	if len(m.Bag) == 0 {
		leftLines = append(leftLines, subtleStyle.Render("  ("+T(m.Lang, "ui.none")+")"))
	} else {
		for i, it := range m.Bag {
			val := it.Value * 2
			marker := "  "
			nameStyle := subtleStyle
			if state.Section == MarketSectionLoot && state.LootIdx == i {
				marker = accentStyle.Render("▶ ")
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
			}
			name := shortenItemName(it.DisplayName(m.Lang), halfW-11)
			leftLines = append(leftLines, fmt.Sprintf("%s%s +%dG", marker, nameStyle.Render(padRight(name, halfW-11)), val))
		}
	}
	sellAllMarker := "  "
	sellAllStyle := healStyle
	if state.Section == MarketSectionLoot && state.LootIdx == len(m.Bag) {
		sellAllMarker = accentStyle.Render("▶ ")
		sellAllStyle = healStyle.Copy().Bold(true).Underline(true)
	}
	leftLines = append(leftLines, "", sellAllMarker+sellAllStyle.Render(fmt.Sprintf("[ %s ]", T(m.Lang, "ui.turn_in"))))

	var rightLines []string
	rightHeader := lipgloss.NewStyle().Bold(true).Render("🛒 " + T(m.Lang, "town.hub_title"))
	if state.Section == MarketSectionShop {
		rightHeader = accentStyle.Render("▶ 🛒 " + T(m.Lang, "town.hub_title"))
	}
	rightLines = append(rightLines, rightHeader, subtleStyle.Render(strings.Repeat("─", halfW)))

	if len(state.Offers) == 0 {
		rightLines = append(rightLines, subtleStyle.Render("  ("+T(m.Lang, "ui.none")+")"))
	} else {
		for i, off := range state.Offers {
			marker := "  "
			itemStyle := goldStyle
			isChosen := (!m.ManualMode && state.Section == MarketSectionShop && state.ShopIdx == i)
			if state.Section == MarketSectionShop && state.ShopIdx == i {
				marker = accentStyle.Render("▶ ")
				itemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
			}
			heroName := "-"
			if off.HeroIdx >= 0 && off.HeroIdx < len(m.Party) {
				heroName = m.Party[off.HeroIdx].DisplayName(m.Lang)
			}
			statDiffStr := fmt.Sprintf("+%d", off.StatDiff)
			if off.IsMissing {
				statDiffStr = subtleStyle.Render(T(m.Lang, "ui.none"))
			} else {
				statDiffStr = healStyle.Render(statDiffStr)
			}
			badge := ""
			if isChosen {
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ КУПЛЕНО")
			}
			name := shortenItemName(off.Item.DisplayName(m.Lang), max(6, halfW-20))
			rightLines = append(rightLines, fmt.Sprintf("%s%s (%s) %sG [%s]%s", marker, itemStyle.Render(name), heroName, goldStyle.Render(fmt.Sprintf("%d", off.Cost)), statDiffStr, badge))
		}
	}

	maxRows := max(len(leftLines), len(rightLines))
	for len(leftLines) < maxRows {
		leftLines = append(leftLines, strings.Repeat(" ", halfW))
	}
	for len(rightLines) < maxRows {
		rightLines = append(rightLines, strings.Repeat(" ", halfW))
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(halfW).Render(strings.Join(leftLines, "\n")),
		" │ ",
		lipgloss.NewStyle().Width(halfW).Render(strings.Join(rightLines, "\n")),
	) + "\n\n")

	if m.ManualMode {
		controls := "[Tab] Секция | [↑/↓] Выбор | [Enter] Сделка | [Пробел] Продать всё | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[Tab] Section | [↑/↓] Select | [Enter] Transact | [Space] Sell All | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 2. МАГИСТРАТ
func (m Model) renderMagistrateScreen(state MagistrateServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🏛 " + T(m.Lang, "town.magistrate"))
	goldStr := goldStyle.Render(fmt.Sprintf("%dG", m.Gold))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStr))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	currentTier := militiaTierForInvestment(m.Legacy.TotalInvested)
	militiaStatus := fmt.Sprintf("🎖 %s: %s (%dG)", T(m.Lang, "stats.militia_tier"), titleStyle.Render(T(m.Lang, currentTier.TitleKey)), m.Legacy.TotalInvested)
	if nxt := nextMilitiaTier(m.Legacy.TotalInvested); nxt != nil {
		militiaStatus += fmt.Sprintf(" → %s (%dG)", T(m.Lang, nxt.TitleKey), nxt.InvestFloor)
	}
	sb.WriteString(militiaStatus + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 "+T(m.Lang, "town.management")+":") + "\n")
	for i, off := range state.Offers {
		marker := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		isChosen := (!m.ManualMode && state.Cursor == i && state.ActionSummary != "")

		if state.Cursor == i {
			marker = accentStyle.Render("▶ ")
			nameStyle = nameStyle.Copy().Bold(true).Underline(true)
		}
		costStyle := goldStyle
		if !off.CanAfford {
			costStyle = subtleStyle
		}
		targetName := T(m.Lang, off.NameKey)
		typeLabel := T(m.Lang, off.DescKey)

		badge := ""
		if isChosen {
			badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ УЛУЧШЕНО")
		}

		var line string
		if off.Target == MagistrateTreasuryGrant {
			line = fmt.Sprintf("%s%s — %s (%s)%s", marker, nameStyle.Render(typeLabel), costStyle.Render(fmt.Sprintf("%dG", off.Cost)), subtleStyle.Render("+100G в Казну"), badge)
		} else {
			line = fmt.Sprintf("%s%s «%s» (%s.%d) — %s%s", marker, nameStyle.Render(typeLabel), targetName, T(m.Lang, "ui.level_short"), off.Level, costStyle.Render(fmt.Sprintf("%dG", off.Cost)), badge)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор улучшения | [Enter] Инвестировать | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select Project | [Enter] Invest | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 3. ХРАМ
func (m Model) renderChurchScreen(state ChurchServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⛪ " + T(m.Lang, m.TownEst.ChurchKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("✨ Омовение и исцеление:") + "\n")
	if len(state.Offers) == 0 {
		sb.WriteString(healStyle.Render("  Все соратники здоровы и полны сил!") + "\n")
	} else {
		for i, off := range state.Offers {
			marker := "  "
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			isChosen := (!m.ManualMode && state.Cursor == i && state.ActionSummary != "")

			if state.Cursor == i {
				marker = accentStyle.Render("▶ ")
				nameStyle = nameStyle.Copy().Bold(true).Underline(true)
			}
			costStr := goldStyle.Render(fmt.Sprintf("%dG", off.Cost))
			if off.Cost == 0 {
				costStr = fountStyle.Render("[БЕСПЛАТНО]")
			} else if !off.CanAfford {
				costStr = subtleStyle.Render(fmt.Sprintf("%dG", off.Cost))
			}
			actionLabel := "Очищение"
			if off.Type == ChurchActionStabilize {
				actionLabel = dangerStyle.Render("Стабилизация")
			}
			badge := ""
			if isChosen {
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ ИСЦЕЛЕНО")
			}
			sb.WriteString(fmt.Sprintf("%s%s: %s [%s] — %s%s\n", marker, actionLabel, nameStyle.Render(off.HeroName), subtleStyle.Render(off.Desc), costStr, badge))
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Исцелить/Очистить | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Heal/Cleanse | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 4. ТАВЕРНА
func (m Model) renderTavernScreen(state TavernServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🍻 " + T(m.Lang, m.TownEst.TavernKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("🛏️ Варианты отдыха:") + "\n")
	for i, off := range state.Offers {
		marker := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		isChosen := (!m.ManualMode && state.Cursor == i && state.ActionSummary != "")

		if state.Cursor == i {
			marker = accentStyle.Render("▶ ")
			nameStyle = nameStyle.Copy().Bold(true).Underline(true)
		}
		costStr := goldStyle.Render(fmt.Sprintf("%dG", off.Cost))
		if !off.CanAfford {
			costStr = subtleStyle.Render(fmt.Sprintf("%dG", off.Cost))
		}
		badge := ""
		if isChosen {
			badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ ОПЛАЧЕНО")
		}
		if off.IsLuxury {
			sb.WriteString(fmt.Sprintf("%s%s — %s (100%% HP & MP)%s\n", marker, nameStyle.Render("Уютные комнаты на втором этаже"), costStr, badge))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s — %s (40%% HP & MP)%s\n", marker, nameStyle.Render("Ночлег на сеновале у очага"), costStr, badge))
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Отдохнуть | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Rest | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 5. ГИЛЬДИЯ
func (m Model) renderGuildScreen(state GuildServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⚔ " + T(m.Lang, m.TownEst.GuildKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 Доукомплектование отряда:") + "\n")
	if len(state.Offers) == 0 {
		sb.WriteString(healStyle.Render("  Отряд укомплектован! Все 5 бойцов в строю.") + "\n")
	} else {
		for i, off := range state.Offers {
			marker := "  "
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			isChosen := (!m.ManualMode && state.Cursor == i && state.ActionSummary != "")

			if state.Cursor == i {
				marker = accentStyle.Render("▶ ")
				nameStyle = nameStyle.Copy().Bold(true).Underline(true)
			}
			clsName := TranslateEnum(m.Lang, "class", string(off.Class)+".short")
			costStr := goldStyle.Render(fmt.Sprintf("%dG", off.Cost))
			recType := "Ветеран"
			if !off.IsVeteran {
				recType = "Ополченец"
				costStr = fountStyle.Render("[БЕСПЛАТНО]")
			} else if !off.CanAfford {
				costStr = subtleStyle.Render(fmt.Sprintf("%dG", off.Cost))
			}
			badge := ""
			if isChosen {
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ ПРИНЯТ В ОТРЯД")
			}
			sb.WriteString(fmt.Sprintf("%s%s (%s, слот %d) — %s%s\n", marker, nameStyle.Render(recType), clsName, off.SlotIdx+1, costStr, badge))
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Нанять | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Recruit | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 6. КУЗНИЦА И КОЖЕВНИК
func (m Model) renderForgeScreen(title string, state ForgeServiceState, isTannery bool) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("⚒ Доступные улучшения:") + "\n")
	rowIdx := 0
	if isTannery && m.BagLevel < len(bagUpgrades)-1 {
		nextBag := bagUpgrades[m.BagLevel+1]
		marker := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		isChosen := (!m.ManualMode && state.Cursor == 0 && state.ActionSummary != "")

		if state.Cursor == rowIdx {
			marker = accentStyle.Render("▶ ")
			nameStyle = nameStyle.Copy().Bold(true).Underline(true)
		}
		costStr := goldStyle.Render(fmt.Sprintf("%dG", nextBag.Cost))
		badge := ""
		if isChosen {
			badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ СШИТО")
		}
		sb.WriteString(fmt.Sprintf("%s%s (%d слотов) — %s%s\n", marker, nameStyle.Render(T(m.Lang, nextBag.NameKey)), nextBag.Capacity, costStr, badge))
		rowIdx++
	}

	if len(state.Offers) == 0 && (!isTannery || m.BagLevel >= len(bagUpgrades)-1) {
		sb.WriteString(subtleStyle.Render("  Нет подходящих предметов для заточки.") + "\n")
	} else {
		for i, off := range state.Offers {
			marker := "  "
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			targetIdx := rowIdx + i
			isChosen := (!m.ManualMode && state.Cursor == targetIdx && state.ActionSummary != "")

			if state.Cursor == targetIdx {
				marker = accentStyle.Render("▶ ")
				nameStyle = nameStyle.Copy().Bold(true).Underline(true)
			}
			heroName := m.Party[off.HeroIdx].DisplayName(m.Lang)
			costStr := goldStyle.Render(fmt.Sprintf("%dG", off.Cost))
			if !off.CanAfford {
				costStr = subtleStyle.Render(fmt.Sprintf("%dG", off.Cost))
			}
			badge := ""
			if isChosen {
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ ЗАТОЧЕНО")
			}
			sb.WriteString(fmt.Sprintf("%s%s (%s +%d) [%s] — %s%s\n", marker, nameStyle.Render(off.ItemName), T(m.Lang, "slot."+string(off.Slot)), off.Level+1, heroName, costStr, badge))
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Улучшить | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Forge | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// 7. ЛАВКА АЛХИМИКА
func (m Model) renderAlchemistScreen(state AlchemistServiceState) string {
	boxW := max(44, m.TermWidth-4)
	innerW := max(38, boxW-4)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🧪 " + T(m.Lang, m.TownEst.AlchemistKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("⚗️️ Мутации и зелья:") + "\n")
	if len(state.Offers) == 0 {
		sb.WriteString(subtleStyle.Render("  Пояса полны, мутации недоступны.") + "\n")
	} else {
		for i, off := range state.Offers {
			marker := "  "
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			isChosen := (!m.ManualMode && state.Cursor == i && state.ActionSummary != "")

			if state.Cursor == i {
				marker = accentStyle.Render("▶ ")
				nameStyle = nameStyle.Copy().Bold(true).Underline(true)
			}
			costStr := goldStyle.Render(fmt.Sprintf("%dG", off.Cost))
			if !off.CanAfford {
				costStr = subtleStyle.Render(fmt.Sprintf("%dG", off.Cost))
			}
			cat := "Мутация"
			if off.Type == AlchemistActionPotion {
				cat = "Зелье"
			}
			badge := ""
			if isChosen {
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ ПРИНЯТО")
			}
			sb.WriteString(fmt.Sprintf("%s[%s] %s (%s) — %s%s\n", marker, cat, nameStyle.Render(off.ItemTitle), off.HeroName, costStr, badge))
		}
	}

	sb.WriteString("\n" + subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Купить/Принять | [Esc] Выйти в поход"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Buy/Apply | [Esc] Depart"
		}
		sb.WriteString(subtleStyle.Render(controls))
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW))
	}

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}
