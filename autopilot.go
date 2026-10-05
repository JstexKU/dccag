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
			// Проверяем условие только если клетка уже открыта на карте
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
			// Считаем опасными только видимых монстров
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

func (m *Model) findNextStep() Point {
	retreatReason := m.evaluateRetreat()
	retreat := retreatReason != RetreatNone
	urgency := m.evaluateHealingUrgency()
	seekingFountain := m.shouldSeekFountain(urgency)
	forceDeeper := (m.CurrentQuest.Type == QuestEscapeTrap) && !m.CurrentQuest.Completed
	bagFull := len(m.Bag) >= m.currentBagCapacity()
	needsGear := m.hasUndergearedHeroes()

	totalCells := m.MapWidth * m.MapHeight

	findPath := func(avoidMonsters bool, targetCondition func(Point, Tile) bool) (Point, bool) {
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

			if curr != m.PartyPos && targetCondition(curr, m.Grid[curr.Y][curr.X]) {
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
						// Ловушки обходятся, только если разведаны
						if m.Tactics.SkipTraps && m.Explored[ny][nx] && m.Grid[ny][nx] == TileTrappedChest && m.CurrentQuest.Type != QuestOpenChests {
							continue
						}
						// Монстров обходим, только если они разведаны
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

	// 1. Отступление к выходу (выход ищется только если он разведан на карте)
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

	// 2. Поиск источника исцеления (только среди открытых фонтанов)
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

	isExplorationQuest := m.CurrentQuest.Type == QuestReachFloor && !m.CurrentQuest.Completed

	// 4. Поиск открытых целей на карте (ЧЕСТНАЯ ПРОВЕРКА m.Explored)
	isKnownTarget := func(p Point, t Tile) bool {
		// Автопилот не видит сквозь туман войны
		if !m.Explored[p.Y][p.X] {
			return false
		}

		_, hasPack := m.Packs[p]
		if hasPack {
			if needsGear {
				return false
			}
			return true
		}
		if seekingFountain && t == TileAltar {
			return false
		}
		if t == TileStairs && isExplorationQuest {
			return true
		}

		if t == TileChest || t == TileTrappedChest {
			if bagFull {
				return false
			}
		}

		if t == TileStairs && !forceDeeper {
			hasVisibleLoot := false
			for y := 0; y < m.MapHeight; y++ {
				for x := 0; x < m.MapWidth; x++ {
					if m.Explored[y][x] {
						tile := m.Grid[y][x]
						if (tile == TileChest || tile == TileRelic) && !bagFull {
							hasVisibleLoot = true
							break
						}
					}
				}
			}
			if hasVisibleLoot {
				return false
			}
		}
		if t == TileTrappedChest && m.Tactics.SkipTraps && m.CurrentQuest.Type != QuestOpenChests {
			return false
		}
		if t == TileAltar && m.Tactics.SkipAltars && (m.CurrentQuest.Type != QuestUseAltar || m.CurrentQuest.Completed) {
			return false
		}
		return t == TileChest || t == TileStairs || t == TileAltar || t == TileFountain || t == TileTrappedChest || t == TileBarrel || t == TileRelic || t == TileEvent
	}

	if needsGear {
		if step, found := findPath(true, isKnownTarget); found {
			return step
		}
	}
	if step, found := findPath(false, isKnownTarget); found {
		return step
	}

	// 5. РЕЖИМ ИССЛЕДОВАТЕЛЯ (Frontier Exploration):
	// Если известных целей нет — идём к ближайшей открытой клетке, граничащей с неразведанной тьмой
	isFrontier := func(p Point, t Tile) bool {
		if !m.Explored[p.Y][p.X] || t == TileWall {
			return false
		}
		dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		for _, d := range dirs {
			nx, ny := p.X+d.X, p.Y+d.Y
			if nx >= 0 && nx < m.MapWidth && ny >= 0 && ny < m.MapHeight {
				if !m.Explored[ny][nx] && m.Grid[ny][nx] != TileWall {
					return true // Клетка граничит с неизведанным проходом
				}
			}
		}
		return false
	}

	if step, found := findPath(true, isFrontier); found {
		return step
	}
	if step, found := findPath(false, isFrontier); found {
		return step
	}

	return m.PartyPos
}
