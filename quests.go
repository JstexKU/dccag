package main

func generateAutoQuest(curFloor int) AutoQuest {
	var candidates []QuestType

	// 1. Базовые и боевые контракты, доступные всегда
	candidates = append(candidates,
		QuestHuntMonster,
		QuestOpenChests,
		QuestReachFloor,
		QuestCollectGold,
		QuestFlawlessFight,
		QuestCriticalStrikes,
		QuestBreakArmor,
		QuestPurgeBiome,
	)

	// 2. Исследовательские контракты со 2-го этажа
	if curFloor >= 2 {
		candidates = append(candidates,
			QuestEscapeTrap,
			QuestDisarmTraps,
			QuestFullMap,
			QuestUseFountain,
		)
	}

	// 3. Продвинутые контракты с 3-го этажа
	if curFloor >= 3 {
		candidates = append(candidates,
			QuestHuntMiniBoss,
			QuestAffixBounty,
			QuestUseAltar,
			QuestFindRelic,
			QuestSecrets,
		)
	}

	// 4. Биомы с бочками (Катакомбы, Пепел, Шахты) или этажи 3+
	biomeCycle := (curFloor - 1) % 11
	if biomeCycle == 0 || biomeCycle == 2 || biomeCycle == 7 || curFloor >= 3 {
		candidates = append(candidates, QuestPyromania)
	}

	// 5. Контракт на Босса — только если текущий или следующий спуск ведёт к боссу (10, 11, 20, 21...)
	nextFloor := curFloor + 1
	if curFloor > 0 && (curFloor%10 == 0 || nextFloor%10 == 0 || curFloor%11 == 0 || nextFloor%11 == 0) {
		candidates = append(candidates, QuestHuntBoss, QuestHuntBoss)
	}

	qType := candidates[rng.Intn(len(candidates))]

	switch qType {
	case QuestHuntMonster:
		var pool []MonsterType
		switch biomeCycle {
		case 0:
			pool = []MonsterType{MobRat, MobGoblin, MobSkeleton}
		case 1:
			pool = []MonsterType{MobSlime, MobDrowned, MobLizard}
		case 2:
			pool = []MonsterType{MobImp, MobOrc, MobSalamander}
		case 3:
			pool = []MonsterType{MobGargoyle, MobGolem, MobPhantom}
		case 4:
			pool = []MonsterType{MobSproutSkeleton, MobDryad, MobBlightEnt}
		case 5:
			pool = []MonsterType{MobSporling, MobTentacle, MobToxicBasil}
		case 6:
			pool = []MonsterType{MobTomeBook, MobScrollMimic, MobArchiveKeeper}
		case 7:
			pool = []MonsterType{MobObsidianBeetle, MobDeepTroll, MobMinerGhoul}
		case 8:
			pool = []MonsterType{MobFallenCrusader, MobBloodCultist, MobShadowInquisitor}
		case 9:
			pool = []MonsterType{MobAstralWeaver, MobChronoPhantom, MobEssenceDevour}
		default: // 10: Трон Бездны
			pool = []MonsterType{MobVoidDemon, MobDeathKnight}
		}
		target := pool[rng.Intn(len(pool))]
		count := rng.Intn(3) + 3
		return AutoQuest{
			Type: QuestHuntMonster, TargetMob: target, TargetCount: count,
			TitleKey: "quest.hunt.title", DescKey: "quest.hunt.desc",
			RewardGold: count * (20 + curFloor*5),
		}

	case QuestHuntMiniBoss:
		return AutoQuest{
			Type: QuestHuntMiniBoss, TargetCount: 1,
			TitleKey: "quest.miniboss.title", DescKey: "quest.miniboss.desc",
			RewardGold: 150 + curFloor*25,
		}

	case QuestOpenChests:
		count := rng.Intn(2) + 2
		return AutoQuest{
			Type: QuestOpenChests, TargetCount: count,
			TitleKey: "quest.chest.title", DescKey: "quest.chest.desc",
			RewardGold: count * (30 + curFloor*3),
		}

	case QuestReachFloor:
		tf := curFloor + 1
		return AutoQuest{
			Type: QuestReachFloor, TargetCount: tf,
			TitleKey: "quest.floor.title", DescKey: "quest.floor.desc",
			RewardGold: tf * 50,
		}

	case QuestFindRelic:
		return AutoQuest{
			Type: QuestFindRelic, TargetCount: 1,
			TitleKey: "quest.relic.title", DescKey: "quest.relic.desc",
			RewardGold: 100 + curFloor*15,
		}

	case QuestEscapeTrap:
		return AutoQuest{
			Type: QuestEscapeTrap, TargetCount: 1,
			TitleKey: "quest.escape.title", DescKey: "quest.escape.desc",
			RewardGold: 110 + curFloor*20,
		}

	case QuestUseAltar:
		return AutoQuest{
			Type: QuestUseAltar, TargetCount: 1,
			TitleKey: "quest.altar.title", DescKey: "quest.altar.desc",
			RewardGold: 80 + curFloor*15,
		}

	case QuestDisarmTraps:
		count := 2
		return AutoQuest{
			Type: QuestDisarmTraps, TargetCount: count,
			TitleKey: "quest.disarm.title", DescKey: "quest.disarm.desc",
			RewardGold: 120 + curFloor*20,
		}

	case QuestUseFountain:
		return AutoQuest{
			Type: QuestUseFountain, TargetCount: 1,
			TitleKey: "quest.fountain.title", DescKey: "quest.fountain.desc",
			RewardGold: 90 + curFloor*15,
		}

	case QuestFullMap:
		return AutoQuest{
			Type: QuestFullMap, TargetCount: 1,
			TitleKey: "quest.map.title", DescKey: "quest.map.desc",
			RewardGold: 140 + curFloor*25,
		}

	case QuestCollectGold:
		targetGold := 120 + curFloor*30
		return AutoQuest{
			Type: QuestCollectGold, TargetCount: targetGold,
			TitleKey: "quest.gold.title", DescKey: "quest.gold.desc",
			RewardGold: 100 + curFloor*20,
		}

	case QuestSecrets:
		count := 2
		return AutoQuest{
			Type: QuestSecrets, TargetCount: count,
			TitleKey: "quest.secrets.title", DescKey: "quest.secrets.desc",
			RewardGold: 130 + curFloor*20,
		}

	case QuestPurgeBiome:
		return AutoQuest{
			Type: QuestPurgeBiome, TargetCount: 1,
			TitleKey: "quest.purge.title", DescKey: "quest.purge.desc",
			RewardGold: 160 + curFloor*30,
		}

	case QuestHuntBoss:
		return AutoQuest{
			Type: QuestHuntBoss, TargetCount: 1,
			TitleKey: "quest.boss.title", DescKey: "quest.boss.desc",
			RewardGold: 300 + curFloor*50,
		}

	case QuestAffixBounty:
		count := rng.Intn(2) + 2
		return AutoQuest{
			Type: QuestAffixBounty, TargetCount: count,
			TitleKey: "quest.affix.title", DescKey: "quest.affix.desc",
			RewardGold: count * (45 + curFloor*10),
		}

	case QuestPyromania:
		count := rng.Intn(2) + 2
		return AutoQuest{
			Type: QuestPyromania, TargetCount: count,
			TitleKey: "quest.pyro.title", DescKey: "quest.pyro.desc",
			RewardGold: count * (40 + curFloor*8),
		}

	case QuestFlawlessFight:
		count := 3
		return AutoQuest{
			Type: QuestFlawlessFight, TargetCount: count,
			TitleKey: "quest.flawless.title", DescKey: "quest.flawless.desc",
			RewardGold: 130 + curFloor*25,
		}

	case QuestCriticalStrikes:
		count := 8 + rng.Intn(5)
		return AutoQuest{
			Type: QuestCriticalStrikes, TargetCount: count,
			TitleKey: "quest.crit.title", DescKey: "quest.crit.desc",
			RewardGold: 110 + curFloor*20,
		}

	default: // QuestBreakArmor
		count := 6 + rng.Intn(5)
		return AutoQuest{
			Type: QuestBreakArmor, TargetCount: count,
			TitleKey: "quest.guard.title", DescKey: "quest.guard.desc",
			RewardGold: 100 + curFloor*18,
		}
	}
}

func (m *Model) checkQuestProgress(action QuestType, targetMob MonsterType, val int) {
	if m.CurrentQuest.Completed {
		return
	}
	if m.CurrentQuest.Type == action {
		if action == QuestHuntMonster && m.CurrentQuest.TargetMob != targetMob {
			return
		}
		if action == QuestReachFloor {
			if val >= m.CurrentQuest.TargetCount {
				m.CurrentQuest.Current = m.CurrentQuest.TargetCount
				m.CurrentQuest.Completed = true
				m.addLog(questStyle.Render(T(m.Lang, "dungeon.log.quest_done", T(m.Lang, m.CurrentQuest.TitleKey))))
			}
			return
		}
		m.CurrentQuest.Current += val
		if m.CurrentQuest.Current >= m.CurrentQuest.TargetCount {
			m.CurrentQuest.Current = m.CurrentQuest.TargetCount
			m.CurrentQuest.Completed = true
			m.addLog(questStyle.Render(T(m.Lang, "dungeon.log.quest_done", T(m.Lang, m.CurrentQuest.TitleKey))))
		}
	}
}