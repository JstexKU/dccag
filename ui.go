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
	replacer := strings.NewReplacer(
		"Пылающий", "Пыл.",
		"Ядовитый", "Ядов.",
		"Леденящий", "Лед.",
		"Громовой", "Гром.",
		"Закаленный", "Зак.",
		"Кровопийцы", "Кров.",
		"Медитации", "Мед.",
		"Ярости", "Яр.",
		"Титана", "Тит.",
		"Железн.", "Жел.",
		"Стальн.", "Стал.",
		"Мифрил.", "Мифр.",
		"Адамант.", "Адам.",
		"Тисов.", "Тис.",
		"Ясенев.", "Ясен.",
		"Кристальн.", "Крист.",
		"Эфирн.", "Эфир.",
		"Сыромятн.", "Сыр.",
		"Варён.", "Вар.",
		"Василиск.", "Вас.",
		"Драконь.", "Драк.",
		"Льнян.", "Лён.",
		"Шёлков.", "Шёлк.",
		"Парчов.", "Парч.",
		"Астральн.", "Астр.",
	)
	res := replacer.Replace(name)
	runes := []rune(res)
	if len(runes) > maxLen {
		if maxLen <= 1 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-1]) + "…"
	}
	return res
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
		return BiomeConfig{Name: BiomeCatacombs, WallColor: lipgloss.Color("240"), FloorColor: lipgloss.Color("236"), FloorRune: '·', EnvHazard: "Сырость и гниль"}
	case 1:
		return BiomeConfig{Name: BiomeGrotto, WallColor: lipgloss.Color("31"), FloorColor: lipgloss.Color("24"), FloorRune: '~', EnvHazard: "Затопленные плиты (-2 Скор)"}
	case 2:
		return BiomeConfig{Name: BiomeInferno, WallColor: lipgloss.Color("124"), FloorColor: lipgloss.Color("52"), FloorRune: '≈', EnvHazard: "Палящий зной (+Огонь)"}
	case 3:
		return BiomeConfig{Name: BiomeCrystal, WallColor: lipgloss.Color("141"), FloorColor: lipgloss.Color("54"), FloorRune: '◊', EnvHazard: "Искажение эфира (+3 к MP)"}
	default:
		return BiomeConfig{Name: BiomeAbyss, WallColor: lipgloss.Color("89"), FloorColor: lipgloss.Color("233"), FloorRune: '×', EnvHazard: "Дыхание Бездны (+10% стресса)"}
	}
}

type BiomeConfig struct {
	Name       BiomeType
	WallColor  lipgloss.Color
	FloorColor lipgloss.Color
	FloorRune  rune
	EnvHazard  string
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
       ___,-------,,-----------------------,,-------,___
  _,-""             ''---------------------''             ""-,_
,'        /\                                           /\        '.
/        /  \         ______________________          /  \         \
|       / /\ \        |                    |         / /\ \        |
|      | |  | |       |     D C C A G      |        |  | | |       |
|      | |__| |       |____________________|        |__| | |       |
|      |______|         _||          ||_            |______|       |
|       |    |         (_||          ||_)           |    |         |
|       |    |          | |          | |            |    |         |
|      _||||_          _|_|_        _|_|_          _||||_          |
`)

	sb.WriteString(banner + "\n")
	sb.WriteString(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag) v2.3.0\n\n"))

	sb.WriteString("Добро пожаловать в мрачный подземельный рогалик!\n")
	sb.WriteString("Ваша задача — провести отряд сквозь БЕСКОНЕЧНЫЕ этажи опаснейших биомов.\n\n")

	sb.WriteString(questStyle.Render("НОВОВВЕДЕНИЯ ВЕРСИИ 2.3.0:\n"))
	sb.WriteString(" • Гендерные глаголы в логах: «Лира нанесла удар», «Бранд парировал выпад».\n")
	sb.WriteString(" • Колоритные заведения: трактир «" + m.TownEst.TavernName + "», кузня «" + m.TownEst.SmithyName + "».\n")
	sb.WriteString(" • Честные материалы: посохи из ясеня и кристаллов, робы из шёлка, клинки из стали.\n")
	sb.WriteString(" • Вкладки Кодекса [I]: быстрое переключение разделов по клавишам [Tab] и [1-4].\n")
	sb.WriteString(" • Пояса зелий и компактные метки (Рога 15, Жрец 15) без переносов строк.\n\n")

	sb.WriteString(dangerStyle.Render(fmt.Sprintf("⏳ Автоматический старт экспедиции через: %d сек...\n\n", m.MenuCountdown)))
	sb.WriteString(healStyle.Render("[ENTER] или [SPACE] — Начать экспедицию немедленно\n"))
	sb.WriteString(subtleStyle.Render("[Q] — Выход из игры"))

	box := menuBoxStyle.Render(sb.String())
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("НАСЛЕДИЕ КОРОЛЕВСТВА:\n"))
	sb.WriteString(fmt.Sprintf(" • В Казну следующего поколения передано: %s\n", goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
	sb.WriteString(fmt.Sprintf(" • Кузня («%s»): Ур.%d | Кожевник («%s»): Ур.%d\n",
		m.TownEst.SmithyName, m.Legacy.SmithyLevel, m.TownEst.TanneryName, m.Legacy.TanneryLevel))
	sb.WriteString(fmt.Sprintf(" • Храм («%s»): Ур.%d | Таверна («%s»): Ур.%d\n\n",
		m.TownEst.ChurchName, m.Legacy.ChurchLevel, m.TownEst.TavernName, m.Legacy.TavernLevel))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("ДОСТИЖЕНИЯ И КОНТРАКТЫ:\n"))
	sb.WriteString(fmt.Sprintf(" • Зачищено этажей: %d | Шагов: %d | Золота: %s\n", m.Stats.FloorsCleared, m.Stats.TotalSteps, goldStyle.Render(fmt.Sprintf("%dG", m.Stats.TotalGoldEarned))))
	sb.WriteString(fmt.Sprintf(" • Выполнено квестов: %d | Заточек: %d | Вскрыто ларей: %d\n\n", m.Stats.QuestsCompleted, m.Stats.UpgradesForged, m.Stats.ChestsOpened))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("ВЫЖИВШИЕ БОЙЦЫ:\n"))
	for _, h := range m.Party {
		if !h.IsDead {
			heroName := h.FullName()
			if h.Title != "" {
				heroName = titleStyle.Render(heroName)
			}
			sb.WriteString(fmt.Sprintf(" • %-20s (%-5s %2d) [%s] (Atk:%2d | Def:%2d | Скор:%2d)\n",
				heroName, h.ShortClass(), h.Level, healStyle.Render("ВЫЖИЛ"), h.TotalAtk(), h.TotalDef(), h.TotalSpeed()))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("КНИГА ПАМЯТИ (ПАВШИЕ: %d):\n", len(m.Stats.FallenHeroes))))
	if len(m.Stats.FallenHeroes) == 0 {
		sb.WriteString(subtleStyle.Render(" Ни один боец не погиб в этом походе.\n"))
	} else {
		for _, f := range m.Stats.FallenHeroes {
			clsStr := string(f.Class)
			if f.Class == ClassRogue {
				clsStr = "Рога"
			} else if f.Class == ClassCleric {
				clsStr = "Жрец"
			}
			sb.WriteString(fmt.Sprintf(" ☠️ %-18s (%-5s) | Этаж: %d | %s\n",
				dangerStyle.Render(f.FullName), clsStr, f.Floor, subtleStyle.Render(f.Cause)))
		}
	}

	if m.State == StateDefeat {
		countdownStr := dangerStyle.Render(fmt.Sprintf("\n⏳ Автоматический перезапуск через: %d сек...", m.RestartCountdown))
		sb.WriteString(countdownStr + "\n")
	}

	sb.WriteString(subtleStyle.Render("\n[↑/↓/PgUp/PgDn] Прокрутка | [R] Перезапуск | [S] Назад | [Q] Выход"))

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

	cSec := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	cSub := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	cMob := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	cBoss := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	cHp := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	cAtk := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	cDef := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	cSpd := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	cTier := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
	cItem := lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
	cNote := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	cVal := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("➔ ")

	tabNames := []string{"1. Бестиарий", "2. Арсенал", "3. Столица", "4. Алхимия"}
	var tabHeaders []string
	for i, name := range tabNames {
		if m.CodexTab == i {
			tabHeaders = append(tabHeaders, lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render("►["+name+"]◄"))
		} else {
			tabHeaders = append(tabHeaders, subtleStyle.Render(" ["+name+"] "))
		}
	}
	tabsRow := strings.Join(tabHeaders, " ")

	fmtMob := func(name string, nameWidth int, hp, atk, def, spd int, growth, extra string) string {
		nameStr := cMob.Render(padRight(name, nameWidth))
		statsStr := fmt.Sprintf(
			"%s %s | %s %s | %s %s | %s %s",
			cHp.Render("HP"), cVal.Render(fmt.Sprintf("%-2d", hp)),
			cAtk.Render("ATK"), cVal.Render(fmt.Sprintf("%-2d", atk)),
			cDef.Render("DEF"), cVal.Render(fmt.Sprintf("%-1d", def)),
			cSpd.Render("СКОР"), cVal.Render(fmt.Sprintf("%-2d", spd)),
		)
		tail := cNote.Render(fmt.Sprintf("| Рост: %s", growth))
		if extra != "" {
			tail += " " + extra
		}
		return fmt.Sprintf("   • %s: %s %s\n", nameStr, statsStr, tail)
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render("═══ КОДЕКС ЗНАНИЙ И БАЗА ДАННЫХ DCCAG ═══\n\n"))
	sb.WriteString(tabsRow + "\n")
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	switch m.CodexTab {
	case 0: // Вкладка 1: Бестиарий и Биомы
		sb.WriteString(cSec.Render("ЦИКЛИЧЕСКИЕ БИОМЫ:") + "\n")
		sb.WriteString(fmt.Sprintf(" • %s: Базовые монстры, сырость и гниль.\n", cSub.Render("Этажи 1, 6, 11... (Гнилые Катакомбы)")))
		sb.WriteString(fmt.Sprintf(" • %s: Слизни, утопленники, ящеры (-2 к скорости группы).\n", cSub.Render("Этажи 2, 7, 12... (Затопленные Гроты)")))
		sb.WriteString(fmt.Sprintf(" • %s: Пепельные бесы, орки, саламандры (+урон огнем).\n", cSub.Render("Этажи 3, 8, 13... (Пепельные Недра)")))
		sb.WriteString(fmt.Sprintf(" • %s: Големы, гаргульи (+3 MP к стоимости способностей).\n", cSub.Render("Этажи 4, 9, 14... (Кристальный Лабиринт)")))
		sb.WriteString(fmt.Sprintf(" • %s: Демоны, Рыцари Смерти, Дракон (+10%% стресса).\n\n", cSub.Render("Этажи 5, 10, 15... (Трон Бездны)")))

		sb.WriteString(cSec.Render("БЕСТИАРИЙ И МОДИФИКАТОРЫ РОСТА:") + "\n")
		sb.WriteString(fmtMob("Чумная крыса", 14, 14, 5, 0, 8, "+15% HP, +10% ATK", ""))
		sb.WriteString(fmtMob("Гоблин", 14, 17, 7, 1, 8, "+15% HP, +12% ATK", ""))
		sb.WriteString(fmtMob("Скелет", 14, 22, 8, 3, 8, "+18% HP, +15% ATK", subtleStyle.Render("(Блок 20%)")))
		sb.WriteString(fmtMob("Болотный слизень", 16, 26, 9, 1, 9, "+20% HP, +12% ATK", stressStyle.Render("(+10 Стр)")))
		sb.WriteString(fmtMob("Пепельный бес", 14, 36, 13, 2, 10, "+22% HP, +20% ATK", fireStyle.Render("(Опаление)")))
		sb.WriteString(fmtMob("Кристальный голем", 17, 60, 18, 7, 11, "+30% HP, +20% ATK", subtleStyle.Render("(Блок 20%)")))
		sb.WriteString(fmtMob("Рыцарь Смерти", 16, 76, 22, 7, 12, "+32% HP, +28% ATK", dangerStyle.Render("(Вампиризм)")))
		sb.WriteString(fmt.Sprintf("   • %s: %s 260+(Lvl*20) | %s 25+Lvl | %s\n\n",
			cBoss.Render("Пепельный Дракон (БОСС)"), cHp.Render("HP"), cAtk.Render("ATK"), fireStyle.Render("Дыхание огнем")))

		sb.WriteString(cSec.Render("АФФИКСЫ МОНСТРОВ:") + "\n")
		sb.WriteString(fmt.Sprintf(" • %s: +3 к базовой атаке; опаляет героя на +4 чистого урона\n", fireStyle.Render("🔥 Огненный")))
		sb.WriteString(fmt.Sprintf(" • %s: Отравляет раны (-3 чистого HP и +14 стресса)\n", stressStyle.Render("☣️ Ядовитый")))
		sb.WriteString(fmt.Sprintf(" • %s: Замедляет инициативу группы и сковывает действия\n", fountStyle.Render("❄️ Ледяной")))
		sb.WriteString(fmt.Sprintf(" • %s: +3 к защите (DEF), +12 к максимальному запасу здоровья\n", subtleStyle.Render("🪨 Каменный")))
		sb.WriteString(fmt.Sprintf(" • %s: Крадет здоровье: восстанавливает 50%% от нанесенного урона\n\n", dangerStyle.Render("🩸 Вампир")))

	case 1: // Вкладка 2: Арсенал и Ремесло
		sb.WriteString(cSec.Render("РЕМЕСЛЕННЫЕ СЕТКИ И ЭКИПИРОВКА:") + "\n")
		sb.WriteString(cNote.Render(" [Кузница: Металлы x1..x4 и Древесина магов x1..x4]") + "\n")
		sb.WriteString(cNote.Render(" [Кожевник: Органика x1..x4 и Зачарованные ткани x1..x4]") + "\n\n")

		sb.WriteString(cSub.Render(" [КУЗНИЦА: ТАНК И ВОИН - Тяжелые латы и сталь]") + "\n")
		sb.WriteString(fmt.Sprintf("   • Танк оружие: %s Гладиус (%s:3) %s%s Палаш %s%s Моргенштерн %s%s Бастионный меч (%s:9, Блок:8)\n",
			cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
		sb.WriteString(fmt.Sprintf("   • Танк доспех: %s Бригантина (%s:4) %s%s Полудоспех %s%s Кираса бастиона %s%s Панцирь цитадели (%s:13, %s:+40)\n",
			cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))
		sb.WriteString(fmt.Sprintf("   • Воин оружие: %s Эспадон (%s:5) %s%s Клеймор %s%s Боевой топор %s%s Фальшион (%s:14, Крит:4)\n",
			cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
		sb.WriteString(fmt.Sprintf("   • Воин доспех: %s Хауберк (%s:3) %s%s Кираса ярости %s%s Чешуйчатый доспех %s%s Нагрудник витязя (%s:9, %s:+32)\n\n",
			cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))

		sb.WriteString(cSub.Render(" [КОЖЕВНИК: РОГА, МАГ, ЖРЕЦ - Кожа, мантии и пошив]") + "\n")
		sb.WriteString(fmt.Sprintf("   • Рога:        %s Охотничьи ножи %s%s Парные стилеты %s%s Кинжалы %s%s Воровские кортики (%s:10, Крит:8)\n",
			cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
		sb.WriteString(fmt.Sprintf("   • Маг мантия:  %s Роба ученика %s%s Мантия чародея %s%s Одеяние эфира %s%s Астральная мантия (%s:8, MP:+45)\n",
			cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
		sb.WriteString(fmt.Sprintf("   • Жрец:        %s Окованная дубина %s%s Боевой молот %s%s Шестопёр %s%s Булава света (%s:10, MP:+26)\n\n",
			cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))

		sb.WriteString(cSec.Render("ПРОГРЕССИЯ УРОВНЕЙ И КЛАССОВЫЙ РОСТ:") + "\n")
		sb.WriteString(" • Опыт монстра делится поровну между всеми выжившими героями.\n")
		sb.WriteString(" • Прирост Танка: +12 HP, +2 MP, +1 Atk, +1 Def каждые 2 ур.\n")
		sb.WriteString(" • Прирост Воина: +8 HP, +3 MP, +2 Atk, +1 Def каждые 3 ур.\n")
		sb.WriteString(" • Прирост Роги: +5 HP, +4 MP, +2 Atk, +1 Скор.\n")
		sb.WriteString(" • Прирост Мага: +4 HP, +8 MP, +3 Atk.\n")
		sb.WriteString(" • Прирост Жреца: +6 HP, +6 MP, +1 Atk.\n\n")

	case 2: // Вкладка 3: Столица и службы
		sb.WriteString(cSec.Render("СТОЛИЧНЫЕ СЛУЖБЫ И ЗАВЕДЕНИЯ:") + "\n")
		sb.WriteString(fmt.Sprintf(" • %s: Точит всё оружие отряда и тяжелые латы Танка/Воина.\n", cSub.Render("Кузница («"+m.TownEst.SmithyName+"»)")))
		sb.WriteString(fmt.Sprintf(" • %s: Выделывает кожу и шёлк (Рога, Маг, Жрец), шьет сумки и ремни.\n", cSub.Render("Кожевник («"+m.TownEst.TanneryName+"»)")))
		sb.WriteString(fmt.Sprintf(" • %s: Ночлег и снятие стресса. При нехватке золота — сеновал (45%% сил).\n", cSub.Render("Таверна («"+m.TownEst.TavernName+"»)")))
		sb.WriteString(fmt.Sprintf(" • %s: Воскрешение павших соратников и очищение от безумия/психозов.\n", cSub.Render("Храм («"+m.TownEst.ChurchName+"»)")))
		sb.WriteString(fmt.Sprintf(" • %s: Сдача выполненных квестов и наём опытных ветеранов.\n\n", cSub.Render("Гильдия («"+m.TownEst.GuildName+"»)")))

		sb.WriteString(cSec.Render("КВОТИРОВАНИЕ БЮДЖЕТА ГОРОДА:") + "\n")
		sb.WriteString(" • Казна отчисляет 10% золота в фонд следующего поколения (Наследие).\n")
		sb.WriteString(" • Остаток делится на восстановление (25%), снаряжение (30%), найм и алхимию.\n\n")

	case 3: // Вкладка 4: Алхимия и Реликвии
		sb.WriteString(cSec.Render("АЛХИМИЧЕСКИЕ МУТАЦИИ:") + "\n")
		sb.WriteString(fmt.Sprintf(" • %s (База 260G): +8 HP, +5 MP, +1 Atk, +1 Def (Универсально)\n", accentStyle.Render("Сыворотка Химеры")))
		sb.WriteString(fmt.Sprintf(" • %s (База 220G): +3 Atk, +2 HP (Приоритет: Рога, Воин, Маг)\n", fireStyle.Render("Эссенция Ярости")))
		sb.WriteString(fmt.Sprintf(" • %s (База 210G): +20 MaxHP (Приоритет: Танк, Воин)\n", healStyle.Render("Кровь Титана")))
		sb.WriteString(fmt.Sprintf(" • %s (База 200G): +14 MaxMP, +1 Atk (Приоритет: Маг, Жрец)\n", fountStyle.Render("Флюид Эфира")))
		sb.WriteString(fmt.Sprintf(" • %s (База 240G): +2 Def, +6 HP (Приоритет: Танк, Жрец)\n\n", healStyle.Render("Эликсир Бастиона")))

		sb.WriteString(cSec.Render("ЛЕГЕНДАРНЫЕ РЕЛИКВИИ:") + "\n")
		sb.WriteString(fmt.Sprintf(" • %s: До +45%% золота, но монстры наносят увеличенный урон.\n", cItem.Render("Компас Алчности")))
		sb.WriteString(fmt.Sprintf(" • %s: При гибели героя живые бойцы получают +3 Atk за каждый уровень реликвии.\n", cItem.Render("Корона Мученика")))
		sb.WriteString(fmt.Sprintf(" • %s: Снижает весь получаемый стресс отряда вплоть до 60%%.\n\n", cItem.Render("Священный Грааль")))
	}

	sb.WriteString(cNote.Render("[Tab / 1-4] Переключение вкладок | [↑/↓/PgUp/PgDn] Прокрутка | [I/Esc] Закрыть"))

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

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("═══ АРСЕНАЛ ОТРЯДА И АЛХИМИЧЕСКИЕ МУТАЦИИ [E] ═══")
	sb.WriteString(header + "\n\n")

	renderSlotInfo := func(slotName string, it *EquipItem) string {
		if it == nil {
			return fmt.Sprintf("   • %-7s: %s\n", slotName, subtleStyle.Render("Пусто"))
		}
		matStr := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("[%s, x%d]", it.Material.Name, it.Material.BonusMult))

		modStr := ""
		if it.Prefix != nil {
			modStr += fmt.Sprintf(" | %s (+%d %s)", it.Prefix.Name, it.Prefix.Bonus, it.Prefix.Element)
		}
		if it.Suffix != nil {
			modStr += fmt.Sprintf(" | %s (+%d %s)", it.Suffix.Name, it.Suffix.Bonus, it.Suffix.Effect)
		}
		if it.SpeedBonus > 0 {
			modStr += fmt.Sprintf(" | +%d Скор", it.SpeedBonus)
		}

		statLabel := "Защ"
		if it.Slot == SlotWeapon {
			statLabel = "Атк"
		}

		upg := ""
		if it.UpgradeLevel > 0 {
			upg = fmt.Sprintf(" +%d", it.UpgradeLevel)
		}

		return fmt.Sprintf("   • %-7s: %s%s %s -> Итого: %s %d%s\n",
			slotName, goldStyle.Render(it.BaseName), upg, matStr, statLabel, it.TotalStat(), subtleStyle.Render(modStr))
	}

	for _, h := range m.Party {
		status := healStyle.Render("[В СТРОЮ]")
		if h.IsDead {
			status = dangerStyle.Render("[МЕРТВ]")
		}

		heroTitle := h.FullName()
		if h.Title != "" {
			heroTitle = titleStyle.Render(heroTitle)
		}

		mutStr := subtleStyle.Render("Нет")
		if h.Mutations.Total() > 0 {
			mutStr = accentStyle.Render(fmt.Sprintf("Всего: %d [Хим:%d Яр:%d Тит:%d Эфир:%d Баст:%d]",
				h.Mutations.Total(), h.Mutations.ChimeraCount, h.Mutations.FuryCount, h.Mutations.TitanCount, h.Mutations.AetherCount, h.Mutations.BastionCount))
		}

		potInfo := "Пуст"
		if len(h.Potions) > 0 {
			var pNames []string
			for _, p := range h.Potions {
				if p != nil {
					pNames = append(pNames, fmt.Sprintf("%s %s", p.Symbol, getPotionName(*p)))
				}
			}
			potInfo = strings.Join(pNames, ", ")
		}

		sb.WriteString(fmt.Sprintf("👤 %s (%s %d, XP:%d/%d, Агро:%d, Скор:%d) %s\n",
			heroTitle, h.ShortClass(), h.Level, h.Exp, h.NextLevelExp(), h.Role.AggroWeight, h.TotalSpeed(), status))
		sb.WriteString(fmt.Sprintf("   🧪 Мутации: %s\n", mutStr))
		sb.WriteString(fmt.Sprintf("   🎒 Пояс зелий: %s (Слотов: %d/%d)\n", potionStyle.Render(potInfo), len(h.Potions), m.MaxPotionSlots()))
		sb.WriteString(renderSlotInfo("Оружие", h.Weapon))
		sb.WriteString(renderSlotInfo("Шлем", h.Head))
		sb.WriteString(renderSlotInfo("Доспех", h.Chest))
		sb.WriteString(renderSlotInfo("Поножи", h.Legs))
		sb.WriteString("\n")
	}

	sb.WriteString(subtleStyle.Render("[↑/↓/PgUp/PgDn] Прокрутка  |  [E/Esc] Закрыть арсенал  |  [Space] Пауза  |  [Q] Выход"))

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
			lvlStr = fmt.Sprintf(" (Ур.%d)", lvl)
		}

		label := fmt.Sprintf("%s %s%s", icon, name, lvlStr)
		if isActive {
			return cActive.Render(fmt.Sprintf("►[%s]◄", label))
		}
		return cIdle.Render(fmt.Sprintf(" [%s] ", label))
	}

	bMarket := fmtBld(TownPhaseSellLoot, "⚖️", "Рынок", 0)
	bMagistrate := fmtBld(TownPhaseMagistrate, "🏛️", "Магистрат", 0)
	bChurch := fmtBld(TownPhaseChurch, "⛪", m.TownEst.ChurchName, m.Legacy.ChurchLevel)
	bTavern := fmtBld(TownPhaseTavern, "🍻", m.TownEst.TavernName, m.Legacy.TavernLevel)
	bGuild := fmtBld(TownPhaseGuild, "⚔️", m.TownEst.GuildName, 0)
	bSmithy := fmtBld(TownPhaseSmithy, "⚒️", m.TownEst.SmithyName, m.Legacy.SmithyLevel)
	bTannery := fmtBld(TownPhaseTannery, "🎒", m.TownEst.TanneryName, m.Legacy.TanneryLevel)
	bAlchemist := fmtBld(TownPhaseAlchemist, "🧪", m.TownEst.AlchemistName, 0)

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
	sb.WriteString(center(cHeader.Render("═══ СТОЛИЧНЫЙ КВАРТАЛ И СЛУЖБЫ ═══")) + "\n\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s", bMarket, bMagistrate)) + "\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s  %s", bChurch, bTavern, bGuild)) + "\n")
	sb.WriteString(center(fmt.Sprintf("%s  %s  %s", bSmithy, bTannery, bAlchemist)) + "\n\n")

	sb.WriteString(cBorder.Render(strings.Repeat("─", maxW)) + "\n")
	sb.WriteString(questStyle.Render(" ЖУРНАЛ ДЕЙСТВИЙ ОТРЯДА В ГОРОДЕ:\n"))

	if len(m.TownHistory) == 0 {
		sb.WriteString(subtleStyle.Render("   • Отряд проходит через городские ворота на привал...\n"))
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

	nameRaw := h.FullName()
	nameRunes := []rune(nameRaw)
	if len(nameRunes) > innerWidth {
		nameRaw = string(nameRunes[:innerWidth-1]) + "…"
	}
	nameStr := nameRaw
	if h.Title != "" {
		nameStr = titleStyle.Render(nameRaw)
	}

	classFormatted := subtleStyle.Render(fmt.Sprintf("(%s %d)", h.ShortClass(), h.Level))

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
		sb.WriteString(fmt.Sprintf("%s %s\n", classFormatted, dangerStyle.Render("[☠️ ПАЛ]")))
		sb.WriteString(subtleStyle.Render(shortenItemName(h.CauseOfDeath, innerWidth)) + "\n")
		sb.WriteString(fmt.Sprintf("Этаж: %d\n", m.Floor))
		sb.WriteString(subtleStyle.Render("[Ждет воскрешения]"))
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
		sb.WriteString(fmt.Sprintf("СТ:%s%3d ⚔%-2d 🛡%-2d\n", stressBar, h.Stress, h.TotalAtk(), h.TotalDef()))

		slotLen := max(4, (innerWidth-6)/2)
		formatEquipSlot := func(item *EquipItem, maxLen int) string {
			if item == nil {
				return "-"
			}
			name := shortenItemName(item.BaseName, maxLen)
			if item.UpgradeLevel > 0 {
				return fmt.Sprintf("+%d%s", item.UpgradeLevel, shortenItemName(item.BaseName, maxLen-2))
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
	sb.WriteString(goldStyle.Render("👑 ОТРЯД И РЕЛИКВИЯ") + "\n")
	relicName := "Нет"
	relicDesc := "Реликвия не найдена"
	if m.Relic != nil {
		relicName = m.Relic.Name
		relicDesc = m.Relic.Description
	}
	sb.WriteString(fmt.Sprintf("Реликвия: %s\n", shortenItemName(relicName, innerWidth-10)))
	sb.WriteString(subtleStyle.Render(shortenItemName(relicDesc, innerWidth)) + "\n")
	sb.WriteString(fmt.Sprintf("🎒 Мешки: %d/%d слотов\n", len(m.Bag), m.currentBagCapacity()))
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)
	sb.WriteString(healStyle.Render(fmt.Sprintf("🏛️ Наследие: +%dG", legacyPart)))

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
			barrelNotice = " [🛢️ ПОРОХ]"
		}
		if m.Combat.Round > 20 {
			barrelNotice += dangerStyle.Render(fmt.Sprintf(" [ЯРОСТЬ R%d]", m.Combat.Round))
		}
		lines = append(lines, dangerStyle.Render(shortenItemName("ВРАЖЕСКАЯ СТАЯ"+barrelNotice+":", innerRightW-2)))
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
				stText = "УБИТ"
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
			mobName := shortenItemName(mob.Name, availName)
			namePadded := padRight(mobName, availName)

			lines = append(lines, fmt.Sprintf("%s%s %s %s", prefix, namePadded, mobBar, stBadge))
			displayed++
		}

		if len(m.Combat.Pack.Members) > displayed && len(lines) < maxLines {
			lines = append(lines, subtleStyle.Render(fmt.Sprintf("... и ещё %d", len(m.Combat.Pack.Members)-displayed)))
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render("СТОЛИЧНОЕ УПРАВЛЕНИЕ:"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("Кузня Ур.%-2d   | Кожевник Ур.%-2d", m.Legacy.SmithyLevel, m.Legacy.TanneryLevel))
		lines = append(lines, fmt.Sprintf("Храм  Ур.%-2d   | Таверна  Ур.%-2d", m.Legacy.ChurchLevel, m.Legacy.TavernLevel))
		lines = append(lines, fmt.Sprintf("Пояс: %d слота  | Сумка: %d сл.", m.MaxPotionSlots(), m.currentBagCapacity()))
		lines = append(lines, subtleStyle.Render(shortenItemName("Квотирование казны активно.", innerRightW-2)))
	} else {
		lines = append(lines, accentStyle.Render("РАЗВЕДКА ТЕРРИТОРИИ:"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("Сделано шагов:   %d", m.Stats.TotalSteps))
		lines = append(lines, fmt.Sprintf("Сундуков найдено: %d", m.Stats.ChestsOpened))
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
		return m.renderStatsScreen("ПОРАЖЕНИЕ: ВЕСЬ ОТРЯД ПАЛ ВО ТЬМЕ...", lipgloss.Color("196"))
	}
	if m.State == StateStatsManual {
		return m.renderStatsScreen("ЭКИПИРОВКА, НАСЛЕДИЕ И КНИГА ПАМЯТИ", lipgloss.Color("214"))
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
	floorTag := fmt.Sprintf("ЭТАЖ %d: %s", m.Floor, biome.Name)
	if m.InTown {
		floorTag = fmt.Sprintf("ЭТАЖ %d: [Лагерь]", m.Floor)
	}

	questTitleShort := shortenItemName(m.CurrentQuest.Title, innerRightW-12)
	var statusBadge string
	if m.CurrentQuest.Completed {
		statusBadge = healStyle.Render("[СДАТЬ]")
	} else {
		statusBadge = subtleStyle.Render(fmt.Sprintf("[%d/%d]", m.CurrentQuest.Current, m.CurrentQuest.TargetCount))
	}
	questLine := padRight(fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge), innerRightW-4)

	var rightPane string
	if availableTopH >= 13 {
		floorBoxContent := fmt.Sprintf(
			"🏰 %s\n%s",
			titleStyle.Render(shortenItemName(floorTag, innerRightW-4)),
			subtleStyle.Render(shortenItemName("Опасность: "+biome.EnvHazard, innerRightW-4)),
		)
		floorBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Width(innerRightW).
			Height(2).
			Render(floorBoxContent)

		statusContent := fmt.Sprintf(
			"💰 Казна: %s\n%s",
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
		lines = append(lines, fmt.Sprintf("🏰 %s | 💰 %s", titleStyle.Render(shortenItemName(floorTag, 14)), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
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
		scrollInfo = fmt.Sprintf(" (Архив: -%d)", m.LogScroll)
	}
	logs.WriteString(lipgloss.NewStyle().Bold(true).Render("ХРОНИКИ ЭКСПЕДИЦИИ" + scrollInfo + ":\n"))

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

	controls := subtleStyle.Render("[Space] Пауза  |  [↑/↓] Логи  |  [F] Побег  |  [Tab] Вкладки [I]  |  [+/-] Скор.  |  [E] Арсенал  |  [I] Кодекс  |  [S] Слава  |  [Q] Выход")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		strings.TrimRight(logs.String(), "\n"),
		controls,
	)
}
