package main

// ============================================================
// 6. КУЗНИЦА И КОЖЕВНИК (FORGE & TANNERY)
// ============================================================

type UpgradeOffer struct {
	HeroIdx   int
	Slot      EquipSlot
	ItemName  string
	Level     int
	Cost      int
	CanAfford bool

	// Зачарование оружия (enchant.go): Level — текущая ступень, EnchantLevel — предлагаемая.
	Enchant      bool
	EnchantKind  EnchantKind
	EnchantLevel int
	Replaces     string // метка заменяемого зачарования, если вид меняется
}

type ForgeServiceState struct {
	Cursor        int
	Offers        []UpgradeOffer
	ActionSummary string
	ChosenIdx     int
	EnchantKind   EnchantKind // выбранный игроком вид зачарования (ручной режим кузницы)
}

// upgradeService описывает мастерскую улучшения снаряжения. Кузница и кожевник — две
// конфигурации одного механизма: различаются порядком слотов, ценой, допуском предметов
// и записями в журнале.
type upgradeService struct {
	slots   []EquipSlot                      // в каком порядке обходить слоты героя
	cost    func(m *Model, it *EquipItem) int // цена следующего уровня предмета
	accepts func(h *Hero, it *EquipItem) bool // берётся ли мастерская за этот предмет
	spend   string                            // категория в отчёте о тратах
	logDone func(m *Model, offer UpgradeOffer)
}

// upgradeCost — цена следующего уровня: квадратичный рост, умноженный на материал,
// за вычетом скидки от развития здания, но не ниже floorPerMat за единицу материала.
func upgradeCost(it *EquipItem, basePerMat, floorPerMat, discount int) int {
	matMult := max(1, it.Material.ValueMult)
	nextLvl := it.UpgradeLevel + 1
	cost := (nextLvl * nextLvl * basePerMat * matMult) - discount
	return max(floorPerMat*matMult, cost)
}

func smithyService() upgradeService {
	return upgradeService{
		slots: []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs},
		cost: func(m *Model, it *EquipItem) int {
			return upgradeCost(it, 45, 35, m.Legacy.SmithyLevel*18)
		},
		accepts: func(h *Hero, it *EquipItem) bool {
			if it.Category == ArmorHeavy {
				return true
			}
			if it.Slot == SlotWeapon {
				switch h.Class {
				case ClassTank, ClassWarrior, ClassPaladin, ClassCleric, ClassBard:
					return true
				}
			}
			return false
		},
		spend: spendForge,
		logDone: func(m *Model, offer UpgradeOffer) {
			smithyName := T(m.Lang, m.TownEst.SmithyKey)
			m.addLog(goldStyle.Render(T(m.Lang, "town.log.smithy_done", smithyName, 1, offer.Cost)))
			m.logTownAction("⚒️", smithyName, T(m.Lang, "town.log.smithy_hist", 1, offer.Cost))
		},
	}
}

func tanneryService() upgradeService {
	return upgradeService{
		slots: []EquipSlot{SlotHead, SlotChest, SlotLegs, SlotWeapon},
		cost: func(m *Model, it *EquipItem) int {
			return upgradeCost(it, 38, 30, m.Legacy.TanneryLevel*14)
		},
		accepts: func(h *Hero, it *EquipItem) bool {
			if it.Category == ArmorMedium || it.Category == ArmorLight {
				return true
			}
			if it.Slot == SlotWeapon {
				switch h.Class {
				case ClassRogue, ClassRanger, ClassMonk, ClassMage, ClassWarlock:
					return true
				}
			}
			return false
		},
		spend: spendTannery,
		logDone: func(m *Model, offer UpgradeOffer) {
			tanneryName := T(m.Lang, m.TownEst.TanneryKey)
			m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_done", tanneryName, 1)))
			m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_hist", 1, offer.Cost))
		},
	}
}

// generateUpgradeOffers перечисляет улучшения, которые мастерская может предложить живым героям.
func (m *Model) generateUpgradeOffers(svc upgradeService) []UpgradeOffer {
	var offers []UpgradeOffer
	for hIdx, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		for _, slot := range svc.slots {
			it := h.GetItemInSlot(slot)
			if it == nil || it.UpgradeLevel >= maxUpgradeLevel || !svc.accepts(h, it) {
				continue
			}
			cost := svc.cost(m, it)
			offers = append(offers, UpgradeOffer{
				HeroIdx:   hIdx,
				Slot:      slot,
				ItemName:  it.DisplayName(m.Lang),
				Level:     it.UpgradeLevel,
				Cost:      cost,
				CanAfford: m.Gold >= cost,
			})
		}
	}
	return offers
}

// executeUpgrade проводит покупку улучшения и возвращает false, если она невозможна.
func (m *Model) executeUpgrade(svc upgradeService, offer UpgradeOffer) bool {
	if m.Gold < offer.Cost || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	it := h.GetItemInSlot(offer.Slot)
	if it == nil || it.UpgradeLevel >= maxUpgradeLevel {
		return false
	}

	m.Gold -= offer.Cost
	it.UpgradeLevel++
	h.SetItemInSlot(offer.Slot, it)
	m.Stats.UpgradesForged++
	globalDebugReport.GoldSpentBreakdown[svc.spend] += offer.Cost

	svc.logDone(m, offer)
	return true
}

func (m *Model) GenerateSmithyOffers() []UpgradeOffer {
	offers := m.generateUpgradeOffers(smithyService())
	return append(offers, m.generateEnchantOffers()...)
}

func (m *Model) ExecuteSmithyUpgrade(offer UpgradeOffer) bool {
	if offer.Enchant {
		return m.executeEnchant(offer)
	}
	return m.executeUpgrade(smithyService(), offer)
}

func (m *Model) GenerateTanneryOffers() []UpgradeOffer {
	return m.generateUpgradeOffers(tanneryService())
}

func (m *Model) ExecuteTanneryUpgrade(offer UpgradeOffer) bool {
	return m.executeUpgrade(tanneryService(), offer)
}

func (m *Model) ExecuteBagUpgrade() bool {
	if m.BagLevel >= len(bagUpgrades)-1 {
		return false
	}
	nextBag := bagUpgrades[m.BagLevel+1]
	maxAllowedTier := m.Legacy.TanneryLevel + 1
	if nextBag.Level > maxAllowedTier || m.Gold < nextBag.Cost {
		return false
	}

	m.Gold -= nextBag.Cost
	m.BagLevel++
	globalDebugReport.GoldSpentBreakdown[spendBags] += nextBag.Cost

	tanneryName := T(m.Lang, m.TownEst.TanneryKey)
	bagName := T(m.Lang, nextBag.NameKey)
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_bag", tanneryName, bagName, nextBag.Capacity, nextBag.Cost)))
	m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_bag_hist", bagName, nextBag.Capacity, nextBag.Cost))
	return true
}
