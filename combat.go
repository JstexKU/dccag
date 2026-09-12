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
				w += 20
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
			w += 20
		}
		curr += w
		if r < curr {
			return h
		}
	}
	return candidates[0]
}

func (m *Model) ApplyDamage(target *Hero, rawDmg int) (actual *Hero, finalDmg int, guarded bool) {
	if target.Class != ClassTank {
		for _, guard := range m.Party {
			if !guard.IsDead && guard.Role.CanGuard && guard != target && guard.HP > guard.MaxHP/4 {
				chance := 35
				mitigation := 0.8
				if guard.Class == ClassTank {
					chance = 65
					mitigation = 0.55
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
		HasBarrel:    rand.Intn(100) < 35,
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
	m.addLog(accentStyle.Render(fmt.Sprintf("⚔️ СХВАТКА! Вражеский отряд (%d тварей)!", pack.LivingCount())))
}

func (m *Model) shouldAttemptFlee() bool {
	if m.Combat == nil || m.Combat.FleeCooldown > 0 {
		return false
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
	return hpPercent < 0.25 || (livingCount == 1 && hpPercent < 0.50)
}

func (m *Model) attemptFlee() {
	globalDebugReport.FleeAttempts++
	chance := 50

	for _, h := range m.Party {
		if !h.IsDead {
			if h.Class == ClassRogue {
				chance += 25
			}
			if h.Class == ClassTank {
				chance += 10
			}
		}
	}

	chance -= m.Floor / 3
	if chance < 35 {
		chance = 35
	}

	roll := rand.Intn(100)
	if roll < chance {
		globalDebugReport.FleeSuccesses++
		m.addLog(healStyle.Render("💨 [ПОБЕГ] Успех! Отряд оторвался от погони под прикрытием завесы."))

		evacuatedCount := 0
		for _, h := range m.Party {
			if h.IsDead {
				h.HP = 1
				h.IsDead = false
				h.CauseOfDeath = ""
				evacuatedCount++
			}
		}
		if evacuatedCount > 0 {
			m.addLog(altarStyle.Render(fmt.Sprintf("🕊️ [Эвакуация] Отряд вынес с поля боя %d павших героев!", evacuatedCount)))
		}

		m.Combat = nil
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = "Отряд отступил в Столицу."
		m.addLog(dangerStyle.Render("🏰 Отряд укрылся за стенами Города!"))
	} else {
		m.addLog(dangerStyle.Render("💥 [ПРОВАЛ ПОБЕГА] Монстры перекрыли отход! Отряд перегруппировался под градом скользящих ударов."))

		for _, h := range m.Party {
			if !h.IsDead {
				chipDamage := int(float64(h.HP) * 0.15)
				if chipDamage < 1 {
					chipDamage = 1
				}
				h.HP -= chipDamage
				h.Feats.DamageTaken += chipDamage
				m.addStress(h, 8)

				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = "Зарублен при неудачном отходе"
					m.recordFallenHero(h)
				}
			}
		}
		m.Combat.FleeCooldown = 3
	}
}

func (m *Model) executeCombatTurn() {
	if m.Combat == nil {
		return
	}

	if m.Combat.FleeCooldown > 0 {
		m.Combat.FleeCooldown--
	}

	if m.shouldAttemptFlee() {
		m.attemptFlee()
		if m.Combat == nil {
			return
		}
	}

	if m.Combat.Pack.LivingCount() == 0 {
		m.addLog(healStyle.Render("💀 Вражеский отряд повержен!"))
		delete(m.Packs, m.Combat.Pos)
		m.PartyPos = m.Combat.Pos
		m.Combat = nil
		m.revealFog()
		return
	}

	if m.isPartyWiped() {
		m.State = StateDefeat
		m.Combat = nil
		m.RestartCountdown = 10
		return
	}

	if m.Combat.TurnIdx >= len(m.Combat.TurnQueue) {
		m.Combat.TurnIdx = 0
		m.Combat.Round++
	}
	current := m.Combat.TurnQueue[m.Combat.TurnIdx]
	m.Combat.TurnIdx++

	biome := getBiome(m.Floor)

	enrageMult := 1.0
	if m.Combat.Round > 20 {
		globalDebugReport.EnrageProcs++
		enrageMult += float64(m.Combat.Round-20) * 0.10
	}

	if current.Type == CombatantHero {
		h := current.HeroRef
		if h.IsDead {
			return
		}

		m.checkAndDrinkPotions(h)

		if h.Affliction == AfflictionParanoid && rand.Intn(100) < 30 {
			m.addLog(stressStyle.Render(fmt.Sprintf("👁️ %s %s в угол (Паранойя)!", h.Name, h.Verb("забился", "забилась"))))
			return
		}

		skillCost := h.SkillCost
		if biome.Name == BiomeCrystal {
			skillCost += 3
		}

		targetMob := m.Combat.Pack.GetLowestHPFocus()

		// 1. ТАНК
		if h.Class == ClassTank {
			if h.MP >= skillCost && !h.IsGuarding {
				h.MP -= skillCost
				h.IsGuarding = true
				h.AddBlock()
				h.PullAggro()
				m.checkAndAwardTitle(h)
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s принимает [Оборонительную Стойку] (+5 Def, блок)!", h.Name)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.Atk >= 12 && rand.Intn(100) < 45 {
				h.MP -= 6
				bashDmg := h.TotalDef() + 4
				targetMob.HP -= bashDmg
				targetMob.Atk = max(2, targetMob.Atk-3)
				h.AddBlock()
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s проводит [Удар щитом] (-%d HP, враг ослаблен на -3 Atk)!", h.Name, bashDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 2. ВОИН
		if h.Class == ClassWarrior {
			if h.MP >= skillCost && !h.IsBerserk {
				h.MP -= skillCost
				h.IsBerserk = true
				m.addLog(fireStyle.Render(fmt.Sprintf("⚔️ %s входит в [Состояние Ярости] (+5 Atk, -2 Def)!", h.Name)))
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
				m.addLog(dangerStyle.Render(fmt.Sprintf("⚔️ %s выполняет [Рассечение] по %d врагам (-%d HP)!", h.Name, hitCount, cleaveDmg)))
				return
			}
		}

		// 3. РАЗБОЙНИК
		if h.Class == ClassRogue {
			if h.MP >= skillCost && !h.IsStealthed {
				h.MP -= skillCost
				h.IsStealthed = true
				h.AddBackstab()
				m.addLog(accentStyle.Render(fmt.Sprintf("🗡️ %s %s в тенях [Скрытность] (100%% крит)!", h.Name, h.Verb("растворился", "растворилась"))))
			} else if h.MP >= 6 && targetMob != nil && targetMob.HP > 20 && rand.Intn(100) < 50 {
				h.MP -= 6
				poisonDmg := h.TotalAtk() + 6
				targetMob.HP -= poisonDmg
				h.Feats.DamageDealt += poisonDmg
				m.addLog(stressStyle.Render(fmt.Sprintf("☣️ %s наносит [Отравленный выпад] (-%d HP)!", h.Name, poisonDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 4. КЛИРИК
		if h.Class == ClassCleric {
			var criticalAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && float64(ally.HP)/float64(ally.MaxHP) <= 0.50 {
					criticalAlly = ally
					break
				}
			}

			if criticalAlly != nil && h.MP >= skillCost && h.Affliction != AfflictionSelfish {
				h.MP -= skillCost
				hAmt := rand.Intn(10) + 16 + (m.Floor * 3)
				if m.Combat.Round > 20 {
					hAmt /= 2
				}
				criticalAlly.HP = min(criticalAlly.MaxHP, criticalAlly.HP+hAmt)
				criticalAlly.Stress = max(0, criticalAlly.Stress-12)
				h.Feats.HealsGiven += hAmt
				m.addLog(healStyle.Render(fmt.Sprintf("✨ %s %s %s (+%d HP)!", h.Name, h.Verb("исцелил", "исцелила"), criticalAlly.Name, hAmt)))
				m.checkAndAwardTitle(h)
				return
			} else if !h.IsAura && h.MP >= (skillCost+6) {
				h.MP -= skillCost
				h.IsAura = true
				m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s раскрывает [Ауру Защиты] (+3 Def отряду)!", h.Name)))
			} else if h.MP >= 7 && targetMob != nil && rand.Intn(100) < 40 {
				h.MP -= 7
				smiteDmg := h.TotalAtk() + 4
				targetMob.HP -= smiteDmg
				if lowest := m.getRandomLivingHero(); lowest != nil && lowest.HP < lowest.MaxHP {
					lowest.HP = min(lowest.MaxHP, lowest.HP+6)
				}
				m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s %s [Священную кару] (-%d HP)!", h.Name, h.Verb("обрушил", "обрушила"), smiteDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 5. МАГ
		if h.Class == ClassMage {
			if m.Combat.HasBarrel && h.MP >= skillCost {
				h.MP -= skillCost
				m.Combat.HasBarrel = false
				m.Stats.BarrelsBlown++
				h.AddManaBurst()
				barrelDmg := 22 + (m.Floor * 3)
				m.addLog(barrelStyle.Render(fmt.Sprintf("💥 %s ПОДРЫВАЕТ БОЧКУ СО СМОЛОЙ (-%d HP отряду врагов)!", h.Name, barrelDmg)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= barrelDmg
						h.Feats.DamageDealt += barrelDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult * 2)
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
				aoeDmg := h.TotalAtk() + 6
				h.AddManaBurst()
				h.AddCCDuration()
				m.addLog(fireStyle.Render(fmt.Sprintf("🔥 %s %s врагов [Огненной Бурей]!", h.Name, h.Verb("накрыл", "накрыла"))))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= aoeDmg
						h.Feats.DamageDealt += aoeDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult * 2)
							m.Gold += g
							m.Stats.TotalGoldEarned += g
							m.Stats.MonsterKills[mob.Type]++
							m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)
						}
					}
				}
				m.checkAndAwardTitle(h)
				return
			} else {
				if eliteMob := m.Combat.Pack.GetHighestHPFocus(); eliteMob != nil {
					targetMob = eliteMob
				}
			}
		}

		if targetMob == nil {
			return
		}

		if (targetMob.Type == MobSkeleton || targetMob.Type == MobGolem || targetMob.Type == MobGargoyle) && rand.Intn(100) < 20 {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ [%s] отразил выпад монолитным блоком!", targetMob.Name)))
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
		if h.IsStealthed {
			critThreshold = 1
		}
		if critThreshold < 14 && !h.IsStealthed {
			critThreshold = 14
		}
		bonusDmg := 0
		armorPierce := 0

		switch h.Class {
		case ClassMage:
			if h.MP >= 5 {
				h.MP -= 5
				bonusDmg += 5
				m.addLog(fireStyle.Render(fmt.Sprintf("🔥 %s выпускает [Огненную стрелу] (+5 ур)!", h.Name)))
			}
		case ClassCleric:
			if h.MP >= 4 {
				h.MP -= 4
				for _, ally := range m.Party {
					if !ally.IsDead && ally.Stress > 0 {
						ally.Stress = max(0, ally.Stress-6)
						h.RemoveDot()
						m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s: [Благословение] (-6 стресса %s)!", h.Name, ally.Name)))
						break
					}
				}
			}
		case ClassWarrior:
			if h.MP >= 4 && h.MP < skillCost {
				h.MP -= 4
				bonusDmg += 3
				armorPierce = 3
				m.addLog(dangerStyle.Render(fmt.Sprintf("⚔️ %s: [Сокрушающий выпад]!", h.Name)))
			}
		case ClassRogue:
			if h.MP >= 3 && h.MP < skillCost {
				h.MP -= 3
				critThreshold = 17
				h.StealLoot()
				m.addLog(accentStyle.Render(fmt.Sprintf("🗡️ %s: [Быстрый порез]!", h.Name)))
			}
		case ClassTank:
			if h.MP >= 4 {
				h.MP -= 4
				h.BaseDef += 2
				h.PullAggro()
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s: [Провокация] (+2 Защ)!", h.Name)))
			}
		}

		isCrit := d20 >= critThreshold || h.IsStealthed
		isFumble := d20 == 1 && !h.IsStealthed

		if isFumble {
			m.addLog(subtleStyle.Render(fmt.Sprintf("💨 %s %s (D20=1)!", h.Name, h.Verb("промахнулся", "промахнулась"))))
			return
		}
		if hitRoll < targetAC && !isCrit {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ Броня [%s] отразила удар %s.", targetMob.Name, h.Name)))
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
			m.addLog(dangerStyle.Render(fmt.Sprintf("💥 КРИТ (D20=%d)! %s сокрушает врага!", d20, h.Name)))
			h.IsStealthed = false
		}

		targetMob.HP -= dmg
		h.Feats.DamageDealt += dmg
		m.checkAndAwardTitle(h)
		m.addLog(fmt.Sprintf("⚔️ %s %s %d урона [%s] (%d HP).", h.Name, h.Verb("нанес", "нанесла"), dmg, targetMob.Name, targetMob.HP))

		if targetMob.HP <= 0 {
			targetMob.HP = 0
			targetMob.IsDead = true
			h.Feats.Kills++
			m.distributePartyExp(targetMob.Exp)
			g := int(float64(targetMob.Exp) * m.Relic.GoldMult * 2)
			m.Gold += g
			m.Stats.TotalGoldEarned += g
			m.Stats.MonsterKills[targetMob.Type]++
			m.checkQuestProgress(QuestHuntMonster, targetMob.Type, 1)

			if targetMob.Type == MobDragon {
				h.Feats.BossKills++
				bossItem := generateItemForClass(h.Class, m.Floor+2)
				bossItem.UpgradeLevel = 4
				m.addLog(accentStyle.Render(fmt.Sprintf("👑 Реликвия Дракона: [%s]!", bossItem.DisplayName())))
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

		isRanged := (mob.Type == MobImp || mob.Type == MobPhantom || mob.Type == MobVoidDemon || mob.Type == MobDragon)
		if isRanged {
			m.addLog(fireStyle.Render(fmt.Sprintf("🎯 [%s] проводит дальнобойную атаку по позициям %s!", mob.Name, victim.Name)))
		} else {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🏃 [%s] сближается вплотную для ближнего боя с %s.", mob.Name, victim.Name)))
		}

		mobRoll := rand.Intn(20) + 1
		mobHit := mobRoll + (mob.Atk / 3)
		heroAC := 10 + victim.TotalDef()

		if mobRoll == 1 {
			m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s ловко %s от выпада [%s]!", victim.Name, victim.Verb("увернулся", "увернулась"), mob.Name)))
			return
		}
		if mobHit < heroAC && mobRoll < 19 {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ Доспехи %s полностью поглотили удар [%s].", victim.Name, mob.Name)))
			return
		}

		rawDmg := mob.Atk - (victim.TotalDef() / 2)
		if victim.IsGuarding {
			rawDmg = int(float64(rawDmg) * 0.6)
		}

		livingCount := m.Combat.Pack.LivingCount()
		if livingCount >= 6 {
			rawDmg = int(float64(rawDmg) * 0.65)
		} else if livingCount >= 4 {
			rawDmg = int(float64(rawDmg) * 0.78)
		}

		if mobRoll >= 19 {
			rawDmg = int(float64(rawDmg) * 1.5)
			m.addStress(victim, 25)
			m.addLog(dangerStyle.Render(fmt.Sprintf("⚡ КРИТИЧЕСКИЙ УДАР от [%s] по %s!", mob.Name, victim.Name)))
		}

		inDmg := max(3, int(float64(rawDmg)*m.Relic.EnemyDmgMod*enrageMult))

		actualHero := victim
		guarded := false
		var dealtDmg int

		if !isRanged {
			actualHero, dealtDmg, guarded = m.ApplyDamage(victim, inDmg)
		} else {
			victim.HP -= inDmg
			dealtDmg = inDmg
		}

		if guarded {
			m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s %s %s от удара [%s], приняв %d урона!",
				actualHero.Name, actualHero.Verb("прикрыл собой", "прикрыла собой"), victim.Name, mob.Name, dealtDmg)))
		} else {
			m.addLog(dangerStyle.Render(fmt.Sprintf("💥 [%s] нанес %d урона по %s!", mob.Name, dealtDmg, actualHero.Name)))
		}

		actualHero.Feats.DamageTaken += dealtDmg
		m.addStress(actualHero, 8)

		if actualHero.HP <= 0 {
			actualHero.HP = 0
			actualHero.CauseOfDeath = fmt.Sprintf("Сражен монстром [%s]", mob.Name)
			m.recordFallenHero(actualHero)
			m.addLog(dangerStyle.Render(fmt.Sprintf("☠️ %s %s в бою от фатального удара [%s]!",
				actualHero.Name, actualHero.Verb("пал", "пала"), mob.Name)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					m.addStress(ally, 20)
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
