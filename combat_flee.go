package main

import (
	"sort"
)

// shouldForceRetreatFromCombat решает, критично ли отступать прямо из боя.
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

	if m.shouldForceRetreatFromCombat(m.evaluateRetreat()) {
		return true
	}

	curHP, maxHP := 0, 0
	livingCount := 0
	for _, h := range m.Party {
		if !h.IsDead && !h.IsDowned {
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
		if !h.IsDead && !h.IsDowned {
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
		var downed []*Hero
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned {
				survivors = append(survivors, h)
			} else if h.IsDowned && !h.IsDead {
				downed = append(downed, h)
			}
		}

		carryCapacity := len(survivors)
		if len(downed) > 0 {
			// Эвристическая сортировка поверженных по ценности
			sort.Slice(downed, func(i, j int) bool {
				score := func(hero *Hero) int {
					s := hero.Mutations.Total()*200 + hero.Level*50
					if hero.Class == ClassTank || hero.Class == ClassPaladin {
						s += 400
					} else if hero.Class == ClassCleric || hero.Class == ClassBard {
						s += 300
					}
					return s
				}
				return score(downed[i]) > score(downed[j])
			})

			rescuedCount := 0
			for i, hero := range downed {
				if i < carryCapacity {
					// Проверка 80% / 20%: донесут ли тяжелораненого бойца сквозь хаос отступления
					if rng.Intn(100) < 80 {
						hero.HP = 0
						hero.IsDowned = true // Тело спасено, ждет помощи в Храме
						rescuedCount++
					} else {
						// 20% неудача при выносе
						hero.IsDead = true
						hero.LostInAbyss = true
						hero.CauseOfDeath = T(m.Lang, "combat.log.evac_failed")
						m.recordFallenHero(hero)
						m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.evac_drop", hero.DisplayName(m.Lang))))
					}
				} else {
					// Не хватило свободных рук
					hero.IsDead = true
					hero.LostInAbyss = true
					hero.CauseOfDeath = T(m.Lang, "combat.log.left_in_abyss")
					m.recordFallenHero(hero)
				}
			}

			if rescuedCount > 0 {
				m.addLog(altarStyle.Render(T(m.Lang, "combat.log.evacuation", rescuedCount)))
			}
			if len(downed) > rescuedCount {
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.left_behind", len(downed)-rescuedCount)))
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
			if !h.IsDead && !h.IsDowned {
				chipDamage := int(float64(h.HP) * 0.20)
				if chipDamage < 2 {
					chipDamage = 2
				}
				h.HP -= chipDamage
				h.Feats.DamageTaken += chipDamage
				m.addStress(h, 15)

				if h.HP <= 0 {
					h.HP = 0
					h.IsDowned = true
					h.IsGuarding = false
					h.IsBerserk = false
					h.IsStealthed = false
					h.IsCharged = false
					h.IsAura = false
					h.CauseOfDeath = T(m.Lang, "combat.log.death_flee")
					verb := TVerb(m.Lang, h.Gender, "рухнул без сознания", "рухнула без сознания", "fell unconscious")
					m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.hero_downed", h.DisplayName(m.Lang), verb, T(m.Lang, "combat.log.flee_pursuit"))))
				}
			}
		}
		m.Combat.FleeCooldown = 3
	}
}