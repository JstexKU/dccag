package main

import "sort"

// ============================================================
// БЛУЖДАЮЩИЕ СТАИ
// ============================================================
//
// Стая не стоит на месте: пока отряд далеко, она бродит возле точки спавна внутри своей комнаты.
// Если отряд попадает в её комнату и оказывается в радиусе обзора, стая бросается в погоню и
// преследует его, не покидая комнату. Потеряв отряд из виду (он вышел из комнаты или отбежал
// слишком далеко), стая возвращается к своему месту. Настигнув отряд, она сама начинает бой.
// Боссы стоят на месте. Стаи, далёкие от отряда, «спят» и не двигаются.

const (
	packSightRadius       = 4  // радиус обзора стаи (в клетках, по большей из осей)
	packLoseRadius        = 6  // дальше этого погоня теряет отряд из виду
	packRoamRadius        = 3  // как далеко от точки спавна стая бродит в покое
	packRoamChance        = 35 // шанс спокойной стаи сделать шаг за ход отряда, %
	packActiveRadius      = 14 // стаи дальше этого от отряда не двигаются
	packDefaultAreaRadius = 6  // область стаи без комнаты (например, на тренировочном этаже)
)

// RoomArea — прямоугольник комнаты, которой принадлежит стая. Нулевая область означает
// «окрестность точки спавна радиусом packDefaultAreaRadius».
type RoomArea struct {
	X, Y, W, H int
}

func chebyshev(a, b Point) int {
	return max(absInt(a.X-b.X), absInt(a.Y-b.Y))
}

func manhattan(a, b Point) int {
	return absInt(a.X-b.X) + absInt(a.Y-b.Y)
}

// inArea — лежит ли точка в области стаи (комната или окрестность спавна).
func (p *MonsterPack) inArea(pt Point) bool {
	if p.Area.W > 0 && p.Area.H > 0 {
		return pt.X >= p.Area.X && pt.X < p.Area.X+p.Area.W && pt.Y >= p.Area.Y && pt.Y < p.Area.Y+p.Area.H
	}
	return chebyshev(pt, p.Home) <= packDefaultAreaRadius
}

// addPack ставит стаю на карту и запоминает её место и комнату для блуждания.
func (m *Model) addPack(pos Point, pack *MonsterPack, area RoomArea) {
	pack.Home = pos
	pack.Area = area
	pack.Chasing = false
	m.Packs[pos] = pack
}

// packCanStand — может ли стая встать в клетку c: пол, свободный от других стай и от отряда.
func (m *Model) packCanStand(c Point, self Point) bool {
	if c.X < 0 || c.X >= m.MapWidth || c.Y < 0 || c.Y >= m.MapHeight {
		return false
	}
	if m.Grid[c.Y][c.X] != TileFloor || c == m.PartyPos {
		return false
	}
	if _, taken := m.Packs[c]; taken && c != self {
		return false
	}
	return true
}

var packDirs = []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

// packFirstStep — первый шаг кратчайшего пути стаи из from к ближайшей клетке, удовлетворяющей
// isGoal; путь идёт только по клеткам её области. Поиск в ширину в порядке packDirs, поэтому
// результат детерминирован.
func (m *Model) packFirstStep(from Point, pack *MonsterPack, isGoal func(Point) bool) (Point, bool) {
	parent := map[Point]Point{from: from}
	queue := []Point{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur != from && isGoal(cur) {
			for parent[cur] != from {
				cur = parent[cur]
			}
			return cur, true
		}
		for _, d := range packDirs {
			next := Point{cur.X + d.X, cur.Y + d.Y}
			if _, seen := parent[next]; seen {
				continue
			}
			if !pack.inArea(next) || !m.packCanStand(next, from) {
				continue
			}
			parent[next] = cur
			queue = append(queue, next)
		}
	}
	return from, false
}

// movePack переносит стаю в соседнюю клетку.
func (m *Model) movePack(from, to Point, pack *MonsterPack) {
	delete(m.Packs, from)
	m.Packs[to] = pack
}

// stepPacks — ход всех стай этажа после шага отряда. Порядок обработки фиксирован (по строкам
// карты), иначе случайный обход map нарушил бы воспроизводимость по сиду.
func (m *Model) stepPacks() {
	if m.InTown || m.Combat != nil || m.State != StatePlaying || len(m.Packs) == 0 {
		return
	}
	type entry struct {
		pos  Point
		pack *MonsterPack
	}
	list := make([]entry, 0, len(m.Packs))
	for pos, pack := range m.Packs {
		list = append(list, entry{pos, pack})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].pos.Y != list[j].pos.Y {
			return list[i].pos.Y < list[j].pos.Y
		}
		return list[i].pos.X < list[j].pos.X
	})

	for _, e := range list {
		if m.Combat != nil {
			return
		}
		if m.Packs[e.pos] != e.pack {
			continue // стая уже ушла со своего места или бой её убрал
		}
		m.stepPack(e.pos, e.pack)
	}
}

// stepPack — ход одной стаи: погоня, возвращение или спокойное блуждание.
func (m *Model) stepPack(pos Point, pack *MonsterPack) {
	if pack.IsBoss || pack.LivingCount() == 0 {
		return
	}

	dist := chebyshev(pos, m.PartyPos)
	if dist > packActiveRadius {
		pack.Chasing = false
		return
	}

	partyInArea := pack.inArea(m.PartyPos)
	if pack.Chasing && !(partyInArea && dist <= packLoseRadius) {
		pack.Chasing = false // потеряла отряд из виду
	}
	if !pack.Chasing && partyInArea && dist <= packSightRadius {
		pack.Chasing = true
		if m.Explored[pos.Y][pos.X] {
			if first := pack.GetFirstLiving(); first != nil {
				m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.pack_notice", m.monsterName(first))))
			}
		}
	}

	if pack.Chasing {
		if manhattan(pos, m.PartyPos) <= 1 {
			m.packCatchesParty(pos, pack)
			return
		}
		// цель погони — свободная клетка рядом с отрядом, а не сама клетка отряда
		next, ok := m.packFirstStep(pos, pack, func(c Point) bool { return manhattan(c, m.PartyPos) <= 1 })
		if ok {
			m.movePack(pos, next, pack)
			if manhattan(next, m.PartyPos) <= 1 {
				m.packCatchesParty(next, pack)
			}
		}
		return
	}

	// Далеко от места — возвращаемся; иначе иногда делаем шаг по комнате.
	if chebyshev(pos, pack.Home) > packRoamRadius {
		if next, ok := m.packFirstStep(pos, pack, func(c Point) bool { return chebyshev(c, pack.Home) <= packRoamRadius }); ok {
			m.movePack(pos, next, pack)
		}
		return
	}
	if rng.Intn(100) >= packRoamChance {
		return
	}
	d := packDirs[rng.Intn(len(packDirs))]
	next := Point{pos.X + d.X, pos.Y + d.Y}
	if chebyshev(next, pack.Home) <= packRoamRadius && pack.inArea(next) &&
		m.packCanStand(next, pos) && manhattan(next, m.PartyPos) > 1 {
		m.movePack(pos, next, pack)
	}
}

// packCatchesParty — стая настигла отряд и начинает бой с клетки, на которой стоит.
func (m *Model) packCatchesParty(pos Point, pack *MonsterPack) {
	if first := pack.GetFirstLiving(); first != nil {
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.pack_caught", m.monsterName(first))))
	}
	pack.Chasing = false
	m.RestTurnsLeft = 0
	m.startCombat(pos, pack)
}
