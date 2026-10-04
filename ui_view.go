package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m Model) renderRightContentLines(innerRightW, maxLines int) []string {
	var lines []string
	if m.Combat != nil {
		lines = append(lines, dangerStyle.Render(shortenItemName(T(m.Lang, "combat.enemy_pack")+":", innerRightW)))
		availRows := max(1, maxLines-1)
		displayed := 0
		for _, mob := range m.Combat.Pack.Members {
			if displayed >= availRows {
				break
			}
			hpText := fmt.Sprintf("%d/%d", mob.HP, mob.MaxHP)
			if mob.IsDead {
				hpText = T(m.Lang, "ui.dead")
			}
			barWidth := 4
			mobBar := renderBar(mob.HP, mob.MaxHP, barWidth, lipgloss.Color("196"), lipgloss.Color("238"))
			hpBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("[%s]", hpText))
			rightBlock := fmt.Sprintf("%s %s", mobBar, hpBadge)
			rightBlockWidth := (barWidth + 2) + 1 + (runewidth.StringWidth(hpText) + 2)
			availName := max(4, innerRightW-rightBlockWidth-3)
			mobName := shortenItemName(T(m.Lang, mob.NameKey), availName)
			mobNamePadded := padRight(mobName, availName)
			marker := "•"
			if m.ManualMode && mob == m.pickTarget() {
				marker = "▶"
			}
			lines = append(lines, fmt.Sprintf("%s %s %s", marker, mobNamePadded, rightBlock))
			displayed++
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render(shortenItemName(T(m.Lang, "town.management")+":", innerRightW)))
		lines = append(lines, shortenItemName(fmt.Sprintf("⚒ %s: %s.%d", T(m.Lang, m.TownEst.SmithyKey), T(m.Lang, "ui.level_short"), m.Legacy.SmithyLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🎒 %s: %s.%d", T(m.Lang, m.TownEst.TanneryKey), T(m.Lang, "ui.level_short"), m.Legacy.TanneryLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🏛️ %s: %s.%d", T(m.Lang, m.TownEst.ChurchKey), T(m.Lang, "ui.level_short"), m.Legacy.ChurchLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🍻 %s: %s.%d", T(m.Lang, m.TownEst.TavernKey), T(m.Lang, "ui.level_short"), m.Legacy.TanneryLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("%s: %d/%d", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity()), innerRightW))
	} else {
		lines = append(lines, accentStyle.Render(shortenItemName(T(m.Lang, "ui.scouting")+":", innerRightW)))
		lines = append(lines, shortenItemName(fmt.Sprintf("%s: %d", T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("%s: %d", T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened), innerRightW))
	}
	return lines
}

func (m Model) renderEnemyStrip(width int) string {
	if m.Combat == nil {
		if m.InTown {
			return subtleStyle.Render(shortenItemName(T(m.Lang, "town.management")+
				fmt.Sprintf(" | ⚒%d 🎒%d 🏛%d 🍻%d",
					m.Legacy.SmithyLevel, m.Legacy.TanneryLevel,
					m.Legacy.ChurchLevel, m.Legacy.TavernLevel), width))
		}
		return subtleStyle.Render(shortenItemName(
			fmt.Sprintf("%s: %d | %s: %d",
				T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps,
				T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened), width,
		))
	}

	var parts []string
	if width >= 55 {
		parts = append(parts, dangerStyle.Render(T(m.Lang, "combat.enemy_pack")+":"))
	} else {
		parts = append(parts, dangerStyle.Render("⚔"))
	}

	for _, mob := range m.Combat.Pack.Members {
		if mob.IsDead {
			continue
		}
		hpStr := fmt.Sprintf("%s %d/%d", T(m.Lang, mob.NameKey), mob.HP, mob.MaxHP)
		if m.ManualMode && mob == m.pickTarget() {
			hpStr = "▶" + hpStr
		}
		parts = append(parts, hpStr)
	}
	joined := strings.Join(parts, "  ")
	return shortenItemName(joined, width)
}

func (m Model) View() string {
	if m.State == StateMenu {
		return m.renderMenuScreen()
	}
	if m.State == StateDefeat {
		return m.renderStatsScreen(T(m.Lang, "defeat.title"), lipgloss.Color("196"))
	}
	if m.State == StateStatsManual {
		return m.renderStatsScreen(T(m.Lang, "stats.manual_title"), lipgloss.Color("214"))
	}
	if m.State == StateInfoBook {
		return m.renderInfoBookScreen()
	}
	if m.State == StateArmory {
		return m.renderArmoryScreen()
	}
	if m.State == StateTactics {
		return m.renderTacticsScreen()
	}
	if m.State == StateCreator {
		return m.renderCreatorScreen()
	}

	termW := max(38, m.TermWidth)
	termH := max(22, m.TermHeight)

	layout := detectLayout(termW, termH)
	var output string
	switch layout {
	case LayoutLandscape:
		output = m.renderLandscape(termW, termH)
	case LayoutPortraitWide:
		output = m.renderPortraitWide(termW, termH)
	default:
		output = m.renderPortrait(termW, termH)
	}

	// 1. Убираем любые CRLF-артефакты Windows
	cleanOutput := strings.ReplaceAll(output, "\r\n", "\n")

	// 2. Гарантируем строгую фиксацию строк по высоте окна терминала
	return truncateLines(cleanOutput, termH)
}

func (m Model) renderLandscape(termW, termH int) string {
	usableW := max(38, termW-2)

	cardsPerRow := landscapeCardsPerRow(usableW)
	cardWidth := max(16, usableW/cardsPerRow)

	cardMode := CardCompact
	if termH >= 48 && termW >= 120 {
		cardMode = detectCardMode(cardWidth - 4)
	} else if termH >= 48 && termW >= 95 {
		cardMode = CardMedium
	}

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardWidth, cardMode))
	}
	cards = append(cards, m.renderPartyBanner(cardWidth, cardMode))

	var cardRows []string
	for i := 0; i < len(cards); i += cardsPerRow {
		end := min(len(cards), i+cardsPerRow)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	middleTier := lipgloss.JoinVertical(lipgloss.Left, cardRows...)
	middleH := lipgloss.Height(middleTier)

	controlsH := 1

	logH := 6
	if termH >= 38 {
		logH = 8
	} else if termH < 26 {
		logH = 4
	}

	availableH := termH - controlsH - logH - 1

	if availableH < middleH+6 {
		middleH = max(6, availableH-6)
		middleTier = truncateLines(middleTier, middleH)
		middleH = lipgloss.Height(middleTier)
	}

	topH := max(5, availableH-middleH)

	sideW := max(24, min(40, int(float64(usableW)*0.28)))
	mapBoxW := max(16, usableW-sideW)
	boxInnerW := max(1, mapBoxW-2)
	boxInnerH := max(1, topH-2)

	mapStr := m.renderMap(boxInnerW, boxInnerH)
	mapLines := truncateLines(mapStr, boxInnerH)
	leftMapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(boxInnerW).
		Height(boxInnerH).
		Render(mapLines)

	sidebarInnerW := max(16, sideW-2)
	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("%s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("%s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}

	questTitleShort := shortenItemName(T(m.Lang, m.CurrentQuest.TitleKey), max(8, sidebarInnerW-10))
	statusBadge := subtleStyle.Render(fmt.Sprintf("[%d/%d]", m.CurrentQuest.Current, m.CurrentQuest.TargetCount))
	if m.CurrentQuest.Completed {
		statusBadge = healStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.turn_in")))
	}

	var sbLines []string
	sbLines = append(sbLines, fmt.Sprintf("🏰 %s", titleStyle.Render(shortenItemName(floorTag, sidebarInnerW))))
	if boxInnerH >= 5 {
		sbLines = append(sbLines, fmt.Sprintf("💰 %s: %s", T(m.Lang, "ui.treasury"), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	}
	if boxInnerH >= 6 {
		sbLines = append(sbLines, fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge))
		sbLines = append(sbLines, subtleStyle.Render(strings.Repeat("─", sidebarInnerW)))
	}
	availRows := max(1, boxInnerH-len(sbLines))
	sbLines = append(sbLines, m.renderRightContentLines(sidebarInnerW, availRows)...)

	sidebarContent := truncateLines(strings.Join(sbLines, "\n"), boxInnerH)

	rightPane := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(sidebarInnerW).
		Height(boxInnerH).
		Render(sidebarContent)

	topTier := lipgloss.JoinHorizontal(lipgloss.Top, leftMapBox, rightPane)
	logBox := m.renderLogBox(usableW-2, logH)
	controls := m.renderControls(termW - 2)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		logBox,
		controls,
	)
}

func (m Model) renderPortraitWide(termW, termH int) string {
	usableW := termW - 2
	controlsH := 1

	cols := 2
	cardW := usableW / cols

	targetMode := CardCompact
	if termH >= 56 && cardW >= 24 {
		targetMode = CardMedium
	}

	exactCardsH := 3 * cardOuterHeight(targetMode)
	nonDynamicH := 1 + 1 + exactCardsH + controlsH
	freeH := max(8, termH-nonDynamicH)

	logH := max(4, min(8, int(float64(freeH)*0.36)))
	mapH := max(6, freeH-logH)

	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("🏰 %s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("🏰 %s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}
	headerLine := fmt.Sprintf(
		"%s | 💰 %dG | 📜 %s",
		shortenItemName(floorTag, usableW-30),
		m.Gold,
		shortenItemName(T(m.Lang, m.CurrentQuest.TitleKey), 14),
	)
	headerBox := titleStyle.Render(shortenItemName(headerLine, usableW))

	mapInnerH := max(1, mapH-2)
	mapStr := m.renderMap(usableW-2, mapInnerH)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 2).
		Height(mapInnerH).
		Render(truncateLines(mapStr, mapInnerH))

	enemyStrip := m.renderEnemyStrip(usableW)

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardW, targetMode))
	}
	cards = append(cards, m.renderPartyBanner(cardW, targetMode))

	var cardRows []string
	for i := 0; i < len(cards); i += cols {
		end := min(len(cards), i+cols)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	cardsBlock := lipgloss.JoinVertical(lipgloss.Left, cardRows...)

	logBox := m.renderLogBox(usableW-2, logH)
	controls := m.renderControls(termW - 2)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerBox,
		mapBox,
		enemyStrip,
		cardsBlock,
		logBox,
		controls,
	)
}

func (m Model) renderPortrait(termW, termH int) string {
	usableW := termW - 2
	controlsH := 1

	cols := 2
	cardW := usableW / cols

	exactCardsH := 3 * cardOuterHeight(CardCompact)
	nonDynamicH := 1 + 1 + exactCardsH + controlsH
	freeH := max(8, termH-nonDynamicH)

	logH := max(3, min(8, int(float64(freeH)*0.38)))
	mapH := max(5, freeH-logH)

	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("🏰 %s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("🏰 %s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}
	headerLine := fmt.Sprintf("%s | 💰 %dG", shortenItemName(floorTag, usableW-14), m.Gold)
	headerBox := titleStyle.Render(shortenItemName(headerLine, usableW))

	mapInnerH := max(1, mapH-2)
	mapStr := m.renderMap(usableW-2, mapInnerH)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 2).
		Height(mapInnerH).
		Render(truncateLines(mapStr, mapInnerH))

	enemyStrip := m.renderEnemyStrip(usableW)

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardW, CardCompact))
	}
	cards = append(cards, m.renderPartyBanner(cardW, CardCompact))

	var cardRows []string
	for i := 0; i < len(cards); i += cols {
		end := min(len(cards), i+cols)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	cardsBlock := lipgloss.JoinVertical(lipgloss.Left, cardRows...)

	logBox := m.renderLogBox(usableW-2, logH)
	controls := m.renderControls(termW - 2)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerBox,
		mapBox,
		enemyStrip,
		cardsBlock,
		logBox,
		controls,
	)
}

func (m Model) renderLogBox(innerW, boxH int) string {
	if boxH < 3 {
		boxH = 3
	}
	contentH := max(1, boxH-3)

	var logContent strings.Builder
	totalLogs := len(m.Logs)
	endIdx := totalLogs - m.LogScroll
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	if endIdx < contentH {
		endIdx = min(totalLogs, contentH)
	}
	startIdx := max(0, endIdx-contentH)

	for i := 0; i < contentH; i++ {
		curIdx := startIdx + i
		if curIdx < endIdx && curIdx < totalLogs {
			logContent.WriteString(fmt.Sprintf("> %s\n", shortenItemName(m.Logs[curIdx], innerW)))
		} else {
			logContent.WriteString("\n")
		}
	}

	scrollBadge := ""
	if m.LogScroll > 0 {
		scrollBadge = fmt.Sprintf(" (-%d)", m.LogScroll)
	}
	logBoxTitle := fmt.Sprintf("📜 %s%s", T(m.Lang, "ui.chronicles"), scrollBadge)

	body := fmt.Sprintf(
		"%s\n%s",
		subtleStyle.Render(shortenItemName(logBoxTitle, innerW)),
		strings.TrimRight(logContent.String(), "\n"),
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("242")).
		Width(innerW).
		Height(boxH - 1).
		Render(truncateLines(body, boxH-1))
}

func (m Model) renderControls(width int) string {
	if text, urgent := m.manualControlsText(); text != "" {
		if urgent {
			return accentStyle.Render(shortenItemName(text, width))
		}
		return subtleStyle.Render(shortenItemName(text, width))
	}
	controlsText := "[Space] Пауза | [F] Побег | [M] Ручной | [T] Тактика | [E] Арсенал | [I] Кодекс | [S] Слава | [+/-] Скор. | [L] Язык | [Q] Выход"
	if m.Lang == LangEN {
		controlsText = "[Space] Pause | [F] Flee | [M] Manual | [T] Tactics | [E] Armory | [I] Codex | [S] Glory | [+/-] Speed | [L] Lang | [Q] Quit"
	}
	return subtleStyle.Render(shortenItemName(controlsText, width))
}
