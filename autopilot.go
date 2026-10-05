package main

func (m *Model) evaluateRetreat() RetreatReason {
	if len(m.Bag) >= m.currentBagCapacity() {
		return RetreatBagFull
	}
	if m.CurrentQuest.Completed {
		return RetreatQuestDone
	}

	living := 0
	curHP, maxHP := 0, 0
	totalPotions := 0
	totalStress := 0
	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		living++
		curHP += h.HP
		maxHP += h.MaxHP
		totalPotions += len(h.Potions)
		totalStress += h.Stress
	}

	if living <= m.Tactics.RetreatMinAlive {
		return RetreatTooFewAlive
	}
	if maxHP > 0 && float64(curHP)/float64(maxHP) < float64(m.Tactics.RetreatHPPct)/100 {
		return RetreatLowHP
	}
	if living > 0 {
		avgStress := totalStress / living
		if totalPotions == 0 && avgStress >= 100 {
			return RetreatNoResources
		}
	}

	return RetreatNone
}

func (m *Model) evaluateHealingUrgency() HealingUrgency {
	living := 0
	criticalCount := 0
	lowCount := 0
	needsMP := 0
	totalStress := 0

	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		living++
		hpPct := float64(h.HP) / float64(h.MaxHP)
		mpPct := 0.0
		if h.MaxMP > 0 {
			mpPct = float64(h.MP) / float64(h.MaxHP)
		}

		if hpPct <= 0.25 || h.Stress >= 160 {
			criticalCount++
		} else if hpPct <= 0.50 || h.Stress >= 100 {
			lowCount++
		}

		if mpPct <= 0.30 {
			switch h.Class {
			case ClassMage, ClassCleric, ClassWarlock, ClassBard, ClassPaladin:
				needsMP++
			}
		}
		totalStress += h.Stress
	}
	if living == 0 {
		return HealingNone
	}
	avgStress := totalStress / living

	switch {
	case criticalCount >= 2 || avgStress >= 130:
		return HealingCritical
	case criticalCount >= 1 || lowCount >= 2 || avgStress >= 80:
		return HealingUrgent
	case lowCount >= 1 || needsMP >= 2 || avgStress >= 50:
		return HealingOptional
	default:
		return HealingNone
	}
}

// distanceToNearest оценивает расстояние только до известных (разведанных) клеток
func (m *Model) distanceToNearest(cond func(Point, Tile) bool) (int, bool) {
	if cond(m.PartyPos, m.Grid[m.PartyPos.Y][m.PartyPos.X]) {
		return 0, true
	}

	totalCells := m.MapWidth * m.MapHeight
	visited := make([]bool, totalCells)
	startIdx := m.PartyPos.Y*m.MapWidth + m.PartyPos.X
	visited[startIdx] = true

	type entry struct {
		p    Point
		dist int
	}
	queue := []entry{{m.PartyPos, 0}}
	dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx := curr.p.X + d.X
			ny := curr.p.Y + d.Y
			if nx < 0 || nx >= m.MapWidth || ny < 0 || ny >= m.MapHeight {
				continue
			}
			idx := ny*m.MapWidth + nx
			if visited[idx] {
				continue
			}
			if m.Grid[ny][nx] == TileWall {
				continue
			}
			if m.Explored[ny][nx] && cond(Point{nx, ny}, m.Grid[ny][nx]) {
				return curr.dist + 1, true
			}
			visited[idx] = true
			queue = append(queue, entry{Point{nx, ny}, curr.dist + 1})
		}
	}
	return 0, false
}

func (m *Model) shouldSeekFountain(urgency HealingUrgency) bool {
	if urgency == HealingNone {
		return false
	}
	dist, found := m.distanceToNearest(func(p Point, t Tile) bool {
		return t == TileFountain
	})
	if !found {
		return false
	}
	switch urgency {
	case HealingCritical:
		return true
	case HealingUrgent:
		return dist <= 25
	case HealingOptional:
		return dist <= 10
	}
	return false
}

func (m *Model) isMonsterNearby() bool {
	for y := m.PartyPos.Y - 3; y <= m.PartyPos.Y+3; y++ {
		for x := m.PartyPos.X - 3; x <= m.PartyPos.X+3; x++ {
			if x < 0 || x >= m.MapWidth || y < 0 || y >= m.MapHeight {
				continue
			}
			if m.Explored[y][x] {
				if _, hasPack := m.Packs[Point{x, y}]; hasPack {
					return true
				}
			}
		}
	}
	return false
}

func (m *Model) hasUndergearedHeroes() bool {
	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		if h.Weapon == nil || h.Head == nil || h.Chest == nil || h.Legs == nil {
			return true
		}
	}
	return false
}

func (m *Model) findEmergencyStep(targetTile Tile) (Point, bool) {
	totalCells := m.MapWidth * m.MapHeight
	visited := make([]bool, totalCells)
	cameFrom := make([]int, totalCells)
	for i := range cameFrom {
		cameFrom[i] = -1
	}

	startIdx := m.PartyPos.Y*m.MapWidth + m.PartyPos.X
	visited[startIdx] = true
	queue := []Point{m.PartyPos}
	dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr != m.PartyPos && m.Grid[curr.Y][curr.X] == targetTile {
			currIdx := curr.Y*m.MapWidth + curr.X
			for cameFrom[currIdx] != startIdx && cameFrom[currIdx] != -1 {
				currIdx = cameFrom[currIdx]
			}
			return Point{X: currIdx % m.MapWidth, Y: currIdx / m.MapWidth}, true
		}

		for _, d := range dirs {
			nx := curr.X + d.X
			ny := curr.Y + d.Y
			if nx >= 0 && nx < m.MapWidth && ny >= 0 && ny < m.MapHeight {
				nIdx := ny*m.MapWidth + nx
				if !visited[nIdx] && m.Grid[ny][nx] != TileWall {
					visited[nIdx] = true
					cameFrom[nIdx] = curr.Y*m.MapWidth + curr.X
					queue = append(queue, Point{nx, ny})
				}
			}
		}
	}
	return m.PartyPos, false
}

// explorationRatio возвращает процент исследованных проходимых клеток (0.0..1.0)
func (m *Model) explorationRatio() float64 {
	totalWalkable := 0
	exploredWalkable := 0
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			if m.Grid[y][x] != TileWall {
				totalWalkable++
				if m.Explored[y][x] {
					exploredWalkable++
				}
			}
		}
	}
	if totalWalkable == 0 {
		return 1.0
	}
	return float64(exploredWalkable) / float64(totalWalkable)
}

func (m *Model) findNextStep() Point {
	retreatReason := m.evaluateRetreat()
	retreat := retreatReason != RetreatNone
	urgency := m.evaluateHealingUrgency()
	seekingFountain := m.shouldSeekFountain(urgency)
	forceDeeper := (m.CurrentQuest.Type == QuestEscapeTrap) && !m.CurrentQuest.Completed
	bagFull := len(m.Bag) >= m.currentBagCapacity()
	needsGear := m.hasUndergearedHeroes()

	totalCells := m.MapWidth * m.MapHeight

	// Вычисляем вектор последнего шага для сохранения инерции (чтобы не вихлять по коридорам)
	lastDir := Point{1, 0}
	if len(m.PathHistory) >= 1 {
		prev := m.PathHistory[len(m.PathHistory)-1]
		dx := m.PartyPos.X - prev.X
		dy := m.PartyPos.Y - prev.Y
		if dx != 0 || dy != 0 {
			lastDir = Point{dx, dy}
		}
	}

	baseDirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	var searchDirs []Point
	searchDirs = append(searchDirs, lastDir)
	for _, d := range baseDirs {
		if d != lastDir && (d.X != -lastDir.X || d.Y != -lastDir.Y) {
			searchDirs = append(searchDirs, d)
		}
	}
	searchDirs = append(searchDirs, Point{-lastDir.X, -lastDir.Y})

	findPath := func(avoidMonsters bool, targetCondition func(Point, Tile) bool) (Point, bool) {
		visited := make([]bool, totalCells)
		cameFrom := make([]int, totalCells)
		for i := range cameFrom {
			cameFrom[i] = -1
		}

		startIdx := m.PartyPos.Y*m.MapWidth + m.PartyPos.X
		visited[startIdx] = true
		queue := []Point{m.PartyPos}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr != m.PartyPos && targetCondition(curr, m.Grid[curr.Y][curr.X]) {
				currIdx := curr.Y*m.MapWidth + curr.X
				for cameFrom[currIdx] != startIdx && cameFrom[currIdx] != -1 {
					currIdx = cameFrom[currIdx]
				}
				return Point{X: currIdx % m.MapWidth, Y: currIdx / m.MapWidth}, true
			}

			for _, d := range searchDirs {
				nx := curr.X + d.X
				ny := curr.Y + d.Y
				if nx >= 0 && nx < m.MapWidth && ny >= 0 && ny < m.MapHeight {
					nIdx := ny*m.MapWidth + nx
					if !visited[nIdx] && m.Grid[ny][nx] != TileWall {
						if m.Tactics.SkipTraps && m.Explored[ny][nx] && m.Grid[ny][nx] == TileTrappedChest && m.CurrentQuest.Type != QuestOpenChests {
							continue
						}
						if avoidMonsters && m.Explored[ny][nx] {
							if _, hasMob := m.Packs[Point{nx, ny}]; hasMob {
								continue
							}
						}
						visited[nIdx] = true
						cameFrom[nIdx] = curr.Y*m.MapWidth + curr.X
						queue = append(queue, Point{nx, ny})
					}
				}
			}
		}
		return m.PartyPos, false
	}

	// 1. Отступление к выходу при критическом уроне/стрессе/полной сумке
	if retreat && !forceDeeper {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return m.Explored[p.Y][p.X] && t == TileExit
		})
		if found {
			return step
		}
		step, found = findPath(false, func(p Point, t Tile) bool {
			return m.Explored[p.Y][p.X] && t == TileExit
		})
		if found {
			return step
		}
	}

	// 2. Срочный поиск источника исцеления
	if seekingFountain {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return m.Explored[p.Y][p.X] && t == TileFountain
		})
		if found {
			return step
		}
	}

	// 3. Поиск открытого снаряжения при его дефиците
	if needsGear && !bagFull {
		step, found := findPath(true, func(p Point, t Tile) bool {
			if !m.Explored[p.Y][p.X] {
				return false
			}
			return t == TileChest || t == TileTrappedChest || t == TileRelic || t == TileEvent
		})
		if found {
			return step
		}
	}

	// 4. Поиск открытых интерактивных объектов (сундуки, алтари, бочки, монстры)
	// ВАЖНО: Лестницы ЗДЕСЬ НЕТ. Отряд не пойдет на лестницу, пока исследует этаж.
	isLootOrTarget := func(p Point, t Tile) bool {
		if !m.Explored[p.Y][p.X] {
			return false
		}

		// Монстров бьем, если открыты
		if _, hasPack := m.Packs[p]; hasPack {
			return !needsGear
		}

		if seekingFountain && t == TileAltar {
			return false
		}

		if (t == TileChest || t == TileTrappedChest) && bagFull {
			return false
		}

		if t == TileTrappedChest && m.Tactics.SkipTraps && m.CurrentQuest.Type != QuestOpenChests {
			return false
		}
		if t == TileAltar && m.Tactics.SkipAltars && (m.CurrentQuest.Type != QuestUseAltar || m.CurrentQuest.Completed) {
			return false
		}

		return t == TileChest || t == TileAltar || t == TileFountain || t == TileTrappedChest || t == TileBarrel || t == TileRelic || t == TileEvent
	}

	if needsGear {
		if step, found := findPath(true, isLootOrTarget); found {
			return step
		}
	}
	if step, found := findPath(false, isLootOrTarget); found {
		return step
	}

	// 5. РЕЖИМ ИССЛЕДОВАТЕЛЯ (Frontier Exploration):
	// Идем открывать неизведанную тьму, пока не изучим хотя бы 70% карты
	exploredPct := m.explorationRatio()
	needMoreExploration := exploredPct < 0.70 && !forceDeeper && !bagFull

	isFrontier := func(p Point, t Tile) bool {
		if !m.Explored[p.Y][p.X] || t == TileWall {
			return false
		}
		dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		for _, d := range dirs {
			nx, ny := p.X + d.X, p.Y + d.Y
			if nx >= 0 && nx < m.MapWidth && ny >= 0 && ny < m.MapHeight {
				if !m.Explored[ny][nx] && m.Grid[ny][nx] != TileWall {
					return true
				}
			}
		}
		return false
	}

	// Если этаж ещё мало исследован (<70%), сначала в приоритете поиск тумана войны!
	if needMoreExploration {
		if step, found := findPath(true, isFrontier); found {
			return step
		}
		if step, found := findPath(false, isFrontier); found {
			return step
		}
	}

	// 6. СПУСК ПО ЛЕСТНИЦЕ:
	// Сюда мы попадаем, если:
	// - Исследовали >= 70% карты
	// - Либо нет доступных границ тумана войны (все залы обойдены)
	// - Либо переполнен мешок / включен режим побега из ловушки
	step, found := findPath(false, func(p Point, t Tile) bool {
		return m.Explored[p.Y][p.X] && t == TileStairs
	})
	if found {
		return step
	}

	// 7. Если лестница ещё не найдена в разведанной зоне, продолжаем открывать остатки карты
	if step, found := findPath(true, isFrontier); found {
		return step
	}
	if step, found := findPath(false, isFrontier); found {
		return step
	}

	return m.PartyPos
}