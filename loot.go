package main

import (
	"fmt"
)

func generateItemForClassSlot(class HeroClass, slot EquipSlot, floor int) EquipItem {
	matIdx := floor / 4
	if matIdx > 3 {
		matIdx = 3
	}

	var mat MaterialTier
	var cat ArmorCategory

	switch class {
	case ClassTank, ClassWarrior, ClassPaladin:
		cat = ArmorHeavy
		mat = MetalMaterials[rng.Intn(matIdx+1)]
	case ClassMage, ClassWarlock:
		cat = ArmorLight
		if slot == SlotWeapon {
			mat = MageWeaponMaterials[rng.Intn(matIdx+1)]
		} else {
			mat = ClothMaterials[rng.Intn(matIdx+1)]
		}
	case ClassMonk:
		cat = ArmorLight
		mat = ClothMaterials[rng.Intn(matIdx+1)]
	case ClassRogue, ClassRanger:
		cat = ArmorMedium
		mat = LeatherMaterials[rng.Intn(matIdx+1)]
	case ClassCleric, ClassBard:
		cat = ArmorMedium
		if slot == SlotWeapon {
			mat = MetalMaterials[rng.Intn(matIdx+1)]
		} else {
			mat = LeatherMaterials[rng.Intn(matIdx+1)]
		}
	default:
		cat = ArmorMedium
		mat = LeatherMaterials[rng.Intn(matIdx+1)]
	}

	upg := 0
	if rng.Intn(100) > 75 {
		upg = rng.Intn(floor/3 + 1)
		if upg > 3 {
			upg = 3
		}
	}

	var pfx *PrefixDef
	if rng.Intn(100) > 60 {
		p := Prefixes[rng.Intn(len(Prefixes))]
		pfx = &p
	}

	var sfx *SuffixDef
	if rng.Intn(100) > 70 {
		s := Suffixes[rng.Intn(len(Suffixes))]
		if pfx != nil && pfx.Element == ElemFrost && s.Effect == SuffFury {
			s = Suffixes[2]
		}
		sfx = &s
	}

	tier := floor / 4
	if tier > 3 {
		tier = 3
	}
	if rng.Intn(100) < 25 && tier < 3 {
		tier++
	}

	baseKey := fmt.Sprintf("item.%s.%s.%d", string(class), string(slot), tier+1)
	baseStat := 2
	bonusMP := 0
	bonusHP := 0
	critBonus := 0
	blockBonus := 0
	stressRes := 0
	speedBonus := 0

	switch class {
	case ClassTank:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			baseStat = 3 + tier*2
			blockBonus = 2 + tier*2
		case SlotChest:
			baseStat = 4 + tier*3
			bonusHP = 10 + tier*10
		case SlotHead:
			baseStat = 2 + tier*2
			blockBonus = 1 + tier
		case SlotLegs:
			baseStat = 2 + tier*2
			bonusHP = 5 + tier*5
		}

	case ClassPaladin:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			blockBonus = 2 + tier
		case SlotChest:
			baseStat = 4 + tier*3
			bonusHP = 10 + tier*8
			bonusMP = 5 + tier*4
		case SlotHead:
			baseStat = 2 + tier*2
			stressRes = 5 + tier*5
		case SlotLegs:
			baseStat = 2 + tier*2
			bonusHP = 5 + tier*4
		}

	case ClassWarrior:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			baseStat = 5 + tier*3
			critBonus = 1 + tier
		case SlotChest:
			baseStat = 3 + tier*2
			bonusHP = 8 + tier*8
		case SlotHead:
			baseStat = 2 + tier*2
			critBonus = 1
		case SlotLegs:
			baseStat = 2 + tier*2
		}

	case ClassMonk:
		cat = ArmorLight
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			speedBonus = 2 + tier
		case SlotChest:
			baseStat = 1 + tier*2
			speedBonus = 1 + tier
		case SlotHead:
			baseStat = 1 + tier
			stressRes = 10 + tier*5
		case SlotLegs:
			baseStat = 1 + tier*2
			speedBonus = 1 + tier
		}

	case ClassRogue:
		cat = ArmorMedium
		if slot != SlotWeapon {
			speedBonus = 1 + tier
		}
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			critBonus = 2 + tier*2
		case SlotChest:
			baseStat = 2 + tier*2
			critBonus = 1 + tier
		case SlotHead:
			baseStat = 1 + tier*2
			critBonus = 1
		case SlotLegs:
			baseStat = 1 + tier*2
		}

	case ClassRanger:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			baseStat = 5 + tier*3
			critBonus = 2 + tier
		case SlotChest:
			baseStat = 2 + tier*2
			speedBonus = 1 + tier
		case SlotHead:
			baseStat = 1 + tier*2
			critBonus = 1 + tier
		case SlotLegs:
			baseStat = 1 + tier*2
			speedBonus = 1 + tier
		}

	case ClassMage:
		cat = ArmorLight
		if slot != SlotWeapon {
			speedBonus = tier
		}
		switch slot {
		case SlotWeapon:
			baseStat = 6 + tier*3
			bonusMP = 10 + tier*10
		case SlotChest:
			baseStat = 2 + tier*2
			bonusMP = 15 + tier*10
		case SlotHead:
			baseStat = 1 + tier*2
			bonusMP = 8 + tier*6
		case SlotLegs:
			baseStat = 1 + tier*2
			bonusMP = 6 + tier*4
		}

	case ClassWarlock:
		cat = ArmorLight
		switch slot {
		case SlotWeapon:
			baseStat = 5 + tier*3
			critBonus = 1 + tier
		case SlotChest:
			baseStat = 2 + tier*2
			bonusMP = 10 + tier*8
		case SlotHead:
			baseStat = 1 + tier*2
			bonusMP = 8 + tier*6
		case SlotLegs:
			baseStat = 1 + tier*2
			bonusHP = 6 + tier*4
		}

	case ClassCleric:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			bonusMP = 8 + tier*6
		case SlotChest:
			baseStat = 3 + tier*2
			stressRes = 10 + tier*5
		case SlotHead:
			baseStat = 2 + tier*2
			stressRes = 5 + tier*5
		case SlotLegs:
			baseStat = 2 + tier*2
			bonusHP = 6 + tier*6
		}

	case ClassBard:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			baseStat = 3 + tier*2
			bonusMP = 6 + tier*4
		case SlotChest:
			baseStat = 2 + tier*2
			stressRes = 12 + tier*6
		case SlotHead:
			baseStat = 1 + tier*2
			stressRes = 8 + tier*4
		case SlotLegs:
			baseStat = 1 + tier*2
			speedBonus = 1 + tier
		}
	}

	val := (baseStat + tier*3) * mat.ValueMult * 12

	return EquipItem{
		BaseNameKey: baseKey, Slot: slot, Category: cat, AllowedClass: class,
		Material: mat, UpgradeLevel: upg, BaseStat: baseStat,
		BonusMP: bonusMP, BonusHP: bonusHP, CritBonus: critBonus,
		BlockBonus: blockBonus, StressRes: stressRes, SpeedBonus: speedBonus, Value: val,
		Prefix: pfx, Suffix: sfx,
	}
}

func generateItemForClass(class HeroClass, floor int) EquipItem {
	slots := []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs}
	chosenSlot := slots[rng.Intn(len(slots))]
	return generateItemForClassSlot(class, chosenSlot, floor)
}
