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
	StateCreator
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
	VisitedFloors    map[int]*FloorState // Кэш посещённых этажей подземелья
	PartyPos         Point
	CameraPos        Point // Плавная позиция центра камеры (устранение тряски экрана)
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
	MarketState      MarketServiceState
	MagistrateState  MagistrateServiceState
	ChurchState      ChurchServiceState
	TavernState      TavernServiceState
	GuildState       GuildServiceState
	SmithyState      ForgeServiceState
	TanneryState     ForgeServiceState
	AlchemistState   AlchemistServiceState

	// Герой-лидер, созданный игроком (Blueprint == nil — отряд случайный), экран создания и меню.
	Blueprint         *HeroBlueprint
	Creator           CreatorState
	MenuGen           int  // поколение таймера меню: устаревшие тики игнорируются
	MenuConfirmRemove bool // в меню нажат X: ждём подтверждения удаления героя
}

type (
	// TickMsg несёт номер поколения цепочки тиков (см. Model.restartTicks).
	TickMsg        struct{ Gen int }
	RestartTickMsg time.Time
	// MenuTickMsg несёт номер поколения таймера меню (см. Model.MenuGen).
	MenuTickMsg struct{ Gen int }
)

func initialModelWithLegacy(legacy TownLegacy) Model {
	return initialModelWith(legacy, nil)
}

// initialModelWith создаёт модель новой экспедиции; с чертежом отряд ведёт созданный игроком лидер.
func initialModelWith(legacy TownLegacy, bp *HeroBlueprint) Model {
	heroes := buildStartingParty(legacy.SmithyLevel, bp)

	activeRelic := generateRelic(1)

	m := Model{
		Lang:             LangRU,
		State:            StateMenu,
		MenuCountdown:    menuCountdownStart,
		Blueprint:        bp,
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
		TownDelayMs:      500,
		StatsScroll:      0,
		LogScroll:        0,
		CodexTab:         0,
		RestartCountdown: 10,
		CurrentQuest:     generateAutoQuest(1),
		Relic:            &activeRelic,
		Legacy:           legacy,
		Stats:            newStats(),
		Packs:            make(map[Point]*MonsterPack),
		VisitedFloors:    make(map[int]*FloorState),
		TermWidth:        120,
		TermHeight:       36,
		RunCounted:       false,
		RestTurnsLeft:    0,
		Tactics:          DefaultTactics(),
	}
	m.relocalizeStart()
	m.Stats.TotalGoldEarned = 50 + legacy.TreasuryGold
	m.initDungeonForFloor(1)
	m.CameraPos = m.PartyPos
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
	return tea.Batch(tea.EnterAltScreen, menuTickCmd(m.MenuGen))
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
	persistState(m.Lang, newLegacy, m.Tactics, m.Blueprint)

	fresh := initialModelWith(newLegacy, m.Blueprint)
	fresh.MenuGen = m.MenuGen + 1
	fresh.Lang = m.Lang
	fresh.Tactics = m.Tactics
	fresh.ManualMode = m.ManualMode
	fresh.TickGen = m.TickGen + 1
	fresh.TermWidth = m.TermWidth
	fresh.TermHeight = m.TermHeight
	fresh.relocalizeStart()

	// Новая экспедиция начинается с меню-заставки; цепочку игровых тиков запустит выход из меню.
	return fresh, menuTickCmd(fresh.MenuGen)
}

func (m Model) resetGame() (Model, tea.Cmd) {
	return resetGameStatic(m)
}
