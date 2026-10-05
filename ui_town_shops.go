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
func (m Model) renderMarketScreen(state MarketServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	halfW := (innerW - 3) / 2

	var sb strings.Builder
	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⚖ " + T(m.Lang, "town.market"))
	goldStr := goldStyle.Render(fmt.Sprintf("%dG", m.Gold))
	bagStr := subtleStyle.Render(fmt.Sprintf("%s: %d/%d", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity()))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s  |  🎒 %s\n", titleStr, goldStr, bagStr))

	if m.ManualMode {
		controls := "[Tab] Секция | [↑/↓] Выбор | [Enter] Сделка | [Пробел] Продать всё | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[Tab] Section | [↑/↓] Select | [Enter] Transact | [Space] Sell All | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	availRows := max(3, viewH-6)

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
			if len(leftLines) >= availRows {
				break
			}
			val := it.Value * 2
			marker := "  "
			nameStyle := subtleStyle
			if state.Section == MarketSectionLoot && state.LootIdx == i {
				marker = accentStyle.Render("▶ ")
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
			}
			name := shortenItemName(it.DisplayName(m.Lang), max(4, halfW-11))
			leftLines = append(leftLines, fmt.Sprintf("%s%s +%dG", marker, nameStyle.Render(padRight(name, halfW-11)), val))
		}
	}
	sellAllMarker := "  "
	sellAllStyle := healStyle
	if state.Section == MarketSectionLoot && state.LootIdx == len(m.Bag) {
		sellAllMarker = accentStyle.Render("▶ ")
		sellAllStyle = healStyle.Copy().Bold(true).Underline(true)
	}
	if len(leftLines) < availRows {
		leftLines = append(leftLines, sellAllMarker+sellAllStyle.Render(fmt.Sprintf("[ %s ]", T(m.Lang, "ui.turn_in"))))
	}

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
			if len(rightLines) >= availRows {
				break
			}
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
				badge = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓")
			}
			name := shortenItemName(off.Item.DisplayName(m.Lang), max(4, halfW-20))
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
	))

	return padTownLines(sb.String(), viewW, viewH)
}

// 2. МАГИСТРАТ
func (m Model) renderMagistrateScreen(state MagistrateServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🏛 " + T(m.Lang, "town.magistrate"))
	goldStr := goldStyle.Render(fmt.Sprintf("%dG", m.Gold))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStr))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Инвестировать | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select Project | [Enter] Invest | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	currentTier := militiaTierForInvestment(m.Legacy.TotalInvested)
	militiaStatus := fmt.Sprintf("🎖 %s: %s (%dG)", T(m.Lang, "stats.militia_tier"), titleStyle.Render(T(m.Lang, currentTier.TitleKey)), m.Legacy.TotalInvested)
	if nxt := nextMilitiaTier(m.Legacy.TotalInvested); nxt != nil {
		militiaStatus += fmt.Sprintf(" → %s (%dG)", T(m.Lang, nxt.TitleKey), nxt.InvestFloor)
	}
	sb.WriteString(militiaStatus + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 "+T(m.Lang, "town.management")+":") + "\n")
	availRows := max(2, viewH-8)
	displayed := 0
	for i, off := range state.Offers {
		if displayed >= availRows {
			break
		}
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

		cleanType := shortenItemName(typeLabel, 26)
		cleanTarget := shortenItemName(targetName, 24)

		var line string
		if off.Target == MagistrateTreasuryGrant {
			line = fmt.Sprintf("%s%s — %s (%s)%s", marker, nameStyle.Render(cleanType), costStyle.Render(fmt.Sprintf("%dG", off.Cost)), subtleStyle.Render("+100G в Казну"), badge)
		} else {
			line = fmt.Sprintf("%s%s «%s» (%s.%d) — %s%s", marker, nameStyle.Render(cleanType), cleanTarget, T(m.Lang, "ui.level_short"), off.Level, costStyle.Render(fmt.Sprintf("%dG", off.Cost)), badge)
		}
		sb.WriteString(line + "\n")
		displayed++
	}

	return padTownLines(sb.String(), viewW, viewH)
}

// 3. ХРАМ
func (m Model) renderChurchScreen(state ChurchServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⛪ " + T(m.Lang, m.TownEst.ChurchKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Исцелить/Очистить | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Heal/Cleanse | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("✨ Омовение и исцеление:") + "\n")
	availRows := max(2, viewH-7)
	displayed := 0
	if len(state.Offers) == 0 {
		sb.WriteString(healStyle.Render("  Все соратники здоровы и полны сил!") + "\n")
	} else {
		for i, off := range state.Offers {
			if displayed >= availRows {
				break
			}
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
			cleanHero := shortenItemName(off.HeroName, 18)
			cleanDesc := shortenItemName(off.Desc, 22)
			sb.WriteString(fmt.Sprintf("%s%s: %s [%s] — %s%s\n", marker, actionLabel, nameStyle.Render(cleanHero), subtleStyle.Render(cleanDesc), costStr, badge))
			displayed++
		}
	}

	return padTownLines(sb.String(), viewW, viewH)
}

// 4. ТАВЕРНА
func (m Model) renderTavernScreen(state TavernServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🍻 " + T(m.Lang, m.TownEst.TavernKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Отдохнуть | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Rest | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

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
		roomDesc := "Ночлег на сеновале у очага"
		pctDesc := "(40% HP & MP)"
		if off.IsLuxury {
			roomDesc = "Уютные комнаты на втором этаже"
			pctDesc = "(100% HP & MP)"
		}
		sb.WriteString(fmt.Sprintf("%s%s — %s %s%s\n", marker, nameStyle.Render(roomDesc), costStr, pctDesc, badge))
	}

	return padTownLines(sb.String(), viewW, viewH)
}

// 5. ГИЛЬДИЯ
func (m Model) renderGuildScreen(state GuildServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⚔ " + T(m.Lang, m.TownEst.GuildKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Нанять | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Recruit | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 Доукомплектование отряда:") + "\n")
	availRows := max(2, viewH-7)
	displayed := 0
	if len(state.Offers) == 0 {
		sb.WriteString(healStyle.Render("  Отряд укомплектован! Все 5 бойцов в строю.") + "\n")
	} else {
		for i, off := range state.Offers {
			if displayed >= availRows {
				break
			}
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
			displayed++
		}
	}

	return padTownLines(sb.String(), viewW, viewH)
}

// 6. КУЗНИЦА И КОЖЕВНИК
func (m Model) renderForgeScreen(title string, state ForgeServiceState, isTannery bool, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Улучшить | [Esc] Дальше"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Forge | [Esc] Next"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("⚒ Доступные улучшения:") + "\n")
	availRows := max(2, viewH-7)
	displayed := 0
	rowIdx := 0

	// 1. Пошив сумки
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
		bagName := shortenItemName(T(m.Lang, nextBag.NameKey), 28)
		sb.WriteString(fmt.Sprintf("%s%s (%d слотов) — %s%s\n", marker, nameStyle.Render(bagName), nextBag.Capacity, costStr, badge))
		rowIdx++
		displayed++
	}

	// 2. Список заточки и выделки экипировки
	if len(state.Offers) == 0 && (!isTannery || m.BagLevel >= len(bagUpgrades)-1) {
		sb.WriteString(subtleStyle.Render("  Нет подходящих предметов для заточки.") + "\n")
	} else {
		for i, off := range state.Offers {
			if displayed >= availRows {
				break
			}
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
			slotName := T(m.Lang, "slot."+string(off.Slot))
			cleanItem := shortenItemName(off.ItemName, max(10, innerW-42))
			sb.WriteString(fmt.Sprintf("%s%s (%s +%d) [%s] — %s%s\n", marker, nameStyle.Render(cleanItem), slotName, off.Level+1, heroName, costStr, badge))
			displayed++
		}
	}

	return padTownLines(sb.String(), viewW, viewH)
}

// 7. ЛАВКА АЛХИМИКА
func (m Model) renderAlchemistScreen(state AlchemistServiceState, viewW, viewH int) string {
	innerW := max(34, viewW)
	var sb strings.Builder

	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🧪 " + T(m.Lang, m.TownEst.AlchemistKey))
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s\n", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))

	if m.ManualMode {
		controls := "[↑/↓] Выбор | [Enter] Купить/Принять | [Esc] Выйти в поход"
		if m.Lang == LangEN {
			controls = "[↑/↓] Select | [Enter] Buy/Apply | [Esc] Depart"
		}
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(renderAutoBanner(state.ActionSummary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("⚗ Мутации и зелья:") + "\n")
	availRows := max(2, viewH-7)
	displayed := 0
	if len(state.Offers) == 0 {
		sb.WriteString(subtleStyle.Render("  Пояса полны, мутации недоступны.") + "\n")
	} else {
		for i, off := range state.Offers {
			if displayed >= availRows {
				break
			}
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
			cleanTitle := shortenItemName(off.ItemTitle, max(10, innerW-38))
			sb.WriteString(fmt.Sprintf("%s[%s] %s (%s) — %s%s\n", marker, cat, nameStyle.Render(cleanTitle), off.HeroName, costStr, badge))
			displayed++
		}
	}

	return padTownLines(sb.String(), viewW, viewH)
}