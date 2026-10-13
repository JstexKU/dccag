package main

// ============================================================
// ХОДЬБА ПО ГОРОДУ И ВИЗИТЫ В ЗАВЕДЕНИЯ
// ============================================================
//
// Отряд ходит по карте города. Автопилот обходит заведения по кругу в порядке
// конвейера (рынок → магистрат → храм → таверна → гильдия → кузница → дубильня →
// алхимик → Врата), пропуская те, где делать нечего. У двери отряд входит: показывается
// карточка заведения; когда дела закончены, отряд выходит на улицу и идёт к следующему.
// В ручном режиме игрок ходит стрелками, а входит, ступив на дверь.

// townNextPhase возвращает фазу, следующую за p в порядке обхода.
func townNextPhase(p TownPhase) TownPhase {
	for i, s := range townSpecs {
		if s.Phase == p && i+1 < len(townSpecs) {
			return townSpecs[i+1].Phase
		}
	}
	return TownPhaseDepart
}

// townNeedsVisit — есть ли смысл идти в заведение. Проверка осторожная: если сомневаемся, идём.
func (m *Model) townNeedsVisit(p TownPhase) bool {
	switch p {
	case TownPhaseMagistrate:
		budget := AllocateBudget(m.Gold).Treasury
		for _, off := range m.GenerateMagistrateOffers() {
			if off.CanAfford && budget >= off.Cost && off.Target != MagistrateTreasuryGrant {
				return true
			}
		}
		return false

	case TownPhaseChurch:
		for _, h := range m.Party {
			if h.IsDead {
				continue
			}
			if h.IsDowned || h.Stress > 20 || h.Affliction != AfflictionNone {
				return true
			}
		}
		return false

	case TownPhaseTavern:
		for _, h := range m.Party {
			if !h.IsDead && !h.IsDowned && (h.HP < h.MaxHP || h.MP < h.MaxMP) {
				return true
			}
		}
		return false

	case TownPhaseGuild:
		if m.CurrentQuest.Completed {
			return true
		}
		for _, h := range m.Party {
			if h.IsDead || h.IsDowned {
				return true
			}
		}
		return false

	case TownPhaseSmithy:
		budget := AllocateBudget(m.Gold).Forge
		for _, off := range m.GenerateSmithyOffers() {
			if off.CanAfford && budget >= off.Cost {
				return true
			}
		}
		return false

	case TownPhaseTannery:
		// Если можно купить сумку — посещаем обязательно
		if m.BagLevel < len(bagUpgrades)-1 {
			next := bagUpgrades[m.BagLevel+1]
			maxAllowedTier := m.Legacy.TanneryLevel + 1
			if next.Level <= maxAllowedTier && m.Gold >= next.Cost {
				return true
			}
		}
		b := AllocateBudget(m.Gold)
		tanneryBudget := b.Tannery + b.Bags
		for _, off := range m.GenerateTanneryOffers() {
			if off.CanAfford && tanneryBudget >= off.Cost {
				return true
			}
		}
		return false

	case TownPhaseAlchemist:
		budget := AllocateBudget(m.Gold).Alchemy
		for _, off := range m.GenerateAlchemistOffers() {
			if off.CanAfford && budget >= off.Cost {
				return true
			}
		}
		return false
	}
	return true // рынок и Врата посещаются всегда
}

// nextWorthPhase пропускает заведения, в которых сейчас нечего делать, и возвращает первое
// нужное (последним всегда идут Врата).
func (m *Model) nextWorthPhase(from TownPhase) TownPhase {
	p := from
	for p != TownPhaseDepart && !m.townNeedsVisit(p) {
		p = townNextPhase(p)
	}
	return p
}

// enterTown — отряд приходит в город из подземелья (выход, отступление, бегство).
// Этаж запоминается, итог вылазки учитывается, тренировочный выход возвращает этаж прогресса.
func (m *Model) enterTown(reason RetreatReason) {
	m.saveCurrentFloorState()
	m.noteTripResult(reason)
	if m.Training {
		m.Training = false
		m.Floor = max(1, m.MaxFloor)
	}

	m.InTown = true
	m.PathHistory = []Point{}
	m.LoopDetectCount = 0
	m.TownDialog = T(m.Lang, "town.log.enter_gate")
	if reason == RetreatTrainingDone {
		m.TownDialog = T(m.Lang, "gate.log.training_done")
	}

	m.TownVisit, m.TownLeaving, m.TownDwell = false, false, 0
	m.TownPhase = TownPhaseSellLoot
	m.TownWalking = false
	if m.Town != nil {
		m.TownWalking = true
		m.TownPos = m.Town.Buildings[m.Town.Gate].Front
		if !m.ManualMode {
			m.TownPhase = m.nextWorthPhase(TownPhaseSellLoot)
		}
	}
}

// initTownScreen готовит карточку заведения.
func (m *Model) initTownScreen(p TownPhase) {
	switch p {
	case TownPhaseSellLoot:
		m.initMarketScreen()
	case TownPhaseMagistrate:
		m.initMagistrateScreen()
	case TownPhaseChurch:
		m.initChurchScreen()
	case TownPhaseTavern:
		m.initTavernScreen()
	case TownPhaseGuild:
		m.initGuildScreen()
	case TownPhaseSmithy:
		m.initSmithyScreen()
	case TownPhaseTannery:
		m.initTanneryScreen()
	case TownPhaseAlchemist:
		m.initAlchemistScreen()
	case TownPhaseDepart:
		m.initGateScreen()
	}
}

// enterBuilding — отряд входит в здание: показывается карточка заведения.
func (m *Model) enterBuilding(idx int) {
	b := m.Town.Buildings[idx]
	m.TownPos = b.Door
	m.TownWalking = false
	m.TownVisit = true
	m.TownLeaving = false
	m.TownPhase = b.Phase
	m.initTownScreen(b.Phase)
	m.TownDwell = 1 // карточка успевает показаться до первого действия
	if b.Phase == TownPhaseDepart {
		m.TownDwell = 2
	}
}

// leaveBuilding — отряд выходит на улицу. done == true: заведение закончено и автопилот
// выбирает следующее; иначе (игрок вышел сам) фаза не меняется.
func (m *Model) leaveBuilding(done bool) {
	idx := -1
	if m.Town != nil {
		idx = m.Town.buildingFor(m.TownPhase)
	}
	m.TownVisit, m.TownLeaving, m.TownDwell = false, false, 0
	m.TownWalking = m.Town != nil
	if idx >= 0 {
		m.TownPos = m.Town.Buildings[idx].Front
	}
	if done && !m.ManualMode {
		m.TownPhase = m.nextWorthPhase(m.TownNextPhase)
	}
}

// stepTownMap — один тик города. Состояния без карты (старые тесты) идут прежним конвейером.
func (m *Model) stepTownMap() {
	switch {
	case m.TownWalking:
		m.stepTownWalk()
	case m.TownVisit:
		m.stepTownVisit()
	default:
		m.stepTown()
	}
}

// stepTownWalk — автопилот шагает к двери следующего заведения.
func (m *Model) stepTownWalk() {
	t := m.Town
	if t == nil {
		m.TownWalking = false
		return
	}
	if m.ManualMode {
		return
	}
	idx := t.buildingFor(m.TownPhase)
	if idx < 0 {
		m.TownWalking = false // нет здания для фазы — работаем по-старому
		return
	}
	door := t.Buildings[idx].Door
	if m.TownPos == door {
		m.enterBuilding(idx)
		return
	}
	next, ok := t.pathStep(m.TownPos, door)
	if !ok {
		next = door // страховка от тупика на карте
	}
	m.TownPos = next
	if next == door {
		m.enterBuilding(idx)
	}
}

// stepTownVisit — автопилот внутри заведения: действия, итог на карточке, выход.
func (m *Model) stepTownVisit() {
	if m.ManualMode {
		return
	}
	if m.TownDwell > 0 {
		m.TownDwell--
		return
	}
	if m.TownLeaving {
		m.leaveBuilding(true)
		return
	}

	phase := m.TownPhase
	if phase == TownPhaseDepart {
		m.stepGateAuto()
		return
	}
	m.stepTown()
	if m.TownPhase != phase {
		// Заведение закончило дела: итог остаётся на карточке ещё на тик, затем выход.
		m.TownNextPhase = m.TownPhase
		m.TownPhase = phase
		m.TownLeaving = true
		m.TownDwell = 1
	}
}

// townManualMove — шаг игрока по улицам; дверь здания открывает его карточку.
func (m *Model) townManualMove(dx, dy int) {
	t := m.Town
	if t == nil || !m.TownWalking {
		return
	}
	next := Point{m.TownPos.X + dx, m.TownPos.Y + dy}
	if !t.walkable(next) {
		return
	}
	m.TownPos = next
	if idx := t.doorAt(next); idx >= 0 {
		m.enterBuilding(idx)
	}
}

// handleTownWalkKey — клавиши ручного режима на улицах города. Второе значение: «обработана».
func (m *Model) handleTownWalkKey(key string) bool {
	switch key {
	case "up":
		m.townManualMove(0, -1)
	case "down":
		m.townManualMove(0, 1)
	case "left":
		m.townManualMove(-1, 0)
	case "right":
		m.townManualMove(1, 0)
	default:
		return false
	}
	return true
}

// tickDelay — пауза до следующего тика: по улицам отряд идёт быстрее, чем выполняются
// действия в заведениях.
func (m Model) tickDelay() int {
	if m.InTown {
		if m.TownWalking {
			return max(30, m.SpeedMs/3)
		}
		return m.TownDelayMs
	}
	return m.SpeedMs
}