package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func typeText(m *Model, text string) {
	for _, r := range text {
		press(m, string(r))
	}
}

func gotoRow(m *Model, row int) {
	for i := 0; i < creatorRows*2 && m.Creator.Row != row; i++ {
		press(m, "down")
	}
}

func menuModel(t *testing.T) *Model {
	t.Helper()
	useTempSave(t)
	seedRNG(31)
	m := initialModel()
	return &m
}

func TestMenuOpensCreatorAndInvalidatesTimer(t *testing.T) {
	m := menuModel(t)
	gen := m.MenuGen
	press(m, "c")
	if m.State != StateCreator {
		t.Fatal("C в меню должна открывать создание героя")
	}
	if m.MenuGen == gen {
		t.Fatal("вход в создание должен обесценивать таймер меню")
	}

	// Тик старого таймера не должен запустить экспедицию прямо из редактора.
	m.MenuCountdown = 1
	next, _ := m.Update(MenuTickMsg{Gen: gen})
	*m = next.(Model)
	if m.State != StateCreator || m.MenuCountdown != 1 {
		t.Fatal("устаревший тик меню должен игнорироваться")
	}
}

func TestCreatorTypingIsTextNotHotkeys(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	typeText(m, "qlcxsmi tr")
	if m.State != StateCreator {
		t.Fatal("буквы в поле имени не должны срабатывать как горячие клавиши")
	}
	if got := m.Creator.Draft.Name; got != "qlcxsmi tr" {
		t.Fatalf("имя %q", got)
	}
	if m.Lang != LangRU {
		t.Fatal("клавиша L не должна менять язык внутри поля имени")
	}

	// Вставленное слово не должно распознаться как команда.
	m.Creator.Draft.Name = ""
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("esc")})
	*m = next.(Model)
	if m.State != StateCreator || m.Creator.Draft.Name != "esc" {
		t.Fatal("вставленный текст «esc» должен стать именем, а не командой")
	}
}

func TestCreatorNameEditing(t *testing.T) {
	m := menuModel(t)
	press(m, "c")

	typeText(m, "  Ann")
	if m.Creator.Draft.Name != "Ann" {
		t.Fatalf("ведущие пробелы должны отбрасываться: %q", m.Creator.Draft.Name)
	}
	typeText(m, "  Lee")
	if m.Creator.Draft.Name != "Ann Lee" {
		t.Fatalf("двойные пробелы должны схлопываться: %q", m.Creator.Draft.Name)
	}
	typeText(m, "%<>\"")
	if m.Creator.Draft.Name != "Ann Lee" {
		t.Fatalf("запрещённые символы должны игнорироваться: %q", m.Creator.Draft.Name)
	}
	press(m, "backspace")
	press(m, "backspace")
	if m.Creator.Draft.Name != "Ann L" {
		t.Fatalf("backspace: %q", m.Creator.Draft.Name)
	}
	typeText(m, strings.Repeat("x", 40))
	if n := len([]rune(m.Creator.Draft.Name)); n != creatorNameMax {
		t.Fatalf("длина имени %d, максимум %d", n, creatorNameMax)
	}
	typeText(m, "Ёё")
	if len([]rune(m.Creator.Draft.Name)) != creatorNameMax {
		t.Fatal("после лимита ввод должен блокироваться")
	}

	m.Creator.Draft.Name = ""
	for i := 0; i < 5; i++ {
		press(m, "backspace")
	}
	typeText(m, "Эдан")
	if m.Creator.Draft.Name != "Эдан" {
		t.Fatalf("кириллица должна вводиться: %q", m.Creator.Draft.Name)
	}
	press(m, "backspace")
	if m.Creator.Draft.Name != "Эда" {
		t.Fatalf("backspace должен удалять символ, а не байт: %q", m.Creator.Draft.Name)
	}
}

func TestCreatorFullFlow(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	typeText(m, "Эдан")

	gotoRow(m, rowGender)
	press(m, "right")
	gotoRow(m, rowRace)
	press(m, "right") // человек -> эльф
	gotoRow(m, rowClass)
	for m.Creator.Draft.Class != ClassMage {
		press(m, "right")
	}
	gotoRow(m, rowCalling)
	press(m, "right")

	gotoRow(m, rowStatFirst+StatAtk)
	press(m, "+")
	press(m, "right")
	gotoRow(m, rowStatFirst+StatHP)
	press(m, "+")
	press(m, "+")
	press(m, "-")

	d := m.Creator.Draft
	if d.Gender != GenderFemale || d.Race != RaceElf || d.Class != ClassMage || d.Calling != CallingStrategist {
		t.Fatalf("выбор не сохранился: %+v", d)
	}
	if d.Points[StatAtk] != 2 || d.Points[StatHP] != 1 || d.PointsLeft() != creatorPoints-3 {
		t.Fatalf("очки распределены неверно: %+v", d.Points)
	}

	gotoRow(m, rowConfirm)
	press(m, "enter")

	if m.State != StateMenu {
		t.Fatalf("после подтверждения нужно вернуться в меню, состояние %v", m.State)
	}
	if m.MenuCountdown != menuCountdownStart {
		t.Fatalf("таймер меню должен начаться заново, а не %d", m.MenuCountdown)
	}
	if m.Blueprint == nil || m.Blueprint.Name != "Эдан" {
		t.Fatalf("чертёж не сохранён: %+v", m.Blueprint)
	}
	lead := m.Party[0]
	if len(m.Party) != 5 || !lead.IsLeader || lead.DisplayName(LangRU) != "Эдан" || lead.Class != ClassMage || lead.Race != RaceElf {
		t.Fatalf("отряд не пересобран вокруг лидера: %+v", lead)
	}

	sd, ok := loadSave()
	if !ok || sd.Hero == nil || sd.Hero.Name != "Эдан" {
		t.Fatalf("герой должен сохраняться на диск сразу: %+v", sd.Hero)
	}

	// Тик нового таймера меню снова работает.
	before := m.MenuCountdown
	next, _ := m.Update(MenuTickMsg{Gen: m.MenuGen})
	*m = next.(Model)
	if m.MenuCountdown != before-1 {
		t.Fatal("после выхода из создания таймер меню должен идти")
	}
}

func TestCreatorRejectsShortName(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	typeText(m, "A")
	gotoRow(m, rowConfirm)
	press(m, "enter")

	if m.State != StateCreator || m.Blueprint != nil {
		t.Fatal("слишком короткое имя не должно приниматься")
	}
	if m.Creator.Err == "" || m.Creator.Row != rowName {
		t.Fatalf("игрок должен увидеть ошибку и вернуться к полю имени (err=%q row=%d)", m.Creator.Err, m.Creator.Row)
	}
	for _, lang := range []Language{LangRU, LangEN} {
		m.Lang = lang
		if !strings.Contains(m.View(), T(lang, m.Creator.Err)) {
			t.Fatalf("%s: ошибка должна отображаться", lang)
		}
	}
	press(m, "x")
	if m.Creator.Err != "" {
		t.Fatal("ошибка должна пропадать при следующем вводе")
	}
}

func TestCreatorEscapeKeepsEverything(t *testing.T) {
	m := menuModel(t)
	old := m.Party
	press(m, "c")
	typeText(m, "Борис")
	m.MenuCountdown = 3
	press(m, "esc")

	if m.State != StateMenu || m.MenuCountdown != menuCountdownStart {
		t.Fatal("Esc возвращает в меню и перезапускает таймер")
	}
	if m.Blueprint != nil {
		t.Fatal("Esc не должна сохранять героя")
	}
	if len(m.Party) != len(old) || m.Party[0] != old[0] {
		t.Fatal("Esc не должна пересоздавать отряд")
	}
}

func TestCreatorEditsExistingHero(t *testing.T) {
	m := menuModel(t)
	bp := HeroBlueprint{Name: "Рагна", Gender: GenderFemale, Race: RaceBeastman, Class: ClassRogue, Calling: CallingMentor}
	bp.AdjustPoint(StatSpd, 1)
	m.Blueprint = &bp

	press(m, "c")
	if m.Creator.Draft.Name != "Рагна" || m.Creator.Draft.Class != ClassRogue || m.Creator.Draft.Points[StatSpd] != 1 {
		t.Fatalf("при правке должны подставляться текущие значения: %+v", m.Creator.Draft)
	}
	typeText(m, "!") // запрещённый символ
	typeText(m, "я")
	if m.Creator.Draft.Name != "Рагная" {
		t.Fatalf("имя дополняется: %q", m.Creator.Draft.Name)
	}
	if m.Blueprint.Name != "Рагна" {
		t.Fatal("пока герой не подтверждён, прежний чертёж не меняется")
	}
}

func TestCreatorRandomAndLanguage(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	press(m, "ctrl+r")
	if !m.Creator.Draft.Valid() || m.Creator.Draft.PointsLeft() != 0 {
		t.Fatalf("Ctrl+R должна давать готового героя: %+v", m.Creator.Draft)
	}
	press(m, "ctrl+l")
	if m.Lang != LangEN || m.State != StateCreator {
		t.Fatal("Ctrl+L переключает язык, не покидая редактор")
	}
}

func TestMenuRemoveHeroNeedsConfirmation(t *testing.T) {
	m := menuModel(t)
	bp := HeroBlueprint{Name: "Рагна", Gender: GenderFemale, Race: RaceElf, Class: ClassMage, Calling: CallingInspirer}
	m.Blueprint = &bp
	m.Party = buildStartingParty(0, m.Blueprint)

	press(m, "x")
	if m.Blueprint == nil || !m.MenuConfirmRemove {
		t.Fatal("первое X только просит подтверждения")
	}
	if !strings.Contains(m.View(), T(m.Lang, "menu.keys_confirm")) {
		t.Fatal("меню должно показывать запрос подтверждения")
	}
	press(m, "p") // любая другая клавиша отменяет
	if m.MenuConfirmRemove || m.Blueprint == nil {
		t.Fatal("другая клавиша должна отменять удаление")
	}

	press(m, "x")
	press(m, "x")
	if m.Blueprint != nil {
		t.Fatal("второе X должно убрать героя")
	}
	for _, h := range m.Party {
		if h.IsLeader {
			t.Fatal("после удаления отряд снова случайный")
		}
	}
	sd, ok := loadSave()
	if !ok || sd.Hero != nil {
		t.Fatal("удаление должно записываться в сохранение")
	}

	press(m, "x") // без героя X ничего не делает
	if m.MenuConfirmRemove {
		t.Fatal("без героя подтверждение не нужно")
	}
}

func TestRestartKeepsLeader(t *testing.T) {
	useTempSave(t)
	seedRNG(41)
	bp := HeroBlueprint{Name: "Эдан", Gender: GenderMale, Race: RaceOlongr, Class: ClassTank, Calling: CallingInspirer}
	m := initialModelWith(TownLegacy{}, &bp)
	m.State = StateDefeat
	m.Party[0].Level = 9 // прогресс забега не переносится

	fresh, cmd := resetGameStatic(m)
	if cmd == nil || fresh.State != StateMenu {
		t.Fatal("после рестарта должно быть меню")
	}
	if fresh.Blueprint == nil || fresh.Blueprint.Name != "Эдан" {
		t.Fatal("чертёж героя должен переживать рестарт")
	}
	lead := fresh.Party[0]
	if !lead.IsLeader || lead.Level != 1 || lead.DisplayName(LangRU) != "Эдан" || lead.Class != ClassTank {
		t.Fatalf("лидер должен быть пересоздан с нуля: %+v", lead)
	}
	if fresh.MenuGen == m.MenuGen {
		t.Fatal("рестарт должен менять поколение таймера меню")
	}
}

func TestCreatorAndMenuScreensRender(t *testing.T) {
	useTempSave(t)
	seedRNG(51)
	sizes := [][2]int{{38, 20}, {60, 24}, {80, 30}, {120, 40}, {160, 50}}
	for _, lang := range []Language{LangRU, LangEN} {
		for _, sz := range sizes {
			m := initialModel()
			m.Lang = lang
			m.TermWidth, m.TermHeight = sz[0], sz[1]

			if m.View() == "" {
				t.Fatalf("%s %v: меню без героя не отрисовалось", lang, sz)
			}

			bp := HeroBlueprint{Name: "Эдан", Gender: GenderMale, Race: RaceElf, Class: ClassMage, Calling: CallingMentor}
			m.Blueprint = &bp
			if v := m.View(); !strings.Contains(v, "Эдан") {
				t.Fatalf("%s %v: меню должно показывать созданного героя", lang, sz)
			}

			m.openCreator()
			m.Creator.Draft = bp
			v := m.View()
			if !strings.Contains(v, "Эдан") {
				t.Fatalf("%s %v: экран создания должен показывать имя", lang, sz)
			}
			for row := 0; row < creatorRows; row++ {
				m.Creator.Row = row
				if m.View() == "" {
					t.Fatalf("%s %v: строка %d не отрисовалась", lang, sz, row)
				}
			}
		}
	}
}

func TestLeaderMarkOnHeroCard(t *testing.T) {
	seedRNG(61)
	bp := HeroBlueprint{Name: "Эдан", Gender: GenderMale, Race: RaceElf, Class: ClassMage, Calling: CallingMentor}
	m := initialModelWith(TownLegacy{}, &bp)
	m.State = StatePlaying
	m.TermWidth, m.TermHeight = 140, 50

	if card := m.renderHeroCard(m.Party[0], 30, CardWide); !strings.Contains(card, "★") {
		t.Fatal("карточка лидера должна быть помечена звёздочкой")
	}
	if card := m.renderHeroCard(m.Party[1], 30, CardWide); strings.Contains(card, "★") {
		t.Fatal("у обычного героя метки лидера быть не должно")
	}
}
