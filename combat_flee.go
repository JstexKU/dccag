package main

import (
	"sort"
)

// shouldForceRetreatFromCombat решает, критично ли отступать прямо из боя.
// BagFull/QuestDone — не критично (сначала добиваем пак).
// LowHP/NoResources/TooFewAlive — критично, телепорт в город.
func (m *Model) shouldForceRetreatFromCombat(reason RetreatReason) bool {
	switch reason {
	case RetreatLowHP, RetreatNoResources, RetreatTooFewAlive:
		return true
	}
	return false
}

// forceRetreatToTown — экстренное отступление в город со штрафом 20% золота.
func (m *Model) forceRetreatToTown(reason RetreatReason) {
	globalDebugReport.FleeAttempts++ // считаем как попытку бегства

	// Штраф −20% золота за паническое бегство
	goldPenalty := int(float64(m.Gold) * 0.20)
	m.Gold -= goldPenalty

	// Запись в историю города
	reasonKey := "town.log.retreat_reason.low_hp"
	switch reason {
	case RetreatLowHP:
		reasonKey = "town.log.retreat_reason.low_hp"
	case RetreatNoResources:
		reasonKey = "town.log.retreat_reason.no_resources"
	case RetreatTooFewAlive:
		reasonKey = "town.log.retreat_reason.too_few_alive"
	}
	m.logTownAction("🏃", T(m.Lang, "town.market"), T(m.Lang, "town.log.retreat_penalty", goldPenalty, T(m.Lang, reasonKey)))

	m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.emergency_retreat", goldPenalty)))

	m.Combat = nil
	m.InTown = true
	m.TownPhase = TownPhaseSellLoot
	m.TownDialog = T(m.Lang, "combat.log.emergency_retreat_dialog")
}

func (m *Model) shouldAttemptFlee() bool {
	if m.Combat == nil || m.Combat.FleeCooldown > 0 {
		return false
	}

	// Если retreat критичен — пытаемся бежать даже при высоком HP
	if m.shouldForceRetreatFromCombat(m.evaluateRetreat()) {
		return true
	}

	curHP, maxHP := 0, 0
	livingCount := 0
	for _, h := range m.Party {
		if !h.IsDead {
			curHP += h.HP
			maxHP += h.MaxHP
			livingCount++
		}
	}
	if maxHP == 0 || livingCount == 0 {
		return false
	}
	hpPercent := float64(curHP) / float64(maxHP)
	fleeAt := float64(m.Tactics.FleeHPPct) / 100
	return hpPercent < fleeAt || (livingCount <= 2 && hpPercent < fleeAt+0.15)
}

func (m *Model) attemptFlee() {
	globalDebugReport.FleeAttempts++
	chance := 45

	for _, h := range m.Party {
		if !h.IsDead {
			if h.Class == ClassRogue || h.Class == ClassRanger {
				chance += 20
			}
			if h.Class == ClassTank || h.Class == ClassPaladin {
				chance += 10
			}
		}
	}

	chance -= m.Floor / 2
	if chance < 20 {
		chance = 20
	}

	roll := rng.Intn(100)
	if roll < chance {
		globalDebugReport.FleeSuccesses++
		m.addLog(healStyle.Render(T(m.Lang, "combat.log.flee_success")))

		var survivors []*Hero
		var fallen []*Hero
		for _, h := range m.Party {
			if !h.IsDead {
				survivors = append(survivors, h)
			} else {
				fallen = append(fallen, h)
			}
		}

		carryCapacity := len(survivors)
		if len(fallen) > 0 {
			// Эвристическая сортировка павших по ценности (мутации > уровень > роль)
			sort.Slice(fallen, func(i, j int) bool {
				score := func(hero *Hero) int {
					s := hero.Mutations.Total()*200 + hero.Level*50
					if hero.Class == ClassTank || hero.Class == ClassPaladin {
						s += 400
					} else if hero.Class == ClassCleric || hero.Class == ClassBard {
						s += 300
					}
					return s
				}
				return score(fallen[i]) > score(fallen[j])
			})

			rescuedCount := 0
			for i, hero := range fallen {
				if i < carryCapacity {
					hero.HP = 1 // Тело спасено, ждёт службы в Храме
					rescuedCount++
				} else {
					hero.CauseOfDeath = T(m.Lang, "combat.log.left_in_abyss")
					hero.LostInAbyss = true // Тело безвозвратно потеряно в Бездне
					m.recordFallenHero(hero)
				}
			}

			if rescuedCount > 0 {
				m.addLog(altarStyle.Render(T(m.Lang, "combat.log.evacuation", rescuedCount)))
			}
			if len(fallen) > rescuedCount {
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.left_behind", len(fallen)-rescuedCount)))
			}
		}

		m.Combat = nil
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = T(m.Lang, "combat.log.retreat_dialog")
		m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.retreat_town")))
	} else {
		m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.flee_fail")))

		for _, h := range m.Party {
			if !h.IsDead {
				chipDamage := int(float64(h.HP) * 0.20)
				if chipDamage < 2 {
					chipDamage = 2
				}
				h.HP -= chipDamage
				h.Feats.DamageTaken += chipDamage
				m.addStress(h, 15)

				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = T(m.Lang, "combat.log.death_flee")
					m.recordFallenHero(h)
				}
			}
		}
		m.Combat.FleeCooldown = 3
	}
}
