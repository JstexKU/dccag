package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// ============================================================
// Layout & Card Mode
// ============================================================

type LayoutMode int

const (
	LayoutLandscape LayoutMode = iota
	LayoutPortraitWide
	LayoutPortrait
)

type CardMode int

const (
	CardWide CardMode = iota
	CardMedium
	CardCompact
)

func detectLayout(termW, termH int) LayoutMode {
	if termW <= 0 || termH <= 0 {
		return LayoutLandscape
	}
	ratio := float64(termW) / float64(termH)
	switch {
	case ratio >= 1.5 || termW >= 160:
		return LayoutLandscape
	case ratio >= 0.9 && termW >= 60:
		return LayoutPortraitWide
	default:
		return LayoutPortrait
	}
}

func detectCardMode(innerW int) CardMode {
	switch {
	case innerW >= 30:
		return CardWide
	case innerW >= 22:
		return CardMedium
	default:
		return CardCompact
	}
}

func cardContentHeight(mode CardMode) int {
	switch mode {
	case CardWide:
		return 10
	case CardMedium:
		return 8
	default:
		return 6
	}
}

func landscapeCardsPerRow(usableW int) int {
	switch {
	case usableW >= 190:
		return 6
	case usableW >= 155:
		return 5
	case usableW >= 125:
		return 4
	case usableW >= 90:
		return 3
	case usableW >= 55:
		return 2
	default:
		return 1
	}
}

// ============================================================
// Утилиты
// ============================================================

func padRightTruncate(s string, targetWidth int) string {
	if targetWidth <= 0 {
		return ""
	}
	w := runewidth.StringWidth(stripANSI(s))
	if w > targetWidth {
		return truncatePlain(s, targetWidth)
	}
	if w == targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

func stripANSI(s string) string {
	var sb strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if c == 'm' {
				inEsc = false
			}
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

func truncatePlain(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	plain := stripANSI(s)
	runes := []rune(plain)
	if len(runes) <= maxW {
		return s
	}
	res := ""
	curW := 0
	for _, r := range runes {
		rw := runewidth.RuneWidth(r)
		if curW+rw >= maxW {
			break
		}
		res += string(r)
		curW += rw
	}
	return res + "…"
}

func blankLines(n, w int) string {
	if n <= 0 {
		return ""
	}
	pad := strings.Repeat(" ", max(0, w))
	var sb strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(pad)
	}
	return sb.String()
}

func truncateLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

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

// ============================================================
// Меню
// ============================================================

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
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) v2.7.0")) + "\n\n")

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
			subtleStyle.Render("[I] — Codex  |  [E] — Armory  |  [S] — Stats  |  [+/-] — Speed  |  [Q] — Quit")
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
			subtleStyle.Render("[I] — Кодекс  |  [E] — Арсенал  |  [S] — Слава  |  [+/-] — Скор.  |  [Q] — Выход")
	}

	alignedText := lipgloss.NewStyle().Align(lipgloss.Left).Render(textBlock)
	sb.WriteString(alignedText)

	box := menuBoxStyle.Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// ============================================================
// Экран статистики
// ============================================================

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.legacy_header") + ":\n"))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", T(m.Lang, "stats.legacy_treasury"), goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
	sb.WriteString(fmt.Sprintf(" • %s «%s»: Ур.%d | %s «%s»: Ур.%d\n",
		T(m.Lang, "town.smithy"), T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel,
		T(m.Lang, "town.tannery"), T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel))
	sb.WriteString(fmt.Sprintf(" • %s «%s»: Ур.%d | %s «%s»: Ур.%d\n",
		T(m.Lang, "town.church"), T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel,
		T(m.Lang, "town.tavern"), T(m.Lang, m.TownEst.TavernKey), m.Legacy.TanneryLevel))

	// NEW: строка о текущем тире ополчения и прогрессе до следующего.
	tier := militiaTierForInvestment(m.Legacy.TotalInvested)
	militiaLine := fmt.Sprintf(" • %s: %s (%dG)",
		T(m.Lang, "stats.militia_tier"),
		titleStyle.Render(T(m.Lang, tier.TitleKey)),
		m.Legacy.TotalInvested)
	if nxt := nextMilitiaTier(m.Legacy.TotalInvested); nxt != nil {
		militiaLine += fmt.Sprintf(" → %s (%dG)",
			T(m.Lang, nxt.TitleKey), nxt.InvestFloor)
	}
	sb.WriteString(militiaLine + "\n\n")

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

// ============================================================
// Экран арсенала
// ============================================================

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

// ============================================================
// Карта и город
// ============================================================

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
	out := make([]string, viewH)
	for i := 0; i < viewH; i++ {
		if i < len(lines) {
			out[i] = lines[i]
		} else {
			out[i] = ""
		}
	}
	return strings.Join(out, "\n")
}

func (m Model) renderMap(viewW, viewH int) string {
	if viewW <= 0 || viewH <= 0 {
		return ""
	}
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

	out := make([]string, viewH)
	for y := 0; y < viewH; y++ {
		mapY := startY + y
		var sb strings.Builder
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
		out[y] = sb.String()
	}
	return strings.Join(out, "\n")
}

// ============================================================
// Карточка героя
// ============================================================

func (m Model) renderHeroCard(h *Hero, cardWidth int) string {
	innerWidth := max(16, cardWidth-4)
	mode := detectCardMode(innerWidth)

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

	nameRaw := shortenItemName(h.FullName(m.Lang), innerWidth)
	nameStr := nameRaw
	if h.TitleKey != "" {
		nameStr = titleStyle.Render(nameRaw)
	}
	sb.WriteString(padRightTruncate(nameStr, innerWidth) + "\n")

	if h.IsDead {
		sb.WriteString(renderDeadHeroCard(h, innerWidth, mode, m.Lang))
	} else {
		switch mode {
		case CardWide:
			sb.WriteString(renderWideHeroCard(h, innerWidth, m))
		case CardMedium:
			sb.WriteString(renderMediumHeroCard(h, innerWidth, m))
		default:
			sb.WriteString(renderCompactHeroCard(h, innerWidth, m))
		}
	}

	style := heroCardStyle.Width(innerWidth)
	if h.IsDead {
		style = heroCardDead.Width(innerWidth)
	} else if isActiveTurn {
		style = heroCardActive.Width(innerWidth)
	}
	return style.Render(sb.String())
}

func renderDeadHeroCard(h *Hero, innerWidth int, mode CardMode, lang Language) string {
	var sb strings.Builder

	infoLine := fmt.Sprintf("[%s | %s]", h.RaceName(lang), h.ShortClass(lang))
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, innerWidth)) + "\n")
	sb.WriteString(dangerStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.dead")), innerWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(h.CauseOfDeath, innerWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.awaiting_revive")), innerWidth)))

	targetLines := cardContentHeight(mode)
	padCount := targetLines - 4
	if padCount > 0 {
		sb.WriteString("\n" + blankLines(padCount, innerWidth))
	}
	return sb.String()
}

// renderWideHeroCard — CardWide, 10 строк контента.
// Резерв обвеса: "HP[" (3) + "]" (1) + "99999/99999" (11) = 15.
func renderWideHeroCard(h *Hero, innerWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s | %s %d]", h.RaceName(m.Lang), h.ShortClass(m.Lang), h.Level), innerWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, innerWidth)) + "\n")

	barLen := innerWidth - 15
	if barLen < 3 {
		barLen = 3
	}

	hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("HP%s%d/%d", hpBar, h.HP, h.MaxHP), innerWidth) + "\n")

	mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP%s%d/%d", mpBar, h.MP, h.MaxMP), innerWidth) + "\n")

	stressColor := lipgloss.Color("135")
	if h.Stress >= 140 {
		stressColor = lipgloss.Color("196")
	}
	stBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST%s%d/200", stBar, h.Stress), innerWidth) + "\n")

	sb.WriteString(renderBeltAndStats(h, innerWidth, m) + "\n")

	sb.WriteString(renderEquipLine("⚔", h.Weapon, innerWidth, m.Lang))
	sb.WriteString(renderEquipLine("🛡", h.Chest, innerWidth, m.Lang))
	sb.WriteString(renderEquipLine("🪖", h.Head, innerWidth, m.Lang))
	sb.WriteString(renderEquipLine("🥾", h.Legs, innerWidth, m.Lang))

	return sb.String()
}

// renderMediumHeroCard — CardMedium, 8 строк контента.
// Резерв: "HP[" (3) + "]" (1) + "99999" (5) = 9.
func renderMediumHeroCard(h *Hero, innerWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s | %s %d]", h.RaceName(m.Lang), h.ShortClass(m.Lang), h.Level), innerWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, innerWidth)) + "\n")

	barLen := innerWidth - 9
	if barLen < 3 {
		barLen = 3
	}

	hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("HP%s%d", hpBar, h.HP), innerWidth) + "\n")

	mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP%s%d", mpBar, h.MP), innerWidth) + "\n")

	stressColor := lipgloss.Color("135")
	if h.Stress >= 140 {
		stressColor = lipgloss.Color("196")
	}
	stBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST%s%d", stBar, h.Stress), innerWidth) + "\n")

	sb.WriteString(renderBeltAndStats(h, innerWidth, m) + "\n")

	halfW := innerWidth / 2
	rightW := innerWidth - halfW

	wCell := formatEquipCell("⚔", h.Weapon, halfW, m.Lang)
	cCell := formatEquipCell("🛡", h.Chest, rightW, m.Lang)
	sb.WriteString(goldStyle.Render(wCell) + goldStyle.Render(cCell) + "\n")

	hCell := formatEquipCell("🪖", h.Head, halfW, m.Lang)
	lCell := formatEquipCell("🥾", h.Legs, rightW, m.Lang)
	sb.WriteString(goldStyle.Render(hCell) + goldStyle.Render(lCell))

	return sb.String()
}

// renderCompactHeroCard — CardCompact, 6 строк контента.
// Без бара — только числа. Статы с 2-значным резервом.
func renderCompactHeroCard(h *Hero, innerWidth int, m Model) string {
	var sb strings.Builder

	infoLine := shortenItemName(fmt.Sprintf("[%s %d]", h.ShortClass(m.Lang), h.Level), innerWidth)
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, innerWidth)) + "\n")

	sb.WriteString(padRightTruncate(fmt.Sprintf("HP %d/%d", h.HP, h.MaxHP), innerWidth) + "\n")
	sb.WriteString(padRightTruncate(fmt.Sprintf("MP %d/%d", h.MP, h.MaxMP), innerWidth) + "\n")
	sb.WriteString(padRightTruncate(fmt.Sprintf("ST %d/200", h.Stress), innerWidth) + "\n")

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
	gapW := innerWidth - leftW - statW
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
	sb.WriteString(goldStyle.Render(padRightTruncate(fmt.Sprintf("⚔%s 🛡%s", wName, cName), innerWidth)))

	return sb.String()
}

// ============================================================
// Хелперы для карточек
// ============================================================

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

// renderBeltAndStats — пояс зелий + бафф + боевые статы.
// Статы идут СРАЗУ после баффа, без растягивания строки.
// Резерв на 3-значные Atk/Def.
func renderBeltAndStats(h *Hero, innerWidth int, m Model) string {
	potSlot := beltSymbols(h, m.MaxPotionSlots())
	styledBuff := buffBadge(h)
	leftBlock := fmt.Sprintf("[%s] %s", potSlot, styledBuff)

	atkStr := fmt.Sprintf("%d", h.TotalAtk())
	defStr := fmt.Sprintf("%d", h.TotalDef())
	statText := fmt.Sprintf("⚔%s 🛡%s", atkStr, defStr)
	styledStats := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(statText)

	// Склеиваем: "[пояс] [бафф] ⚔12 🛡9" — один пробел между блоками.
	line := leftBlock + " " + styledStats

	// Паддим до innerWidth, но если строка длиннее — обрезаем.
	return padRightTruncate(line, innerWidth)
}

// renderEquipLine — Wide: иконка-эмодзи + имя + пробел + стат, БЕЗ trailing.
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

	// icon(2) + пробел(1) + имя + пробел(1) + стат, БЕЗ trailing.
	line := goldStyle.Render(icon+" "+name) + " " + subtleStyle.Render(statVal)
	return padRightTruncate(line, innerWidth) + "\n"
}

// formatEquipCell — Medium: ячейка с эмодзи-иконкой и именем.
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

// ============================================================
// Баннер отряда
// ============================================================

func (m Model) renderPartyBanner(cardWidth int) string {
	innerWidth := max(16, cardWidth-4)
	mode := detectCardMode(innerWidth)

	relicName := T(m.Lang, "ui.none")
	relicDesc := T(m.Lang, "ui.no_relic")
	if m.Relic != nil {
		relicName = T(m.Lang, m.Relic.NameKey)
		relicDesc = T(m.Lang, m.Relic.DescKey)
	}
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)

	var sb strings.Builder

	switch mode {
	case CardWide:
		sb.WriteString(goldStyle.Render(padRightTruncate("👑 ОТРЯД И РЕЛИКВИЯ", innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("%s: %s", T(m.Lang, "ui.relic_short"), relicName), innerWidth) + "\n")
		sb.WriteString(subtleStyle.Render(padRightTruncate(relicDesc, innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("🎒 %s: %d/%d %s", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")), innerWidth) + "\n")
		sb.WriteString(healStyle.Render(padRightTruncate(fmt.Sprintf("🏛️ +%dG", legacyPart), innerWidth)) + "\n")
		sb.WriteString(subtleStyle.Render(padRightTruncate(T(m.Lang, "town.tax_active"), innerWidth)) + "\n")
		sb.WriteString(blankLines(4, innerWidth))

	case CardMedium:
		sb.WriteString(goldStyle.Render(padRightTruncate("👑 ОТРЯД И РЕЛИКВИЯ", innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("%s: %s", T(m.Lang, "ui.relic_short"), shortenItemName(relicName, innerWidth-10)), innerWidth) + "\n")
		sb.WriteString(subtleStyle.Render(padRightTruncate(shortenItemName(relicDesc, innerWidth), innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("🎒 %s: %d/%d %s", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")), innerWidth) + "\n")
		sb.WriteString(healStyle.Render(padRightTruncate(fmt.Sprintf("🏛️ +%dG", legacyPart), innerWidth)) + "\n")
		sb.WriteString(blankLines(3, innerWidth))

	default: // Compact
		sb.WriteString(goldStyle.Render(padRightTruncate("👑 ОТРЯД", innerWidth)) + "\n")

		shortRelic := []rune(relicName)
		if len(shortRelic) > 14 {
			shortRelic = append(shortRelic[:13], '…')
		}
		sb.WriteString(padRightTruncate(string(shortRelic), innerWidth) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("🎒 %d/%d", len(m.Bag), m.currentBagCapacity()), innerWidth) + "\n")
		sb.WriteString(healStyle.Render(padRightTruncate(fmt.Sprintf("🏛️ +%dG", legacyPart), innerWidth)) + "\n")
		sb.WriteString(blankLines(3, innerWidth))
	}

	return lipgloss.NewStyle().
		Width(innerWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Render(sb.String())
}

// ============================================================
// Правый контент / полоска врагов
// ============================================================

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
			lines = append(lines, fmt.Sprintf("• %s %s", mobNamePadded, rightBlock))
			displayed++
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render(shortenItemName(T(m.Lang, "town.management")+":", innerRightW)))
		lines = append(lines, shortenItemName(fmt.Sprintf("⚒️ %s: Ур.%d", T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🎒 %s: Ур.%d", T(m.Lang, m.TownEst.TanneryKey), m.Legacy.TanneryLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🏛️ %s: Ур.%d", T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🍻 %s: Ур.%d", T(m.Lang, m.TownEst.TavernKey), m.Legacy.TanneryLevel), innerRightW))
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
				T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened), width))
	}

	var parts []string
	parts = append(parts, dangerStyle.Render(T(m.Lang, "combat.enemy_pack")+":"))
	for _, mob := range m.Combat.Pack.Members {
		if mob.IsDead {
			continue
		}
		hpStr := fmt.Sprintf("%s %d/%d", T(m.Lang, mob.NameKey), mob.HP, mob.MaxHP)
		parts = append(parts, hpStr)
	}
	joined := strings.Join(parts, "  ")
	return shortenItemName(joined, width)
}

// ============================================================
// View
// ============================================================

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

	layout := detectLayout(termW, termH)
	switch layout {
	case LayoutLandscape:
		return m.renderLandscape(termW, termH)
	case LayoutPortraitWide:
		return m.renderPortraitWide(termW, termH)
	default:
		return m.renderPortrait(termW, termH)
	}
}

// ============================================================
// Landscape
// ============================================================

func (m Model) renderLandscape(termW, termH int) string {
	usableW := termW - 2

	cardsPerRow := landscapeCardsPerRow(usableW)
	singleCardWidth := max(20, (usableW-(cardsPerRow-1)*1)/cardsPerRow)

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, singleCardWidth))
	}
	cards = append(cards, m.renderPartyBanner(singleCardWidth))

	var cardRows []string
	for i := 0; i < len(cards); i += cardsPerRow {
		end := min(len(cards), i+cardsPerRow)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	middleTier := lipgloss.JoinVertical(lipgloss.Left, cardRows...)
	middleH := lipgloss.Height(middleTier)

	logH := 5
	if termH < 30 {
		logH = 3
	}
	if termH < 24 {
		logH = 2
	}
	controlsH := 1

	topMin := 8
	totalTop := termH - logH - controlsH - 1

	if middleH > totalTop-topMin {
		middleH = totalTop - topMin
		if middleH < 6 {
			middleH = 6
		}
		middleTier = truncateLines(middleTier, middleH)
		middleH = lipgloss.Height(middleTier)
	}

	topH := totalTop - middleH
	if topH < 4 {
		topH = 4
	}

	sideW := max(28, min(44, int(float64(usableW)*0.30)))
	mapBoxW := usableW - sideW - 1

	mapInnerW := max(10, mapBoxW-4)
	mapInnerH := max(2, topH-2)

	mapStr := m.renderMap(mapInnerW, mapInnerH)
	leftMapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(mapInnerW).
		Height(topH - 2).
		Render(mapStr)

	sidebarInnerW := max(16, sideW-4)

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

	var sbLines []string
	sbLines = append(sbLines, fmt.Sprintf("🏰 %s", titleStyle.Render(shortenItemName(floorTag, sidebarInnerW))))
	sbLines = append(sbLines, fmt.Sprintf("💰 %s: %s", T(m.Lang, "ui.treasury"), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sbLines = append(sbLines, fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge))
	sbLines = append(sbLines, subtleStyle.Render(strings.Repeat("─", sidebarInnerW)))

	availRows := max(1, topH-2-len(sbLines))
	sbLines = append(sbLines, m.renderRightContentLines(sidebarInnerW, availRows)...)

	rightPane := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(sidebarInnerW).
		Height(topH - 2).
		Render(strings.Join(sbLines, "\n"))

	topTier := lipgloss.JoinHorizontal(lipgloss.Top, leftMapBox, " ", rightPane)

	logBox := m.renderLogBox(usableW-4, logH-2)
	controls := m.renderControls(termW - 2)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		logBox,
		controls,
	)
}

// ============================================================
// PortraitWide
// ============================================================

func (m Model) renderPortraitWide(termW, termH int) string {
	usableW := termW - 2

	mapH := max(6, int(float64(termH)*0.28))
	logH := 5
	if termH < 40 {
		logH = 3
	}
	controlsH := 1

	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("🏰 %s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("🏰 %s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}
	headerLine := fmt.Sprintf("%s | 💰 %dG | 📜 %s",
		shortenItemName(floorTag, usableW-30),
		m.Gold,
		shortenItemName(T(m.Lang, m.CurrentQuest.TitleKey), 14),
	)
	headerBox := titleStyle.Render(shortenItemName(headerLine, usableW))

	mapStr := m.renderMap(usableW-4, mapH-2)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 4).
		Height(mapH - 2).
		Render(mapStr)

	enemyStrip := m.renderEnemyStrip(usableW)

	cols := 2
	cardW := (usableW - (cols - 1)) / cols

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardW))
	}
	cards = append(cards, m.renderPartyBanner(cardW))

	var cardRows []string
	for i := 0; i < len(cards); i += cols {
		end := min(len(cards), i+cols)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	cardsBlock := lipgloss.JoinVertical(lipgloss.Left, cardRows...)

	reservedH := 1 + mapH + 1 + logH + controlsH
	maxCardsH := termH - reservedH
	if maxCardsH < 6 {
		maxCardsH = 6
	}
	cardsBlock = truncateLines(cardsBlock, maxCardsH)

	logBox := m.renderLogBox(usableW-4, logH-2)
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

// ============================================================
// Portrait
// ============================================================

func (m Model) renderPortrait(termW, termH int) string {
	usableW := termW - 2

	mapH := max(5, int(float64(termH)*0.22))
	logH := 5
	if termH < 40 {
		logH = 3
	}
	controlsH := 1

	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("🏰 %s %d: %s", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "biome."+string(biome.Name)))
	if m.InTown {
		floorTag = fmt.Sprintf("🏰 %s %d: [%s]", T(m.Lang, "ui.floor"), m.Floor, T(m.Lang, "town.camp"))
	}
	headerLine := fmt.Sprintf("%s | 💰 %dG", shortenItemName(floorTag, usableW-14), m.Gold)
	headerBox := titleStyle.Render(shortenItemName(headerLine, usableW))

	mapStr := m.renderMap(usableW-4, mapH-2)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 4).
		Height(mapH - 2).
		Render(mapStr)

	enemyStrip := m.renderEnemyStrip(usableW)

	cardW := usableW
	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardW))
	}
	cards = append(cards, m.renderPartyBanner(cardW))
	cardsBlock := lipgloss.JoinVertical(lipgloss.Left, cards...)

	reservedH := 1 + mapH + 1 + logH + controlsH
	maxCardsH := termH - reservedH
	if maxCardsH < 6 {
		maxCardsH = 6
	}
	cardsBlock = truncateLines(cardsBlock, maxCardsH)

	logBox := m.renderLogBox(usableW-4, logH-2)
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

// ============================================================
// Общие блоки
// ============================================================

func (m Model) renderLogBox(innerW, innerH int) string {
	if innerH < 1 {
		innerH = 1
	}
	var logContent strings.Builder
	totalLogs := len(m.Logs)
	endIdx := totalLogs - m.LogScroll
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	if endIdx < innerH {
		endIdx = min(totalLogs, innerH)
	}
	startIdx := max(0, endIdx-innerH)

	for i := 0; i < innerH; i++ {
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

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("242")).
		Width(innerW).
		Height(innerH).
		Render(fmt.Sprintf("%s\n%s",
			subtleStyle.Render(shortenItemName(logBoxTitle, innerW)),
			strings.TrimRight(logContent.String(), "\n"),
		))
}

func (m Model) renderControls(width int) string {
	controlsText := "[Space] Пауза | [F] Побег | [E] Арсенал | [I] Кодекс | [S] Слава | [+/-] Скор. | [L] Язык | [Q] Выход"
	if m.Lang == LangEN {
		controlsText = "[Space] Pause | [F] Flee | [E] Armory | [I] Codex | [S] Glory | [+/-] Speed | [L] Lang | [Q] Quit"
	}
	return subtleStyle.Render(shortenItemName(controlsText, width))
}