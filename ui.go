package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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
	w := runewidth.StringWidth(name)
	if w <= maxLen {
		return name
	}
	runes := []rune(name)
	res := ""
	curW := 0
	for _, r := range runes {
		rw := runewidth.RuneWidth(r)
		if curW+rw >= maxLen {
			break
		}
		res += string(r)
		curW += rw
	}
	return res + "…"
}

func padRight(s string, targetWidth int) string {
	w := runewidth.StringWidth(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

func getRaceGlyph(r RaceType, lang Language) string {
	switch r {
	case RaceElf:
		if lang == LangEN {
			return "E"
		}
		return "Э"
	case RaceBeastman:
		if lang == LangEN {
			return "B"
		}
		return "З"
	case RaceOlongr:
		if lang == LangEN {
			return "O"
		}
		return "О"
	default:
		if lang == LangEN {
			return "H"
		}
		return "Ч"
	}
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
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) v2.5.0")) + "\n\n")

	var textBlock string
	if m.Lang == LangEN {
		textBlock = "Welcome to the grim tactical dungeon crawler!\n" +
			"Guide your autonomous squad through INFINITE floors and ruthless biomes.\n\n" +
			questStyle.Render("SYSTEM FEATURES:") + "\n" +
			" • Bilingual Engine (RU / EN) switched on the fly with [L].\n" +
			" • 10 Unique Classes across 4 Diverse Races with innate perks.\n" +
			" • Dynamic establishments: «" + T(m.Lang, m.TownEst.TavernKey) + "», «" + T(m.Lang, m.TownEst.SmithyKey) + "».\n" +
			" • Strategic evacuation: rescue veteran bodies equal to survivors count.\n\n" +
			dangerStyle.Render(fmt.Sprintf("⏳ Expedition autostarts in: %d sec...\n\n", m.MenuCountdown)) +
			healStyle.Render("[Space] or [Enter] — Embark immediately") + "\n" +
			accentStyle.Render("[L] — Switch language (RU / EN)") + "\n" +
			subtleStyle.Render("[I] — Codex  |  [E] — Armory  |  [S] — Stats  |  [Q] — Quit")
	} else {
		textBlock = "Добро пожаловать в мрачный тактический подземельный рогалик!\n" +
			"Ваша задача — провести отряд сквозь БЕСКОНЕЧНЫЕ этажи опаснейших биомов.\n\n" +
			questStyle.Render("ОСОБЕННОСТИ СИСТЕМЫ:") + "\n" +
			" • Двуязычный движок (RU / EN) с переключением на лету клавишей [L].\n" +
			" • 10 специализированных классов и 4 расы со своими врождёнными дарами.\n" +
			" • Колоритные заведения: «" + T(m.Lang, m.TownEst.TavernKey) + "», «" + T(m.Lang, m.TownEst.SmithyKey) + "».\n" +
			" • Эвакуация при побеге: вынос тел ценных ветеранов по числу выживших.\n\n" +
			dangerStyle.Render(fmt.Sprintf("⏳ Автоматический старт экспедиции через: %d сек...\n\n", m.MenuCountdown)) +
			healStyle.Render("[Пробел] или [Enter] — Начать экспедицию немедленно") + "\n" +
			accentStyle.Render("[L] — Сменить язык (RU / EN)") + "\n" +
			subtleStyle.Render("[I] — Кодекс  |  [E] — Арсенал  |  [S] — Слава  |  [Q] — Выход")
	}

	alignedText := lipgloss.NewStyle().Align(lipgloss.Left).Render(textBlock)
	sb.WriteString(alignedText)

	box := menuBoxStyle.Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.legacy_header") + ":\n"))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", T(m.Lang, "stats.legacy_treasury"), goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
	sb.WriteString(fmt.Sprintf(" • %s «%s»: Ур.%d | %s «%s»: Ур.%d\n",
		T(m.Lang, "town.smithy"), T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel,
		T(m.Lang, "town.tannery"), T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel))
	sb.WriteString(fmt.Sprintf(" • %s «%s»: Ур.%d | %s «%s»: Ур.%d\n\n",
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
			raceStr := h.RaceName(m.Lang)
			sb.WriteString(fmt.Sprintf(" • %-16s (%s %s %d) [%s] (Atk:%2d Def:%2d)\n",
				shortenItemName(heroName, 16), raceStr, h.ShortClass(m.Lang), h.Level, healStyle.Render(T(m.Lang, "ui.alive")),
				h.TotalAtk(), h.TotalDef()))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s (%d):\n", T(m.Lang, "stats.fallen_heroes"), len(m.Stats.FallenHeroes))))
	if len(m.Stats.FallenHeroes) == 0 {
		sb.WriteString(subtleStyle.Render(" " + T(m.Lang, "stats.no_fallen") + "\n"))
	} else {
		for _, f := range m.Stats.FallenHeroes {
			clsStr := TranslateEnum(m.Lang, "class", string(f.Class)+".short")
			sb.WriteString(fmt.Sprintf(" ☠️ %-16s (%s) | %s:%d | %s\n",
				dangerStyle.Render(shortenItemName(f.FullName, 16)), clsStr, T(m.Lang, "ui.floor"), f.Floor, subtleStyle.Render(f.Cause)))
		}
	}

	if m.State == StateDefeat {
		countdownStr := dangerStyle.Render(fmt.Sprintf(T(m.Lang, "defeat.restart_timer"), m.RestartCountdown))
		sb.WriteString("\n" + countdownStr + "\n")
	}

	sb.WriteString(subtleStyle.Render("\n[↑/↓/PgUp/PgDn] " + T(m.Lang, "ui.scroll") + " | [R] " + T(m.Lang, "ui.btn_restart") + " | [S] " + T(m.Lang, "ui.btn_back") + " | [L] Lang | [Q] " + T(m.Lang, "ui.quit")))

	lines := strings.Split(sb.String(), "\n")
	maxVisibleLines := max(8, m.TermHeight-4)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]

	box := statsBoxStyle.Width(max(40, m.TermWidth-4)).Render(strings.Join(visibleLines, "\n"))
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderArmoryScreen() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(T(m.Lang, "armory.title") + " [E]")
	sb.WriteString(header + "\n\n")

	renderSlotInfo := func(slotName string, it *EquipItem) string {
		if it == nil {
			return fmt.Sprintf("   • %-7s: %s\n", slotName, subtleStyle.Render("-"))
		}
		statLabel := T(m.Lang, "ui.def_short")
		if it.Slot == SlotWeapon {
			statLabel = T(m.Lang, "ui.atk_short")
		}
		upg := ""
		if it.UpgradeLevel > 0 {
			upg = fmt.Sprintf("+%d ", it.UpgradeLevel)
		}
		return fmt.Sprintf("   • %-7s: %s%s (%s %d)\n",
			slotName, upg, goldStyle.Render(shortenItemName(T(m.Lang, it.BaseNameKey), 18)), statLabel, it.TotalStat())
	}

	for _, h := range m.Party {
		status := healStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.alive")))
		if h.IsDead {
			status = dangerStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.dead")))
		}

		heroTitle := shortenItemName(h.FullName(m.Lang), 20)
		if h.TitleKey != "" {
			heroTitle = titleStyle.Render(heroTitle)
		}

		sb.WriteString(fmt.Sprintf("👤 %s (%s, %s %d) %s\n", heroTitle, h.RaceName(m.Lang), h.ShortClass(m.Lang), h.Level, status))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.weapon"), h.Weapon))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.head"), h.Head))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.chest"), h.Chest))
		sb.WriteString(renderSlotInfo(T(m.Lang, "slot.legs"), h.Legs))
		sb.WriteString("\n")
	}

	sb.WriteString(subtleStyle.Render("[↑/↓/PgUp/PgDn] " + T(m.Lang, "ui.scroll") + " | [E/Esc] " + T(m.Lang, "ui.btn_back") + " | [L] Lang | [Q] " + T(m.Lang, "ui.quit")))

	lines := strings.Split(sb.String(), "\n")
	maxVisibleLines := max(8, m.TermHeight-4)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]

	box := statsBoxStyle.Width(max(40, m.TermWidth-4)).Render(strings.Join(visibleLines, "\n"))
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderTownHub(viewW, viewH int) string {
	cActive := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	cIdle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	cHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	fmtBld := func(phase TownPhase, icon, bNameKey string) string {
		isActive := m.TownPhase == phase
		name := T(m.Lang, bNameKey)
		label := fmt.Sprintf("%s «%s»", icon, name)
		if isActive {
			return cActive.Render(fmt.Sprintf("►[%s]◄", label))
		}
		return cIdle.Render(fmt.Sprintf(" [%s] ", label))
	}

	bMarket := cIdle.Render(" [⚖️ " + T(m.Lang, "town.market") + "] ")
	if m.TownPhase == TownPhaseSellLoot {
		bMarket = cActive.Render("►[⚖️ " + T(m.Lang, "town.market") + "]◄")
	}

	bMagistrate := cIdle.Render(" [🏛️ " + T(m.Lang, "town.magistrate") + "] ")
	if m.TownPhase == TownPhaseMagistrate {
		bMagistrate = cActive.Render("►[🏛️ " + T(m.Lang, "town.magistrate") + "]◄")
	}

	bChurch := fmtBld(TownPhaseChurch, "⛪", m.TownEst.ChurchKey)
	bTavern := fmtBld(TownPhaseTavern, "🍻", m.TownEst.TavernKey)
	bGuild := fmtBld(TownPhaseGuild, "⚔️", m.TownEst.GuildKey)
	bSmithy := fmtBld(TownPhaseSmithy, "⚒️", m.TownEst.SmithyKey)
	bTannery := fmtBld(TownPhaseTannery, "🎒", m.TownEst.TanneryKey)
	bAlchemist := fmtBld(TownPhaseAlchemist, "🧪", m.TownEst.AlchemistKey)

	var sb strings.Builder
	sb.WriteString(cHeader.Render(shortenItemName(T(m.Lang, "town.hub_title"), viewW)) + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", bMarket, bMagistrate))
	sb.WriteString(fmt.Sprintf("%s %s %s\n", bChurch, bTavern, bGuild))
	sb.WriteString(fmt.Sprintf("%s %s %s\n", bSmithy, bTannery, bAlchemist))

	if len(m.TownHistory) > 0 {
		sb.WriteString(subtleStyle.Render(strings.Repeat("─", viewW)) + "\n")
		for _, act := range m.TownHistory {
			sb.WriteString(fmt.Sprintf("• %s\n", shortenItemName(act, viewW-4)))
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
	innerWidth := max(16, cardWidth-4)

	nameRaw := shortenItemName(h.FullName(m.Lang), innerWidth)
	nameStr := nameRaw
	if h.TitleKey != "" {
		nameStr = titleStyle.Render(nameRaw)
	}

	rGlyph := getRaceGlyph(h.Race, m.Lang)
	classFormatted := subtleStyle.Render(fmt.Sprintf("[%s|%s %d]", rGlyph, h.ShortClass(m.Lang), h.Level))

	potSlot := ""
	for i := 0; i < m.MaxPotionSlots(); i++ {
		if i < len(h.Potions) && h.Potions[i] != nil {
			potSlot += h.Potions[i].Symbol
		} else {
			potSlot += "·"
		}
	}

	var sb strings.Builder
	if h.IsDead {
		sb.WriteString(fmt.Sprintf("%s\n", nameStr))
		sb.WriteString(fmt.Sprintf("%s %s\n", classFormatted, dangerStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.dead")))))
		sb.WriteString(subtleStyle.Render(shortenItemName(h.CauseOfDeath, innerWidth)) + "\n")
		sb.WriteString(subtleStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.awaiting_revive"))))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s\n", nameStr, classFormatted))

		barLen := max(2, (innerWidth-16)/2)
		hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
		mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("HP:%s%2d MP:%s%2d\n", hpBar, h.HP, mpBar, h.MP))

		stressColor := lipgloss.Color("135")
		if h.Stress >= 140 {
			stressColor = lipgloss.Color("196")
		}
		stressBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("ST:%s%3d ⚔%-2d 🛡%-2d\n", stressBar, h.Stress, h.TotalAtk(), h.TotalDef()))

		wStr := "-"
		if h.Weapon != nil {
			wStr = shortenItemName(T(m.Lang, h.Weapon.BaseNameKey), 8)
		}
		chStr := "-"
		if h.Chest != nil {
			chStr = shortenItemName(T(m.Lang, h.Chest.BaseNameKey), 8)
		}
		equipLine := shortenItemName(fmt.Sprintf("⚔%s 🛡%s [%s]", wStr, chStr, potSlot), innerWidth)
		sb.WriteString(subtleStyle.Render(equipLine))
	}

	style := heroCardStyle.Width(innerWidth)
	if h.IsDead {
		style = heroCardDead.Width(innerWidth)
	}
	return style.Render(sb.String())
}

func (m Model) renderPartyBanner(cardWidth int) string {
	innerWidth := max(16, cardWidth-4)
	var sb strings.Builder
	sb.WriteString(goldStyle.Render("👑 "+T(m.Lang, "ui.party_and_relic")) + "\n")

	relicName := T(m.Lang, "ui.none")
	if m.Relic != nil {
		relicName = T(m.Lang, m.Relic.NameKey)
	}
	sb.WriteString(fmt.Sprintf("%s: %s\n", T(m.Lang, "ui.relic_short"), shortenItemName(relicName, innerWidth-8)))
	sb.WriteString(fmt.Sprintf("🎒 %s: %d/%d %s\n", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")))
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)
	sb.WriteString(healStyle.Render(fmt.Sprintf("🏛️ +%dG", legacyPart)))

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
		lines = append(lines, dangerStyle.Render(shortenItemName(T(m.Lang, "combat.enemy_pack")+":", innerRightW)))
		availRows := max(1, maxLines-1)
		displayed := 0
		for _, mob := range m.Combat.Pack.Members {
			if displayed >= availRows {
				break
			}
			stText := fmt.Sprintf("%d/%d", mob.HP, mob.MaxHP)
			if mob.IsDead {
				stText = T(m.Lang, "ui.dead")
			}
			mobBar := renderBar(mob.HP, mob.MaxHP, 2, lipgloss.Color("196"), lipgloss.Color("238"))

			mobNameFixed := padRight(shortenItemName(T(m.Lang, mob.NameKey), 12), 12)
			lines = append(lines, fmt.Sprintf("• %s %s [%s]", mobNameFixed, mobBar, stText))
			displayed++
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render(T(m.Lang, "town.management")+":"))
		lines = append(lines, fmt.Sprintf("⚒️ %s: Ур.%-2d", shortenItemName(T(m.Lang, m.TownEst.SmithyKey), 14), m.Legacy.SmithyLevel))
		lines = append(lines, fmt.Sprintf("🎒 %s: Ур.%-2d", shortenItemName(T(m.Lang, m.TownEst.TanneryKey), 14), m.Legacy.TanneryLevel))
		lines = append(lines, fmt.Sprintf("🏛️ %s: Ур.%-2d", shortenItemName(T(m.Lang, m.TownEst.ChurchKey), 14), m.Legacy.ChurchLevel))
		lines = append(lines, fmt.Sprintf("🍻 %s: Ур.%-2d", shortenItemName(T(m.Lang, m.TownEst.TavernKey), 14), m.Legacy.TavernLevel))
		lines = append(lines, fmt.Sprintf("%s: %d/%d", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity()))
	} else {
		lines = append(lines, accentStyle.Render(T(m.Lang, "ui.scouting")+":"))
		lines = append(lines, fmt.Sprintf("%s: %d", T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps))
		lines = append(lines, fmt.Sprintf("%s: %d", T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened))
	}
	return lines
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

	termW := max(38, m.TermWidth)
	termH := max(22, m.TermHeight)
	usableW := termW - 2

	isMobile := termW < 80

	cardsPerRow := 3
	if usableW < 75 {
		cardsPerRow = 1
	} else if usableW < 125 {
		cardsPerRow = 2
	} else if usableW >= 150 {
		cardsPerRow = 6
	}

	singleCardWidth := max(20, (usableW-(cardsPerRow-1)*1)/cardsPerRow)

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
	realMiddleH := lipgloss.Height(middleTier)

	desiredBottomH := max(4, int(float64(termH)*0.16))
	logInnerH := max(1, desiredBottomH-2)

	var logContent strings.Builder
	totalLogs := len(m.Logs)
	endIdx := totalLogs - m.LogScroll
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	if endIdx < logInnerH {
		endIdx = min(totalLogs, logInnerH)
	}
	startIdx := max(0, endIdx-logInnerH)

	for i := 0; i < logInnerH; i++ {
		curIdx := startIdx + i
		if curIdx < endIdx && curIdx < totalLogs {
			logContent.WriteString(fmt.Sprintf("> %s\n", shortenItemName(m.Logs[curIdx], usableW-4)))
		} else {
			logContent.WriteString("\n")
		}
	}

	scrollBadge := ""
	if m.LogScroll > 0 {
		scrollBadge = fmt.Sprintf(" (-%d)", m.LogScroll)
	}
	logBoxTitle := fmt.Sprintf("📜 %s%s", T(m.Lang, "ui.chronicles"), scrollBadge)

	logBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("242")).
		Width(usableW - 2).
		Height(logInnerH)

	renderedLogBox := logBoxStyle.Render(fmt.Sprintf("%s\n%s",
		subtleStyle.Render(logBoxTitle),
		strings.TrimRight(logContent.String(), "\n"),
	))
	realBottomH := lipgloss.Height(renderedLogBox) + 1

	realTopH := termH - realMiddleH - realBottomH
	if realTopH < 5 {
		realTopH = 5
	}

	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("%s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("%s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}

	questTitleShort := shortenItemName(T(m.Lang, m.CurrentQuest.TitleKey), 16)
	statusBadge := subtleStyle.Render(fmt.Sprintf("[%d/%d]", m.CurrentQuest.Current, m.CurrentQuest.TargetCount))
	if m.CurrentQuest.Completed {
		statusBadge = healStyle.Render(fmt.Sprintf("[%s]", T(m.Lang, "ui.turn_in")))
	}

	var topTier string

	if isMobile {
		headerLine := fmt.Sprintf("🏰 %s | 💰 %dG | 📜 %s %s",
			shortenItemName(floorTag, 14), m.Gold, questTitleShort, statusBadge)
		headerBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Width(usableW).
			Render(shortenItemName(headerLine, usableW))

		mapH := max(3, realTopH-3)
		mapStr := m.renderMap(usableW-2, mapH)
		mapBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(usableW - 2).
			Height(mapH).
			Render(mapStr)

		topTier = lipgloss.JoinVertical(lipgloss.Left, headerBox, mapBox)
	} else {
		sideW := max(24, min(40, int(float64(usableW)*0.28)))
		mapBoxW := usableW - sideW - 1

		mapInnerW := max(10, mapBoxW-2)
		mapInnerH := max(3, realTopH-2)

		mapStr := m.renderMap(mapInnerW, mapInnerH)
		leftMapBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(mapInnerW).
			Height(mapInnerH).
			Render(mapStr)

		sidebarInnerW := max(16, sideW-2)
		sidebarInnerH := mapInnerH

		var sbLines []string
		sbLines = append(sbLines, fmt.Sprintf("🏰 %s", titleStyle.Render(shortenItemName(floorTag, sidebarInnerW))))
		sbLines = append(sbLines, fmt.Sprintf("💰 %s: %s", T(m.Lang, "ui.treasury"), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
		sbLines = append(sbLines, fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge))
		sbLines = append(sbLines, subtleStyle.Render(strings.Repeat("─", sidebarInnerW)))

		availRows := max(1, sidebarInnerH-len(sbLines))
		sbLines = append(sbLines, m.renderRightContentLines(sidebarInnerW, availRows)...)

		if len(sbLines) > sidebarInnerH {
			sbLines = sbLines[:sidebarInnerH]
		}
		for len(sbLines) < sidebarInnerH {
			sbLines = append(sbLines, "")
		}

		rightPane := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(sidebarInnerW).
			Height(sidebarInnerH).
			Render(strings.Join(sbLines, "\n"))

		topTier = lipgloss.JoinHorizontal(lipgloss.Top, leftMapBox, " ", rightPane)
	}

	controlsText := "[Space] Пауза | [F] Побег | [E] Арсенал | [I] Кодекс | [S] Слава | [+/-] Скор. | [L] Язык | [Q] Выход"
	if m.Lang == LangEN {
		controlsText = "[Space] Pause | [F] Flee | [E] Armory | [I] Codex | [S] Glory | [+/-] Speed | [L] Lang | [Q] Quit"
	}
	controls := subtleStyle.Render(shortenItemName(controlsText, termW-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		renderedLogBox,
		controls,
	)
}