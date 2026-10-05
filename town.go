package main

import (
	"fmt"
)

func (m *Model) stepTown() {
	switch m.TownPhase {
	case TownPhaseSellLoot:
		m.initMarketScreen()
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
			m.MarketState.ActionSummary = fmt.Sprintf("Продано: %s (+%dG)", item.DisplayName(m.Lang), price)
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

		m.MarketState.ActionSummary = "Рюкзак пуст. Снаряжение отряда в порядке."
		m.TownPhase = TownPhaseMagistrate

	case TownPhaseMagistrate:
		m.initMagistrateScreen()
		if m.ManualMode {
			return
		}

		budget := AllocateBudget(m.Gold).Treasury
		offers := m.MagistrateState.Offers

		// Ищем доступное здание с минимальным уровнем
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
			m.MagistrateState.ActionSummary = fmt.Sprintf("Инвестировано %dG в «%s» (Ур.%d)", targetOffer.Cost, bldName, targetOffer.Level+1)
			return
		}

		m.MagistrateState.ActionSummary = "Казна распределена. Инвестиции в Магистрате завершены."
		m.TownPhase = TownPhaseChurch

	case TownPhaseChurch:
		m.initChurchScreen()
		if m.ManualMode {
			return
		}

		// Обработка брошенных в бездне тел выжившими
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

		// Пошаговая стабилизация и очищение
		offers := m.ChurchState.Offers
		for i, off := range offers {
			if off.CanAfford {
				m.ChurchState.Cursor = i
				m.ChurchState.ChosenIdx = i

				m.ExecuteChurchAction(off)
				if off.Type == ChurchActionStabilize {
					m.ChurchState.ActionSummary = fmt.Sprintf("Стабилизирован тяжелораненый боец %s", off.HeroName)
				} else {
					m.ChurchState.ActionSummary = fmt.Sprintf("Очищен разум бойца %s (-60 стресса)", off.HeroName)
				}
				return
			}
		}

		m.ChurchState.ActionSummary = "Отряд благословлен, раненых нет."
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		m.initTavernScreen()
		if m.ManualMode {
			return
		}

		offers := m.TavernState.Offers
		// Выбираем лучший доступный ночлег
		chosenIdx := -1
		if offers[0].CanAfford {
			chosenIdx = 0
		} else if offers[1].CanAfford {
			chosenIdx = 1
		}

		if chosenIdx != -1 && m.TavernState.ChosenIdx == -1 {
			m.TavernState.Cursor = chosenIdx
			m.TavernState.ChosenIdx = chosenIdx

			m.ExecuteTavernRest(offers[chosenIdx])
			if chosenIdx == 0 {
				m.TavernState.ActionSummary = fmt.Sprintf("Отряд отдохнул в комфортных покоях (-%dG, 100%% HP/MP)", offers[0].Cost)
			} else {
				m.TavernState.ActionSummary = "Казна пуста. Ночлег на сеновале у очага (40% сил)."
			}
			return
		}

		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
		m.initGuildScreen()
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
		// Пошаговый найм на свободные вакансии
		for i, off := range offers {
			// Предпочитаем ветерана, если хватает денег, иначе берем ополчение
			if off.IsVeteran && off.CanAfford {
				m.GuildState.Cursor = i
				m.GuildState.ChosenIdx = i
				m.ExecuteGuildRecruit(off)
				m.GuildState.ActionSummary = fmt.Sprintf("Нанят ветеран %s (%s, -%dG)", m.Party[off.SlotIdx].DisplayName(m.Lang), off.Class, off.Cost)
				return
			} else if !off.IsVeteran {
				m.GuildState.Cursor = i
				m.GuildState.ChosenIdx = i
				m.ExecuteGuildRecruit(off)
				m.GuildState.ActionSummary = fmt.Sprintf("Призван ополченец %s (%s, бесплатно)", m.Party[off.SlotIdx].DisplayName(m.Lang), off.Class)
				return
			}
		}

		m.GuildState.ActionSummary = "Все позиции укомплектованы. Отряд полон."
		m.TownPhase = TownPhaseSmithy

	case TownPhaseSmithy:
		m.initSmithyScreen()
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
			m.SmithyState.ActionSummary = fmt.Sprintf("Заточено: %s (+%d) для %s (-%dG)",
				targetOffer.ItemName, targetOffer.Level+1, m.Party[targetOffer.HeroIdx].DisplayName(m.Lang), targetOffer.Cost)
			return
		}

		m.SmithyState.ActionSummary = "Лимит бюджета исчерпан или экипировка максимально заточена."
		m.TownPhase = TownPhaseTannery

	case TownPhaseTannery:
		m.initTanneryScreen()
		if m.ManualMode {
			return
		}

		budget := AllocateBudget(m.Gold)
		tanneryBudget := budget.Tannery + budget.Bags
		offers := m.TanneryState.Offers

		// 1. Проверяем пошив сумки
		if m.BagLevel < len(bagUpgrades)-1 {
			nextBag := bagUpgrades[m.BagLevel+1]
			maxAllowedTier := m.Legacy.TanneryLevel + 1
			if nextBag.Level <= maxAllowedTier && tanneryBudget >= nextBag.Cost && m.Gold >= nextBag.Cost && m.TanneryState.ChosenIdx != 0 {
				m.TanneryState.Cursor = 0
				m.TanneryState.ChosenIdx = 0
				m.ExecuteBagUpgrade()
				bagName := T(m.Lang, nextBag.NameKey)
				m.TanneryState.ActionSummary = fmt.Sprintf("Сшита %s (%d слотов, -%dG)", bagName, nextBag.Capacity, nextBag.Cost)
				return
			}
		}

		// 2. Проверяем выделку легких/средних доспехов
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
			m.TanneryState.ActionSummary = fmt.Sprintf("Выделка: %s (+%d) для %s (-%dG)",
				targetOffer.ItemName, targetOffer.Level+1, m.Party[targetOffer.HeroIdx].DisplayName(m.Lang), targetOffer.Cost)
			return
		}

		m.TanneryState.ActionSummary = "Сумка расширена, доспехи в отличном состоянии."
		m.TownPhase = TownPhaseAlchemist

	case TownPhaseAlchemist:
		m.initAlchemistScreen()
		if m.ManualMode {
			return
		}

		budget := AllocateBudget(m.Gold).Alchemy
		offers := m.AlchemistState.Offers

		// 1. Ищем подходящие мутации
		for i, off := range offers {
			if off.Type == AlchemistActionMutation && off.CanAfford && budget >= off.Cost {
				m.AlchemistState.Cursor = i
				m.AlchemistState.ChosenIdx = i
				m.ExecuteAlchemistOffer(off)
				m.AlchemistState.ActionSummary = fmt.Sprintf("Принята мутация: %s для %s (-%dG)", off.ItemTitle, off.HeroName, off.Cost)
				return
			}
		}

		// 2. Ищем докупку зелий
		for i, off := range offers {
			if off.Type == AlchemistActionPotion && off.CanAfford && budget >= off.Cost {
				m.AlchemistState.Cursor = i
				m.AlchemistState.ChosenIdx = i
				m.ExecuteAlchemistOffer(off)
				m.AlchemistState.ActionSummary = fmt.Sprintf("Куплено зелье %s для %s (-%dG)", off.ItemTitle, off.HeroName, off.Cost)
				return
			}
		}

		m.AlchemistState.ActionSummary = "Пояса укомплектованы зельями. Мутации не требуются."
		m.TownPhase = TownPhaseDepart

	case TownPhaseDepart:
		for _, h := range m.Party {
			h.ReviveCooldown = 0
		}
		m.TownHistory = []string{}
		m.InTown = false
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		m.addLog(accentStyle.Render(T(m.Lang, "town.log.depart")))
	}
}

