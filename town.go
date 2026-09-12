package main

import (
	"fmt"
	"math/rand"
	"sort"
)

// --- Столичные заведения и вспомогательные функции ---

func generateTownEstablishments() TownEstablishments {
	smithies := []string{"Драконий Вздох", "Пылающий Горн", "Молот и Наковальня", "Стальная Искра", "Удар Титана"}
	tanneries := []string{"Вторая Кожа", "Прочный Стежок", "Дубленый Лев", "Лоскут и Заклепка", "Северный Олень"}
	taverns := []string{"Пьяный Дракон", "Приют Пройдохи", "Золотой Кубок", "Последний Приют", "Кабанья Голова"}
	guilds := []string{"Железный Контракт", "Орден Рассвета", "Союз Четырех Ветров", "Искатели Судеб", "Гвардия Удачи"}
	alchemists := []string{"Магия Эфира", "Зеленый Флакон", "Корень Мандрагоры", "Философский Камень", "Капля Света"}
	churches := []string{"Храм Вечного Рассвета", "Обитель Семи Светил", "Монастырь Безмолвия", "Часовня Упавшей Звезды", "Святыня Живой Воды"}

	return TownEstablishments{
		SmithyName:    smithies[rand.Intn(len(smithies))],
		TanneryName:   tanneries[rand.Intn(len(tanneries))],
		TavernName:    taverns[rand.Intn(len(taverns))],
		GuildName:     guilds[rand.Intn(len(guilds))],
		AlchemistName: alchemists[rand.Intn(len(alchemists))],
		ChurchName:    churches[rand.Intn(len(churches))],
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

func getPotionName(p Potion) string {
	return fmt.Sprintf("%s %s", p.Size, p.Type)
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
			m.addLog(goldStyle.Render(fmt.Sprintf("⚖️ [Рынок] Трофеи проданы на +%dG.", soldGold)))
			m.logTownAction("⚖️", "Рыночная площадь", fmt.Sprintf("Сбыт трофеев на +%dG. Дух укреплен (-40 стресса)", soldGold))
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
				Name  string
				Level int
			}
			buildings := []Building{
				{Name: m.TownEst.SmithyName, Level: m.Legacy.SmithyLevel},
				{Name: m.TownEst.TanneryName, Level: m.Legacy.TanneryLevel},
				{Name: m.TownEst.ChurchName, Level: m.Legacy.ChurchLevel},
				{Name: m.TownEst.TavernName, Level: m.Legacy.TavernLevel},
			}
			sort.Slice(buildings, func(i, j int) bool { return buildings[i].Level < buildings[j].Level })

			targetBld := &buildings[0]
			switch targetBld.Name {
			case m.TownEst.SmithyName:
				m.Legacy.SmithyLevel++
			case m.TownEst.TanneryName:
				m.Legacy.TanneryLevel++
			case m.TownEst.ChurchName:
				m.Legacy.ChurchLevel++
			case m.TownEst.TavernName:
				m.Legacy.TavernLevel++
			}
			m.addLog(titleStyle.Render(fmt.Sprintf("🏛️ [Магистрат] Отчислено %dG на развитие города («%s» Ур.%d)!",
				investAmt, targetBld.Name, targetBld.Level+1)))
			m.logTownAction("🏛️", "Магистрат Столицы", fmt.Sprintf("Внесено %dG («%s» улучшена до Ур.%d)", investAmt, targetBld.Name, targetBld.Level+1))
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
				m.addLog(fountStyle.Render(fmt.Sprintf("⛪ [%s] Литургия проведена (-%dG, поднято: %d)!", m.TownEst.ChurchName, blessCost, revived)))
				m.logTownAction("⛪", m.TownEst.ChurchName, fmt.Sprintf("Литургия исцеления: поднято %d бойцов, снят стресс (-%dG)", revived, blessCost))
			}
		}
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		tavernCost := max(35, (20*m.Floor)-(m.Legacy.TavernLevel*8))
		if m.Gold >= tavernCost {
			m.Gold -= tavernCost
			globalDebugReport.GoldSpentBreakdown["Таверна (ночлег)"] += tavernCost
			for _, h := range m.Party {
				if !h.IsDead {
					h.HP = h.MaxHP
					h.MP = h.MaxMP
				}
			}
			m.addLog(healStyle.Render(fmt.Sprintf("🍻 [%s] Полноценный отдых (-%dG). Отряд полон сил.", m.TownEst.TavernName, tavernCost)))
			m.logTownAction("🍻", m.TownEst.TavernName, fmt.Sprintf("Ночлег в уютных покоях (-%dG). Силы восстановлены", tavernCost))
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
			m.addLog(dangerStyle.Render("🏚️ [Сеновал] Казна истощена! Ночлег на сеновале (45% сил)."))
			m.logTownAction("🏚️", m.TownEst.TavernName, "Казна пуста! Ночлег на сеновале (45% сил)")
		}
		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
		if m.CurrentQuest.Completed {
			reward := m.CurrentQuest.RewardGold * 2
			m.Gold += reward
			m.Stats.TotalGoldEarned += reward
			m.Stats.QuestsCompleted++
			m.addLog(questStyle.Render(fmt.Sprintf("📜 [%s] Контракт закрыт: +%dG!", m.TownEst.GuildName, reward)))
			m.logTownAction("📜", m.TownEst.GuildName, fmt.Sprintf("Закрыт контракт: получена награда +%dG", reward))
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
					m.addLog(healStyle.Render(fmt.Sprintf("⚔️ [%s] Нанят ветеран %s (%s, Ур.%d) за %dG!",
						m.TownEst.GuildName, m.Party[i].Name, m.Party[i].ShortClass(), m.Party[i].Level, recruitCost)))
					m.logTownAction("⚔️", m.TownEst.GuildName, fmt.Sprintf("Принят контракт ветерана %s (%s, Ур.%d) за %dG",
						m.Party[i].Name, m.Party[i].ShortClass(), m.Party[i].Level, recruitCost))
				} else {
					m.Party[i] = createHero(newClass, 1, 0)
					m.Party[i].Title = "Ополченец"
					m.addLog(subtleStyle.Render(fmt.Sprintf("🤝 [%s] Ополченец %s (%s) встал в строй бесплатно.",
						m.TownEst.GuildName, m.Party[i].Name, m.Party[i].ShortClass())))
					m.logTownAction("🤝", m.TownEst.GuildName, fmt.Sprintf("Ополченец %s (%s) встал в строй без оплаты",
						m.Party[i].Name, m.Party[i].ShortClass()))
				}
			}
		}
		m.TownPhase = TownPhaseSmithy

	case TownPhaseSmithy:
		budget := AllocateBudget(m.Gold)
		smithyBudget := budget.Forge
		upgradesCount := 0
		totalSpent := 0

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
					// Кузница точит ВСЁ оружие отряда (металл и дерево магов) + тяжелые латы
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
			m.addLog(goldStyle.Render(fmt.Sprintf("⚒️ [%s] Заточено оружия и лат: %d шт. (-%dG)!", m.TownEst.SmithyName, upgradesCount, totalSpent)))
			m.logTownAction("⚒️", m.TownEst.SmithyName, fmt.Sprintf("Заточено предметов арсенала: %d шт. (-%dG)", upgradesCount, totalSpent))
		}
		m.TownPhase = TownPhaseTannery

	case TownPhaseTannery:
		budget := AllocateBudget(m.Gold)
		tanneryBudget := budget.Tannery + budget.Bags
		totalSpent := 0

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
				m.addLog(goldStyle.Render(fmt.Sprintf("🎒 [%s] Сшит %s (%d сл.) за %dG!", m.TownEst.TanneryName, nextBag.Name, nextBag.Capacity, nextBag.Cost)))
				m.logTownAction("🎒", m.TownEst.TanneryName, fmt.Sprintf("Сшит %s (%d слотов) за %dG", nextBag.Name, nextBag.Capacity, nextBag.Cost))
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
				// Кожевник обслуживает только доспехи (не оружие!)
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
				m.addLog(goldStyle.Render(fmt.Sprintf("🎒 [%s] Укреплено кожи и ткани: %d шт.!", m.TownEst.TanneryName, craftedItems)))
				m.logTownAction("🎒", m.TownEst.TanneryName, fmt.Sprintf("Выделка легких и средних доспехов: %d шт. (-%dG)", craftedItems, totalSpent))
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

				verbApplied := target.Verb("принял", "приняла")

				switch pref {
				case MutChimera:
					target.Mutations.ChimeraCount++
					target.MaxHP += 8
					target.HP += 8
					target.MaxMP += 5
					target.MP += 5
					target.BaseAtk += 1
					target.BaseDef += 1
					m.addLog(accentStyle.Render(fmt.Sprintf("⚗️ %s %s [Сыворотку Химеры] (+8 HP, +5 MP, +1 Atk, +1 Def) за %dG!", target.Name, verbApplied, cost)))
				case MutFury:
					target.Mutations.FuryCount++
					target.BaseAtk += 3
					target.MaxHP += 2
					target.HP += 2
					m.addLog(fireStyle.Render(fmt.Sprintf("⚗️ %s %s [Эссенцию Ярости] (+3 Atk, +2 HP) за %dG!", target.Name, verbApplied, cost)))
				case MutTitan:
					target.Mutations.TitanCount++
					target.MaxHP += 20
					target.HP += 20
					m.addLog(healStyle.Render(fmt.Sprintf("⚗️ %s %s [Кровь Титана] (+20 HP) за %dG!", target.Name, verbApplied, cost)))
				case MutAether:
					target.Mutations.AetherCount++
					target.MaxMP += 14
					target.MP += 14
					target.BaseAtk += 1
					m.addLog(fountStyle.Render(fmt.Sprintf("⚗️ %s %s [Флюид Эфира] (+14 MP, +1 Atk) за %dG!", target.Name, verbApplied, cost)))
				case MutBastion:
					target.Mutations.BastionCount++
					target.BaseDef += 2
					target.MaxHP += 6
					target.HP += 6
					m.addLog(healStyle.Render(fmt.Sprintf("⚗️ %s %s [Эликсир Бастиона] (+2 Def, +6 HP) за %dG!", target.Name, verbApplied, cost)))
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
			m.logTownAction("🧪", m.TownEst.AlchemistName, fmt.Sprintf("Сварены сыворотки и настойки в пояса на сумму %dG", spentAlch))
		}
		m.TownPhase = TownPhaseDepart

	case TownPhaseDepart:
		m.TownHistory = []string{}
		m.InTown = false
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		m.addLog(accentStyle.Render("🛡️ Отряд снаряжен и спускается на глубину!"))
	}
}
