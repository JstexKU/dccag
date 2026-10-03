package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// press прогоняет клавишу через настоящий Update, как это делает Bubble Tea.
func press(m *Model, key string) tea.Cmd {
	next, cmd := m.Update(keyMsg(key))
	*m = next.(Model)
	return cmd
}

func logsText(m *Model) string { return strings.Join(m.Logs, "\n") }

// newManualCombat создаёт ручной бой, в котором первыми ходят герои.
func newManualCombat(t *testing.T, classes ...HeroClass) *Model {
	t.Helper()
	m := partyWith(t, classes...)
	m.ManualMode = true
	m.AutoMode = true
	m.State = StatePlaying
	m.InTown = false
	m.startCombat(Point{X: 1, Y: 1}, spawnMonsterPack(false, 1))

	var heroes, mobs []TurnOrderEntry
	for _, e := range m.Combat.TurnQueue {
		if e.Type == CombatantHero {
			heroes = append(heroes, e)
		} else {
			mobs = append(mobs, e)
		}
	}
	m.Combat.TurnQueue = append(heroes, mobs...)
	m.Combat.TurnIdx = 0
	for _, mob := range m.Combat.Pack.Members {
		mob.HP = mob.MaxHP * 5 // чтобы ни один удар не закончил бой
		mob.MaxHP = mob.HP
	}
	return m
}

func TestToggleManualKey(t *testing.T) {
	seedRNG(11)
	m := initialModel()
	m.State = StatePlaying
	gen := m.TickGen

	cmd := press(&m, "m")
	if !m.ManualMode || cmd == nil || m.TickGen == gen {
		t.Fatal("M должна включать ручной режим и перезапускать тики")
	}
	press(&m, "m")
	if m.ManualMode {
		t.Fatal("повторное M должно возвращать автопилот")
	}
}

func TestManualExploringDoesNotAutoStep(t *testing.T) {
	seedRNG(12)
	m := initialModel()
	m.State = StatePlaying
	press(&m, "m")

	before := m.Stats.TotalSteps
	pos := m.PartyPos
	next, cmd := m.Update(TickMsg{Gen: m.TickGen})
	m = next.(Model)
	if m.Stats.TotalSteps != before || m.PartyPos != pos {
		t.Fatal("в ручном режиме тики не должны водить отряд по подземелью")
	}
	if cmd == nil {
		t.Fatal("цепочка тиков должна жить, чтобы работали ходы монстров и город")
	}
}

func TestManualMoveWalksAndStopsAtWalls(t *testing.T) {
	dirs := []struct {
		key    string
		dx, dy int
	}{{"up", 0, -1}, {"down", 0, 1}, {"left", -1, 0}, {"right", 1, 0}}

	for seed := int64(30); seed < 40; seed++ {
		seedRNG(seed)
		m := initialModel()
		m.State = StatePlaying
		m.ManualMode = true

		var walk, wall string
		for _, d := range dirs {
			p := Point{X: m.PartyPos.X + d.dx, Y: m.PartyPos.Y + d.dy}
			if p.Y < 0 || p.Y >= len(m.Grid) || p.X < 0 || p.X >= len(m.Grid[p.Y]) {
				continue
			}
			if _, pack := m.Packs[p]; pack {
				continue
			}
			switch m.Grid[p.Y][p.X] {
			case TileFloor:
				walk = d.key
			case TileWall:
				wall = d.key
			}
		}
		if walk == "" || wall == "" {
			continue
		}

		start := m.PartyPos
		steps := m.Stats.TotalSteps
		press(&m, wall)
		if m.PartyPos != start || m.Stats.TotalSteps != steps {
			t.Fatal("в стену ходить нельзя, и шаг не должен засчитываться")
		}
		press(&m, walk)
		if m.PartyPos == start {
			t.Fatal("шаг по свободной клетке должен перемещать отряд")
		}
		if m.Stats.TotalSteps != steps+1 {
			t.Fatal("шаг должен засчитываться в статистику")
		}
		return
	}
	t.Skip("не нашлось стартовой клетки со стеной и свободным соседом")
}

func TestManualCombatWaitsForCommand(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage)
	if !m.awaitingCommand() {
		t.Fatal("первым ходит герой — игра должна ждать команду")
	}

	steps, logs := m.Stats.TotalSteps, len(m.Logs)
	next, cmd := m.Update(TickMsg{Gen: m.TickGen})
	*m = next.(Model)
	if m.Stats.TotalSteps != steps || len(m.Logs) != logs || m.Combat.TurnIdx != 0 {
		t.Fatal("пока игрок не выбрал действие, бой не должен продвигаться")
	}
	if cmd == nil {
		t.Fatal("тики должны продолжать идти вхолостую")
	}

	// 'n' (шаг на паузе) тоже не должен отнимать ход у игрока.
	m.AutoMode = false
	press(m, "n")
	if m.Combat.TurnIdx != 0 {
		t.Fatal("шаг на паузе не должен выполнять ход героя за игрока")
	}
}

func TestManualStrikeUsesFocusAndNoMana(t *testing.T) {
	m := newManualCombat(t, ClassMage)
	mobs := m.livingMobs()
	if len(mobs) == 0 {
		t.Fatal("в стае нет живых монстров")
	}
	hero := m.pendingHero()
	hero.MP = hero.MaxMP
	mp := hero.MP

	press(m, "right") // выбор цели
	focus := m.pickTarget()
	if focus == nil {
		t.Fatal("после выбора цели должен быть фокус")
	}
	hp := focus.HP
	others := map[*Monster]int{}
	for _, mob := range mobs {
		if mob != focus {
			others[mob] = mob.HP
		}
	}

	press(m, "a")
	if hero.MP != mp {
		t.Fatalf("простой удар не должен тратить ману: %d -> %d", mp, hero.MP)
	}
	if m.Combat.TurnIdx != 1 {
		t.Fatal("после команды ход должен перейти дальше")
	}
	if focus.HP > hp {
		t.Fatal("цель не может выздороветь от удара")
	}
	for mob, was := range others {
		if mob.HP != was {
			t.Fatal("удар по выбранной цели не должен задевать остальных")
		}
	}
}

func TestManualSkillSpendsManaAndPlainStrikeDoesNot(t *testing.T) {
	m := newManualCombat(t, ClassWarrior)
	hero := m.pendingHero()
	hero.MP = hero.MaxMP
	hero.IsBerserk = false
	mp := hero.MP
	press(m, " ")
	if hero.MP >= mp {
		t.Fatal("команда «Навык» воина должна потратить ману")
	}

	m = newManualCombat(t, ClassWarrior)
	hero = m.pendingHero()
	hero.MP = hero.MaxMP
	mp = hero.MP
	press(m, "a")
	if hero.MP != mp {
		t.Fatal("команда «Удар» не должна тратить ману")
	}
}

func TestManualSkillWithoutManaSaysSo(t *testing.T) {
	m := newManualCombat(t, ClassWarrior)
	hero := m.pendingHero()
	hero.MP = 0
	press(m, " ")
	want := T(m.Lang, "ctl.no_skill", hero.DisplayName(m.Lang))
	if !strings.Contains(logsText(m), want) {
		t.Fatalf("в журнале должно быть пояснение %q", want)
	}
}

func TestManualGuard(t *testing.T) {
	m := newManualCombat(t, ClassMage)
	hero := m.pendingHero()
	hero.MP = 0
	hero.IsGuarding = false
	press(m, "d")
	if !hero.IsGuarding {
		t.Fatal("защитная стойка должна включать IsGuarding")
	}
	if hero.MP != min(hero.MaxMP, 3) {
		t.Fatalf("защита должна возвращать 3 MP, получено %d", hero.MP)
	}
}

func TestManualPotionSelfMenu(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage)
	hero := m.pendingHero()
	p := createPotion(PotionHP, SizeMedium, 1)
	hero.Potions = []*Potion{&p}
	hero.HP = 1

	press(m, "p")
	if !m.Combat.PotionMenu {
		t.Fatal("P должна открывать меню зелий")
	}
	if text, urgent := m.manualControlsText(); text == "" || !urgent {
		t.Fatal("в меню зелий должна быть подсказка")
	}
	press(m, "1")
	if hero.HP <= 1 {
		t.Fatal("лечебное зелье должно вылечить героя")
	}
	if len(hero.Potions) != 0 {
		t.Fatal("выпитое зелье должно исчезнуть из пояса")
	}
	if m.Combat.PotionMenu {
		t.Fatal("после выбора меню должно закрыться")
	}
}

func TestManualPotionMenuCancelAndEmptyBelt(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage)
	hero := m.pendingHero()
	hero.Potions = nil
	press(m, "p")
	if m.Combat.PotionMenu {
		t.Fatal("без зелий меню открываться не должно")
	}

	p := createPotion(PotionMP, SizeSmall, 1)
	hero.Potions = []*Potion{&p}
	press(m, "p")
	press(m, "esc")
	if m.Combat.PotionMenu || len(hero.Potions) != 1 {
		t.Fatal("Esc должна закрывать меню и не тратить зелье")
	}
	if m.Combat.TurnIdx != 0 {
		t.Fatal("отмена не должна тратить ход")
	}
}

func TestDrinkPotionOnAlly(t *testing.T) {
	m := partyWith(t, ClassCleric, ClassWarrior)
	user, ally := m.Party[0], m.Party[1]
	p := createPotion(PotionHP, SizeMedium, 1)
	user.Potions = []*Potion{&p}
	ally.HP = 1
	userHP := user.HP

	m.drinkPotion(user, ally, 0)
	if ally.HP <= 1 {
		t.Fatal("союзник должен получить лечение")
	}
	if user.HP != userHP || len(user.Potions) != 0 {
		t.Fatal("зелье тратится у владельца и не лечит его самого")
	}
}

func TestManualCombatHasNoAutoRetreat(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage, ClassCleric)
	// Монстры ходят первыми, а отряд еле жив: автопилот давно бы сбежал.
	var heroes, mobs []TurnOrderEntry
	for _, e := range m.Combat.TurnQueue {
		if e.Type == CombatantHero {
			heroes = append(heroes, e)
		} else {
			mobs = append(mobs, e)
		}
	}
	m.Combat.TurnQueue = append(mobs, heroes...)
	m.Combat.TurnIdx = 0
	for _, h := range m.Party {
		h.HP = 1
	}
	m.step()
	if m.InTown {
		t.Fatal("в ручном режиме автопилот не должен уводить отряд в город")
	}
}

func TestManualCampAndCooldown(t *testing.T) {
	seedRNG(13)
	m := initialModel()
	m.State = StatePlaying
	m.ManualMode = true
	for k := range m.Packs {
		delete(m.Packs, k)
	}
	m.Stats.TotalSteps = 100
	m.ManualLastCamp = 0
	for _, h := range m.Party {
		h.HP = 1
		h.Stress = 50
	}

	press(&m, "c")
	for _, h := range m.Party {
		if h.HP <= 1 || h.Stress >= 50 {
			t.Fatal("привал должен лечить и снимать стресс")
		}
	}

	hp := m.Party[0].HP
	press(&m, "c")
	if m.Party[0].HP != hp {
		t.Fatal("повторный привал сразу после первого должен быть заблокирован")
	}
}

func TestManualCampBlockedByMonsters(t *testing.T) {
	seedRNG(14)
	m := initialModel()
	m.State = StatePlaying
	m.ManualMode = true
	m.Stats.TotalSteps = 100
	m.Packs[m.PartyPos] = spawnMonsterPack(false, 1)
	m.Party[0].HP = 1

	press(&m, "c")
	if m.Party[0].HP != 1 {
		t.Fatal("рядом с монстрами привал невозможен")
	}
}

func TestManualLeaveNeedsExitTile(t *testing.T) {
	seedRNG(15)
	m := initialModel()
	m.State = StatePlaying
	m.ManualMode = true

	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileFloor
	press(&m, "x")
	if m.InTown {
		t.Fatal("уйти в город можно только с клетки выхода")
	}

	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileExit
	press(&m, "x")
	if !m.InTown || m.TownPhase != TownPhaseSellLoot {
		t.Fatal("на клетке выхода отряд должен вернуться в город")
	}
}

func TestManualControlsTextCoversAllPhases(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage)
	for _, lang := range []Language{LangRU, LangEN} {
		m.Lang = lang
		if text, urgent := m.manualControlsText(); text == "" || !urgent {
			t.Fatalf("%s: в ожидании команды подсказка должна быть выделена", lang)
		}
		m.Combat.Cmd = ManualCmd{Kind: CmdStrike}
		if text, urgent := m.manualControlsText(); text == "" || urgent {
			t.Fatalf("%s: во время хода противника подсказка нейтральная", lang)
		}
		m.Combat.Cmd = ManualCmd{}

		saved := m.Combat
		m.Combat = nil
		if text, _ := m.manualControlsText(); text == "" {
			t.Fatalf("%s: при исследовании должна быть подсказка", lang)
		}
		m.Combat = saved
	}

	m.ManualMode = false
	if text, _ := m.manualControlsText(); text != "" {
		t.Fatal("вне ручного режима подсказки быть не должно")
	}
}

func TestManualScreensRender(t *testing.T) {
	m := newManualCombat(t, ClassWarrior, ClassMage)
	m.TermWidth, m.TermHeight = 120, 40
	if m.View() == "" {
		t.Fatal("экран боя в ручном режиме должен рендериться")
	}
	m.Combat = nil
	if m.View() == "" {
		t.Fatal("экран исследования в ручном режиме должен рендериться")
	}
}

// Фазз: игрок жмёт случайные команды. Паники и зависания недопустимы,
// а город проходится автоматически.
func TestManualFuzzDoesNotPanicOrHang(t *testing.T) {
	moves := []string{"up", "down", "left", "right"}
	combat := []string{"a", " ", "d", "p", "1", "2", "left", "right", "esc", "f"}

	for seed := int64(50); seed < 56; seed++ {
		seedRNG(seed)
		m := initialModel()
		m.State = StatePlaying
		m.ManualMode = true

		for i := 0; i < 4000 && m.State == StatePlaying; i++ {
			switch {
			case m.InTown:
				m.step()
			case m.Combat != nil:
				if m.awaitingCommand() {
					press(&m, combat[rng.Intn(len(combat))])
				} else {
					m.step()
				}
			default:
				if rng.Intn(25) == 0 {
					press(&m, "c")
				} else if rng.Intn(60) == 0 {
					press(&m, "x")
				} else {
					press(&m, moves[rng.Intn(len(moves))])
				}
			}
		}
		if m.Stats.TotalSteps == 0 {
			t.Fatalf("seed %d: ни одного шага", seed)
		}
	}
}
