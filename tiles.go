package main

import (
	"fmt"
)

func generateRelic(level int) PartyRelic {
	chosen := AllRelics[rng.Intn(len(AllRelics))]

	nameKey := fmt.Sprintf("relic.%s.name.%d", string(chosen), level)
	descKey := fmt.Sprintf("relic.%s.desc", string(chosen))

	relic := PartyRelic{
		Type:        chosen,
		NameKey:     nameKey,
		Level:       level,
		DescKey:     descKey,
		GoldMult:    1.0,
		EnemyDmgMod: 1.0,
	}

	switch chosen {
	case RelicGreedCompass:
		relic.GoldMult = 1.15 + (float64(level) * 0.10)
		relic.EnemyDmgMod = 1.20 - (float64(level) * 0.02)
	case RelicMartyrCrown:
		relic.MartyrFury = true
	case RelicHolyGrail:
		relic.StressRes = 10 + (level * 10)
	case RelicPioneerEye:
		relic.VisionBonus = 1 + (level / 2) // +1 радиус на ур.1-2, +2 на ур.3
	case RelicQuicksilverVial:
		relic.SpeedBonus = 1 + level // +2..+4 к скорости всего отряда
		relic.FirstStrike = true
	case RelicObsidianAnvil:
		relic.GearTempered = true // закалка T2+ экипировки
	case RelicWhisperingSkull:
		relic.PartyCritBonus = 1 + level // -2..-4 к порогу крита
		relic.CritStressRisk = max(5, 20-level*3)
	case RelicPhoenixCore:
		relic.PhoenixCharged = true // заряд возрождения при смертельном уроне
	case RelicHornOfPlenty:
		relic.PotionPotency = 1.15 + (float64(level) * 0.10) // +25%..+45% силы зелий и привала
	case RelicInquisitorSeal:
		relic.BaneDmgBonus = 3 + (level * 2) // +5..+9 чистый урон по нежити/демонам/элите
	}

	return relic
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
			m.hazardDeath(target)
			verb := TVerb(m.Lang, target.Gender, "fall")
			m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.altar_death", hName, verb)))
		} else {
			verb := TVerb(m.Lang, target.Gender, "sacrifice")
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
		gold := int(float64(rng.Intn(25)+15) * m.goldMult())
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
					m.hazardDeath(h)
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