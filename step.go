package main

import (
	"fmt"
)

func (m *Model) step() {
	if m.State != StatePlaying {
		return
	}
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.RestartCountdown = 10
		return
	}

	// 0. Город — отдельный поток, шаги подземелья не считаются
	if m.InTown {
		m.stepTownMap()
		return
	}

	m.Stats.TotalSteps++

	// 1. Продолжение привала (RestTurnsLeft > 0) — сидим на месте
	if m.RestTurnsLeft > 0 {
		m.RestTurnsLeft--
		recoveryStep := 1
		// Реликвия «Рог Изобилия»: ускоренное восстановление на привале
		if m.Relic != nil && m.Relic.PotionPotency > 1.0 {
			recoveryStep = 2
		}

		for _, h := range m.Party {
			if h.IsDead || h.IsDowned {
				continue
			}
			if h.HP < h.MaxHP {
				h.HP = min(h.MaxHP, h.HP+recoveryStep)
			}
			if h.MP < h.MaxMP {
				h.MP = min(h.MaxMP, h.MP+recoveryStep)
			}
			if h.Stress > 0 {
				h.Stress = max(0, h.Stress-recoveryStep)
			}
		}
		return
	}

	// 2. Бой — раньше всего остального
	if m.Combat != nil {
		m.executeCombatTurn()
		return
	}

	// 3. Проверка retreat: если отряд на выходе и причина есть — уходим в город
	if reason := m.evaluateRetreat(); reason != RetreatNone &&
		m.Grid[m.PartyPos.Y][m.PartyPos.X] == TileExit &&
		m.Stats.TotalSteps > 0 {
		m.enterTown(reason)
		return
	}

	// 4. Привал только при реальной необходимости (ранения, мана, стресс)
	needsRest := false
	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		hpPct := float64(h.HP) / float64(h.MaxHP)
		mpPct := 1.0
		if h.MaxMP > 0 {
			mpPct = float64(h.MP) / float64(h.MaxMP)
		}
		if hpPct <= 0.65 || mpPct <= 0.40 || h.Stress >= 45 {
			needsRest = true
			break
		}
	}

	if needsRest &&
		m.Stats.TotalSteps%30 == 0 &&
		m.evaluateRetreat() == RetreatNone &&
		!m.isMonsterNearby() {

		m.RestTurnsLeft = rng.Intn(5) + 4 // 4..8 шагов
		biome := getBiome(m.Floor)
		variant := rng.Intn(4) + 1
		key := fmt.Sprintf("dungeon.log.rest.%s.%d", biome.Name, variant)
		m.addLog(healStyle.Render(T(m.Lang, key)))
		return
	}

	// 5. Зелья
	for _, h := range m.Party {
		m.checkAndDrinkPotions(h)
	}

	// 6. Поиск следующего шага
	var next Point

	if m.LoopDetectCount >= 5 {
		targetTile := TileStairs
		if m.evaluateRetreat() != RetreatNone || (m.Training && !m.ManualMode) {
			targetTile = TileExit
		}
		forcedStep, found := m.findEmergencyStep(targetTile)
		if found {
			next = forcedStep
		} else {
			if m.Training && !m.ManualMode {
				m.enterTown(RetreatTrainingDone)
				return
			}
			m.saveCurrentFloorState()
			m.Floor++
			m.Stats.FloorsCleared++
			m.checkQuestProgress(QuestReachFloor, "", m.Floor)
			m.checkQuestProgress(QuestEscapeTrap, "", 1)
			m.PathHistory = []Point{}
			m.LoopDetectCount = 0
			m.initDungeonForFloor(m.Floor)
			return
		}
	} else {
		next = m.findNextStep()
	}

	isOscillating := false
	histLen := len(m.PathHistory)
	if histLen >= 2 && m.PathHistory[histLen-2] == next {
		isOscillating = true
	}

	visitCount := 0
	for _, p := range m.PathHistory {
		if p == next {
			visitCount++
		}
	}

	if isOscillating || visitCount >= 3 || next == m.PartyPos {
		m.LoopDetectCount++
	} else if m.LoopDetectCount > 0 && m.LoopDetectCount < 5 {
		m.LoopDetectCount--
	}

	if m.LoopDetectCount == 5 {
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.collision_break")))
	}

	if m.LoopDetectCount >= 10 {
		if m.Training && !m.ManualMode {
			m.enterTown(RetreatTrainingDone)
			return
		}
		m.saveCurrentFloorState()
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		return
	}

	m.PathHistory = append(m.PathHistory, next)
	if len(m.PathHistory) > 36 {
		m.PathHistory = m.PathHistory[1:]
	}

	m.moveTo(next)
}

func (m *Model) moveTo(next Point) {
	if pack, exists := m.Packs[next]; exists {
		m.startCombat(next, pack)
		return
	}

	if next == m.PartyPos {
		if m.Training && !m.ManualMode {
			m.enterTown(RetreatTrainingDone)
			return
		}
		m.saveCurrentFloorState()
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(T(m.Lang, "dungeon.log.floor_cleared_boss", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(T(m.Lang, "dungeon.log.floor_cleared", m.Floor)))
		}

		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		return
	}

	switch m.Grid[next.Y][next.X] {
	case TileChest:
		if len(m.Bag) >= m.currentBagCapacity() {
			m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full_skip")))
			m.PartyPos = next
			m.revealFog()
			return
		}
		m.Stats.ChestsOpened++
		m.checkQuestProgress(QuestOpenChests, "", 1)
		gold := int(float64(rng.Intn(16)+10+(m.Floor*2)) * m.goldMult())
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		// Квест «Золотая лихорадка»: учёт золота из сундука
		m.checkQuestProgress(QuestCollectGold, "", gold)

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			m.checkAndAwardTitle(lh)
		}
		item := generateItemForClass(targetClass, m.Floor)
		m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.chest_open", gold, item.DisplayName(m.Lang))))
		m.equipOrBag(item)
		m.Grid[next.Y][next.X] = TileFloor

	case TileRelic:
		m.handleRelicTile()
		m.Grid[next.Y][next.X] = TileFloor

	case TileAltar:
		m.handleAltar()
		m.Grid[next.Y][next.X] = TileFloor

	case TileFountain:
		m.handleFountain()
		m.Grid[next.Y][next.X] = TileFloor

	case TileTrappedChest:
		if len(m.Bag) >= m.currentBagCapacity() {
			m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full_skip")))
			m.PartyPos = next
			m.revealFog()
			return
		}
		m.handleTrappedChest()
		m.Grid[next.Y][next.X] = TileFloor

	case TileBarrel:
		// Квест «Огненный шквал»: разрушение/подрыв бочки при наступлении
		m.Stats.BarrelsBlown++
		m.checkQuestProgress(QuestPyromania, "", 1)
		m.Grid[next.Y][next.X] = TileFloor

	case TileEvent:
		m.Grid[next.Y][next.X] = TileFloor
		m.handleEventTile(next)

	case TileStairs:
		if m.Training && !m.ManualMode {
			break
		}

		// Квест «Полная зачистка»: если на этаже не осталось живых стай монстров
		if len(m.Packs) == 0 {
			m.checkQuestProgress(QuestPurgeBiome, "", 1)
		}

		// Квест «Топография Бездны»: если разведано >= 90% территории
		if m.calcExploredPercent() >= 90 {
			m.checkQuestProgress(QuestFullMap, "", 1)
		}

		m.saveCurrentFloorState()
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(T(m.Lang, "dungeon.log.floor_cleared_boss", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(T(m.Lang, "dungeon.log.stairs_descend", m.Floor)))
		}

		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		return
	}

	m.PartyPos = next
	m.revealFog()
}

func (m *Model) calcExploredPercent() int {
	if len(m.Grid) == 0 || len(m.Grid[0]) == 0 {
		return 0
	}
	total := 0
	explored := 0
	for y := 0; y < len(m.Grid); y++ {
		for x := 0; x < len(m.Grid[y]); x++ {
			if m.Grid[y][x] != TileWall {
				total++
				if len(m.Explored) > y && len(m.Explored[y]) > x && m.Explored[y][x] {
					explored++
				}
			}
		}
	}
	if total == 0 {
		return 0
	}
	return (explored * 100) / total
}