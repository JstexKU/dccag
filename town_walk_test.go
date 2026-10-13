package main

import "testing"

func townModel(t *testing.T) *Model {
	t.Helper()
	useTempSave(t)
	seedRNG(81)
	m := initialModel()
	m.State = StatePlaying
	return &m
}

func phaseIndex(p TownPhase) int {
	for i, s := range townSpecs {
		if s.Phase == p {
			return i
		}
	}
	return -1
}

// healthyParty приводит отряд в безупречное состояние: лечить и воскрешать некого.
func healthyParty(m *Model) {
	for _, h := range m.Party {
		h.HP, h.MP, h.Stress = h.MaxHP, h.MaxMP, 0
		h.Affliction = AfflictionNone
		h.IsDowned, h.IsDead = false, false
	}
	m.CurrentQuest.Completed = false
}

// runTownAuto прогоняет автопилот по городу до выхода в подземелье и возвращает порядок визитов.
func runTownAuto(t *testing.T, m *Model, maxSteps int) []TownPhase {
	t.Helper()
	var visited []TownPhase
	for i := 0; i < maxSteps && m.InTown; i++ {
		before := m.TownVisit
		m.step()
		if m.TownVisit && !before {
			visited = append(visited, m.TownPhase)
		}
	}
	if m.InTown {
		t.Fatal("отряд не вышел из города за разумное число шагов")
	}
	return visited
}

func TestEnterTownPlacesPartyAtTheGate(t *testing.T) {
	m := townModel(t)
	m.Floor, m.MaxFloor = 4, 4
	m.enterTown(RetreatLowHP)

	if !m.InTown || !m.TownWalking || m.TownVisit {
		t.Fatal("отряд приходит в город и идёт по улицам")
	}
	if m.TownPos != m.Town.Buildings[m.Town.Gate].Front {
		t.Fatalf("отряд появляется у Врат, а он на %v", m.TownPos)
	}
	if m.TownPhase != TownPhaseSellLoot {
		t.Fatalf("обход начинается с рынка, а не с %v", m.TownPhase)
	}
}

func TestEnterTownAfterTrainingRestoresProgressFloor(t *testing.T) {
	m := townModel(t)
	m.MaxFloor, m.Floor = 7, 3
	m.Training = true
	m.FailStreak = 4
	m.enterTown(RetreatTrainingDone)

	if m.Training || m.Floor != 7 {
		t.Fatalf("после тренировки этаж прогресса должен вернуться: training=%v floor=%d", m.Training, m.Floor)
	}
	if m.FailStreak != 0 {
		t.Fatalf("успешная тренировка сбрасывает счётчик неудач, а он %d", m.FailStreak)
	}
	if m.TownDialog != T(m.Lang, "gate.log.training_done") {
		t.Fatal("после тренировки должна быть своя реплика")
	}
}

func TestAutopilotVisitsNeededBuildingsInOrderAndLeavesThroughGate(t *testing.T) {
	m := townModel(t)
	m.Tactics.DisableTraining = true
	m.Gold = 2000
	for _, h := range m.Party {
		h.HP = max(1, h.MaxHP/3)
		h.Stress = 150 // рынок снимает часть стресса, но храм всё равно нужен
	}
	m.enterTown(RetreatLowHP)
	visited := runTownAuto(t, m, 6000)

	if len(visited) < 3 {
		t.Fatalf("раненый отряд с золотом должен обойти несколько заведений, а визитов %d", len(visited))
	}
	if visited[0] != TownPhaseSellLoot {
		t.Fatalf("первым идёт рынок, а не %v", visited[0])
	}
	if visited[len(visited)-1] != TownPhaseDepart {
		t.Fatalf("последними идут Врата, а не %v", visited[len(visited)-1])
	}
	last := -1
	seen := map[TownPhase]bool{}
	for _, p := range visited {
		idx := phaseIndex(p)
		if idx <= last {
			t.Fatalf("порядок обхода нарушен: %v после здания с номером %d", p, last)
		}
		last = idx
		seen[p] = true
	}
	if !seen[TownPhaseChurch] || !seen[TownPhaseTavern] {
		t.Fatalf("раненый и напуганный отряд обязан зайти в храм и таверну: %v", visited)
	}
	if m.InTown || m.TownWalking || m.TownVisit {
		t.Fatal("после Врат отряд должен быть в подземелье")
	}
}

func TestHealthyPartyWithoutGoldSkipsIdleBuildings(t *testing.T) {
	m := townModel(t)
	m.Tactics.DisableTraining = true
	m.Gold, m.Bag = 0, nil
	healthyParty(m)
	m.enterTown(RetreatNone)

	visited := runTownAuto(t, m, 3000)
	if len(visited) != 2 || visited[0] != TownPhaseSellLoot || visited[1] != TownPhaseDepart {
		t.Fatalf("здоровому отряду без золота нужны только рынок и Врата, а визиты: %v", visited)
	}
}

func TestNeedsVisitRules(t *testing.T) {
	m := townModel(t)
	m.Gold, m.Bag = 0, nil
	healthyParty(m)

	for _, p := range []TownPhase{TownPhaseSellLoot, TownPhaseDepart} {
		if !m.townNeedsVisit(p) {
			t.Errorf("%v посещается всегда", p)
		}
	}
	for _, p := range []TownPhase{TownPhaseMagistrate, TownPhaseChurch, TownPhaseTavern, TownPhaseGuild, TownPhaseSmithy, TownPhaseTannery, TownPhaseAlchemist} {
		if m.townNeedsVisit(p) {
			t.Errorf("здоровому отряду без золота %v не нужен", p)
		}
	}

	m.Party[0].HP = 1
	if !m.townNeedsVisit(TownPhaseTavern) {
		t.Error("раненому нужна таверна")
	}
	healthyParty(m)
	m.Party[1].Stress = 60
	if !m.townNeedsVisit(TownPhaseChurch) {
		t.Error("напуганному нужен храм")
	}
	healthyParty(m)
	m.Party[2].IsDowned = true
	if !m.townNeedsVisit(TownPhaseChurch) || !m.townNeedsVisit(TownPhaseGuild) {
		t.Error("павшему нужны храм и гильдия")
	}
	healthyParty(m)
	m.CurrentQuest.Completed = true
	if !m.townNeedsVisit(TownPhaseGuild) {
		t.Error("за выполненный контракт награда выдаётся в гильдии")
	}

	if got := m.nextWorthPhase(TownPhaseMagistrate); got != TownPhaseGuild {
		t.Errorf("после магистрата ближайшее нужное — гильдия, получено %v", got)
	}
	healthyParty(m)
	if got := m.nextWorthPhase(TownPhaseMagistrate); got != TownPhaseDepart {
		t.Errorf("если нужных заведений нет, остаются Врата, получено %v", got)
	}
}

func TestManualTownWalkingEnteringAndLeaving(t *testing.T) {
	m := townModel(t)
	m.ManualMode = true
	m.enterTown(RetreatNone)
	b := m.Town.Buildings[0]

	// Стена здания не пускает.
	m.TownPos = Point{b.Door.X - 1, b.Door.Y + 1}
	here := m.TownPos
	press(m, "up")
	if m.TownPos != here || m.TownVisit {
		t.Fatal("в стену здания ходить нельзя")
	}

	// Дверь открывает карточку.
	m.TownPos = b.Front
	press(m, "up")
	if !m.TownVisit || m.TownWalking || m.TownPhase != b.Phase || m.TownPos != b.Door {
		t.Fatalf("шаг на дверь должен открывать заведение: visit=%v walking=%v", m.TownVisit, m.TownWalking)
	}

	// Esc выводит на улицу перед дверью, фаза остаётся прежней.
	press(m, "esc")
	if m.TownVisit || !m.TownWalking || m.TownPos != b.Front || m.TownPhase != b.Phase {
		t.Fatal("Esc должна выводить на улицу перед дверью")
	}

	// Стрелки двигают по улице.
	press(m, "right")
	if m.TownPos != (Point{b.Front.X + 1, b.Front.Y}) {
		t.Fatalf("шаг вправо по улице: %v", m.TownPos)
	}
}

func TestManualBuildingThatAdvancesPhaseSendsPartyToTheStreet(t *testing.T) {
	m := townModel(t)
	m.ManualMode = true
	m.enterTown(RetreatNone)
	idx := m.Town.buildingFor(TownPhaseMagistrate)
	m.enterBuilding(idx)
	if m.TownPhase != TownPhaseMagistrate {
		t.Fatal("вошли в магистрат")
	}

	// Старый обработчик по Esc сам переводит фазу «дальше»: на карте это значит «вышел на улицу».
	press(m, "esc")
	if m.TownVisit || !m.TownWalking {
		t.Fatal("после Esc отряд на улице")
	}
	if m.TownPhase != TownPhaseMagistrate {
		t.Fatalf("фаза не должна уезжать вперёд, получено %v", m.TownPhase)
	}
}

func TestLegacyPipelineStillWorksWithoutMapState(t *testing.T) {
	m := townModel(t)
	m.InTown = true
	m.TownPhase = TownPhaseChurch // как в старых тестах: карта не задействована
	for i := 0; i < 100 && m.TownPhase == TownPhaseChurch; i++ {
		m.step()
	}
	if m.TownPhase == TownPhaseChurch {
		t.Fatal("прежний конвейер должен продвигаться без карты")
	}
	if m.TownWalking || m.TownVisit {
		t.Fatal("старый путь не должен включать ходьбу")
	}
}

func TestTownTickDelays(t *testing.T) {
	m := townModel(t)
	m.SpeedMs, m.TownDelayMs = 270, 500
	if m.tickDelay() != 270 {
		t.Fatal("в подземелье пауза равна скорости игры")
	}
	m.InTown = true
	m.TownWalking = true
	if got := m.tickDelay(); got != 90 {
		t.Fatalf("по улицам идём втрое быстрее: %d", got)
	}
	m.SpeedMs = 30
	if got := m.tickDelay(); got < 30 {
		t.Fatalf("слишком быстро: %d", got)
	}
	m.TownWalking = false
	if m.tickDelay() != 500 {
		t.Fatal("в заведении пауза города")
	}
}

func TestBarnRestNeverLowersHealth(t *testing.T) {
	m := townModel(t)
	m.Gold = 100
	for _, h := range m.Party {
		h.HP, h.MP = h.MaxHP*9/10, h.MaxMP*9/10
	}
	before := m.Party[0].HP
	m.ExecuteTavernRest(TavernOffer{Cost: 5, IsLuxury: false})
	if m.Party[0].HP < before {
		t.Fatalf("сарай отнял здоровье: %d → %d", before, m.Party[0].HP)
	}

	m.Party[1].HP, m.Party[1].MP = 1, 0
	m.ExecuteTavernRest(TavernOffer{Cost: 5, IsLuxury: false})
	if m.Party[1].HP <= 1 {
		t.Fatal("тяжело раненому сарай должен помочь")
	}
}
