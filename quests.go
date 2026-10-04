package main

func generateAutoQuest(curFloor int) AutoQuest {
	qType := QuestType(rng.Intn(7))
	switch qType {
	case QuestHuntMonster:
		var pool []MonsterType
		biomeCycle := (curFloor - 1) % 11
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
	default:
		return AutoQuest{
			Type: QuestUseAltar, TargetCount: 1,
			TitleKey: "quest.altar.title", DescKey: "quest.altar.desc",
			RewardGold: 80 + curFloor*15,
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
