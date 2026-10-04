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
		if h.IsDead {
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
		if h.IsDead {
			continue
		}
		living++
		hpPct := float64(h.HP) / float64(h.MaxHP)
		mpPct := 0.0
		if h.MaxMP > 0 {
			mpPct = float64(h.MP) / float64(h.MaxMP)
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

func (m *Model) distanceToNearest(cond func(Point, Tile) bool) (int, bool) {
	if cond(m.PartyPos, m.Grid[m.PartyPos.Y][m.PartyPos.X]) {
		return 0, true
	}
	visited := make(map[Point]bool)
	visited[m.PartyPos] = true
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
			next := Point{curr.p.X + d.X, curr.p.Y + d.Y}
			if next.X < 0 || next.X >= m.MapWidth || next.Y < 0 || next.Y >= m.MapHeight {
				continue
			}
			if visited[next] {
				continue
			}
			if m.Grid[next.Y][next.X] == TileWall {
				continue
			}
			if cond(next, m.Grid[next.Y][next.X]) {
				return curr.dist + 1, true
			}
			visited[next] = true
			queue = append(queue, entry{next, curr.dist + 1})
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
		return dist <= 15
	case HealingOptional:
		return dist <= 6
	}
	return false
}

func (m *Model) isMonsterNearby() bool {
	for y := m.PartyPos.Y - 3; y <= m.PartyPos.Y+3; y++ {
		for x := m.PartyPos.X - 3; x <= m.PartyPos.X+3; x++ {
			if x < 0 || x >= m.MapWidth || y < 0 || y >= m.MapHeight {
				continue
			}
			if _, hasPack := m.Packs[Point{x, y}]; hasPack {
				return true
			}
		}
	}
	return false
}

func (m *Model) hasUndergearedHeroes() bool {
	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		if h.Weapon == nil || h.Head == nil || h.Chest == nil || h.Legs == nil {
			return true
		}
	}
	return false
}

func (m *Model) findEmergencyStep(targetTile Tile) (Point, bool) {
	queue := []Point{m.PartyPos}
	visited := make(map[Point]bool)
	cameFrom := make(map[Point]Point)
	visited[m.PartyPos] = true
	dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr != m.PartyPos && m.Grid[curr.Y][curr.X] == targetTile {
			step := curr
			for cameFrom[step] != m.PartyPos {
				step = cameFrom[step]
			}
			return step, true
		}

		for _, d := range dirs {
			np := Point{curr.X + d.X, curr.Y + d.Y}
			if np.X >= 0 && np.X < m.MapWidth && np.Y >= 0 && np.Y < m.MapHeight {
				if !visited[np] && m.Grid[np.Y][np.X] != TileWall {
					visited[np] = true
					cameFrom[np] = curr
					queue = append(queue, np)
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

	findPath := func(avoidMonsters bool, targetCondition func(Point, Tile) bool) (Point, bool) {
		queue := []Point{m.PartyPos}
		visited := make(map[Point]bool)
		cameFrom := make(map[Point]Point)
		visited[m.PartyPos] = true
		dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr != m.PartyPos && targetCondition(curr, m.Grid[curr.Y][curr.X]) {
				step := curr
				for cameFrom[step] != m.PartyPos {
					step = cameFrom[step]
				}
				return step, true
			}

			for _, d := range dirs {
				next := Point{curr.X + d.X, curr.Y + d.Y}
				if next.X >= 0 && next.X < m.MapWidth && next.Y >= 0 && next.Y < m.MapHeight {
					if !visited[next] && m.Grid[next.Y][next.X] != TileWall {
						if m.Tactics.SkipTraps && m.Grid[next.Y][next.X] == TileTrappedChest && m.CurrentQuest.Type != QuestOpenChests {
							continue
						}
						if avoidMonsters {
							if _, hasMob := m.Packs[next]; hasMob {
								continue
							}
						}
						visited[next] = true
						cameFrom[next] = curr
						queue = append(queue, next)
					}
				}
			}
		}
		return m.PartyPos, false
	}

	if retreat && !forceDeeper {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
		step, found = findPath(false, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
	}

	if seekingFountain {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileFountain
		})
		if found {
			return step
		}
	}

	if needsGear && !bagFull {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileChest || t == TileTrappedChest || t == TileRelic || t == TileEvent
		})
		if found {
			return step
		}
	}

	isExplorationQuest := m.CurrentQuest.Type == QuestReachFloor && !m.CurrentQuest.Completed

	isTarget := func(p Point, t Tile) bool {
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
		step, found := findPath(true, isTarget)
		if found {
			return step
		}
	}

	step, found := findPath(false, isTarget)
	if found {
		return step
	}

	return m.PartyPos
}
