package main

import (
	"fmt"
)

// Пороги детектора зацикливания отряда.
const (
	loopForceThreshold = 5  // с этого значения отряд идёт аварийным маршрутом
	loopAbortThreshold = 10 // с этого значения отряд принудительно покидает этаж
	pathHistoryMax     = 36 // сколько последних клеток помнит детектор
)

// isAutoTraining — отряд тренируется на пройденном этаже в автопилоте.
func (m *Model) isAutoTraining() bool {
	return m.Training && !m.ManualMode
}

// step — один шаг автопилота. Этапы идут строго по приоритету; этап возвращает true,
// если он занял этот шаг и дальше идти не нужно.
func (m *Model) step() {
	if m.State != StatePlaying {
		return
	}
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.RestartCountdown = 10
		return
	}

	// Город — отдельный поток, шаги подземелья не считаются
	if m.InTown {
		m.stepTownMap()
		return
	}

	m.Stats.TotalSteps++

	if m.stepRest() || m.stepCombat() || m.stepRetreat() || m.stepCamp() {
		return
	}
	m.stepPotions()
	m.stepMove()
}

// stepRest — продолжение привала (RestTurnsLeft > 0): сидим на месте и лечимся.
func (m *Model) stepRest() bool {
	if m.RestTurnsLeft <= 0 {
		return false
	}
	m.RestTurnsLeft--
	recoveryStep := 1
	// Реликвия «Рог Изобилия»: ускоренное восстановление на привале
	if m.Relic != nil && m.Relic.PotionPotency > 1.0 {
		recoveryStep = 2
	}

	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		if h.HP < h.MaxHP {
			h.HP = min(h.MaxHP, h.HP+recoveryStep)
		}
		if h.MP < h.MaxMP {
			h.MP = min(h.MaxMP, h.MP+recoveryStep)
		}
		if h.Stress > 0 {
			h.Stress = max(0, h.Stress-recoveryStep)
		}
	}
	m.stepPacks() // пока отряд отдыхает, стаи не дремлют
	return true
}

// stepCombat — идущий бой важнее всего остального.
func (m *Model) stepCombat() bool {
	if m.Combat == nil {
		return false
	}
	m.executeCombatTurn()
	return true
}

// stepRetreat — если отряд стоит на выходе и есть причина отступить, уходим в город.
func (m *Model) stepRetreat() bool {
	reason := m.evaluateRetreat()
	if reason == RetreatNone ||
		m.Grid[m.PartyPos.Y][m.PartyPos.X] != TileExit ||
		m.Stats.TotalSteps <= 0 {
		return false
	}
	m.enterTown(reason)
	return true
}

// partyNeedsRest — привал нужен только при реальной необходимости: тяжёлые ранения,
// нехватка маны у кастеров или сильный стресс.
func (m *Model) partyNeedsRest() bool {
	for _, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		hpPct := float64(h.HP) / float64(h.MaxHP)
		mpPct := 1.0
		if h.MaxMP > 0 {
			mpPct = float64(h.MP) / float64(h.MaxMP)
		}
		isCaster := (h.Class == ClassMage || h.Class == ClassCleric || h.Class == ClassWarlock || h.Class == ClassBard || h.Class == ClassPaladin)

		if hpPct <= 0.45 || (isCaster && mpPct <= 0.25) || h.Stress >= 80 {
			return true
		}
	}
	return false
}

// stepCamp — ставим привал, если он нужен и вокруг тихо.
func (m *Model) stepCamp() bool {
	if !m.partyNeedsRest() ||
		m.Stats.TotalSteps%30 != 0 ||
		m.evaluateRetreat() != RetreatNone ||
		m.isMonsterNearby() {
		return false
	}

	m.RestTurnsLeft = rng.Intn(5) + 4 // 4..8 шагов
	biome := getBiome(m.Floor)
	variant := rng.Intn(4) + 1
	key := fmt.Sprintf("dungeon.log.rest.%s.%d", biome.Name, variant)
	m.addLog(healStyle.Render(T(m.Lang, key)))
	return true
}

// stepPotions — каждый герой при необходимости пьёт зелье.
func (m *Model) stepPotions() {
	for _, h := range m.Party {
		m.checkAndDrinkPotions(h)
	}
}

// stepMove — выбор следующей клетки, защита от зацикливания и сам шаг.
func (m *Model) stepMove() {
	var next Point

	if m.LoopDetectCount >= loopForceThreshold {
		targetTile := TileStairs
		if m.evaluateRetreat() != RetreatNone || m.isAutoTraining() {
			targetTile = TileExit
		}
		forcedStep, found := m.findEmergencyStep(targetTile)
		if !found {
			m.breakLoop()
			return
		}
		next = forcedStep
	} else {
		next = m.findNextStep()
	}

	m.trackLoop(next)
	if m.LoopDetectCount >= loopAbortThreshold {
		m.breakLoop()
		return
	}

	m.PathHistory = append(m.PathHistory, next)
	if len(m.PathHistory) > pathHistoryMax {
		m.PathHistory = m.PathHistory[1:]
	}

	m.moveTo(next)
	m.stepPacks()
}

// trackLoop обновляет счётчик зацикливания по выбранной клетке.
func (m *Model) trackLoop(next Point) {
	isOscillating := false
	histLen := len(m.PathHistory)
	if histLen >= 2 && m.PathHistory[histLen-2] == next {
		isOscillating = true
	}

	visitCount := 0
	for _, p := range m.PathHistory {
		if p == next {
			visitCount++
		}
	}

	if isOscillating || visitCount >= 3 || next == m.PartyPos {
		m.LoopDetectCount++
	} else if m.LoopDetectCount > 0 && m.LoopDetectCount < loopForceThreshold {
		m.LoopDetectCount--
	}

	if m.LoopDetectCount == loopForceThreshold {
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.collision_break")))
	}
}

// breakLoop — выход из тупика: тренировка завершается, иначе отряд уходит на этаж ниже.
func (m *Model) breakLoop() {
	if m.isAutoTraining() {
		m.enterTown(RetreatTrainingDone)
		return
	}
	m.descend("")
}

// descend переводит отряд на следующий этаж. normalKey — ключ записи в журнале
// (для каждого 10-го этажа вместо него пишется запись о боссе); пустой ключ — без записи.
func (m *Model) descend(normalKey string) {
	m.saveCurrentFloorState()
	m.Floor++
	m.Stats.FloorsCleared++
	m.checkQuestProgress(QuestReachFloor, "", m.Floor)
	m.checkQuestProgress(QuestEscapeTrap, "", 1)

	if normalKey != "" {
		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(T(m.Lang, "dungeon.log.floor_cleared_boss", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(T(m.Lang, normalKey, m.Floor)))
		}
	}

	m.PathHistory = []Point{}
	m.LoopDetectCount = 0
	m.initDungeonForFloor(m.Floor)
}

// tileHandler — что происходит, когда отряд наступает на плитку. Возвращает true,
// если отряд ушёл с этажа и ставить его на плитку не нужно.
type tileHandler func(m *Model, at Point) (leftFloor bool)

// tileHandlers — обработчики плиток. Новая плитка — одна запись здесь и один метод.
// Заполняется в init(), чтобы таблица не образовала цикл инициализации с step().
var tileHandlers map[Tile]tileHandler

func init() {
	tileHandlers = map[Tile]tileHandler{
		TileChest:        (*Model).walkOnChest,
		TileTrappedChest: (*Model).walkOnTrappedChest,
		TileRelic:        consumeTile((*Model).handleRelicTile),
		TileAltar:        consumeTile((*Model).handleAltar),
		TileFountain:     consumeTile((*Model).handleFountain),
		TileLore:         consumeTile((*Model).handleLoreTile),
		TileBarrel:       (*Model).walkOnBarrel,
		TileEvent:        (*Model).walkOnEvent,
		TileStairs:       (*Model).walkOnStairs,
	}
}

// consumeTile — обработчик «выполнить действие, затем плитка становится полом».
func consumeTile(action func(*Model)) tileHandler {
	return func(m *Model, at Point) bool {
		action(m)
		m.Grid[at.Y][at.X] = TileFloor
		return false
	}
}

func (m *Model) bagFull() bool {
	return len(m.Bag) >= m.currentBagCapacity()
}

func (m *Model) walkOnChest(at Point) bool {
	if m.bagFull() {
		m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full_skip")))
		return false
	}
	m.Stats.ChestsOpened++
	m.checkQuestProgress(QuestOpenChests, "", 1)
	gold := int(float64(rng.Intn(16)+10+(m.Floor*2)) * m.goldMult())
	m.Gold += gold
	m.Stats.TotalGoldEarned += gold

	// Квест «Золотая лихорадка»: учёт золота из сундука
	m.checkQuestProgress(QuestCollectGold, "", gold)

	targetClass := ClassWarrior
	if lh := m.getRandomLivingHero(); lh != nil {
		targetClass = lh.Class
		lh.AddTreasure()
		m.checkAndAwardTitle(lh)
	}
	item := generateItemForClass(targetClass, m.Floor)
	m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.chest_open", gold, item.DisplayName(m.Lang))))
	m.equipOrBag(item)
	m.Grid[at.Y][at.X] = TileFloor
	return false
}

func (m *Model) walkOnTrappedChest(at Point) bool {
	if m.bagFull() {
		m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full_skip")))
		return false
	}
	m.handleTrappedChest()
	m.Grid[at.Y][at.X] = TileFloor
	return false
}

// walkOnBarrel — квест «Огненный шквал»: разрушение/подрыв бочки при наступлении.
func (m *Model) walkOnBarrel(at Point) bool {
	m.Stats.BarrelsBlown++
	m.checkQuestProgress(QuestPyromania, "", 1)
	m.Grid[at.Y][at.X] = TileFloor
	return false
}

// walkOnEvent — плитка очищается до обработки: событие может изменить этаж.
func (m *Model) walkOnEvent(at Point) bool {
	m.Grid[at.Y][at.X] = TileFloor
	m.handleEventTile(at)
	return false
}

func (m *Model) walkOnStairs(at Point) bool {
	if m.isAutoTraining() {
		return false
	}

	// Блокировка: если цель — полная зачистка, а монстры ещё живы, не спускаемся
	if m.CurrentQuest.Type == QuestPurgeBiome && !m.CurrentQuest.Completed && len(m.Packs) > 0 {
		return false
	}

	// Квест «Полная зачистка»: если на этаже не осталось живых стай монстров
	if len(m.Packs) == 0 {
		m.checkQuestProgress(QuestPurgeBiome, "", 1)
	}

	// Квест «Топография Бездны»: если разведано >= 90% территории
	if m.calcExploredPercent() >= 90 {
		m.checkQuestProgress(QuestFullMap, "", 1)
	}

	m.descend("dungeon.log.stairs_descend")
	return true
}

func (m *Model) moveTo(next Point) {
	if pack, exists := m.Packs[next]; exists {
		m.startCombat(next, pack)
		return
	}

	if next == m.PartyPos {
		if m.isAutoTraining() {
			m.enterTown(RetreatTrainingDone)
			return
		}
		m.descend("dungeon.log.floor_cleared")
		return
	}

	if handler, ok := tileHandlers[m.Grid[next.Y][next.X]]; ok && handler(m, next) {
		return
	}

	m.PartyPos = next
	m.revealFog()
}

func (m *Model) calcExploredPercent() int {
	if len(m.Grid) == 0 || len(m.Grid[0]) == 0 {
		return 0
	}
	total := 0
	explored := 0
	for y := 0; y < len(m.Grid); y++ {
		for x := 0; x < len(m.Grid[y]); x++ {
			if m.Grid[y][x] != TileWall {
				total++
				if len(m.Explored) > y && len(m.Explored[y]) > x && m.Explored[y][x] {
					explored++
				}
			}
		}
	}
	if total == 0 {
		return 0
	}
	return (explored * 100) / total
}