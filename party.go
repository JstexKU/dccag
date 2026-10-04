package main

func (m *Model) currentBagCapacity() int {
	return bagUpgrades[m.BagLevel].Capacity
}

func (m *Model) isPartyWiped() bool {
	for _, h := range m.Party {
		if !h.IsDead {
			return false
		}
	}
	return true
}

func (m *Model) getRandomLivingHero() *Hero {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return nil
	}
	return living[rng.Intn(len(living))]
}

func (m *Model) addStress(h *Hero, amt int) {
	if h.IsDead {
		return
	}
	if m.Relic != nil && m.Relic.StressRes > 0 {
		amt = amt * (100 - m.Relic.StressRes) / 100
	}

	itemRes := 0
	for _, it := range []*EquipItem{h.Head, h.Chest, h.Legs} {
		if it != nil {
			itemRes += it.StressRes
		}
	}
	if itemRes > 50 {
		itemRes = 50
	}
	amt = amt * (100 - itemRes) / 100

	h.Stress += amt

	if amt >= 15 {
		splash := amt / 2
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += splash
			}
		}
	}

	if h.Stress >= 100 && h.Affliction == AfflictionNone {
		virtueChance := 20
		if h.Race == RaceOlongr {
			virtueChance = 40
		}

		if rng.Intn(100) < virtueChance {
			h.Affliction = AfflictionVirtuous
			h.Stress = 0
			h.HP = h.MaxHP
			verb1 := TVerb(m.Lang, h.Gender, "превозмог", "превозмогла", "conquered")
			verb2 := TVerb(m.Lang, h.Gender, "обрел", "обрела", "gained")
			m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.virtue", h.FullName(m.Lang), verb1, verb2)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress = max(0, ally.Stress-20)
				}
			}
		} else {
			affs := []AfflictionType{AfflictionParanoid, AfflictionSelfish, AfflictionManiac}
			h.Affliction = affs[rng.Intn(len(affs))]
			verb := TVerb(m.Lang, h.Gender, "сломлен", "сломлена", "broken")
			affName := T(m.Lang, "affliction."+string(h.Affliction))
			m.addLog(stressStyle.Render(T(m.Lang, "dungeon.log.affliction", h.FullName(m.Lang), verb, affName)))
		}
	}

	if h.Stress >= 200 && h.Affliction != AfflictionVirtuous {
		h.Stress = 200
		if h.HP <= 1 {
			h.HP = 0
			h.CauseOfDeath = T(m.Lang, "dungeon.death.heart_attack", T(m.Lang, "affliction."+string(h.Affliction)))
			m.recordFallenHero(h)
			m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.heart_attack_death", h.DisplayName(m.Lang))))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress += 35
				}
			}
			return
		}

		h.HP = 1
		h.Stress = 160
		verb := TVerb(m.Lang, h.Gender, "схватился", "схватилась", "clutched")
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.heart_attack", h.FullName(m.Lang), verb)))
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += 20
			}
		}
	}
}

func (m *Model) checkAndDrinkPotions(h *Hero) {
	if h.IsDead || len(h.Potions) == 0 || m.InTown {
		return
	}
	if m.ManualMode && m.Combat != nil {
		return
	}

	remainingPotions := []*Potion{}
	for _, p := range h.Potions {
		if p == nil {
			continue
		}

		shouldDrink := false
		switch p.Type {
		case PotionHP:
			missingHP := h.MaxHP - h.HP
			if float64(h.HP)/float64(h.MaxHP) <= float64(m.Tactics.PotionHPPct)/100 || missingHP >= p.Power {
				shouldDrink = true
			}
		case PotionMP:
			missingMP := h.MaxMP - h.MP
			skillNeeded := h.SkillCost
			if h.MP < skillNeeded || (h.MaxMP > 0 && float64(h.MP)/float64(h.MaxMP) <= 0.30) || missingMP >= p.Power {
				shouldDrink = true
			}
		case PotionStress:
			if h.Stress >= 70 || h.Affliction != AfflictionNone {
				shouldDrink = true
			}
		}

		if shouldDrink {
			verb := TVerb(m.Lang, h.Gender, "выпил", "выпила", "drank")
			pName := getPotionName(*p, m.Lang)
			hName := h.DisplayName(m.Lang)
			switch p.Type {
			case PotionHP:
				h.HP = min(h.MaxHP, h.HP+p.Power)
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_hp", hName, verb, pName, p.Power)))
			case PotionMP:
				h.MP = min(h.MaxHP, h.MP+p.Power)
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_mp", hName, verb, pName, p.Power)))
			case PotionStress:
				h.Stress = max(0, h.Stress-p.Power)
				h.Affliction = AfflictionNone
				verbStress := TVerb(m.Lang, h.Gender, "принял", "приняла", "took")
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_stress", hName, verbStress, pName, p.Power)))
			}
		} else {
			remainingPotions = append(remainingPotions, p)
		}
	}
	h.Potions = remainingPotions
}

func createHero(class HeroClass, floor int, smithyLvl int) *Hero {
	targetLevel := max(1, floor/2)
	race := AllRaces[rng.Intn(len(AllRaces))]

	maxHP := 42
	baseDef := 2
	speed := 10
	skillNameKey := "skill.strike"
	skillCost := 10
	maxMP := 35

	switch class {
	case ClassTank:
		maxHP = 60
		baseDef = 4
		speed = 8
		skillNameKey = "skill.tank_stance"
		skillCost = 8
		maxMP = 30
	case ClassPaladin:
		maxHP = 54
		baseDef = 3
		speed = 9
		skillNameKey = "skill.paladin_holy"
		skillCost = 10
		maxMP = 35
	case ClassWarrior:
		maxHP = 48
		baseDef = 2
		speed = 10
		skillNameKey = "skill.warrior_rage"
		skillCost = 10
		maxMP = 25
	case ClassMonk:
		maxHP = 44
		baseDef = 1
		speed = 13
		skillNameKey = "skill.monk_flurry"
		skillCost = 8
		maxMP = 30
	case ClassRogue:
		maxHP = 35
		baseDef = 1
		speed = 15
		skillNameKey = "skill.rogue_stealth"
		skillCost = 12
		maxMP = 35
	case ClassRanger:
		maxHP = 38
		baseDef = 1
		speed = 13
		skillNameKey = "skill.ranger_shot"
		skillCost = 9
		maxMP = 30
	case ClassMage:
		maxHP = 28
		baseDef = 0
		speed = 11
		skillNameKey = "skill.mage_charge"
		skillCost = 15
		maxMP = 45
	case ClassWarlock:
		maxHP = 34
		baseDef = 1
		speed = 10
		skillNameKey = "skill.warlock_curse"
		skillCost = 11
		maxMP = 40
	case ClassCleric:
		maxHP = 34
		baseDef = 2
		speed = 9
		skillNameKey = "skill.cleric_aura"
		skillCost = 12
		maxMP = 40
	case ClassBard:
		maxHP = 36
		baseDef = 1
		speed = 12
		skillNameKey = "skill.bard_song"
		skillCost = 9
		maxMP = 40
	}

	raceMod := GetRaceModifiers(race)
	if raceMod.MaxHPBonusPercent != 0 {
		maxHP += int(float64(maxHP) * raceMod.MaxHPBonusPercent)
	}

	nameDef := getRandomHeroName()

	h := &Hero{
		NameKey:      nameDef.NameKey,
		Race:         race,
		Gender:       nameDef.Gender,
		Class:        class,
		Role:         GetClassRole(class),
		Level:        1,
		Exp:          0,
		MaxHP:        maxHP,
		HP:           maxHP,
		MaxMP:        maxMP,
		MP:           maxMP,
		BaseAtk:      6 + smithyLvl,
		BaseDef:      baseDef,
		Speed:        speed,
		SkillNameKey: skillNameKey,
		SkillCost:    skillCost,
		IsDead:       false,
		Potions:      []*Potion{},
	}

	for h.Level < targetLevel {
		h.GainExp(h.NextLevelExp())
	}

	itemFloor := max(1, floor)
	w := generateItemForClassSlot(class, SlotWeapon, itemFloor)
	w.UpgradeLevel = smithyLvl
	h.Weapon = &w

	head := generateItemForClassSlot(class, SlotHead, itemFloor)
	h.Head = &head
	ch := generateItemForClassSlot(class, SlotChest, itemFloor)
	h.Chest = &ch
	legs := generateItemForClassSlot(class, SlotLegs, itemFloor)
	h.Legs = &legs

	return h
}

func (m *Model) distributePartyExp(expAmt int) {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return
	}

	expPerHero := expAmt / len(living)
	if expPerHero < 1 {
		expPerHero = 1
	}

	for _, h := range living {
		actualExp := expPerHero
		raceMod := GetRaceModifiers(h.Race)
		if raceMod.ExpBonusPercent > 0 {
			actualExp += int(float64(actualExp) * raceMod.ExpBonusPercent)
		}

		if h.GainExp(actualExp) {
			verb := TVerb(m.Lang, h.Gender, "достиг", "достигла", "reached")
			m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.lvl_up", h.DisplayName(m.Lang), verb, h.Level)))
		}
	}
}

func (m *Model) equipOrBag(item EquipItem) {
	var bestHero *Hero
	maxDiff := -99999

	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		if item.AllowedClass != "" && item.AllowedClass != h.Class {
			continue
		}

		currItem := h.GetItemInSlot(item.Slot)
		currStat := 0
		if currItem != nil {
			currStat = currItem.TotalStat()
		}

		diff := item.TotalStat() - currStat
		if diff > maxDiff {
			maxDiff = diff
			bestHero = h
		}
	}

	if bestHero != nil && maxDiff > 0 {
		oldItem := bestHero.GetItemInSlot(item.Slot)
		if oldItem != nil {
			// При смене экипировки старый предмет идёт в мешок.
			// Если мешок полон — старый предмет теряется (не кладём сверх лимита).
			if len(m.Bag) < m.currentBagCapacity() {
				m.Bag = append(m.Bag, *oldItem)
			} else {
				m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full", oldItem.DisplayName(m.Lang))))
			}
		}
		newItem := item
		bestHero.SetItemInSlot(newItem.Slot, &newItem)
		verb := TVerb(m.Lang, bestHero.Gender, "сменил", "сменила", "equipped")
		slotName := T(m.Lang, "slot."+string(item.Slot))
		m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.equip_swap",
			bestHero.DisplayName(m.Lang), verb, slotName, newItem.DisplayName(m.Lang), newItem.TotalStat())))
	} else {
		// Страховка: не класть в переполненный мешок.
		if len(m.Bag) >= m.currentBagCapacity() {
			m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full", item.DisplayName(m.Lang))))
			return
		}
		m.Bag = append(m.Bag, item)
		m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_stored", item.DisplayName(m.Lang))))
	}
}

func (m *Model) recordFallenHero(h *Hero) {
	m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
		FullName: h.FullName(m.Lang),
		Class:    h.Class,
		Cause:    h.CauseOfDeath,
		Floor:    m.Floor,
	})

	h.IsDead = true
	h.IsGuarding = false
	h.IsBerserk = false
	h.IsStealthed = false
	h.IsCharged = false
	h.IsAura = false
	h.Weapon = nil
	h.Head = nil
	h.Chest = nil
	h.Legs = nil
	h.Potions = []*Potion{}

	if m.Relic != nil && m.Relic.MartyrFury {
		m.addLog(fireStyle.Render(T(m.Lang, "dungeon.log.martyr_crown")))
		for _, ally := range m.Party {
			if !ally.IsDead {
				ally.BaseAtk += 4
			}
		}
	}
}
