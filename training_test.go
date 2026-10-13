package main

import "testing"

func levelParty(t *testing.T, maxFloor int, levels ...int) *Model {
	t.Helper()
	m := townModel(t)
	m.MaxFloor, m.Floor = maxFloor, maxFloor
	for i, h := range m.Party {
		h.Level = levels[i%len(levels)]
	}
	return m
}

func TestRecommendedLevel(t *testing.T) {
	want := map[int]int{1: 1, 2: 1, 3: 2, 4: 2, 5: 3, 10: 5, 11: 6, 20: 10}
	for floor, lvl := range want {
		if got := recommendedLevel(floor); got != lvl {
			t.Errorf("этаж %d: рекомендуемый уровень %d, ожидался %d", floor, got, lvl)
		}
	}
	if recommendedLevel(0) != 1 || recommendedLevel(-3) != 1 {
		t.Error("рекомендуемый уровень не бывает ниже первого")
	}
}

func TestWeakHeroes(t *testing.T) {
	m := levelParty(t, 10, 5, 4, 3, 2, 1) // рекомендуемый уровень 5, отставание от 2
	weak := m.weakHeroes()
	if len(weak) != 3 {
		t.Fatalf("слабых %d, ожидалось 3", len(weak))
	}
	m.Party[4].IsDead = true
	if len(m.weakHeroes()) != 2 {
		t.Fatal("павший не считается")
	}
	m.Party[3].LostInAbyss = true
	if len(m.weakHeroes()) != 1 {
		t.Fatal("оставленный в Бездне не считается")
	}
	m.Tactics.TrainingGap = 4
	if len(m.weakHeroes()) != 0 {
		t.Fatal("при допуске 4 уровня слабых нет")
	}
}

func TestPlanDepartureContinuesWhenPartyIsReady(t *testing.T) {
	m := levelParty(t, 10, 5)
	p := m.planDeparture()
	if p.Training || p.Floor != 10 || p.Reason != "gate.reason.continue" {
		t.Fatalf("готовый отряд продолжает спуск: %+v", p)
	}

	m = levelParty(t, 1, 1)
	if p := m.planDeparture(); p.Training || p.Floor != 1 {
		t.Fatalf("на первом этаже тренироваться негде: %+v", p)
	}
}

func TestPlanDepartureTrainsWeakParty(t *testing.T) {
	m := levelParty(t, 10, 5, 4, 3, 2, 1)
	p := m.planDeparture()
	if !p.Training || p.Reason != "gate.reason.weak" {
		t.Fatalf("при трёх слабых нужна тренировка: %+v", p)
	}
	if p.Floor != 3 { // 2*1+1 для самого младшего
		t.Fatalf("тренироваться стоит на этаже %d, ожидался 3", p.Floor)
	}
	if p.Floor >= m.MaxFloor {
		t.Fatal("тренировочный этаж должен быть пройденным")
	}
	// Текст пояснения подставляется без ошибок форматирования.
	if text := T(m.Lang, p.Reason, p.Args...); len(text) < 10 || text == p.Reason {
		t.Fatalf("пояснение %q", text)
	}
}

func TestPlanDepartureTrainingFloorFollowsLowestLevel(t *testing.T) {
	m := levelParty(t, 14, 7, 5, 4, 4, 4) // рекомендуемый 7: слабы те, кто на 4 и ниже (отставание 3)
	m.Tactics.TrainingGap = 3
	p := m.planDeparture()
	if !p.Training || p.Floor != 9 { // 2*4+1
		t.Fatalf("ожидалась тренировка на этаже 9, получено %+v", p)
	}

	m = levelParty(t, 6, 1, 1, 1, 1, 1) // слабы все, пройдено пять этажей
	p = m.planDeparture()
	if !p.Training || p.Floor > 5 {
		t.Fatalf("этаж тренировки не может быть глубже пройденных: %+v", p)
	}
}

func TestPlanDepartureOneWeakNeedsFailStreak(t *testing.T) {
	m := levelParty(t, 10, 5, 5, 5, 5, 1)
	if p := m.planDeparture(); p.Training {
		t.Fatalf("один слабый боец без череды неудач — не повод: %+v", p)
	}
	m.FailStreak = 1
	if p := m.planDeparture(); p.Training {
		t.Fatal("одной неудачи мало")
	}
	m.FailStreak = 2
	p := m.planDeparture()
	if !p.Training || p.Reason != "gate.reason.streak" {
		t.Fatalf("после двух неудач слабого нужно качать: %+v", p)
	}
	if text := T(m.Lang, p.Reason, p.Args...); text == p.Reason {
		t.Fatalf("у пояснения нет текста: %+v", p)
	}
}

func TestPlanDepartureRespectsTacticsSwitch(t *testing.T) {
	m := levelParty(t, 10, 1)
	m.Tactics.DisableTraining = true
	p := m.planDeparture()
	if p.Training || p.Floor != 10 || p.Reason != "gate.reason.disabled" {
		t.Fatalf("при отключённой тренировке отряд идёт вглубь: %+v", p)
	}
}

func TestPlanDepartureSyncsMaxFloor(t *testing.T) {
	m := levelParty(t, 1, 5)
	m.Floor = 6 // состояние, собранное вручную
	p := m.planDeparture()
	if m.MaxFloor != 6 || p.Floor != 6 {
		t.Fatalf("самый глубокий этаж должен подтянуться: max=%d, план %+v", m.MaxFloor, p)
	}
}

func TestNoteTripResult(t *testing.T) {
	m := townModel(t)
	m.MaxFloor, m.TripStartMaxFloor = 5, 5

	m.noteTripResult(RetreatLowHP)
	m.noteTripResult(RetreatTooFewAlive)
	if m.FailStreak != 2 {
		t.Fatalf("две неудачные вылазки подряд: %d", m.FailStreak)
	}
	m.noteTripResult(RetreatBagFull)
	if m.FailStreak != 2 {
		t.Fatal("заполненный рюкзак — не неудача")
	}
	m.MaxFloor = 6 // продвинулись
	m.noteTripResult(RetreatLowHP)
	if m.FailStreak != 0 {
		t.Fatal("продвижение сбрасывает счётчик")
	}

	m.MaxFloor, m.FailStreak = 5, 3
	m.Training = true
	m.noteTripResult(RetreatLowHP)
	if m.FailStreak != 3 {
		t.Fatal("отступление с тренировки не копит неудачи")
	}
	m.noteTripResult(RetreatTrainingDone)
	if m.FailStreak != 0 {
		t.Fatal("законченная тренировка сбрасывает счётчик")
	}
}

func TestSpawnSmallPack(t *testing.T) {
	for floor := 1; floor <= 30; floor++ {
		for i := 0; i < 20; i++ {
			p := spawnSmallPack(floor)
			if len(p.Members) < 1 || len(p.Members) > trainingMaxPack {
				t.Fatalf("этаж %d: в малой стае %d монстров", floor, len(p.Members))
			}
			if p.IsMiniBoss || p.IsBoss {
				t.Fatalf("этаж %d: в малой стае не бывает боссов", floor)
			}
		}
	}
}

func TestRepopulateTrainingFloor(t *testing.T) {
	m := levelParty(t, 8, 1)
	m.Floor = 3
	m.initDungeonForFloor(3)
	m.Packs = make(map[Point]*MonsterPack)

	m.repopulateTrainingFloor()
	if len(m.Packs) < trainingMinPacks || len(m.Packs) > trainingMaxPacks {
		t.Fatalf("на этаже %d стай — вне пределов %d..%d", len(m.Packs), trainingMinPacks, trainingMaxPacks)
	}
	var spots []Point
	for pos, p := range m.Packs {
		if m.Grid[pos.Y][pos.X] != TileFloor {
			t.Fatalf("стая стоит не на полу: %v", pos)
		}
		if absInt(pos.X-m.PartyPos.X)+absInt(pos.Y-m.PartyPos.Y) < trainingSafeDist {
			t.Fatalf("стая слишком близко ко входу: %v", pos)
		}
		if len(p.Members) > trainingMaxPack || p.IsMiniBoss {
			t.Fatalf("стая не «малая»: %+v", p)
		}
		spots = append(spots, pos)
	}
	for i := range spots {
		for j := range spots[:i] {
			if absInt(spots[i].X-spots[j].X)+absInt(spots[i].Y-spots[j].Y) < trainingPackGap {
				t.Fatalf("стаи слишком близки друг к другу: %v и %v", spots[i], spots[j])
			}
		}
	}

	// Повторное заселение доводит число стай до нормы и не трогает живых.
	count := len(m.Packs)
	m.repopulateTrainingFloor()
	if len(m.Packs) != count {
		t.Fatalf("повторное заселение при полном этаже ничего не меняет: %d → %d", count, len(m.Packs))
	}
}

func TestRepopulateKeepsExistingPacksAndTopsUp(t *testing.T) {
	m := levelParty(t, 8, 1)
	m.Floor = 3
	m.initDungeonForFloor(3)
	m.Packs = make(map[Point]*MonsterPack)
	m.repopulateTrainingFloor()
	full := len(m.Packs)

	// Часть стай «побеждена».
	removed := 0
	for pos := range m.Packs {
		if removed == 3 {
			break
		}
		delete(m.Packs, pos)
		removed++
	}
	kept := map[Point]*MonsterPack{}
	for pos, p := range m.Packs {
		kept[pos] = p
	}

	m.repopulateTrainingFloor()
	if len(m.Packs) < full-1 {
		t.Fatalf("недостающие стаи должны вернуться: было %d, стало %d", full, len(m.Packs))
	}
	for pos, p := range kept {
		if m.Packs[pos] != p {
			t.Fatalf("существующая стая %v должна остаться на месте", pos)
		}
	}
}

func TestTrainingFinishedRules(t *testing.T) {
	m := levelParty(t, 10, 1)
	m.Floor = 3
	m.initDungeonForFloor(3)

	if m.trainingFinished() {
		t.Fatal("вне тренировки она не может быть закончена")
	}
	m.Training = true
	m.Packs = map[Point]*MonsterPack{{X: 5, Y: 5}: spawnSmallPack(3)}
	if m.trainingFinished() {
		t.Fatal("есть слабые бойцы и живая стая — тренировка продолжается")
	}
	m.Packs = map[Point]*MonsterPack{}
	if !m.trainingFinished() {
		t.Fatal("на этаже не осталось стай — тренировка закончена")
	}
	m.Packs = map[Point]*MonsterPack{{X: 5, Y: 5}: spawnSmallPack(3)}
	for _, h := range m.Party {
		h.Level = 9
	}
	if !m.trainingFinished() {
		t.Fatal("слабых не осталось — тренировка закончена")
	}
}

func TestEvaluateRetreatReportsTrainingDone(t *testing.T) {
	m := levelParty(t, 10, 1)
	m.Floor = 3
	m.initDungeonForFloor(3)
	healthyParty(m)
	m.Training = true
	m.Packs = map[Point]*MonsterPack{}

	if got := m.evaluateRetreat(); got != RetreatTrainingDone {
		t.Fatalf("получено %v, ожидалось окончание тренировки", got)
	}
	m.Training = false
	if got := m.evaluateRetreat(); got == RetreatTrainingDone {
		t.Fatal("вне тренировки такой причины нет")
	}
}

func TestDepartFromTownContinueAndTraining(t *testing.T) {
	m := levelParty(t, 8, 1)
	m.InTown = true

	m.departFromTown(departurePlan{Floor: 8})
	if m.InTown || m.Training || m.Floor != 8 || m.TripStartMaxFloor != 8 {
		t.Fatalf("продолжение спуска: town=%v training=%v floor=%d", m.InTown, m.Training, m.Floor)
	}

	m.InTown = true
	m.MaxFloor = 8
	m.departFromTown(departurePlan{Floor: 3, Training: true})
	if !m.Training || m.Floor != 3 || m.MaxFloor != 8 {
		t.Fatalf("тренировка: training=%v floor=%d max=%d", m.Training, m.Floor, m.MaxFloor)
	}
	if len(m.Packs) < trainingMinPacks {
		t.Fatalf("тренировочный этаж должен быть заселён: стай %d", len(m.Packs))
	}

	m.InTown = true
	m.departFromTown(departurePlan{Floor: 99})
	if m.Floor != 8 {
		t.Fatalf("глубже самого глубокого этажа спуститься нельзя: %d", m.Floor)
	}
}

func TestTrainingFloorRestoresVisitedLayout(t *testing.T) {
	m := levelParty(t, 1, 1)
	m.initDungeonForFloor(1)
	m.saveCurrentFloorState()
	exitBefore := m.VisitedFloors[1].ExitPos

	m.Floor, m.MaxFloor = 4, 4
	m.initDungeonForFloor(4)
	m.saveCurrentFloorState()

	m.InTown = true
	m.departFromTown(departurePlan{Floor: 1, Training: true})
	if m.VisitedFloors[1] == nil || m.VisitedFloors[1].ExitPos != exitBefore {
		t.Fatal("пройденный этаж должен вернуться с прежней планировкой")
	}
	if m.Floor != 1 || !m.Training {
		t.Fatal("тренировка на первом этаже")
	}
}

func TestTrainingRunNeverDescendsAndGoesHome(t *testing.T) {
	for seed := int64(90); seed < 93; seed++ {
		m := levelParty(t, 8, 1)
		seedRNG(seed)
		plan := m.planDeparture()
		if !plan.Training {
			t.Fatalf("seed %d: слабому отряду нужна тренировка", seed)
		}
		m.InTown = true
		m.departFromTown(plan)
		floor := m.Floor

		for i := 0; i < 5000 && !m.InTown && m.State == StatePlaying; i++ {
			m.step()
			if !m.InTown && m.Floor != floor {
				t.Fatalf("seed %d, шаг %d: на тренировке отряд сменил этаж %d → %d", seed, i, floor, m.Floor)
			}
		}
		if m.InTown {
			if m.Training {
				t.Fatalf("seed %d: после возвращения в город тренировка должна быть сброшена", seed)
			}
			if m.Floor != m.MaxFloor {
				t.Fatalf("seed %d: в городе этаж прогресса %d, получено %d", seed, m.MaxFloor, m.Floor)
			}
		}
	}
}

func TestClearedTrainingFloorSendsPartyHome(t *testing.T) {
	m := levelParty(t, 8, 1)
	m.InTown = true
	m.departFromTown(departurePlan{Floor: 3, Training: true})
	m.Packs = make(map[Point]*MonsterPack) // этаж «зачищен»
	healthyParty(m)

	for i := 0; i < 4000 && !m.InTown; i++ {
		m.step()
	}
	if !m.InTown {
		t.Fatal("после зачистки отряд должен вернуться в город")
	}
	if m.Training || m.Floor != 8 {
		t.Fatalf("training=%v floor=%d", m.Training, m.Floor)
	}
}

func TestGateAutopilotUsesPlan(t *testing.T) {
	m := levelParty(t, 8, 1)
	m.enterTown(RetreatLowHP)
	m.enterBuilding(m.Town.Gate)
	if m.TownPhase != TownPhaseDepart || !m.TownVisit {
		t.Fatal("отряд у Врат")
	}
	if m.GateState.ActionSummary == "" || !m.GateState.Plan.Training {
		t.Fatalf("автопилот должен показать решение: %+v", m.GateState)
	}
	for i := 0; i < 10 && m.InTown; i++ {
		m.step()
	}
	if m.InTown || !m.Training {
		t.Fatalf("слабый отряд должен уйти на тренировку: town=%v training=%v", m.InTown, m.Training)
	}
}

func TestTacticsTrainingRows(t *testing.T) {
	if tacticsCount != 9 {
		t.Fatalf("строк тактики %d, ожидалось 9", tacticsCount)
	}
	tc := DefaultTactics()
	if tc.DisableTraining || tc.TrainingGap != 2 {
		t.Fatalf("по умолчанию тренировка включена, допуск 2: %+v", tc)
	}
	if tc.value(7) != 1 {
		t.Fatal("строка «тренировка» по умолчанию «ВКЛ»")
	}
	tc.adjust(7, +1)
	if !tc.DisableTraining || tc.value(7) != 0 {
		t.Fatal("переключатель должен выключать тренировку")
	}
	tc.adjust(7, +1)
	if tc.DisableTraining {
		t.Fatal("и включать обратно")
	}
	for i := 0; i < 20; i++ {
		tc.adjust(8, +1)
	}
	if tc.TrainingGap != tacticRanges[8].max {
		t.Fatalf("допуск ограничен сверху: %d", tc.TrainingGap)
	}
	bad := Tactics{TrainingGap: 0, RetreatHPPct: 35, RetreatMinAlive: 2, FleeHPPct: 25, PotionHPPct: 40}
	if bad.Clamped().TrainingGap != tacticRanges[8].min {
		t.Fatal("нулевой допуск поднимается до минимума")
	}
}

func TestOldSaveGetsDefaultTrainingGap(t *testing.T) {
	path := useTempSave(t)
	writeTestFile(t, path, `{"version":1,"lang":"ru","tactics":{"retreat_hp_pct":40,"retreat_min_alive":2,"flee_hp_pct":25,"potion_hp_pct":40}}`)
	got, ok := loadSave()
	if !ok {
		t.Fatal("старое сохранение должно читаться")
	}
	if got.Tactics.TrainingGap != DefaultTactics().TrainingGap || got.Tactics.DisableTraining {
		t.Fatalf("сохранения до 3.3 получают тренировку с допуском по умолчанию: %+v", got.Tactics)
	}
}
