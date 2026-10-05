package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// runHeadless прогоняет игру без интерфейса и возвращает итоговую модель.
func runHeadless(seed int64, steps int) Model {
	seedRNG(seed)
	m := initialModel()
	m.State = StatePlaying
	for i := 0; i < steps && m.State == StatePlaying; i++ {
		m.step()
	}
	return m
}

// Дымовой тест: игра не должна паниковать на разных зёрнах и при разных тактиках.
func TestSimulationDoesNotPanic(t *testing.T) {
	for seed := int64(1); seed <= 8; seed++ {
		seedRNG(seed)
		m := initialModel()
		m.State = StatePlaying
		switch seed % 4 {
		case 1:
			m.Tactics.SkipTraps = true
			m.Tactics.SkipAltars = true
		case 2:
			m.Tactics.GambleEvents = false
			m.Tactics.RetreatHPPct = 60
		case 3:
			m.Tactics.RetreatMinAlive = 0
			m.Tactics.FleeHPPct = 10
		}
		for i := 0; i < 6000 && m.State == StatePlaying; i++ {
			m.step()
		}
		if m.Stats.TotalSteps == 0 {
			t.Errorf("seed %d: отряд не сделал ни одного шага", seed)
		}
	}
}

// Одинаковое зерно обязано давать одинаковый забег — на этом держится воспроизводимость багов.
func TestSimulationIsDeterministic(t *testing.T) {
	a := runHeadless(42, 2500)
	b := runHeadless(42, 2500)
	if a.Floor != b.Floor || a.Gold != b.Gold || a.Stats.TotalSteps != b.Stats.TotalSteps ||
		a.Stats.ChestsOpened != b.Stats.ChestsOpened || len(a.Logs) != len(b.Logs) {
		t.Fatalf("забеги с одним зерном разошлись:\n a: floor=%d gold=%d steps=%d\n b: floor=%d gold=%d steps=%d",
			a.Floor, a.Gold, a.Stats.TotalSteps, b.Floor, b.Gold, b.Stats.TotalSteps)
	}
}

// Устаревший тик (от предыдущей цепочки) не должен двигать игру — это и был баг удвоения скорости.
func TestStaleTickIsIgnored(t *testing.T) {
	seedRNG(5)
	m := initialModel()
	m.State = StatePlaying

	m.restartTicks(10)
	stale := m.TickGen
	m.restartTicks(10)
	current := m.TickGen
	if stale == current {
		t.Fatal("restartTicks должен менять поколение")
	}

	before := m.Stats.TotalSteps
	next, cmd := m.Update(TickMsg{Gen: stale})
	m = next.(Model)
	if m.Stats.TotalSteps != before || cmd != nil {
		t.Fatal("устаревший тик не должен выполнять шаг и перепланировать цепочку")
	}

	next, cmd = m.Update(TickMsg{Gen: current})
	m = next.(Model)
	if m.Stats.TotalSteps != before+1 || cmd == nil {
		t.Fatal("актуальный тик должен выполнить шаг и запланировать следующий")
	}
}

func TestOpeningAndClosingOverlayInvalidatesOldTicks(t *testing.T) {
	seedRNG(6)
	m := initialModel()
	m.State = StatePlaying
	m.restartTicks(10)
	old := m.TickGen

	// Открыли и сразу закрыли Арсенал: старый тик ещё «в полёте».
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = next.(Model)
	if cmd == nil || m.TickGen == old {
		t.Fatal("закрытие окна должно запускать новую цепочку тиков")
	}

	before := m.Stats.TotalSteps
	next, _ = m.Update(TickMsg{Gen: old})
	m = next.(Model)
	if m.Stats.TotalSteps != before {
		t.Fatal("тик старой цепочки не должен выполняться")
	}
}

func TestRestartOnlyFromDefeatOrGlory(t *testing.T) {
	seedRNG(9)
	m := initialModel()
	m.State = StateArmory
	gold := m.Gold
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = next.(Model)
	if m.State != StateArmory || m.Gold != gold {
		t.Fatal("R в Арсенале не должна сбрасывать забег")
	}
}

func TestTacticsScreenRoundTrip(t *testing.T) {
	seedRNG(10)
	m := initialModel()
	m.State = StatePlaying

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = next.(Model)
	if m.State != StateTactics {
		t.Fatal("T должна открывать экран тактики")
	}
	if m.View() == "" {
		t.Fatal("экран тактики должен рендериться")
	}

	before := m.Tactics.RetreatHPPct
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	if m.Tactics.RetreatHPPct != before+tacticRanges[0].step {
		t.Fatalf("стрелка вправо должна увеличивать порог: %d -> %d", before, m.Tactics.RetreatHPPct)
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.State != StatePlaying || cmd == nil {
		t.Fatal("Esc должна возвращать в игру и перезапускать тики")
	}
}

// TestFloorStatePersistenceAndRestoration проверяет сохранение изменений на этаже при повторном входе.
func TestFloorStatePersistenceAndRestoration(t *testing.T) {
	seedRNG(101)
	m := initialModel()
	m.State = StatePlaying

	// 1. Модифицируем 1-й этаж: открываем сундук, убираем пак монстров, отмечаем разведанную область
	var chestPt Point
	chestFound := false
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			if m.Grid[y][x] == TileChest {
				chestPt = Point{x, y}
				chestFound = true
				break
			}
		}
		if chestFound {
			break
		}
	}
	if chestFound {
		m.Grid[chestPt.Y][chestPt.X] = TileFloor // симулируем вскрытие сундука
	}

	initialPackCount := len(m.Packs)
	for pt := range m.Packs {
		delete(m.Packs, pt) // симулируем уничтожение одного пака
		break
	}
	m.Explored[10][10] = true

	// Сохраняем состояние этажа 1
	m.saveCurrentFloorState()

	// 2. Переходим на этаж 2
	m.Floor = 2
	m.initDungeonForFloor(2)
	if m.Floor != 2 || m.VisitedFloors[1] == nil {
		t.Fatal("этаж 2 должен инициализироваться, а этаж 1 оставаться в памяти")
	}

	// 3. Возвращаемся обратно на этаж 1
	m.Floor = 1
	m.initDungeonForFloor(1)

	if chestFound && m.Grid[chestPt.Y][chestPt.X] != TileFloor {
		t.Fatal("открытый сундук не должен восстанавливаться при возврате на этаж")
	}
	if len(m.Packs) != initialPackCount-1 {
		t.Fatalf("уничтоженные монстры не должны возрождаться: паков %d, ожидалось %d", len(m.Packs), initialPackCount-1)
	}
	if !m.Explored[10][10] {
		t.Fatal("разведанные участки карты должны сохраняться")
	}
}

// TestDungeonConnectivityAndScale проверяет масштаб карты и доступность лестницы через BFS.
func TestDungeonConnectivityAndScale(t *testing.T) {
	for seed := int64(200); seed <= 205; seed++ {
		seedRNG(seed)
		m := initialModel()
		m.initDungeonForFloor(5) // проверяем глубокий этаж

		if m.MapWidth < 110 || m.MapHeight < 45 {
			t.Fatalf("seed %d: масштаб карты недостаточен (%dx%d)", seed, m.MapWidth, m.MapHeight)
		}

		// Ищем лестницу спуска (TileStairs) с помощью BFS
		_, found := m.findEmergencyStep(TileStairs)
		if !found {
			t.Fatalf("seed %d: лестница спуска недостижима из точки спавна отряда", seed)
		}
	}
}
