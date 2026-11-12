package main

// ============================================================
// 1. РЫНОК (MARKET)
// ============================================================

type MarketSection int

const (
	MarketSectionLoot MarketSection = iota
	MarketSectionShop
)

const MarketLootPageSize = 6 // Количество одновременно видимых слотов лута в окне рынка

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
	Section          MarketSection
	LootIdx          int
	LootScrollOffset int // Смещение для прокрутки длинного инвентаря
	ShopIdx          int
	Offers           []MarketOffer
	ActionSummary    string
	ChosenIdx        int
}

// AdjustLootScroll удерживает выбранный LootIdx внутри видимого скользящего окна MarketLootPageSize
func (s *MarketServiceState) AdjustLootScroll(totalItems int) {
	if totalItems <= MarketLootPageSize {
		s.LootScrollOffset = 0
		return
	}
	if s.LootIdx < s.LootScrollOffset {
		s.LootScrollOffset = s.LootIdx
	} else if s.LootIdx >= s.LootScrollOffset+MarketLootPageSize {
		s.LootScrollOffset = s.LootIdx - MarketLootPageSize + 1
	}
	maxOffset := totalItems - MarketLootPageSize
	if s.LootScrollOffset > maxOffset {
		s.LootScrollOffset = maxOffset
	}
	if s.LootScrollOffset < 0 {
		s.LootScrollOffset = 0
	}
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

	// Корректируем индекс и скролл после продажи
	if m.MarketState.LootIdx >= len(m.Bag) {
		m.MarketState.LootIdx = max(0, len(m.Bag)-1)
	}
	m.MarketState.AdjustLootScroll(len(m.Bag))
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
	m.MarketState.LootIdx = 0
	m.MarketState.LootScrollOffset = 0

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
	globalDebugReport.GoldSpentBreakdown[spendGear] += offer.Cost

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
