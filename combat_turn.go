package main

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

	// 2. Поражение (все погибли или лежат без сознания)
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.Combat = nil
		m.RestartCountdown = 10
		return
	}

	// 3–4. Экстренное отступление и побег — решения автопилота; в ручном режиме их принимает игрок (F).
	if !m.ManualMode {
		retreatReason := m.evaluateRetreat()
		if m.shouldForceRetreatFromCombat(retreatReason) {
			m.forceRetreatToTown(retreatReason)
			return
		}

		if m.shouldAttemptFlee() {
			m.attemptFlee()
			if m.Combat == nil {
				return
			}
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
		if h.IsDead || h.IsDowned {
			return
		}

		// Команда игрока (ручной режим) забирается один раз и сбрасывается.
		var cmd ManualCmd
		if m.ManualMode {
			cmd = m.Combat.Cmd
			m.Combat.Cmd = ManualCmd{}
			m.Combat.PotionMenu = false
		}
		forceSkill := cmd.Kind == CmdSkill
		useSkills := cmd.Kind != CmdStrike

		// Расовая пассивка: регенерация маны эльфа
		raceMod := GetRaceModifiers(h.Race)
		if raceMod.ManaRegen > 0 && h.MP < h.MaxMP {
			h.MP = min(h.MaxMP, h.MP+raceMod.ManaRegen)
		}

		m.checkAndDrinkPotions(h)
		hName := h.DisplayName(m.Lang)

		if h.Affliction == AfflictionParanoid && rng.Intn(100) < 35 {
			verb := TVerb(m.Lang, h.Gender, "забился", "забилась", "cowered")
			m.addLog(stressStyle.Render(T(m.Lang, "combat.log.paranoid", hName, verb)))
			return
		}

		switch cmd.Kind {
		case CmdGuard:
			m.manualGuard(h)
			return
		case CmdPotion:
			m.manualPotion(h, cmd)
			return
		}
		mpBefore := h.MP

		skillCost := h.SkillCost
		if biome.Name == BiomeCrystal {
			skillCost += 3
		}

		targetMob := m.pickTarget()

		// ==================== БОЛЬШИЕ И СРЕДНИЕ УМЕНИЯ (10 КЛАССОВ) ====================

		// 1. ТАНК
		if useSkills && h.Class == ClassTank {
			if h.MP >= skillCost && !h.IsGuarding {
				h.MP -= skillCost
				h.IsGuarding = true
				h.AddBlock()
				h.PullAggro()
				m.checkAndAwardTitle(h)
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.tank_stance", hName)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.Atk >= 12 && (forceSkill || rng.Intn(100) < 45) {
				h.MP -= 6
				bashDmg := h.TotalDef() + 4
				targetMob.HP -= bashDmg
				targetMob.Atk = max(2, targetMob.Atk-3)
				h.AddBlock()
				m.addLog(healStyle.Render(T(m.Lang, "combat.log.tank_bash", hName, bashDmg)))
				if targetMob.HP <= 0 {
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		// 2. ПАЛАДИН
		if useSkills && h.Class == ClassPaladin {
			var woundedAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && !ally.IsDowned && float64(ally.HP)/float64(ally.MaxHP) <= 0.45 {
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
			} else if h.MP >= 6 && targetMob != nil && (forceSkill || rng.Intn(100) < 50) {
				h.MP -= 6
				smiteDmg := h.TotalAtk() + (h.TotalDef() / 3) + 3
				targetMob.HP -= smiteDmg
				h.IsGuarding = true
				h.AddBlock()
				m.addLog(goldStyle.Render(T(m.Lang, "combat.log.paladin_smite", hName, smiteDmg)))
				if targetMob.HP <= 0 {
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		// 3. ВОИН
		if useSkills && h.Class == ClassWarrior {
			if h.MP >= skillCost && !h.IsBerserk {
				h.MP -= skillCost
				h.IsBerserk = true
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.warrior_rage", hName)))
			} else if h.MP >= 7 && m.Combat.Pack.LivingCount() >= 2 && (forceSkill || rng.Intn(100) < 55) {
				h.MP -= 7
				cleaveDmg := h.TotalAtk() + 2
				hitCount := 0
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && hitCount < 2 {
						mob.HP -= cleaveDmg
						h.Feats.DamageDealt += cleaveDmg
						if mob.HP <= 0 {
							m.onMonsterKilled(h, mob)
						}
						hitCount++
					}
				}
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.warrior_cleave", hName, hitCount, cleaveDmg)))
				return
			}
		}

		// 4. МОНАХ
		if useSkills && h.Class == ClassMonk {
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
							m.onMonsterKilled(h, mob)
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
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		// 5. РАЗБОЙНИК
		if useSkills && h.Class == ClassRogue {
			if h.MP >= skillCost && !h.IsStealthed {
				h.MP -= skillCost
				h.IsStealthed = true
				h.AddBackstab()
				verb := TVerb(m.Lang, h.Gender, "растворился", "растворилась", "vanished")
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.rogue_stealth", hName, verb)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.HP > 20 && (forceSkill || rng.Intn(100) < 50) {
				h.MP -= 6
				poisonDmg := h.TotalAtk() + 6
				targetMob.HP -= poisonDmg
				h.Feats.DamageDealt += poisonDmg
				m.addLog(stressStyle.Render(T(m.Lang, "combat.log.rogue_poison", hName, poisonDmg)))
				if targetMob.HP <= 0 {
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		// 6. СЛЕДОПЫТ
		if useSkills && h.Class == ClassRanger {
			if targetMob != nil && h.MP >= skillCost && targetMob.Atk >= 12 && (forceSkill || rng.Intn(100) < 60) {
				h.MP -= skillCost
				trapDmg := 8 + (m.Floor * 2)
				targetMob.HP -= trapDmg
				targetMob.Atk = max(2, targetMob.Atk-5)
				h.AddCCDuration()
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.ranger_trap", hName, T(m.Lang, targetMob.NameKey), trapDmg)))
				if targetMob.HP <= 0 {
					m.onMonsterKilled(h, targetMob)
				}
				return
			} else if h.MP >= 6 && m.Combat.Pack.LivingCount() >= 2 {
				h.MP -= 6
				volleyDmg := h.TotalAtk() + 1
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && (mob.Type == MobImp || mob.Type == MobPhantom || mob.Type == MobVoidDemon || mob.Type == MobDryad || mob.Type == MobTomeBook || mob.Type == MobAstralWeaver) {
						mob.HP -= volleyDmg
						h.Feats.DamageDealt += volleyDmg
						if mob.HP <= 0 {
							m.onMonsterKilled(h, mob)
						}
					}
				}
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.ranger_volley", hName, volleyDmg)))
				return
			}
		}

		// 7. КЛИРИК
		if useSkills && h.Class == ClassCleric {
			// А. Приоритет: поднятие на ноги союзника без сознания
			var downedAlly *Hero
			for _, ally := range m.Party {
				if ally.IsDowned && !ally.IsDead {
					downedAlly = ally
					break
				}
			}

			if downedAlly != nil && h.MP >= (skillCost+4) && h.ReviveCooldown <= 0 {
				h.MP -= (skillCost + 4)
				h.ReviveCooldown = 3
				downedAlly.IsDowned = false
				reviveHP := max(12, downedAlly.MaxHP/4)
				downedAlly.HP = reviveHP
				downedAlly.Stress = max(0, downedAlly.Stress-20)
				h.AddRevive()
				m.checkAndAwardTitle(h)
				m.Stats.Resurrections++

				// Добавляем бойца в спасённые
				m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
					FullName: downedAlly.FullName(m.Lang),
					Class:    downedAlly.Class,
					Cause:    T(m.Lang, "combat.log.cleric_revive_cause"),
					Floor:    m.Floor,
					Revived:  true,
				})

				m.addLog(healStyle.Render(T(m.Lang, "combat.log.cleric_revive", hName, downedAlly.DisplayName(m.Lang), reviveHP)))
				return
			}

			// Б. Обычное исцеление раненых
			var criticalAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && !ally.IsDowned && float64(ally.HP)/float64(ally.MaxHP) <= 0.45 {
					criticalAlly = ally
					break
				}
			}

			if criticalAlly != nil && h.MP >= skillCost && h.Affliction != AfflictionSelfish {
				h.MP -= skillCost
				hAmt := rng.Intn(8) + 12 + (m.Floor * 2)
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
			} else if h.MP >= 7 && targetMob != nil && (forceSkill || rng.Intn(100) < 40) {
				h.MP -= 7
				smiteDmg := h.TotalAtk() + 4
				targetMob.HP -= smiteDmg
				if lowest := m.getRandomLivingHero(); lowest != nil && lowest.HP < lowest.MaxHP {
					lowest.HP = min(lowest.MaxHP, lowest.HP+4)
				}
				verb := TVerb(m.Lang, h.Gender, "обрушил", "обрушила", "unleashed")
				m.addLog(fountStyle.Render(T(m.Lang, "combat.log.cleric_smite", hName, verb, smiteDmg)))
				if targetMob.HP <= 0 {
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		// 8. БАРД
		if useSkills && h.Class == ClassBard {
			hasHighStress := false
			for _, ally := range m.Party {
				if !ally.IsDead && !ally.IsDowned && ally.Stress >= 50 {
					hasHighStress = true
					break
				}
			}

			if hasHighStress && h.MP >= skillCost {
				h.MP -= skillCost
				for _, ally := range m.Party {
					if !ally.IsDead && !ally.IsDowned {
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
							m.onMonsterKilled(h, mob)
						}
					}
				}
				m.addLog(stressStyle.Render(T(m.Lang, "combat.log.bard_dissonance", hName, dissonanceDmg)))
				return
			}
		}

		// 9. МАГ
		if useSkills && h.Class == ClassMage {
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
							m.onMonsterKilled(h, mob)
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
				m.addLog(fireStyle.Render(T(m.Lang, "combat.log.mage_storm", hName, verb, aoeDmg)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= aoeDmg
						h.Feats.DamageDealt += aoeDmg
						if mob.HP <= 0 {
							m.onMonsterKilled(h, mob)
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
							m.onMonsterKilled(h, mob)
						}
						hits++
					}
				}
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.mage_lightning", hName, chainDmg)))
				return
			} else {
				if eliteMob := m.Combat.Pack.GetHighestHPFocus(); eliteMob != nil && !m.ManualMode {
					targetMob = eliteMob
				}
			}
		}

		// 10. ЧЕРНОКНИЖНИК
		if useSkills && h.Class == ClassWarlock {
			if h.HP > 15 && h.MP >= skillCost {
				needMana := false
				for _, ally := range m.Party {
					if !ally.IsDead && !ally.IsDowned && ally != h && float64(ally.MP)/float64(ally.MaxMP) <= 0.30 {
						needMana = true
						break
					}
				}

				if needMana {
					h.MP -= skillCost
					h.HP -= 6
					for _, ally := range m.Party {
						if !ally.IsDead && !ally.IsDowned && ally != h {
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
					m.onMonsterKilled(h, targetMob)
				}
				return
			}
		}

		if targetMob == nil {
			return
		}

		if cmd.Kind == CmdSkill && h.MP == mpBefore {
			m.addLog(subtleStyle.Render(T(m.Lang, "ctl.no_skill", hName)))
		}

		mobDisplayName := T(m.Lang, targetMob.NameKey)

		if (targetMob.Type == MobSkeleton || targetMob.Type == MobGolem || targetMob.Type == MobGargoyle || targetMob.Type == MobSproutSkeleton || targetMob.Type == MobObsidianBeetle || targetMob.Type == MobFallenCrusader) && rng.Intn(100) < 25 {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.mob_block", mobDisplayName)))
			return
		}

		d20 := rng.Intn(20) + 1
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
		minorClass := h.Class
		if !useSkills {
			minorClass = ""
		}
		switch minorClass {
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
					if !ally.IsDead && !ally.IsDowned && ally.Stress > 0 {
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
					if !ally.IsDead && !ally.IsDowned && (ally.Class == ClassMage || ally.Class == ClassCleric || ally.Class == ClassWarlock) && ally.MP < ally.MaxMP {
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
		dmg := h.TotalAtk() + rng.Intn(varRand) - (effectiveDef / 2) + bonusDmg
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
		m.addLog(T(m.Lang, "combat.log.hit", hName, hitVerb, dmg, mobDisplayName, targetMob.HP))

		if targetMob.HP <= 0 {
			m.onMonsterKilled(h, targetMob)
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
		isRanged := (mob.Type == MobImp || mob.Type == MobPhantom || mob.Type == MobVoidDemon || mob.Type == MobDragon ||
			mob.Type == MobDryad || mob.Type == MobTomeBook || mob.Type == MobBloodCultist || mob.Type == MobAstralWeaver)
		if isRanged {
			m.addLog(fireStyle.Render(T(m.Lang, "combat.log.mob_ranged", mobDisplayName, victimName)))
		} else {
			m.addLog(subtleStyle.Render(T(m.Lang, "combat.log.mob_melee", mobDisplayName, victimName)))
		}

		mobRoll := rng.Intn(20) + 1
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
			actualHero.IsDowned = true
			actualHero.IsGuarding = false
			actualHero.IsBerserk = false
			actualHero.IsStealthed = false
			actualHero.IsCharged = false
			actualHero.IsAura = false
			actualHero.CauseOfDeath = T(m.Lang, "combat.log.death_cause_mob", mobDisplayName)

			verb := TVerb(m.Lang, actualHero.Gender, "рухнул без сознания", "рухнула без сознания", "fell unconscious")
			m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.hero_downed", actualHeroName, verb, mobDisplayName)))
			m.onLeaderDowned(actualHero)
			for _, ally := range m.Party {
				if !ally.IsDead && !ally.IsDowned {
					m.addStress(ally, 18)
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
