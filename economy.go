package main

import (
	"fmt"
)

func generateTownEstablishments() TownEstablishments {
	return TownEstablishments{
		SmithyKey:    fmt.Sprintf("town.smithy.%d", rng.Intn(5)+1),
		TanneryKey:   fmt.Sprintf("town.tannery.%d", rng.Intn(5)+1),
		TavernKey:    fmt.Sprintf("town.tavern.%d", rng.Intn(5)+1),
		GuildKey:     fmt.Sprintf("town.guild.%d", rng.Intn(5)+1),
		AlchemistKey: fmt.Sprintf("town.alchemist.%d", rng.Intn(5)+1),
		ChurchKey:    fmt.Sprintf("town.church.%d", rng.Intn(5)+1),
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

	// 2. Индикаторы дефицита живучести
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
		// Приоритет 2:1 в пользу Fury (muts.FuryCount <= muts.ChimeraCount*2).
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

func (m *Model) buyMissingOrBetterGear() {
	slots := []EquipSlot{SlotWeapon, SlotChest, SlotHead, SlotLegs}

	for _, h := range m.Party {
		if h.IsDead {
			continue
		}

		for _, slot := range slots {
			currItem := h.GetItemInSlot(slot)

			// 1. Слот пуст — покупка необходимого снаряжения в приоритете
			if currItem == nil {
				cost := 35 + (m.Floor * 15)
				if m.Gold >= cost {
					m.Gold -= cost
					globalDebugReport.GoldSpentBreakdown["Покупка снаряжения"] += cost

					newItem := generateItemForClassSlot(h.Class, slot, m.Floor)
					h.SetItemInSlot(slot, &newItem)

					hName := h.DisplayName(m.Lang)
					slotName := T(m.Lang, "slot."+string(slot))
					itemName := newItem.DisplayName(m.Lang)

					m.addLog(goldStyle.Render(T(m.Lang, "town.log.bought_missing", hName, slotName, itemName, cost)))
					m.logTownAction("🛒", T(m.Lang, "town.market"), T(m.Lang, "town.log.bought_missing_hist", hName, itemName, cost))
				}
				continue
			}

			// 2. Слот заполнен, но вещь слабая — плановый апгрейд при достатке казны
			upgradeThreshold := 120 + (m.Floor * 25)
			if m.Gold > upgradeThreshold {
				shopOffer := generateItemForClassSlot(h.Class, slot, m.Floor+1)
				if shopOffer.TotalStat() > currItem.TotalStat()+3 && m.Gold >= shopOffer.Value {
					m.Gold -= shopOffer.Value
					globalDebugReport.GoldSpentBreakdown["Покупка снаряжения"] += shopOffer.Value

					if len(m.Bag) < m.currentBagCapacity() {
						m.Bag = append(m.Bag, *currItem)
					} else {
						m.Gold += currItem.Value
					}

					h.SetItemInSlot(slot, &shopOffer)
					hName := h.DisplayName(m.Lang)
					itemName := shopOffer.DisplayName(m.Lang)

					m.addLog(goldStyle.Render(T(m.Lang, "town.log.bought_upgrade", hName, itemName, shopOffer.Value)))
					m.logTownAction("🛒", T(m.Lang, "town.market"), T(m.Lang, "town.log.bought_upgrade_hist", hName, itemName, shopOffer.Value))
				}
			}
		}
	}
}
