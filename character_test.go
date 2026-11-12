package main

import (
	"strings"
	"testing"
)

func charModel(t *testing.T) Model {
	t.Helper()
	seedRNG(111)
	m := initialModel()
	m.State = StatePlaying
	m.Lang = LangRU
	m.TermWidth, m.TermHeight = 120, 45
	return m
}

func sheetText(m Model, h *Hero) string {
	return stripANSI(strings.Join(m.charSheetLines(h), "\n"))
}

// Порог крита считается в одном месте для боя и для окна: предметы, раса, реликвия, скрытность.
func TestCritThresholdRules(t *testing.T) {
	m := charModel(t)
	h := m.Party[1]
	clearGear(h)
	h.Race, h.IsStealthed = RaceHuman, false

	if got := m.critThreshold(h); got != 19 {
		t.Errorf("базовый порог %d, ожидалось 19", got)
	}
	h.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, CritBonus: 2})
	if got := m.critThreshold(h); got != 17 {
		t.Errorf("с предметом +2: %d, ожидалось 17", got)
	}
	h.SetItemInSlot(SlotLegs, &EquipItem{Slot: SlotLegs, CritBonus: 9}) // поножи на крит не влияют
	if got := m.critThreshold(h); got != 17 {
		t.Errorf("поножи не должны менять порог: %d", got)
	}
	h.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, CritBonus: 20})
	if got := m.critThreshold(h); got != 14 {
		t.Errorf("порог не должен опускаться ниже 14: %d", got)
	}
	h.SetItemInSlot(SlotChest, nil)

	m.Relic = &PartyRelic{PartyCritBonus: 1}
	if got := m.critThreshold(h); got != 16 {
		t.Errorf("с реликвией: %d, ожидалось 16", got)
	}
	m.Relic = nil
	h.Race = RaceElf // +2 к шансу крита
	if got := m.critThreshold(h); got != 15 {
		t.Errorf("эльф: %d, ожидалось 15", got)
	}
	h.IsStealthed = true
	if got := m.critThreshold(h); got != 1 {
		t.Errorf("скрытность: %d, ожидалось 1", got)
	}

	for thr, want := range map[int]int{19: 10, 17: 20, 14: 35, 1: 100, 25: 5} {
		if got := critChancePct(thr); got != want {
			t.Errorf("шанс крита при пороге %d: %d%%, ожидалось %d%%", thr, got, want)
		}
	}
}

// Шкала опыта показывает, сколько осталось до следующего уровня.
func TestCharSheetExperienceBar(t *testing.T) {
	m := charModel(t)
	h := m.Party[1]
	h.Level, h.Exp = 2, 153
	txt := sheetText(m, h)

	if !strings.Contains(txt, "153/300 (51%)") {
		t.Errorf("нет шкалы опыта 153/300 (51%%):\n%s", txt)
	}
	if !strings.Contains(txt, "До уровня 3 осталось 147 очков опыта") {
		t.Errorf("нет строки «до следующего уровня»:\n%s", txt)
	}
	m.Lang = LangEN
	if !strings.Contains(sheetText(m, h), "To reach level 3: 147 XP left") {
		t.Error("английская строка опыта не найдена")
	}
}

// Карточка собирается для любого героя на всех языках, без сырых ключей и «%!».
func TestCharSheetIsCompleteInAllLanguages(t *testing.T) {
	m := charModel(t)
	h := m.Party[0]
	h.IsLeader, h.Calling = true, CallingMentor
	h.ExpElixirPct = 80
	h.Mutations.FuryCount, h.Mutations.TitanCount = 2, 1
	h.PoisonTurns, h.PoisonDmg, h.ChilledTurns = 2, 3, 1
	h.IsGuarding = true
	h.Potions = []*Potion{{Type: PotionHP, Size: SizeSmall, Power: 25, Symbol: "🔴"}}
	w := plainWeapon(h)
	w.Enchant = &Enchant{Kind: EnchantLightning, Level: 2}
	w.Prefix = &PrefixDef{Key: "prefix.fire", Element: ElemFire, Bonus: 3}
	h.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, BaseNameKey: "armor.head.1", BonusHP: 20, BonusMP: 5,
		CritBonus: 1, SpeedBonus: 1, StressRes: 10, Prefix: &PrefixDef{Key: "prefix.frost", Element: ElemFrost, Bonus: 2},
		Suffix: &SuffixDef{Key: "suffix.fury", Effect: SuffFury, Bonus: 4}})
	h.Feats.Kills, h.Feats.BossKills, h.Feats.DamageDealt = 12, 1, 340

	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		m.Lang = lang
		for i, hero := range m.Party {
			txt := sheetText(m, hero)
			if strings.Contains(txt, "char.") || strings.Contains(txt, "%!") || strings.Contains(txt, "affliction.") {
				t.Errorf("%s, герой %d: сырые ключи или ошибки формата:\n%s", lang, i, txt)
			}
			for _, key := range []string{"char.sec.vitals", "char.sec.combat", "char.sec.abilities", "char.sec.bonuses",
				"char.sec.equipment", "char.sec.belt", "char.sec.records"} {
				if !strings.Contains(txt, T(lang, key)) {
					t.Errorf("%s, герой %d: нет раздела %s", lang, i, key)
				}
			}
		}
	}

	m.Lang = LangRU
	txt := sheetText(m, h)
	for _, want := range []string{"★", "+80%", "×2", "Мутации", "Отравлен", "защитной стойке", "⚡ Молния 2", "(врождённый)", "здоровье +20", "Побеждено монстров: ", "340"} {
		if !strings.Contains(txt, want) {
			t.Errorf("в карточке нет %q:\n%s", want, txt)
		}
	}
}

// Пограничные герои: павший, без сознания, потерянный, без снаряжения, навыка и зелий.
func TestCharSheetEdgeHeroes(t *testing.T) {
	m := charModel(t)
	bare := &Hero{CustomName: "Новичок", Race: RaceHuman, Class: ClassWarrior, Level: 1, MaxHP: 30, HP: 30}
	dead := m.Party[1]
	dead.IsDead, dead.HP = true, 0
	downed := m.Party[2]
	downed.IsDowned, downed.HP, downed.ReviveCooldown = true, 0, 3
	lost := m.Party[3]
	lost.LostInAbyss = true
	m.Party = append(m.Party, bare)

	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		m.Lang = lang
		for i, h := range m.Party {
			txt := sheetText(m, h)
			if strings.Contains(txt, "%!") || strings.Contains(txt, "char.") {
				t.Errorf("%s, герой %d: %s", lang, i, txt)
			}
		}
	}
	m.Lang = LangRU
	if !strings.Contains(sheetText(m, dead), "МЁРТВ") && !strings.Contains(sheetText(m, dead), T(LangRU, "ui.dead")) {
		t.Error("павший герой не помечен")
	}
	if !strings.Contains(sheetText(m, downed), T(LangRU, "char.downed")) || !strings.Contains(sheetText(m, lost), T(LangRU, "char.lost")) {
		t.Error("герой без сознания или потерянный не помечен")
	}
	if !strings.Contains(sheetText(m, bare), T(LangRU, "char.empty_slot")) {
		t.Error("пустые слоты снаряжения не показаны")
	}
}

func TestCharacterKeyHandling(t *testing.T) {
	m := charModel(t)
	n := len(m.Party)
	m.openCharacterSheet()
	if m.State != StateCharacter || m.CharIndex != 0 {
		t.Fatalf("окно не открылось: state=%v index=%d", m.State, m.CharIndex)
	}

	m.handleCharacterKey("tab")
	if m.CharIndex != 1 {
		t.Errorf("tab: индекс %d", m.CharIndex)
	}
	m.handleCharacterKey("shift+tab")
	m.handleCharacterKey("shift+tab")
	if m.CharIndex != n-1 {
		t.Errorf("shift+tab с первого героя должен переходить к последнему: %d", m.CharIndex)
	}
	m.handleCharacterKey("right")
	if m.CharIndex != 0 {
		t.Errorf("right с последнего героя должен возвращать к первому: %d", m.CharIndex)
	}
	m.handleCharacterKey("3")
	if m.CharIndex != 2 {
		t.Errorf("цифра 3: индекс %d", m.CharIndex)
	}
	m.handleCharacterKey("9")
	if m.CharIndex != 2 {
		t.Error("цифра за пределами отряда должна игнорироваться")
	}

	m.StatsScroll = 0
	m.handleCharacterKey("up")
	if m.StatsScroll != 0 {
		t.Error("прокрутка вверх не должна уходить в минус")
	}
	m.TermHeight = 14 // карточка заведомо длиннее окна
	for i := 0; i < 200; i++ {
		m.handleCharacterKey("down")
	}
	if m.StatsScroll != m.charMaxScroll() || m.StatsScroll == 0 {
		t.Errorf("прокрутка %d, максимум %d", m.StatsScroll, m.charMaxScroll())
	}
	m.handleCharacterKey("tab")
	if m.StatsScroll != 0 {
		t.Error("смена героя должна сбрасывать прокрутку")
	}

	// Выход, запасные клавиши и «проглатывание» остальных.
	if _, handled := m.handleCharacterKey("enter"); !handled {
		t.Error("enter за окном не должен доходить до игры")
	}
	for _, key := range []string{"q", "l", "ctrl+c"} {
		if _, handled := m.handleCharacterKey(key); handled {
			t.Errorf("%q должна работать и в окне", key)
		}
	}
	for _, key := range []string{"p", "h", "esc"} {
		m.State = StateCharacter
		if _, handled := m.handleCharacterKey(key); !handled || m.State != StatePlaying {
			t.Errorf("%q должна закрывать окно", key)
		}
	}
}

// Через настоящий Update: P открывает, Tab листает, Enter за окном ничего не делает, P закрывает.
func TestCharacterWindowThroughUpdate(t *testing.T) {
	m := charModel(t)
	m.AutoMode = false

	press(&m, "p")
	if m.State != StateCharacter {
		t.Fatalf("P не открыла окно: %v", m.State)
	}
	press(&m, "tab")
	if m.CharIndex != 1 {
		t.Errorf("Tab: индекс %d", m.CharIndex)
	}
	press(&m, "enter")
	if m.State != StateCharacter || m.AutoMode {
		t.Error("Enter за окном не должен менять игру")
	}
	if view := stripANSI(m.View()); !strings.Contains(view, m.Party[1].DisplayName(LangRU)) || !strings.Contains(view, "ПЕРСОНАЖИ") {
		t.Error("в окне нет имени выбранного героя или заголовка")
	}
	press(&m, "p")
	if m.State != StatePlaying {
		t.Errorf("P не закрыла окно: %v", m.State)
	}

	// Русская раскладка: «з» — это P.
	press(&m, "з")
	if m.State != StateCharacter {
		t.Error("клавиша «з» (P в русской раскладке) не открыла окно")
	}
	press(&m, "esc")
	if m.State != StatePlaying {
		t.Error("Esc не закрыла окно")
	}

	// H открывает окно и в ручном режиме, где P занята зельями.
	m.ManualMode = true
	press(&m, "h")
	if m.State != StateCharacter {
		t.Error("H не открыла окно в ручном режиме")
	}
}

func TestCharacterWindowRendersInAllSizes(t *testing.T) {
	m := charModel(t)
	m.State = StateCharacter
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		m.Lang = lang
		for _, size := range [][2]int{{120, 45}, {80, 24}, {50, 14}} {
			m.TermWidth, m.TermHeight = size[0], size[1]
			for i := range m.Party {
				m.CharIndex = i
				if view := m.View(); strings.TrimSpace(stripANSI(view)) == "" {
					t.Errorf("%s %dx%d герой %d: пустой экран", lang, size[0], size[1], i)
				}
			}
		}
	}
}
