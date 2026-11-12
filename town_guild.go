package main

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

	usedClasses := make(map[HeroClass]bool)
	for _, ally := range m.Party {
		if !ally.IsDead && !ally.IsDowned && !ally.LostInAbyss {
			usedClasses[ally.Class] = true
		}
	}

	for i, h := range m.Party {
		if h.IsDead || h.LostInAbyss || (h.IsDowned && !h.IsLeader) {
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
		globalDebugReport.GoldSpentBreakdown[spendRecruit] += offer.Cost
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
