package main

import (
	"os"
	"regexp"
	"testing"
)

// Версия в main.go совпадает с последним выпуском в CHANGELOG.md, а «Unreleased» в истории один.
func TestVersionMatchesChangelog(t *testing.T) {
	raw, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Skip("CHANGELOG.md недоступен: " + err.Error())
	}
	text := string(raw)
	// Считаем именно заголовки (в начале строки): само это слово может встретиться в тексте записи.
	if n := len(regexp.MustCompile(`(?m)^## \[Unreleased\]`).FindAllString(text, -1)); n > 1 {
		t.Errorf("заголовок [Unreleased] встречается %d раз", n)
	}
	m := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`).FindStringSubmatch(text)
	if m == nil {
		t.Fatal("в CHANGELOG.md нет ни одного выпуска")
	}
	if m[1] != version {
		t.Errorf("version=%q, а последний выпуск в CHANGELOG.md — %q", version, m[1])
	}
}

func TestMutationLimitByLevel(t *testing.T) {
	for level, want := range map[int]int{1: 5, 9: 5, 10: 5, 12: 6, 40: 20, 183: 91} {
		if got := maxHeroMutations(level); got != want {
			t.Errorf("уровень %d: предел %d, ожидалось %d", level, got, want)
		}
	}
}

func mutationOffers(m *Model, heroIdx int) []AlchemistOffer {
	var out []AlchemistOffer
	for _, o := range m.GenerateAlchemistOffers() {
		if o.Type == AlchemistActionMutation && o.HeroIdx == heroIdx {
			out = append(out, o)
		}
	}
	return out
}

// Герой на пределе мутаций не получает предложения; поднял уровень — получает снова.
func TestAlchemistStopsOfferingMutationsAtTheLimit(t *testing.T) {
	m := elixirModel(t)
	h := m.Party[1]
	h.Level = 1

	h.Mutations = HeroMutations{FuryCount: 4}
	if len(mutationOffers(&m, 1)) != 1 {
		t.Fatal("на 4 мутациях из 5 предложение должно быть")
	}
	h.Mutations = HeroMutations{FuryCount: 5}
	if len(mutationOffers(&m, 1)) != 0 {
		t.Fatal("на пределе мутаций предложения быть не должно")
	}
	h.Level = 12 // предел 6
	offers := mutationOffers(&m, 1)
	if len(offers) != 1 {
		t.Fatal("после роста уровня предложение должно вернуться")
	}

	// Покупка до предела проходит, устаревшее предложение сверх предела не продаётся и не списывает золото.
	if !m.ExecuteAlchemistOffer(offers[0]) || h.Mutations.Total() != 6 {
		t.Fatalf("покупка не прошла: мутаций %d", h.Mutations.Total())
	}
	gold := m.Gold
	if m.ExecuteAlchemistOffer(offers[0]) || m.Gold != gold || h.Mutations.Total() != 6 {
		t.Error("устаревшее предложение сверх предела должно отклоняться без списания золота")
	}
}

// Автопилот не приходит к алхимику, если продавать нечего: лимит виден и townNeedsVisit.
func TestAutopilotSkipsAlchemistWhenEveryoneIsCapped(t *testing.T) {
	m := elixirModel(t) // отряд ровный: автоэликсир не нужен
	m.Gold = 1_000_000
	for _, h := range m.Party {
		h.Level = 1
		h.Mutations = HeroMutations{FuryCount: 5}
		h.Potions = nil
		for len(h.Potions) < m.MaxPotionSlots() {
			p := createPotion(PotionHP, SizeSmall, 1)
			h.Potions = append(h.Potions, &p)
		}
	}
	m.MaxFloor = 1

	if m.townNeedsVisit(TownPhaseAlchemist) {
		t.Error("все на пределе мутаций, пояса полны: лавка алхимика не должна быть нужна")
	}
	m.Party[0].Mutations = HeroMutations{}
	if !m.townNeedsVisit(TownPhaseAlchemist) {
		t.Error("у героя без мутаций автопилот должен зайти к алхимику")
	}
}

func TestFleeChance(t *testing.T) {
	mk := func(floor int, classes ...HeroClass) *Model {
		seedRNG(131)
		m := initialModel()
		m.State = StatePlaying
		m.Floor = floor
		for i, h := range m.Party {
			h.IsDead, h.IsDowned = false, false
			h.Class = ClassWarrior
			if i < len(classes) {
				h.Class = classes[i]
			}
		}
		return &m
	}

	if got := mk(1).fleeChance(); got != 55 {
		t.Errorf("обычный отряд на первом этаже: %d%%, ожидалось 55%%", got)
	}
	if got := mk(105).fleeChance(); got != 45 {
		t.Errorf("на 105 этаже шанс не должен падать ниже 45%%, получили %d%%", got)
	}
	// 55 + 15 + 15 + 10 + 10 − 40/4 = 95, но предел — 90%.
	m := mk(40, ClassRogue, ClassRanger, ClassTank, ClassPaladin)
	if got := m.fleeChance(); got != 90 {
		t.Errorf("с ловкими и защитниками на 40 этаже: %d%%, ожидалось 90%% (предел)", got)
	}
	// Без предела было бы заметно: на 80 этаже 105 − 20 = 85.
	if got := mk(80, ClassRogue, ClassRanger, ClassTank, ClassPaladin).fleeChance(); got != 85 {
		t.Errorf("на 80 этаже: %d%%, ожидалось 85%%", got)
	}
	if got := mk(1, ClassRogue, ClassRanger, ClassTank, ClassPaladin).fleeChance(); got != 90 {
		t.Errorf("шанс не должен превышать 90%%: %d%%", got)
	}
	// Павшие и бессознательные не помогают.
	m = mk(1, ClassRogue, ClassRogue, ClassWarrior)
	m.Party[0].IsDead = true
	m.Party[1].IsDowned = true
	if got := m.fleeChance(); got != 55 {
		t.Errorf("павшие не должны давать бонус: %d%%", got)
	}
}

// Неудачный побег стоит 10% здоровья (не меньше 2) и задаёт паузу 8 ходов.
func TestFailedFleeCostAndCooldown(t *testing.T) {
	for seed := int64(1); seed <= 400; seed++ {
		seedRNG(seed)
		m := initialModel()
		m.State = StatePlaying
		m.Combat = &ActiveCombat{Pack: ratPack()}
		before := map[*Hero]int{}
		for _, h := range m.Party {
			h.IsDead, h.IsDowned = false, false
			h.HP = h.MaxHP
			before[h] = h.HP
		}

		m.attemptFlee()
		if m.Combat == nil {
			continue // побег удался — ищем неудачный
		}
		if m.Combat.FleeCooldown != 8 {
			t.Errorf("пауза после неудачи %d, ожидалось 8", m.Combat.FleeCooldown)
		}
		for _, h := range m.Party {
			chip := max(2, int(float64(before[h])*0.10))
			if h.HP != before[h]-chip {
				t.Errorf("урон от неудачного побега: HP %d -> %d, ожидалось -%d", before[h], h.HP, chip)
			}
		}
		return
	}
	t.Fatal("за 400 зёрен не нашлось ни одного неудачного побега")
}

// Ветеран, переросший этаж, не идёт тренироваться из-за разницы в пару уровней.
func TestVeteransAreNotWeakOverTinyGaps(t *testing.T) {
	m := elixirModel(t)
	m.MaxFloor = 105 // рекомендуемый уровень около 53
	levels := []int{181, 183, 185, 185, 183}
	for i, h := range m.Party {
		h.Level = levels[i%len(levels)]
	}
	if m.targetLevel() != 183 {
		t.Fatalf("целевой уровень %d, ожидалось 183 (в тесте подобраны уровни)", m.targetLevel())
	}
	if weak := m.weakHeroes(); len(weak) != 0 {
		t.Fatalf("ветеран 181 уровня не должен считаться слабым: %d", len(weak))
	}

	// Настоящий отстающий всё ещё слабый.
	m.Party[2].Level = 100
	if weak := m.weakHeroes(); len(weak) != 1 || weak[0] != m.Party[2] {
		t.Errorf("отстающий на 80 уровней должен считаться слабым: %d", len(weak))
	}

	// На малых уровнях старое правило работает как раньше (цель 6, герой 4: разница 2 из 6 — это слабый).
	m = elixirModel(t)
	m.Party[3].Level = 4
	if weak := m.weakHeroes(); len(weak) != 1 {
		t.Errorf("на малых уровнях отставание на 2 должно считаться слабостью: %d", len(weak))
	}
}
