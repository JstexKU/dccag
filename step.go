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
		m.stepTown()
		return
	}

	m.Stats.TotalSteps++

	// 1. Продолжение привала (RestTurnsLeft > 0) — сидим на месте
	if m.RestTurnsLeft > 0 {
		m.RestTurnsLeft--
		for _, h := range m.Party {
			if h.IsDead {
				continue
			}
			if h.HP < h.MaxHP {
				h.HP++
			}
			if h.MP < h.MaxMP {
				h.MP++
			}
			if h.Stress > 0 {
				h.Stress--
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
	if m.evaluateRetreat() != RetreatNone &&
		m.Grid[m.PartyPos.Y][m.PartyPos.X] == TileExit &&
		m.Stats.TotalSteps > 0 {
		m.saveCurrentFloorState()
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = T(m.Lang, "town.log.enter_gate")
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		return
	}

	// 4. Возможный старт нового привала
	urgency := m.evaluateHealingUrgency()
	restInterval := 50
	if urgency == HealingCritical {
		restInterval = 25
	}
	if m.Stats.TotalSteps%restInterval == 0 &&
		m.evaluateRetreat() == RetreatNone &&
		!m.isMonsterNearby() {

		m.RestTurnsLeft = rng.Intn(6) + 5 // 5..10 шагов
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
		if m.evaluateRetreat() != RetreatNone {
			targetTile = TileExit
		}
		forcedStep, found := m.findEmergencyStep(targetTile)
		if found {
			next = forcedStep
		} else {
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
		m.Grid[next.Y][next.X] = TileFloor

	case TileEvent:
		m.Grid[next.Y][next.X] = TileFloor
		m.handleEventTile(next)

	case TileStairs:
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
