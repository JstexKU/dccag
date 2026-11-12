package main

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
		globalDebugReport.GoldSpentBreakdown[spendTavernInn] += offer.Cost
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				h.HP = h.MaxHP
				h.MP = h.MaxMP
			}
		}
		m.addLog(healStyle.Render(T(m.Lang, "town.log.tavern_rest", tavernName, offer.Cost)))
		m.logTownAction("🍻", tavernName, T(m.Lang, "town.log.tavern_hist", offer.Cost))
	} else {
		globalDebugReport.GoldSpentBreakdown[spendTavernBarn] += offer.Cost
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				// Сарай даёт отдых, но не отнимает силы у тех, кто уже в порядке.
				h.HP = max(h.HP, max(1, int(float64(h.MaxHP)*0.40)))
				h.MP = max(h.MP, int(float64(h.MaxHP)*0.40))
			}
		}
		m.addLog(dangerStyle.Render(T(m.Lang, "town.log.tavern_barn")))
		m.logTownAction("🏚️", tavernName, T(m.Lang, "town.log.tavern_barn_hist"))
	}
	return true
}
