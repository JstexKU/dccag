package main

import "testing"

func TestDefaultTacticsMatchLegacyBehaviour(t *testing.T) {
	d := DefaultTactics()
	if d.RetreatHPPct != 35 || d.RetreatMinAlive != 2 || d.FleeHPPct != 25 || d.PotionHPPct != 40 {
		t.Fatalf("значения по умолчанию разошлись с исходной логикой: %+v", d)
	}
	if d.SkipTraps || d.SkipAltars || !d.GambleEvents {
		t.Fatalf("флаги по умолчанию неверны: %+v", d)
	}
	if d != d.Clamped() {
		t.Fatal("настройки по умолчанию должны укладываться в допустимые диапазоны")
	}
}

func TestTacticsAdjustClampsAndToggles(t *testing.T) {
	tc := DefaultTactics()
	for i := 0; i < 50; i++ {
		tc.adjust(0, +1)
	}
	if tc.RetreatHPPct != tacticRanges[0].max {
		t.Errorf("верхняя граница не соблюдена: %d", tc.RetreatHPPct)
	}
	for i := 0; i < 50; i++ {
		tc.adjust(0, -1)
	}
	if tc.RetreatHPPct != tacticRanges[0].min {
		t.Errorf("нижняя граница не соблюдена: %d", tc.RetreatHPPct)
	}

	before := tc.SkipTraps
	tc.adjust(4, +1)
	if tc.SkipTraps == before {
		t.Error("булева настройка должна переключаться")
	}
	tc.adjust(4, -1)
	if tc.SkipTraps != before {
		t.Error("повторное переключение должно вернуть исходное значение")
	}
}

func TestTacticsClampedFixesGarbage(t *testing.T) {
	bad := Tactics{RetreatHPPct: 999, RetreatMinAlive: -5, FleeHPPct: -1, PotionHPPct: 5000}
	fixed := bad.Clamped()
	for i := 0; i < tacticsCount; i++ {
		r := tacticRanges[i]
		if v := fixed.value(i); v < r.min || v > r.max {
			t.Errorf("настройка %d вне диапазона: %d", i, v)
		}
	}
}

func TestEvaluateRetreatRespectsTactics(t *testing.T) {
	seedRNG(1)
	m := initialModel()
	m.State = StatePlaying
	for _, h := range m.Party {
		h.HP = h.MaxHP * 30 / 100
		h.Stress = 0
	}

	m.Tactics.RetreatHPPct = 35
	if got := m.evaluateRetreat(); got != RetreatLowHP {
		t.Fatalf("при пороге 35%% и HP 30%% ожидалось отступление, получено %v", got)
	}
	m.Tactics.RetreatHPPct = 20
	if got := m.evaluateRetreat(); got != RetreatNone {
		t.Fatalf("при пороге 20%% и HP 30%% отступать не нужно, получено %v", got)
	}
}

func TestShouldAttemptFleeRespectsTactics(t *testing.T) {
	seedRNG(2)
	m := initialModel()
	m.State = StatePlaying
	m.Tactics.RetreatHPPct = 10 // чтобы сработал именно обычный побег, а не экстренное отступление
	m.startCombat(Point{1, 1}, spawnMonsterPack(false, 1))
	for _, h := range m.Party {
		h.HP = h.MaxHP * 30 / 100
	}

	m.Tactics.FleeHPPct = 25
	if m.shouldAttemptFlee() {
		t.Error("при HP 30% и пороге 25% бежать не нужно")
	}
	m.Tactics.FleeHPPct = 40
	if !m.shouldAttemptFlee() {
		t.Error("при HP 30% и пороге 40% отряд должен бежать")
	}
}
