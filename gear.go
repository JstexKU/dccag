package main

// ============================================================
// БОНУСЫ СНАРЯЖЕНИЯ
// ============================================================
//
// Здоровье и мана с предметов добавляются к максимумам героя, пока предмет надет.
// Суффиксы: «Ярости» усиливает удары раненого героя, «Медитации» восстанавливает ману каждый
// ход боя. Стихийный префикс на броне даёт сопротивление мутантам (на оружии — слабое
// зачарование, см. enchant.go).

const (
	furyHealthPct       = 50 // «Ярость» действует, пока здоровье не выше этой доли, %
	fireResistPerPiece  = 15 // сопротивление огненным мутантам от одной вещи «Пылающей» брони, %
	fireResistCap       = 60
	frostResistPerPiece = 25 // шанс устоять перед холодом от одной вещи «Леденящей» брони, %
)

func (h *Hero) gearItems() []*EquipItem {
	return []*EquipItem{h.Weapon, h.Head, h.Chest, h.Legs}
}

// refreshGearBonuses приводит максимумы здоровья и маны в соответствие с надетым снаряжением:
// меняется только разница с уже учтённым вкладом (GearHP/GearMP), поэтому вызывать можно
// сколько угодно раз. Здоровый и полный герой остаётся полным, раненый — нет.
func (h *Hero) refreshGearBonuses() {
	hp, mp := 0, 0
	for _, it := range h.gearItems() {
		if it != nil {
			hp += it.BonusHP
			mp += it.BonusMP
		}
	}

	if d := hp - h.GearHP; d != 0 {
		wasFull := h.HP >= h.MaxHP
		h.MaxHP = max(1, h.MaxHP+d)
		if wasFull {
			h.HP = h.MaxHP
		} else {
			h.HP = min(h.HP, h.MaxHP)
		}
		h.GearHP = hp
	}
	if d := mp - h.GearMP; d != 0 {
		wasFull := h.MP >= h.MaxMP
		h.MaxMP = max(0, h.MaxMP+d)
		if wasFull {
			h.MP = h.MaxMP
		} else {
			h.MP = min(h.MP, h.MaxMP)
		}
		h.GearMP = mp
	}
}

// suffixTotal — суммарный бонус суффикса данного вида по всему снаряжению героя.
func (h *Hero) suffixTotal(effect SuffixType) int {
	total := 0
	for _, it := range h.gearItems() {
		if it != nil && it.Suffix != nil && it.Suffix.Effect == effect {
			total += it.Suffix.Bonus
		}
	}
	return total
}

// furyBonus — прибавка к урону от суффикса «Ярости»: действует, пока герой ранен.
func (h *Hero) furyBonus() int {
	if h.MaxHP <= 0 || h.HP*100 > h.MaxHP*furyHealthPct {
		return 0
	}
	return h.suffixTotal(SuffFury)
}

// manaRegen — мана, возвращаемая суффиксом «Медитации» за ход боя.
func (h *Hero) manaRegen() int {
	return h.suffixTotal(SuffMana)
}

// elementPieces — сколько частей брони (шлем, доспех, поножи) несут префикс данной стихии.
func (h *Hero) elementPieces(e ElementType) int {
	n := 0
	for _, it := range []*EquipItem{h.Head, h.Chest, h.Legs} {
		if it != nil && it.Prefix != nil && it.Prefix.Element == e {
			n++
		}
	}
	return n
}

// resistAffixDamage режет урон огненного мутанта по сопротивлению «Пылающей» брони.
func (h *Hero) resistAffixDamage(mob *Monster, dmg int) int {
	if mob == nil || mob.Affix != AffixFire {
		return dmg
	}
	res := min(fireResistCap, fireResistPerPiece*h.elementPieces(ElemFire))
	return dmg * (100 - res) / 100
}
