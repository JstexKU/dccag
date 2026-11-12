package main

import "testing"

func stressModel(t *testing.T) Model {
	t.Helper()
	seedRNG(31)
	m := initialModel()
	m.State = StatePlaying
	m.Relic = nil
	return m
}

// Предметы снижают стресс не более чем на 50%, раса и предметы применяются по очереди.
func TestReduceStressItemCapAndRace(t *testing.T) {
	m := stressModel(t)
	h := m.Party[1]
	h.IsLeader = false
	h.Head, h.Chest, h.Legs = &EquipItem{StressRes: 40}, &EquipItem{StressRes: 40}, nil

	raceRes := GetRaceModifiers(h.Race).StressResPercent
	want := 100
	if raceRes > 0 {
		want = want * (100 - raceRes) / 100
	}
	want = want * 50 / 100 // предметы: 80% режутся до 50%

	if m.activeCalling() == CallingInspirer {
		t.Skip("призвание Вдохновитель добавляет свой множитель")
	}
	if got := m.reduceStress(h, 100); got != want {
		t.Errorf("reduceStress=%d, ожидалось %d", got, want)
	}

	// Без предметов и без реликвии остаётся только расовое сопротивление.
	h.Head, h.Chest = nil, nil
	wantRace := 100
	if raceRes > 0 {
		wantRace = wantRace * (100 - raceRes) / 100
	}
	if got := m.reduceStress(h, 100); got != wantRace {
		t.Errorf("без предметов reduceStress=%d, ожидалось %d", got, wantRace)
	}
}

// Реликвия со StressRes режет стресс до расы и предметов.
func TestReduceStressRelic(t *testing.T) {
	m := stressModel(t)
	h := m.Party[0] // лидер: призвание на него не действует
	h.Head, h.Chest, h.Legs = nil, nil, nil
	m.Relic = &PartyRelic{StressRes: 50}

	want := 50
	if raceRes := GetRaceModifiers(h.Race).StressResPercent; raceRes > 0 {
		want = want * (100 - raceRes) / 100
	}
	if got := m.reduceStress(h, 100); got != want {
		t.Errorf("reduceStress=%d, ожидалось %d", got, want)
	}
}

// Союзники получают половину стресса, но только от 15 и выше.
func TestApplyStressSplash(t *testing.T) {
	m := stressModel(t)
	h := m.Party[1]
	for _, p := range m.Party {
		p.Stress = 0
	}

	m.applyStressSplash(h, 14)
	for i, p := range m.Party {
		if p.Stress != 0 {
			t.Errorf("герой %d получил стресс %d при слабом ударе", i, p.Stress)
		}
	}

	m.Party[2].IsDowned = true
	m.applyStressSplash(h, 40)
	for i, p := range m.Party {
		want := 20
		if p == h || p.IsDowned || p.IsDead {
			want = 0
		}
		if p.Stress != want {
			t.Errorf("герой %d: стресс %d, ожидалось %d", i, p.Stress, want)
		}
	}
}

// Сердечный приступ: с одним очком здоровья герой погибает, иначе остаётся на грани;
// добродетельный приступу не подвержен.
func TestCheckHeartAttack(t *testing.T) {
	m := stressModel(t)
	h := m.Party[1]
	h.Stress, h.HP, h.Affliction = 200, 40, AfflictionNone
	m.checkHeartAttack(h)
	if h.HP != 1 || h.Stress != 160 || h.IsDowned {
		t.Errorf("приступ без смерти: HP=%d Stress=%d Downed=%v", h.HP, h.Stress, h.IsDowned)
	}

	h.Stress, h.HP = 200, 1
	m.checkHeartAttack(h)
	if !h.IsDowned || h.HP != 0 {
		t.Errorf("приступ со смертью: HP=%d Downed=%v", h.HP, h.IsDowned)
	}

	v := m.Party[3]
	v.Stress, v.HP, v.Affliction = 200, 1, AfflictionVirtuous
	m.checkHeartAttack(v)
	if v.IsDowned || v.HP != 1 || v.Stress != 200 {
		t.Errorf("добродетельный не должен страдать: HP=%d Stress=%d Downed=%v", v.HP, v.Stress, v.IsDowned)
	}
}

// Срыв наступает один раз и только при 100 стресса у героя без состояния.
func TestCheckBreakingPoint(t *testing.T) {
	m := stressModel(t)
	h := m.Party[1]
	h.Affliction, h.Stress = AfflictionNone, 99
	m.checkBreakingPoint(h)
	if h.Affliction != AfflictionNone {
		t.Fatal("срыв до 100 стресса")
	}

	h.Stress = 100
	m.checkBreakingPoint(h)
	if h.Affliction == AfflictionNone {
		t.Fatal("на 100 стресса состояние не изменилось")
	}
	if h.Affliction != AfflictionVirtuous && h.Stress < 100 {
		t.Errorf("психоз не должен сбрасывать стресс: %d", h.Stress)
	}

	after, stress := h.Affliction, h.Stress
	m.checkBreakingPoint(h)
	if h.Affliction != after || h.Stress != stress {
		t.Error("повторный вызов не должен менять героя с состоянием")
	}
}
