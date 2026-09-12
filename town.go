package main

import (
	"fmt"
	"math/rand"
	"sort"
)

// --- Столичные заведения и вспомогательные функции ---

func generateTownEstablishments() TownEstablishments {
	return TownEstablishments{
		SmithyKey:    fmt.Sprintf("town.smithy.%d", rand.Intn(5)+1),
		TanneryKey:   fmt.Sprintf("town.tannery.%d", rand.Intn(5)+1),
		TavernKey:    fmt.Sprintf("town.tavern.%d", rand.Intn(5)+1),
		GuildKey:     fmt.Sprintf("town.guild.%d", rand.Intn(5)+1),
		AlchemistKey: fmt.Sprintf("town.alchemist.%d", rand.Intn(5)+1),
		ChurchKey:    fmt.Sprintf("town.church.%d", rand.Intn(5)+1),
	}
}

func AllocateBudget(gold int) TownBudget {
	treasury := int(float64(gold) * 0.10)
	rem := gold - treasury
	return TownBudget{
		Treasury: treasury,
		Bags:     int(float64(rem) * 0.05),
		Recovery: int(float64(rem) * 0.25),
		Recruit:  int(float64(rem) * 0.15),
		Forge:    int(float64(rem) * 0.15),
		Tannery:  int(float64(rem) * 0.15),
		Alchemy:  int(float64(rem) * 0.15),
	}
}

func CalculatePotionCost(basePrice, floor int) int {
	mult := 1.0 + (float64(floor) * 0.08)
	cost := int(float64(basePrice) * mult)
	maxCost := basePrice * 5
	if cost > maxCost {
		return maxCost
	}
	return cost
}

func createPotion(pType PotionType, size PotionSize, floor int) Potion {
	baseCost := 40
	power := 20
	switch size {
	case SizeSmall:
		power = 20
		baseCost = 35
	case SizeMedium:
		power = 45
		baseCost = 65
	case SizeLarge:
		power = 80
		baseCost = 110
	case SizeGrand:
		power = 140
		baseCost = 190
	}

	cost := CalculatePotionCost(baseCost, floor)
	symbol := "🟢"
	if pType == PotionMP {
		symbol = "🔵"
	} else if pType == PotionStress {
		symbol = "🟣"
	}

	return Potion{Type: pType, Size: size, Power: power, Cost: cost, Symbol: symbol}
}

func getPotionName(p Potion, lang Language) string {
	return T(lang, fmt.Sprintf("potion.%s.%s", p.Size, p.Type))
}

func GetClassMutationPreference(class HeroClass, m HeroMutations) MutationType {
	switch class {
	case ClassTank:
		if m.TitanCount <= m.BastionCount {
			return MutTitan
		}
		if m.BastionCount < 3 {
			return MutBastion
		}
		return MutChimera
	case ClassWarrior:
		if m.FuryCount <= m.TitanCount {
			return MutFury
		}
		return MutTitan
	case ClassRogue:
		if m.FuryCount <= m.ChimeraCount*2 {
			return MutFury
		}
		return MutChimera
	case ClassMage:
		if m.AetherCount <= m.FuryCount {
			return MutAether
		}
		return MutFury
	case ClassCleric:
		if m.AetherCount <= m.BastionCount {
			return MutAether
		}
		if m.BastionCount <= m.ChimeraCount {
			return MutBastion
		}
		return MutChimera
	default:
		return MutChimera
	}
}

func CalculateMutationCost(basePrice, floor, heroMutationsCount int) int {
	floorMult := 1.0 + (float64(floor) * 0.07)
	heroMult := 1.0 + (float64(heroMutationsCount) * 0.12)
	cost := int(float64(basePrice) * floorMult * heroMult)

	hardCap := basePrice * 6
	if cost > hardCap {
		cost = hardCap
	}
	return cost
}

func (m *Model) MaxPotionSlots() int {
	if m.Legacy.TanneryLevel >= 5 {
		return 3
	}
	if m.Legacy.TanneryLevel >= 3 {
		return 2
	}
	return 1
}

func (m *Model) logTownAction(icon, building, desc string) {
	entry := fmt.Sprintf("%s %-20s: %s", icon, building, desc)
	m.TownHistory = append(m.TownHistory, entry)
	if len(m.TownHistory) > 6 {
		m.TownHistory = m.TownHistory[len(m.TownHistory)-6:]
	}
}

// --- Пайплайн пребывания в городе ---

func (m *Model) stepTown() {
	switch m.TownPhase {
	case TownPhaseSellLoot:
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
			if !h.IsDead {
				h.Stress = max(0, h.Stress-40)
				if h.Stress < 60 {
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
		}
		m.TownPhase = TownPhaseChurch

	case TownPhaseChurch:
		hasDead := false
		hasStress := false
		for _, h := range m.Party {
			if h.HP == 1 || h.IsDead {
				hasDead = true
			}
			if h.Stress > 0 {
				hasStress = true
			}
		}

		if hasDead || hasStress {
			blessCost := max(80, (100*m.Floor)-(m.Legacy.ChurchLevel*30))
			churchName := T(m.Lang, m.TownEst.ChurchKey)
			if m.Gold >= blessCost {
				m.Gold -= blessCost
				globalDebugReport.GoldSpentBreakdown["Церковь (исцеление/воскрешение)"] += blessCost
				revived := 0
				for _, h := range m.Party {
					if h.HP == 1 || h.IsDead {
						h.IsDead = false
						h.HP = h.MaxHP / 2
						revived++
					}
					h.Stress = 0
					h.Affliction = AfflictionNone
				}
				m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, blessCost, revived)))
				m.logTownAction("⛪", churchName, T(m.Lang, "town.log.church_hist", revived, blessCost))
			}
		}
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		tavernCost := max(35, (20*m.Floor)-(m.Legacy.TavernLevel*8))
		tavernName := T(m.Lang, m.TownEst.TavernKey)
		if m.Gold >= tavernCost {
			m.Gold -= tavernCost
			globalDebugReport.GoldSpentBreakdown["Таверна (ночлег)"] += tavernCost
			for _, h := range m.Party {
				if !h.IsDead {
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
				if !h.IsDead {
					h.HP = max(1, int(float64(h.MaxHP)*0.45))
					h.MP = int(float64(h.MaxMP) * 0.45)
				}
			}
			m.addLog(dangerStyle.Render(T(m.Lang, "town.log.tavern_barn")))
			m.logTownAction("🏚️", tavernName, T(m.Lang, "town.log.tavern_barn_hist"))
		}
		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
		guildName := T(m.Lang, m.TownEst.GuildKey)
		if m.CurrentQuest.Completed {
			reward := m.CurrentQuest.RewardGold * 2
			m.Gold += reward
			m.Stats.TotalGoldEarned += reward
			m.Stats.QuestsCompleted++
			m.addLog(questStyle.Render(T(m.Lang, "town.log.guild_quest", guildName, reward)))
			m.logTownAction("📜", guildName, T(m.Lang, "town.log.guild_quest_hist", reward))
			m.CurrentQuest = generateAutoQuest(m.Floor)
		}

		recruitCost := max(35, 50+(m.Floor*22)-(m.Legacy.ChurchLevel*6))
		for i, h := range m.Party {
			if h.IsDead {
				classes := []HeroClass{ClassTank, ClassWarrior, ClassRogue, ClassMage, ClassCleric}
				newClass := classes[rand.Intn(len(classes))]

				if m.Gold >= recruitCost {
					m.Gold -= recruitCost
					globalDebugReport.GoldSpentBreakdown["Найм ветеранов"] += recruitCost
					m.Party[i] = createHero(newClass, m.Floor, m.Legacy.SmithyLevel)
					m.addLog(healStyle.Render(T(m.Lang, "town.log.guild_veteran",
						guildName, m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost)))
					m.logTownAction("⚔️", guildName, T(m.Lang, "town.log.guild_vet_hist",
						m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost))
				} else {
					m.Party[i] = createHero(newClass, 1, 0)
					m.Party[i].TitleKey = "title.militia"
					m.addLog(subtleStyle.Render(T(m.Lang, "town.log.guild_militia",
						guildName, m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang))))
					m.logTownAction("🤝", guildName, T(m.Lang, "town.log.guild_mil_hist",
						m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang)))
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
			cost := (nextLvl * nextLvl * 35 * matMult) - (m.Legacy.SmithyLevel * 15)
			return max(30*matMult, cost)
		}

		for smithyBudget > 0 {
			var bestHero *Hero
			var bestSlot EquipSlot
			var bestItem *EquipItem
			minCost := 999999

			for _, h := range m.Party {
				if h.IsDead {
					continue
				}
				for _, slot := range []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs} {
					it := h.GetItemInSlot(slot)
					canForge := it != nil && it.UpgradeLevel < 6 && (it.Slot == SlotWeapon || it.Category == ArmorHeavy)
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

		// 1. Пошив новой сумки
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

		// 2. Выделка доспехов Роги, Жреца и мантий Мага
		getUpgradeCost := func(it *EquipItem) int {
			if it == nil {
				return 999999
			}
			matMult := max(1, it.Material.ValueMult)
			nextLvl := it.UpgradeLevel + 1
			cost := (nextLvl * nextLvl * 30 * matMult) - (m.Legacy.TanneryLevel * 12)
			return max(25*matMult, cost)
		}

		craftedItems := 0
		for tanneryBudget > 0 {
			var bestHero *Hero
			var bestSlot EquipSlot
			var bestItem *EquipItem
			minCost := 999999

			for _, h := range m.Party {
				if h.IsDead {
					continue
				}
				for _, slot := range []EquipSlot{SlotHead, SlotChest, SlotLegs} {
					it := h.GetItemInSlot(slot)
					canTan := it != nil && (it.Category == ArmorMedium || it.Category == ArmorLight) && it.UpgradeLevel < 6
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
				if !h.IsDead {
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

			pref := GetClassMutationPreference(target.Class, target.Mutations)
			baseCost := 220
			switch pref {
			case MutChimera:
				baseCost = 260
			case MutBastion:
				baseCost = 240
			case MutFury:
				baseCost = 220
			case MutTitan:
				baseCost = 210
			case MutAether:
				baseCost = 200
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
			if h.IsDead {
				continue
			}
			for h.HasFreePotionSlot(maxPots) {
				pType := PotionHP
				if h.Stress > 25 || h.Affliction != AfflictionNone {
					pType = PotionStress
				} else if h.Class == ClassMage || h.Class == ClassCleric {
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
		m.TownHistory = []string{}
		m.InTown = false
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		m.addLog(accentStyle.Render(T(m.Lang, "town.log.depart")))
	}
}