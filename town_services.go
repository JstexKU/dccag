package main

import (
	"fmt"
)

// ============================================================
// 1. РЫНОК (MARKET)
// ============================================================

type MarketSection int

const (
	MarketSectionLoot MarketSection = iota
	MarketSectionShop
)

type MarketOffer struct {
	Item      EquipItem
	Cost      int
	HeroIdx   int
	Slot      EquipSlot
	StatDiff  int
	IsUpgrade bool
	IsMissing bool
}

type MarketServiceState struct {
	Section       MarketSection
	LootIdx       int
	ShopIdx       int
	Offers        []MarketOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateMarketOffers() []MarketOffer {
	var offers []MarketOffer
	slots := []EquipSlot{SlotWeapon, SlotChest, SlotHead, SlotLegs}

	for heroIdx, h := range m.Party {
		if h.IsDead {
			continue
		}

		for _, slot := range slots {
			currItem := h.GetItemInSlot(slot)
			currStat := 0
			if currItem != nil {
				currStat = currItem.TotalStat()
			}

			itemFloor := m.Floor + 1
			if currItem == nil {
				itemFloor = m.Floor
			}
			offerItem := generateItemForClassSlot(h.Class, slot, itemFloor)
			offerStat := offerItem.TotalStat()
			diff := offerStat - currStat

			cost := offerItem.Value
			if currItem == nil {
				cost = 35 + (m.Floor * 15)
			}

			isMissing := currItem == nil
			isUpgrade := diff > 0

			if isMissing || diff >= 2 {
				offers = append(offers, MarketOffer{
					Item:      offerItem,
					Cost:      cost,
					HeroIdx:   heroIdx,
					Slot:      slot,
					StatDiff:  diff,
					IsUpgrade: isUpgrade,
					IsMissing: isMissing,
				})
			}
		}
	}
	return offers
}

func (m *Model) ExecuteMarketSellLoot(idx int) bool {
	if idx < 0 || idx >= len(m.Bag) {
		return false
	}
	item := m.Bag[idx]
	price := item.Value * 2

	m.Gold += price
	m.Stats.TotalGoldEarned += price

	m.Bag = append(m.Bag[:idx], m.Bag[idx+1:]...)
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.market_sold", price)))
	m.logTownAction("⚖️", T(m.Lang, "town.market"), T(m.Lang, "town.log.market_history", price))
	return true
}

func (m *Model) ExecuteMarketSellAll() int {
	soldGold := 0
	for _, it := range m.Bag {
		soldGold += it.Value * 2
	}
	if soldGold > 0 {
		m.Gold += soldGold
		m.Stats.TotalGoldEarned += soldGold
		m.addLog(goldStyle.Render(T(m.Lang, "town.log.market_sold", soldGold)))
		m.logTownAction("⚖️", T(m.Lang, "town.market"), T(m.Lang, "town.log.market_history", soldGold))
	}
	m.Bag = []EquipItem{}

	for _, h := range m.Party {
		if !h.IsDead && !h.IsDowned {
			h.Stress = max(0, h.Stress-30)
			if h.Stress < 50 {
				h.Affliction = AfflictionNone
			}
		}
	}
	return soldGold
}

func (m *Model) ExecuteMarketBuy(offer MarketOffer) bool {
	if m.Gold < offer.Cost {
		return false
	}
	if offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	if h.IsDead {
		return false
	}

	m.Gold -= offer.Cost
	globalDebugReport.GoldSpentBreakdown["Покупка снаряжения"] += offer.Cost

	currItem := h.GetItemInSlot(offer.Slot)
	boughtItem := offer.Item

	if currItem != nil {
		if len(m.Bag) < m.currentBagCapacity() {
			m.Bag = append(m.Bag, *currItem)
		} else {
			m.Gold += currItem.Value
		}
	}

	h.SetItemInSlot(offer.Slot, &boughtItem)

	hName := h.DisplayName(m.Lang)
	itemName := boughtItem.DisplayName(m.Lang)
	slotName := T(m.Lang, "slot."+string(offer.Slot))

	if offer.IsMissing {
		m.addLog(goldStyle.Render(T(m.Lang, "town.log.bought_missing", hName, slotName, itemName, offer.Cost)))
		m.logTownAction("🛒", T(m.Lang, "town.market"), T(m.Lang, "town.log.bought_missing_hist", hName, itemName, offer.Cost))
	} else {
		m.addLog(goldStyle.Render(T(m.Lang, "town.log.bought_upgrade", hName, itemName, offer.Cost)))
		m.logTownAction("🛒", T(m.Lang, "town.market"), T(m.Lang, "town.log.bought_upgrade_hist", hName, itemName, offer.Cost))
	}
	return true
}

// ============================================================
// 2. МАГИСТРАТ (MAGISTRATE)
// ============================================================

type MagistrateTarget int

const (
	MagistrateSmithy MagistrateTarget = iota
	MagistrateTannery
	MagistrateChurch
	MagistrateTavern
	MagistrateTreasuryGrant
)

type MagistrateOffer struct {
	Target    MagistrateTarget
	NameKey   string
	Level     int
	Cost      int
	DescKey   string
	CanAfford bool
}

type MagistrateServiceState struct {
	Cursor        int
	Offers        []MagistrateOffer
	ActionSummary string
	ChosenIdx     int
}

func BuildingUpgradeCost(currentLevel int) int {
	return 150 + (currentLevel * 120)
}

func (m *Model) GenerateMagistrateOffers() []MagistrateOffer {
	offers := []MagistrateOffer{
		{
			Target:    MagistrateSmithy,
			NameKey:   m.TownEst.SmithyKey,
			Level:     m.Legacy.SmithyLevel,
			Cost:      BuildingUpgradeCost(m.Legacy.SmithyLevel),
			DescKey:   "town.smithy",
			CanAfford: m.Gold >= BuildingUpgradeCost(m.Legacy.SmithyLevel),
		},
		{
			Target:    MagistrateTannery,
			NameKey:   m.TownEst.TanneryKey,
			Level:     m.Legacy.TanneryLevel,
			Cost:      BuildingUpgradeCost(m.Legacy.TanneryLevel),
			DescKey:   "town.tannery",
			CanAfford: m.Gold >= BuildingUpgradeCost(m.Legacy.TanneryLevel),
		},
		{
			Target:    MagistrateChurch,
			NameKey:   m.TownEst.ChurchKey,
			Level:     m.Legacy.ChurchLevel,
			Cost:      BuildingUpgradeCost(m.Legacy.ChurchLevel),
			DescKey:   "town.church",
			CanAfford: m.Gold >= BuildingUpgradeCost(m.Legacy.ChurchLevel),
		},
		{
			Target:    MagistrateTavern,
			NameKey:   m.TownEst.TavernKey,
			Level:     m.Legacy.TavernLevel,
			Cost:      BuildingUpgradeCost(m.Legacy.TavernLevel),
			DescKey:   "town.tavern",
			CanAfford: m.Gold >= BuildingUpgradeCost(m.Legacy.TavernLevel),
		},
	}

	grantCost := 100
	offers = append(offers, MagistrateOffer{
		Target:    MagistrateTreasuryGrant,
		NameKey:   "town.magistrate",
		Level:     0,
		Cost:      grantCost,
		DescKey:   "stats.legacy_treasury",
		CanAfford: m.Gold >= grantCost,
	})

	return offers
}

func (m *Model) ExecuteMagistrateInvest(offer MagistrateOffer) bool {
	if m.Gold < offer.Cost {
		return false
	}

	m.Gold -= offer.Cost
	globalDebugReport.GoldSpentBreakdown["Инвестиции в Магистрат"] += offer.Cost

	prevTier := militiaTierForInvestment(m.Legacy.TotalInvested)
	m.Legacy.TotalInvested += offer.Cost

	var bldName string
	var newLevel int

	switch offer.Target {
	case MagistrateSmithy:
		m.Legacy.SmithyLevel++
		bldName = T(m.Lang, m.TownEst.SmithyKey)
		newLevel = m.Legacy.SmithyLevel
	case MagistrateTannery:
		m.Legacy.TanneryLevel++
		bldName = T(m.Lang, m.TownEst.TanneryKey)
		newLevel = m.Legacy.TanneryLevel
	case MagistrateChurch:
		m.Legacy.ChurchLevel++
		bldName = T(m.Lang, m.TownEst.ChurchKey)
		newLevel = m.Legacy.ChurchLevel
	case MagistrateTavern:
		m.Legacy.TavernLevel++
		bldName = T(m.Lang, m.TownEst.TavernKey)
		newLevel = m.Legacy.TavernLevel
	case MagistrateTreasuryGrant:
		m.Legacy.TreasuryGold += offer.Cost
		bldName = T(m.Lang, "stats.legacy_treasury")
		newLevel = m.Legacy.TreasuryGold
	}

	m.addLog(titleStyle.Render(T(m.Lang, "town.log.magistrate_tax", offer.Cost, bldName, newLevel)))
	m.logTownAction("🏛️️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.magistrate_hist", offer.Cost, bldName, newLevel))

	if newTier := militiaTierForInvestment(m.Legacy.TotalInvested); newTier.TitleKey != prevTier.TitleKey {
		tierName := T(m.Lang, newTier.TitleKey)
		m.addLog(titleStyle.Render(T(m.Lang, "town.log.militia_upgrade", tierName)))
		m.logTownAction("🎖️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.militia_upgrade_hist", tierName))
	}
	return true
}

// ============================================================
// 3. ХРАМ (CHURCH)
// ============================================================

type ChurchActionType int

const (
	ChurchActionStabilize ChurchActionType = iota
	ChurchActionCleanse
)

type ChurchOffer struct {
	Type      ChurchActionType
	HeroIdx   int
	HeroName  string
	Cost      int
	CanAfford bool
	Desc      string
}

type ChurchServiceState struct {
	Cursor        int
	Offers        []ChurchOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) ChurchBaseCost() int {
	c := 120 + (m.Floor * 30) - (m.Legacy.ChurchLevel * 15)
	if c < 60 {
		return 60
	}
	return c
}

func (m *Model) GenerateChurchOffers() []ChurchOffer {
	var offers []ChurchOffer
	baseCost := m.ChurchBaseCost()

	for idx, h := range m.Party {
		if h.IsDead {
			continue
		}
		if h.IsDowned {
			cost := baseCost
			if h.IsLeader {
				cost = 0
			}
			offers = append(offers, ChurchOffer{
				Type:      ChurchActionStabilize,
				HeroIdx:   idx,
				HeroName:  h.DisplayName(m.Lang),
				Cost:      cost,
				CanAfford: h.IsLeader || m.Gold >= cost,
				Desc:      T(m.Lang, "ui.downed"),
			})
		} else if h.Stress > 20 || h.Affliction != AfflictionNone {
			cost := baseCost / 3
			desc := fmt.Sprintf("Stress: %d", h.Stress)
			if h.Affliction != AfflictionNone {
				desc = fmt.Sprintf("Stress: %d (%s)", h.Stress, h.Affliction)
			}
			offers = append(offers, ChurchOffer{
				Type:      ChurchActionCleanse,
				HeroIdx:   idx,
				HeroName:  h.DisplayName(m.Lang),
				Cost:      cost,
				CanAfford: m.Gold >= cost,
				Desc:      desc,
			})
		}
	}
	return offers
}

func (m *Model) ExecuteChurchAction(offer ChurchOffer) bool {
	if offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	churchName := T(m.Lang, m.TownEst.ChurchKey)

	if offer.Type == ChurchActionStabilize {
		if !h.IsLeader && m.Gold < offer.Cost {
			return false
		}
		if h.IsLeader {
			m.addLog(fountStyle.Render(T(m.Lang, "leader.church_free", churchName, h.DisplayName(m.Lang))))
		} else {
			m.Gold -= offer.Cost
			globalDebugReport.GoldSpentBreakdown["Храм (стабилизация/очищение)"] += offer.Cost
		}
		h.IsDowned = false
		h.HP = h.MaxHP / 2
		h.MP = h.MaxMP / 2
		h.Stress = 70
		h.CauseOfDeath = ""
		m.Stats.Resurrections++

		m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, offer.Cost, 1)))
		m.logTownAction("⛪", churchName, T(m.Lang, "town.log.church_hist", 1, offer.Cost))
		return true
	}

	if offer.Type == ChurchActionCleanse {
		if m.Gold < offer.Cost {
			return false
		}
		m.Gold -= offer.Cost
		globalDebugReport.GoldSpentBreakdown["Храм (стабилизация/очищение)"] += offer.Cost
		h.Stress = max(0, h.Stress-60)
		h.Affliction = AfflictionNone
		m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, offer.Cost, 0)))
		return true
	}
	return false
}

// ============================================================
// 4. ТАВЕРНА (TAVERN)
// ============================================================

type TavernOffer struct {
	IsLuxury  bool
	Cost      int
	CanAfford bool
}

type TavernServiceState struct {
	Cursor        int
	Offers        []TavernOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateTavernOffers() []TavernOffer {
	luxCost := max(40, (25*m.Floor)-(m.Legacy.TavernLevel*10))
	barnCost := max(5, luxCost/4)

	return []TavernOffer{
		{IsLuxury: true, Cost: luxCost, CanAfford: m.Gold >= luxCost},
		{IsLuxury: false, Cost: barnCost, CanAfford: m.Gold >= barnCost},
	}
}

func (m *Model) ExecuteTavernRest(offer TavernOffer) bool {
	if m.Gold < offer.Cost {
		return false
	}
	tavernName := T(m.Lang, m.TownEst.TavernKey)
	m.Gold -= offer.Cost

	if offer.IsLuxury {
		globalDebugReport.GoldSpentBreakdown["Таверна (ночлег)"] += offer.Cost
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				h.HP = h.MaxHP
				h.MP = h.MaxMP
			}
		}
		m.addLog(healStyle.Render(T(m.Lang, "town.log.tavern_rest", tavernName, offer.Cost)))
		m.logTownAction("🍻", tavernName, T(m.Lang, "town.log.tavern_hist", offer.Cost))
	} else {
		globalDebugReport.GoldSpentBreakdown["Таверна (сарай)"] += offer.Cost
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				h.HP = max(1, int(float64(h.MaxHP)*0.40))
				h.MP = int(float64(h.MaxMP) * 0.40)
			}
		}
		m.addLog(dangerStyle.Render(T(m.Lang, "town.log.tavern_barn")))
		m.logTownAction("🏚️", tavernName, T(m.Lang, "town.log.tavern_barn_hist"))
	}
	return true
}

// ============================================================
// 5. ГИЛЬДИЯ (GUILD)
// ============================================================

type GuildRecruitOffer struct {
	SlotIdx   int
	Class     HeroClass
	IsVeteran bool
	Cost      int
	CanAfford bool
}

type GuildServiceState struct {
	Cursor        int
	Offers        []GuildRecruitOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateGuildOffers() []GuildRecruitOffer {
	var offers []GuildRecruitOffer
	recruitCost := max(45, 60+(m.Floor*25)-(m.Legacy.ChurchLevel*8))

	for i, h := range m.Party {
		if (h.IsDead || h.IsDowned) && !(h.IsLeader && !h.IsDead) {
			usedClasses := make(map[HeroClass]bool)
			for _, ally := range m.Party {
				if !ally.IsDead && !ally.IsDowned {
					usedClasses[ally.Class] = true
				}
			}
			var availableClasses []HeroClass
			for _, c := range AllClasses {
				if !usedClasses[c] {
					availableClasses = append(availableClasses, c)
				}
			}
			if len(availableClasses) == 0 {
				availableClasses = AllClasses
			}
			classCandidate := availableClasses[rng.Intn(len(availableClasses))]

			offers = append(offers, GuildRecruitOffer{
				SlotIdx:   i,
				Class:     classCandidate,
				IsVeteran: true,
				Cost:      recruitCost,
				CanAfford: m.Gold >= recruitCost,
			})
			offers = append(offers, GuildRecruitOffer{
				SlotIdx:   i,
				Class:     classCandidate,
				IsVeteran: false,
				Cost:      0,
				CanAfford: true,
			})
		}
	}
	return offers
}

func (m *Model) ExecuteGuildRecruit(offer GuildRecruitOffer) bool {
	if offer.SlotIdx < 0 || offer.SlotIdx >= len(m.Party) {
		return false
	}
	if offer.IsVeteran && m.Gold < offer.Cost {
		return false
	}
	guildName := T(m.Lang, m.TownEst.GuildKey)

	if offer.IsVeteran {
		m.Gold -= offer.Cost
		globalDebugReport.GoldSpentBreakdown["Найм ветеранов"] += offer.Cost
		m.Party[offer.SlotIdx] = createHero(offer.Class, m.Floor, m.Legacy.SmithyLevel)
		newHero := m.Party[offer.SlotIdx]
		m.addLog(healStyle.Render(T(m.Lang, "town.log.guild_veteran",
			guildName, newHero.DisplayName(m.Lang), newHero.RaceName(m.Lang), newHero.ShortClass(m.Lang), newHero.Level, offer.Cost)))
		m.logTownAction("⚔️", guildName, T(m.Lang, "town.log.guild_vet_hist",
			newHero.DisplayName(m.Lang), newHero.ShortClass(m.Lang), newHero.Level, offer.Cost))
	} else {
		m.Party[offer.SlotIdx] = m.createMilitiaForGuild(offer.Class)
		newHero := m.Party[offer.SlotIdx]
		tierName := T(m.Lang, newHero.TitleKey)
		m.addLog(subtleStyle.Render(T(m.Lang, "town.log.guild_militia",
			guildName, newHero.DisplayName(m.Lang), tierName, newHero.RaceName(m.Lang), newHero.ShortClass(m.Lang))))
		m.logTownAction("🤝", guildName, T(m.Lang, "town.log.guild_mil_hist",
			newHero.DisplayName(m.Lang), tierName, newHero.ShortClass(m.Lang)))
	}
	return true
}

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
}

type ForgeServiceState struct {
	Cursor        int
	Offers        []UpgradeOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateSmithyOffers() []UpgradeOffer {
	var offers []UpgradeOffer
	getCost := func(it *EquipItem) int {
		if it == nil {
			return 999999
		}
		matMult := max(1, it.Material.ValueMult)
		nextLvl := it.UpgradeLevel + 1
		cost := (nextLvl * nextLvl * 45 * matMult) - (m.Legacy.SmithyLevel * 18)
		return max(35*matMult, cost)
	}

	for hIdx, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		for _, slot := range []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs} {
			it := h.GetItemInSlot(slot)
			if it == nil || it.UpgradeLevel >= 6 {
				continue
			}

			canForge := false
			if it.Category == ArmorHeavy {
				canForge = true
			} else if it.Slot == SlotWeapon {
				switch h.Class {
				case ClassTank, ClassWarrior, ClassPaladin, ClassCleric, ClassBard:
					canForge = true
				}
			}

			if canForge {
				cost := getCost(it)
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
	}
	return offers
}

func (m *Model) ExecuteSmithyUpgrade(offer UpgradeOffer) bool {
	if m.Gold < offer.Cost || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	it := h.GetItemInSlot(offer.Slot)
	if it == nil || it.UpgradeLevel >= 6 {
		return false
	}

	m.Gold -= offer.Cost
	it.UpgradeLevel++
	h.SetItemInSlot(offer.Slot, it)
	m.Stats.UpgradesForged++
	globalDebugReport.GoldSpentBreakdown["Кузница (заточки)"] += offer.Cost

	smithyName := T(m.Lang, m.TownEst.SmithyKey)
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.smithy_done", smithyName, 1, offer.Cost)))
	m.logTownAction("⚒️", smithyName, T(m.Lang, "town.log.smithy_hist", 1, offer.Cost))
	return true
}

func (m *Model) GenerateTanneryOffers() []UpgradeOffer {
	var offers []UpgradeOffer
	getCost := func(it *EquipItem) int {
		if it == nil {
			return 999999
		}
		matMult := max(1, it.Material.ValueMult)
		nextLvl := it.UpgradeLevel + 1
		cost := (nextLvl * nextLvl * 38 * matMult) - (m.Legacy.TanneryLevel * 14)
		return max(30*matMult, cost)
	}

	for hIdx, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		for _, slot := range []EquipSlot{SlotHead, SlotChest, SlotLegs, SlotWeapon} {
			it := h.GetItemInSlot(slot)
			if it == nil || it.UpgradeLevel >= 6 {
				continue
			}

			canTan := false
			if it.Category == ArmorMedium || it.Category == ArmorLight {
				canTan = true
			} else if it.Slot == SlotWeapon {
				switch h.Class {
				case ClassRogue, ClassRanger, ClassMonk, ClassMage, ClassWarlock:
					canTan = true
				}
			}

			if canTan {
				cost := getCost(it)
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
	}
	return offers
}

func (m *Model) ExecuteTanneryUpgrade(offer UpgradeOffer) bool {
	if m.Gold < offer.Cost || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	it := h.GetItemInSlot(offer.Slot)
	if it == nil || it.UpgradeLevel >= 6 {
		return false
	}

	m.Gold -= offer.Cost
	it.UpgradeLevel++
	h.SetItemInSlot(offer.Slot, it)
	m.Stats.UpgradesForged++
	globalDebugReport.GoldSpentBreakdown["Кожевник (выделка и сумки)"] += offer.Cost

	tanneryName := T(m.Lang, m.TownEst.TanneryKey)
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_done", tanneryName, 1)))
	m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_hist", 1, offer.Cost))
	return true
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
	globalDebugReport.GoldSpentBreakdown["Улучшение сумок"] += nextBag.Cost

	tanneryName := T(m.Lang, m.TownEst.TanneryKey)
	bagName := T(m.Lang, nextBag.NameKey)
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_bag", tanneryName, bagName, nextBag.Capacity, nextBag.Cost)))
	m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_bag_hist", bagName, nextBag.Capacity, nextBag.Cost))
	return true
}

// ============================================================
// 7. ЛАВКА АЛХИМИКА (ALCHEMIST)
// ============================================================

type AlchemistActionType int

const (
	AlchemistActionMutation AlchemistActionType = iota
	AlchemistActionPotion
)

type AlchemistOffer struct {
	Type      AlchemistActionType
	HeroIdx   int
	HeroName  string
	ItemTitle string
	Cost      int
	CanAfford bool
	MutType   MutationType
	PotType   PotionType
}

type AlchemistServiceState struct {
	Cursor        int
	Offers        []AlchemistOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateAlchemistOffers() []AlchemistOffer {
	var offers []AlchemistOffer

	for idx, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		pref := GetClassMutationPreference(h, m.Floor, m.Stats.FallenHeroes)
		baseCost := 240
		switch pref {
		case MutChimera:
			baseCost = 280
		case MutBastion:
			baseCost = 260
		case MutFury:
			baseCost = 240
		case MutTitan:
			baseCost = 230
		case MutAether:
			baseCost = 220
		}
		cost := CalculateMutationCost(baseCost, m.Floor, h.Mutations.Total())
		offers = append(offers, AlchemistOffer{
			Type:      AlchemistActionMutation,
			HeroIdx:   idx,
			HeroName:  h.DisplayName(m.Lang),
			ItemTitle: T(m.Lang, "mut."+string(pref)+".name"),
			Cost:      cost,
			CanAfford: m.Gold >= cost,
			MutType:   pref,
		})

		maxPots := m.MaxPotionSlots()
		if h.HasFreePotionSlot(maxPots) {
			pType := PotionHP
			if h.Stress > 30 || h.Affliction != AfflictionNone {
				pType = PotionStress
			} else if h.Class == ClassMage || h.Class == ClassCleric || h.Class == ClassWarlock || h.Class == ClassBard {
				pType = PotionMP
			}
			cand := createPotion(pType, SizeSmall, m.Floor)
			offers = append(offers, AlchemistOffer{
				Type:      AlchemistActionPotion,
				HeroIdx:   idx,
				HeroName:  h.DisplayName(m.Lang),
				ItemTitle: fmt.Sprintf("%s (%s)", cand.Symbol, pType),
				Cost:      cand.Cost,
				CanAfford: m.Gold >= cand.Cost,
				PotType:   pType,
			})
		}
	}
	return offers
}

func (m *Model) ExecuteAlchemistOffer(offer AlchemistOffer) bool {
	if m.Gold < offer.Cost || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	m.Gold -= offer.Cost
	globalDebugReport.GoldSpentBreakdown["Алхимия и зелья"] += offer.Cost

	if offer.Type == AlchemistActionMutation {
		verbApplied := TVerb(m.Lang, h.Gender, "принял", "приняла", "took")
		hName := h.DisplayName(m.Lang)

		switch offer.MutType {
		case MutChimera:
			h.Mutations.ChimeraCount++
			h.MaxHP += 8
			h.HP += 8
			h.MaxMP += 5
			h.MP += 5
			h.BaseAtk += 1
			h.BaseDef += 1
			m.addLog(accentStyle.Render(T(m.Lang, "town.log.mut_chimera", hName, verbApplied, T(m.Lang, "mut.chimera.name"), offer.Cost)))
		case MutFury:
			h.Mutations.FuryCount++
			h.BaseAtk += 3
			h.MaxHP += 2
			h.HP += 2
			m.addLog(fireStyle.Render(T(m.Lang, "town.log.mut_fury", hName, verbApplied, T(m.Lang, "mut.fury.name"), offer.Cost)))
		case MutTitan:
			h.Mutations.TitanCount++
			h.MaxHP += 20
			h.HP += 20
			m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_titan", hName, verbApplied, T(m.Lang, "mut.titan.name"), offer.Cost)))
		case MutAether:
			h.Mutations.AetherCount++
			h.MaxMP += 14
			h.MP += 14
			h.BaseAtk += 1
			m.addLog(fountStyle.Render(T(m.Lang, "town.log.mut_aether", hName, verbApplied, T(m.Lang, "mut.aether.name"), offer.Cost)))
		case MutBastion:
			h.Mutations.BastionCount++
			h.BaseDef += 2
			h.MaxHP += 6
			h.HP += 6
			m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_bastion", hName, verbApplied, T(m.Lang, "mut.bastion.name"), offer.Cost)))
		}
		return true
	}

	if offer.Type == AlchemistActionPotion {
		pot := createPotion(offer.PotType, SizeSmall, m.Floor)
		h.Potions = append(h.Potions, &pot)
		m.logTownAction("🧪", T(m.Lang, m.TownEst.AlchemistKey), T(m.Lang, "town.log.alch_hist", offer.Cost))
		return true
	}
	return false
}
