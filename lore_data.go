package main

func init() {
	AllLoreVolumes = buildLoreDatabase()
}

func buildLoreDatabase() []LoreVolume {
	var list []LoreVolume

	// ============================================================
	// РАЗДЕЛ 1. ФОЛИАНТЫ МАСТЕРСТВА И РОДОВЫЕ СВИТКИ (30 + 7)
	// ============================================================

	// 1. ТАНК
	list = append(list,
		LoreVolume{
			ID: "mastery_tank_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.tank.1.title", AuthorKey: "lore.author.brand_1",
			ExcerptKey: "lore.mastery.tank.1.excerpt", ContentKey: "lore.mastery.tank.1.text",
			BiomeOrigin: BiomeCatacombs,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassTank, Value: 1, DescKey: "lore.bonus.tank_1"},
		},
		LoreVolume{
			ID: "mastery_tank_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.tank.2.title", AuthorKey: "lore.author.brand_1",
			ExcerptKey: "lore.mastery.tank.2.excerpt", ContentKey: "lore.mastery.tank.2.text",
			BiomeOrigin: BiomeMines,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassTank, Value: 2, DescKey: "lore.bonus.tank_2"},
		},
		LoreVolume{
			ID: "mastery_tank_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.tank.3.title", AuthorKey: "lore.author.brand_1",
			ExcerptKey: "lore.mastery.tank.3.excerpt", ContentKey: "lore.mastery.tank.3.text",
			BiomeOrigin: BiomeAbyss,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassTank, Value: 3, DescKey: "lore.bonus.tank_3"},
		},
	)

	// 2. ПАЛАДИН
	list = append(list,
		LoreVolume{
			ID: "mastery_paladin_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.paladin.1.title", AuthorKey: "lore.author.dawn_order",
			ExcerptKey: "lore.mastery.paladin.1.excerpt", ContentKey: "lore.mastery.paladin.1.text",
			BiomeOrigin: BiomeSanctuary,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassPaladin, Value: 5, DescKey: "lore.bonus.paladin_1"},
		},
		LoreVolume{
			ID: "mastery_paladin_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.paladin.2.title", AuthorKey: "lore.author.dawn_order",
			ExcerptKey: "lore.mastery.paladin.2.excerpt", ContentKey: "lore.mastery.paladin.2.text",
			BiomeOrigin: BiomeInferno,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassPaladin, Value: 15, DescKey: "lore.bonus.paladin_2"},
		},
		LoreVolume{
			ID: "mastery_paladin_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.paladin.3.title", AuthorKey: "lore.author.dawn_order",
			ExcerptKey: "lore.mastery.paladin.3.excerpt", ContentKey: "lore.mastery.paladin.3.text",
			BiomeOrigin: BiomeSanctuary,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassPaladin, Value: 2, DescKey: "lore.bonus.paladin_3"},
		},
	)

	// 3. ВОИН
	list = append(list,
		LoreVolume{
			ID: "mastery_warrior_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.warrior.1.title", AuthorKey: "lore.author.north_berserk",
			ExcerptKey: "lore.mastery.warrior.1.excerpt", ContentKey: "lore.mastery.warrior.1.text",
			BiomeOrigin: BiomeDeadwood,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarrior, Value: 1, DescKey: "lore.bonus.warrior_1"},
		},
		LoreVolume{
			ID: "mastery_warrior_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.warrior.2.title", AuthorKey: "lore.author.north_berserk",
			ExcerptKey: "lore.mastery.warrior.2.excerpt", ContentKey: "lore.mastery.warrior.2.text",
			BiomeOrigin: BiomeInferno,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarrior, Value: 2, DescKey: "lore.bonus.warrior_2"},
		},
		LoreVolume{
			ID: "mastery_warrior_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.warrior.3.title", AuthorKey: "lore.author.north_berserk",
			ExcerptKey: "lore.mastery.warrior.3.excerpt", ContentKey: "lore.mastery.warrior.3.text",
			BiomeOrigin: BiomeAbyss,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarrior, Value: 3, DescKey: "lore.bonus.warrior_3"},
		},
	)

	// 4. МОНАХ
	list = append(list,
		LoreVolume{
			ID: "mastery_monk_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.monk.1.title", AuthorKey: "lore.author.grotto_hermit",
			ExcerptKey: "lore.mastery.monk.1.excerpt", ContentKey: "lore.mastery.monk.1.text",
			BiomeOrigin: BiomeGrotto,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMonk, Value: 1, DescKey: "lore.bonus.monk_1"},
		},
		LoreVolume{
			ID: "mastery_monk_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.monk.2.title", AuthorKey: "lore.author.grotto_hermit",
			ExcerptKey: "lore.mastery.monk.2.excerpt", ContentKey: "lore.mastery.monk.2.text",
			BiomeOrigin: BiomeCrystal,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMonk, Value: 2, DescKey: "lore.bonus.monk_2"},
		},
		LoreVolume{
			ID: "mastery_monk_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.monk.3.title", AuthorKey: "lore.author.grotto_hermit",
			ExcerptKey: "lore.mastery.monk.3.excerpt", ContentKey: "lore.mastery.monk.3.text",
			BiomeOrigin: BiomeAstral,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMonk, Value: 3, DescKey: "lore.bonus.monk_3"},
		},
	)

	// 5. РАЗБОЙНИК
	list = append(list,
		LoreVolume{
			ID: "mastery_rogue_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.rogue.1.title", AuthorKey: "lore.author.shadow_broker",
			ExcerptKey: "lore.mastery.rogue.1.excerpt", ContentKey: "lore.mastery.rogue.1.text",
			BiomeOrigin: BiomeCatacombs,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRogue, Value: 1, DescKey: "lore.bonus.rogue_1"},
		},
		LoreVolume{
			ID: "mastery_rogue_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.rogue.2.title", AuthorKey: "lore.author.shadow_broker",
			ExcerptKey: "lore.mastery.rogue.2.excerpt", ContentKey: "lore.mastery.rogue.2.text",
			BiomeOrigin: BiomeFungal,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRogue, Value: 2, DescKey: "lore.bonus.rogue_2"},
		},
		LoreVolume{
			ID: "mastery_rogue_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.rogue.3.title", AuthorKey: "lore.author.shadow_broker",
			ExcerptKey: "lore.mastery.rogue.3.excerpt", ContentKey: "lore.mastery.rogue.3.text",
			BiomeOrigin: BiomeArchives,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRogue, Value: 3, DescKey: "lore.bonus.rogue_3"},
		},
	)

	// 6. СЛЕДОПЫТ
	list = append(list,
		LoreVolume{
			ID: "mastery_ranger_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.ranger.1.title", AuthorKey: "lore.author.deep_hunter",
			ExcerptKey: "lore.mastery.ranger.1.excerpt", ContentKey: "lore.mastery.ranger.1.text",
			BiomeOrigin: BiomeDeadwood,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRanger, Value: 1, DescKey: "lore.bonus.ranger_1"},
		},
		LoreVolume{
			ID: "mastery_ranger_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.ranger.2.title", AuthorKey: "lore.author.deep_hunter",
			ExcerptKey: "lore.mastery.ranger.2.excerpt", ContentKey: "lore.mastery.ranger.2.text",
			BiomeOrigin: BiomeFungal,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRanger, Value: 2, DescKey: "lore.bonus.ranger_2"},
		},
		LoreVolume{
			ID: "mastery_ranger_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.ranger.3.title", AuthorKey: "lore.author.deep_hunter",
			ExcerptKey: "lore.mastery.ranger.3.excerpt", ContentKey: "lore.mastery.ranger.3.text",
			BiomeOrigin: BiomeMines,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassRanger, Value: 3, DescKey: "lore.bonus.ranger_3"},
		},
	)

	// 7. КЛИРИК
	list = append(list,
		LoreVolume{
			ID: "mastery_cleric_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.cleric.1.title", AuthorKey: "lore.author.arch_bishop",
			ExcerptKey: "lore.mastery.cleric.1.excerpt", ContentKey: "lore.mastery.cleric.1.text",
			BiomeOrigin: BiomeSanctuary,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassCleric, Value: 1, DescKey: "lore.bonus.cleric_1"},
		},
		LoreVolume{
			ID: "mastery_cleric_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.cleric.2.title", AuthorKey: "lore.author.arch_bishop",
			ExcerptKey: "lore.mastery.cleric.2.excerpt", ContentKey: "lore.mastery.cleric.2.text",
			BiomeOrigin: BiomeCatacombs,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassCleric, Value: 2, DescKey: "lore.bonus.cleric_2"},
		},
		LoreVolume{
			ID: "mastery_cleric_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.cleric.3.title", AuthorKey: "lore.author.arch_bishop",
			ExcerptKey: "lore.mastery.cleric.3.excerpt", ContentKey: "lore.mastery.cleric.3.text",
			BiomeOrigin: BiomeAbyss,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassCleric, Value: 3, DescKey: "lore.bonus.cleric_3"},
		},
	)

	// 8. БАРД
	list = append(list,
		LoreVolume{
			ID: "mastery_bard_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.bard.1.title", AuthorKey: "lore.author.royal_minstrel",
			ExcerptKey: "lore.mastery.bard.1.excerpt", ContentKey: "lore.mastery.bard.1.text",
			BiomeOrigin: BiomeGrotto,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassBard, Value: 1, DescKey: "lore.bonus.bard_1"},
		},
		LoreVolume{
			ID: "mastery_bard_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.bard.2.title", AuthorKey: "lore.author.royal_minstrel",
			ExcerptKey: "lore.mastery.bard.2.excerpt", ContentKey: "lore.mastery.bard.2.text",
			BiomeOrigin: BiomeCrystal,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassBard, Value: 2, DescKey: "lore.bonus.bard_2"},
		},
		LoreVolume{
			ID: "mastery_bard_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.bard.3.title", AuthorKey: "lore.author.royal_minstrel",
			ExcerptKey: "lore.mastery.bard.3.excerpt", ContentKey: "lore.mastery.bard.3.text",
			BiomeOrigin: BiomeAstral,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassBard, Value: 3, DescKey: "lore.bonus.bard_3"},
		},
	)

	// 9. МАГ
	list = append(list,
		LoreVolume{
			ID: "mastery_mage_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.mage.1.title", AuthorKey: "lore.author.arch_mage",
			ExcerptKey: "lore.mastery.mage.1.excerpt", ContentKey: "lore.mastery.mage.1.text",
			BiomeOrigin: BiomeCrystal,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMage, Value: 1, DescKey: "lore.bonus.mage_1"},
		},
		LoreVolume{
			ID: "mastery_mage_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.mage.2.title", AuthorKey: "lore.author.arch_mage",
			ExcerptKey: "lore.mastery.mage.2.excerpt", ContentKey: "lore.mastery.mage.2.text",
			BiomeOrigin: BiomeArchives,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMage, Value: 2, DescKey: "lore.bonus.mage_2"},
		},
		LoreVolume{
			ID: "mastery_mage_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.mage.3.title", AuthorKey: "lore.author.arch_mage",
			ExcerptKey: "lore.mastery.mage.3.excerpt", ContentKey: "lore.mastery.mage.3.text",
			BiomeOrigin: BiomeAstral,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassMage, Value: 3, DescKey: "lore.bonus.mage_3"},
		},
	)

	// 10. ЧЕРНОКНИЖНИК
	list = append(list,
		LoreVolume{
			ID: "mastery_warlock_1", Category: LoreCatMastery, VolumeNum: 1,
			TitleKey: "lore.mastery.warlock.1.title", AuthorKey: "lore.author.void_apostle",
			ExcerptKey: "lore.mastery.warlock.1.excerpt", ContentKey: "lore.mastery.warlock.1.text",
			BiomeOrigin: BiomeInferno,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarlock, Value: 1, DescKey: "lore.bonus.warlock_1"},
		},
		LoreVolume{
			ID: "mastery_warlock_2", Category: LoreCatMastery, VolumeNum: 2,
			TitleKey: "lore.mastery.warlock.2.title", AuthorKey: "lore.author.void_apostle",
			ExcerptKey: "lore.mastery.warlock.2.excerpt", ContentKey: "lore.mastery.warlock.2.text",
			BiomeOrigin: BiomeSanctuary,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarlock, Value: 2, DescKey: "lore.bonus.warlock_2"},
		},
		LoreVolume{
			ID: "mastery_warlock_3", Category: LoreCatMastery, VolumeNum: 3,
			TitleKey: "lore.mastery.warlock.3.title", AuthorKey: "lore.author.void_apostle",
			ExcerptKey: "lore.mastery.warlock.3.excerpt", ContentKey: "lore.mastery.warlock.3.text",
			BiomeOrigin: BiomeAbyss,
			Bonus: LoreBonus{Type: LoreBonusClassPerk, TargetClass: ClassWarlock, Value: 3, DescKey: "lore.bonus.warlock_3"},
		},
	)

	// РОДОВЫЕ СВИТКИ (7 РАС)
	races := []struct {
		Race HeroRace
		ID   string
	}{
		{RaceHuman, "lineage_human"},
		{RaceElf, "lineage_elf"},
		{RaceBeastman, "lineage_beastman"},
		{RaceOlongr, "lineage_olongr"},
		{RaceDwarf, "lineage_dwarf"},
		{RaceDemonkin, "lineage_demonkin"},
		{RaceDragonkin, "lineage_dragonkin"},
	}
	for i, r := range races {
		list = append(list, LoreVolume{
			ID: r.ID, Category: LoreCatMastery, VolumeNum: i + 1,
			TitleKey: "lore." + r.ID + ".title", AuthorKey: "lore.author.ancients",
			ExcerptKey: "lore." + r.ID + ".excerpt", ContentKey: "lore." + r.ID + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusRacePerk, TargetRace: r.Race, Value: 1, DescKey: "lore.bonus." + r.ID},
		})
	}

	// ============================================================
	// РАЗДЕЛ 2. ХРОНИКИ БЫТИЯ (8 ТОМОВ)
	// ============================================================
	for v := 1; v <= 8; v++ {
		list = append(list, LoreVolume{
			ID: "genesis_vol_" + itoa(v), Category: LoreCatGenesis, VolumeNum: v,
			TitleKey: "lore.genesis." + itoa(v) + ".title", AuthorKey: "lore.author.chronicler",
			ExcerptKey: "lore.genesis." + itoa(v) + ".excerpt", ContentKey: "lore.genesis." + itoa(v) + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusPartyExp, Value: 60 * v, DescKey: "lore.bonus.genesis_exp"},
		})
	}

	// ============================================================
	// РАЗДЕЛ 3. ЭТНОГРАФИЯ И МИФЫ НАРОДОВ (7 РАС)
	// ============================================================
	ethnoRaces := []string{"human", "elf", "beastman", "olongr", "dwarf", "demonkin", "dragonkin"}
	for i, rName := range ethnoRaces {
		list = append(list, LoreVolume{
			ID: "ethno_" + rName, Category: LoreCatEthnography, VolumeNum: i + 1,
			TitleKey: "lore.ethno." + rName + ".title", AuthorKey: "lore.author.scholar_walter",
			ExcerptKey: "lore.ethno." + rName + ".excerpt", ContentKey: "lore.ethno." + rName + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusStressHeal, Value: 20, DescKey: "lore.bonus.ethno_stress"},
		})
	}

	// ============================================================
	// РАЗДЕЛ 4. БЕЛЛЕТРИСТИКА И ЛИРИКА (8 ТОМОВ)
	// ============================================================
	for v := 1; v <= 8; v++ {
		list = append(list, LoreVolume{
			ID: "belles_vol_" + itoa(v), Category: LoreCatBellesLettres, VolumeNum: v,
			TitleKey: "lore.belles." + itoa(v) + ".title", AuthorKey: "lore.author.tavern_poet",
			ExcerptKey: "lore.belles." + itoa(v) + ".excerpt", ContentKey: "lore.belles." + itoa(v) + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusStressHeal, Value: 30, DescKey: "lore.bonus.belles_stress"},
		})
	}

	// ============================================================
	// РАЗДЕЛ 5. ЭКСПЕДИЦИОННЫЕ ДНЕВНИКИ БИОМОВ (11 БИОМОВ)
	// ============================================================
	biomeList := []struct {
		Biome BiomeType
		Slug  string
	}{
		{BiomeCatacombs, "catacombs"},
		{BiomeGrotto, "grotto"},
		{BiomeInferno, "inferno"},
		{BiomeCrystal, "crystal"},
		{BiomeDeadwood, "deadwood"},
		{BiomeFungal, "fungal"},
		{BiomeArchives, "archives"},
		{BiomeMines, "mines"},
		{BiomeSanctuary, "sanctuary"},
		{BiomeAstral, "astral"},
		{BiomeAbyss, "abyss"},
	}
	for i, bInfo := range biomeList {
		list = append(list, LoreVolume{
			ID: "diary_" + bInfo.Slug, Category: LoreCatDiaries, VolumeNum: i + 1,
			TitleKey: "lore.diary." + bInfo.Slug + ".title", AuthorKey: "lore.author.lost_scout",
			ExcerptKey: "lore.diary." + bInfo.Slug + ".excerpt", ContentKey: "lore.diary." + bInfo.Slug + ".text",
			BiomeOrigin: bInfo.Biome,
			Bonus: LoreBonus{Type: LoreBonusPartyExp, Value: 100, DescKey: "lore.bonus.diary_exp"},
		})
	}

	// ============================================================
	// РАЗДЕЛ 6. БЕСТИАРИЙ И ТРАКТАТЫ ОБ АФФИКСАХ (5 АФФИКСОВ + 4 БОССА)
	// ============================================================
	affixes := []struct {
		Affix MonsterAffix
		Slug  string
	}{
		{AffixFire, "fire"},
		{AffixPoison, "poison"},
		{AffixFrost, "frost"},
		{AffixStone, "stone"},
		{AffixVampiric, "vampiric"},
	}
	for i, aff := range affixes {
		list = append(list, LoreVolume{
			ID: "bestiary_affix_" + aff.Slug, Category: LoreCatBestiary, VolumeNum: i + 1,
			TitleKey: "lore.bestiary.affix." + aff.Slug + ".title", AuthorKey: "lore.author.tactician",
			ExcerptKey: "lore.bestiary.affix." + aff.Slug + ".excerpt", ContentKey: "lore.bestiary.affix." + aff.Slug + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusBaneDmg, TargetAffix: aff.Affix, Value: 3, DescKey: "lore.bonus.affix_bane"},
		})
	}

	bosses := []struct {
		Mob  MonsterType
		Slug string
	}{
		{MobDragon, "dragon"},
		{MobDeathKnight, "death_knight"},
		{MobMiniExecutioner, "executioner"},
		{MobMiniColossus, "colossus"},
	}
	for i, b := range bosses {
		list = append(list, LoreVolume{
			ID: "bestiary_boss_" + b.Slug, Category: LoreCatBestiary, VolumeNum: len(affixes) + i + 1,
			TitleKey: "lore.bestiary.boss." + b.Slug + ".title", AuthorKey: "lore.author.tactician",
			ExcerptKey: "lore.bestiary.boss." + b.Slug + ".excerpt", ContentKey: "lore.bestiary.boss." + b.Slug + ".text",
			BiomeOrigin: "",
			Bonus: LoreBonus{Type: LoreBonusBaneDmg, TargetMob: b.Mob, Value: 4, DescKey: "lore.bonus.boss_bane"},
		})
	}

	return list
}

// itoa — вспомогательная функция без внешних зависимостей для сборки ключей
func itoa(i int) string {
	if i >= 0 && i <= 9 {
		return string(rune('0' + i))
	}
	return "0"
}

// FindLoreVolume возвращает том по его ID
func FindLoreVolume(id string) *LoreVolume {
	for i := range AllLoreVolumes {
		if AllLoreVolumes[i].ID == id {
			return &AllLoreVolumes[i]
		}
	}
	return nil
}

// GetVolumesByCategory возвращает отфильтрованный срез томов конкретного раздела
func GetVolumesByCategory(cat LoreCategory) []LoreVolume {
	var res []LoreVolume
	for _, v := range AllLoreVolumes {
		if v.Category == cat {
			res = append(res, v)
		}
	}
	return res
}