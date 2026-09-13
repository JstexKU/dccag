package main

import (
	"fmt"
	"math/rand"
	"sort"
)

func (m *Model) SelectTarget() *Hero {
	var candidates []*Hero
	totalWeight := 0

	for _, h := range m.Party {
		if !h.IsDead {
			w := h.Role.AggroWeight
			if h.HP < h.MaxHP/3 {
				w += 25
			}
			candidates = append(candidates, h)
			totalWeight += w
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	r := rand.Intn(totalWeight)
	curr := 0
	for _, h := range candidates {
		w := h.Role.AggroWeight
		if h.HP < h.MaxHP/3 {
			w += 25
		}
		curr += w
		if r < curr {
			return h
		}
	}
	return candidates[0]
}

func (m *Model) ApplyDamage(target *Hero, rawDmg int) (actual *Hero, finalDmg int, guarded bool) {
	if target.Class != ClassTank && target.Class != ClassPaladin {
		for _, guard := range m.Party {
			if !guard.IsDead && guard.Role.CanGuard && guard != target && guard.HP > guard.MaxHP/4 {
				chance := 30
				mitigation := 0.75
				if guard.Class == ClassTank {
					chance = 60
					mitigation = 0.50
				} else if guard.Class == ClassPaladin {
					chance = 50
					mitigation = 0.60
				}

				if rand.Intn(100) < chance {
					mitigated := int(float64(rawDmg) * mitigation)
					guard.HP -= mitigated
					globalDebugReport.DamageGuarded += (rawDmg - mitigated)
					guard.AddBlock()
					return guard, mitigated, true
				}
			}
		}
	}

	target.HP -= rawDmg
	return target, rawDmg, false
}

func (m *Model) startCombat(pos Point, pack *MonsterPack) {
	combat := &ActiveCombat{
		Pos:          pos,
		Pack:         pack,
		HasBarrel:    rand.Intn(100) < 30,
		FleeCooldown: 0,
		Round:        1,
	}

	biome := getBiome(m.Floor)
	speedPenalty := 0
	if biome.Name == BiomeGrotto {
		speedPenalty = 2
	}

	for _, h := range m.Party {
		if !h.IsDead {
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
			initRoll := rand.Intn(20) + 1 + h.TotalSpeed() + m.Legacy.TanneryLevel - speedPenalty
			combat.TurnQueue = append(combat.TurnQueue, TurnOrderEntry{
				Type: CombatantHero, HeroRef: h, Initiative: initRoll,
			})
		}
	}

	for _, mob := range pack.Members {
		if !mob.IsDead {
			initRoll := rand.Intn(20) + 1 + mob.Speed
			combat.TurnQueue = append(combat.TurnQueue, TurnOrderEntry{
				Type: CombatantMonster, MonsterRef: mob, Initiative: initRoll,
			})
		}
	}

	sort.Slice(combat.TurnQueue, func(i, j int) bool {
		return combat.TurnQueue[i].Initiative > combat.TurnQueue[j].Initiative
	})

	m.Combat = combat
	m.addLog(accentStyle.Render(T(m.Lang, "combat.log.start", pack.LivingCount())))
}

// ============================================================
// RETREAT ИЗ БОЯ — экстренное бегство в город
// ============================================================

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

// ============================================================
// FLEE — обычный побег от пака
// ============================================================

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
	return hpPercent < 0.25 || (livingCount <= 2 && hpPercent < 0.40)
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

	roll := rand.Intn(100)
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

// ============================================================
// EXECUTE COMBAT TURN
// ============================================================

func (m *Model) executeCombatTurn() {
	if m.Combat == nil {
		return
	}

	if m.Combat.FleeCooldown > 0 {
		m.Combat.FleeCooldown--
	}

	// 1. Победа
	if m.Combat.Pack.LivingCount() == 0 {
		m.addLog(healStyle.Render(T(m.Lang, "combat.log.pack_defeated")))
		delete(m.Packs, m.Combat.Pos)
		m.PartyPos = m.Combat.Pos
		m.Combat = nil
		m.revealFog()
		return
	}

	// 2. Поражение
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.Combat = nil
		m.RestartCountdown = 10
		return
	}

	// 3. Экстренный retreat из боя (критические причины)
	retreatReason := m.evaluateRetreat()
	if m.shouldForceRetreatFromCombat(retreatReason) {
		m.forceRetreatToTown(retreatReason)
		return
	}

	// 4. Обычная попытка побега
	if m.shouldAttemptFlee() {
		m.attemptFlee()
		if m.Combat == nil {
			return
		}
	}

	// 5. Обычный ход
	if m.Combat.TurnIdx >= len(m.Combat.TurnQueue) {
		m.Combat.TurnIdx = 0
		m.Combat.Round++
	}
	current := m.Combat.TurnQueue[m.Combat.TurnIdx]
	m.Combat.TurnIdx++

	biome := getBiome(m.Floor)

	enrageMult := 1.0
	if m.Combat.Round > 15 {
		globalDebugReport.EnrageProcs++
		enrageMult += float64(m.Combat.Round-15) * 0.15
	}

	if current.Type == CombatantHero {
		h := current.HeroRef
		if h.IsDead {
			return
		}

		// Расовая пассивка: регенерация маны эльфа
		raceMod := GetRaceModifiers(h.Race)
		if raceMod.ManaRegen > 0 && h.MP < h.MaxMP {
			h.MP = min(h.MaxMP, h.MP+raceMod.ManaRegen)
		}

		m.checkAndDrinkPotions(h)
		hName := h.DisplayName(m.Lang)

		if h.Affliction == AfflictionParanoid && rand.Intn(100) < 35 {
			verb := TVerb(m.Lang, h.Gender, "забился", "забилась", "cowered")
			m.addLog(stressStyle.Render(T(m.Lang, "combat.log.paranoid", hName, verb)))
			return
		}

		skillCost := h.SkillCost
		if biome.Name == BiomeCrystal {
			skillCost += 3
		}

		targetMob := m.Combat.Pack.GetLowestHPFocus()

		// ==================== БОЛЬШИЕ И СРЕДНИЕ УМЕНИЯ (10 КЛАССОВ) ====================

		// 1. ТАНК
		if h.Class == ClassTank {
			if h.MP >= skillCost && !h.IsGuarding {
				h.MP -= skillCost
				h.IsGuarding = true
				h.AddBlock()
				h.PullAggro()
				m.checkAndAwardTitle(h)
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.tank_stance", hName)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.Atk >= 12 && rand.Intn(100) < 45 {
				h.MP -= 6
				bashDmg := h.TotalDef() + 4
				targetMob.HP -= bashDmg
				targetMob.Atk = max(2, targetMob.Atk-3)
				h.AddBlock()
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.tank_bash", hName, bashDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 2. ПАЛАДИН
		if h.Class == ClassPaladin {
			var woundedAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && float64(ally.HP)/float64(ally.MaxHP) <= 0.45 {
					woundedAlly = ally
					break
				}
			}

			if woundedAlly != nil && h.MP >= skillCost {
				h.MP -= skillCost
				healAmt := 14 + (m.Floor * 2) + (h.TotalDef() / 2)
				woundedAlly.HP = min(woundedAlly.MaxHP, woundedAlly.HP+healAmt)
				woundedAlly.Stress = max(0, woundedAlly.Stress-10)
				h.Feats.HealsGiven += healAmt
				verb := TVerb(m.Lang, h.Gender, "совершил", "совершила", "performed")
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.paladin_heal", hName, verb, woundedAlly.DisplayName(m.Lang), healAmt)))
				m.checkAndAwardTitle(h)
				return
			} else if h.MP >= 6 && targetMob != nil && rand.Intn(100) < 50 {
				h.MP -= 6
				smiteDmg := h.TotalAtk() + (h.TotalDef() / 3) + 3
				targetMob.HP -= smiteDmg
				h.IsGuarding = true
				h.AddBlock()
				m.addLog(goldStyle.Render(T(m.Lang, "combat.log.paladin_smite", hName, smiteDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 3. ВОИН
		if h.Class == ClassWarrior {
			if h.MP >= skillCost && !h.IsBerserk {
				h.MP -= skillCost
				h.IsBerserk = true
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.warrior_rage", hName)))
			} else if h.MP >= 7 && m.Combat.Pack.LivingCount() >= 2 && rand.Intn(100) < 55 {
				h.MP -= 7
				cleaveDmg := h.TotalAtk() + 2
				hitCount := 0
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && hitCount < 2 {
						mob.HP -= cleaveDmg
						h.Feats.DamageDealt += cleaveDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
						hitCount++
					}
				}
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.warrior_cleave", hName, hitCount, cleaveDmg)))
				return
			}
		}

		// 4. МОНАХ
		if h.Class == ClassMonk {
			if m.Combat.Pack.LivingCount() > 0 && h.MP >= skillCost {
				h.MP -= skillCost
				h.IsCharged = true
				flurryDmg := (h.TotalAtk() / 2) + 3
				hits := 0
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && hits < 3 {
						mob.HP -= flurryDmg
						h.Feats.DamageDealt += flurryDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
						hits++
					}
				}
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.monk_flurry", hName, hits, flurryDmg)))
				return
			} else if h.MP >= 6 && targetMob != nil {
				h.MP -= 6
				palmDmg := h.TotalAtk() + 3
				targetMob.HP -= palmDmg
				targetMob.Speed = max(1, targetMob.Speed-4)
				h.AddCCDuration()
				m.addLog(fountStyle.Render(T(m.Lang, "combat.log.monk_palm", hName, palmDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 5. РАЗБОЙНИК
		if h.Class == ClassRogue {
			if h.MP >= skillCost && !h.IsStealthed {
				h.MP -= skillCost
				h.IsStealthed = true
				h.AddBackstab()
				verb := TVerb(m.Lang, h.Gender, "растворился", "растворилась", "vanished")
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.rogue_stealth", hName, verb)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.HP > 20 && rand.Intn(100) < 50 {
				h.MP -= 6
				poisonDmg := h.TotalAtk() + 6
				targetMob.HP -= poisonDmg
				h.Feats.DamageDealt += poisonDmg
				m.addLog(stressStyle.Render(T(m.Lang, "combat.log.rogue_poison", hName, poisonDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 6. СЛЕДОПЫТ
		if h.Class == ClassRanger {
			if targetMob != nil && h.MP >= skillCost && targetMob.Atk >= 12 && rand.Intn(100) < 60 {
				h.MP -= skillCost
				trapDmg := 8 + (m.Floor * 2)
				targetMob.HP -= trapDmg
				targetMob.Atk = max(2, targetMob.Atk-5)
				h.AddCCDuration()
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.ranger_trap", hName, T(m.Lang, targetMob.NameKey), trapDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			} else if h.MP >= 6 && m.Combat.Pack.LivingCount() >= 2 {
				h.MP -= 6
				volleyDmg := h.TotalAtk() + 1
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && (mob.Type == MobImp || mob.Type == MobPhantom || mob.Type == MobVoidDemon) {
						mob.HP -= volleyDmg
						h.Feats.DamageDealt += volleyDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
					}
				}
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.ranger_volley", hName, volleyDmg)))
				return
			}
		}

		// 7. КЛИРИК
		if h.Class == ClassCleric {
			var criticalAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && float64(ally.HP)/float64(ally.MaxHP) <= 0.45 {
					criticalAlly = ally
					break
				}
			}

			if criticalAlly != nil && h.MP >= skillCost && h.Affliction != AfflictionSelfish {
				h.MP -= skillCost
				hAmt := rand.Intn(8) + 12 + (m.Floor * 2)
				if m.Combat.Round > 15 {
					hAmt /= 2
				}
				criticalAlly.HP = min(criticalAlly.MaxHP, criticalAlly.HP+hAmt)
				criticalAlly.Stress = max(0, criticalAlly.Stress-10)
				h.Feats.HealsGiven += hAmt
				verb := TVerb(m.Lang, h.Gender, "исцелил", "исцелила", "healed")
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.cleric_heal", hName, verb, criticalAlly.DisplayName(m.Lang), hAmt)))
				m.checkAndAwardTitle(h)
				return
			} else if !h.IsAura && h.MP >= (skillCost+6) {
				h.MP -= skillCost
				h.IsAura = true
				m.addLog(fountStyle.Render(T(m.Lang, "combat.log.cleric_aura", hName)))
			} else if h.MP >= 7 && targetMob != nil && rand.Intn(100) < 40 {
				h.MP -= 7
				smiteDmg := h.TotalAtk() + 4
				targetMob.HP -= smiteDmg
				if lowest := m.getRandomLivingHero(); lowest != nil && lowest.HP < lowest.MaxHP {
					lowest.HP = min(lowest.MaxHP, lowest.HP+4)
				}
				verb := TVerb(m.Lang, h.Gender, "обрушил", "обрушила", "unleashed")
				m.addLog(fountStyle.Render(T(m.Lang, "combat.log.cleric_smite", hName, verb, smiteDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 8. БАРД
		if h.Class == ClassBard {
			hasHighStress := false
			for _, ally := range m.Party {
				if !ally.IsDead && ally.Stress >= 50 {
					hasHighStress = true
					break
				}
			}

			if hasHighStress && h.MP >= skillCost {
				h.MP -= skillCost
				for _, ally := range m.Party {
					if !ally.IsDead {
						ally.Stress = max(0, ally.Stress-15)
						if ally.Stress < 60 {
							ally.Affliction = AfflictionNone
						}
					}
				}
				h.RemoveDot()
				m.addLog(potionStyle.Render(T(m.Lang, "combat.log.bard_ballad", hName)))
				return
			} else if h.MP >= 6 && m.Combat.Pack.LivingCount() >= 2 {
				h.MP -= 6
				dissonanceDmg := (h.TotalAtk() / 2) + 2
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= dissonanceDmg
						mob.Atk = max(1, mob.Atk-2)
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
					}
				}
				m.addLog(stressStyle.Render(T(m.Lang, "combat.log.bard_dissonance", hName, dissonanceDmg)))
				return
			}
		}

		// 9. МАГ
		if h.Class == ClassMage {
			if m.Combat.HasBarrel && h.MP >= skillCost {
				h.MP -= skillCost
				m.Combat.HasBarrel = false
				m.Stats.BarrelsBlown++
				h.AddManaBurst()
				barrelDmg := 22 + (m.Floor * 3)
				m.addLog(barrelStyle.Render(T(m.Lang, "combat.log.mage_barrel", hName, barrelDmg)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= barrelDmg
						h.Feats.DamageDealt += barrelDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult)
							m.Gold += g
							m.Stats.TotalGoldEarned += g
							m.Stats.MonsterKills[mob.Type]++
							m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)
						}
					}
				}
				m.checkAndAwardTitle(h)
				return
			} else if m.Combat.Pack.LivingCount() >= 2 && h.MP >= skillCost {
				h.MP -= skillCost
				aoeDmg := h.TotalAtk() + 5
				h.AddManaBurst()
				h.AddCCDuration()
				verb := TVerb(m.Lang, h.Gender, "накрыл", "накрыла", "blanketed")
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.mage_storm", hName, verb)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= aoeDmg
						h.Feats.DamageDealt += aoeDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult)
							m.Gold += g
							m.Stats.TotalGoldEarned += g
							m.Stats.MonsterKills[mob.Type]++
							m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)
						}
					}
				}
				m.checkAndAwardTitle(h)
				return
			} else if h.MP >= 6 && m.Combat.Pack.LivingCount() >= 2 {
				h.MP -= 6
				chainDmg := h.TotalAtk() + 3
				hits := 0
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && hits < 2 {
						mob.HP -= chainDmg
						h.Feats.DamageDealt += chainDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
						hits++
					}
				}
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.mage_lightning", hName, chainDmg)))
				return
			} else {
				if eliteMob := m.Combat.Pack.GetHighestHPFocus(); eliteMob != nil {
					targetMob = eliteMob
				}
			}
		}

		// 10. ЧЕРНОКНИЖНИК
		if h.Class == ClassWarlock {
			if h.HP > 15 && h.MP >= skillCost {
				needMana := false
				for _, ally := range m.Party {
					if !ally.IsDead && ally != h && float64(ally.MP)/float64(ally.MaxMP) <= 0.30 {
						needMana = true
						break
					}
				}

				if needMana {
					h.MP -= skillCost
					h.HP -= 6
					for _, ally := range m.Party {
						if !ally.IsDead && ally != h {
							ally.MP = min(ally.MaxMP, ally.MP+12)
						}
					}
					h.AddManaBurst()
					m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.warlock_sacrifice", hName)))
					return
				}
			}

			if h.MP >= 6 && targetMob != nil {
				h.MP -= 6
				drainDmg := h.TotalAtk() + 4
				targetMob.HP -= drainDmg
				healSelf := drainDmg / 2
				h.HP = min(h.MaxHP, h.HP+healSelf)
				h.Feats.DamageDealt += drainDmg
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.warlock_drain", hName, drainDmg, healSelf)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		if targetMob == nil {
			return
		}

		mobDisplayName := T(m.Lang, targetMob.NameKey)

		if (targetMob.Type == MobSkeleton || targetMob.Type == MobGolem || targetMob.Type == MobGargoyle) && rand.Intn(100) < 25 {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.mob_block", mobDisplayName)))
			return
		}

		d20 := rand.Intn(20) + 1
		hitRoll := d20 + (h.TotalAtk() / 3)
		targetAC := 10 + targetMob.Defense
		critThreshold := 19
		for _, it := range []*EquipItem{h.Weapon, h.Head, h.Chest} {
			if it != nil {
				critThreshold -= it.CritBonus
			}
		}
		if raceMod.CritChanceBonus > 0 {
			critThreshold -= raceMod.CritChanceBonus
		}
		if h.IsStealthed {
			critThreshold = 1
		}
		if critThreshold < 14 && !h.IsStealthed {
			critThreshold = 14
		}
		bonusDmg := 0
		armorPierce := 0

		// ==================== МАЛЫЕ НАВЫКИ (3-5 MP) ====================
		switch h.Class {
		case ClassMage:
			if h.MP >= 5 {
				h.MP -= 5
				bonusDmg += 5
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.fire_arrow", hName)))
			}
		case ClassCleric:
			if h.MP >= 4 {
				h.MP -= 4
				for _, ally := range m.Party {
					if !ally.IsDead && ally.Stress > 0 {
						ally.Stress = max(0, ally.Stress-6)
						h.RemoveDot()
						m.addLog(fountStyle.Render(T(m.Lang, "combat.log.blessing", hName, ally.DisplayName(m.Lang))))
						break
					}
				}
			}
		case ClassWarrior:
			if h.MP >= 4 && h.MP < skillCost {
				h.MP -= 4
				bonusDmg += 3
				armorPierce = 3
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.crush_strike", hName)))
			}
		case ClassRogue:
			if h.MP >= 3 && h.MP < skillCost {
				h.MP -= 3
				critThreshold = 17
				h.StealLoot()
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.quick_cut", hName)))
			}
		case ClassTank:
			if h.MP >= 4 {
				h.MP -= 4
				h.BaseDef += 2
				h.PullAggro()
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.taunt", hName)))
			}
		case ClassPaladin:
			if h.MP >= 3 {
				h.MP -= 3
				h.BaseDef += 2
				h.Stress = max(0, h.Stress-4)
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.paladin_minor", hName)))
			}
		case ClassRanger:
			if h.MP >= 4 {
				h.MP -= 4
				armorPierce = 4
				bonusDmg += 3
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.ranger_minor", hName)))
			}
		case ClassMonk:
			if h.MP >= 3 {
				h.MP -= 3
				bonusDmg += 4
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.monk_minor", hName)))
			}
		case ClassBard:
			if h.MP >= 3 {
				h.MP -= 3
				for _, ally := range m.Party {
					if !ally.IsDead && (ally.Class == ClassMage || ally.Class == ClassCleric || ally.Class == ClassWarlock) && ally.MP < ally.MaxMP {
						ally.MP = min(ally.MaxMP, ally.MP+5)
						m.addLog(potionStyle.Render(T(m.Lang, "combat.log.bard_minor", hName, ally.DisplayName(m.Lang))))
						break
					}
				}
			}
		case ClassWarlock:
			if h.MP >= 4 {
				h.MP -= 4
				armorPierce = 6
				bonusDmg += 2
				m.addLog(stressStyle.Render(T(m.Lang, "combat.log.warlock_minor", hName)))
			}
		}

		isCrit := d20 >= critThreshold || h.IsStealthed
		isFumble := d20 == 1 && !h.IsStealthed

		if isFumble {
			verb := TVerb(m.Lang, h.Gender, "промахнулся", "промахнулась", "missed")
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.fumble", hName, verb)))
			return
		}
		if hitRoll < targetAC && !isCrit {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.armor_deflect", mobDisplayName, hName)))
			return
		}

		effectiveDef := max(0, targetMob.Defense-armorPierce)
		varRand := max(4, h.TotalAtk()/2)
		dmg := h.TotalAtk() + rand.Intn(varRand) - (effectiveDef / 2) + bonusDmg
		dmg = max(4+(m.Floor/2), dmg)

		if isCrit {
			dmg = int(float64(dmg) * 1.8)
			h.Feats.CritsLanded++
			h.AddCriticalStrike()
			m.checkAndAwardTitle(h)
			m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.crit", d20, hName)))
			h.IsStealthed = false
		}

		targetMob.HP -= dmg
		h.Feats.DamageDealt += dmg

		// Вампиризм (Зверолюди или экипировка)
		lifesteal := raceMod.LifeStealPercent
		if it := h.Weapon; it != nil && it.Suffix != nil && it.Suffix.Effect == SuffVampirism {
			lifesteal += 0.25
		}
		if lifesteal > 0 {
			leechHP := int(float64(dmg) * lifesteal)
			if leechHP > 0 {
				h.HP = min(h.MaxHP, h.HP+leechHP)
			}
		}

		m.checkAndAwardTitle(h)
		hitVerb := TVerb(m.Lang, h.Gender, "нанес", "нанесла", "dealt")
		m.addLog(fmt.Sprintf("⚔️ %s %s %d урона [%s] (%d HP).", hName, hitVerb, dmg, mobDisplayName, targetMob.HP))

		if targetMob.HP <= 0 {
			targetMob.HP = 0
			targetMob.IsDead = true
			h.Feats.Kills++
			m.distributePartyExp(targetMob.Exp)
			g := int(float64(targetMob.Exp) * m.Relic.GoldMult)
			m.Gold += g
			m.Stats.TotalGoldEarned += g
			m.Stats.MonsterKills[targetMob.Type]++
			m.checkQuestProgress(QuestHuntMonster, targetMob.Type, 1)

			if targetMob.Type == MobDragon {
				h.Feats.BossKills++
				bossItem := generateItemForClass(h.Class, m.Floor+2)
				bossItem.UpgradeLevel = 4
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.boss_relic", bossItem.DisplayName(m.Lang))))
				m.equipOrBag(bossItem)
			}
			m.checkAndAwardTitle(h)
		}
		return
	}

	if current.Type == CombatantMonster {
		mob := current.MonsterRef
		if mob.IsDead {
			return
		}

		victim := m.SelectTarget()
		if victim == nil {
			return
		}

		m.checkAndDrinkPotions(victim)

		mobDisplayName := T(m.Lang, mob.NameKey)
		victimName := victim.DisplayName(m.Lang)
		isRanged := (mob.Type == MobImp || mob.Type == MobPhantom || mob.Type == MobVoidDemon || mob.Type == MobDragon)
		if isRanged {
			m.addLog(fireStyle.Render(T(m.Lang, "combat.log.mob_ranged", mobDisplayName, victimName)))
		} else {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.mob_melee", mobDisplayName, victimName)))
		}

		mobRoll := rand.Intn(20) + 1
		mobHit := mobRoll + (mob.Atk / 3)
		heroAC := 10 + victim.TotalDef()

		if mobRoll == 1 {
			verb := TVerb(m.Lang, victim.Gender, "увернулся", "увернулась", "dodged")
			m.addLog(healStyle.Render(T(m.Lang, "combat.log.dodge", victimName, verb, mobDisplayName)))
			return
		}
		if mobHit < heroAC && mobRoll < 18 {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.armor_absorb", victimName, mobDisplayName)))
			return
		}

		// Смертоносная формула урона монстров v2.4.3
		rawDmg := int(float64(mob.Atk)*1.25) - (victim.TotalDef() / 2)
		if victim.IsGuarding {
			rawDmg = int(float64(rawDmg) * 0.6)
		}

		livingCount := m.Combat.Pack.LivingCount()
		if livingCount >= 6 {
			rawDmg = int(float64(rawDmg) * 0.85)
		} else if livingCount >= 4 {
			rawDmg = int(float64(rawDmg) * 0.92)
		}

		if mobRoll >= 18 {
			rawDmg = int(float64(rawDmg) * 2.0)
			m.addStress(victim, 35)
			m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.mob_crit", mobDisplayName, victimName)))
		}

		floorScale := 1.0 + float64(m.Floor)*0.015
		inDmg := max(5, int(float64(rawDmg)*m.Relic.EnemyDmgMod*enrageMult*floorScale))

		actualHero := victim
		guarded := false
		var dealtDmg int

		if !isRanged {
			actualHero, dealtDmg, guarded = m.ApplyDamage(victim, inDmg)
		} else {
			victim.HP -= inDmg
			dealtDmg = inDmg
		}

		actualHeroName := actualHero.DisplayName(m.Lang)

		if guarded {
			verb := TVerb(m.Lang, actualHero.Gender, "прикрыл собой", "прикрыла собой", "shielded")
			m.addLog(healStyle.Render(T(m.Lang, "combat.log.tank_guard", actualHeroName, verb, victimName, mobDisplayName, dealtDmg)))
		} else {
			m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.mob_hit", mobDisplayName, dealtDmg, actualHeroName)))
		}

		actualHero.Feats.DamageTaken += dealtDmg
		m.addStress(actualHero, 12)

		if actualHero.HP <= 0 {
			actualHero.HP = 0
			actualHero.CauseOfDeath = T(m.Lang, "combat.log.death_cause_mob", mobDisplayName)
			m.recordFallenHero(actualHero)
			verb := TVerb(m.Lang, actualHero.Gender, "пал", "пала", "fell")
			m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.hero_slain", actualHeroName, verb, mobDisplayName)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					m.addStress(ally, 25)
				}
			}
		} else {
			if float64(actualHero.HP)/float64(actualHero.MaxHP) <= 0.20 {
				actualHero.Feats.NearDeathEscapes++
			}
			m.checkAndAwardTitle(actualHero)
		}
	}
}