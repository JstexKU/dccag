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
	case innerW >= 28:
		return CardWide
	case innerW >= 20:
		return CardMedium
	default:
		return CardCompact
	}
}

// cardContentHeight возвращает точное число строк содержимого внутри карточки (без рамки)
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

// cardOuterHeight возвращает полную внешнюю высоту карточки с учетом рамки (top + bottom = +2 строки)
func cardOuterHeight(mode CardMode) int {
	return cardContentHeight(mode) + 2
}

func landscapeCardsPerRow(usableW int) int {
	switch {
	case usableW >= 125:
		return 6
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

// normalizeLines добивает или урезает текст строго до targetCount строк одинаковой ширины innerWidth
func normalizeLines(content string, innerWidth, targetCount int) string {
	rawLines := strings.Split(content, "\n")
	res := make([]string, targetCount)
	for i := 0; i < targetCount; i++ {
		if i < len(rawLines) {
			res[i] = padRightTruncate(rawLines[i], innerWidth)
		} else {
			res[i] = strings.Repeat(" ", innerWidth)
		}
	}
	return strings.Join(res, "\n")
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
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) v2.8.*")) + "\n\n")

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
	sb.WriteString(fmt.Sprintf(" • %s «%s»: %s.%d | %s «%s»: %s.%d\n",
		T(m.Lang, "town.smithy"), T(m.Lang, m.TownEst.SmithyKey), T(m.Lang, "ui.level_short"), m.Legacy.SmithyLevel,
		T(m.Lang, "town.tannery"), T(m.Lang, m.TownEst.TanneryKey), T(m.Lang, "ui.level_short"), m.Legacy.TanneryLevel))
	sb.WriteString(fmt.Sprintf(" • %s «%s»: %s.%d | %s «%s»: %s.%d\n",
		T(m.Lang, "town.church"), T(m.Lang, m.TownEst.ChurchKey), T(m.Lang, "ui.level_short"), m.Legacy.ChurchLevel,
		T(m.Lang, "town.tavern"), T(m.Lang, m.TownEst.TavernKey), T(m.Lang, "ui.level_short"), m.Legacy.TavernLevel))

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
		{TownPhaseSellLoot, "⚖️", T(m.Lang, "town.market"), 0},
		{TownPhaseMagistrate, "🏛️", T(m.Lang, "town.magistrate"), 0},
		{TownPhaseChurch, "⛪", T(m.Lang, m.TownEst.ChurchKey), m.Legacy.ChurchLevel},
		{TownPhaseTavern, "🍻", T(m.Lang, m.TownEst.TavernKey), m.Legacy.TavernLevel},
		{TownPhaseGuild, "⚔", T(m.Lang, m.TownEst.GuildKey), 0},
		{TownPhaseSmithy, "⚒️", T(m.Lang, m.TownEst.SmithyKey), m.Legacy.SmithyLevel},
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
			m.Legacy.ChurchLevel, m.Legacy.TavernLevel), viewW)) + "\n\n")

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
		if i < len(lines) {
			out[i] = padRightTruncate(lines[i], width)
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
			case TileEvent:
				sb.WriteString(eventTileStyle.Render("?"))
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
	innerWidth := max(14, cardWidth-4)
	mode := detectCardMode(innerWidth)
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

	// Строгая нормализация количества строк контента под заданную высоту карточки
	normalizedBody := normalizeLines(sb.String(), innerWidth, targetLines)

	baseStyle := heroCardStyle
	if h.IsDead {
		baseStyle = heroCardDead
	} else if isActiveTurn {
		baseStyle = heroCardActive
	}

	return baseStyle.
		Width(innerWidth).
		Height(targetLines).
		Render(normalizedBody)
}

func renderDeadHeroCard(h *Hero, innerWidth int, mode CardMode, lang Language) string {
	var sb strings.Builder

	infoLine := fmt.Sprintf("[%s | %s]", h.RaceName(lang), h.ShortClass(lang))
	sb.WriteString(subtleStyle.Render(padRightTruncate(infoLine, innerWidth)) + "\n")
	sb.WriteString(dangerStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.dead")), innerWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(h.CauseOfDeath, innerWidth)) + "\n")
	sb.WriteString(subtleStyle.Render(padRightTruncate(fmt.Sprintf("[%s]", T(lang, "ui.awaiting_revive")), innerWidth)))

	return sb.String()
}

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

// ============================================================
// Баннер отряда
// ============================================================

func (m Model) renderPartyBanner(cardWidth int) string {
	innerWidth := max(14, cardWidth-4)
	mode := detectCardMode(innerWidth)
	targetLines := cardContentHeight(mode)

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
		sb.WriteString(subtleStyle.Render(padRightTruncate(T(m.Lang, "town.tax_active"), innerWidth)))

	case CardMedium:
		sb.WriteString(goldStyle.Render(padRightTruncate("👑 ОТРЯД И РЕЛИКВИЯ", innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("%s: %s", T(m.Lang, "ui.relic_short"), shortenItemName(relicName, innerWidth-10)), innerWidth) + "\n")
		sb.WriteString(subtleStyle.Render(padRightTruncate(shortenItemName(relicDesc, innerWidth), innerWidth)) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("🎒 %s: %d/%d %s", T(m.Lang, "ui.bag"), len(m.Bag), m.currentBagCapacity(), T(m.Lang, "ui.slots_short")), innerWidth) + "\n")
		sb.WriteString(healStyle.Render(padRightTruncate(fmt.Sprintf("🏛️ +%dG", legacyPart), innerWidth)))

	default:
		sb.WriteString(goldStyle.Render(padRightTruncate("👑 ОТРЯД", innerWidth)) + "\n")
		shortRelic := []rune(relicName)
		if len(shortRelic) > 14 {
			shortRelic = append(shortRelic[:13], '…')
		}
		sb.WriteString(padRightTruncate(string(shortRelic), innerWidth) + "\n")
		sb.WriteString(padRightTruncate(fmt.Sprintf("🎒 %d/%d", len(m.Bag), m.currentBagCapacity()), innerWidth) + "\n")
		sb.WriteString(healStyle.Render(padRightTruncate(fmt.Sprintf("🏛️ +%dG", legacyPart), innerWidth)))
	}

	normalizedBody := normalizeLines(sb.String(), innerWidth, targetLines)

	return lipgloss.NewStyle().
		Width(innerWidth).
		Height(targetLines).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Render(normalizedBody)
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
		lines = append(lines, shortenItemName(fmt.Sprintf("⚒️ %s: %s.%d", T(m.Lang, m.TownEst.SmithyKey), T(m.Lang, "ui.level_short"), m.Legacy.SmithyLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🎒 %s: %s.%d", T(m.Lang, m.TownEst.TanneryKey), T(m.Lang, "ui.level_short"), m.Legacy.TanneryLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🏛️ %s: %s.%d", T(m.Lang, m.TownEst.ChurchKey), T(m.Lang, "ui.level_short"), m.Legacy.ChurchLevel), innerRightW))
		lines = append(lines, shortenItemName(fmt.Sprintf("🍻 %s: %s.%d", T(m.Lang, m.TownEst.TavernKey), T(m.Lang, "ui.level_short"), m.Legacy.TavernLevel), innerRightW))
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
	if m.State == StateTactics {
		return m.renderTacticsScreen()
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
	usableW := max(38, termW-2)

	cardsPerRow := landscapeCardsPerRow(usableW)
	// Единый расчет внешней ширины для ВСЕХ карточек
	cardWidth := max(16, usableW/cardsPerRow)

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, cardWidth))
	}
	cards = append(cards, m.renderPartyBanner(cardWidth))

	var cardRows []string
	for i := 0; i < len(cards); i += cardsPerRow {
		end := min(len(cards), i+cardsPerRow)
		cardRows = append(cardRows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	middleTier := lipgloss.JoinVertical(lipgloss.Left, cardRows...)
	middleH := lipgloss.Height(middleTier)

	controlsH := 1
	logH := 5
	if termH < 30 {
		logH = 4
	}
	if termH < 25 {
		logH = 3
	}

	availableH := termH - controlsH - logH - 1
	topMinH := 7
	if availableH < topMinH+middleH {
		middleH = max(6, availableH-topMinH)
		middleTier = truncateLines(middleTier, middleH)
		middleH = lipgloss.Height(middleTier)
	}
	topH := max(topMinH, availableH-middleH)

	sideW := max(26, min(40, int(float64(usableW)*0.28)))
	mapBoxW := max(20, usableW-sideW)
	boxInnerW := max(1, mapBoxW-2)
	boxInnerH := max(1, topH-2)

	mapStr := m.renderMap(boxInnerW, boxInnerH)
	leftMapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(boxInnerW).
		Height(boxInnerH).
		Render(mapStr)

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
	sbLines = append(sbLines, fmt.Sprintf("💰 %s: %s", T(m.Lang, "ui.treasury"), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	sbLines = append(sbLines, fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge))
	sbLines = append(sbLines, subtleStyle.Render(strings.Repeat("─", sidebarInnerW)))
	availRows := max(1, boxInnerH-len(sbLines))
	sbLines = append(sbLines, m.renderRightContentLines(sidebarInnerW, availRows)...)

	rightPane := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(sidebarInnerW).
		Height(boxInnerH).
		Render(strings.Join(sbLines, "\n"))

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

	mapStr := m.renderMap(usableW-2, mapH-2)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 2).
		Height(mapH - 2).
		Render(mapStr)

	enemyStrip := m.renderEnemyStrip(usableW)

	cols := 2
	cardW := usableW / cols

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

	mapStr := m.renderMap(usableW-2, mapH-2)
	mapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(usableW - 2).
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

// ============================================================
// Общие блоки
// ============================================================

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

	body := fmt.Sprintf("%s\n%s",
		subtleStyle.Render(shortenItemName(logBoxTitle, innerW)),
		strings.TrimRight(logContent.String(), "\n"),
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("242")).
		Width(innerW).
		Render(body)
}

func (m Model) renderControls(width int) string {
	controlsText := "[Space] Пауза | [F] Побег | [T] Тактика | [E] Арсенал | [I] Кодекс | [S] Слава | [+/-] Скор. | [L] Язык | [Q] Выход"
	if m.Lang == LangEN {
		controlsText = "[Space] Pause | [F] Flee | [T] Tactics | [E] Armory | [I] Codex | [S] Glory | [+/-] Speed | [L] Lang | [Q] Quit"
	}
	return subtleStyle.Render(shortenItemName(controlsText, width))
}
