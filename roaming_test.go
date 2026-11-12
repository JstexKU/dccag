package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// arenaModel — модель с пустой картой: всё стена, кроме прямоугольного зала; всё разведано.
func arenaModel(t *testing.T, x0, y0, w, h int) Model {
	t.Helper()
	seedRNG(91)
	m := initialModel()
	m.State = StatePlaying
	m.Lang = LangRU
	m.Packs = map[Point]*MonsterPack{}
	m.Combat = nil
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			m.Grid[y][x] = TileWall
			m.Explored[y][x] = true
		}
	}
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			m.Grid[y][x] = TileFloor
		}
	}
	return m
}

func ratPack() *MonsterPack {
	return &MonsterPack{Members: []*Monster{{
		Type: MobRat, NameKey: "mob.rat", Level: 1, Glyph: 'r', Color: "137",
		HP: 20, MaxHP: 20, Atk: 5, Defense: 0, Speed: 5, Exp: 10,
	}}}
}

func packPositions(m *Model) string {
	var pts []Point
	for p := range m.Packs {
		pts = append(pts, p)
	}
	sort.Slice(pts, func(i, j int) bool {
		if pts[i].Y != pts[j].Y {
			return pts[i].Y < pts[j].Y
		}
		return pts[i].X < pts[j].X
	})
	return fmt.Sprint(pts)
}

// Далёкие стаи спят: пока отряд за радиусом активности, они не двигаются.
func TestSleepingPacksStayPut(t *testing.T) {
	m := arenaModel(t, 5, 5, 30, 20)
	pack := ratPack()
	m.addPack(Point{10, 10}, pack, RoomArea{5, 5, 30, 20})
	m.PartyPos = Point{25, 22} // дальше packActiveRadius

	for i := 0; i < 200; i++ {
		m.stepPacks()
	}
	if m.Packs[Point{10, 10}] != pack || len(m.Packs) != 1 {
		t.Fatalf("спящая стая сдвинулась: %s", packPositions(&m))
	}
}

// Спокойная стая бродит возле точки спавна и не покидает свою область.
func TestPackRoamsNearHome(t *testing.T) {
	m := arenaModel(t, 5, 5, 30, 20)
	home := Point{10, 10}
	pack := ratPack()
	m.addPack(home, pack, RoomArea{5, 5, 30, 20})
	m.PartyPos = Point{22, 12} // отряд близко к стае, но вне её обзора

	visited := map[Point]bool{}
	for i := 0; i < 400; i++ {
		m.stepPacks()
		if len(m.Packs) != 1 {
			t.Fatalf("шаг %d: стай %d вместо одной", i, len(m.Packs))
		}
		for pos := range m.Packs {
			visited[pos] = true
			if chebyshev(pos, home) > packRoamRadius || !pack.inArea(pos) {
				t.Fatalf("шаг %d: стая ушла слишком далеко: %v", i, pos)
			}
		}
		if pack.Chasing || m.Combat != nil {
			t.Fatal("стая не должна была заметить отряд")
		}
	}
	if len(visited) < 3 {
		t.Errorf("стая почти не двигалась: посещено клеток %d", len(visited))
	}
}

// Отряд в комнате и в радиусе обзора: стая преследует, настигает и начинает бой со своей клетки.
func TestPackChasesAndCatchesParty(t *testing.T) {
	m := arenaModel(t, 5, 5, 12, 8)
	pack := ratPack()
	m.addPack(Point{8, 8}, pack, RoomArea{5, 5, 12, 8})
	m.PartyPos = Point{12, 8}

	start := Point{8, 8}
	for i := 0; i < 12 && m.Combat == nil; i++ {
		m.stepPacks()
		if i == 0 {
			if !pack.Chasing {
				t.Fatal("стая не заметила отряд в радиусе обзора")
			}
			var moved bool
			for pos := range m.Packs {
				moved = pos != start && manhattan(pos, m.PartyPos) < manhattan(start, m.PartyPos)
			}
			if !moved && m.Combat == nil {
				t.Errorf("стая не приблизилась: %s", packPositions(&m))
			}
		}
	}
	if m.Combat == nil {
		t.Fatalf("стая так и не настигла отряд: %s", packPositions(&m))
	}
	if m.Combat.Pack != pack || m.Packs[m.Combat.Pos] != pack || manhattan(m.Combat.Pos, m.PartyPos) > 1 {
		t.Errorf("бой начался не там: pos=%v, отряд %v", m.Combat.Pos, m.PartyPos)
	}
	if !strings.Contains(strings.Join(m.Logs, "\n"), "👁") {
		t.Error("в журнале нет записи о том, что стая заметила отряд")
	}
}

// Вне комнаты стаи отряд не преследуют, даже стоя в радиусе обзора.
func TestPackIgnoresPartyOutsideItsRoom(t *testing.T) {
	m := arenaModel(t, 5, 5, 40, 10)
	pack := ratPack()
	m.addPack(Point{14, 10}, pack, RoomArea{8, 8, 8, 6}) // комната: x 8..15
	m.PartyPos = Point{16, 10}                           // рядом, но за стеной комнаты

	for i := 0; i < 100; i++ {
		m.stepPacks()
		if pack.Chasing || m.Combat != nil {
			t.Fatalf("шаг %d: стая преследует отряд вне своей комнаты", i)
		}
	}
}

func TestPackLosesTrackAndReturnsHome(t *testing.T) {
	m := arenaModel(t, 5, 5, 12, 8)
	pack := ratPack()
	m.addPack(Point{8, 8}, pack, RoomArea{5, 5, 12, 8})
	pack.Chasing = true
	m.PartyPos = Point{15, 8} // дальше packLoseRadius
	m.stepPacks()
	if pack.Chasing {
		t.Error("стая должна потерять отряд, убежавший за радиус погони")
	}

	// Далеко от дома без погони: стая возвращается шаг за шагом.
	m = arenaModel(t, 5, 5, 30, 12)
	home := Point{10, 10}
	pack = ratPack()
	m.addPack(home, pack, RoomArea{5, 5, 30, 12})
	delete(m.Packs, home)
	m.Packs[Point{15, 10}] = pack
	m.PartyPos = Point{27, 10}
	m.stepPacks()
	m.stepPacks()
	for pos := range m.Packs {
		if chebyshev(pos, home) > packRoamRadius {
			t.Errorf("за два хода стая не вернулась домой: %v", pos)
		}
	}
}

func TestBossesStandStill(t *testing.T) {
	m := arenaModel(t, 5, 5, 12, 8)
	boss := ratPack()
	boss.IsBoss = true
	m.addPack(Point{8, 8}, boss, RoomArea{5, 5, 12, 8})
	m.PartyPos = Point{11, 8}

	for i := 0; i < 60; i++ {
		m.stepPacks()
	}
	if m.Packs[Point{8, 8}] != boss || m.Combat != nil || boss.Chasing {
		t.Errorf("босс не должен ни бродить, ни преследовать: %s", packPositions(&m))
	}
}

// Стаи никогда не делят клетку, не наступают на отряд и не пропадают с карты.
func TestPacksNeverOverlapOrVanish(t *testing.T) {
	m := arenaModel(t, 5, 5, 8, 5)
	area := RoomArea{5, 5, 8, 5}
	for i, pos := range []Point{{5, 5}, {6, 5}, {12, 5}, {5, 9}, {12, 9}, {9, 7}} {
		_ = i
		m.addPack(pos, ratPack(), area)
	}
	m.PartyPos = Point{8, 6}

	for i := 0; i < 300 && m.Combat == nil; i++ {
		m.stepPacks()
		if len(m.Packs) != 6 {
			t.Fatalf("шаг %d: стай %d вместо 6: %s", i, len(m.Packs), packPositions(&m))
		}
		for pos := range m.Packs {
			if pos == m.PartyPos {
				t.Fatalf("шаг %d: стая встала на клетку отряда", i)
			}
		}
	}
}

// Одно и то же зерно даёт одно и то же блуждание (обход map не влияет на результат).
func TestPackRoamingIsDeterministic(t *testing.T) {
	run := func() string {
		m := arenaModel(t, 5, 5, 40, 20)
		area := RoomArea{5, 5, 40, 20}
		for _, pos := range []Point{{10, 8}, {14, 8}, {10, 12}, {20, 15}, {30, 10}, {33, 18}} {
			m.addPack(pos, ratPack(), area)
		}
		m.PartyPos = Point{25, 22}
		seedRNG(123)
		for i := 0; i < 150; i++ {
			m.PartyPos = Point{20 + i%7, 20}
			m.stepPacks()
		}
		return packPositions(&m)
	}
	first := run()
	for i := 0; i < 3; i++ {
		if got := run(); got != first {
			t.Fatalf("прогон %d разошёлся:\n%s\n%s", i, first, got)
		}
	}
}

// Каждая стая, созданная генератором, знает свою точку спавна и комнату.
func TestGeneratedPacksKnowTheirRoom(t *testing.T) {
	for floor := 1; floor <= 12; floor++ {
		seedRNG(int64(500 + floor))
		m := initialModel()
		m.initDungeonForFloor(floor)
		if len(m.Packs) == 0 {
			t.Fatalf("этаж %d: нет стай", floor)
		}
		for pos, pack := range m.Packs {
			if pack.Home != pos || pack.Area.W <= 0 || pack.Area.H <= 0 || !pack.inArea(pos) {
				t.Errorf("этаж %d: стая в %v: Home=%v Area=%+v", floor, pos, pack.Home, pack.Area)
			}
		}
	}
}

// Стая без комнаты (тренировочный этаж) бродит в окрестности точки спавна.
func TestPackWithoutRoomUsesDefaultArea(t *testing.T) {
	pack := ratPack()
	pack.Home = Point{20, 20}
	if !pack.inArea(Point{20 + packDefaultAreaRadius, 20}) || pack.inArea(Point{20 + packDefaultAreaRadius + 1, 20}) {
		t.Error("область стаи без комнаты должна быть радиусом packDefaultAreaRadius")
	}
}

// Ход стай не срабатывает в городе и во время боя.
func TestStepPacksSkipsTownAndCombat(t *testing.T) {
	m := arenaModel(t, 5, 5, 12, 8)
	pack := ratPack()
	m.addPack(Point{8, 8}, pack, RoomArea{5, 5, 12, 8})
	m.PartyPos = Point{12, 8}

	m.InTown = true
	m.stepPacks()
	m.InTown = false
	m.Combat = &ActiveCombat{Pack: pack}
	m.stepPacks()
	m.Combat = nil
	if m.Packs[Point{8, 8}] != pack || pack.Chasing {
		t.Error("стая двигалась в городе или во время боя")
	}
}
