package main

import (
	"sort"
)

func (m *Model) calculateDungeonSize(floor int) (int, int) {
	// Масштабные размеры карты: ширина увеличена в 2 раза (220..440), высота 45..90
	w := 220 + (floor-1)*12
	h := 45 + (floor-1)*3
	if w > 440 {
		w = 440
	}
	if h > 90 {
		h = 90
	}
	return w, h
}

// saveCurrentFloorState сохраняет снимок текущего этажа в память
func (m *Model) saveCurrentFloorState() {
	if m.MapWidth == 0 || m.MapHeight == 0 || len(m.Grid) == 0 {
		return
	}
	if m.VisitedFloors == nil {
		m.VisitedFloors = make(map[int]*FloorState)
	}

	gridCopy := make([][]Tile, m.MapHeight)
	exploredCopy := make([][]bool, m.MapHeight)
	for y := 0; y < m.MapHeight; y++ {
		gridCopy[y] = make([]Tile, m.MapWidth)
		exploredCopy[y] = make([]bool, m.MapWidth)
		copy(gridCopy[y], m.Grid[y])
		copy(exploredCopy[y], m.Explored[y])
	}

	packsCopy := make(map[Point]*MonsterPack, len(m.Packs))
	for pt, pack := range m.Packs {
		if pack != nil && pack.LivingCount() > 0 {
			packsCopy[pt] = pack
		}
	}

	var exitPos, stairsPos Point
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			if m.Grid[y][x] == TileExit {
				exitPos = Point{x, y}
			} else if m.Grid[y][x] == TileStairs {
				stairsPos = Point{x, y}
			}
		}
	}

	m.VisitedFloors[m.Floor] = &FloorState{
		Floor:     m.Floor,
		Width:     m.MapWidth,
		Height:    m.MapHeight,
		Grid:      gridCopy,
		Explored:  exploredCopy,
		Packs:     packsCopy,
		ExitPos:   exitPos,
		StairsPos: stairsPos,
	}
}

// restoreFloorState восстанавливает карту и состояние сохранённого этажа
func (m *Model) restoreFloorState(st *FloorState) {
	m.MapWidth = st.Width
	m.MapHeight = st.Height

	m.Grid = make([][]Tile, st.Height)
	m.Explored = make([][]bool, st.Height)
	for y := 0; y < st.Height; y++ {
		m.Grid[y] = make([]Tile, st.Width)
		m.Explored[y] = make([]bool, st.Width)
		copy(m.Grid[y], st.Grid[y])
		copy(m.Explored[y], st.Explored[y])
	}

	m.Packs = make(map[Point]*MonsterPack, len(st.Packs))
	for pt, pack := range st.Packs {
		if pack != nil && pack.LivingCount() > 0 {
			m.Packs[pt] = pack
		}
	}
	m.Combat = nil
}

func (m *Model) initDungeonForFloor(floor int) {
	if m.VisitedFloors == nil {
		m.VisitedFloors = make(map[int]*FloorState)
	}

	// 1. Если этаж уже посещался — загружаем его сохранённое состояние
	if saved, exists := m.VisitedFloors[floor]; exists && saved != nil {
		m.restoreFloorState(saved)
		// Если отряд возвращается на этаж — безопасно ставим его на свободный пол около выхода
		if saved.ExitPos.X > 0 && saved.ExitPos.Y > 0 {
			placed := false
			dirs := []Point{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
			for _, d := range dirs {
				np := Point{saved.ExitPos.X + d.X, saved.ExitPos.Y + d.Y}
				if np.X >= 0 && np.X < m.MapWidth && np.Y >= 0 && np.Y < m.MapHeight && m.Grid[np.Y][np.X] == TileFloor {
					m.PartyPos = np
					placed = true
					break
				}
			}
			if !placed {
				m.PartyPos = saved.ExitPos
			}
		}
		m.revealFog()
		return
	}

	// 2. Иначе генерируем новый масштабный этаж
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

	// Сохраняем свежесгенерированный этаж в кэш
	m.saveCurrentFloorState()
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

	type Rect struct {
		X, Y, W, H int
		Shape      int // 0 - прямоугольная, 1 - овальный грот, 2 - зал с колоннами
	}
	rooms := []Rect{}

	// Плотность комнат под увеличенный масштаб карты
	attempts := (m.MapWidth * m.MapHeight) / 36
	if attempts < 28 {
		attempts = 28
	}

	overlaps := func(r1, r2 Rect) bool {
		pad := 2
		return !(r1.X+r1.W+pad < r2.X || r2.X+r2.W+pad < r1.X ||
			r1.Y+r1.H+pad < r2.Y || r2.Y+r2.H+pad < r1.Y)
	}

	for i := 0; i < attempts; i++ {
		rw := rng.Intn(9) + 7
		rh := rng.Intn(5) + 5
		if m.MapWidth-rw-3 <= 1 || m.MapHeight-rh-3 <= 1 {
			continue
		}
		rx := rng.Intn(m.MapWidth-rw-3) + 2
		ry := rng.Intn(m.MapHeight-rh-3) + 2
		cand := Rect{X: rx, Y: ry, W: rw, H: rh, Shape: rng.Intn(3)}

		collision := false
		for _, ex := range rooms {
			if overlaps(cand, ex) {
				collision = true
				break
			}
		}
		if collision {
			continue
		}

		rooms = append(rooms, cand)

		// Вырезаем комнату заданной геометрии
		switch cand.Shape {
		case 1: // Овальный грот
			cx := float64(rx) + float64(rw)/2.0
			cy := float64(ry) + float64(rh)/2.0
			rxR := float64(rw) / 2.0
			ryR := float64(rh) / 2.0
			for y := ry; y < ry+rh; y++ {
				for x := rx; x < rx+rw; x++ {
					dx := (float64(x) - cx) / rxR
					dy := (float64(y) - cy) / ryR
					if dx*dx+dy*dy <= 1.05 {
						m.Grid[y][x] = TileFloor
					}
				}
			}
		case 2: // Зал с колоннами
			for y := ry; y < ry+rh; y++ {
				for x := rx; x < rx+rw; x++ {
					if (x-rx)%3 == 2 && (y-ry)%3 == 2 && x > rx+1 && x < rx+rw-2 && y > ry+1 && y < ry+rh-2 {
						m.Grid[y][x] = TileWall
					} else {
						m.Grid[y][x] = TileFloor
					}
				}
			}
		default: // Прямоугольный зал
			for y := ry; y < ry+rh; y++ {
				for x := rx; x < rx+rw; x++ {
					m.Grid[y][x] = TileFloor
				}
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
			if cy+1 < m.MapHeight-1 {
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
			if cx+1 < m.MapWidth-1 {
				m.Grid[cy+1][cx] = TileFloor
			}
			if y2 > cy {
				cy++
			} else {
				cy--
			}
		}
	}

	// 3. Построение связного графа с петлями и развилками (MST + loopback)
	type Edge struct {
		u, v int
		dist int
	}
	var allEdges []Edge
	for i := 0; i < len(rooms); i++ {
		c1 := Point{rooms[i].X + rooms[i].W/2, rooms[i].Y + rooms[i].H/2}
		for j := i + 1; j < len(rooms); j++ {
			c2 := Point{rooms[j].X + rooms[j].W/2, rooms[j].Y + rooms[j].H/2}
			dx := c1.X - c2.X
			dy := c1.Y - c2.Y
			d := dx*dx + dy*dy
			allEdges = append(allEdges, Edge{i, j, d})
		}
	}

	sort.Slice(allEdges, func(i, j int) bool {
		return allEdges[i].dist < allEdges[j].dist
	})

	parent := make([]int, len(rooms))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] == i {
			return i
		}
		parent[i] = find(parent[i])
		return parent[i]
	}
	union := func(i, j int) bool {
		rootI, rootJ := find(i), find(j)
		if rootI != rootJ {
			parent[rootI] = rootJ
			return true
		}
		return false
	}

	var connectedEdges []Edge
	for _, e := range allEdges {
		if union(e.u, e.v) {
			connectedEdges = append(connectedEdges, e)
		}
	}

	extraLoops := len(rooms) / 4
	addedLoops := 0
	for _, e := range allEdges {
		if addedLoops >= extraLoops {
			break
		}
		alreadyConnected := false
		for _, ce := range connectedEdges {
			if (ce.u == e.u && ce.v == e.v) || (ce.u == e.v && ce.v == e.u) {
				alreadyConnected = true
				break
			}
		}
		// Порог поднят с 800 до 2500 с учётом удвоенной ширины карты
		if !alreadyConnected && e.dist < 2500 {
			connectedEdges = append(connectedEdges, e)
			addedLoops++
		}
	}

	for _, e := range connectedEdges {
		r1, r2 := rooms[e.u], rooms[e.v]
		x1, y1 := r1.X+r1.W/2, r1.Y+r1.H/2
		x2, y2 := r2.X+r2.W/2, r2.Y+r2.H/2
		carveWideCorridor(x1, y1, x2, y2)
	}

	// 4. Расстановка входа, выхода и лестницы
	exitRoom := rooms[0]
	exitPos := Point{exitRoom.X + exitRoom.W/2, exitRoom.Y + exitRoom.H/2}
	if m.CurrentQuest.Type == QuestEscapeTrap && !m.CurrentQuest.Completed {
		m.Grid[exitPos.Y][exitPos.X] = TileFloor
	} else {
		m.Grid[exitPos.Y][exitPos.X] = TileExit
	}

	m.PartyPos = Point{exitPos.X + 1, exitPos.Y}
	if m.PartyPos.X >= m.MapWidth || m.Grid[m.PartyPos.Y][m.PartyPos.X] == TileWall {
		m.PartyPos = exitPos
	}
	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileFloor

	endRoom := rooms[len(rooms)-1]
	stairsPos := Point{endRoom.X + endRoom.W/2, endRoom.Y + endRoom.H/2}
	m.Grid[stairsPos.Y][stairsPos.X] = TileStairs

	relicSpawned := false

	// 5. Плотное и многослойное наполнение комнат (монстры, тайники, бочки, алтари)
	for i := 1; i < len(rooms); i++ {
		r := rooms[i]
		isBossRoom := (m.Floor%10 == 0 && i == len(rooms)-1)
		center := Point{r.X + r.W/2, r.Y + r.H/2}

		// А. Спавн монстров: в каждой комнате есть пак, в больших — шанс на второй
		if isBossRoom {
			m.Packs[center] = spawnMonsterPack(true, m.Floor)
		} else {
			m.Packs[center] = spawnMonsterPack(false, m.Floor)

			if (r.W >= 10 || r.H >= 7) && rng.Intn(100) < 45 {
				secondPos := Point{r.X + 2, r.Y + r.H - 2}
				if m.Grid[secondPos.Y][secondPos.X] == TileFloor && secondPos != center {
					m.Packs[secondPos] = spawnMonsterPack(false, m.Floor)
				}
			}
		}

		// Б. Точки расстановки интерактивных объектов
		spawnPoints := []Point{
			{r.X + 1, r.Y + 1},
			{r.X + r.W - 2, r.Y + 1},
			{r.X + 1, r.H + r.Y - 2},
			{r.X + r.W - 2, r.H + r.Y - 2},
		}

		// Реликвия для квеста или редкая находка
		if !relicSpawned && (m.CurrentQuest.Type == QuestFindRelic || rng.Intn(14) == 0) {
			p := spawnPoints[0]
			if m.Grid[p.Y][p.X] == TileFloor {
				m.Grid[p.Y][p.X] = TileRelic
				relicSpawned = true
				spawnPoints = spawnPoints[1:]
			}
		}

		// Тематическая специализация зала
		roomTheme := rng.Intn(12)

		for idx, pt := range spawnPoints {
			if pt.X <= 0 || pt.X >= m.MapWidth-1 || pt.Y <= 0 || pt.Y >= m.MapHeight-1 {
				continue
			}
			if m.Grid[pt.Y][pt.X] != TileFloor || pt == center {
				continue
			}

			switch roomTheme {
			case 0, 1, 2, 3: // Сокровищница: связка сундуков и ловушек
				if idx == 0 {
					m.Grid[pt.Y][pt.X] = TileChest
				} else if idx == 1 {
					if rng.Intn(100) < 50 {
						m.Grid[pt.Y][pt.X] = TileTrappedChest
					} else {
						m.Grid[pt.Y][pt.X] = TileChest
					}
				} else if idx == 2 && rng.Intn(100) < 45 {
					m.Grid[pt.Y][pt.X] = TileBarrel
				}

			case 4, 5: // Святилище: Фонтан или Алтарь + бочки со смолой
				if idx == 0 {
					if rng.Intn(100) < 60 {
						m.Grid[pt.Y][pt.X] = TileFountain
					} else {
						m.Grid[pt.Y][pt.X] = TileAltar
					}
				} else if idx == 1 {
					m.Grid[pt.Y][pt.X] = TileBarrel
				} else if idx == 2 && rng.Intn(100) < 50 {
					m.Grid[pt.Y][pt.X] = TileChest
				}

			case 6, 7: // Зал событий (?) + сундук
				if idx == 0 {
					m.Grid[pt.Y][pt.X] = TileEvent
				} else if idx == 1 {
					m.Grid[pt.Y][pt.X] = TileChest
				} else if idx == 2 && rng.Intn(100) < 50 {
					m.Grid[pt.Y][pt.X] = TileBarrel
				}

			default: // Боевой зал: регулярный сундук, ловушка или бочки
				roll := rng.Intn(10)
				switch {
				case roll <= 3:
					m.Grid[pt.Y][pt.X] = TileChest
				case roll == 4:
					m.Grid[pt.Y][pt.X] = TileTrappedChest
				case roll == 5 || roll == 6:
					m.Grid[pt.Y][pt.X] = TileBarrel
				case roll == 7:
					m.Grid[pt.Y][pt.X] = TileEvent
				}
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