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
		biomeCycle := (floor - 1) % 11
		mbLvl := floor + 1
		mbHP := (110 + (mbLvl * 16)) * scaleMult

		var mbType MonsterType
		var mbNameKey string
		var mbGlyph rune
		var mbColor string
		var mbAtk, mbDef int
		var mbAffix MonsterAffix = AffixStone

		switch biomeCycle {
		case 0, 1:
			mbType, mbNameKey, mbGlyph, mbColor, mbAffix = MobMiniExecutioner, "mob.mini_executioner", 'E', "196", AffixPoison
			mbAtk, mbDef = (18+floor*2)*scaleMult, (6+floor/3)*scaleMult
		case 2, 7:
			mbType, mbNameKey, mbGlyph, mbColor, mbAffix = MobMiniReaver, "mob.mini_reaver", 'R', "208", AffixFire
			mbAtk, mbDef = (22+floor*2)*scaleMult, (4+floor/3)*scaleMult
		case 3, 6:
			mbType, mbNameKey, mbGlyph, mbColor, mbAffix = MobMiniColossus, "mob.mini_colossus", 'C', "51", AffixStone
			mbAtk, mbDef = (17+floor*2)*scaleMult, (9+floor/3)*scaleMult
		case 4, 5:
			mbType, mbNameKey, mbGlyph, mbColor, mbAffix = MobMiniBoneblight, "mob.mini_boneblight", 'W', "118", AffixPoison
			mbAtk, mbDef = (20+floor*2)*scaleMult, (7+floor/3)*scaleMult
		default: // 8, 9, 10
			mbType, mbNameKey, mbGlyph, mbColor, mbAffix = MobMiniStalker, "mob.mini_stalker", 'S', "93", AffixVampiric
			mbAtk, mbDef = (24+floor*2)*scaleMult, (5+floor/3)*scaleMult
		}

		pack.Members = append(pack.Members, &Monster{
			ID: 1, Type: mbType, NameKey: mbNameKey, Level: mbLvl, Affix: mbAffix,
			Glyph: mbGlyph, Color: mbColor, HP: mbHP, MaxHP: mbHP,
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

		var mType MonsterType
		var glyph rune
		var color string
		var baseAtk, baseDef, baseHP int

		biomeCycle := (floor - 1) % 11

		switch biomeCycle {
		case 0:
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobRat, 'r', "137", 6, 0, 15
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGoblin, 'g', "118", 8, 1, 19
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSkeleton, 's', "252", 10, 3, 25
			}
		case 1:
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSlime, 'c', "43", 11, 1, 30
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDrowned, 'u', "31", 13, 2, 38
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobLizard, 'l', "29", 14, 3, 34
			}
		case 2:
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobImp, 'i', "208", 15, 2, 40
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobOrc, 'o', "130", 17, 4, 52
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSalamander, 'm', "196", 18, 3, 46
			}
		case 3:
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGargoyle, 'G', "102", 19, 6, 58
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGolem, 'C', "141", 20, 7, 68
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobPhantom, 'p', "159", 22, 2, 50
			}
		case 4: // Затонувший Лес
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSproutSkeleton, 's', "106", 12, 3, 28
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDryad, 'd', "83", 16, 1, 38
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobBlightEnt, 'T', "94", 18, 6, 56
			}
		case 5: // Грибные Топи
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSporling, 'x', "142", 13, 2, 32
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobTentacle, 't', "65", 15, 3, 42
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobToxicBasil, 'b', "35", 19, 4, 48
			}
		case 6: // Забытые Архивы
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobTomeBook, 'b', "221", 16, 1, 35
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobScrollMimic, 'm', "178", 18, 4, 44
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobArchiveKeeper, 'A', "75", 21, 5, 54
			}
		case 7: // Обсидиановые Шахты
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobObsidianBeetle, 'o', "238", 17, 8, 45
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobMinerGhoul, 'z', "243", 20, 3, 48
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDeepTroll, 'O', "130", 23, 6, 65
			}
		case 8: // Осквернённый Санктуарий
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobBloodCultist, 'c', "161", 19, 2, 42
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobShadowInquisitor, 'i', "125", 22, 5, 52
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobFallenCrusader, 'F', "196", 24, 7, 68
			}
		case 9: // Астральный Разлом
			roll := rng.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobAstralWeaver, 'w', "69", 20, 3, 46
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobChronoPhantom, 'p', "183", 22, 2, 50
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobEssenceDevour, 'D', "135", 25, 4, 60
			}
		default: // 10: Трон Бездны
			roll := rng.Intn(2)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobVoidDemon, 'V', "161", 25, 5, 78
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDeathKnight, 'K', "89", 27, 8, 88
			}
		}

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
