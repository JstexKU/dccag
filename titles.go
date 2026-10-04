package main

func (m *Model) checkAndAwardTitle(h *Hero) {
	if !isMilitiaTitle(h.TitleKey) {
		return
	}

	newTitleKey := ""
	switch {
	case h.Feats.BossKills >= 3:
		newTitleKey = "title.slayer_of_beasts"
	case h.Feats.BossKills >= 1:
		newTitleKey = "title.dragonslayer"
	case h.Feats.NearDeathEscapes >= 6:
		newTitleKey = "title.luck_cursed"
	case h.Feats.NearDeathEscapes >= 3:
		newTitleKey = "title.immortal"
	case h.Feats.TreasureFound >= 5:
		newTitleKey = "title.treasure_seeker"
	case h.Feats.SecretsRevealed >= 4:
		newTitleKey = "title.secret_keeper"

	// 1. Танк
	case h.Class == ClassTank && h.Feats.Blocks >= 10:
		newTitleKey = "title.impenetrable"
	case h.Class == ClassTank && h.Feats.AggroPulled >= 8:
		newTitleKey = "title.storm_shield"
	case h.Class == ClassTank && h.Feats.DamageTaken >= 75:
		newTitleKey = "title.the_wall"

	// 2. Паладин
	case h.Class == ClassPaladin && h.Feats.HealsGiven >= 80:
		newTitleKey = "title.paladin_redeemer"
	case h.Class == ClassPaladin && h.Feats.Blocks >= 8:
		newTitleKey = "title.paladin_bastion"
	case h.Class == ClassPaladin && h.Feats.Kills >= 5:
		newTitleKey = "title.paladin_crusader"

	// 3. Воин
	case h.Class == ClassWarrior && h.Feats.Kills >= 12:
		newTitleKey = "title.blood_blade"
	case h.Class == ClassWarrior && h.Feats.Kills >= 5:
		newTitleKey = "title.executioner"
	case h.Class == ClassWarrior && h.Feats.DamageDealt >= 90:
		newTitleKey = "title.axe"
	case h.Class == ClassWarrior && h.Feats.CriticalStrikes >= 4:
		newTitleKey = "title.rank_cleaver"

	// 4. Монах
	case h.Class == ClassMonk && h.Feats.CCDuration >= 20:
		newTitleKey = "title.monk_calm"
	case h.Class == ClassMonk && h.Feats.CritsLanded >= 6:
		newTitleKey = "title.monk_fist"
	case h.Class == ClassMonk && h.Feats.DamageDealt >= 85:
		newTitleKey = "title.monk_wind"

	// 5. Разбойник
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 6:
		newTitleKey = "title.phantom_strike"
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 3:
		newTitleKey = "title.shadow"
	case h.Class == ClassRogue && h.Feats.Kills >= 4:
		newTitleKey = "title.blade"
	case h.Class == ClassRogue && h.Feats.Backstabs >= 5:
		newTitleKey = "title.knife_in_the_back"
	case h.Class == ClassRogue && h.Feats.LootStolen >= 3:
		newTitleKey = "title.deft_hand"

	// 6. Следопыт
	case h.Class == ClassRanger && h.Feats.CCDuration >= 25:
		newTitleKey = "title.ranger_trapper"
	case h.Class == ClassRanger && h.Feats.Kills >= 6:
		newTitleKey = "title.ranger_sniper"
	case h.Class == ClassRanger && h.Feats.CriticalStrikes >= 5:
		newTitleKey = "title.ranger_hawkeye"

	// 7. Маг
	case h.Class == ClassMage && h.Feats.DamageDealt >= 150:
		newTitleKey = "title.stormbringer"
	case h.Class == ClassMage && h.Feats.DamageDealt >= 90:
		newTitleKey = "title.cinder"
	case h.Class == ClassMage && h.Feats.CCDuration >= 40:
		newTitleKey = "title.chains_of_the_void"
	case h.Class == ClassMage && h.Feats.ManaBursts >= 3:
		newTitleKey = "title.flash"

	// 8. Чернокнижник
	case h.Class == ClassWarlock && h.Feats.DamageDealt >= 110:
		newTitleKey = "title.warlock_harvester"
	case h.Class == ClassWarlock && h.Feats.ManaBursts >= 4:
		newTitleKey = "title.warlock_void"
	case h.Class == ClassWarlock && h.Feats.Kills >= 5:
		newTitleKey = "title.warlock_curser"

	// 9. Клирик
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 120:
		newTitleKey = "title.grace"
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 70:
		newTitleKey = "title.holy"
	case h.Class == ClassCleric && h.Feats.Revives >= 3:
		newTitleKey = "title.resurrector"
	case h.Class == ClassCleric && h.Feats.DoTsRemoved >= 4:
		newTitleKey = "title.purifier"

	// 10. Бард
	case h.Class == ClassBard && h.Feats.DoTsRemoved >= 5:
		newTitleKey = "title.bard_virtuoso"
	case h.Class == ClassBard && h.Feats.CCDuration >= 20:
		newTitleKey = "title.bard_siren"
	case h.Class == ClassBard && h.Feats.DamageDealt >= 60:
		newTitleKey = "title.bard_rhapsodist"
	}

	if newTitleKey != "" {
		h.TitleKey = newTitleKey
		verb := TVerb(m.Lang, h.Gender, "заслужил", "заслужила", "earned")
		m.addLog(titleStyle.Render(T(m.Lang, "dungeon.log.title_awarded", h.DisplayName(m.Lang), verb, T(m.Lang, newTitleKey))))
	}
}
