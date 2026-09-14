package main

// ============================================================
// MILITIA PROGRESSION — мета-прогрессия городского ополчения
// ============================================================
//
// Гильдия выставляет бесплатного бойца, если у отряда не хватает
// золота на наём ветерана. Сила этого ополченца масштабируется от
// суммарных вложений в столицу за все прогоны (m.Legacy.TotalInvested).
//
// Правила:
//   1. Ополчение ВСЕГДА отстаёт от наёмного ветерана минимум на 2
//      «этажных эквивалента» (level/gear). Т.е. на этаже 10 ветеран
//      создаётся через createHero(class, 10, smithy), а ополчение —
//      максимум через createHero(class, 8, smithy/2).
//   2. Кузнечный бонус к базовой атаке у ополчения — половина от
//      уровня городской кузницы (ополчение — не регулярная армия).
//   3. Класс ополченца определяется вызывающим кодом (Гильдия) —
//      тем же пулом availableClasses, что и у ветерана, чтобы
//      сохранить уникальность классов в отряде.
//   4. Титул ополчения зависит от тира и может быть заменён любым
//      боевым титулом через checkAndAwardTitle (см. isMilitiaTitle).

// MilitiaTier — ступень городского ополчения.
type MilitiaTier struct {
	TitleKey    string // ключ локализации для титула
	FloorEquiv  int    // «этаж» для createHero — задаёт уровень и качество снаряжения
	InvestFloor int    // минимальный порог накопленных вложений (G)
}

// militiaTiers отсортирован по УБЫВАНИЮ InvestFloor.
// Первая подходящая ступень в militiaTierForInvestment — победитель.
var militiaTiers = []MilitiaTier{
	{TitleKey: "title.militia_capital_guard", FloorEquiv: 16, InvestFloor: 12000},
	{TitleKey: "title.militia_elite", FloorEquiv: 12, InvestFloor: 6000},
	{TitleKey: "title.militia_veteran", FloorEquiv: 8, InvestFloor: 2500},
	{TitleKey: "title.militia_sergeant", FloorEquiv: 5, InvestFloor: 800},
	{TitleKey: "title.militia_regular", FloorEquiv: 3, InvestFloor: 200},
	{TitleKey: "title.militia", FloorEquiv: 1, InvestFloor: 0},
}

// militiaTierForInvestment возвращает ступень ополчения для указанного
// объёма накопленных вложений в столицу. Всегда возвращает валидный
// тир (fallback — тир 0 «Ополченец»).
func militiaTierForInvestment(totalInvested int) MilitiaTier {
	for _, t := range militiaTiers {
		if totalInvested >= t.InvestFloor {
			return t
		}
	}
	return militiaTiers[len(militiaTiers)-1]
}

// nextMilitiaTier возвращает следующую (более высокую) ступень относительно
// текущего объёма вложений, либо nil, если достигнут максимум.
// Используется в UI для отображения прогресса.
func nextMilitiaTier(totalInvested int) *MilitiaTier {
	// Идём от младших тиров к старшим, ищем первый, порог которого не достигнут.
	for i := len(militiaTiers) - 1; i >= 0; i-- {
		if totalInvested < militiaTiers[i].InvestFloor {
			return &militiaTiers[i]
		}
	}
	return nil
}

// isMilitiaTitle возвращает true для любого титула ополчения (включая
// пустой ключ и базовый "title.militia"). Такие титулы разрешено
// перезаписывать боевыми достижениями через checkAndAwardTitle.
func isMilitiaTitle(titleKey string) bool {
	if titleKey == "" || titleKey == "title.militia" {
		return true
	}
	for _, t := range militiaTiers {
		if titleKey == t.TitleKey {
			return true
		}
	}
	return false
}

// createMilitiaForGuild создаёт бесплатного ополченца, масштабированного
// от суммарных вложений в столицу. Класс передаётся вызывающим кодом.
//
// Ограничения:
//   - floorEquiv ≤ m.Floor - 2  (ополчение не догоняет ветерана);
//   - floorEquiv ≥ 1            (не уходим в 0/отрицательные значения);
//   - кузница даёт только половину бонуса к базовой атаке.
func (m *Model) createMilitiaForGuild(class HeroClass) *Hero {
	tier := militiaTierForInvestment(m.Legacy.TotalInvested)

	// Жёсткий кап: минимум на 2 этажа позади текущего прогресса.
	capFloor := m.Floor - 2
	if capFloor < 1 {
		capFloor = 1
	}
	floorEquiv := tier.FloorEquiv
	if floorEquiv > capFloor {
		floorEquiv = capFloor
	}

	// Кузница влияет на базовую атаку через createHero → BaseAtk = 6 + smithyLvl.
	// Ополчению даём только половину бонуса: город оружие выдаёт, но по остаточному принципу.
	militiaSmithy := m.Legacy.SmithyLevel / 2

	h := createHero(class, floorEquiv, militiaSmithy)
	h.TitleKey = tier.TitleKey
	return h
}