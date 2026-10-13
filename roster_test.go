package main

import "testing"

func mkBP(name string, class HeroClass) HeroBlueprint {
	b := defaultBlueprint()
	b.Name, b.Class = name, class
	return b
}

func TestCreateCustomHeroLeaderAndCompanion(t *testing.T) {
	seedRNG(1)
	b := HeroBlueprint{Name: "Эдан", Gender: GenderFemale, Race: RaceElf, Class: ClassMage, Calling: CallingMentor}
	b.AdjustPoint(StatAtk, 2)

	leader := createCustomHero(b, 0, true)
	comp := createCustomHero(b, 0, false)

	if !leader.IsLeader || leader.Calling != CallingMentor {
		t.Error("лидер получает призвание")
	}
	if comp.IsLeader || comp.Calling != "" {
		t.Errorf("спутник не лидер и без призвания: %+v", comp.Calling)
	}
	for _, h := range []*Hero{leader, comp} {
		if h.DisplayName(LangRU) != "Эдан" || h.Class != ClassMage || h.Race != RaceElf || h.Gender != GenderFemale {
			t.Errorf("параметры героя не совпали: %+v", h)
		}
		if h.BaseAtk != comp.BaseAtk {
			t.Error("очки характеристик должны применяться одинаково")
		}
	}
	if createLeaderHero(b, 0).IsLeader != true {
		t.Error("createLeaderHero создаёт лидера")
	}
}

func TestBuildStartingPartyWithCompanions(t *testing.T) {
	for seed := int64(1); seed <= 25; seed++ {
		seedRNG(seed)
		leader := mkBP("Эдан", ClassTank)
		party := buildStartingParty(0, &leader, mkBP("Рагна", ClassMage), mkBP("Борис", ClassRogue))

		if len(party) != partySize {
			t.Fatalf("seed %d: героев %d", seed, len(party))
		}
		if !party[0].IsLeader || party[0].DisplayName(LangRU) != "Эдан" {
			t.Fatalf("seed %d: первым идёт лидер", seed)
		}
		if party[1].DisplayName(LangRU) != "Рагна" || party[1].Class != ClassMage || party[1].IsLeader {
			t.Fatalf("seed %d: вторым идёт первый спутник", seed)
		}
		if party[2].DisplayName(LangRU) != "Борис" || party[2].Class != ClassRogue {
			t.Fatalf("seed %d: третьим идёт второй спутник", seed)
		}
		used := map[HeroClass]bool{ClassTank: true, ClassMage: true, ClassRogue: true}
		leaders := 0
		for i, h := range party {
			if h.IsLeader {
				leaders++
			}
			if i >= 3 {
				if used[h.Class] {
					t.Fatalf("seed %d: случайный герой %d повторяет занятый класс %s", seed, i, h.Class)
				}
				used[h.Class] = true
				if h.CustomName != "" {
					t.Fatalf("seed %d: случайный герой не должен иметь пользовательского имени", seed)
				}
			}
		}
		if leaders != 1 {
			t.Fatalf("seed %d: лидеров %d", seed, leaders)
		}
	}
}

func TestBuildStartingPartyFullCustom(t *testing.T) {
	seedRNG(7)
	leader := mkBP("Один", ClassTank)
	party := buildStartingParty(0, &leader,
		mkBP("Два", ClassWarrior), mkBP("Три", ClassRogue), mkBP("Четыре", ClassMage), mkBP("Пять", ClassCleric), mkBP("Лишний", ClassBard))
	if len(party) != partySize {
		t.Fatalf("героев %d", len(party))
	}
	want := []string{"Один", "Два", "Три", "Четыре", "Пять"}
	for i, h := range party {
		if h.DisplayName(LangRU) != want[i] {
			t.Errorf("герой %d: %q, ожидалось %q", i, h.DisplayName(LangRU), want[i])
		}
	}
}

func TestBuildStartingPartyAllowsDuplicateCustomClassesAndSkipsInvalid(t *testing.T) {
	seedRNG(8)
	leader := mkBP("Один", ClassWarrior)
	party := buildStartingParty(0, &leader, mkBP("Два", ClassWarrior), mkBP("", ClassMage), mkBP("А", ClassRogue))
	if len(party) != partySize {
		t.Fatalf("героев %d", len(party))
	}
	if party[1].DisplayName(LangRU) != "Два" || party[1].Class != ClassWarrior {
		t.Error("повтор класса среди созданных героев допустим")
	}
	for _, h := range party[2:] {
		if h.CustomName != "" {
			t.Errorf("герои с негодными именами должны быть заменены случайными: %q", h.CustomName)
		}
		if h.Class == ClassWarrior {
			t.Error("случайные бойцы не повторяют занятые классы")
		}
	}
}

func TestBuildStartingPartyWithoutLeaderIgnoresCompanions(t *testing.T) {
	seedRNG(9)
	party := buildStartingParty(0, nil, mkBP("Рагна", ClassMage))
	for _, h := range party {
		if h.IsLeader || h.CustomName != "" {
			t.Fatal("без лидера отряд случайный")
		}
	}
}

func TestRandomPartyBlueprints(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		seedRNG(seed)
		for _, lang := range []Language{LangRU, LangEN} {
			slots := randomPartyBlueprints(lang)
			seen := map[HeroClass]bool{}
			for i, b := range slots {
				if !b.Valid() || b != b.Sanitized() {
					t.Fatalf("seed %d: слот %d негоден: %+v", seed, i, b)
				}
				if seen[b.Class] {
					t.Fatalf("seed %d: повтор класса %s", seed, b.Class)
				}
				seen[b.Class] = true
			}
		}
	}
}

func TestSanitizeCompanions(t *testing.T) {
	leader := mkBP("Один", ClassTank)
	list := []HeroBlueprint{mkBP("Два", ClassMage), mkBP("", ClassRogue), mkBP("Три", ClassBard), mkBP("Четыре", ClassCleric), mkBP("Пять", ClassMonk), mkBP("Шесть", ClassRanger)}
	got := sanitizeCompanions(&leader, list)
	if len(got) != partySize-1 {
		t.Fatalf("спутников %d, максимум %d", len(got), partySize-1)
	}
	if got[0].Name != "Два" || got[1].Name != "Три" {
		t.Errorf("порядок и фильтрация: %+v", got)
	}
	if sanitizeCompanions(nil, list) != nil {
		t.Error("без лидера спутников нет")
	}
	if sanitizeCompanions(&leader, nil) != nil {
		t.Error("пустой список остаётся пустым")
	}
}

func TestSaveRoundTripWithCompanions(t *testing.T) {
	useTempSave(t)
	leader := mkBP("Эдан", ClassTank)
	comp := []HeroBlueprint{mkBP("Рагна", ClassMage), mkBP("Борис", ClassRogue)}
	persistState(LangRU, TownLegacy{}, DefaultTactics(), &leader, comp...)

	got, ok := loadSave()
	if !ok || got.Hero == nil || len(got.Companions) != 2 {
		t.Fatalf("отряд должен сохраняться: %+v", got)
	}
	if got.Companions[0].Name != "Рагна" || got.Companions[1].Class != ClassRogue {
		t.Errorf("спутники изменились: %+v", got.Companions)
	}

	persistState(LangRU, TownLegacy{}, DefaultTactics(), nil, comp...)
	got, _ = loadSave()
	if got.Hero != nil || len(got.Companions) != 0 {
		t.Error("без лидера спутники не сохраняются")
	}
}

func TestLoadSaveDropsBadCompanions(t *testing.T) {
	path := useTempSave(t)
	writeTestFile(t, path, `{"version":1,"lang":"ru","hero":{"name":"Эдан","race":"elf","class":"mage"},`+
		`"companions":[{"name":"%%","class":"tank"},{"name":"Рагна","race":"dragon","class":"hacker"}]}`)
	got, ok := loadSave()
	if !ok || got.Hero == nil {
		t.Fatal("сохранение читается")
	}
	if len(got.Companions) != 1 || got.Companions[0].Name != "Рагна" {
		t.Fatalf("негодный спутник отбрасывается, годный исправляется: %+v", got.Companions)
	}
	if !validRace(got.Companions[0].Race) || !validClass(got.Companions[0].Class) {
		t.Error("значения спутника должны быть исправлены")
	}
}

func TestInitialModelWithCompanions(t *testing.T) {
	useTempSave(t)
	seedRNG(11)
	leader := mkBP("Эдан", ClassTank)
	m := initialModelWith(TownLegacy{}, &leader, mkBP("Рагна", ClassMage))
	if len(m.Party) != partySize || len(m.Companions) != 1 || m.Party[1].DisplayName(LangRU) != "Рагна" {
		t.Fatalf("модель должна собираться вокруг созданного отряда: %+v", m.Companions)
	}
	fresh, _ := resetGameStatic(m)
	if len(fresh.Companions) != 1 || fresh.Party[1].DisplayName(LangRU) != "Рагна" || fresh.Party[1].Level != 1 {
		t.Fatal("созданный отряд переживает рестарт")
	}
}

// ---- Редактор ----

func fillSlot(m *Model, name string, class HeroClass) {
	typeText(m, name)
	gotoRow(m, rowClass)
	for guard := 0; m.Creator.Draft.Class != class && guard < 20; guard++ {
		press(m, "right")
	}
	m.Creator.Row = rowName
}

func TestCreatorFullPartyFlow(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	if m.Creator.Row != rowName || m.Creator.Slot != 0 {
		t.Fatal("редактор открывается на имени лидера")
	}

	fillSlot(m, "Эдан", ClassTank)
	press(m, "pgdown")
	if m.Creator.Slot != 1 || m.Creator.Draft.Name != "" {
		t.Fatalf("PgDn переключает на следующего героя: слот %d, имя %q", m.Creator.Slot, m.Creator.Draft.Name)
	}
	fillSlot(m, "Рагна", ClassMage)
	press(m, "pgdown")
	press(m, "pgdown") // слот 3 пропускаем
	fillSlot(m, "Борис", ClassRogue)

	// Возврат назад не теряет введённое.
	press(m, "pgup")
	press(m, "pgup")
	press(m, "pgup")
	if m.Creator.Slot != 0 || m.Creator.Draft.Name != "Эдан" || m.Creator.Draft.Class != ClassTank {
		t.Fatalf("лидер должен сохраниться: %+v", m.Creator.Draft)
	}

	gotoRow(m, rowConfirm)
	press(m, "enter")
	if m.State != StateMenu {
		t.Fatalf("после подтверждения нужно меню, а состояние %v, ошибка %q", m.State, m.Creator.Err)
	}
	if m.Blueprint == nil || m.Blueprint.Name != "Эдан" || len(m.Companions) != 2 {
		t.Fatalf("лидер и два спутника: %+v / %+v", m.Blueprint, m.Companions)
	}
	if m.Companions[0].Name != "Рагна" || m.Companions[1].Name != "Борис" {
		t.Errorf("порядок спутников: %+v", m.Companions)
	}
	if len(m.Party) != partySize || !m.Party[0].IsLeader || m.Party[1].DisplayName(LangRU) != "Рагна" || m.Party[2].DisplayName(LangRU) != "Борис" {
		t.Fatal("отряд пересобран вокруг созданных героев")
	}

	sd, ok := loadSave()
	if !ok || sd.Hero == nil || len(sd.Companions) != 2 {
		t.Fatalf("отряд сохранён на диск: %+v", sd)
	}
	if m.menuHeroLine() == "" {
		t.Fatal("в меню должна быть строка об отряде")
	}
}

func TestCreatorNeedsLeaderBeforeCompanions(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	press(m, "pgdown")
	fillSlot(m, "Рагна", ClassMage)
	gotoRow(m, rowConfirm)
	press(m, "enter")

	if m.State != StateCreator || m.Blueprint != nil {
		t.Fatal("без лидера отряд не создаётся")
	}
	if m.Creator.Err != "creator.err.leader" || m.Creator.Slot != 0 || m.Creator.Row != rowName {
		t.Fatalf("редактор возвращает к лидеру: err=%q slot=%d row=%d", m.Creator.Err, m.Creator.Slot, m.Creator.Row)
	}
	for _, lang := range []Language{LangRU, LangEN} {
		m.Lang = lang
		if m.View() == "" {
			t.Fatal("экран с ошибкой не рисуется")
		}
	}
	// Рагна не потеряна.
	press(m, "pgdown")
	if m.Creator.Draft.Name != "Рагна" {
		t.Fatalf("введённый спутник должен сохраниться: %q", m.Creator.Draft.Name)
	}
}

func TestCreatorRandomPartyAndClear(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	press(m, "ctrl+t")
	if n := len(m.Creator.filledSlots()); n != partySize {
		t.Fatalf("Ctrl+T заполняет все слоты, а заполнено %d", n)
	}
	seen := map[HeroClass]bool{}
	for _, b := range m.Creator.Slots {
		if seen[b.Class] {
			t.Fatalf("в случайном отряде повтор класса %s", b.Class)
		}
		seen[b.Class] = true
	}

	press(m, "pgdown")
	press(m, "delete")
	if m.Creator.Draft.Name != "" {
		t.Fatal("Delete очищает героя в слоте")
	}
	press(m, "pgup")
	gotoRow(m, rowConfirm)
	press(m, "enter")
	if m.State != StateMenu || len(m.Companions) != partySize-2 {
		t.Fatalf("очищенный слот остаётся случайным: спутников %d", len(m.Companions))
	}
}

func TestCreatorCallingOnlyForLeader(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	press(m, "pgdown")
	before := m.Creator.Draft.Calling
	gotoRow(m, rowCalling)
	press(m, "right")
	if m.Creator.Draft.Calling != before {
		t.Fatal("у спутника призвание не меняется")
	}
	press(m, "pgup")
	gotoRow(m, rowCalling)
	press(m, "right")
	if m.Creator.Draft.Calling == before {
		t.Fatal("у лидера призвание меняется")
	}
}

func TestCreatorSlotRowSwitchesWithArrows(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	fillSlot(m, "Эдан", ClassTank)
	gotoRow(m, rowSlot)
	press(m, "right")
	if m.Creator.Slot != 1 {
		t.Fatalf("стрелка вправо на строке героев: слот %d", m.Creator.Slot)
	}
	press(m, "left")
	press(m, "left")
	if m.Creator.Slot != partySize-1 {
		t.Fatalf("переключение идёт по кругу: слот %d", m.Creator.Slot)
	}
}

func TestCreatorOpensWithExistingParty(t *testing.T) {
	m := menuModel(t)
	leader := mkBP("Эдан", ClassTank)
	m.Blueprint = &leader
	m.Companions = []HeroBlueprint{mkBP("Рагна", ClassMage)}
	press(m, "c")
	if m.Creator.Slots[0].Name != "Эдан" || m.Creator.Slots[1].Name != "Рагна" || m.Creator.Slots[2].Name != "" {
		t.Fatalf("слоты должны подставиться: %+v", m.Creator.Slots)
	}
	if m.Creator.filledSlots()[1] != 1 {
		t.Fatal("заполненные слоты определяются верно")
	}
}

func TestRemoveHeroClearsCompanionsToo(t *testing.T) {
	m := menuModel(t)
	leader := mkBP("Эдан", ClassTank)
	m.Blueprint = &leader
	m.Companions = []HeroBlueprint{mkBP("Рагна", ClassMage)}
	m.Party = buildStartingParty(0, m.Blueprint, m.Companions...)

	press(m, "x")
	press(m, "x")
	if m.Blueprint != nil || len(m.Companions) != 0 {
		t.Fatal("удаление убирает весь созданный отряд")
	}
	sd, ok := loadSave()
	if !ok || sd.Hero != nil || len(sd.Companions) != 0 {
		t.Fatal("и из сохранения")
	}
}

func TestCreatorScreenRendersEverySlotAndLanguage(t *testing.T) {
	m := menuModel(t)
	press(m, "c")
	for _, lang := range []Language{LangRU, LangEN} {
		m.Lang = lang
		for slot := 0; slot < partySize; slot++ {
			m.switchSlot(slot)
			for row := 0; row < creatorRows; row++ {
				m.Creator.Row = row
				for _, size := range [][2]int{{60, 22}, {100, 30}, {160, 50}} {
					m.TermWidth, m.TermHeight = size[0], size[1]
					if m.View() == "" {
						t.Fatalf("%s слот %d строка %d %v: пусто", lang, slot, row, size)
					}
				}
			}
		}
	}
}
