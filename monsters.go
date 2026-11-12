package main

func spawnMonsterPack(isBoss bool, floor int) *MonsterPack {
	pack := &MonsterPack{IsBoss: isBoss}
	scaleMult := 1 + (floor / 8)

	// 1. Главный босс декады
	if isBoss && floor%10 == 0 {
		dragonLvl := floor
		dragonHP := (300 + (floor * 20)) * scaleMult
		dragon := &Monster{
			ID: 1, Type: MobDragon, NameKey: "mob.boss_dragon", Level: dragonLvl, Affix: AffixFire,
			Glyph: 'D', Color: "196", HP: dragonHP, MaxHP: dragonHP,
			Atk: (30 + floor*2) * scaleMult, Defense: (10 + floor/2) * scaleMult, Speed: 12, Exp: 450 * scaleMult,
		}
		pack.Members = append(pack.Members, dragon)
		for i := 1; i <= 2; i++ {
			pack.Members = append(pack.Members, &Monster{
				ID: i + 1, Type: MobDeathKnight, NameKey: "mob.death_knight", Level: dragonLvl - 1, Affix: AffixVampiric,
				Glyph: 'K', Color: "89", HP: (85 + floor*6) * scaleMult, MaxHP: (85 + floor*6) * scaleMult,
				Atk: (22 + floor*2) * scaleMult, Defense: (7 + floor/3) * scaleMult, Speed: 10, Exp: 75 * scaleMult,
			})
		}
		return pack
	}

	// 2. Спавн Мини-босса (шанс 20% на этажах от 3-го)
	isMiniBoss := isBoss || (floor >= 3 && rng.Intn(100) < 20)
	if isMiniBoss {
		pack.IsMiniBoss = true
		mb := getBiomeDef(floor).MiniBoss
		mbLvl := floor + 1
		mbHP := (110 + (mbLvl * 16)) * scaleMult
		mbAtk := (mb.Atk + floor*2) * scaleMult
		mbDef := (mb.Def + floor/3) * scaleMult

		pack.Members = append(pack.Members, &Monster{
			ID: 1, Type: mb.Type, NameKey: mb.NameKey, Level: mbLvl, Affix: mb.Affix,
			Glyph: mb.Glyph, Color: mb.Color, HP: mbHP, MaxHP: mbHP,
			Atk: mbAtk, Defense: mbDef, Speed: 10, Exp: (120 + mbLvl*8) * scaleMult,
		})

		for i := 1; i <= 2; i++ {
			pack.Members = append(pack.Members, &Monster{
				ID: i + 1, Type: MobSkeleton, NameKey: "mob.skeleton", Level: floor, Affix: AffixNone,
				Glyph: 's', Color: "245", HP: (30 + floor*5) * scaleMult, MaxHP: (30 + floor*5) * scaleMult,
				Atk: (12 + floor*2) * scaleMult, Defense: 4 * scaleMult, Speed: 8, Exp: (25 + floor*3) * scaleMult,
			})
		}
		return pack
	}

	// 3. Обычные стаи (11 биомов)
	maxExtra := floor / 6
	if maxExtra > 4 {
		maxExtra = 4
	}
	packSize := rng.Intn(3) + 3 + maxExtra
	affixes := []MonsterAffix{AffixNone, AffixFire, AffixPoison, AffixFrost, AffixStone, AffixVampiric}

	for i := 1; i <= packSize; i++ {
		mobLvl := floor
		if rng.Intn(100) < 35 {
			mobLvl++
		}
		aff := AffixNone
		if floor >= 2 && rng.Intn(100) < (25+(floor*5)) {
			aff = affixes[rng.Intn(len(affixes))]
		}

		mobs := getBiomeDef(floor).Mobs
		mob := mobs[rng.Intn(len(mobs))]
		mType, glyph, color := mob.Type, mob.Glyph, mob.Color
		baseAtk, baseDef, baseHP := mob.Atk, mob.Def, mob.HP

		hp := (baseHP + (mobLvl * 6)) * scaleMult
		atk := (baseAtk + (mobLvl * 3)) * scaleMult
		def := (baseDef + (mobLvl / 2)) * scaleMult

		if aff == AffixFire {
			atk += 4 * scaleMult
		} else if aff == AffixStone {
			def += 4 * scaleMult
			hp += 15 * scaleMult
		}

		pack.Members = append(pack.Members, &Monster{
			ID: i, Type: mType, NameKey: "mob." + string(mType), Level: mobLvl, Affix: aff,
			Glyph: glyph, Color: color, HP: hp, MaxHP: hp, Atk: atk, Defense: def, Speed: 8 + mobLvl/2, Exp: (14 + mobLvl*4) * scaleMult,
		})
	}
	return pack
}
