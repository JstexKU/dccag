package main

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var ruToEnKey = map[rune]rune{
	'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'і': 's', 'т': 'n', 'ь': 'm', 'б': ',', 'ю': '.', 'ё': '`',
	'ў': ']', '’': '\'',
	'Й': 'q', 'Ц': 'w', 'У': 'e', 'К': 'r', 'Е': 't', 'Н': 'y', 'Г': 'u', 'Ш': 'i', 'Щ': 'o', 'З': 'p', 'Х': '[', 'Ъ': ']',
	'Ф': 'a', 'Ы': 's', 'В': 'd', 'А': 'f', 'П': 'g', 'Р': 'h', 'О': 'j', 'Л': 'k', 'Д': 'l', 'Ж': ';', 'Э': '\'',
	'Я': 'z', 'Ч': 'x', 'С': 'c', 'М': 'v', 'И': 'b', 'І': 's', 'Т': 'n', 'Ь': 'm', 'Б': ',', 'Ю': '.', 'Ё': '`',
	'Ў': ']',
}

// normalizeKey приводит одиночные символы с русской/белорусской раскладки к их QWERTY-эквивалентам,
// а также приводит заглавные латинские буквы к нижнему регистру (на случай Caps Lock).
func normalizeKey(key string) string {
	runes := []rune(key)
	if len(runes) == 1 {
		r := runes[0]
		if en, ok := ruToEnKey[r]; ok {
			return string(en)
		}
		if r >= 'A' && r <= 'Z' {
			return string(r + ('a' - 'A'))
		}
	}
	return key
}

func tickCmd(speedMs, gen int) tea.Cmd {
	return tea.Tick(time.Duration(speedMs)*time.Millisecond, func(time.Time) tea.Msg {
		return TickMsg{Gen: gen}
	})
}

// restartTicks запускает новую цепочку тиков и обесценивает все ранее запланированные.
// Без этого быстрое закрытие окна или снятие паузы порождало вторую параллельную
// цепочку, и игра ускорялась вдвое.
func (m *Model) restartTicks(delayMs int) tea.Cmd {
	m.TickGen++
	return tickCmd(delayMs, m.TickGen)
}

func restartTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return RestartTickMsg(t)
	})
}

func menuTickCmd(gen int) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return MenuTickMsg{Gen: gen}
	})
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
		// Тики устаревшего таймера меню (после входа в создание героя) отбрасываются.
		if msg.Gen != m.MenuGen {
			return m, nil
		}
		if m.State == StateMenu {
			m.MenuCountdown--
			if m.MenuCountdown <= 0 {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
			return m, menuTickCmd(m.MenuGen)
		}

	case tea.KeyMsg:
		key := msg.String()

		// Экран создания героя забирает все клавиши без нормализации: буквы — это текст имени.
		if m.State == StateCreator {
			cmd := m.handleCreatorKey(msg)
			return m, cmd
		}

		// Для всех игровых экранов нормализуем раскладку клавиш (ЙЦУКЕН -> QWERTY)
		key = normalizeKey(key)

		if m.State == StateMenu {
			if cmd, handled := m.handleMenuKey(key); handled {
				return m, cmd
			}
		}

		// Экран тактики перехватывает свои клавиши раньше общей обработки.
		if m.State == StateTactics {
			if cmd, handled := m.handleTacticsKey(key); handled {
				return m, cmd
			}
		}

		// Экран библиотеки литературы перехватывает свои клавиши раньше общей обработки.
		if m.State == StateLibrary {
			if cmd, handled := m.handleLibraryKey(key); handled {
				return m, cmd
			}
		}

		// Ручное управление в подземелье: движение, команды героям, привал, выход.
		if m.manualDungeon() {
			if cmd, handled := m.handleManualKey(key); handled {
				return m, cmd
			}
		}

		// Ручное управление на улицах города: стрелки, вход через дверь.
		if m.ManualMode && m.InTown && !m.manualDungeon() && m.TownWalking {
			if m.handleTownWalkKey(key) {
				return m, nil
			}
		}

		// Выход из заведения на улицу (Esc); у Врат Esc в списке этажей сначала возвращает в меню Врат.
		if m.ManualMode && m.InTown && m.TownVisit && key == "esc" {
			if m.TownPhase == TownPhaseDepart && m.GateState.Mode == gateModeFloors {
				m.GateState.Mode = gateModeMain
				return m, nil
			}
			m.leaveBuilding(false)
			return m, nil
		}

		// Ручное управление в заведениях города (когда отряд не исследует подземелье).
		if m.ManualMode && m.InTown && !m.manualDungeon() && !m.TownWalking {
			var cmd tea.Cmd
			var handled bool
			phaseBefore := m.TownPhase
			switch m.TownPhase {
			case TownPhaseSellLoot:
				cmd, handled = m.handleMarketKey(key)
			case TownPhaseMagistrate:
				cmd, handled = m.handleMagistrateKey(key)
			case TownPhaseChurch:
				cmd, handled = m.handleChurchKey(key)
			case TownPhaseTavern:
				cmd, handled = m.handleTavernKey(key)
			case TownPhaseGuild:
				cmd, handled = m.handleGuildKey(key)
			case TownPhaseSmithy:
				cmd, handled = m.handleSmithyKey(key)
			case TownPhaseTannery:
				cmd, handled = m.handleTanneryKey(key)
			case TownPhaseAlchemist:
				cmd, handled = m.handleAlchemistKey(key)
			case TownPhaseDepart:
				handled = m.handleGateKey(key)
			}
			// Обработчики прежнего конвейера сами переводят фазу «дальше» (например, после отдыха
			// в таверне). На карте это означает «дела закончены»: возвращаем фазу и выходим на улицу.
			if m.TownVisit && m.InTown && m.TownPhase != phaseBefore {
				m.TownPhase = phaseBefore
				m.leaveBuilding(false)
			}
			if handled {
				return m, cmd
			}
		}

		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "l":
			m.Lang = NextLanguage(m.Lang)
		case "1", "2", "3", "4", "5", "6", "7":
			if m.State == StateInfoBook {
				tabIdx, _ := strconv.Atoi(key)
				m.CodexTab = (tabIdx - 1) % codexTabCount
				m.StatsScroll = 0
			} else {
				if key == "1" {
					m.SpeedMs = 280
				} else if key == "2" {
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
			if m.State == StatePlaying && m.Combat != nil && m.Combat.FleeCooldown == 0 {
				m.attemptFlee()
			}
		case "enter", " ":
			if m.State == StateMenu {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
			if m.State == StatePlaying {
				m.AutoMode = !m.AutoMode
				if m.AutoMode {
					delay := m.tickDelay()
					cmd := m.restartTicks(delay)
					return m, cmd
				}
			}
		case "r":
			// Рестарт доступен только с экранов, где он указан в подсказке: поражение и «Слава».
			if m.State == StateDefeat || m.State == StateStatsManual {
				return m.resetGame()
			}
		case "s":
			if m.State == StatePlaying {
				m.State = StateStatsManual
				m.StatsScroll = 0
			} else if m.State == StateStatsManual {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
		case "i":
			if m.State == StatePlaying {
				m.State = StateInfoBook
				m.StatsScroll = 0
			} else if m.State == StateInfoBook {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
		case "e":
			if m.State == StatePlaying {
				m.State = StateArmory
				m.StatsScroll = 0
			} else if m.State == StateArmory {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
		case "b":
			if m.State == StatePlaying {
				m.State = StateLibrary
				m.LibraryState.Cursor = 0
				m.LibraryState.ScrollOffset = 0
				m.LibraryState.TextScrollY = 0
				m.LibraryState.ViewingDetails = false
			} else if m.State == StateLibrary {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
			}
		case "t":
			if m.State == StatePlaying {
				m.State = StateTactics
			}
		case "m":
			if m.State == StatePlaying {
				cmd := m.toggleManual()
				return m, cmd
			}
		case "esc":
			if m.State == StateArmory || m.State == StateStatsManual || m.State == StateInfoBook || m.State == StateLibrary {
				m.State = StatePlaying
				cmd := m.restartTicks(m.SpeedMs)
				return m, cmd
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
				m.LogScroll = max(0, min(len(m.Logs)-3, m.LogScroll+5))
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
			if m.State == StatePlaying && !m.AutoMode && !m.awaitingCommand() {
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
		// Тики устаревших цепочек (см. restartTicks) молча отбрасываются.
		if msg.Gen != m.TickGen {
			return m, nil
		}
		if m.State == StatePlaying && m.AutoMode {
			// Ручной режим: ход в подземелье делает игрок.
			if m.manualDungeon() && (m.Combat == nil || m.awaitingCommand()) {
				return m, tickCmd(m.SpeedMs, m.TickGen)
			}
			// В ручном режиме город не прокручивается по таймеру сам — ждём решений игрока.
			if m.ManualMode && m.InTown && !m.manualDungeon() {
				return m, tickCmd(m.tickDelay(), m.TickGen)
			}
			m.step()
			if m.State == StateDefeat {
				return m, restartTickCmd()
			}
			delay := m.tickDelay()
			return m, tickCmd(delay, m.TickGen)
		}
	}
	return m, nil
}

func (m *Model) handleLibraryKey(key string) (tea.Cmd, bool) {
	st := &m.LibraryState
	volumes := GetVolumesByCategory(st.ActiveCategory)

	switch key {
	case "b", "esc":
		if st.ViewingDetails {
			st.ViewingDetails = false
			st.TextScrollY = 0
			return nil, true
		}
		m.State = StatePlaying
		cmd := m.restartTicks(m.SpeedMs)
		return cmd, true

	case "tab":
		st.ActiveCategory = (st.ActiveCategory + 1) % 6
		st.Cursor = 0
		st.ScrollOffset = 0
		st.TextScrollY = 0
		st.ViewingDetails = false
		return nil, true

	case "shift+tab":
		st.ActiveCategory = (st.ActiveCategory + 5) % 6
		st.Cursor = 0
		st.ScrollOffset = 0
		st.TextScrollY = 0
		st.ViewingDetails = false
		return nil, true

	case "1", "2", "3", "4", "5", "6":
		catIdx, _ := strconv.Atoi(key)
		st.ActiveCategory = LoreCategory(catIdx - 1)
		st.Cursor = 0
		st.ScrollOffset = 0
		st.TextScrollY = 0
		st.ViewingDetails = false
		return nil, true

	case "enter", "right":
		if len(volumes) > 0 {
			st.ViewingDetails = !st.ViewingDetails
			st.TextScrollY = 0
		}
		return nil, true

	case "left":
		if st.ViewingDetails {
			st.ViewingDetails = false
			st.TextScrollY = 0
			return nil, true
		}

	case "up", "k":
		if st.ViewingDetails {
			if st.TextScrollY > 0 {
				st.TextScrollY--
			}
		} else if len(volumes) > 0 {
			st.Cursor = (st.Cursor + len(volumes) - 1) % len(volumes)
			st.AdjustScroll(len(volumes), 12)
			st.TextScrollY = 0
		}
		return nil, true

	case "down", "j":
		if st.ViewingDetails {
			st.TextScrollY++
		} else if len(volumes) > 0 {
			st.Cursor = (st.Cursor + 1) % len(volumes)
			st.AdjustScroll(len(volumes), 12)
			st.TextScrollY = 0
		}
		return nil, true

	case "pgup":
		if st.ViewingDetails {
			st.TextScrollY = max(0, st.TextScrollY-6)
		} else if len(volumes) > 0 {
			st.Cursor = max(0, st.Cursor-5)
			st.AdjustScroll(len(volumes), 12)
			st.TextScrollY = 0
		}
		return nil, true

	case "pgdown":
		if st.ViewingDetails {
			st.TextScrollY += 6
		} else if len(volumes) > 0 {
			st.Cursor = min(len(volumes)-1, st.Cursor+5)
			st.AdjustScroll(len(volumes), 12)
			st.TextScrollY = 0
		}
		return nil, true
	}

	return nil, false
}