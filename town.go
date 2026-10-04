package main

import (
	"sort"
)

func (m *Model) stepTown() {
	switch m.TownPhase {
	case TownPhaseSellLoot:
		// Сначала покупаем недостающие или лучшие вещи
		m.buyMissingOrBetterGear()

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
		m.TownPhase = TownPhaseMagistrate

	case TownPhaseMagistrate:
		budget := AllocateBudget(m.Gold)
		investAmt := budget.Treasury
		if investAmt > 0 {
			m.Gold -= investAmt
			globalDebugReport.GoldSpentBreakdown["Инвестиции в Магистрат"] += investAmt

			// Параллельно копим мета-вложения в столицу (прогрессия ополчения).
			prevTier := militiaTierForInvestment(m.Legacy.TotalInvested)
			m.Legacy.TotalInvested += investAmt

			type Building struct {
				Key   string
				Level int
			}
			buildings := []Building{
				{Key: m.TownEst.SmithyKey, Level: m.Legacy.SmithyLevel},
				{Key: m.TownEst.TanneryKey, Level: m.Legacy.TanneryLevel},
				{Key: m.TownEst.ChurchKey, Level: m.Legacy.ChurchLevel},
				{Key: m.TownEst.TavernKey, Level: m.Legacy.TavernLevel},
			}
			sort.Slice(buildings, func(i, j int) bool { return buildings[i].Level < buildings[j].Level })

			targetBld := &buildings[0]
			switch targetBld.Key {
			case m.TownEst.SmithyKey:
				m.Legacy.SmithyLevel++
			case m.TownEst.TanneryKey:
				m.Legacy.TanneryLevel++
			case m.TownEst.ChurchKey:
				m.Legacy.ChurchLevel++
			case m.TownEst.TavernKey:
				m.Legacy.TavernLevel++
			}
			bldName := T(m.Lang, targetBld.Key)
			m.addLog(titleStyle.Render(T(m.Lang, "town.log.magistrate_tax", investAmt, bldName, targetBld.Level+1)))
			m.logTownAction("🏛️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.magistrate_hist", investAmt, bldName, targetBld.Level+1))

			// Если тир ополчения поднялся — торжественный лог.
			if newTier := militiaTierForInvestment(m.Legacy.TotalInvested); newTier.TitleKey != prevTier.TitleKey {
				tierName := T(m.Lang, newTier.TitleKey)
				m.addLog(titleStyle.Render(T(m.Lang, "town.log.militia_upgrade", tierName)))
				m.logTownAction("🎖️", T(m.Lang, "town.magistrate"), T(m.Lang, "town.log.militia_upgrade_hist", tierName))
			}
		}
		m.TownPhase = TownPhaseChurch

	case TownPhaseChurch:
		churchName := T(m.Lang, m.TownEst.ChurchKey)
		healBaseCost := 120 + (m.Floor * 30) - (m.Legacy.ChurchLevel * 15)
		if healBaseCost < 60 {
			healBaseCost = 60
		}

		// 1. Проверяем эвакуацию тяжелораненых (IsDowned): 1 живой на ногах выносит 1 бойца
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
		if len(downedHeroes) > 0 {
			// Сортировка поверженных по ценности
			sort.Slice(downedHeroes, func(i, j int) bool {
				score := func(hero *Hero) int {
					if hero.IsLeader {
						return 1 << 20 // лидера выносят первым
					}
					s := hero.Mutations.Total()*200 + hero.Level*50
					if hero.Class == ClassTank || hero.Class == ClassPaladin {
						s += 400
					} else if hero.Class == ClassCleric || hero.Class == ClassBard {
						s += 300
					}
					return s
				}
				return score(downedHeroes[i]) > score(downedHeroes[j])
			})

			for _, h := range downedHeroes {
				// Если не хватило свободных рук — брошен в Бездне
				if carryCapacity <= 0 && !h.IsLeader {
					h.IsDead = true
					h.IsDowned = false
					h.LostInAbyss = true
					h.CauseOfDeath = T(m.Lang, "combat.log.left_in_abyss")
					m.recordFallenHero(h)
					continue
				}
				carryCapacity--

				// Бросок 80% / 20%: донесли ли раненого до ворот
				if !h.IsLeader && rng.Intn(100) >= 80 {
					// 20% неудача при подъёме на поверхность
					h.IsDead = true
					h.IsDowned = false
					h.LostInAbyss = true
					h.CauseOfDeath = T(m.Lang, "combat.log.evac_failed")
					m.recordFallenHero(h)
					m.addLog(dangerStyle.Render(T(m.Lang, "town.log.evac_failed", h.DisplayName(m.Lang))))
				}
			}
		}

		stabilizedCount := 0
		totalChurchSpent := 0

		// 2. Храм исцеляет и стабилизирует успешно доставленных бойцов (IsDowned)
		for _, h := range m.Party {
			if h.IsDowned && !h.IsDead && (h.IsLeader || m.Gold >= healBaseCost) {
				if h.IsLeader {
					// Лидера ставят на ноги за счёт ордена.
					m.addLog(fountStyle.Render(T(m.Lang, "leader.church_free", churchName, h.DisplayName(m.Lang))))
				} else {
					m.Gold -= healBaseCost
					totalChurchSpent += healBaseCost
				}
				h.IsDowned = false
				h.HP = h.MaxHP / 2
				h.MP = h.MaxMP / 2
				h.Stress = 70
				h.CauseOfDeath = ""
				stabilizedCount++
				m.Stats.Resurrections++

				found := false
				for idx := len(m.Stats.FallenHeroes) - 1; idx >= 0; idx-- {
					if m.Stats.FallenHeroes[idx].FullName == h.FullName(m.Lang) && !m.Stats.FallenHeroes[idx].Revived {
						m.Stats.FallenHeroes[idx].Revived = true
						found = true
						break
					}
				}
				if !found {
					m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
						FullName: h.FullName(m.Lang),
						Class:    h.Class,
						Cause:    T(m.Lang, "ui.stabilized"),
						Floor:    m.Floor,
						Revived:  true,
					})
				}
			} else if !h.IsDead && !h.IsDowned && (h.Stress > 20 || h.Affliction != AfflictionNone) {
				cleanseCost := healBaseCost / 3
				if m.Gold >= cleanseCost {
					m.Gold -= cleanseCost
					totalChurchSpent += cleanseCost
					h.Stress = max(0, h.Stress-60)
					h.Affliction = AfflictionNone
				}
			}
		}

		if totalChurchSpent > 0 {
			globalDebugReport.GoldSpentBreakdown["Храм (стабилизация/очищение)"] += totalChurchSpent
			m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, totalChurchSpent, stabilizedCount)))
			m.logTownAction("⛪", churchName, T(m.Lang, "town.log.church_hist", stabilizedCount, totalChurchSpent))
		}
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		tavernCost := max(40, (25*m.Floor)-(m.Legacy.TavernLevel*10))
		tavernName := T(m.Lang, m.TownEst.TavernKey)
		if m.Gold >= tavernCost {
			m.Gold -= tavernCost
			globalDebugReport.GoldSpentBreakdown["Таверна (ночлег)"] += tavernCost
			for _, h := range m.Party {
				if !h.IsDead && !h.IsDowned {
					h.HP = h.MaxHP
					h.MP = h.MaxMP
				}
			}
			m.addLog(healStyle.Render(T(m.Lang, "town.log.tavern_rest", tavernName, tavernCost)))
			m.logTownAction("🍻", tavernName, T(m.Lang, "town.log.tavern_hist", tavernCost))
		} else {
			barnCost := max(5, tavernCost/4)
			if m.Gold >= barnCost {
				m.Gold -= barnCost
				globalDebugReport.GoldSpentBreakdown["Таверна (сарай)"] += barnCost
			}
			for _, h := range m.Party {
				if !h.IsDead && !h.IsDowned {
					h.HP = max(1, int(float64(h.MaxHP)*0.40))
					h.MP = int(float64(h.MaxMP) * 0.40)
				}
			}
			m.addLog(dangerStyle.Render(T(m.Lang, "town.log.tavern_barn")))
			m.logTownAction("🏚️", tavernName, T(m.Lang, "town.log.tavern_barn_hist"))
		}
		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
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

		recruitCost := max(45, 60+(m.Floor*25)-(m.Legacy.ChurchLevel*8))
		for i, h := range m.Party {
			// Заменяем окончательно погибших или оставшихся без помощи; лидера Гильдия не заменяет
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

				newClass := availableClasses[rng.Intn(len(availableClasses))]

				if m.Gold >= recruitCost {
					m.Gold -= recruitCost
					globalDebugReport.GoldSpentBreakdown["Найм ветеранов"] += recruitCost
					m.Party[i] = createHero(newClass, m.Floor, m.Legacy.SmithyLevel)
					m.addLog(healStyle.Render(T(m.Lang, "town.log.guild_veteran",
						guildName, m.Party[i].DisplayName(m.Lang), m.Party[i].RaceName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost)))
					m.logTownAction("⚔️", guildName, T(m.Lang, "town.log.guild_vet_hist",
						m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost))
				} else {
					m.Party[i] = m.createMilitiaForGuild(newClass)
					newHero := m.Party[i]
					tierName := T(m.Lang, newHero.TitleKey)
					m.addLog(subtleStyle.Render(T(m.Lang, "town.log.guild_militia",
						guildName, newHero.DisplayName(m.Lang), tierName, newHero.RaceName(m.Lang), newHero.ShortClass(m.Lang))))
					m.logTownAction("🤝", guildName, T(m.Lang, "town.log.guild_mil_hist",
						newHero.DisplayName(m.Lang), tierName, newHero.ShortClass(m.Lang)))
				}
			}
		}
		m.TownPhase = TownPhaseSmithy

	case TownPhaseSmithy:
		budget := AllocateBudget(m.Gold)
		smithyBudget := budget.Forge
		upgradesCount := 0
		totalSpent := 0
		smithyName := T(m.Lang, m.TownEst.SmithyKey)

		getUpgradeCost := func(it *EquipItem) int {
			if it == nil {
				return 999999
			}
			matMult := max(1, it.Material.ValueMult)
			nextLvl := it.UpgradeLevel + 1
			cost := (nextLvl * nextLvl * 45 * matMult) - (m.Legacy.SmithyLevel * 18)
			return max(35*matMult, cost)
		}

		for smithyBudget > 0 {
			var bestHero *Hero
			var bestSlot EquipSlot
			var bestItem *EquipItem
			minCost := 999999

			for _, h := range m.Party {
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
						cost := getUpgradeCost(it)
						if cost < minCost && smithyBudget >= cost && m.Gold >= cost {
							minCost = cost
							bestHero = h
							bestSlot = slot
							bestItem = it
						}
					}
				}
			}

			if bestItem == nil {
				break
			}

			m.Gold -= minCost
			smithyBudget -= minCost
			totalSpent += minCost
			bestItem.UpgradeLevel++
			bestHero.SetItemInSlot(bestSlot, bestItem)
			upgradesCount++
			m.Stats.UpgradesForged++
		}

		if totalSpent > 0 {
			globalDebugReport.GoldSpentBreakdown["Кузница (заточки)"] += totalSpent
			m.addLog(goldStyle.Render(T(m.Lang, "town.log.smithy_done", smithyName, upgradesCount, totalSpent)))
			m.logTownAction("⚒️", smithyName, T(m.Lang, "town.log.smithy_hist", upgradesCount, totalSpent))
		}
		m.TownPhase = TownPhaseTannery

	case TownPhaseTannery:
		budget := AllocateBudget(m.Gold)
		tanneryBudget := budget.Tannery + budget.Bags
		totalSpent := 0
		tanneryName := T(m.Lang, m.TownEst.TanneryKey)

		// 1. Пошив сумки
		if m.BagLevel < len(bagUpgrades)-1 {
			nextBag := bagUpgrades[m.BagLevel+1]
			maxAllowedTier := m.Legacy.TanneryLevel + 1
			if nextBag.Level <= maxAllowedTier && tanneryBudget >= nextBag.Cost && m.Gold >= nextBag.Cost {
				m.Gold -= nextBag.Cost
				tanneryBudget -= nextBag.Cost
				totalSpent += nextBag.Cost
				globalDebugReport.GoldSpentBreakdown["Улучшение сумок"] += nextBag.Cost
				m.BagLevel++
				bagName := T(m.Lang, nextBag.NameKey)
				m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_bag", tanneryName, bagName, nextBag.Capacity, nextBag.Cost)))
				m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_bag_hist", bagName, nextBag.Capacity, nextBag.Cost))
			}
		}

		// 2. Выделка легкой и средней брони, а также легкого оружия
		getUpgradeCost := func(it *EquipItem) int {
			if it == nil {
				return 999999
			}
			matMult := max(1, it.Material.ValueMult)
			nextLvl := it.UpgradeLevel + 1
			cost := (nextLvl * nextLvl * 38 * matMult) - (m.Legacy.TanneryLevel * 14)
			return max(30*matMult, cost)
		}

		craftedItems := 0
		for tanneryBudget > 0 {
			var bestHero *Hero
			var bestSlot EquipSlot
			var bestItem *EquipItem
			minCost := 999999

			for _, h := range m.Party {
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
						cost := getUpgradeCost(it)
						if cost < minCost && tanneryBudget >= cost && m.Gold >= cost {
							minCost = cost
							bestHero = h
							bestSlot = slot
							bestItem = it
						}
					}
				}
			}

			if bestItem == nil {
				break
			}

			m.Gold -= minCost
			tanneryBudget -= minCost
			totalSpent += minCost
			bestItem.UpgradeLevel++
			bestHero.SetItemInSlot(bestSlot, bestItem)
			craftedItems++
			m.Stats.UpgradesForged++
		}

		if totalSpent > 0 {
			globalDebugReport.GoldSpentBreakdown["Кожевник (выделка и сумки)"] += totalSpent
			if craftedItems > 0 {
				m.addLog(goldStyle.Render(T(m.Lang, "town.log.tannery_done", tanneryName, craftedItems)))
				m.logTownAction("🎒", tanneryName, T(m.Lang, "town.log.tannery_hist", craftedItems, totalSpent))
			}
		}
		m.TownPhase = TownPhaseAlchemist

	case TownPhaseAlchemist:
		budget := AllocateBudget(m.Gold)
		alchBudget := budget.Alchemy
		spentAlch := 0

		for alchBudget > 0 {
			var candidates []*Hero
			for _, h := range m.Party {
				if !h.IsDead && !h.IsDowned {
					candidates = append(candidates, h)
				}
			}
			if len(candidates) == 0 {
				break
			}

			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].Mutations.Total() < candidates[j].Mutations.Total()
			})
			target := candidates[0]

			pref := GetClassMutationPreference(target, m.Floor, m.Stats.FallenHeroes)
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

			cost := CalculateMutationCost(baseCost, m.Floor, target.Mutations.Total())
			if alchBudget >= cost && m.Gold >= cost {
				m.Gold -= cost
				alchBudget -= cost
				spentAlch += cost

				verbApplied := TVerb(m.Lang, target.Gender, "принял", "приняла", "took")
				hName := target.DisplayName(m.Lang)

				switch pref {
				case MutChimera:
					target.Mutations.ChimeraCount++
					target.MaxHP += 8
					target.HP += 8
					target.MaxMP += 5
					target.MP += 5
					target.BaseAtk += 1
					target.BaseDef += 1
					m.addLog(accentStyle.Render(T(m.Lang, "town.log.mut_chimera", hName, verbApplied, T(m.Lang, "mut.chimera.name"), cost)))
				case MutFury:
					target.Mutations.FuryCount++
					target.BaseAtk += 3
					target.MaxHP += 2
					target.HP += 2
					m.addLog(fireStyle.Render(T(m.Lang, "town.log.mut_fury", hName, verbApplied, T(m.Lang, "mut.fury.name"), cost)))
				case MutTitan:
					target.Mutations.TitanCount++
					target.MaxHP += 20
					target.HP += 20
					m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_titan", hName, verbApplied, T(m.Lang, "mut.titan.name"), cost)))
				case MutAether:
					target.Mutations.AetherCount++
					target.MaxMP += 14
					target.MP += 14
					target.BaseAtk += 1
					m.addLog(fountStyle.Render(T(m.Lang, "town.log.mut_aether", hName, verbApplied, T(m.Lang, "mut.aether.name"), cost)))
				case MutBastion:
					target.Mutations.BastionCount++
					target.BaseDef += 2
					target.MaxHP += 6
					target.HP += 6
					m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_bastion", hName, verbApplied, T(m.Lang, "mut.bastion.name"), cost)))
				}
			} else {
				break
			}
		}

		maxPots := m.MaxPotionSlots()
		for _, h := range m.Party {
			if h.IsDead || h.IsDowned {
				continue
			}
			for h.HasFreePotionSlot(maxPots) {
				pType := PotionHP
				if h.Stress > 30 || h.Affliction != AfflictionNone {
					pType = PotionStress
				} else if h.Class == ClassMage || h.Class == ClassCleric || h.Class == ClassWarlock || h.Class == ClassBard {
					pType = PotionMP
				}

				cand := createPotion(pType, SizeSmall, m.Floor)
				if alchBudget >= cand.Cost && m.Gold >= cand.Cost {
					m.Gold -= cand.Cost
					alchBudget -= cand.Cost
					spentAlch += cand.Cost
					pot := cand
					h.Potions = append(h.Potions, &pot)
				} else {
					break
				}
			}
		}

		if spentAlch > 0 {
			globalDebugReport.GoldSpentBreakdown["Алхимия и зелья"] += spentAlch
			m.logTownAction("🧪", T(m.Lang, m.TownEst.AlchemistKey), T(m.Lang, "town.log.alch_hist", spentAlch))
		}
		m.TownPhase = TownPhaseDepart

	case TownPhaseDepart:
		// Сбрасываем кулдаун воскрешения Клирика к новому походу
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
