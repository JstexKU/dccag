package main

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
			desc := T(m.Lang, "church.desc.stress", h.Stress)
			if h.Affliction != AfflictionNone {
				desc = T(m.Lang, "church.desc.stress_affliction", h.Stress, T(m.Lang, "affliction."+string(h.Affliction)))
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
			globalDebugReport.GoldSpentBreakdown[spendChurch] += offer.Cost
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
		globalDebugReport.GoldSpentBreakdown[spendChurch] += offer.Cost
		h.Stress = max(0, h.Stress-60)
		h.Affliction = AfflictionNone
		m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, offer.Cost, 0)))
		return true
	}
	return false
}
