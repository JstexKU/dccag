package main

import "testing"

// Чем сильнее отставание, тем больше прибавка и дешевле эликсир; опытным героям он стоит дороже.
func TestElixirTiers(t *testing.T) {
	prevBonus, prevMult := 1<<30, -1
	for gap := 6; gap >= -3; gap-- {
		bonus, mult := elixirTier(gap)
		if bonus > prevBonus || mult < prevMult {
			t.Errorf("gap=%d: бонус %d и множитель %d не монотонны (было %d и %d)", gap, bonus, mult, prevBonus, prevMult)
		}
		prevBonus, prevMult = bonus, mult
	}
	if b, _ := elixirTier(5); b != 150 {
		t.Errorf("сильно отстающий должен получать +150%%, получил %d", b)
	}
	if b, _ := elixirTier(-5); b >= 25 {
		t.Errorf("опережающий должен получать меньше +25%%, получил %d", b)
	}

	// При равном уровне новичку (отстаёт) дешевле, чем опытному (опережает).
	novice, _ := elixirCost(10, 3)
	par, _ := elixirCost(10, 0)
	ahead, _ := elixirCost(10, -2)
	if !(novice < par && par < ahead) {
		t.Errorf("цены %d < %d < %d нарушены", novice, par, ahead)
	}

	// Эликсир не бывает дешёвым: даже самому слабому он стоит больше найма ветерана (85G).
	if cost, _ := elixirCost(1, 5); cost < 120 {
		t.Errorf("слишком дёшево: %d", cost)
	}
}

func elixirModel(t *testing.T) Model {
	t.Helper()
	seedRNG(61)
	m := initialModel()
	m.State = StatePlaying
	m.Lang = LangRU
	m.MaxFloor = 11 // рекомендуемый уровень 6
	m.Gold = 100000
	for _, h := range m.Party {
		h.IsDead, h.IsDowned, h.LostInAbyss = false, false, false
		h.Level, h.Exp = 6, 0
		h.ExpElixirPct = 0
	}
	return m
}

// Предложение есть у каждого живого героя без эликсира; покупка действует до конца вылазки.
func TestElixirOfferPurchaseAndExpiry(t *testing.T) {
	m := elixirModel(t)
	laggard := m.Party[2]
	laggard.Level = 1

	var offer *AlchemistOffer
	count := 0
	for _, o := range m.GenerateAlchemistOffers() {
		if o.Type == AlchemistActionElixir {
			count++
			if o.HeroIdx == 2 {
				cp := o
				offer = &cp
			}
		}
	}
	if count != len(m.Party) || offer == nil {
		t.Fatalf("предложений эликсира %d из %d героев", count, len(m.Party))
	}
	if offer.ExpPct != 150 || offer.Gap != 5 {
		t.Errorf("для отставания на 5 уровней: +%d%%, gap=%d", offer.ExpPct, offer.Gap)
	}

	gold := m.Gold
	if !m.ExecuteAlchemistOffer(*offer) {
		t.Fatal("покупка не удалась")
	}
	if laggard.ExpElixirPct != 150 || m.Gold != gold-offer.Cost {
		t.Errorf("после покупки: бонус %d, золото %d (ожидалось %d)", laggard.ExpElixirPct, m.Gold, gold-offer.Cost)
	}

	// Повторное предложение герою с действующим эликсиром не выдаётся, устаревшее — не списывает золото.
	for _, o := range m.GenerateAlchemistOffers() {
		if o.Type == AlchemistActionElixir && o.HeroIdx == 2 {
			t.Error("герою с эликсиром предлагают второй")
		}
	}
	gold = m.Gold
	if m.ExecuteAlchemistOffer(*offer) || m.Gold != gold {
		t.Error("устаревшее предложение не должно списывать золото")
	}

	// Возвращение в город снимает эффект.
	m.enterTown(RetreatNone)
	if laggard.ExpElixirPct != 0 {
		t.Errorf("эликсир пережил вылазку: %d%%", laggard.ExpElixirPct)
	}
}

// Опыт героя с эликсиром растёт на указанный процент, у остальных не меняется.
func TestElixirBoostsExperience(t *testing.T) {
	gain := func(elixir int) (boosted, other int) {
		m := elixirModel(t)
		m.Party[1].ExpElixirPct = elixir
		levels := []int{m.Party[1].Level, m.Party[3].Level}
		m.distributePartyExp(100)
		if m.Party[1].Level != levels[0] || m.Party[3].Level != levels[1] {
			t.Fatal("в тесте не должно быть повышения уровня")
		}
		return m.Party[1].Exp, m.Party[3].Exp
	}
	baseBoosted, baseOther := gain(0)
	boosted, other := gain(100)
	if other != baseOther {
		t.Errorf("у героя без эликсира опыт изменился: %d -> %d", baseOther, other)
	}
	if boosted != baseBoosted*2 {
		t.Errorf("+100%%: опыт %d, ожидалось %d", boosted, baseBoosted*2)
	}
}

// Автопилот покупает эликсир только отстающему и только при достаточном бюджете.
func TestAutopilotElixirTargetsLaggards(t *testing.T) {
	m := elixirModel(t)
	if idx := m.pickAutoElixir(m.GenerateAlchemistOffers()); idx != -1 {
		t.Fatalf("при ровном отряде автопилот купил эликсир (индекс %d)", idx)
	}

	m.Party[2].Level = 1
	offers := m.GenerateAlchemistOffers()
	idx := m.pickAutoElixir(offers)
	if idx == -1 || offers[idx].HeroIdx != 2 || offers[idx].Type != AlchemistActionElixir {
		t.Fatalf("автопилот должен выбрать эликсир отстающему, получил %d", idx)
	}

	m.Gold = 0
	if idx := m.pickAutoElixir(m.GenerateAlchemistOffers()); idx != -1 {
		t.Error("без золота эликсир купить нельзя")
	}
}

// targetLevel и weakHeroes остаются согласованными после выноса общей функции.
func TestWeakHeroesUseTargetLevel(t *testing.T) {
	m := elixirModel(t)
	if len(m.weakHeroes()) != 0 {
		t.Fatal("ровный отряд не должен считаться слабым")
	}
	if m.targetLevel() != 6 {
		t.Errorf("целевой уровень %d, ожидалось 6", m.targetLevel())
	}
	m.Party[3].Level = 1
	weak := m.weakHeroes()
	if len(weak) != 1 || weak[0] != m.Party[3] {
		t.Errorf("слабых: %d", len(weak))
	}
}

// Эликсир виден на карточке героя.
func TestElixirBadge(t *testing.T) {
	m := elixirModel(t)
	h := m.Party[1]
	h.IsDowned, h.ReviveCooldown, h.Affliction = false, 0, AfflictionNone
	h.IsGuarding, h.IsBerserk, h.IsStealthed, h.IsCharged, h.IsAura = false, false, false, false, false
	h.PoisonTurns, h.ChilledTurns = 0, 0
	h.ExpElixirPct = 80
	if got := stripANSI(buffBadge(h)); got != "[📚]" {
		t.Errorf("значок эликсира: %q", got)
	}
}
