package main

func (m *Model) stepTown() {
	switch m.TownPhase {
	case TownPhaseSellLoot:
		if len(m.MarketState.Offers) == 0 && len(m.Bag) > 0 && m.MarketState.ActionSummary == "" {
			m.initMarketScreen()
		}
		if m.ManualMode {
			return
		}

		// 1. Пошаговая продажа добычи из рюкзака
		if len(m.Bag) > 0 {
			lastIdx := len(m.Bag) - 1
			item := m.Bag[lastIdx]
			price := item.Value * 2

			m.MarketState.Section = MarketSectionLoot
			m.MarketState.LootIdx = lastIdx
			m.MarketState.ChosenIdx = lastIdx

			m.ExecuteMarketSellLoot(lastIdx)
			m.MarketState.ActionSummary = T(m.Lang, "town.sum.market_sold", item.DisplayName(m.Lang), price)
			return
		}

		// 2. Снятие стресса после продажи
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				h.Stress = max(0, h.Stress-30)
				if h.Stress < 50 {
					h.Affliction = AfflictionNone
				}
			}
		}

		m.MarketState.ActionSummary = T(m.Lang, "town.sum.market_empty")
		m.TownPhase = TownPhaseMagistrate

	case TownPhaseMagistrate:
		if len(m.MagistrateState.Offers) == 0 && m.MagistrateState.ActionSummary == "" {
			m.initMagistrateScreen()
		}
		if m.ManualMode {
			return
		}

		budget := AllocateBudget(m.Gold).Treasury
		offers := m.MagistrateState.Offers

		bestIdx := -1
		minLevel := 999999
		for i, off := range offers {
			if off.CanAfford && budget >= off.Cost && off.Target != MagistrateTreasuryGrant {
				if off.Level < minLevel {
					minLevel = off.Level
					bestIdx = i
				}
			}
		}

		if bestIdx != -1 {
			targetOffer := offers[bestIdx]
			m.MagistrateState.Cursor = bestIdx
			m.MagistrateState.ChosenIdx = bestIdx

			m.ExecuteMagistrateInvest(targetOffer)
			bldName := T(m.Lang, targetOffer.NameKey)
			m.MagistrateState.ActionSummary = T(m.Lang, "town.sum.magistrate_invest", targetOffer.Cost, bldName, targetOffer.Level+1)
			m.MagistrateState.Offers = m.GenerateMagistrateOffers()
			return
		}

		m.MagistrateState.ActionSummary = T(m.Lang, "town.sum.magistrate_done")
		m.TownPhase = TownPhaseChurch

	case TownPhaseChurch:
		if len(m.ChurchState.Offers) == 0 && m.ChurchState.ActionSummary == "" {
			m.initChurchScreen()
		}
		if m.ManualMode {
			return
		}

		// Обработка брошенных в бездне тел
		var livingHeroes []*Hero
		var downedHeroes []*Hero
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				livingHeroes = append(livingHeroes, h)
			} else if h.IsDowned && !h.IsDead {
				downedHeroes = append(downedHeroes, h)
			}
		}

		carryCapacity := len(livingHeroes)
		for _, h := range downedHeroes {
			if carryCapacity <= 0 && !h.IsLeader {
				h.IsDead = true
				h.IsDowned = false
				h.LostInAbyss = true
				h.CauseOfDeath = T(m.Lang, "combat.log.left_in_abyss")
				m.recordFallenHero(h)
				continue
			}
			carryCapacity--
		}

		offers := m.ChurchState.Offers
		for i, off := range offers {
			if off.CanAfford {
				m.ChurchState.Cursor = i
				m.ChurchState.ChosenIdx = i

				m.ExecuteChurchAction(off)
				if off.Type == ChurchActionStabilize {
					m.ChurchState.ActionSummary = T(m.Lang, "town.sum.church_stabilize", off.HeroName)
				} else {
					m.ChurchState.ActionSummary = T(m.Lang, "town.sum.church_cleanse", off.HeroName)
				}
				m.ChurchState.Offers = m.GenerateChurchOffers()
				return
			}
		}

		m.ChurchState.ActionSummary = T(m.Lang, "town.sum.church_done")
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		if len(m.TavernState.Offers) == 0 && m.TavernState.ActionSummary == "" {
			m.initTavernScreen()
		}
		if m.ManualMode {
			return
		}

		offers := m.TavernState.Offers
		if len(offers) == 0 {
			m.initTavernScreen()
			offers = m.TavernState.Offers
		}

		if m.TavernState.ActionSummary != "" {
			m.TownPhase = TownPhaseGuild
			return
		}

		chosenIdx := -1
		if offers[0].CanAfford {
			chosenIdx = 0
		} else if offers[1].CanAfford {
			chosenIdx = 1
		}

		if chosenIdx != -1 {
			m.TavernState.Cursor = chosenIdx
			m.TavernState.ChosenIdx = chosenIdx

			m.ExecuteTavernRest(offers[chosenIdx])
			if chosenIdx == 0 {
				m.TavernState.ActionSummary = T(m.Lang, "town.sum.tavern_inn", offers[0].Cost)
			} else {
				m.TavernState.ActionSummary = T(m.Lang, "town.sum.tavern_barn")
			}
			return
		}

		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
		if len(m.GuildState.Offers) == 0 && m.GuildState.ActionSummary == "" {
			m.initGuildScreen()
		}
		if m.ManualMode {
			return
		}

		guildName := T(m.Lang, m.TownEst.GuildKey)
		if m.CurrentQuest.Completed {
			reward := m.CurrentQuest.RewardGold
			m.Gold += reward
			m.Stats.TotalGoldEarned += reward
			m.Stats.QuestsCompleted++
			m.addLog(questStyle.Render(T(m.Lang, "town.log.guild_quest", guildName, reward)))
			m.logTownAction("📜", guildName, T(m.Lang, "town.log.guild_quest_hist", reward))
			m.CurrentQuest = generateAutoQuest(m.Floor)
		}

		offers := m.GuildState.Offers
		for i, off := range offers {
			if off.IsVeteran && off.CanAfford {
				m.GuildState.Cursor = i
				m.GuildState.ChosenIdx = i
				m.ExecuteGuildRecruit(off)
				m.GuildState.ActionSummary = T(m.Lang, "town.sum.guild_hire", m.Party[off.SlotIdx].DisplayName(m.Lang), T(m.Lang, "class."+string(off.Class)+".name"), off.Cost)
				m.GuildState.Offers = m.GenerateGuildOffers()
				return
			} else if !off.IsVeteran {
				m.GuildState.Cursor = i
				m.GuildState.ChosenIdx = i
				m.ExecuteGuildRecruit(off)
				m.GuildState.ActionSummary = T(m.Lang, "town.sum.guild_militia", m.Party[off.SlotIdx].DisplayName(m.Lang), T(m.Lang, "class."+string(off.Class)+".name"))
				m.GuildState.Offers = m.GenerateGuildOffers()
				return
			}
		}

		m.GuildState.ActionSummary = T(m.Lang, "town.sum.guild_done")
		m.TownPhase = TownPhaseSmithy

	case TownPhaseSmithy:
		if len(m.SmithyState.Offers) == 0 && m.SmithyState.ActionSummary == "" {
			m.initSmithyScreen()
		}
		if m.ManualMode {
			return
		}

		offers := m.SmithyState.Offers
		budget := AllocateBudget(m.Gold).Forge

		bestIdx := -1
		minCost := 999999
		for i, off := range offers {
			if off.CanAfford && budget >= off.Cost && off.Cost < minCost {
				minCost = off.Cost
				bestIdx = i
			}
		}

		if bestIdx != -1 {
			targetOffer := offers[bestIdx]
			m.SmithyState.Cursor = bestIdx
			m.SmithyState.ChosenIdx = bestIdx

			m.ExecuteSmithyUpgrade(targetOffer)
			m.SmithyState.ActionSummary = T(m.Lang, "town.sum.smithy_upgrade",
				targetOffer.ItemName, targetOffer.Level+1, m.Party[targetOffer.HeroIdx].DisplayName(m.Lang), targetOffer.Cost)
			m.SmithyState.Offers = m.GenerateSmithyOffers()
			return
		}

		m.SmithyState.ActionSummary = T(m.Lang, "town.sum.smithy_done")
		m.TownPhase = TownPhaseTannery

	case TownPhaseTannery:
		if len(m.TanneryState.Offers) == 0 && m.TanneryState.ActionSummary == "" {
			m.initTanneryScreen()
		}
		if m.ManualMode {
			return
		}

		// 1. Пошив сумки
		if m.BagLevel < len(bagUpgrades)-1 {
			nextBag := bagUpgrades[m.BagLevel+1]
			maxAllowedTier := m.Legacy.TanneryLevel + 1
			if nextBag.Level <= maxAllowedTier && m.Gold >= nextBag.Cost && m.TanneryState.ChosenIdx != 999 {
				m.TanneryState.Cursor = 0
				m.TanneryState.ChosenIdx = 999
				if m.ExecuteBagUpgrade() {
					bagName := T(m.Lang, nextBag.NameKey)
					m.TanneryState.ActionSummary = T(m.Lang, "town.sum.tannery_bag", bagName, nextBag.Capacity, nextBag.Cost)
					m.TanneryState.Offers = m.GenerateTanneryOffers()
					return
				}
			}
		}

		budget := AllocateBudget(m.Gold)
		tanneryBudget := budget.Tannery + budget.Bags
		offers := m.TanneryState.Offers

		// 2. Выделка доспехов
		bestIdx := -1
		minCost := 999999
		for i, off := range offers {
			if off.CanAfford && tanneryBudget >= off.Cost && off.Cost < minCost {
				minCost = off.Cost
				bestIdx = i
			}
		}

		if bestIdx != -1 {
			targetOffer := offers[bestIdx]
			offset := 0
			if m.BagLevel < len(bagUpgrades)-1 {
				offset = 1
			}
			m.TanneryState.Cursor = offset + bestIdx
			m.TanneryState.ChosenIdx = offset + bestIdx

			m.ExecuteTanneryUpgrade(targetOffer)
			m.TanneryState.ActionSummary = T(m.Lang, "town.sum.tannery_upgrade",
				targetOffer.ItemName, targetOffer.Level+1, m.Party[targetOffer.HeroIdx].DisplayName(m.Lang), targetOffer.Cost)
			m.TanneryState.Offers = m.GenerateTanneryOffers()
			return
		}

		m.TanneryState.ActionSummary = T(m.Lang, "town.sum.tannery_done")
		m.TownPhase = TownPhaseAlchemist

	case TownPhaseAlchemist:
		if len(m.AlchemistState.Offers) == 0 && m.AlchemistState.ActionSummary == "" {
			m.initAlchemistScreen()
		}
		if m.ManualMode {
			return
		}

		budget := AllocateBudget(m.Gold).Alchemy
		offers := m.AlchemistState.Offers

		// 1. Честное распределение мутаций: ищем доступного героя с НАИМЕНЬШИМ числом мутаций
		bestMutIdx := -1
		minMutations := 999999
		for i, off := range offers {
			if off.Type == AlchemistActionMutation && off.CanAfford && budget >= off.Cost {
				if off.HeroIdx >= 0 && off.HeroIdx < len(m.Party) {
					heroMutCount := m.Party[off.HeroIdx].Mutations.Total()
					if heroMutCount < minMutations {
						minMutations = heroMutCount
						bestMutIdx = i
					}
				}
			}
		}

		if bestMutIdx != -1 {
			targetOffer := offers[bestMutIdx]
			m.AlchemistState.Cursor = bestMutIdx
			m.AlchemistState.ChosenIdx = bestMutIdx
			m.ExecuteAlchemistOffer(targetOffer)
			m.AlchemistState.ActionSummary = T(m.Lang, "town.sum.alchemist_mutation", targetOffer.ItemTitle, targetOffer.HeroName, targetOffer.Cost)
			m.AlchemistState.Offers = m.GenerateAlchemistOffers()
			return
		}

		// 2. Честное распределение зелий: отдаем зелье тому, у кого меньше всего зелий в поясе
		bestPotIdx := -1
		minPotions := 999999
		for i, off := range offers {
			if off.Type == AlchemistActionPotion && off.CanAfford && budget >= off.Cost {
				if off.HeroIdx >= 0 && off.HeroIdx < len(m.Party) {
					heroPotCount := len(m.Party[off.HeroIdx].Potions)
					if heroPotCount < minPotions {
						minPotions = heroPotCount
						bestPotIdx = i
					}
				}
			}
		}

		if bestPotIdx != -1 {
			targetOffer := offers[bestPotIdx]
			m.AlchemistState.Cursor = bestPotIdx
			m.AlchemistState.ChosenIdx = bestPotIdx
			m.ExecuteAlchemistOffer(targetOffer)
			m.AlchemistState.ActionSummary = T(m.Lang, "town.sum.alchemist_potion", targetOffer.ItemTitle, targetOffer.HeroName, targetOffer.Cost)
			m.AlchemistState.Offers = m.GenerateAlchemistOffers()
			return
		}

		m.AlchemistState.ActionSummary = T(m.Lang, "town.sum.alchemist_done")
		m.TownPhase = TownPhaseDepart

	case TownPhaseDepart:
		m.departFromTown(m.planDeparture())
	}
}