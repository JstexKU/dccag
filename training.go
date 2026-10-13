package main

// ============================================================
// ТРЕНИРОВКА НА ПРОЙДЕННЫХ ЭТАЖАХ
// ============================================================
//
// Когда в отряде много слабых бойцов (новобранцы, пришедшие на смену павшим), спуск
// на самый глубокий этаж превращается в цепочку смертей и отступлений. Выход:
// спуститься на один из пройденных этажей, который заново заселяется небольшими
// стаями монстров, подходящих этому этажу, набрать уровни и вернуться.

const (
	trainingMinPacks = 8  // сколько стай держится на тренировочном этаже
	trainingMaxPacks = 26 //
	trainingCellsPer = 260
	trainingMaxPack  = 2  // монстров в одной стае
	trainingSafeDist = 12 // рядом со входом стаи не селятся
	trainingPackGap  = 7  // минимальное расстояние между стаями
)

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// recommendedLevel — уровень героя, при котором этаж считается посильным.
// Совпадает с соглашением генерации героев: уровень ≈ этаж / 2.
func recommendedLevel(floor int) int {
	return max(1, (floor+1)/2)
}

// syncMaxFloor подтягивает самый глубокий этаж к текущему: так состояния, собранные
// вручную (например, в тестах), остаются согласованными.
func (m *Model) syncMaxFloor() {
	if m.Floor > m.MaxFloor {
		m.MaxFloor = m.Floor
	}
	if m.MaxFloor < 1 {
		m.MaxFloor = 1
	}
}

// weakHeroes возвращает живых героев, отстающих от рекомендуемого уровня самого
// глубокого этажа на TrainingGap и больше.
func (m *Model) weakHeroes() []*Hero {
	rec := recommendedLevel(m.MaxFloor)
	gap := max(1, m.Tactics.TrainingGap)
	var out []*Hero
	for _, h := range m.Party {
		if h.IsDead || h.LostInAbyss {
			continue
		}
		if rec-h.Level >= gap {
			out = append(out, h)
		}
	}
	return out
}

// trainingFloors — пройденные этажи, на которые можно вернуться за опытом.
func (m *Model) trainingFloors() []int {
	var out []int
	for f := 1; f < m.MaxFloor; f++ {
		out = append(out, f)
	}
	return out
}

// departurePlan — решение о том, куда отправляется отряд из города.
type departurePlan struct {
	Floor    int
	Training bool
	Reason   string // ключ словаря с пояснением
	Args     []any
}

// planDeparture решает за автопилота: продолжить спуск или тренироваться.
// Тренировка нужна, если отстают двое и больше, либо отстаёт хотя бы один и
// предыдущие вылазки заканчивались отступлением без продвижения (петля неудач).
func (m *Model) planDeparture() departurePlan {
	m.syncMaxFloor()
	deepest := max(1, m.MaxFloor)
	cont := departurePlan{Floor: deepest, Reason: "gate.reason.continue", Args: []any{deepest}}

	if m.Tactics.DisableTraining {
		return departurePlan{Floor: deepest, Reason: "gate.reason.disabled", Args: []any{deepest}}
	}
	floors := m.trainingFloors()
	if len(floors) == 0 {
		return cont
	}

	weak := m.weakHeroes()
	if len(weak) < 2 && !(len(weak) >= 1 && m.FailStreak >= 2) {
		return cont
	}

	lowest := weak[0].Level
	for _, h := range weak {
		lowest = min(lowest, h.Level)
	}
	target := clampInt(2*lowest+1, 1, floors[len(floors)-1])

	if len(weak) < 2 {
		return departurePlan{Floor: target, Training: true, Reason: "gate.reason.streak", Args: []any{lowest, target}}
	}
	return departurePlan{Floor: target, Training: true, Reason: "gate.reason.weak", Args: []any{len(weak), lowest, target}}
}

// trainingFinished — тренировка выполнила свою задачу: слабых бойцов не осталось
// или на этаже не осталось живых стай.
func (m *Model) trainingFinished() bool {
	if !m.Training {
		return false
	}
	if len(m.weakHeroes()) == 0 {
		return true
	}
	for _, p := range m.Packs {
		if p.LivingCount() > 0 {
			return false
		}
	}
	return true
}

// spawnSmallPack создаёт небольшую стаю, подходящую этажу: без мини-боссов и не больше
// trainingMaxPack монстров.
func spawnSmallPack(floor int) *MonsterPack {
	var p *MonsterPack
	for i := 0; i < 8; i++ {
		p = spawnMonsterPack(false, floor)
		if !p.IsMiniBoss {
			break
		}
	}
	if len(p.Members) > trainingMaxPack {
		p.Members = p.Members[:trainingMaxPack]
	}
	return p
}

// repopulateTrainingFloor заселяет текущий этаж небольшими стаями до целевого числа.
// Стаи не появляются рядом с отрядом и рядом друг с другом; при достаточной
// разведке этажа выбираются уже открытые клетки, чтобы автопилот видел цели.
func (m *Model) repopulateTrainingFloor() {
	if len(m.Grid) == 0 {
		return
	}

	living := 0
	for _, p := range m.Packs {
		if p.LivingCount() > 0 {
			living++
		}
	}

	var all, seen []Point
	for y := 1; y < m.MapHeight-1; y++ {
		for x := 1; x < m.MapWidth-1; x++ {
			if m.Grid[y][x] != TileFloor {
				continue
			}
			p := Point{x, y}
			all = append(all, p)
			if m.Explored[y][x] {
				seen = append(seen, p)
			}
		}
	}
	if len(all) == 0 {
		return
	}
	pool := seen
	if len(seen) < len(all)/3 {
		pool = all
	}

	want := clampInt(len(all)/trainingCellsPer, trainingMinPacks, trainingMaxPacks)
	need := want - living
	if need <= 0 {
		return
	}

	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	for _, p := range pool {
		if need == 0 {
			break
		}
		if absInt(p.X-m.PartyPos.X)+absInt(p.Y-m.PartyPos.Y) < trainingSafeDist {
			continue
		}
		if _, taken := m.Packs[p]; taken {
			continue
		}
		tooClose := false
		for q := range m.Packs {
			if absInt(q.X-p.X)+absInt(q.Y-p.Y) < trainingPackGap {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}
		m.Packs[p] = spawnSmallPack(m.Floor)
		need--
	}
}

// noteTripResult учитывает итог вылазки: череда отступлений без продвижения копит
// «счётчик неудач», который подталкивает автопилот к тренировке.
func (m *Model) noteTripResult(reason RetreatReason) {
	switch {
	case m.Training:
		if reason == RetreatTrainingDone {
			m.FailStreak = 0
		}
	case m.MaxFloor > m.TripStartMaxFloor:
		m.FailStreak = 0
	case reason == RetreatLowHP || reason == RetreatTooFewAlive || reason == RetreatNoResources:
		m.FailStreak++
	}
}
