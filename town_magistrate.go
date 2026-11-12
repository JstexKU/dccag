package main

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
	globalDebugReport.GoldSpentBreakdown[spendMagistrate] += offer.Cost

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
	m.logTownAction("🏛️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.magistrate_hist", offer.Cost, bldName, newLevel))

	if newTier := militiaTierForInvestment(m.Legacy.TotalInvested); newTier.TitleKey != prevTier.TitleKey {
		tierName := T(m.Lang, newTier.TitleKey)
		m.addLog(titleStyle.Render(T(m.Lang, "town.log.militia_upgrade", tierName)))
		m.logTownAction("🎖️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.militia_upgrade_hist", tierName))
	}
	return true
}
