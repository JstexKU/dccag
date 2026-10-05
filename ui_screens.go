package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMenuModeToggle() string {
	modeLabel := "⚙️ Автоигра"
	if m.Lang == LangEN {
		modeLabel = "⚙️ Auto-play"
	}
	modeStyle := subtleStyle

	if m.ManualMode {
		modeLabel = "🎮 Ручное управление"
		if m.Lang == LangEN {
			modeLabel = "🎮 Manual control"
		}
		modeStyle = accentStyle.Copy().Bold(true)
	}

	prefix := "[M] Режим:"
	if m.Lang == LangEN {
		prefix = "[M] Mode:"
	}
	return fmt.Sprintf("%s %s", prefix, modeStyle.Render(modeLabel))
}

func (m Model) renderMenuScreen() string {
	termW := max(38, m.TermWidth)
	termH := max(20, m.TermHeight)

	isFullMode := termW >= 120 && termH >= 28
	isMediumMode := termW >= 66 && termH >= 24

	var sb strings.Builder

	switch {
	case isFullMode:
		// 1. Полноразмерный режим (Wide / Tall)
		castleLines := []string{
			`         / \                                                 / \       `,
			`        /   \                  |>>>                         /   \      `,
			`       /_____\                 |                           /_____\     `,
			`      |  .-.  |            _  _|_  _                      |  .-.  |    `,
			`      |  | |  |           |;|_|;|_|;|                     |  | |  |    `,
			`      |  '-'  |           \\.    .  /                     |  '-'  |    `,
			`      |       |            \\:  .  /                      |       |    `,
			`    ,-'-------'-,           ||:   |                     ,-'-------'-,  `,
			`  ,'  /═══════\  '.         ||:.  |                   ,'  /═══════\  '.`,
			` /   /         \   \        ||:  .|                  /   /         \   \`,
			`|   |           |   |       ||:   | ____            |   |           |   |`,
			`|   |           |   |       ||: , !_|__|            |   |           |   |`,
			`|===|===========|===|   ____||_ | |    |            |===|===========|===|`,
			`|   |   DCCAG   |   |  |___|__|_|_|_   |  ____      |   |   DCCAG   |   |`,
			`|   |           |   |      |        |  | |____|     |   |           |   |`,
			`|___|___________|___|      |________|__|_|    |     |___|___________|___|`,
			`  [═══════════════]           /════════\  |___|       [═══════════════]  `,
		}
		sb.WriteString(townArtStyle.Render(strings.Join(castleLines, "\n")) + "\n")
		sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) "+displayVersion())) + "\n\n")

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
				m.renderMenuModeToggle() + "\n" +
				accentStyle.Render("[L] — Switch language (RU / EN)") + "   " + m.menuHeroKeys() + "\n" +
				subtleStyle.Render("[+/-] — Speed  |  [Q] — Quit")
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
				m.renderMenuModeToggle() + "\n" +
				accentStyle.Render("[L] — Сменить язык (RU / EN)") + "   " + m.menuHeroKeys() + "\n" +
				subtleStyle.Render("[+/-] — Скор.  |  [Q] — Выход")
		}
		sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Left).Render(textBlock))
		if line := m.menuHeroLine(); line != "" {
			sb.WriteString("\n" + line)
		}

		box := menuBoxStyle.Render(sb.String())
		return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)

	case isMediumMode:
		// 2. Средний режим (66 - 79 колонок)
		logoLines := []string{
			`╔══════════════════════════════╗`,
			`║ ░░░░▄ ▄░░░ ▄░░░  ▄░░░▄ ▄░░░  ║`,
			`║ ░░ ░░ ░░ ▀ ░░ ▀  ░░ ░░ ░░ ▀  ║`,
			`║ ▒▒ ▒▒ ▒▒   ▒▒    ▒▒▒▒▒ ▒▒ ▄▄ ║`,
			`║ ▓█ ▓█ ▓█   ▓█    ▓█ ▓█ ▓█ ▓█ ║`,
			`║ ▄▄ ▄▄ ▄▄   ▄▄    ▄▄ ▄▄ ▄▄ ▄▄ ║`,
			`║ ░░ ░░ ░░   ░░    ░░ ░░ ░░ ░░ ║`,
			`║ ▒▒ ▒▒ ▒▒ ▄ ▒▒ ▄  ▒▒ ▒▒ ▒▒ ▒▒ ║`,
			`║ ▓▓▓▓▀ ▀▓▓▓ ▀▓▓▓ ▄▓▓ ▓▓ ▀▓▓▓▓ ║`,
			`╚══════════════════════════════╝`,
		}
		sb.WriteString(townArtStyle.Render(strings.Join(logoLines, "\n")) + "\n\n")

		timerStr := dangerStyle.Render(fmt.Sprintf("⏳ %d sec...", m.MenuCountdown))
		startBtn := healStyle.Render("[Space / Enter] Start Expedition")
		if m.Lang == LangRU {
			timerStr = dangerStyle.Render(fmt.Sprintf("⏳ Старт через: %d сек...", m.MenuCountdown))
			startBtn = healStyle.Render("[Пробел / Enter] Начать экспедицию")
		}

		sb.WriteString(startBtn + "    " + timerStr + "\n")
		sb.WriteString(m.renderMenuModeToggle() + "\n\n")
		if line := m.menuHeroLine(); line != "" {
			sb.WriteString(line + "\n")
		}
		sb.WriteString(accentStyle.Render("[L] Language (RU / EN)") + "  " + m.menuHeroKeys() + "\n")
		sb.WriteString(subtleStyle.Render("[Q] Quit"))

		box := menuBoxStyle.Render(sb.String())
		return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)

	default:
		// 3. Компактный режим (Ширина < 66 или Высота < 24)
		innerW := max(34, termW-6)

		headerTitle := "⚔ DCCAG " + displayVersion() + " ⚔"
		tagline := "Tactical Auto Dungeon Crawler"
		if m.Lang == LangRU {
			tagline = "Автономный тактический рогалик"
		}

		sb.WriteString(titleStyle.Render(padRightTruncate(headerTitle, innerW)) + "\n")
		sb.WriteString(subtleStyle.Render(padRightTruncate(tagline, innerW)) + "\n\n")

		timerStr := dangerStyle.Render(fmt.Sprintf("⏳ %d sec", m.MenuCountdown))
		if m.Lang == LangRU {
			timerStr = dangerStyle.Render(fmt.Sprintf("⏳ Автостарт: %d сек", m.MenuCountdown))
		}
		sb.WriteString(timerStr + "\n")
		sb.WriteString(m.renderMenuModeToggle() + "\n\n")
		if line := m.menuHeroLine(); line != "" {
			sb.WriteString(line + "\n")
		}
		sb.WriteString(m.menuHeroKeys() + "\n")

		if m.Lang == LangEN {
			sb.WriteString(healStyle.Render("[Space] Start") + "\n")
			sb.WriteString(accentStyle.Render("[L] Language (EN/RU)") + "\n")
			sb.WriteString(subtleStyle.Render("[Q] Quit"))
		} else {
			sb.WriteString(healStyle.Render("[Пробел] Начать") + "\n")
			sb.WriteString(accentStyle.Render("[L] Язык (RU/EN)") + "\n")
			sb.WriteString(subtleStyle.Render("[Q] Выход"))
		}

		box := menuBoxStyle.Width(innerW).Render(sb.String())
		return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)
	}
}

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	// 1. Наследие
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.legacy_header") + ":"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n",
		T(m.Lang, "stats.legacy_treasury"),
		goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
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

	// 2. Достижения
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.achievements_header") + ":"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(" • %s: %d | %s: %d | %s: %s\n",
		T(m.Lang, "stats.floors_cleared"), m.Stats.FloorsCleared,
		T(m.Lang, "stats.total_steps"), m.Stats.TotalSteps,
		T(m.Lang, "stats.gold_earned"), goldStyle.Render(fmt.Sprintf("%dG", m.Stats.TotalGoldEarned))))
	sb.WriteString(fmt.Sprintf(" • %s: %d | %s: %d | %s: %d\n\n",
		T(m.Lang, "stats.contracts_closed"), m.Stats.QuestsCompleted,
		T(m.Lang, "stats.upgrades_forged"), m.Stats.UpgradesForged,
		T(m.Lang, "stats.chests_opened"), m.Stats.ChestsOpened))

	// 3. Выжившие бойцы
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(T(m.Lang, "stats.survivors_header") + ":"))
	sb.WriteString("\n")
	nameColWidth := 20
	for _, h := range m.Party {
		if !h.IsDead && !h.IsDowned {
			rawName := shortenItemName(h.FullName(m.Lang), nameColWidth)
			paddedName := padRight(rawName, nameColWidth)
			raceStr := h.RaceName(m.Lang)
			sb.WriteString(fmt.Sprintf(" • %s (%s %s %d) [%s] (Atk:%2d Def:%2d)\n",
				paddedName,
				raceStr, h.ShortClass(m.Lang), h.Level, healStyle.Render(T(m.Lang, "ui.alive")),
				h.TotalAtk(), h.TotalDef()))
		}
	}
	sb.WriteString("\n")

	var deadForever []FallenHeroRecord
	var savedList []FallenHeroRecord

	for _, f := range m.Stats.FallenHeroes {
		if f.Revived {
			savedList = append(savedList, f)
		} else {
			deadForever = append(deadForever, f)
		}
	}

	// 4. Павшие навсегда
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("☠️ %s (%d):", T(m.Lang, "stats.fallen_heroes"), len(deadForever))))
	sb.WriteString("\n")
	if len(deadForever) == 0 {
		sb.WriteString("   " + subtleStyle.Render(T(m.Lang, "stats.no_fallen")) + "\n")
	} else {
		for _, f := range deadForever {
			clsStr := TranslateEnum(m.Lang, "class", string(f.Class)+".short")
			rawName := shortenItemName(f.FullName, nameColWidth)
			sb.WriteString(fmt.Sprintf(" • %s (%s) | %s:%d | %s\n",
				dangerStyle.Render(padRight(rawName, nameColWidth)),
				clsStr, T(m.Lang, "ui.floor"), f.Floor, subtleStyle.Render(f.Cause)))
		}
	}
	sb.WriteString("\n")

	// 5. Спасённые и исцелённые
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("✨ %s (%d):", T(m.Lang, "stats.saved_heroes"), len(savedList))))
	sb.WriteString("\n")
	if len(savedList) == 0 {
		sb.WriteString("   " + subtleStyle.Render(T(m.Lang, "stats.no_saved")) + "\n")
	} else {
		for _, f := range savedList {
			clsStr := TranslateEnum(m.Lang, "class", string(f.Class)+".short")
			rawName := shortenItemName(f.FullName, nameColWidth)
			sb.WriteString(fmt.Sprintf(" • %s (%s) | %s:%d | %s\n",
				healStyle.Render(padRight(rawName, nameColWidth)),
				clsStr, T(m.Lang, "ui.floor"), f.Floor, subtleStyle.Render(T(m.Lang, "ui.stabilized"))))
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

	box := statsBoxStyle.
		Width(max(40, m.TermWidth-4)).
		Height(maxVisibleLines).
		Render(strings.Join(visibleLines, "\n"))

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