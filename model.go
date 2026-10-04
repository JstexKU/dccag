package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateDefeat
	StateStatsManual
	StateArmory
	StateInfoBook
	StateTactics
)

type TownPhase int

const (
	TownPhaseSellLoot TownPhase = iota
	TownPhaseMagistrate
	TownPhaseChurch
	TownPhaseTavern
	TownPhaseGuild
	TownPhaseSmithy
	TownPhaseTannery
	TownPhaseAlchemist
	TownPhaseDepart
)

type Model struct {
	Lang             Language
	State            GameState
	MenuCountdown    int
	TermWidth        int
	TermHeight       int
	MapWidth         int
	MapHeight        int
	Grid             [][]Tile
	Explored         [][]bool
	Packs            map[Point]*MonsterPack
	PartyPos         Point
	PathHistory      []Point
	LoopDetectCount  int
	Party            []*Hero
	Bag              []EquipItem
	BagLevel         int
	Gold             int
	Floor            int
	InTown           bool
	TownPhase        TownPhase
	TownDialog       string
	TownHistory      []string
	TownEst          TownEstablishments
	Combat           *ActiveCombat
	Logs             []string
	AutoMode         bool
	SpeedMs          int
	TownDelayMs      int
	StatsScroll      int
	LogScroll        int
	CodexTab         int
	RestartCountdown int
	CurrentQuest     AutoQuest
	Relic            *PartyRelic
	Legacy           TownLegacy
	Stats            RunStats
	RunCounted       bool // защита от двойного учёта рана в accumulateDebugReport
	RestTurnsLeft    int  // сколько шагов отряд ещё сидит на привале (0 — не на привале)
	TickGen          int  // поколение цепочки тиков: устаревшие тики игнорируются
	Tactics          Tactics
	TacticsSel       int
	ManualMode       bool // ручное управление: игрок ходит по карте и командует в бою
	ManualLastCamp   int  // Stats.TotalSteps на момент последнего ручного привала
}

type (
	// TickMsg несёт номер поколения цепочки тиков (см. Model.restartTicks).
	TickMsg        struct{ Gen int }
	RestartTickMsg time.Time
	MenuTickMsg    time.Time
)

func initialModelWithLegacy(legacy TownLegacy) Model {
	// rand.Seed убран: начиная с Go 1.20 глобальный math/rand автосидируется.

	// Перемешиваем все 10 классов и отбираем ровно 5 уникальных
	shuffledClasses := make([]HeroClass, len(AllClasses))
	copy(shuffledClasses, AllClasses)
	rng.Shuffle(len(shuffledClasses), func(i, j int) {
		shuffledClasses[i], shuffledClasses[j] = shuffledClasses[j], shuffledClasses[i]
	})

	var heroes []*Hero
	for _, c := range shuffledClasses[:5] {
		heroes = append(heroes, createHero(c, 1, legacy.SmithyLevel))
	}

	activeRelic := generateRelic(1)

	m := Model{
		Lang:             LangRU,
		State:            StateMenu,
		MenuCountdown:    20,
		Party:            heroes,
		Bag:              []EquipItem{},
		BagLevel:         0,
		Gold:             50 + legacy.TreasuryGold,
		Floor:            1,
		InTown:           false,
		TownPhase:        TownPhaseSellLoot,
		TownDialog:       "",
		TownHistory:      []string{},
		TownEst:          generateTownEstablishments(),
		Combat:           nil,
		Logs:             []string{},
		AutoMode:         true,
		SpeedMs:          260,
		TownDelayMs:      2200,
		StatsScroll:      0,
		LogScroll:        0,
		CodexTab:         0,
		RestartCountdown: 10,
		CurrentQuest:     generateAutoQuest(1),
		Relic:            &activeRelic,
		Legacy:           legacy,
		Stats:            newStats(),
		Packs:            make(map[Point]*MonsterPack),
		TermWidth:        120,
		TermHeight:       36,
		RunCounted:       false,
		RestTurnsLeft:    0,
		Tactics:          DefaultTactics(),
	}
	m.relocalizeStart()
	m.Stats.TotalGoldEarned = 50 + legacy.TreasuryGold
	m.initDungeonForFloor(1)
	return m
}

// relocalizeStart обновляет стартовые тексты под текущий язык (нужно после смены Lang).
func (m *Model) relocalizeStart() {
	m.TownDialog = T(m.Lang, "town.log.enter_gate")
	m.Logs = []string{T(m.Lang, "dungeon.log.start")}
}

func initialModel() Model {
	return initialModelWithLegacy(TownLegacy{TreasuryGold: 0, SmithyLevel: 0, TanneryLevel: 0, ChurchLevel: 0, TavernLevel: 0})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, menuTickCmd())
}

func (m *Model) addLog(msg string) {
	m.Logs = append(m.Logs, msg)
	if len(m.Logs) > 100 {
		m.Logs = m.Logs[len(m.Logs)-100:]
	}
	m.LogScroll = 0
}

func resetGameStatic(m Model) (Model, tea.Cmd) {
	m.accumulateDebugReport()
	saveDebugReportToFile()

	// Налог с добычи уходит в казну столицы и сохраняется на диск.
	newLegacy := legacyAfterRun(m)
	persistState(m.Lang, newLegacy, m.Tactics)

	fresh := initialModelWithLegacy(newLegacy)
	fresh.Lang = m.Lang
	fresh.Tactics = m.Tactics
	fresh.ManualMode = m.ManualMode
	fresh.TickGen = m.TickGen + 1
	fresh.TermWidth = m.TermWidth
	fresh.TermHeight = m.TermHeight
	fresh.relocalizeStart()

	// Новая экспедиция начинается с меню-заставки; цепочку игровых тиков запустит выход из меню.
	return fresh, menuTickCmd()
}

func (m Model) resetGame() (Model, tea.Cmd) {
	return resetGameStatic(m)
}
