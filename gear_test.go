package main

import "testing"

func clearGear(h *Hero) {
	for _, slot := range []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs} {
		h.SetItemInSlot(slot, nil)
	}
}

// Здоровье и мана с предметов добавляются к максимумам и убираются вместе с предметом.
func TestGearBonusesAffectMaxHPAndMP(t *testing.T) {
	seedRNG(72)
	h := buildHero(ClassTank, RaceHuman, "", GenderMale, 1, 0)

	sumHP, sumMP := 0, 0
	for _, it := range h.gearItems() {
		if it != nil {
			sumHP += it.BonusHP
			sumMP += it.BonusMP
		}
	}
	if h.GearHP != sumHP || h.GearMP != sumMP {
		t.Fatalf("учтено HP %d/MP %d, на предметах %d/%d", h.GearHP, h.GearMP, sumHP, sumMP)
	}
	if h.HP != h.MaxHP || h.MP != h.MaxMP {
		t.Fatalf("новый герой должен быть полным: HP %d/%d MP %d/%d", h.HP, h.MaxHP, h.MP, h.MaxMP)
	}
	baseHP, baseMP := h.MaxHP-sumHP, h.MaxMP-sumMP

	clearGear(h)
	if h.MaxHP != baseHP || h.MaxMP != baseMP || h.GearHP != 0 || h.GearMP != 0 {
		t.Fatalf("без снаряжения: HP %d (ждали %d), MP %d (ждали %d)", h.MaxHP, baseHP, h.MaxMP, baseMP)
	}
	if h.HP > h.MaxHP || h.MP > h.MaxMP {
		t.Fatal("текущие значения превысили максимум после снятия")
	}

	// Полный герой, надевший вещь, остаётся полным.
	h.HP, h.MP = h.MaxHP, h.MaxMP
	h.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, BonusHP: 30, BonusMP: 12})
	if h.MaxHP != baseHP+30 || h.MaxMP != baseMP+12 || h.HP != h.MaxHP || h.MP != h.MaxMP {
		t.Errorf("после надевания: HP %d/%d MP %d/%d", h.HP, h.MaxHP, h.MP, h.MaxMP)
	}

	// Раненый не лечится от смены вещи; замена пересчитывает только разницу.
	h.HP = 10
	h.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, BonusHP: 50})
	if h.MaxHP != baseHP+50 || h.HP != 10 || h.MaxMP != baseMP {
		t.Errorf("замена: HP %d/%d MaxMP %d (ждали %d)", h.HP, h.MaxHP, h.MaxMP, baseMP)
	}

	h.HP = h.MaxHP
	h.SetItemInSlot(SlotChest, nil)
	if h.MaxHP != baseHP || h.HP > h.MaxHP {
		t.Errorf("после снятия: HP %d/%d", h.HP, h.MaxHP)
	}

	// Повторное обновление ничего не меняет.
	before := h.MaxHP
	h.refreshGearBonuses()
	h.refreshGearBonuses()
	if h.MaxHP != before {
		t.Error("refreshGearBonuses не идемпотентен")
	}
}

func TestFuryAndManaSuffixes(t *testing.T) {
	seedRNG(73)
	h := buildHero(ClassWarrior, RaceHuman, "", GenderMale, 1, 0)
	clearGear(h)
	fury := &SuffixDef{Key: "suffix.fury", Effect: SuffFury, Bonus: 4}
	mana := &SuffixDef{Key: "suffix.mana", Effect: SuffMana, Bonus: 3}
	h.SetItemInSlot(SlotWeapon, &EquipItem{Slot: SlotWeapon, Suffix: fury})
	h.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, Suffix: fury})
	h.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, Suffix: mana})

	h.MaxHP = 100
	for hp, want := range map[int]int{100: 0, 51: 0, 50: 8, 1: 8} {
		h.HP = hp
		if got := h.furyBonus(); got != want {
			t.Errorf("ярость при HP %d: %d, ожидалось %d", hp, got, want)
		}
	}

	if got := h.manaRegen(); got != 3 {
		t.Errorf("мана за ход: %d, ожидалось 3", got)
	}
	h.SetItemInSlot(SlotLegs, &EquipItem{Slot: SlotLegs, Suffix: mana})
	if got := h.manaRegen(); got != 6 {
		t.Errorf("мана за ход с двумя вещами: %d, ожидалось 6", got)
	}
}

// Стихийные префиксы на броне дают сопротивление мутантам и скорость; на оружии — нет.
func TestArmorElementPrefixes(t *testing.T) {
	seedRNG(74)
	h := buildHero(ClassWarrior, RaceHuman, "", GenderMale, 1, 0)
	clearGear(h)
	prefix := func(e ElementType) *PrefixDef { return &PrefixDef{Key: "prefix.x", Element: e} }

	h.SetItemInSlot(SlotWeapon, &EquipItem{Slot: SlotWeapon, Prefix: prefix(ElemFire)})
	if h.elementPieces(ElemFire) != 0 {
		t.Fatal("оружие не считается бронёй")
	}

	fireMob, plain := &Monster{Affix: AffixFire}, &Monster{Affix: AffixPoison}
	if h.resistAffixDamage(fireMob, 100) != 100 {
		t.Error("без брони сопротивления нет")
	}
	h.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, Prefix: prefix(ElemFire)})
	h.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, Prefix: prefix(ElemFire)})
	if got := h.resistAffixDamage(fireMob, 100); got != 70 {
		t.Errorf("две вещи: урон %d, ожидалось 70", got)
	}
	if h.resistAffixDamage(plain, 100) != 100 {
		t.Error("сопротивление действует только против огненных")
	}
	h.SetItemInSlot(SlotLegs, &EquipItem{Slot: SlotLegs, Prefix: prefix(ElemFire)})
	if got := h.resistAffixDamage(fireMob, 100); got != 55 {
		t.Errorf("три вещи: урон %d, ожидалось 55", got)
	}

	base := h.TotalSpeed()
	h.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, Prefix: prefix(ElemLightning)})
	if h.TotalSpeed() != base+1 {
		t.Errorf("громовая броня: скорость %d, ожидалось %d", h.TotalSpeed(), base+1)
	}
	h.SetItemInSlot(SlotWeapon, &EquipItem{Slot: SlotWeapon, Prefix: prefix(ElemLightning)})
	if h.TotalSpeed() != base+1 {
		t.Error("громовое оружие не должно давать скорость")
	}
}

// Ядовитая броня сокращает яд, леденящая даёт шанс устоять перед холодом.
func TestArmorResistsMutantStatuses(t *testing.T) {
	m := affixModel(t)
	victim := m.Party[1]
	clearGear(victim)
	victim.HP = victim.MaxHP
	prefix := func(e ElementType) *PrefixDef { return &PrefixDef{Key: "prefix.x", Element: e} }

	victim.SetItemInSlot(SlotHead, &EquipItem{Slot: SlotHead, Prefix: prefix(ElemPoison)})
	victim.SetItemInSlot(SlotChest, &EquipItem{Slot: SlotChest, Prefix: prefix(ElemPoison)})
	m.applyAffixOnHit(&Monster{Affix: AffixPoison, NameKey: "mob.rat"}, victim, 10)
	if victim.PoisonTurns != 1 {
		t.Errorf("две ядовитые вещи: ходов яда %d, ожидалось 1", victim.PoisonTurns)
	}

	clearGear(victim)
	for _, slot := range []EquipSlot{SlotHead, SlotChest, SlotLegs} {
		victim.SetItemInSlot(slot, &EquipItem{Slot: slot, Prefix: prefix(ElemFrost)})
	}
	chilled := 0
	for i := 0; i < 1000; i++ {
		victim.ChilledTurns = 0
		m.applyAffixOnHit(&Monster{Affix: AffixFrost, NameKey: "mob.rat"}, victim, 10)
		if victim.ChilledTurns > 0 {
			chilled++
		}
	}
	// Три вещи — 75% устоять, значит холод ложится примерно в 25% случаев.
	if chilled < 190 || chilled > 310 {
		t.Errorf("холод лёг %d раз из 1000, ожидалось около 250", chilled)
	}
}
