package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// ЭКРАН СОЗДАНИЯ ГЕРОЯ-ЛИДЕРА
// ============================================================

const (
	rowName = iota
	rowGender
	rowRace
	rowClass
	rowCalling
	rowStatFirst // строки rowStatFirst … rowStatFirst+statCount-1 — очки характеристик
	rowConfirm   = rowStatFirst + statCount
	creatorRows  = rowConfirm + 1
)

const menuCountdownStart = 20

// CreatorState — состояние экрана создания героя.
type CreatorState struct {
	Draft HeroBlueprint // редактируемый чертёж
	Row   int           // выбранная строка
	Err   string        // ключ словаря с текстом ошибки (пусто — ошибки нет)
}

func cycle[E comparable](list []E, cur E, dir int) E {
	idx := 0
	for i, v := range list {
		if v == cur {
			idx = i
			break
		}
	}
	return list[(idx+dir+len(list))%len(list)]
}

// openCreator открывает экран; существующий герой подставляется для правки.
func (m *Model) openCreator() {
	draft := defaultBlueprint()
	if m.Blueprint != nil {
		draft = *m.Blueprint
	}
	m.Creator = CreatorState{Draft: draft}
	m.State = StateCreator
	m.MenuGen++ // отложенный тик меню устарел и не должен сработать
}

// closeCreator возвращает в главное меню и заново запускает его таймер.
func (m *Model) closeCreator() tea.Cmd {
	m.State = StateMenu
	m.MenuCountdown = menuCountdownStart
	m.MenuGen++
	return menuTickCmd(m.MenuGen)
}

func (m *Model) confirmCreator() tea.Cmd {
	bp := m.Creator.Draft.Sanitized()
	if !bp.Valid() {
		m.Creator.Err = "creator.err.name"
		m.Creator.Row = rowName
		return nil
	}
	m.Blueprint = &bp
	m.Party = buildStartingParty(m.Legacy.SmithyLevel, m.Blueprint)
	persistState(m.Lang, m.Legacy, m.Tactics, m.Blueprint)
	return m.closeCreator()
}

// removeHero убирает сохранённого героя: следующие забеги снова идут со случайным отрядом.
func (m *Model) removeHero() {
	if m.Blueprint == nil {
		return
	}
	m.Blueprint = nil
	m.Party = buildStartingParty(m.Legacy.SmithyLevel, nil)
	persistState(m.Lang, m.Legacy, m.Tactics, nil)
}

// handleMenuKey — клавиши главного меню, связанные с героем. Второе значение: «обработана».
// Удаление героя требует повторного нажатия X, любая другая клавиша отменяет его.
func (m *Model) handleMenuKey(key string) (tea.Cmd, bool) {
	if key != "x" {
		m.MenuConfirmRemove = false
	}
	switch key {
	case "c":
		m.openCreator()
		return nil, true
	case "x":
		if m.Blueprint == nil {
			return nil, true
		}
		if !m.MenuConfirmRemove {
			m.MenuConfirmRemove = true
			return nil, true
		}
		m.MenuConfirmRemove = false
		m.removeHero()
		return nil, true
	}
	return nil, false
}

// creatorAdjust меняет значение выбранной строки (dir = +1 / -1).
func (m *Model) creatorAdjust(dir int) {
	c := &m.Creator
	d := &c.Draft
	switch {
	case c.Row == rowGender:
		d.Gender = 1 - d.Gender
	case c.Row == rowRace:
		d.Race = cycle(AllRaces, d.Race, dir)
	case c.Row == rowClass:
		d.Class = cycle(AllClasses, d.Class, dir)
	case c.Row == rowCalling:
		d.Calling = cycle(AllCallings, d.Calling, dir)
	case c.Row >= rowStatFirst && c.Row < rowConfirm:
		d.AdjustPoint(c.Row-rowStatFirst, dir)
	}
}

// creatorTypeRune обрабатывает введённый символ: текст имени или «+»/«-» на строках очков.
func (m *Model) creatorTypeRune(r rune) {
	c := &m.Creator
	switch {
	case c.Row == rowName:
		if !allowedNameRune(r) {
			return
		}
		name := []rune(c.Draft.Name)
		if len(name) >= creatorNameMax {
			return
		}
		if r == ' ' && (len(name) == 0 || name[len(name)-1] == ' ') {
			return
		}
		c.Draft.Name += string(r)
	case c.Row >= rowStatFirst && c.Row < rowConfirm:
		stat := c.Row - rowStatFirst
		switch r {
		case '+', '=':
			c.Draft.AdjustPoint(stat, +1)
		case '-', '_':
			c.Draft.AdjustPoint(stat, -1)
		}
	}
}

// handleCreatorKey перехватывает все клавиши, пока открыт экран создания:
// буквы здесь — это текст имени, а не горячие клавиши игры.
func (m *Model) handleCreatorKey(msg tea.KeyMsg) tea.Cmd {
	c := &m.Creator
	c.Err = ""

	// Печатные символы — всегда текст, даже если пришла пачка из нескольких рун:
	// вставленное слово «esc» или «up» не должно сработать как команда.
	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		if msg.Alt {
			return nil
		}
		runes := msg.Runes
		if msg.Type == tea.KeySpace && len(runes) == 0 {
			runes = []rune{' '}
		}
		for _, r := range runes {
			m.creatorTypeRune(r)
		}
		return nil
	}

	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		return m.closeCreator()
	case "ctrl+r":
		c.Draft = randomBlueprint(m.Lang)
		return nil
	case "ctrl+l":
		if m.Lang == LangRU {
			m.Lang = LangEN
		} else {
			m.Lang = LangRU
		}
		return nil
	case "up", "shift+tab":
		c.Row = (c.Row + creatorRows - 1) % creatorRows
		return nil
	case "down", "tab":
		c.Row = (c.Row + 1) % creatorRows
		return nil
	case "enter":
		if c.Row == rowConfirm {
			return m.confirmCreator()
		}
		c.Row++
		return nil
	case "left":
		m.creatorAdjust(-1)
		return nil
	case "right":
		m.creatorAdjust(+1)
		return nil
	case "backspace":
		if c.Row == rowName {
			r := []rune(c.Draft.Name)
			if len(r) > 0 {
				c.Draft.Name = string(r[:len(r)-1])
			}
		}
		return nil
	}

	return nil
}

// callingValue — числовой бонус призвания (для подстановки в описание).
func callingValue(c LeaderCalling) int {
	switch c {
	case CallingStrategist:
		return callingStrategistGoldPct
	case CallingMentor:
		return callingMentorExpPct
	}
	return callingInspirerStressCut
}

func (m Model) raceLabel(r RaceType) string {
	return T(m.Lang, "race."+string(r)+".name")
}

func (m Model) classLabel(c HeroClass) string {
	return T(m.Lang, "class."+string(c)+".name")
}

func (m Model) callingLabel(c LeaderCalling) string {
	return T(m.Lang, "calling."+string(c)+".name")
}

func (m Model) genderLabel(g Gender) string {
	if g == GenderFemale {
		return T(m.Lang, "creator.gender.female")
	}
	return T(m.Lang, "creator.gender.male")
}

// menuHeroLine — строка главного меню о созданном герое (пусто, если героя нет).
func (m Model) menuHeroLine() string {
	if m.Blueprint == nil {
		return ""
	}
	bp := *m.Blueprint
	return healStyle.Render(T(m.Lang, "menu.hero_set", bp.Name, m.raceLabel(bp.Race), m.classLabel(bp.Class), m.callingLabel(bp.Calling)))
}

// menuHeroKeys — подсказка по клавишам героя для главного меню.
func (m Model) menuHeroKeys() string {
	switch {
	case m.Blueprint == nil:
		return accentStyle.Render(T(m.Lang, "menu.keys_create"))
	case m.MenuConfirmRemove:
		return dangerStyle.Render(T(m.Lang, "menu.keys_confirm"))
	}
	return accentStyle.Render(T(m.Lang, "menu.keys_edit"))
}

func (m Model) renderCreatorScreen() string {
	c := m.Creator
	d := c.Draft
	prev := d.Preview(m.Legacy.SmithyLevel)
	boxW := max(40, m.TermWidth-4)
	innerW := max(30, boxW-6)
	tight := m.TermHeight < 30
	gap := "\n" // пустая строка между блоками; в тесном окне убирается
	if tight {
		gap = ""
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(T(m.Lang, "creator.title")) + "\n")
	if !tight {
		sb.WriteString(subtleStyle.Render(T(m.Lang, "creator.subtitle")) + "\n")
	}
	sb.WriteString(gap)

	row := func(idx int, label, value string) {
		marker := "  "
		if c.Row == idx {
			marker = accentStyle.Render("▶ ")
		}
		sb.WriteString(marker + fmt.Sprintf("%-12s", label) + " " + value + "\n")
	}
	arrows := func(text string) string { return "◀ " + text + " ▶" }

	cursor := ""
	if c.Row == rowName {
		cursor = "_"
	}
	nameBox := fmt.Sprintf("[ %-*s ]", creatorNameMax+1, d.Name+cursor)
	row(rowName, T(m.Lang, "creator.field.name"), nameBox)
	row(rowGender, T(m.Lang, "creator.field.gender"), arrows(m.genderLabel(d.Gender)))
	row(rowRace, T(m.Lang, "creator.field.race"),
		arrows(m.raceLabel(d.Race))+"  "+subtleStyle.Render(T(m.Lang, "creator.race."+string(d.Race))))
	row(rowClass, T(m.Lang, "creator.field.class"),
		arrows(m.classLabel(d.Class))+"  "+subtleStyle.Render(T(m.Lang, prev.SkillKey)))
	row(rowCalling, T(m.Lang, "creator.field.calling"),
		arrows(m.callingLabel(d.Calling))+"  "+subtleStyle.Render(T(m.Lang, "calling."+string(d.Calling)+".desc", callingValue(d.Calling))))

	pointsLine := T(m.Lang, "creator.points", d.PointsLeft(), creatorPoints)
	if d.PointsLeft() > 0 {
		pointsLine = accentStyle.Render(pointsLine)
	} else {
		pointsLine = subtleStyle.Render(pointsLine)
	}
	sb.WriteString("  " + pointsLine + "\n")

	statValues := [statCount]int{prev.HP, prev.MP, prev.Atk, prev.Def, prev.Spd}
	for i := 0; i < statCount; i++ {
		spec := statSpecs[i]
		bonus := ""
		if d.Points[i] > 0 {
			bonus = fmt.Sprintf(" (+%d)", d.Points[i]*spec.perPoint)
		}
		row(rowStatFirst+i, T(m.Lang, spec.labelKey),
			fmt.Sprintf("◀ %d/%d ▶  %d%s", d.Points[i], spec.maxPoints, statValues[i], bonus))
	}

	confirm := "[ " + T(m.Lang, "creator.confirm") + " ]"
	if c.Row == rowConfirm {
		confirm = healStyle.Render(confirm)
	} else {
		confirm = subtleStyle.Render(confirm)
	}
	marker := "  "
	if c.Row == rowConfirm {
		marker = accentStyle.Render("▶ ")
	}
	sb.WriteString(gap + marker + confirm + "\n")

	role := T(m.Lang, "creator.role", prev.Role.AggroWeight)
	if prev.Role.CanGuard {
		role += ", " + T(m.Lang, "creator.flag.guard")
	}
	if prev.Role.CanHeal {
		role += ", " + T(m.Lang, "creator.flag.heal")
	}
	sb.WriteString(gap + T(m.Lang, "creator.skill", T(m.Lang, prev.SkillKey), prev.SkillCost) + "\n")
	sb.WriteString(role + "\n")

	if !tight {
		rules := lipgloss.NewStyle().Width(innerW).Render(T(m.Lang, "creator.rules"))
		sb.WriteString("\n" + subtleStyle.Render(rules) + "\n")
	}
	if c.Err != "" {
		sb.WriteString("\n" + dangerStyle.Render(T(m.Lang, c.Err)) + "\n")
	}
	sb.WriteString(gap + subtleStyle.Render(T(m.Lang, "creator.controls")))

	box := statsBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(max(38, m.TermWidth), max(20, m.TermHeight), lipgloss.Center, lipgloss.Center, box)
}
