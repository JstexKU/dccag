package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderBar(current, max int, totalBars int, filledColor, emptyColor lipgloss.Color) string {
	if totalBars < 2 {
		totalBars = 2
	}
	if max <= 0 {
		max = 1
	}
	ratio := float64(current) / float64(max)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(totalBars))
	empty := totalBars - filled
	fStr := lipgloss.NewStyle().Foreground(filledColor).Render(strings.Repeat("█", filled))
	eStr := lipgloss.NewStyle().Foreground(emptyColor).Render(strings.Repeat("░", empty))
	return "[" + fStr + eStr + "]"
}

func shortenItemName(name string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(name)
	if len(runes) > maxLen {
		if maxLen <= 1 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-1]) + "…"
	}
	return name
}

func padRight(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

func getBiome(floor int) BiomeConfig {
	cycle := (floor - 1) % 5
	switch cycle {
	case 0:
		return BiomeConfig{Name: BiomeCatacombs, WallColor: lipgloss.Color("240"), FloorColor: lipgloss.Color("236"), FloorRune: '·', EnvHazardKey: "hazard.catacombs"}
	case 1:
		return BiomeConfig{Name: BiomeGrotto, WallColor: lipgloss.Color("31"), FloorColor: lipgloss.Color("24"), FloorRune: '~', EnvHazardKey: "hazard.grotto"}
	case 2:
		return BiomeConfig{Name: BiomeInferno, WallColor: lipgloss.Color("124"), FloorColor: lipgloss.Color("52"), FloorRune: '≈', EnvHazardKey: "hazard.inferno"}
	case 3:
		return BiomeConfig{Name: BiomeCrystal, WallColor: lipgloss.Color("141"), FloorColor: lipgloss.Color("54"), FloorRune: '◊', EnvHazardKey: "hazard.crystal"}
	default:
		return BiomeConfig{Name: BiomeAbyss, WallColor: lipgloss.Color("89"), FloorColor: lipgloss.Color("233"), FloorRune: '×', EnvHazardKey: "hazard.abyss"}
	}
}

type BiomeConfig struct {
	Name         BiomeType
	WallColor    lipgloss.Color
	FloorColor   lipgloss.Color
	FloorRune    rune
	EnvHazardKey string
}

func getAffixIcon(a MonsterAffix) string {
	switch a {
	case AffixFire:
		return "🔥"
	case AffixPoison:
		return "☣️"
	case AffixFrost:
		return "❄️"
	case AffixStone:
		return "🪨"
	case AffixVampiric:
		return "🩸"
	default:
		return ""
	}
}

func (m Model) renderMenuScreen() string {
	var sb strings.Builder

	banner := townArtStyle.Render(`
           / \                                                 / \
          /   \                  |>>>                         /   \
         /_____\                 |                           /_____\
        |  .-.  |            _  _|_  _                      |  .-.  |
        |  | |  |           |;|_|;|_|;|                     |  | |  |
        |  '-'  |           \\.    .  /                     |  '-'  |
        |       |            \\:  .  /                      |       |
      ,-'-------'-,           ||:   |                     ,-'-------'-,
    ,'  /═══════\  '.         ||:.  |                   ,'  /═══════\  '.
   /   /         \   \        ||:  .|                  /   /         \   \
  |   |           |   |       ||:   | ____            |   |           |   |
  |   |           |   |       ||: , !_|__|            |   |           |   |
  |===|===========|===|   ____||_ | |    |            |===|===========|===|
  |   |  D C C AG |   |  |___|__|_|_|_   |  ____      |   |  D C C AG |   |
  |   |           |   |      |        |  | |____|     |   |           |   |
  |___|___________|___|      |________|__|_|    |     |___|___________|___|
    [═══════════════]           /════════\  |___|       [═══════════════]
`)

	sb.WriteString(banner + "\n")
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) v2.3.0")) + "\n\n")

	var textBlock string
	if m.Lang == LangEN {
		textBlock = "Welcome to the grim tactical dungeon crawler!\n" +
			"Guide your autonomous squad through INFINITE floors and ruthless biomes.\n\n" +
			questStyle.Render("SYSTEM FEATURES:") + "\n" +
			" • Bilingual Engine (RU / EN) switched on the fly with [L].\n" +
			" • Dynamic establishments: «" + T(m.Lang, m.TownEst.TavernKey) + "», «" + T(m.Lang, m.TownEst.SmithyKey) + "».\n" +
			" • Distinct Artisan materials: wood, silk, rawhide, and tempered steel.\n" +
			" • Multi-tabbed Codex [I]: fast navigation with [Tab] and [1-4].\n\n" +
			dangerStyle.Render(fmt.Sprintf("⏳ Expedition autostarts in: %d sec...\n\n", m.MenuCountdown)) +
			healStyle.Render("[ENTER] or [SPACE] — Embark immediately") + "\n" +
			accentStyle.Render("[L] — Switch language (RU / EN)") + "\n" +
			subtleStyle.Render("[Q] — Quit")
	} else {
		textBlock = "Добро пожаловать в мрачный тактический подземельный рогалик!\n" +
			"Ваша задача — провести отряд сквозь БЕСКОНЕЧНЫЕ этажи опаснейших биомов.\n\n" +
			questStyle.Render("ОСОБЕННОСТИ СИСТЕМЫ:") + "\n" +
			" • Двуязычный движок (RU / EN) с переключением на лету клавишей [L].\n" +
			" • Колоритные заведения: трактир «" + T(m.Lang, m.TownEst.TavernKey) + "», кузня «" + T(m.Lang, m.TownEst.SmithyKey) + "».\n" +
			" • Разделение ремёсел: посохи из тиса, робы из шёлка, латы из стали.\n" +
			" • Вкладки Кодекса [I]: быстрое переключение разделов по [Tab] и [1-4].\n\n" +
			dangerStyle.Render(fmt.Sprintf("⏳ Автоматический старт экспедиции через: %d сек...\n\n", m.MenuCountdown)) +
			healStyle.Render("[ENTER] или [SPACE] — Начать экспедицию немедленно") + "\n" +
			accentStyle.Render("[L] — Сменить язык (RU / EN)") + "\n" +
			subtleStyle.Render("[Q] — Выход из игры")
	}

	// Выравниваем текстовый блок по левому краю внутри центрированного бокса
	alignedText := lipgloss.NewStyle().Align(lipgloss.Left).Render(textBlock)
	sb.WriteString(alignedText)

	box := menuBoxStyle.Render(sb.String())
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.legacy_header") + ":\n"))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", T(m.Lang, "stats.legacy_treasury"), goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
	sb.WriteString(fmt.Sprintf(" • %s («%s»): Ур.%d | %s («%s»): Ур.%d\n",
		T(m.Lang, "town.smithy"), T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel,
		T(m.Lang, "town.tannery"), T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel))
	sb.WriteString(fmt.Sprintf(" • %s («%s»): Ур.%d | %s («%s»): Ур.%d\n\n",
		T(m.Lang, "town.church"), T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel,
		T(m.Lang, "town.tavern"), T(m.Lang, m.TownEst.TavernKey), m.Legacy.TavernLevel))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.achievements_header") + ":\n"))
	sb.WriteString(fmt.Sprintf(" • %s: %d | %s: %d | %s: %s\n",
		T(m.Lang, "stats.floors_cleared"), m.Stats.FloorsCleared,
		T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps,
		T(m.Lang, "stats.gold_earned"), goldStyle.Render(fmt.Sprintf("%dG", m.Stats.TotalGoldEarned))))
	sb.WriteString(fmt.Sprintf(" • %s: %d | %s: %d | %s: %d\n\n",
		T(m.Lang, "stats.contracts_closed"), m.Stats.QuestsCompleted,
		T(m.Lang, "stats.upgrades_forged"), m.Stats.UpgradesForged,
		T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.survivors_header") + ":\n"))
	for _, h := range m.Party {
		if !h.IsDead {
			heroName := h.FullName(m.Lang)
			if h.TitleKey != "" {
				heroName = titleStyle.Render(heroName)
			}
			sb.WriteString(fmt.Sprintf(" • %-20s (%-5s %2d) [%s] (Atk:%2d | Def:%2d | %s:%2d)\n",
				heroName, h.ShortClass(m.Lang), h.Level, healStyle.Render(T(m.Lang, "ui.alive")),
				h.TotalAtk(), h.TotalDef(), T(m.Lang, "ui.speed_short"), h.TotalSpeed()))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s (%d):\n", T(m.Lang, "stats.fallen_heroes"), len(m.Stats.FallenHeroes))))
	if len(m.Stats.FallenHeroes) == 0 {
		sb.WriteString(subtleStyle.Render(" " + T(m.Lang, "stats.no_fallen") + "\n"))
	} else {
		for _, f := range m.Stats.FallenHeroes {
			clsStr := TranslateEnum(m.Lang, "class", string(f.Class)+".short")
			sb.WriteString(fmt.Sprintf(" ☠️ %-18s (%-5s) | %s: %d | %s\n",
				dangerStyle.Render(f.FullName), clsStr, T(m.Lang, "ui.floor"), f.Floor, subtleStyle.Render(f.Cause)))
		}
	}

	if m.State == StateDefeat {
		countdownStr := dangerStyle.Render(fmt.Sprintf(T(m.Lang, "defeat.restart_timer"), m.RestartCountdown))
		sb.WriteString("\n" + countdownStr + "\n")
	}

	sb.WriteString(subtleStyle.Render("\n[↑/↓/PgUp/PgDn] " + T(m.Lang, "ui.scroll") + " | [R] " + T(m.Lang, "ui.btn_restart") + " | [S] " + T(m.Lang, "ui.btn_back") + " | [L] Lang | [Q] " + T(m.Lang, "ui.quit")))

	fullText := sb.String()
	lines := strings.Split(fullText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderInfoBookScreen() string {
	boxW := m.TermWidth - 8
	if boxW > 108 {
		boxW = 108
	}
	if boxW < 60 {
		boxW = 60
	}
	innerW := boxW - 6

	tabNames := []string{
		T(m.Lang, "codex.tab.classes"),
		T(m.Lang, "codex.tab.combat"),
		T(m.Lang, "codex.tab.town"),
		T(m.Lang, "codex.tab.relics"),
	}

	var tabHeaders []string
	for i, name := range tabNames {
		if m.CodexTab == i {
			tabHeaders = append(tabHeaders, lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render("►["+name+"]◄"))
		} else {
			tabHeaders = append(tabHeaders, subtleStyle.Render(" ["+name+"] "))
		}
	}
	tabsRow := strings.Join(tabHeaders, " ")

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render(T(m.Lang, "codex.header") + "\n\n"))
	sb.WriteString(tabsRow + "\n")
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	switch m.CodexTab {
	case 0:
		sb.WriteString(T(m.Lang, "codex.content.classes") + "\n\n")
	case 1:
		sb.WriteString(T(m.Lang, "codex.content.combat") + "\n\n")
	case 2:
		sb.WriteString(T(m.Lang, "codex.content.town") + "\n\n")
	case 3:
		sb.WriteString(T(m.Lang, "codex.content.relics") + "\n\n")
	}

	sb.WriteString(subtleStyle.Render(fmt.Sprintf("[%s / %s] %s | [↑/↓/PgUp/PgDn] %s | [L] Lang | [I/Esc] %s",
		accentStyle.Render("Tab"), accentStyle.Render("1-4"), T(m.Lang, "codex.switch_tabs"),
		T(m.Lang, "ui.scroll"), T(m.Lang, "ui.btn_back"))))

	wrappedText := lipgloss.NewStyle().Width(innerW).Render(sb.String())
	lines := strings.Split(wrappedText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Width(boxW).Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderArmoryScreen() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(T(m.Lang, "armory.title") + " [E]")
	sb.WriteString(header + "\n\n")

	renderSlotInfo := func(slotName string, it *EquipItem) string {
		if it == nil {
			return fmt.Sprintf("   • %-7s: %s\n", slotName, subtleStyle.Render("-"))
		}
		matStr := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("[%s, x%d]", T(m.Lang, it.Material.Key), it.Material.BonusMult))

		modStr := ""
		if it.Prefix != nil {
			modStr += fmt.Sprintf(" | %s (+%d)", T(m.Lang, it.Prefix.Key), it.Prefix.Bonus)
		}
		if it.Suffix != nil {
			modStr += fmt.Sprintf(" | %s (+%d)", T(m.Lang, it.Suffix.Key), it.Suffix.Bonus)
		}
		if it.SpeedBonus > 0 {
			modStr += fmt.Sprintf(" | +%d %s", it.SpeedBonus, T(m.Lang, "ui.speed_short"))
		}

		statLabel := T(m.Lang, "ui.def_short")
		if it.Slot == SlotWeapon {
			statLabel = T(m.Lang, "ui.atk_short")
		}

		upg := ""
		if it.UpgradeLevel > 0 {
			upg = fmt.Sprintf(" +%d", it.UpgradeLevel)
		}

		return fmt.Sprintf("   • %-7s: %s%s %s -> %s: %s %d%s\n",
			slotName, goldStyle.Render(T(m.Lang, it.BaseNameKey)), upg, matStr,
			T(m.Lang, "ui.total"), statLabel, it.TotalStat(), subtleStyle.Render(modStr))
	}

	for _, h := range m.Party {
		status := healStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.alive")))
		if h.IsDead {
			status = dangerStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.dead")))
		}

		heroTitle := h.FullName(m.Lang)
		if h.TitleKey != "" {
			heroTitle = titleStyle.Render(heroTitle)
		}

		mutStr := subtleStyle.Render(T(m.Lang, "ui.none"))
		if h.Mutations.Total() > 0 {
			mutStr = accentStyle.Render(fmt.Sprintf("%s: %d [Ch:%d Fu:%d Ti:%d Ae:%d Ba:%d]",
				T(m.Lang, "ui.total"), h.Mutations.Total(), h.Mutations.ChimeraCount, h.Mutations.FuryCount,
				h.Mutations.TitanCount, h.Mutations.AetherCount, h.Mutations.BastionCount))
		}

		potInfo := T(m.Lang, "ui.none")
		if len(h.Potions) > 0 {
			var pNames []string
			for _, p := range h.Potions {
				if p != nil {
					pNames = append(pNames, fmt.Sprintf("%s %s", p.Symbol, T(m.Lang, fmt.Sprintf("potion.%s.%s", p.Size, p.Type))))
				}
			}
			potInfo = strings.Join(pNames, ", ")
		}

		sb.WriteString(fmt.Sprintf("👤 %s (%s %d, XP:%d/%d, Aggro:%d, %s:%d) %s\n",
			heroTitle, h.ShortClass(m.Lang), h.Level, h.Exp, h.NextLevelExp(), h.Role.AggroWeight,
			T(m.Lang, "ui.speed_short"), h.TotalSpeed(), status))
		sb.WriteString(fmt.Sprintf("   🧪 %s: %s\n", T(m.Lang, "ui.mutations"), mutStr))
		sb.WriteString(fmt.Sprintf("   🎒 %s: %s (%s: %d/%d)\n",
			T(m.Lang, "ui.potions_belt"), potionStyle.Render(potInfo), T(m.Lang, "ui.slots"), len(h.Potions), m.MaxPotionSlots()))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.weapon"), h.Weapon))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.head"), h.Head))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.chest"), h.Chest))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.legs"), h.Legs))
		sb.WriteString("\n")
	}

	sb.WriteString(subtleStyle.Render("[↑/↓/PgUp/PgDn] " + T(m.Lang, "ui.scroll") + " | [E/Esc] " + T(m.Lang, "ui.btn_back") + " | [L] Lang | [Space] Pause | [Q] " + T(m.Lang, "ui.quit")))

	fullText := sb.String()
	lines := strings.Split(fullText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderTownHub(viewW, viewH int) string {
	cActive := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	cIdle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	cBorder := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	cHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	fmtBld := func(phase TownPhase, icon, name string, lvl int) string {
		isActive := m.TownPhase == phase
		lvlStr := ""
		if lvl > 0 {
			lvlStr = fmt.Sprintf(" (%s.%d)", T(m.Lang, "ui.level_short"), lvl)
		}

		label := fmt.Sprintf("%s %s%s", icon, name, lvlStr)
		if isActive {
			return cActive.Render(fmt.Sprintf("►[%s]◄", label))
		}
		return cIdle.Render(fmt.Sprintf(" [%s] ", label))
	}

	bMarket := fmtBld(TownPhaseSellLoot, "⚖️", T(m.Lang, "town.market"), 0)
	bMagistrate := fmtBld(TownPhaseMagistrate, "🏛️", T(m.Lang, "town.magistrate"), 0)
	bChurch := fmtBld(TownPhaseChurch, "⛪", T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel)
	bTavern := fmtBld(TownPhaseTavern, "🍻", T(m.Lang, m.TownEst.TavernKey), m.Legacy.TavernLevel)
	bGuild := fmtBld(TownPhaseGuild, "⚔️", T(m.Lang, m.TownEst.GuildKey), 0)
	bSmithy := fmtBld(TownPhaseSmithy, "⚒️", T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel)
	bTannery := fmtBld(TownPhaseTannery, "🎒", T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel)
	bAlchemist := fmtBld(TownPhaseAlchemist, "🧪", T(m.Lang, m.TownEst.AlchemistKey), 0)

	maxW := max(36, viewW-2)
	center := func(s string) string {
		w := lipgloss.Width(s)
		if w >= maxW {
			return s
		}
		pad := (maxW - w) / 2
		return strings.Repeat(" ", pad) + s
	}

	var sb strings.Builder
	sb.WriteString(center(cHeader.Render(T(m.Lang, "town.hub_title"))) + "\n\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s", bMarket, bMagistrate)) + "\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s  %s", bChurch, bTavern, bGuild)) + "\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s  %s", bSmithy, bTannery, bAlchemist)) + "\n\n")

	sb.WriteString(cBorder.Render(strings.Repeat("─", maxW)) + "\n")
	sb.WriteString(questStyle.Render(" " + T(m.Lang, "town.log_title") + ":\n"))

	if len(m.TownHistory) == 0 {
		sb.WriteString(subtleStyle.Render("   • " + T(m.Lang, "town.log.enter_gate") + "\n"))
	} else {
		for _, act := range m.TownHistory {
			sb.WriteString(fmt.Sprintf("   • %s\n", shortenItemName(act, maxW-6)))
		}
	}

	lines := strings.Split(sb.String(), "\n")
	if len(lines) > viewH {
		lines = lines[:viewH]
	}
	for len(lines) < viewH {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderMap(viewW, viewH int) string {
	if m.InTown {
		return m.renderTownHub(viewW, viewH)
	}

	biome := getBiome(m.Floor)
	customWall := lipgloss.NewStyle().Foreground(biome.WallColor).Render("#")
	customFloor := lipgloss.NewStyle().Foreground(biome.FloorColor).Render(string(biome.FloorRune))

	startX := m.PartyPos.X - viewW/2
	startY := m.PartyPos.Y - viewH/2

	if startX+viewW > m.MapWidth {
		startX = m.MapWidth - viewW
	}
	if startY+viewH > m.MapHeight {
		startY = m.MapHeight - viewH
	}
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	var sb strings.Builder
	for y := 0; y < viewH; y++ {
		mapY := startY + y
		for x := 0; x < viewW; x++ {
			mapX := startX + x
			if mapX >= m.MapWidth || mapY >= m.MapHeight {
				sb.WriteString(" ")
				continue
			}

			if mapX == m.PartyPos.X && mapY == m.PartyPos.Y {
				sb.WriteString(partyStyle.Render("@"))
				continue
			}
			if !m.Explored[mapY][mapX] {
				sb.WriteString(" ")
				continue
			}
			pos := Point{mapX, mapY}
			if pack, exists := m.Packs[pos]; exists {
				first := pack.GetFirstLiving()
				if first != nil {
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(first.Color)).Bold(true).Render(string(first.Glyph)))
				} else {
					sb.WriteString(customFloor)
				}
				continue
			}
			switch m.Grid[mapY][mapX] {
			case TileWall:
				sb.WriteString(customWall)
			case TileFloor:
				sb.WriteString(customFloor)
			case TileChest:
				sb.WriteString(goldStyle.Render("$"))
			case TileRelic:
				sb.WriteString(relicTileStyle.Render("*"))
			case TileAltar:
				sb.WriteString(altarStyle.Render("_"))
			case TileFountain:
				sb.WriteString(fountStyle.Render("~"))
			case TileTrappedChest:
				sb.WriteString(trappedChestStyle.Render("T"))
			case TileBarrel:
				sb.WriteString(barrelStyle.Render("o"))
			case TileStairs:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render(">"))
			case TileExit:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render("<"))
			default:
				sb.WriteRune(rune(m.Grid[mapY][mapX]))
			}
		}
		if y < viewH-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) renderHeroCard(h *Hero, cardWidth int) string {
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
	innerWidth := max(18, cardWidth-4)

	nameRaw := h.FullName(m.Lang)
	nameRunes := []rune(nameRaw)
	if len(nameRunes) > innerWidth {
		nameRaw = string(nameRunes[:innerWidth-1]) + "…"
	}
	nameStr := nameRaw
	if h.TitleKey != "" {
		nameStr = titleStyle.Render(nameRaw)
	}

	classFormatted := subtleStyle.Render(fmt.Sprintf("(%s %d)", h.ShortClass(m.Lang), h.Level))

	potSlot := ""
	maxSlots := m.MaxPotionSlots()
	for i := 0; i < maxSlots; i++ {
		if i < len(h.Potions) && h.Potions[i] != nil {
			potSlot += h.Potions[i].Symbol
		} else {
			potSlot += "·"
		}
	}
	potSlotRendered := fmt.Sprintf("[%s]", potSlot)

	buffSlot := subtleStyle.Render("[-]")
	switch {
	case h.Affliction != AfflictionNone:
		buffSlot = stressStyle.Render("[👁]")
	case h.IsGuarding:
		buffSlot = healStyle.Render("[🛡]")
	case h.IsBerserk:
		buffSlot = fireStyle.Render("[⚔ ]")
	case h.IsStealthed:
		buffSlot = accentStyle.Render("[🗡]")
	case h.IsCharged:
		buffSlot = accentStyle.Render("[🔮]")
	case h.IsAura:
		buffSlot = fountStyle.Render("[✨]")
	}

	turnSlot := subtleStyle.Render("[ ]")
	if isActiveTurn {
		turnSlot = accentStyle.Render("[⚡]")
	}

	if h.IsDead {
		sb.WriteString(fmt.Sprintf("%s\n", nameStr))
		sb.WriteString(fmt.Sprintf("%s %s\n", classFormatted, dangerStyle.Render(fmt.Sprintf("[☠️ %s]", T(m.Lang, "ui.dead")))))
		sb.WriteString(subtleStyle.Render(shortenItemName(h.CauseOfDeath, innerWidth)) + "\n")
		sb.WriteString(fmt.Sprintf("%s: %d\n", T(m.Lang, "ui.floor"), m.Floor))
		sb.WriteString(subtleStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.awaiting_revive"))))
	} else {
		sb.WriteString(fmt.Sprintf("%s\n", nameStr))
		sb.WriteString(fmt.Sprintf("%s %s %s %s\n", classFormatted, potSlotRendered, buffSlot, turnSlot))

		barLen := max(2, (innerWidth-18)/2)
		hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
		mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("HP:%s%2d MP:%s%2d\n", hpBar, h.HP, mpBar, h.MP))

		stressColor := lipgloss.Color("135")
		if h.Stress >= 140 {
			stressColor = lipgloss.Color("196")
		}
		stressBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("ST:%s%3d ⚔%-2d 🛡%-2d\n", stressBar, h.Stress, h.TotalAtk(), h.TotalDef()))

		slotLen := max(4, (innerWidth-6)/2)
		formatEquipSlot := func(item *EquipItem, maxLen int) string {
			if item == nil {
				return "-"
			}
			name := shortenItemName(T(m.Lang, item.BaseNameKey), maxLen)
			if item.UpgradeLevel > 0 {
				return fmt.Sprintf("+%d%s", item.UpgradeLevel, shortenItemName(T(m.Lang, item.BaseNameKey), maxLen-2))
			}
			return name
		}

		wName := padRight(formatEquipSlot(h.Weapon, slotLen), slotLen)
		hName := padRight(formatEquipSlot(h.Head, slotLen), slotLen)
		chName := padRight(formatEquipSlot(h.Chest, slotLen), slotLen)
		lName := padRight(formatEquipSlot(h.Legs, slotLen), slotLen)

		equipBlock := fmt.Sprintf("⚔%s 🪖%s\n🛡%s 🥾%s", wName, hName, chName, lName)
		sb.WriteString(subtleStyle.Render(equipBlock))
	}

	style := heroCardStyle.Width(innerWidth)
	if h.IsDead {
		style = heroCardDead.Width(innerWidth)
	} else if isActiveTurn {
		style = heroCardActive.Width(innerWidth)
	}
	return style.Render(sb.String())
}

func (m Model) renderPartyBanner(cardWidth int) string {
	innerWidth := max(18, cardWidth-4)
	var sb strings.Builder
	sb.WriteString(goldStyle.Render("👑 "+T(m.Lang, "ui.party_and_relic")) + "\n")
	relicName := T(m.Lang, "ui.none")
	relicDesc := T(m.Lang, "ui.no_relic")
	if m.Relic != nil {
		relicName = T(m.Lang, m.Relic.NameKey)
		relicDesc = T(m.Lang, m.Relic.DescKey)
	}
	sb.WriteString(fmt.Sprintf("%s: %s\n", T(m.Lang, "ui.relic_short"), shortenItemName(relicName, innerWidth-10)))
	sb.WriteString(subtleStyle.Render(shortenItemName(relicDesc, innerWidth)) + "\n")
	sb.WriteString(fmt.Sprintf("🎒 %s: %d/%d %s\n", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")))
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)
	sb.WriteString(healStyle.Render(fmt.Sprintf("🏛️ %s: +%dG", T(m.Lang, "stats.legacy_treasury"), legacyPart)))

	return lipgloss.NewStyle().
		Width(innerWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Render(sb.String())
}

func (m Model) renderRightContentLines(innerRightW, maxLines int) []string {
	var lines []string
	if m.Combat != nil {
		barrelNotice := ""
		if m.Combat.HasBarrel {
			barrelNotice = " [🛢️ " + T(m.Lang, "combat.barrel") + "]"
		}
		if m.Combat.Round > 20 {
			barrelNotice += dangerStyle.Render(fmt.Sprintf(" [%s R%d]", T(m.Lang, "status.rage"), m.Combat.Round))
		}
		lines = append(lines, dangerStyle.Render(shortenItemName(T(m.Lang, "combat.enemy_pack")+barrelNotice+":", innerRightW-2)))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))

		availRows := max(1, maxLines-len(lines))
		displayed := 0
		for _, mob := range m.Combat.Pack.Members {
			if displayed >= availRows {
				break
			}

			statusWidth := 7
			stText := fmt.Sprintf("%d/%d", mob.HP, mob.MaxHP)
			stColor := lipgloss.Color("245")
			if mob.IsDead {
				stText = T(m.Lang, "ui.dead")
				stColor = lipgloss.Color("239")
			}
			stBadge := lipgloss.NewStyle().Foreground(stColor).Render(padRight(fmt.Sprintf("[%s]", stText), statusWidth))

			barLen := 3
			if innerRightW > 36 {
				barLen = 4
			}
			mobBar := renderBar(mob.HP, mob.MaxHP, barLen, lipgloss.Color("196"), lipgloss.Color("238"))

			affIcon := getAffixIcon(mob.Affix)
			if affIcon != "" {
				affIcon += " "
			}

			prefix := fmt.Sprintf("• %s", affIcon)
			fixedW := lipgloss.Width(prefix) + 1 + lipgloss.Width(mobBar) + 1 + lipgloss.Width(stBadge)
			availName := max(3, innerRightW-fixedW)
			mobName := shortenItemName(T(m.Lang, mob.NameKey), availName)
			namePadded := padRight(mobName, availName)

			lines = append(lines, fmt.Sprintf("%s%s %s %s", prefix, namePadded, mobBar, stBadge))
			displayed++
		}

		if len(m.Combat.Pack.Members) > displayed && len(lines) < maxLines {
			lines = append(lines, subtleStyle.Render(fmt.Sprintf("... +%d", len(m.Combat.Pack.Members)-displayed)))
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render(T(m.Lang, "town.management")+":"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("%s %s.%-2d   | %s %s.%-2d",
			T(m.Lang, "town.smithy"), T(m.Lang, "ui.level_short"), m.Legacy.SmithyLevel,
			T(m.Lang, "town.tannery"), T(m.Lang, "ui.level_short"), m.Legacy.TanneryLevel))
		lines = append(lines, fmt.Sprintf("%s  %s.%-2d   | %s  %s.%-2d",
			T(m.Lang, "town.church"), T(m.Lang, "ui.level_short"), m.Legacy.ChurchLevel,
			T(m.Lang, "town.tavern"), T(m.Lang, "ui.level_short"), m.Legacy.TavernLevel))
		lines = append(lines, fmt.Sprintf("%s: %d  | %s: %d",
			T(m.Lang, "ui.potions_belt"), m.MaxPotionSlots(), T(m.Lang, "ui.bag"), m.currentBagCapacity()))
		lines = append(lines, subtleStyle.Render(shortenItemName(T(m.Lang, "town.tax_active"), innerRightW-2)))
	} else {
		lines = append(lines, accentStyle.Render(T(m.Lang, "ui.scouting")+":"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("%s:   %d", T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps))
		lines = append(lines, fmt.Sprintf("%s: %d", T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened))
	}
	return lines
}

func (m Model) renderRightContent(innerRightW, infoInnerH int) string {
	lines := m.renderRightContentLines(innerRightW, infoInnerH)
	for len(lines) < infoInnerH {
		lines = append(lines, "")
	}
	if len(lines) > infoInnerH {
		lines = lines[:infoInnerH]
	}
	return strings.Join(lines, "\n")
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

	termW := max(70, m.TermWidth)
	termH := max(20, m.TermHeight)
	usableW := termW - 2

	cardsPerRow := 3
	if usableW >= 150 || (termH < 34 && usableW >= 115) {
		cardsPerRow = 6
	} else if usableW < 90 {
		cardsPerRow = 2
	}

	singleCardWidth := (usableW - (cardsPerRow-1)*1) / cardsPerRow
	numRows := (6 + cardsPerRow - 1) / cardsPerRow

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, singleCardWidth))
	}
	cards = append(cards, m.renderPartyBanner(singleCardWidth))

	var cardRowElements []string
	for i := 0; i < len(cards); i += cardsPerRow {
		end := min(len(cards), i+cardsPerRow)
		cardRowElements = append(cardRowElements, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	middleTier := lipgloss.JoinVertical(lipgloss.Left, cardRowElements...)

	cardsHeight := numRows * 8
	maxLogs := 5
	if termH < 32 {
		maxLogs = 3
	}
	if termH < 26 {
		maxLogs = 2
	}
	fixedBottomHeight := maxLogs + 4

	availableTopH := max(6, termH-cardsHeight-fixedBottomHeight)

	sideW := int(float64(usableW) * 0.34)
	if sideW < 28 {
		sideW = 28
	}
	if sideW > 46 {
		sideW = 46
	}
	mapBoxW := max(32, usableW-sideW-1)

	viewW := max(10, mapBoxW-4)
	viewH := max(4, availableTopH-2)

	mapStr := m.renderMap(viewW, viewH)
	leftMapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(mapBoxW - 2).
		Height(viewH).
		Render(mapStr)

	innerRightW := max(22, sideW-4)
	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("%s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("%s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}

	questTitleShort := shortenItemName(T(m.Lang, m.CurrentQuest.TitleKey), innerRightW-12)
	var statusBadge string
	if m.CurrentQuest.Completed {
		statusBadge = healStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.turn_in")))
	} else {
		statusBadge = subtleStyle.Render(fmt.Sprintf("[%d/%d]", m.CurrentQuest.Current, m.CurrentQuest.TargetCount))
	}
	questLine := padRight(fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge), innerRightW-4)

	var rightPane string
	if availableTopH >= 13 {
		floorBoxContent := fmt.Sprintf(
			"🏰 %s\n%s",
			titleStyle.Render(shortenItemName(floorTag, innerRightW-4)),
			subtleStyle.Render(shortenItemName(T(m.Lang, "ui.hazard")+": "+T(m.Lang, biome.EnvHazardKey), innerRightW-4)),
		)
		floorBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Width(innerRightW).
			Height(2).
			Render(floorBoxContent)

		statusContent := fmt.Sprintf(
			"💰 %s: %s\n%s",
			T(m.Lang, "ui.treasury"),
			goldStyle.Render(fmt.Sprintf("%dG", m.Gold)),
			questLine,
		)
		statusBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Width(innerRightW).
			Height(2).
			Render(statusContent)

		infoInnerH := max(3, availableTopH-10)
		rightInfoBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(innerRightW).
			Height(infoInnerH).
			Render(m.renderRightContent(innerRightW, infoInnerH))

		rightPane = lipgloss.JoinVertical(lipgloss.Left, rightInfoBox, statusBox, floorBox)
	} else {
		compactH := max(4, availableTopH-2)
		var lines []string
		lines = append(lines, fmt.Sprintf("🏰 %s | 💰 %s",
			titleStyle.Render(shortenItemName(floorTag, 14)),
			goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
		lines = append(lines, fmt.Sprintf("📜 %s", questLine))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, m.renderRightContentLines(innerRightW, compactH-3)...)

		rightPane = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(innerRightW).
			Height(compactH).
			Render(strings.Join(lines, "\n"))
	}

	topTier := lipgloss.JoinHorizontal(lipgloss.Top, leftMapBox, " ", rightPane)

	var logs strings.Builder
	scrollInfo := ""
	if m.LogScroll > 0 {
		scrollInfo = fmt.Sprintf(" (%s: -%d)", T(m.Lang, "ui.archive"), m.LogScroll)
	}
	logs.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "ui.chronicles") + scrollInfo + ":\n"))

	totalLogs := len(m.Logs)
	endIdx := totalLogs - m.LogScroll
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	if endIdx < maxLogs {
		endIdx = min(totalLogs, maxLogs)
	}
	startIdx := max(0, endIdx-maxLogs)

	for i := 0; i < maxLogs; i++ {
		curIdx := startIdx + i
		if curIdx < endIdx && curIdx < totalLogs {
			logs.WriteString(fmt.Sprintf("> %s\n", shortenItemName(m.Logs[curIdx], termW-8)))
		} else {
			logs.WriteString("\n")
		}
	}

	controls := subtleStyle.Render(fmt.Sprintf(
		"[Space] %s | [↑/↓] %s | [F] %s | [Tab] %s | [+/-] %s | [E] %s | [I] %s | [S] %s | [L] Lang | [Q] %s",
		T(m.Lang, "ui.pause"),
		T(m.Lang, "ui.logs"),
		T(m.Lang, "ui.flee"),
		T(m.Lang, "ui.tabs"),
		T(m.Lang, "ui.speed"),
		T(m.Lang, "ui.btn_armory"),
		T(m.Lang, "ui.btn_codex"),
		T(m.Lang, "ui.glory"),
		T(m.Lang, "ui.quit"),
	))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		strings.TrimRight(logs.String(), "\n"),
		controls,
	)
}

