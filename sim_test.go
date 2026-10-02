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
