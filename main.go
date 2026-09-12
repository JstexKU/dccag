package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) accumulateDebugReport() {
	globalDebugReport.TotalRuns++
	if m.Floor > globalDebugReport.MaxFloorReached {
		globalDebugReport.MaxFloorReached = m.Floor
	}

	globalDebugReport.TotalGoldSpent += (m.Stats.TotalGoldEarned - m.Gold)

	for _, fallen := range m.Stats.FallenHeroes {
		globalDebugReport.DeathCauses[fallen.Cause]++
		globalDebugReport.ClassDeaths[string(fallen.Class)]++
		globalDebugReport.DeathsByFloor[fallen.Floor]++
	}

	for _, h := range m.Party {
		if !h.IsDead {
			globalDebugReport.AverageAtk[string(h.Class)] = float64(h.TotalAtk())
		}
	}
}

func saveDebugReportToFile() {
	if !*debugReportFlag {
		return
	}
	fileData, err := json.MarshalIndent(globalDebugReport, "", "  ")
	if err == nil {
		_ = os.WriteFile("dccag_report.json", fileData, 0o644)
		fmt.Println("\n[DEBUG] Подробный отчет телеметрии сохранен в файл: dccag_report.json")
	}
}

func tickCmd(speedMs int) tea.Cmd {
	return tea.Tick(time.Duration(speedMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func restartTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return RestartTickMsg(t)
	})
}

func menuTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return MenuTickMsg(t)
	})
}

func createHero(class HeroClass, floor int, smithyLvl int) *Hero {
	targetLevel := max(1, floor/2)

	maxHP := 45
	baseDef := 2
	speed := 10
	skillName := "Удар"
	skillCost := 10
	maxMP := 35

	switch class {
	case ClassTank:
		maxHP = 65
		baseDef = 4
		speed = 8
		skillName = "Стойка"
		skillCost = 8
		maxMP = 30
	case ClassWarrior:
		maxHP = 50
		baseDef = 2
		speed = 10
		skillName = "Ярость"
		skillCost = 10
		maxMP = 25
	case ClassRogue:
		maxHP = 38
		baseDef = 1
		speed = 15
		skillName = "Тень"
		skillCost = 12
		maxMP = 35
	case ClassMage:
		maxHP = 30
		baseDef = 0
		speed = 11
		skillName = "Заряд"
		skillCost = 15
		maxMP = 45
	case ClassCleric:
		maxHP = 36
		baseDef = 2
		speed = 9
		skillName = "Аура"
		skillCost = 12
		maxMP = 40
	}

	nameDef := getRandomHeroName()

	h := &Hero{
		Name:      nameDef.Name,
		Gender:    nameDef.Gender,
		Class:     class,
		Role:      GetClassRole(class),
		Level:     1,
		Exp:       0,
		MaxHP:     maxHP,
		HP:        maxHP,
		MaxMP:     maxMP,
		MP:        maxMP,
		BaseAtk:   6 + smithyLvl,
		BaseDef:   baseDef,
		Speed:     speed,
		SkillName: skillName,
		SkillCost: skillCost,
		IsDead:    false,
		Potions:   []*Potion{},
	}

	for h.Level < targetLevel {
		h.GainExp(h.NextLevelExp())
	}

	itemFloor := max(1, floor)
	w := generateItemForClassSlot(class, SlotWeapon, itemFloor)
	w.UpgradeLevel = smithyLvl
	h.Weapon = &w

	head := generateItemForClassSlot(class, SlotHead, itemFloor)
	h.Head = &head
	ch := generateItemForClassSlot(class, SlotChest, itemFloor)
	h.Chest = &ch
	legs := generateItemForClassSlot(class, SlotLegs, itemFloor)
	h.Legs = &legs

	return h
}

func initialModelWithLegacy(legacy TownLegacy) Model {
	rand.Seed(time.Now().UnixNano())

	classes := []HeroClass{ClassTank, ClassWarrior, ClassRogue, ClassMage, ClassCleric}
	var heroes []*Hero
	for _, c := range classes {
		heroes = append(heroes, createHero(c, 1, legacy.SmithyLevel))
	}

	activeRelic := generateRelic(1)

	m := Model{
		State:            StateMenu,
		MenuCountdown:    20,
		Party:            heroes,
		Bag:              []EquipItem{},
		BagLevel:         0,
		Gold:             50 + legacy.TreasuryGold,
		Floor:            1,
		InTown:           false,
		TownPhase:        TownPhaseSellLoot,
		TownDialog:       "Отряд вошел через ворота Столицы на привал.",
		TownHistory:      []string{},
		TownEst:          generateTownEstablishments(),
		Combat:           nil,
		Logs:             []string{"[Хроники] Отряд ступил во мрак подземелья dccag."},
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
	}
	m.Stats.TotalGoldEarned = 50 + legacy.TreasuryGold
	m.initDungeonForFloor(1)
	return m
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

func (m *Model) distributePartyExp(expAmt int) {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return
	}

	expPerHero := expAmt / len(living)
	if expPerHero < 1 {
		expPerHero = 1
	}

	for _, h := range living {
		if h.GainExp(expPerHero) {
			m.addLog(healStyle.Render(fmt.Sprintf("⭐ [УРОВЕНЬ] %s %s Ур.%d! Характеристики возросли!",
				h.Name, h.Verb("достиг", "достигла"), h.Level)))
		}
	}
}

func (m *Model) checkQuestProgress(action QuestType, targetMob MonsterType, val int) {
	if m.CurrentQuest.Completed {
		return
	}
	if m.CurrentQuest.Type == action {
		if action == QuestHuntMonster && m.CurrentQuest.TargetMob != targetMob {
			return
		}
		if action == QuestReachFloor {
			if val >= m.CurrentQuest.TargetCount {
				m.CurrentQuest.Current = m.CurrentQuest.TargetCount
				m.CurrentQuest.Completed = true
				m.addLog(questStyle.Render(fmt.Sprintf("📜 [КОНТРАКТ] %s выполнен!", m.CurrentQuest.Title)))
			}
			return
		}
		m.CurrentQuest.Current += val
		if m.CurrentQuest.Current >= m.CurrentQuest.TargetCount {
			m.CurrentQuest.Current = m.CurrentQuest.TargetCount
			m.CurrentQuest.Completed = true
			m.addLog(questStyle.Render(fmt.Sprintf("📜 [КОНТРАКТ] %s выполнен!", m.CurrentQuest.Title)))
		}
	}
}

func (m *Model) equipOrBag(item EquipItem) {
	var bestHero *Hero
	maxDiff := -99999

	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		if item.AllowedClass != "" && item.AllowedClass != h.Class {
			continue
		}

		currItem := h.GetItemInSlot(item.Slot)
		currStat := 0
		if currItem != nil {
			currStat = currItem.TotalStat()
		}

		diff := item.TotalStat() - currStat
		if diff > maxDiff {
			maxDiff = diff
			bestHero = h
		}
	}

	if bestHero != nil && maxDiff > 0 {
		oldItem := bestHero.GetItemInSlot(item.Slot)
		if oldItem != nil {
			m.Bag = append(m.Bag, *oldItem)
		}
		newItem := item
		bestHero.SetItemInSlot(newItem.Slot, &newItem)
		m.addLog(healStyle.Render(fmt.Sprintf("✨ %s %s [%s] на [%s] (Мощь: %d)!",
			bestHero.Name, bestHero.Verb("сменил", "сменила"), item.Slot, newItem.DisplayName(), newItem.TotalStat())))
	} else {
		m.Bag = append(m.Bag, item)
		m.addLog(subtleStyle.Render(fmt.Sprintf("📦 %s сложен в сумку.", item.DisplayName())))
	}
}

func (m *Model) recordFallenHero(h *Hero) {
	m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
		FullName: h.FullName(),
		Class:    h.Class,
		Cause:    h.CauseOfDeath,
		Floor:    m.Floor,
	})

	h.IsDead = true
	h.IsGuarding = false
	h.IsBerserk = false
	h.IsStealthed = false
	h.IsCharged = false
	h.IsAura = false
	h.Weapon = nil
	h.Head = nil
	h.Chest = nil
	h.Legs = nil
	h.Potions = []*Potion{}

	if m.Relic != nil && m.Relic.MartyrFury {
		m.addLog(fireStyle.Render("👑 [Корона] Ярость павшего усилила живых (+4 Atk)!"))
		for _, ally := range m.Party {
			if !ally.IsDead {
				ally.BaseAtk += 4
			}
		}
	}
}

func (m *Model) step() {
	if m.State != StatePlaying {
		return
	}
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.RestartCountdown = 10
		return
	}

	m.Stats.TotalSteps++

	if m.Stats.TotalSteps > 0 && m.Stats.TotalSteps%25 == 0 {
		healedCount := 0
		for _, h := range m.Party {
			if h.IsDead {
				continue
			}
			changed := false
			if h.HP < h.MaxHP {
				h.HP++
				changed = true
			}
			if h.MP < h.MaxMP {
				h.MP++
				changed = true
			}
			if h.Stress > 0 {
				h.Stress--
				changed = true
			}
			if changed {
				healedCount++
			}
		}
		if healedCount > 0 {
			m.addLog(healStyle.Render("🌿 [Привал] В тишине подземелья отряд немного передохнул (+1 HP/MP, -1 Стресс)."))
		}
	}

	for _, h := range m.Party {
		m.checkAndDrinkPotions(h)
	}

	if m.Combat != nil {
		m.executeCombatTurn()
		return
	}

	if m.InTown {
		m.stepTown()
		return
	}

	if m.checkRetreat() && m.Grid[m.PartyPos.Y][m.PartyPos.X] == TileExit && m.Stats.TotalSteps > 0 {
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = "Отряд вернулся в город."
		return
	}

	next := m.findNextStep()

	loopHit := false
	for _, p := range m.PathHistory {
		if p == next {
			m.LoopDetectCount++
			loopHit = true
			break
		}
	}
	if !loopHit {
		m.LoopDetectCount = 0
	}

	if m.LoopDetectCount > 4 {
		m.addLog(dangerStyle.Render("⚠️ [Коллизия] Экстренный прорыв к свободному залу."))
		foundSafeSpot := false
		for y := 0; y < m.MapHeight && !foundSafeSpot; y++ {
			for x := 0; x < m.MapWidth && !foundSafeSpot; x++ {
				if m.Grid[y][x] == TileFloor || m.Grid[y][x] == TileExit {
					dx := x - m.PartyPos.X
					dy := y - m.PartyPos.Y
					if dx*dx+dy*dy <= 25 && dx*dx+dy*dy > 1 {
						m.PartyPos = Point{x, y}
						foundSafeSpot = true
					}
				}
			}
		}
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.revealFog()
		return
	} else {
		m.PathHistory = append(m.PathHistory, next)
		if len(m.PathHistory) > 10 {
			m.PathHistory = m.PathHistory[1:]
		}
	}

	if pack, exists := m.Packs[next]; exists {
		m.startCombat(next, pack)
		return
	}

	if next == m.PartyPos {
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(fmt.Sprintf("🌟 ЭТАЖ %d ПРОЙДЕН! Бездна зовет...", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(fmt.Sprintf("Этаж %d пройден! Спуск глубже.", m.Floor)))
		}

		m.initDungeonForFloor(m.Floor)
		return
	}

	switch m.Grid[next.Y][next.X] {
	case TileChest:
		m.Stats.ChestsOpened++
		m.checkQuestProgress(QuestOpenChests, "", 1)
		gold := int(float64(rand.Intn(16)+10+(m.Floor*2)) * m.Relic.GoldMult * 2)
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			m.checkAndAwardTitle(lh)
		}
		item := generateItemForClass(targetClass, m.Floor)
		m.addLog(goldStyle.Render(fmt.Sprintf("🎁 Сундук: +%dG и [%s].", gold, item.DisplayName())))
		m.equipOrBag(item)
		m.Grid[next.Y][next.X] = TileFloor

	case TileRelic:
		m.handleRelicTile()
		m.Grid[next.Y][next.X] = TileFloor

	case TileAltar:
		m.handleAltar()
		m.Grid[next.Y][next.X] = TileFloor

	case TileFountain:
		m.handleFountain()
		m.Grid[next.Y][next.X] = TileFloor

	case TileTrappedChest:
		m.handleTrappedChest()
		m.Grid[next.Y][next.X] = TileFloor

	case TileBarrel:
		m.Grid[next.Y][next.X] = TileFloor

	case TileStairs:
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(fmt.Sprintf("🌟 ЭТАЖ %d ПРОЙДЕН! Бездна зовет...", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(fmt.Sprintf("Спуск на этаж %d!", m.Floor)))
		}

		m.initDungeonForFloor(m.Floor)
		return
	}

	m.PartyPos = next
	m.revealFog()
}

func resetGameStatic(m Model) (Model, tea.Cmd) {
	m.accumulateDebugReport()
	saveDebugReportToFile()

	savedGold := int(float64(m.Gold) * LegacyTaxRate)
	newLegacy := m.Legacy
	newLegacy.TreasuryGold = savedGold

	fresh := initialModelWithLegacy(newLegacy)
	fresh.TermWidth = m.TermWidth
	fresh.TermHeight = m.TermHeight

	return fresh, tea.Batch(tickCmd(fresh.SpeedMs), menuTickCmd())
}

func (m Model) resetGame() (Model, tea.Cmd) {
	return resetGameStatic(m)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermWidth = msg.Width
		m.TermHeight = msg.Height
		if len(m.Grid) == 0 {
			m.initDungeonForFloor(m.Floor)
		}

	case MenuTickMsg:
		if m.State == StateMenu {
			m.MenuCountdown--
			if m.MenuCountdown <= 0 {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
			return m, menuTickCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3", "4":
			if m.State == StateInfoBook {
				tabIdx, _ := strconv.Atoi(msg.String())
				m.CodexTab = tabIdx - 1
				m.StatsScroll = 0
			} else {
				if msg.String() == "1" {
					m.SpeedMs = 280
				} else if msg.String() == "2" {
					m.SpeedMs = 120
				}
			}
		case "tab":
			if m.State == StateInfoBook {
				m.CodexTab = (m.CodexTab + 1) % 4
				m.StatsScroll = 0
			}
		case "f":
			if m.State == StatePlaying && m.Combat != nil {
				m.attemptFlee()
			}
		case "enter", " ":
			if m.State == StateMenu {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
			if m.State == StatePlaying {
				m.AutoMode = !m.AutoMode
				if m.AutoMode {
					delay := m.SpeedMs
					if m.InTown {
						delay = m.TownDelayMs
					}
					return m, tickCmd(delay)
				}
			}
		case "r":
			if m.State != StatePlaying && m.State != StateMenu {
				return m.resetGame()
			}
		case "s":
			if m.State == StatePlaying {
				m.State = StateStatsManual
				m.StatsScroll = 0
			} else if m.State == StateStatsManual {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "i":
			if m.State == StatePlaying {
				m.State = StateInfoBook
				m.StatsScroll = 0
			} else if m.State == StateInfoBook {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "e":
			if m.State == StatePlaying {
				m.State = StateArmory
				m.StatsScroll = 0
			} else if m.State == StateArmory {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "esc":
			if m.State == StateArmory || m.State == StateStatsManual || m.State == StateInfoBook {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "up", "k":
			if m.State == StatePlaying {
				if m.LogScroll < len(m.Logs)-3 {
					m.LogScroll++
				}
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				if m.StatsScroll > 0 {
					m.StatsScroll--
				}
			}
		case "down", "j":
			if m.State == StatePlaying {
				if m.LogScroll > 0 {
					m.LogScroll--
				}
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				if m.StatsScroll < 500 {
					m.StatsScroll++
				}
			}
		case "pgup":
			if m.State == StatePlaying {
				m.LogScroll = min(len(m.Logs)-3, m.LogScroll+5)
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				m.StatsScroll = max(0, m.StatsScroll-6)
			}
		case "pgdown":
			if m.State == StatePlaying {
				m.LogScroll = max(0, m.LogScroll-5)
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				m.StatsScroll += 6
			}
		case "+", "=":
			if m.SpeedMs > 60 {
				m.SpeedMs -= 30
			}
		case "-", "_":
			if m.SpeedMs < 800 {
				m.SpeedMs += 40
			}
		case "n":
			if m.State == StatePlaying && !m.AutoMode {
				m.step()
			}
		}

	case RestartTickMsg:
		if m.State == StateDefeat {
			m.RestartCountdown--
			if m.RestartCountdown <= 0 {
				return m.resetGame()
			}
			return m, restartTickCmd()
		}

	case TickMsg:
		if m.State == StatePlaying && m.AutoMode {
			m.step()
			if m.State == StateDefeat {
				return m, restartTickCmd()
			}
			delay := m.SpeedMs
			if m.InTown {
				delay = m.TownDelayMs
			}
			return m, tickCmd(delay)
		}
	}
	return m, nil
}

func main() {
	flag.Parse()

	m := initialModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Ошибка запуска: %v\n", err)
		os.Exit(1)
	}

	if *debugReportFlag {
		if fm, ok := finalModel.(*Model); ok {
			fm.accumulateDebugReport()
		} else if fmVal, ok := finalModel.(Model); ok {
			fmVal.accumulateDebugReport()
		}
		saveDebugReportToFile()
	}
}
