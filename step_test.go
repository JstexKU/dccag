package main

import "testing"

// Тесты ветвей step.go, до которых не доходят эталонные забеги.

func newPlayingModel(t *testing.T) Model {
	t.Helper()
	seedRNG(3)
	m := initialModel()
	m.State = StatePlaying
	m.Training = false
	m.ManualMode = false
	return m
}

// Выход из зацикливания в обычной экспедиции — спуск на этаж ниже без возврата в город.
func TestBreakLoopDescends(t *testing.T) {
	m := newPlayingModel(t)
	floor := m.Floor
	m.PathHistory = []Point{{1, 1}, {1, 2}}
	m.LoopDetectCount = loopAbortThreshold

	m.breakLoop()

	if m.Floor != floor+1 || m.Stats.FloorsCleared != 1 {
		t.Errorf("этаж %d (ждали %d), пройдено %d", m.Floor, floor+1, m.Stats.FloorsCleared)
	}
	if m.InTown {
		t.Error("отряд не должен уходить в город")
	}
	if len(m.PathHistory) != 0 || m.LoopDetectCount != 0 {
		t.Error("история пути и счётчик зацикливания должны сбрасываться")
	}
}

// Тренировка в автопилоте при зацикливании завершается возвратом в город.
func TestBreakLoopEndsTraining(t *testing.T) {
	m := newPlayingModel(t)
	m.Training = true
	floor := m.Floor

	m.breakLoop()

	if !m.InTown || m.Training {
		t.Errorf("InTown=%v Training=%v: ждали возврат в город и конец тренировки", m.InTown, m.Training)
	}
	if m.Stats.FloorsCleared != 0 || m.Floor < floor {
		t.Error("тренировка не должна засчитывать пройденные этажи")
	}
}

// При полной сумке сундук остаётся на месте и золото не выдаётся.
func TestFullBagLeavesChests(t *testing.T) {
	for _, tile := range []Tile{TileChest, TileTrappedChest} {
		m := newPlayingModel(t)
		m.Bag = make([]EquipItem, m.currentBagCapacity())
		at := Point{1, 1}
		m.Grid[at.Y][at.X] = tile
		gold := m.Gold

		var left bool
		if tile == TileChest {
			left = m.walkOnChest(at)
		} else {
			left = m.walkOnTrappedChest(at)
		}

		if left {
			t.Errorf("%q: отряд не должен покидать этаж", tile)
		}
		if m.Grid[at.Y][at.X] != tile || m.Gold != gold || m.Stats.ChestsOpened != 0 {
			t.Errorf("%q: сундук должен остаться закрытым при полной сумке", tile)
		}
	}
}

// Квест «Полная зачистка» не пускает вниз, пока на этаже живы стаи.
func TestStairsBlockedByPurgeQuest(t *testing.T) {
	m := newPlayingModel(t)
	m.CurrentQuest = AutoQuest{Type: QuestPurgeBiome}
	m.Packs[Point{2, 2}] = &MonsterPack{}
	floor := m.Floor

	if m.walkOnStairs(Point{3, 3}) {
		t.Error("спуск должен быть заблокирован")
	}
	if m.Floor != floor {
		t.Errorf("этаж изменился: %d -> %d", floor, m.Floor)
	}
}

// Каждая плитка с поведением имеет обработчик, а обычный пол — нет.
func TestTileHandlersTable(t *testing.T) {
	for _, tile := range []Tile{TileChest, TileTrappedChest, TileRelic, TileAltar, TileFountain,
		TileLore, TileBarrel, TileEvent, TileStairs} {
		if tileHandlers[tile] == nil {
			t.Errorf("у плитки %q нет обработчика", tile)
		}
	}
	if tileHandlers[TileFloor] != nil || tileHandlers[TileWall] != nil {
		t.Error("у пола и стены обработчика быть не должно")
	}
}
