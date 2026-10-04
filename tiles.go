package main

import (
	"fmt"
)

func generateRelic(level int) PartyRelic {
	relicKeys := []string{"greed_compass", "martyr_crown", "holy_grail"}
	chosen := relicKeys[rng.Intn(len(relicKeys))]

	nameKey := fmt.Sprintf("relic.%s.name.%d", chosen, level)
	descKey := fmt.Sprintf("relic.%s.desc", chosen)

	switch chosen {
	case "greed_compass":
		gMult := 1.15 + (float64(level) * 0.10)
		dmgMod := 1.20 - (float64(level) * 0.02)
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: gMult, EnemyDmgMod: dmgMod, MartyrFury: false,
			DescKey: descKey,
		}
	case "martyr_crown":
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: true,
			DescKey: descKey,
		}
	default:
		res := 10 + (level * 10)
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: false, StressRes: res,
			DescKey: descKey,
		}
	}
}

func (m *Model) handleAltar() {
	m.Stats.AltarsUsed++
	m.checkQuestProgress(QuestUseAltar, "", 1)
	target := m.getRandomLivingHero()

	if target != nil {
		bloodCost := max(15, target.HP*35/100)
		target.HP -= bloodCost
		target.BaseAtk += 3
		m.addStress(target, 25)

		hName := target.DisplayName(m.Lang)
		if target.HP <= 0 {
			target.HP = 0
			target.CauseOfDeath = T(m.Lang, "dungeon.death.altar")
			m.recordFallenHero(target)
			verb := TVerb(m.Lang, target.Gender, "пал", "пала", "fell")
			m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.altar_death", hName, verb)))
		} else {
			verb := TVerb(m.Lang, target.Gender, "пожертвовал", "пожертвовала", "sacrificed")
			m.addLog(altarStyle.Render(T(m.Lang, "dungeon.log.altar_success", hName, verb, bloodCost)))
		}
	}
}

func (m *Model) handleFountain() {
	m.Stats.FountainsUsed++
	for _, h := range m.Party {
		if !h.IsDead {
			h.HP = h.MaxHP
			h.MP = h.MaxMP
			h.Stress = max(0, h.Stress-60)
			if h.Stress == 0 {
				h.Affliction = AfflictionNone
			}
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
		}
	}
	m.addLog(fountStyle.Render(T(m.Lang, "dungeon.log.fountain")))
}

func (m *Model) handleTrappedChest() {
	m.Stats.ChestsOpened++
	m.checkQuestProgress(QuestOpenChests, "", 1)
	d20 := rng.Intn(20) + 1

	rogueBonus := 0
	for _, h := range m.Party {
		if (h.Class == ClassRogue || h.Class == ClassRanger) && !h.IsDead {
			rogueBonus = 3
			break
		}
	}

	rollSuccess := (d20+rogueBonus >= 12)

	if rollSuccess {
		m.Stats.TrapsDisarmed++
		gold := int(float64(rng.Intn(25)+15) * m.Relic.GoldMult)
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			lh.RevealSecret()
			m.checkAndAwardTitle(lh)
		}
		rareItem := generateItemForClass(targetClass, m.Floor+1)
		m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.trapped_chest_success", gold, rareItem.DisplayName(m.Lang))))
		m.equipOrBag(rareItem)
	} else {
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.trapped_chest_boom", d20)))
		for _, h := range m.Party {
			if !h.IsDead {
				trapDmg := rng.Intn(12) + 10
				h.HP -= trapDmg
				m.addStress(h, 25)
				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = T(m.Lang, "dungeon.death.trap")
					m.recordFallenHero(h)
				}
			}
		}
	}
}

func (m *Model) handleRelicTile() {
	m.checkQuestProgress(QuestFindRelic, "", 1)

	newLevel := 2
	switch {
	case m.Floor >= 11:
		newLevel = 3
	case m.Floor >= 6 && m.Relic != nil && m.Relic.Level >= 2:
		newLevel = 3
	}

	newRelic := generateRelic(newLevel)

	if m.Relic == nil || newRelic.Level > m.Relic.Level {
		m.Relic = &newRelic
		m.addLog(relicTileStyle.Render(T(m.Lang, "dungeon.log.relic_found", T(m.Lang, newRelic.NameKey))))
	} else {
		m.Gold += 80
		m.Stats.TotalGoldEarned += 80
		m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.relic_salvaged")))
	}
}
