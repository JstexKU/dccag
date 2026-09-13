package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) accumulateDebugReport() {
	// Защита от двойного учёта рана.
	if m.RunCounted {
		return
	}
	m.RunCounted = true

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
		fmt.Println("\n[DEBUG] Telemetry report saved to: dccag_report.json")
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
	race := AllRaces[rand.Intn(len(AllRaces))]

	maxHP := 42
	baseDef := 2
	speed := 10
	skillNameKey := "skill.strike"
	skillCost := 10
	maxMP := 35

	switch class {
	case ClassTank:
		maxHP = 60
		baseDef = 4
		speed = 8
		skillNameKey = "skill.tank_stance"
		skillCost = 8
		maxMP = 30
	case ClassPaladin:
		maxHP = 54
		baseDef = 3
		speed = 9
		skillNameKey = "skill.paladin_holy"
		skillCost = 10
		maxMP = 35
	case ClassWarrior:
		maxHP = 48
		baseDef = 2
		speed = 10
		skillNameKey = "skill.warrior_rage"
		skillCost = 10
		maxMP = 25
	case ClassMonk:
		maxHP = 44
		baseDef = 1
		speed = 13
		skillNameKey = "skill.monk_flurry"
		skillCost = 8
		maxMP = 30
	case ClassRogue:
		maxHP = 35
		baseDef = 1
		speed = 15
		skillNameKey = "skill.rogue_stealth"
		skillCost = 12
		maxMP = 35
	case ClassRanger:
		maxHP = 38
		baseDef = 1
		speed = 13
		skillNameKey = "skill.ranger_shot"
		skillCost = 9
		maxMP = 30
	case ClassMage:
		maxHP = 28
		baseDef = 0
		speed = 11
		skillNameKey = "skill.mage_charge"
		skillCost = 15
		maxMP = 45
	case ClassWarlock:
		maxHP = 34
		baseDef = 1
		speed = 10
		skillNameKey = "skill.warlock_curse"
		skillCost = 11
		maxMP = 40
	case ClassCleric:
		maxHP = 34
		baseDef = 2
		speed = 9
		skillNameKey = "skill.cleric_aura"
		skillCost = 12
		maxMP = 40
	case ClassBard:
		maxHP = 36
		baseDef = 1
		speed = 12
		skillNameKey = "skill.bard_song"
		skillCost = 9
		maxMP = 40
	}

	raceMod := GetRaceModifiers(race)
	if raceMod.MaxHPBonusPercent != 0 {
		maxHP += int(float64(maxHP) * raceMod.MaxHPBonusPercent)
	}

	nameDef := getRandomHeroName()

	h := &Hero{
		NameKey:      nameDef.NameKey,
		Race:         race,
		Gender:       nameDef.Gender,
		Class:        class,
		Role:         GetClassRole(class),
		Level:        1,
		Exp:          0,
		MaxHP:        maxHP,
		HP:           maxHP,
		MaxMP:        maxMP,
		MP:           maxMP,
		BaseAtk:      6 + smithyLvl,
		BaseDef:      baseDef,
		Speed:        speed,
		SkillNameKey: skillNameKey,
		SkillCost:    skillCost,
		IsDead:       false,
		Potions:      []*Potion{},
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
	// rand.Seed убран: начиная с Go 1.20 глобальный math/rand автосидируется.

	// Перемешиваем все 10 классов и отбираем ровно 5 уникальных
	shuffledClasses := make([]HeroClass, len(AllClasses))
	copy(shuffledClasses, AllClasses)
	rand.Shuffle(len(shuffledClasses), func(i, j int) {
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
	}
	m.TownDialog = T(m.Lang, "town.log.enter_gate")
	m.Logs = append(m.Logs, T(m.Lang, "dungeon.log.start"))
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
		actualExp := expPerHero
		raceMod := GetRaceModifiers(h.Race)
		if raceMod.ExpBonusPercent > 0 {
			actualExp += int(float64(actualExp) * raceMod.ExpBonusPercent)
		}

		if h.GainExp(actualExp) {
			verb := TVerb(m.Lang, h.Gender, "достиг", "достигла", "reached")
			m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.lvl_up", h.DisplayName(m.Lang), verb, h.Level)))
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
				m.addLog(questStyle.Render(T(m.Lang, "dungeon.log.quest_done", T(m.Lang, m.CurrentQuest.TitleKey))))
			}
			return
		}
		m.CurrentQuest.Current += val
		if m.CurrentQuest.Current >= m.CurrentQuest.TargetCount {
			m.CurrentQuest.Current = m.CurrentQuest.TargetCount
			m.CurrentQuest.Completed = true
			m.addLog(questStyle.Render(T(m.Lang, "dungeon.log.quest_done", T(m.Lang, m.CurrentQuest.TitleKey))))
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
			// При смене экипировки старый предмет идёт в мешок.
			// Если мешок полон — старый предмет теряется (не кладём сверх лимита).
			if len(m.Bag) < m.currentBagCapacity() {
				m.Bag = append(m.Bag, *oldItem)
			} else {
				m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full", oldItem.DisplayName(m.Lang))))
			}
		}
		newItem := item
		bestHero.SetItemInSlot(newItem.Slot, &newItem)
		verb := TVerb(m.Lang, bestHero.Gender, "сменил", "сменила", "equipped")
		slotName := T(m.Lang, "slot."+string(item.Slot))
		m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.equip_swap",
			bestHero.DisplayName(m.Lang), verb, slotName, newItem.DisplayName(m.Lang), newItem.TotalStat())))
	} else {
		// Страховка: не класть в переполненный мешок.
		if len(m.Bag) >= m.currentBagCapacity() {
			m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_full", item.DisplayName(m.Lang))))
			return
		}
		m.Bag = append(m.Bag, item)
		m.addLog(subtleStyle.Render(T(m.Lang, "dungeon.log.bag_stored", item.DisplayName(m.Lang))))
	}
}

func (m *Model) recordFallenHero(h *Hero) {
	m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
		FullName: h.FullName(m.Lang),
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
		m.addLog(fireStyle.Render(T(m.Lang, "dungeon.log.martyr_crown")))
		for _, ally := range m.Party {
			if !ally.IsDead {
				ally.BaseAtk += 4
			}
		}
	}
}

func resetGameStatic(m Model) (Model, tea.Cmd) {
	m.accumulateDebugReport()
	saveDebugReportToFile()

	savedGold := int(float64(m.Gold) * LegacyTaxRate)
	newLegacy := m.Legacy
	newLegacy.TreasuryGold = savedGold

	fresh := initialModelWithLegacy(newLegacy)
	fresh.Lang = m.Lang
	fresh.TermWidth = m.TermWidth
	fresh.TermHeight = m.TermHeight

	return fresh, tea.Batch(tickCmd(fresh.SpeedMs), menuTickCmd())
}

func (m Model) resetGame() (Model, tea.Cmd) {
	return resetGameStatic(m)
}

// step() живёт в dungeon.go.
// resetGameStatic() и resetGame() — выше, в этом же файле.

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
		case "l":
			if m.Lang == LangRU {
				m.Lang = LangEN
			} else {
				m.Lang = LangRU
			}
		case "1", "2", "3", "4", "5", "6", "7":
			if m.State == StateInfoBook {
				tabIdx, _ := strconv.Atoi(msg.String())
				m.CodexTab = (tabIdx - 1) % codexTabCount
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
				m.CodexTab = (m.CodexTab + 1) % codexTabCount
				m.StatsScroll = 0
			}
		case "shift+tab":
			if m.State == StateInfoBook {
				m.CodexTab = (m.CodexTab + codexTabCount - 1) % codexTabCount
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
	// Фиксация 1 потока для стабильной работы на 32-битных архитектурах и в iSH
	runtime.GOMAXPROCS(1)

	flag.Parse()

	m := initialModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Startup error: %v\n", err)
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