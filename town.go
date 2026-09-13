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

// GetClassMutationPreference динамически оценивает уязвимость героя и историю потерь
func GetClassMutationPreference(h *Hero, floor int, fallenHistory []FallenHeroRecord) MutationType {
	muts := h.Mutations

	// 1. Проверяем, погибал ли этот класс в последних записях Книги Памяти
	recentlyDied := false
	for i := len(fallenHistory) - 1; i >= 0 && i >= len(fallenHistory)-6; i-- {
		if fallenHistory[i].Class == h.Class {
			recentlyDied = true
			break
		}
	}

	// 2. Индикаторы дефицита живучести (v2.4.3):
	hpThreshold := 30 + (floor * 6)
	isCriticallyWounded := float64(h.HP)/float64(h.MaxHP) <= 0.45
	isFragile := h.MaxHP < hpThreshold

	if recentlyDied || isCriticallyWounded || isFragile {
		// Безоговорочный приоритет на спасение жизни — Кровь Титана (+20 MaxHP)
		return MutTitan
	}

	// 3. Базовая ролевая прогрессия под 10 классов
	switch h.Class {
	case ClassTank:
		if muts.BastionCount < muts.TitanCount {
			return MutBastion // +2 Def, +6 HP
		}
		return MutChimera // +8 HP, +5 MP, +1 Atk, +1 Def

	case ClassPaladin:
		if muts.BastionCount <= muts.TitanCount {
			return MutBastion
		}
		if muts.AetherCount < 2 {
			return MutAether
		}
		return MutTitan

	case ClassWarrior:
		if muts.FuryCount <= muts.TitanCount {
			return MutFury // +3 Atk, +2 HP
		}
		return MutChimera

	case ClassMonk:
		if muts.FuryCount <= muts.ChimeraCount {
			return MutFury
		}
		if muts.TitanCount < 2 {
			return MutTitan
		}
		return MutChimera

	case ClassRogue:
		if muts.FuryCount <= muts.ChimeraCount*2 {
			return MutFury
		}
		return MutChimera

	case ClassRanger:
		if muts.FuryCount <= muts.ChimeraCount {
			return MutFury
		}
		return MutChimera

	case ClassMage:
		if muts.AetherCount <= muts.FuryCount {
			return MutAether // +14 MP, +1 Atk
		}
		return MutFury

	case ClassWarlock:
		if muts.AetherCount <= muts.TitanCount {
			return MutAether
		}
		if muts.TitanCount < 3 {
			return MutTitan // Варлоку нужно HP для жертвенных навыков
		}
		return MutFury

	case ClassCleric:
		if muts.AetherCount <= muts.BastionCount {
			return MutAether
		}
		return MutBastion

	case ClassBard:
		if muts.AetherCount <= muts.ChimeraCount {
			return MutAether
		}
		if muts.ChimeraCount < 2 {
			return MutChimera
		}
		return MutBastion

	default:
		return MutChimera
	}
}

func CalculateMutationCost(basePrice, floor, heroMutationsCount int) int {
	floorMult := 1.0 + (float64(floor) * 0.08)
	heroMult := 1.0 + (float64(heroMutationsCount) * 0.15)
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
		churchName := T(m.Lang, m.TownEst.ChurchKey)
		reviveBaseCost := 120 + (m.Floor * 30) - (m.Legacy.ChurchLevel * 15)
		if reviveBaseCost < 60 {
			reviveBaseCost = 60
		}

		revivedCount := 0
		totalChurchSpent := 0

		for _, h := range m.Party {
			if h.IsDead && m.Gold >= reviveBaseCost {
				m.Gold -= reviveBaseCost
				totalChurchSpent += reviveBaseCost
				h.IsDead = false
				h.HP = h.MaxHP / 2
				h.MP = h.MaxMP / 2
				h.Stress = 90
				h.CauseOfDeath = ""
				revivedCount++
				m.Stats.Resurrections++
			} else if !h.IsDead && (h.Stress > 20 || h.Affliction != AfflictionNone) {
				cleanseCost := reviveBaseCost / 3
				if m.Gold >= cleanseCost {
					m.Gold -= cleanseCost
					totalChurchSpent += cleanseCost
					h.Stress = max(0, h.Stress-60)
					h.Affliction = AfflictionNone
				}
			}
		}

		if totalChurchSpent > 0 {
			globalDebugReport.GoldSpentBreakdown["Церковь (исцеление/воскрешение)"] += totalChurchSpent
			m.addLog(fountStyle.Render(T(m.Lang, "town.log.church_liturgy", churchName, totalChurchSpent, revivedCount)))
			m.logTownAction("⛪", churchName, T(m.Lang, "town.log.church_hist", revivedCount, totalChurchSpent))
		}
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		tavernCost := max(40, (25*m.Floor)-(m.Legacy.TavernLevel*10))
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
			if h.IsDead {
				// Автоматический найм любого из 10 классов
				newClass := AllClasses[rand.Intn(len(AllClasses))]

				if m.Gold >= recruitCost {
					m.Gold -= recruitCost
					globalDebugReport.GoldSpentBreakdown["Найм ветеранов"] += recruitCost
					m.Party[i] = createHero(newClass, m.Floor, m.Legacy.SmithyLevel)
					m.addLog(healStyle.Render(T(m.Lang, "town.log.guild_veteran",
						guildName, m.Party[i].DisplayName(m.Lang), m.Party[i].RaceName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost)))
					m.logTownAction("⚔️", guildName, T(m.Lang, "town.log.guild_vet_hist",
						m.Party[i].DisplayName(m.Lang), m.Party[i].ShortClass(m.Lang), m.Party[i].Level, recruitCost))
				} else {
					m.Party[i] = createHero(newClass, 1, 0)
					m.Party[i].TitleKey = "title.militia"
					m.addLog(subtleStyle.Render(T(m.Lang, "town.log.guild_militia",
						guildName, m.Party[i].DisplayName(m.Lang), m.Party[i].RaceName(m.Lang), m.Party[i].ShortClass(m.Lang))))
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
			cost := (nextLvl * nextLvl * 45 * matMult) - (m.Legacy.SmithyLevel * 18)
			return max(35*matMult, cost)
		}

		// Кузница точит металл и тяжелые латы: Танк, Воин, Паладин
		// + Металлическое оружие ближнего боя Клирика и Барда
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

		// 2. Выделка легкой и средней брони, а также дистанционного/магического оружия
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
				if h.IsDead {
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

			// Адаптивный расчет мутации с учетом выживаемости и дефицита HP (v2.4.3)
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
			if h.IsDead {
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
		m.TownHistory = []string{}
		m.InTown = false
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		m.addLog(accentStyle.Render(T(m.Lang, "town.log.depart")))
	}
}