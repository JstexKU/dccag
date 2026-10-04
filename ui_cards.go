package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var heroCardDowned = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("208")).
	Padding(0, 1)

func (m Model) renderHeroCard(h *Hero, cardWidth int, forceMode ...CardMode) string {
	innerWidth := max(14, cardWidth-4)
	textWidth := max(12, innerWidth-2)

	mode := detectCardMode(innerWidth)
	if len(forceMode) > 0 {
		mode = forceMode[0]
	}
	targetLines := cardContentHeight(mode)

	isActiveTurn := false
	if m.Combat != nil && len(m.Combat.TurnQueue) > 0 {
		idx := m.Combat.TurnIdx
		if idx >= len(m.Combat.TurnQueue) {
			idx = 0
		}
		if m.Combat.TurnQueue[idx].Type == CombatantHero && m.Combat.TurnQueue[idx].HeroRef == h {
			isActiveTurn = true
		}
	}

	var sb strings.Builder

	fullName := h.FullName(m.Lang)
	if h.IsLeader {
		fullName = "★ " + fullName
	}
	nameRaw := shortenItemName(fullName, textWidth)
	nameStr := nameRaw
	if h.TitleKey != "" {
		nameStr = titleStyle.Render(nameRaw)
	}
	sb.WriteString(padRightTruncate(nameStr, textWidth) + "\n")

	if h.IsDead {
		sb.WriteString(renderDeadHeroCard(h, textWidth, mode, m.Lang))
	} else if h.IsDowned {
		sb.WriteString(renderDownedHeroCard(h, textWidth, mode, m.Lang))
	} else {
		switch mode {
		case CardWide:
			sb.WriteString(renderWideHeroCard(h, textWidth, m))
		case CardMedium:
			sb.WriteString(renderMediumHeroCard(h, textWidth, m))
		default:
			sb.WriteString(renderCompactHeroCard(h, textWidth, m))
		}
	}

	normalizedBody := normalizeLines(sb.String(), textWidth, targetLines)

	baseStyle := heroCardStyle
	if h.IsDead {
		baseStyle = heroCardDead
	} else if h.IsDowned {
		baseStyle = heroCardDowned
	} else if isActiveTurn {
		baseStyle = heroCardActive
	}

	return baseStyle.
		Width(innerWidth).
		Height(targetLines).
		MaxHeight(targetLines + 2).
		Render(normalizedBody)
}

func renderDeadHeroCard(h *Hero, textWidth int, mode CardMode, lang Language) string {
	var sb strings.Builder

	infoLine := fmt.Sprintf("[%s | %s]", h.RaceName(lang), h.ShortClass(lang))
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, textWidth)) + "\n")
	sb.WriteString(dangerStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.dead")), textWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(h.CauseOfDeath, textWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.left_in_abyss_status")), textWidth)))

	return sb.String()
}

func renderDownedHeroCard(h *Hero, textWidth int, mode CardMode, lang Language) string {
	var sb strings.Builder

	infoLine := fmt.Sprintf("[%s | %s %d]", h.RaceName(lang), h.ShortClass(lang), h.Level)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, textWidth)) + "\n")
	sb.WriteString(goldStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.downed")), textWidth)) + "\n")
	sb.WriteString(dangerStyle.Render(padRightTruncate(fmt.Sprintf("HP: 0/%d", h.MaxHP), textWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.awaiting_evac")), textWidth)))

	return sb.String()
}

func renderWideHeroCard(h *Hero, textWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s | %s %d]", h.RaceName(m.Lang), h.ShortClass(m.Lang), h.Level), textWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, textWidth)) + "\n")

	barLen := textWidth - 15
	if barLen < 3 {
		barLen = 3
	}

	hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("HP%s%d/%d", hpBar, h.HP, h.MaxHP), textWidth) + "\n")

	mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP%s%d/%d", mpBar, h.MP, h.MaxMP), textWidth) + "\n")

	stressColor := lipgloss.Color("135")
	if h.Stress >= 140 {
		stressColor = lipgloss.Color("196")
	}
	stBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST%s%d/200", stBar, h.Stress), textWidth) + "\n")

	sb.WriteString(renderBeltAndStats(h, textWidth, m) + "\n")

	sb.WriteString(renderEquipLine("⚔", h.Weapon, textWidth, m.Lang))
	sb.WriteString(renderEquipLine("🛡", h.Chest, textWidth, m.Lang))
	sb.WriteString(renderEquipLine("🪖", h.Head, textWidth, m.Lang))
	sb.WriteString(renderEquipLine("🥾", h.Legs, textWidth, m.Lang))

	return sb.String()
}

func renderMediumHeroCard(h *Hero, textWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s | %s %d]", h.RaceName(m.Lang), h.ShortClass(m.Lang), h.Level), textWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, textWidth)) + "\n")

	barLen := textWidth - 9
	if barLen < 3 {
		barLen = 3
	}

	hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("HP%s%d", hpBar, h.HP), textWidth) + "\n")

	mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP%s%d", mpBar, h.MP), textWidth) + "\n")

	stressColor := lipgloss.Color("135")
	if h.Stress >= 140 {
		stressColor = lipgloss.Color("196")
	}
	stBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST%s%d", stBar, h.Stress), textWidth) + "\n")

	sb.WriteString(renderBeltAndStats(h, textWidth, m) + "\n")

	halfW := textWidth / 2
	rightW := textWidth - halfW

	wCell := formatEquipCell("⚔", h.Weapon, halfW, m.Lang)
	cCell := formatEquipCell("🛡", h.Chest, rightW, m.Lang)
	sb.WriteString(goldStyle.Render(wCell) + goldStyle.Render(cCell) + "\n")

	hCell := formatEquipCell("🪖", h.Head, halfW, m.Lang)
	lCell := formatEquipCell("🥾", h.Legs, rightW, m.Lang)
	sb.WriteString(goldStyle.Render(hCell) + goldStyle.Render(lCell))

	return sb.String()
}

func renderCompactHeroCard(h *Hero, textWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s %d]", h.ShortClass(m.Lang), h.Level), textWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, textWidth)) + "\n")

	sb.WriteString(padRightTruncate(fmt.Sprintf("HP %d/%d", h.HP, h.MaxHP), textWidth) + "\n")
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP %d/%d", h.MP, h.MaxMP), textWidth) + "\n")
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST %d/200", h.Stress), textWidth) + "\n")

	potSlot := beltSymbols(h, m.MaxPotionSlots())
	leftBlock := fmt.Sprintf("[%s]", potSlot)

	atkStr := fmt.Sprintf("%d", h.TotalAtk())
	defStr := fmt.Sprintf("%d", h.TotalDef())
	if len(atkStr) > 2 {
		atkStr = atkStr[:2] + "+"
	}
	if len(defStr) > 2 {
		defStr = defStr[:2] + "+"
	}
	statText := fmt.Sprintf("A:%s D:%s", atkStr, defStr)

	leftW := runewidth.StringWidth(leftBlock)
	statW := runewidth.StringWidth(statText)
	gapW := textWidth - leftW - statW
	if gapW < 1 {
		gapW = 1
	}
	sb.WriteString(leftBlock + strings.Repeat(" ", gapW) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(statText) + "\n")

	wName := "-"
	if h.Weapon != nil {
		wName = shortenItemName(T(m.Lang, h.Weapon.BaseNameKey), 6)
	}
	cName := "-"
	if h.Chest != nil {
		cName = shortenItemName(T(m.Lang, h.Chest.BaseNameKey), 6)
	}
	sb.WriteString(goldStyle.Render(padRightTruncate(fmt.Sprintf("⚔%s 🛡%s", wName, cName), textWidth)))

	return sb.String()
}

func beltSymbols(h *Hero, maxSlots int) string {
	var sb strings.Builder
	for i := 0; i < maxSlots; i++ {
		if i < len(h.Potions) && h.Potions[i] != nil {
			sb.WriteString(h.Potions[i].Symbol)
		} else {
			sb.WriteString("·")
		}
	}
	return sb.String()
}

func buffBadge(h *Hero) string {
	switch {
	case h.IsDowned:
		return goldStyle.Render("[💤]")
	case h.ReviveCooldown > 0:
		return subtleStyle.Render("[⏳]")
	case h.Affliction != AfflictionNone:
		return stressStyle.Render("[👁]")
	case h.IsGuarding:
		return healStyle.Render("[🛡]")
	case h.IsBerserk:
		return fireStyle.Render("[⚔]")
	case h.IsStealthed:
		return accentStyle.Render("[🗡]")
	case h.IsCharged:
		return accentStyle.Render("[🔮]")
	case h.IsAura:
		return fountStyle.Render("[✨]")
	default:
		return subtleStyle.Render("[-]")
	}
}

func renderBeltAndStats(h *Hero, innerWidth int, m Model) string {
	potSlot := beltSymbols(h, m.MaxPotionSlots())
	styledBuff := buffBadge(h)
	leftBlock := fmt.Sprintf("[%s] %s", potSlot, styledBuff)

	atkStr := fmt.Sprintf("%d", h.TotalAtk())
	defStr := fmt.Sprintf("%d", h.TotalDef())
	statText := fmt.Sprintf("⚔%s 🛡%s", atkStr, defStr)
	styledStats := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(statText)

	line := leftBlock + " " + styledStats
	return padRightTruncate(line, innerWidth)
}

func renderEquipLine(icon string, it *EquipItem, innerWidth int, lang Language) string {
	if it == nil {
		return subtleStyle.Render(padRightTruncate(icon+" -", innerWidth)) + "\n"
	}
	const iconW = 2
	statVal := fmt.Sprintf("%d", it.TotalStat())
	statW := runewidth.StringWidth(statVal)
	maxNameLen := innerWidth - iconW - statW - 2
	if maxNameLen < 3 {
		maxNameLen = 3
	}
	name := shortenItemName(T(lang, it.BaseNameKey), maxNameLen)

	line := goldStyle.Render(icon+" "+name) + " " + subtleStyle.Render(statVal)
	return padRightTruncate(line, innerWidth) + "\n"
}

func formatEquipCell(icon string, it *EquipItem, cellW int, lang Language) string {
	if it == nil {
		return padRightTruncate(icon+" -", cellW)
	}
	const iconW = 2
	maxNameLen := cellW - iconW - 1
	if maxNameLen < 2 {
		maxNameLen = 2
	}
	name := shortenItemName(T(lang, it.BaseNameKey), maxNameLen)
	return padRightTruncate(icon+" "+name, cellW)
}

func (m Model) renderPartyBanner(cardWidth int, forceMode ...CardMode) string {
	innerWidth := max(14, cardWidth-4)
	textWidth := max(12, innerWidth-2)

	mode := detectCardMode(innerWidth)
	if len(forceMode) > 0 {
		mode = forceMode[0]
	}
	targetLines := cardContentHeight(mode)

	relicName := T(m.Lang, "ui.none")
	relicDesc := T(m.Lang, "ui.no_relic")
	if m.Relic != nil {
		relicName = T(m.Lang, m.Relic.NameKey)
		relicDesc = T(m.Lang, m.Relic.DescKey)
	}
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)

	var sb strings.Builder

	titleText := "👑 " + T(m.Lang, "ui.party_and_relic")
	if mode == CardCompact {
		titleText = "👑 " + shortenItemName(T(m.Lang, "ui.party_and_relic"), textWidth-2)
	}
	sb.WriteString(goldStyle.Render(shortenItemName(titleText, textWidth)) + "\n")

	switch mode {
	case CardWide:
		sb.WriteString(shortenItemName(fmt.Sprintf("%s: %s", T(m.Lang, "ui.relic_short"), relicName), textWidth) + "\n")
		sb.WriteString(subtleStyle.Render(shortenItemName(relicDesc, textWidth)) + "\n")
		sb.WriteString("\n")
		sb.WriteString(shortenItemName(fmt.Sprintf("🎒 %s: %d/%d %s", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")), textWidth) + "\n")
		sb.WriteString(healStyle.Render(shortenItemName(fmt.Sprintf("🏛️ +%dG (%s)", legacyPart, T(m.Lang, "ui.treasury")), textWidth)) + "\n")
		sb.WriteString(subtleStyle.Render(shortenItemName(T(m.Lang, "town.tax_active"), textWidth)) + "\n")
		sb.WriteString("\n\n")

	case CardMedium:
		sb.WriteString(shortenItemName(fmt.Sprintf("%s: %s", T(m.Lang, "ui.relic_short"), relicName), textWidth) + "\n")
		sb.WriteString(subtleStyle.Render(shortenItemName(relicDesc, textWidth)) + "\n")
		sb.WriteString("\n")
		sb.WriteString(shortenItemName(fmt.Sprintf("🎒 %s: %d/%d", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity()), textWidth) + "\n")
		sb.WriteString(healStyle.Render(shortenItemName(fmt.Sprintf("🏛️ +%dG", legacyPart), textWidth)) + "\n")
		sb.WriteString(subtleStyle.Render(shortenItemName(T(m.Lang, "town.tax_active"), textWidth)) + "\n")
		sb.WriteString("\n")

	default: // CardCompact
		sb.WriteString(shortenItemName(relicName, textWidth) + "\n")
		sb.WriteString(shortenItemName(fmt.Sprintf("🎒 %d/%d", len(m.Bag), m.currentBagCapacity()), textWidth) + "\n")
		sb.WriteString(healStyle.Render(shortenItemName(fmt.Sprintf("🏛️ +%dG", legacyPart), textWidth)) + "\n")
		sb.WriteString(subtleStyle.Render(shortenItemName(T(m.Lang, "town.tax_active"), textWidth)) + "\n")
		sb.WriteString("\n")
	}

	normalizedBody := normalizeLines(sb.String(), textWidth, targetLines)

	return heroCardStyle.
		Width(innerWidth).
		Height(targetLines).
		MaxHeight(targetLines + 2).
		Render(normalizedBody)
}
