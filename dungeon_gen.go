package main

func (m *Model) calculateDungeonSize(floor int) (int, int) {
	w := 70 + (floor-1)*3
	h := 30 + (floor-1)*1
	if w > 130 {
		w = 130
	}
	if h > 55 {
		h = 55
	}
	return w, h
}

func (m *Model) initDungeonForFloor(floor int) {
	w, h := m.calculateDungeonSize(floor)
	m.MapWidth = w
	m.MapHeight = h
	m.Grid = make([][]Tile, h)
	m.Explored = make([][]bool, h)
	for y := 0; y < h; y++ {
		m.Grid[y] = make([]Tile, w)
		m.Explored[y] = make([]bool, w)
	}
	m.generateDungeon()
	m.revealFog()
}

func (m *Model) generateDungeon() {
	m.Packs = make(map[Point]*MonsterPack)
	m.Combat = nil
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			m.Grid[y][x] = TileWall
			m.Explored[y][x] = false
		}
	}

	type Rect struct{ X, Y, W, H int }
	rooms := []Rect{}
	attempts := (m.MapWidth * m.MapHeight) / 50
	if attempts < 8 {
		attempts = 8
	}

	for i := 0; i < attempts; i++ {
		w := rng.Intn(7) + 7
		h := rng.Intn(4) + 4
		if m.MapWidth-w-2 <= 1 || m.MapHeight-h-2 <= 1 {
			continue
		}
		x := rng.Intn(m.MapWidth-w-2) + 1
		y := rng.Intn(m.MapHeight-h-2) + 1
		rooms = append(rooms, Rect{x, y, w, h})
		for ry := y; ry < y+h; ry++ {
			for rx := x; rx < x+w; rx++ {
				m.Grid[ry][rx] = TileFloor
			}
		}
	}

	if len(rooms) == 0 {
		return
	}

	carveWideCorridor := func(x1, y1, x2, y2 int) {
		cx, cy := x1, y1
		for cx != x2 {
			m.Grid[cy][cx] = TileFloor
			if cy+1 < m.MapHeight {
				m.Grid[cy+1][cx] = TileFloor
			}
			if x2 > cx {
				cx++
			} else {
				cx--
			}
		}
		for cy != y2 {
			m.Grid[cy][cx] = TileFloor
			if cx+1 < m.MapWidth {
				m.Grid[cy+1][cx] = TileFloor
			}
			if y2 > cy {
				cy++
			} else {
				cy--
			}
		}
	}

	for i := 0; i < len(rooms)-1; i++ {
		x1, y1 := rooms[i].X+rooms[i].W/2, rooms[i].Y+rooms[i].H/2
		x2, y2 := rooms[i+1].X+rooms[i+1].W/2, rooms[i+1].Y+rooms[i+1].H/2
		carveWideCorridor(x1, y1, x2, y2)
	}

	exitPos := Point{rooms[0].X + 1, rooms[0].Y + 1}
	if m.CurrentQuest.Type == QuestEscapeTrap && !m.CurrentQuest.Completed {
		m.Grid[exitPos.Y][exitPos.X] = TileFloor
	} else {
		m.Grid[exitPos.Y][exitPos.X] = TileExit
	}
	if exitPos.X+1 < rooms[0].X+rooms[0].W-1 {
		m.PartyPos = Point{exitPos.X + 1, exitPos.Y}
	} else {
		m.PartyPos = Point{exitPos.X, exitPos.Y + 1}
	}
	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileFloor

	endRoom := rooms[len(rooms)-1]
	m.Grid[endRoom.Y+endRoom.H/2][endRoom.X+endRoom.W/2] = TileStairs

	relicSpawned := false

	for i := 1; i < len(rooms); i++ {
		r := rooms[i]
		isBoss := (m.Floor%10 == 0 && i == len(rooms)-1)
		pos := Point{r.X + r.W/2, r.Y + r.H/2}
		if isBoss {
			m.Packs[pos] = spawnMonsterPack(true, m.Floor)
		} else {
			m.Packs[pos] = spawnMonsterPack(false, m.Floor)
			eventRoll := rng.Intn(13)
			cornerPos := Point{r.X + 1, r.Y + 1}
			switch {
			case (!relicSpawned && m.CurrentQuest.Type == QuestFindRelic) || (eventRoll == 12 && !relicSpawned):
				m.Grid[cornerPos.Y][cornerPos.X] = TileRelic
				relicSpawned = true
			case eventRoll == 0:
				m.Grid[cornerPos.Y][cornerPos.X] = TileAltar
			case eventRoll == 1:
				m.Grid[cornerPos.Y][cornerPos.X] = TileFountain
			case eventRoll == 2:
				m.Grid[cornerPos.Y][cornerPos.X] = TileTrappedChest
			case eventRoll == 3:
				m.Grid[cornerPos.Y][cornerPos.X] = TileBarrel
			case eventRoll == 4 || eventRoll == 5:
				m.Grid[cornerPos.Y][cornerPos.X] = TileEvent
			case eventRoll >= 7:
				m.Grid[cornerPos.Y][cornerPos.X] = TileChest
			}
		}
	}
}

func (m *Model) revealFog() {
	radius := 6
	for y := m.PartyPos.Y - radius; y <= m.PartyPos.Y+radius; y++ {
		for x := m.PartyPos.X - radius; x <= m.PartyPos.X+radius; x++ {
			if x >= 0 && x < m.MapWidth && y >= 0 && y < m.MapHeight {
				dx := x - m.PartyPos.X
				dy := y - m.PartyPos.Y
				if dx*dx+dy*dy <= radius*radius {
					m.Explored[y][x] = true
				}
			}
		}
	}
}
