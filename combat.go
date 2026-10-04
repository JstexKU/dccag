package main

import (
	"sort"
)

func (m *Model) SelectTarget() *Hero {
	var candidates []*Hero
	totalWeight := 0

	for _, h := range m.Party {
		if !h.IsDead && !h.IsDowned {
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

	r := rng.Intn(totalWeight)
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
			if !guard.IsDead && !guard.IsDowned && guard.Role.CanGuard && guard != target && guard.HP > guard.MaxHP/4 {
				chance := 30
				mitigation := 0.75
				if guard.Class == ClassTank {
					chance = 60
					mitigation = 0.50
				} else if guard.Class == ClassPaladin {
					chance = 50
					mitigation = 0.60
				}

				if rng.Intn(100) < chance {
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
		HasBarrel:    rng.Intn(100) < 30,
		FleeCooldown: 0,
		Round:        1,
	}

	biome := getBiome(m.Floor)
	speedPenalty := 0
	if biome.Name == BiomeGrotto {
		speedPenalty = 2
	}

	for _, h := range m.Party {
		if !h.IsDead && !h.IsDowned {
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
			initRoll := rng.Intn(20) + 1 + h.TotalSpeed() + m.Legacy.TanneryLevel - speedPenalty
			combat.TurnQueue = append(combat.TurnQueue, TurnOrderEntry{
				Type: CombatantHero, HeroRef: h, Initiative: initRoll,
			})
		}
	}

	for _, mob := range pack.Members {
		if !mob.IsDead {
			initRoll := rng.Intn(20) + 1 + mob.Speed
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

// onMonsterKilled — единая точка обработки гибели монстра от руки героя h.
func (m *Model) onMonsterKilled(h *Hero, mob *Monster) {
	mob.HP = 0
	mob.IsDead = true
	h.Feats.Kills++
	m.distributePartyExp(mob.Exp)

	g := int(float64(mob.Exp) * m.goldMult())
	m.Gold += g
	m.Stats.TotalGoldEarned += g
	m.Stats.MonsterKills[mob.Type]++
	m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)

	// Награда и прогресс за уничтожение Мини-босса
	if m.Combat.Pack.IsMiniBoss && mob.ID == 1 {
		m.checkQuestProgress(QuestHuntMiniBoss, mob.Type, 1)
		miniBossItem := generateItemForClass(h.Class, m.Floor+1)
		miniBossItem.UpgradeLevel = 2
		m.addLog(accentStyle.Render(T(m.Lang, "combat.log.miniboss_trophy", miniBossItem.DisplayName(m.Lang))))
		m.equipOrBag(miniBossItem)
	}

	if mob.Type == MobDragon {
		h.Feats.BossKills++
		bossItem := generateItemForClass(h.Class, m.Floor+2)
		bossItem.UpgradeLevel = 4
		m.addLog(accentStyle.Render(T(m.Lang, "combat.log.boss_relic", bossItem.DisplayName(m.Lang))))
		m.equipOrBag(bossItem)
	}
	m.checkAndAwardTitle(h)
}
