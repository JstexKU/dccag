package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ============================================================
// РУЧНОЕ УПРАВЛЕНИЕ (клавиша M)
// ============================================================
//
// В ручном режиме игрок сам ведёт отряд по подземелью (стрелки) и отдаёт
// команду каждому герою в его ход. Автопилот при этом не принимает решений
// об отступлении и побеге, а зелья в бою пьёт только по команде игрока.
// Город по-прежнему проходится автоматически.
//
// Команды в бою:
//   A       — простой удар по выбранной цели (без навыков и расхода маны)
//   Space   — навык класса (шансовые условия навыков отключены; если ничего не
//             подходит — обычный удар с пометкой в журнале)
//   D       — защитная стойка: герой принимает удар на себя и переводит дух
//   P       — зелье (1-9 — номер, ←/→ — кому)
//   F       — попытка побега (уже была в игре)
//   ←/→     — выбор цели среди живых врагов

type ManualCmdKind int

const (
	CmdNone ManualCmdKind = iota
	CmdSkill
	CmdStrike
	CmdGuard
	CmdPotion
)

// ManualCmd — команда игрока для героя, чей ход наступил.
type ManualCmd struct {
	Kind      ManualCmdKind
	PotionIdx int
	Target    *Hero // кому достанется зелье (nil — самому себе)
}

const (
	manualCampTurns    = 8  // сколько «ходов» отдыха даёт один привал
	manualCampCooldown = 25 // шагов между привалами (иначе привал — бесплатное лечение)
)

// manualDungeon — ручной режим активен и отряд сейчас в подземелье.
func (m *Model) manualDungeon() bool {
	return m.ManualMode && m.State == StatePlaying && !m.InTown
}

// pendingHero возвращает живого героя, чей ход следующий в бою (или nil).
// Победа, поражение и ходы монстров/павших обрабатываются обычным step().
func (m *Model) pendingHero() *Hero {
	c := m.Combat
	if c == nil || len(c.TurnQueue) == 0 || c.Pack.LivingCount() == 0 || m.isPartyWiped() {
		return nil
	}
	idx := c.TurnIdx
	if idx >= len(c.TurnQueue) {
		idx = 0
	}
	e := c.TurnQueue[idx]
	if e.Type == CombatantHero && e.HeroRef != nil && !e.HeroRef.IsDead {
		return e.HeroRef
	}
	return nil
}

// awaitingCommand — игра ждёт команду игрока для очередного героя.
func (m *Model) awaitingCommand() bool {
	return m.manualDungeon() && m.Combat != nil && m.Combat.Cmd.Kind == CmdNone && m.pendingHero() != nil
}

// pickTarget выбирает цель одиночных умений: фокус игрока в ручном режиме,
// иначе — самый раненый враг, как у автопилота.
func (m *Model) pickTarget() *Monster {
	c := m.Combat
	if m.ManualMode && c.Focus != nil && !c.Focus.IsDead {
		return c.Focus
	}
	return c.Pack.GetLowestHPFocus()
}

func (m *Model) livingMobs() []*Monster {
	var out []*Monster
	if m.Combat == nil {
		return out
	}
	for _, mob := range m.Combat.Pack.Members {
		if !mob.IsDead {
			out = append(out, mob)
		}
	}
	return out
}

// cycleFocus переключает цель среди живых врагов (dir = +1 / -1).
func (m *Model) cycleFocus(dir int) {
	mobs := m.livingMobs()
	if len(mobs) == 0 {
		return
	}
	cur := m.pickTarget()
	idx := 0
	for i, mob := range mobs {
		if mob == cur {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(mobs)) % len(mobs)
	m.Combat.Focus = mobs[idx]
}

// cyclePotionTarget переключает получателя зелья среди живых героев.
func (m *Model) cyclePotionTarget(dir int) {
	c := m.Combat
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return
	}
	idx := 0
	for i, h := range living {
		if h == c.PotionTo {
			idx = i
			break
		}
	}
	c.PotionTo = living[(idx+dir+len(living))%len(living)]
}

// manualGuard — защитная стойка: герой переводит дух и принимает удары на себя.
func (m *Model) manualGuard(h *Hero) {
	h.IsGuarding = true
	h.AddBlock()
	h.MP = min(h.MaxMP, h.MP+3)
	h.Stress = max(0, h.Stress-2)
	m.addLog(healStyle.Render(T(m.Lang, "ctl.guard", h.DisplayName(m.Lang))))
}

// manualPotion — герой тратит ход на зелье из собственного пояса.
func (m *Model) manualPotion(h *Hero, cmd ManualCmd) {
	if cmd.PotionIdx < 0 || cmd.PotionIdx >= len(h.Potions) || h.Potions[cmd.PotionIdx] == nil {
		m.addLog(subtleStyle.Render(T(m.Lang, "ctl.potion_none", h.DisplayName(m.Lang))))
		return
	}
	target := cmd.Target
	if target == nil || target.IsDead {
		target = h
	}
	m.drinkPotion(h, target, cmd.PotionIdx)
}

// drinkPotion применяет зелье номер idx из пояса user к герою target.
func (m *Model) drinkPotion(user, target *Hero, idx int) {
	p := user.Potions[idx]
	rest := make([]*Potion, 0, len(user.Potions))
	for i, q := range user.Potions {
		if i != idx {
			rest = append(rest, q)
		}
	}
	user.Potions = rest

	pName := getPotionName(*p, m.Lang)
	tName := target.DisplayName(m.Lang)
	if target != user {
		m.addLog(potionStyle.Render(T(m.Lang, "ctl.potion_give", user.DisplayName(m.Lang), pName, tName)))
	}

	switch p.Type {
	case PotionHP:
		target.HP = min(target.MaxHP, target.HP+p.Power)
		verb := TVerb(m.Lang, target.Gender, "выпил", "выпила", "drank")
		m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_hp", tName, verb, pName, p.Power)))
	case PotionMP:
		target.MP = min(target.MaxMP, target.MP+p.Power)
		verb := TVerb(m.Lang, target.Gender, "выпил", "выпила", "drank")
		m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_mp", tName, verb, pName, p.Power)))
	case PotionStress:
		target.Stress = max(0, target.Stress-p.Power)
		target.Affliction = AfflictionNone
		verb := TVerb(m.Lang, target.Gender, "принял", "приняла", "took")
		m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_stress", tName, verb, pName, p.Power)))
	}
}

// manualMove — один шаг отряда в соседнюю клетку.
func (m *Model) manualMove(dx, dy int) {
	if m.State != StatePlaying || m.InTown || m.Combat != nil {
		return
	}
	next := Point{X: m.PartyPos.X + dx, Y: m.PartyPos.Y + dy}
	if next.Y < 0 || next.Y >= len(m.Grid) || next.X < 0 || next.X >= len(m.Grid[next.Y]) {
		return
	}
	if m.Grid[next.Y][next.X] == TileWall {
		return
	}

	m.RestTurnsLeft = 0
	m.Stats.TotalSteps++
	for _, h := range m.Party {
		m.checkAndDrinkPotions(h)
	}
	m.moveTo(next)

	if m.isPartyWiped() {
		m.State = StateDefeat
		m.RestartCountdown = 10
	}
}

// manualCamp — привал: короткий отдых с ограничением по частоте.
func (m *Model) manualCamp() {
	if m.Combat != nil || m.InTown {
		return
	}
	if m.isMonsterNearby() {
		m.addLog(dangerStyle.Render(T(m.Lang, "ctl.camp_danger")))
		return
	}
	if passed := m.Stats.TotalSteps - m.ManualLastCamp; passed < manualCampCooldown {
		m.addLog(subtleStyle.Render(T(m.Lang, "ctl.camp_wait", manualCampCooldown-passed)))
		return
	}

	m.Stats.TotalSteps += manualCampTurns
	m.ManualLastCamp = m.Stats.TotalSteps
	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		h.HP = min(h.MaxHP, h.HP+manualCampTurns)
		h.MP = min(h.MaxMP, h.MP+manualCampTurns)
		h.Stress = max(0, h.Stress-manualCampTurns)
	}
	biome := getBiome(m.Floor)
	key := fmt.Sprintf("dungeon.log.rest.%s.%d", biome.Name, rng.Intn(4)+1)
	m.addLog(healStyle.Render(T(m.Lang, key)))
	m.addLog(subtleStyle.Render(T(m.Lang, "ctl.camp_done", manualCampTurns)))
}

// manualLeave — возвращение в город; возможно только на клетке выхода.
func (m *Model) manualLeave() {
	if m.Combat != nil || m.InTown {
		return
	}
	if m.Grid[m.PartyPos.Y][m.PartyPos.X] != TileExit {
		m.addLog(subtleStyle.Render(T(m.Lang, "ctl.not_on_exit")))
		return
	}
	m.saveCurrentFloorState()
	m.InTown = true
	m.TownPhase = TownPhaseSellLoot
	m.TownDialog = T(m.Lang, "town.log.enter_gate")
	m.PathHistory = []Point{}
	m.LoopDetectCount = 0
}

// toggleManual включает или выключает ручной режим и перезапускает тики.
func (m *Model) toggleManual() tea.Cmd {
	m.ManualMode = !m.ManualMode
	m.PathHistory = []Point{}
	m.LoopDetectCount = 0
	if m.Combat != nil {
		m.Combat.Cmd = ManualCmd{}
		m.Combat.PotionMenu = false
		if !m.ManualMode {
			m.Combat.Focus = nil
		}
	}
	if m.ManualMode {
		m.RestTurnsLeft = 0
		m.addLog(accentStyle.Render(T(m.Lang, "ctl.mode_manual")))
	} else {
		m.addLog(accentStyle.Render(T(m.Lang, "ctl.mode_auto")))
	}
	m.AutoMode = true
	delay := m.SpeedMs
	if m.InTown {
		delay = m.TownDelayMs
	}
	return m.restartTicks(delay)
}

// commitCommand выполняет ход героя с уже записанной командой.
func (m *Model) commitCommand() tea.Cmd {
	m.step()
	if m.State == StateDefeat {
		return restartTickCmd()
	}
	return nil
}

// handleManualKey обрабатывает клавиши ручного режима в подземелье.
// Второе значение — «клавиша обработана».
func (m *Model) handleManualKey(key string) (tea.Cmd, bool) {
	if m.Combat == nil {
		switch key {
		case "up":
			m.manualMove(0, -1)
		case "down":
			m.manualMove(0, 1)
		case "left":
			m.manualMove(-1, 0)
		case "right":
			m.manualMove(1, 0)
		case "c":
			m.manualCamp()
		case "x":
			m.manualLeave()
		default:
			return nil, false
		}
		if m.State == StateDefeat {
			return restartTickCmd(), true
		}
		return nil, true
	}

	c := m.Combat
	hero := m.pendingHero()
	awaiting := hero != nil && c.Cmd.Kind == CmdNone

	// Меню зелий.
	if c.PotionMenu && awaiting {
		switch key {
		case "esc", "p":
			c.PotionMenu = false
			return nil, true
		case "left":
			m.cyclePotionTarget(-1)
			return nil, true
		case "right":
			m.cyclePotionTarget(1)
			return nil, true
		case " ", "enter":
			return nil, true // не даём случайно поставить игру на паузу
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(key[0] - '1')
			if idx < len(hero.Potions) {
				c.Cmd = ManualCmd{Kind: CmdPotion, PotionIdx: idx, Target: c.PotionTo}
				return m.commitCommand(), true
			}
			return nil, true
		}
		return nil, false
	}

	switch key {
	case "left":
		m.cycleFocus(-1)
		return nil, true
	case "right":
		m.cycleFocus(1)
		return nil, true
	}
	if !awaiting {
		return nil, false
	}

	switch key {
	case "a":
		c.Cmd = ManualCmd{Kind: CmdStrike}
		return m.commitCommand(), true
	case " ", "enter":
		c.Cmd = ManualCmd{Kind: CmdSkill}
		return m.commitCommand(), true
	case "d":
		c.Cmd = ManualCmd{Kind: CmdGuard}
		return m.commitCommand(), true
	case "p":
		if len(hero.Potions) == 0 {
			m.addLog(subtleStyle.Render(T(m.Lang, "ctl.potion_none", hero.DisplayName(m.Lang))))
			return nil, true
		}
		c.PotionMenu = true
		c.PotionTo = hero
		return nil, true
	}
	return nil, false
}

// manualControlsText возвращает подсказку по клавишам для нижней строки.
// Второе значение true — игра ждёт действия игрока (подсказка выделяется).
func (m Model) manualControlsText() (string, bool) {
	if !m.manualDungeon() {
		return "", false
	}
	if m.Combat == nil {
		return T(m.Lang, "ctl.bar.explore"), false
	}
	hero := m.pendingHero()
	if hero == nil || m.Combat.Cmd.Kind != CmdNone {
		return T(m.Lang, "ctl.bar.wait"), false
	}
	name := hero.DisplayName(m.Lang)
	if m.Combat.PotionMenu {
		var items []string
		for i, p := range hero.Potions {
			if p != nil {
				items = append(items, fmt.Sprintf("[%d] %s", i+1, getPotionName(*p, m.Lang)))
			}
		}
		to := name
		if m.Combat.PotionTo != nil {
			to = m.Combat.PotionTo.DisplayName(m.Lang)
		}
		return T(m.Lang, "ctl.bar.potion", name, strings.Join(items, " "), to), true
	}
	return T(m.Lang, "ctl.bar.combat", name), true
}
