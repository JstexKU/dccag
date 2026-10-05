package main

import "testing"

func TestMarketGenerateOffersAndBuy(t *testing.T) {
	seedRNG(42)
	m := initialModel()
	m.State = StatePlaying
	m.InTown = true
	m.TownPhase = TownPhaseSellLoot
	m.Gold = 1000

	m.initMarketScreen()
	offers := m.MarketState.Offers
	if len(offers) == 0 {
		t.Fatal("рынок должен генерировать предложения экипировки")
	}

	firstOffer := offers[0]
	hero := m.Party[firstOffer.HeroIdx]
	initialStat := 0
	if oldItem := hero.GetItemInSlot(firstOffer.Slot); oldItem != nil {
		initialStat = oldItem.TotalStat()
	}

	goldBefore := m.Gold
	success := m.ExecuteMarketBuy(firstOffer)
	if !success {
		t.Fatal("покупка должна пройти при наличии золота")
	}

	if m.Gold != goldBefore-firstOffer.Cost {
		t.Fatalf("золото списано некорректно: было %d, стало %d, цена %d", goldBefore, m.Gold, firstOffer.Cost)
	}

	newItem := hero.GetItemInSlot(firstOffer.Slot)
	if newItem == nil || newItem.TotalStat() <= initialStat {
		t.Fatal("предмет должен быть надет и давать прирост характеристик")
	}
}

func TestMarketSellLootAndSellAll(t *testing.T) {
	seedRNG(43)
	m := initialModel()
	m.Bag = []EquipItem{
		generateItemForClass(ClassWarrior, 1),
		generateItemForClass(ClassMage, 1),
	}
	m.Gold = 100
	itemVal := m.Bag[0].Value * 2

	// Одиночная продажа
	m.ExecuteMarketSellLoot(0)
	if len(m.Bag) != 1 {
		t.Fatalf("в мешке должен остаться 1 предмет, осталось %d", len(m.Bag))
	}
	if m.Gold != 100+itemVal {
		t.Fatalf("золото после продажи: %d, ожидалось %d", m.Gold, 100+itemVal)
	}

	// Оптовая продажа
	goldBefore := m.Gold
	restVal := m.Bag[0].Value * 2
	m.ExecuteMarketSellAll()
	if len(m.Bag) != 0 {
		t.Fatal("мешок должен быть пуст после SellAll")
	}
	if m.Gold != goldBefore+restVal {
		t.Fatalf("золото после SellAll: %d, ожидалось %d", m.Gold, goldBefore+restVal)
	}
}

func TestMarketNavigationKeys(t *testing.T) {
	seedRNG(44)
	m := initialModel()
	m.ManualMode = true
	m.InTown = true
	m.TownPhase = TownPhaseSellLoot
	m.Bag = []EquipItem{generateItemForClass(ClassRogue, 1)}
	m.initMarketScreen()

	// Tab переключает секцию
	press(&m, "tab")
	if m.MarketState.Section != MarketSectionShop {
		t.Fatal("Tab должен переключать секцию на Shop")
	}
	press(&m, "tab")
	if m.MarketState.Section != MarketSectionLoot {
		t.Fatal("повторный Tab должен возвращать на Loot")
	}

	// Esc переводит фазу города дальше
	press(&m, "esc")
	if m.TownPhase != TownPhaseMagistrate {
		t.Fatalf("Esc должен переводить фазу в Magistrate, текущая: %v", m.TownPhase)
	}
}

func TestMagistrateInvestAndNavigation(t *testing.T) {
	seedRNG(50)
	m := initialModel()
	m.ManualMode = true
	m.InTown = true
	m.TownPhase = TownPhaseMagistrate
	m.Gold = 1000
	m.initMagistrateScreen()

	if len(m.MagistrateState.Offers) == 0 {
		t.Fatal("магистрат должен генерировать предложения улучшений")
	}

	initialLvl := m.Legacy.SmithyLevel
	// Покупка первого улучшения (Кузница)
	press(&m, "enter")
	if m.Legacy.SmithyLevel != initialLvl+1 {
		t.Fatalf("уровень кузницы должен вырасти: было %d, стало %d", initialLvl, m.Legacy.SmithyLevel)
	}

	// Esc переводит в Храм
	press(&m, "esc")
	if m.TownPhase != TownPhaseChurch {
		t.Fatalf("Esc должен переводить в Church, текущая фаза: %v", m.TownPhase)
	}
}

func TestChurchCleanseAndStabilize(t *testing.T) {
	seedRNG(51)
	m := initialModel()
	m.ManualMode = true
	m.InTown = true
	m.TownPhase = TownPhaseChurch
	m.Gold = 500

	// Делаем одного бойца поверженным, а второго стрессованным
	m.Party[0].IsDowned = true
	m.Party[1].Stress = 80

	m.initChurchScreen()
	if len(m.ChurchState.Offers) < 2 {
		t.Fatalf("ожидалось минимум 2 предложения в храме, получено %d", len(m.ChurchState.Offers))
	}

	// Стабилизируем первого бойца
	press(&m, "enter")
	if m.Party[0].IsDowned {
		t.Fatal("герой должен быть стабилизирован")
	}

	// Esc переводит в Таверну
	press(&m, "esc")
	if m.TownPhase != TownPhaseTavern {
		t.Fatalf("Esc должен переводить в Tavern, текущая фаза: %v", m.TownPhase)
	}
}

func TestTavernRest(t *testing.T) {
	seedRNG(52)
	m := initialModel()
	m.ManualMode = true
	m.InTown = true
	m.TownPhase = TownPhaseTavern
	m.Gold = 200

	m.Party[0].HP = 10
	m.initTavernScreen()

	// Выбираем люксовый отдых
	press(&m, "enter")
	if m.Party[0].HP != m.Party[0].MaxHP {
		t.Fatalf("отдых в таверне должен восстановить полное HP: %d / %d", m.Party[0].HP, m.Party[0].MaxHP)
	}
	if m.TownPhase != TownPhaseGuild {
		t.Fatalf("после отдыха фаза должна стать Guild, текущая: %v", m.TownPhase)
	}
}

func TestForgeAndTanneryUpgrades(t *testing.T) {
	seedRNG(53)
	m := initialModel()
	m.ManualMode = true
	m.InTown = true
	m.TownPhase = TownPhaseSmithy
	m.Gold = 1000
	m.Legacy.TanneryLevel = 2 // разблокируем уровни сумок у кожевника

	m.initSmithyScreen()
	if len(m.SmithyState.Offers) > 0 {
		initialLvl := m.SmithyState.Offers[0].Level
		press(&m, "enter")
		if m.SmithyState.Offers[0].Level <= initialLvl && m.Party[0].GetItemInSlot(SlotWeapon).UpgradeLevel <= initialLvl {
			t.Fatal("предмет должен улучшиться после ковки")
		}
	}

	// Переход к Кожевнику
	press(&m, "esc")
	if m.TownPhase != TownPhaseTannery {
		t.Fatalf("Esc должен переводить в Tannery, текущая фаза: %v", m.TownPhase)
	}

	// Кожевник: улучшение сумки
	m.initTanneryScreen()
	initBagLvl := m.BagLevel
	press(&m, "enter")
	if m.BagLevel != initBagLvl+1 {
		t.Fatalf("уровень сумки должен вырасти: было %d, стало %d", initBagLvl, m.BagLevel)
	}

	// Переход к Алхимику
	press(&m, "esc")
	if m.TownPhase != TownPhaseAlchemist {
		t.Fatalf("Esc должен переводить в Alchemist, текущая фаза: %v", m.TownPhase)
	}
}
