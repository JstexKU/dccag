package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// ОКНО «ПЕРСОНАЖИ» (клавиша P)
// ============================================================
//
// Подробная карточка каждого героя: состояние, шкала опыта до следующего уровня, боевые
// характеристики, способности, бонусы и состояния, снаряжение, пояс зелий и подвиги.
// Tab и стрелки ←/→ листают героев, цифры 1–9 выбирают героя, ↑/↓ прокручивают карточку.

const (
	charBarWidth = 22
	charLabelW   = 24
)

// critThreshold — порог крита на d20: удар критический, если бросок не ниже порога
// (чем меньше порог, тем чаще крит). Общий для боя и для окна «Персонажи».
func (m *Model) critThreshold(h *Hero) int {
	raceMod := GetRaceModifiers(h.Race)
	critThreshold := 19
	for _, it := range []*EquipItem{h.Weapon, h.Head, h.Chest} {
		if it != nil {
			critThreshold -= it.CritBonus
		}
	}
	if raceMod.CritChanceBonus > 0 {
		critThreshold -= raceMod.CritChanceBonus
	}
	// Реликвия «Шепчущий Череп»: снижение порога крита
	if m.Relic != nil && m.Relic.PartyCritBonus > 0 {
		critThreshold -= m.Relic.PartyCritBonus
	}
	if h.IsStealthed {
		critThreshold = 1
	}
	if critThreshold < 14 && !h.IsStealthed {
		critThreshold = 14
	}
	return critThreshold
}

// critChancePct — шанс крита в процентах при данном пороге d20.
func critChancePct(threshold int) int {
	n := 21 - threshold
	n = max(1, min(20, n))
	return n * 5
}

// heroPower — грубый боевой рейтинг героя для сравнения внутри отряда.
func heroPower(h *Hero) int {
	return h.TotalAtk()*3 + h.TotalDef()*3 + h.MaxHP/4 + h.TotalSpeed()*2
}

// stressResistParts — из чего складывается сопротивление стрессу: предметы (не больше 50%), раса, реликвия.
func (m Model) stressResistParts(h *Hero) (items, race, relic int) {
	for _, it := range []*EquipItem{h.Head, h.Chest, h.Legs} {
		if it != nil {
			items += it.StressRes
		}
	}
	items = min(50, items)
	race = GetRaceModifiers(h.Race).StressResPercent
	if m.Relic != nil {
		relic = m.Relic.StressRes
	}
	return items, race, relic
}

func (m Model) yesNo(v bool) string {
	if v {
		return T(m.Lang, "char.yes")
	}
	return T(m.Lang, "char.no")
}

func (m Model) heroIndex(h *Hero) int {
	for i, p := range m.Party {
		if p == h {
			return i
		}
	}
	return -1
}

// charStatusTag — жив ли герой, без сознания или потерян.
func (m Model) charStatusTag(h *Hero) string {
	switch {
	case h.LostInAbyss:
		return dangerStyle.Render("[" + T(m.Lang, "char.lost") + "]")
	case h.IsDead:
		return dangerStyle.Render("[" + T(m.Lang, "ui.dead") + "]")
	case h.IsDowned:
		return stressStyle.Render("[" + T(m.Lang, "char.downed") + "]")
	}
	return healStyle.Render("[" + T(m.Lang, "ui.alive") + "]")
}

// charRacePerks — особенности расы героя.
func (m Model) charRacePerks(r RaceType) []string {
	mod := GetRaceModifiers(r)
	pct := func(f float64) int {
		if f < 0 {
			return int(f*100 - 0.5)
		}
		return int(f*100 + 0.5)
	}
	var out []string
	if mod.ExpBonusPercent != 0 {
		out = append(out, T(m.Lang, "char.race.exp", pct(mod.ExpBonusPercent)))
	}
	if mod.SpeedFlat != 0 {
		out = append(out, T(m.Lang, "char.race.spd", mod.SpeedFlat))
	}
	if mod.CritChanceBonus != 0 {
		out = append(out, T(m.Lang, "char.race.crit", mod.CritChanceBonus))
	}
	if mod.LifeStealPercent != 0 {
		out = append(out, T(m.Lang, "char.race.lifesteal", pct(mod.LifeStealPercent)))
	}
	if mod.DefBonusPercent != 0 {
		out = append(out, T(m.Lang, "char.race.def", pct(mod.DefBonusPercent)))
	}
	if mod.MaxHPBonusPercent != 0 {
		out = append(out, T(m.Lang, "char.race.hp", pct(mod.MaxHPBonusPercent)))
	}
	if mod.ManaRegen != 0 {
		out = append(out, T(m.Lang, "char.race.mana", mod.ManaRegen))
	}
	if mod.StunImmune {
		out = append(out, T(m.Lang, "char.race.stun"))
	}
	if mod.StressResPercent != 0 {
		out = append(out, T(m.Lang, "char.race.stress", mod.StressResPercent))
	}
	if mod.LowHPAtkBonus != 0 {
		out = append(out, T(m.Lang, "char.race.lowhp", mod.LowHPAtkBonus))
	}
	if mod.FireImmune {
		out = append(out, T(m.Lang, "char.race.fire"))
	}
	if mod.RetaliateFire != 0 {
		out = append(out, T(m.Lang, "char.race.retaliate", mod.RetaliateFire))
	}
	if len(out) == 0 {
		out = append(out, T(m.Lang, "char.race.none"))
	}
	return out
}

// charItemLines — строки предмета: название и характеристика, затем бонусы и аффиксы.
func (m Model) charItemLines(slot EquipSlot, it *EquipItem) []string {
	slotName := T(m.Lang, "slot."+string(slot))
	if it == nil {
		return []string{fmt.Sprintf(" • %s %s", padRight(slotName+":", 9), subtleStyle.Render(T(m.Lang, "char.empty_slot")))}
	}

	statLabel := T(m.Lang, "ui.def_short")
	if slot == SlotWeapon {
		statLabel = T(m.Lang, "ui.atk_short")
	}
	out := []string{fmt.Sprintf(" • %s %s (%s %d)", padRight(slotName+":", 9),
		goldStyle.Render(it.DisplayName(m.Lang)), statLabel, it.TotalStat())}

	var parts []string
	if it.BonusHP > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.hp", it.BonusHP))
	}
	if it.BonusMP > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.mp", it.BonusMP))
	}
	if it.CritBonus > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.crit", it.CritBonus))
	}
	if it.SpeedBonus > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.spd", it.SpeedBonus))
	}
	if it.BlockBonus > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.block", it.BlockBonus))
	}
	if it.StressRes > 0 {
		parts = append(parts, T(m.Lang, "char.bonus.stress", it.StressRes))
	}
	if it.Suffix != nil {
		parts = append(parts, T(m.Lang, "char.suffix."+string(it.Suffix.Effect)))
	}
	if it.Prefix != nil {
		if slot == SlotWeapon {
			if kind, ok := prefixEnchantKind(it.Prefix.Element); ok {
				parts = append(parts, enchantIcon(kind)+" "+T(m.Lang, "char.weapon_innate"))
			}
		} else if it.Prefix.Element != ElemNone {
			parts = append(parts, T(m.Lang, "char.armor."+string(it.Prefix.Element)))
		}
	}
	if it.Enchant != nil {
		parts = append(parts, enchantLabel(m.Lang, it.Enchant.Kind)+" "+fmt.Sprint(it.Enchant.Level))
	}
	if len(parts) > 0 {
		out = append(out, "     "+subtleStyle.Render(strings.Join(parts, " | ")))
	}
	return out
}

// charSheetLines — все строки карточки героя (без прокрутки и рамки).
func (m Model) charSheetLines(h *Hero) []string {
	var out []string
	line := func(s string) { out = append(out, s) }
	section := func(key string) {
		out = append(out, "", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("── "+T(m.Lang, key)+" ──"))
	}
	label := func(key string) string { return padRight(T(m.Lang, key), charLabelW) }
	indent := strings.Repeat(" ", charLabelW)

	// --- Шапка героя
	name := h.FullName(m.Lang)
	if h.IsLeader {
		name = "★ " + name
	}
	line(lipgloss.NewStyle().Bold(true).Render(name))
	line(fmt.Sprintf("%s · %s · %s %d   %s", h.RaceName(m.Lang), h.ShortClass(m.Lang), T(m.Lang, "char.level"), h.Level, m.charStatusTag(h)))

	// --- Состояние: здоровье, мана, стресс, опыт
	section("char.sec.vitals")
	line(label("char.hp") + renderBar(h.HP, h.MaxHP, charBarWidth, "46", "238") + fmt.Sprintf(" %d/%d", h.HP, h.MaxHP))
	if h.MaxMP > 0 {
		line(label("char.mp") + renderBar(h.MP, h.MaxMP, charBarWidth, "39", "238") + fmt.Sprintf(" %d/%d", h.MP, h.MaxMP))
	}
	stressColor := lipgloss.Color("214")
	if h.Stress >= 70 {
		stressColor = lipgloss.Color("196")
	}
	affliction := h.Affliction
	if affliction == "" {
		affliction = AfflictionNone
	}
	line(label("char.stress") + renderBar(h.Stress, 100, charBarWidth, stressColor, "238") +
		fmt.Sprintf(" %d  (%s)", h.Stress, T(m.Lang, "affliction."+string(affliction))))

	need := max(1, h.NextLevelExp())
	line(label("char.exp") + renderBar(h.Exp, need, charBarWidth, "177", "238") +
		fmt.Sprintf(" %d/%d (%d%%)", h.Exp, need, h.Exp*100/need))
	line(indent + subtleStyle.Render(T(m.Lang, "char.exp_left", h.Level+1, max(0, need-h.Exp))))

	if h.ExpElixirPct > 0 {
		line(indent + accentStyle.Render(T(m.Lang, "char.elixir_active", h.ExpElixirPct)))
	} else if idx := m.heroIndex(h); idx >= 0 && !h.IsDead && !h.LostInAbyss {
		if offer := m.elixirOffer(idx, h); offer.Gap >= 1 {
			line(indent + dangerStyle.Render(T(m.Lang, "char.lagging", offer.Gap)))
			line(indent + subtleStyle.Render(T(m.Lang, "char.elixir_hint", offer.ExpPct, offer.Cost)))
		}
	}

	// --- Боевые характеристики
	section("char.sec.combat")
	line(label("char.atk") + fmt.Sprintf("%d  (%s %d)", h.TotalAtk(), T(m.Lang, "char.base"), h.BaseAtk))
	line(label("char.def") + fmt.Sprintf("%d  (%s %d)", h.TotalDef(), T(m.Lang, "char.base"), h.BaseDef))
	line(label("char.spd") + fmt.Sprintf("%d  (%s %d)", h.TotalSpeed(), T(m.Lang, "char.base"), h.Speed))
	thr := m.critThreshold(h)
	line(label("char.crit") + T(m.Lang, "char.crit_fmt", critChancePct(thr), thr))
	ires, rres, relic := m.stressResistParts(h)
	line(label("char.stressres") + T(m.Lang, "char.stressres_fmt", ires, rres, relic))
	line(label("char.role") + T(m.Lang, "char.role_fmt", h.Role.AggroWeight, m.yesNo(h.Role.CanGuard), m.yesNo(h.Role.CanHeal)))

	total, living := 0, 0
	for _, p := range m.Party {
		if !p.IsDead && !p.LostInAbyss {
			total += heroPower(p)
			living++
		}
	}
	avg := 0
	if living > 0 {
		avg = total / living
	}
	line(label("char.power") + T(m.Lang, "char.power_fmt", heroPower(h), avg))

	// --- Способности
	section("char.sec.abilities")
	if h.SkillNameKey != "" {
		state := T(m.Lang, "char.skill_ready")
		if h.MP < h.SkillCost {
			state = T(m.Lang, "char.skill_nomana")
		}
		line(" • " + T(m.Lang, "char.skill", T(m.Lang, h.SkillNameKey), h.SkillCost, state))
	} else {
		line(" • " + subtleStyle.Render(T(m.Lang, "char.skill_none")))
	}
	if h.IsLeader && h.Calling != "" {
		c := string(h.Calling)
		line(" • " + T(m.Lang, "char.leader_calling", T(m.Lang, "calling."+c+".name"), T(m.Lang, "calling."+c+".desc", callingValue(h.Calling))))
	}
	for _, perk := range m.charRacePerks(h.Race) {
		line(" • " + perk)
	}

	// --- Бонусы и состояния
	section("char.sec.bonuses")
	var muts []string
	for _, mm := range []struct {
		kind  MutationType
		count int
	}{
		{MutChimera, h.Mutations.ChimeraCount}, {MutFury, h.Mutations.FuryCount}, {MutTitan, h.Mutations.TitanCount},
		{MutAether, h.Mutations.AetherCount}, {MutBastion, h.Mutations.BastionCount},
	} {
		if mm.count > 0 {
			muts = append(muts, fmt.Sprintf("%s ×%d", T(m.Lang, "mut."+string(mm.kind)+".name"), mm.count))
		}
	}
	if len(muts) > 0 {
		line(" • " + T(m.Lang, "char.mutations") + ": " + strings.Join(muts, ", "))
	}
	if h.GearHP > 0 {
		line(" • " + T(m.Lang, "char.gear_hp", h.GearHP))
	}
	if h.GearMP > 0 {
		line(" • " + T(m.Lang, "char.gear_mp", h.GearMP))
	}
	if fury := h.suffixTotal(SuffFury); fury > 0 {
		line(" • " + T(m.Lang, "char.gear_fury", fury))
	}
	if regen := h.manaRegen(); regen > 0 {
		line(" • " + T(m.Lang, "char.gear_mana", regen))
	}
	var elems []string
	for _, e := range []ElementType{ElemFire, ElemFrost, ElemPoison, ElemLightning} {
		if n := h.elementPieces(e); n > 0 {
			kind, _ := prefixEnchantKind(e)
			elems = append(elems, fmt.Sprintf("%s×%d", enchantIcon(kind), n))
		}
	}
	if len(elems) > 0 {
		line(" • " + T(m.Lang, "char.gear_elements", strings.Join(elems, " ")))
	}
	var fx []string
	for _, ef := range h.weaponEffects() {
		s := fmt.Sprintf("%s %d", enchantLabel(m.Lang, ef.Kind), ef.Level)
		if ef.Innate {
			s += " (" + T(m.Lang, "char.innate") + ")"
		}
		fx = append(fx, s)
	}
	if len(fx) > 0 {
		line(" • " + T(m.Lang, "char.weapon_fx", strings.Join(fx, ", ")))
	}

	var states []string
	if h.PoisonTurns > 0 {
		states = append(states, T(m.Lang, "char.state.poison", h.PoisonTurns, h.PoisonDmg))
	}
	if h.ChilledTurns > 0 {
		states = append(states, T(m.Lang, "char.state.chill", h.ChilledTurns))
	}
	for _, s := range []struct {
		on  bool
		key string
	}{
		{h.IsGuarding, "char.state.guard"}, {h.IsBerserk, "char.state.berserk"}, {h.IsStealthed, "char.state.stealth"},
		{h.IsCharged, "char.state.charged"}, {h.IsAura, "char.state.aura"},
	} {
		if s.on {
			states = append(states, T(m.Lang, s.key))
		}
	}
	if h.ReviveCooldown > 0 {
		states = append(states, T(m.Lang, "char.state.revive", h.ReviveCooldown))
	}
	if len(states) == 0 {
		states = append(states, subtleStyle.Render(T(m.Lang, "char.state.none")))
	}
	for _, s := range states {
		line(" • " + s)
	}

	// --- Снаряжение
	section("char.sec.equipment")
	for _, pair := range []struct {
		slot EquipSlot
		it   *EquipItem
	}{{SlotWeapon, h.Weapon}, {SlotHead, h.Head}, {SlotChest, h.Chest}, {SlotLegs, h.Legs}} {
		for _, l := range m.charItemLines(pair.slot, pair.it) {
			line(l)
		}
	}

	// --- Пояс зелий и рюкзак отряда
	section("char.sec.belt")
	line(" " + T(m.Lang, "char.belt", len(h.Potions), m.MaxPotionSlots()))
	if len(h.Potions) == 0 {
		line("   " + subtleStyle.Render(T(m.Lang, "char.belt_empty")))
	}
	for _, p := range h.Potions {
		if p != nil {
			line(fmt.Sprintf("   • %s %s (%d)", p.Symbol, getPotionName(*p, m.Lang), p.Power))
		}
	}
	line(" " + T(m.Lang, "char.bag", len(m.Bag), m.currentBagCapacity()))

	// --- Подвиги
	section("char.sec.records")
	f := h.Feats
	recorded := 0
	for _, r := range []struct {
		key string
		val int
	}{
		{"char.rec.kills", f.Kills}, {"char.rec.bosses", f.BossKills}, {"char.rec.dealt", f.DamageDealt},
		{"char.rec.taken", f.DamageTaken}, {"char.rec.crits", f.CritsLanded}, {"char.rec.heals", f.HealsGiven},
		{"char.rec.escapes", f.NearDeathEscapes}, {"char.rec.treasure", f.TreasureFound},
		{"char.rec.blocks", f.Blocks}, {"char.rec.revives", f.Revives},
	} {
		if r.val > 0 {
			line(fmt.Sprintf(" • %s %d", padRight(T(m.Lang, r.key)+":", 32), r.val))
			recorded++
		}
	}
	if recorded == 0 {
		line(" " + subtleStyle.Render(T(m.Lang, "char.rec_none")))
	}
	return out
}

// charTabStrip — полоса с именами героев; выбранный подсвечен.
func (m Model) charTabStrip() string {
	var parts []string
	for i, h := range m.Party {
		label := fmt.Sprintf(" %d %s ", i+1, shortenItemName(h.DisplayName(m.Lang), 10))
		switch {
		case i == m.CharIndex:
			parts = append(parts, accentStyle.Bold(true).Reverse(true).Render(label))
		case h.IsDead || h.LostInAbyss:
			parts = append(parts, subtleStyle.Render(label))
		default:
			parts = append(parts, label)
		}
	}
	return strings.Join(parts, " ")
}

// charVisibleLines — сколько строк карточки помещается под шапкой и над подсказкой.
func (m Model) charVisibleLines() int {
	return max(6, m.TermHeight-9)
}

func (m Model) charMaxScroll() int {
	if len(m.Party) == 0 {
		return 0
	}
	idx := max(0, min(m.CharIndex, len(m.Party)-1))
	return max(0, len(m.charSheetLines(m.Party[idx]))-m.charVisibleLines())
}

func (m Model) renderCharacterScreen() string {
	if len(m.Party) == 0 {
		return ""
	}
	idx := max(0, min(m.CharIndex, len(m.Party)-1))
	h := m.Party[idx]

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(T(m.Lang, "char.title") + " [P]")
	header := title + subtleStyle.Render(fmt.Sprintf("   ◄ %d/%d ►", idx+1, len(m.Party))) + "\n" + m.charTabStrip()

	body := m.charSheetLines(h)
	visible := m.charVisibleLines()
	scroll := max(0, min(m.StatsScroll, max(0, len(body)-visible)))
	end := min(len(body), scroll+visible)
	shown := body[scroll:end]

	var sb strings.Builder
	sb.WriteString(header + "\n")
	sb.WriteString(strings.Join(shown, "\n"))
	sb.WriteString(strings.Repeat("\n", visible-len(shown)+1))
	sb.WriteString(subtleStyle.Render(T(m.Lang, "char.controls")))

	box := statsBoxStyle.Width(max(40, m.TermWidth-4)).Render(sb.String())
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// openCharacterSheet открывает окно (последний просмотренный герой запоминается).
func (m *Model) openCharacterSheet() {
	m.State = StateCharacter
	m.StatsScroll = 0
	if m.CharIndex < 0 || m.CharIndex >= len(m.Party) {
		m.CharIndex = 0
	}
}

func (m *Model) charStep(d int) {
	n := len(m.Party)
	if n == 0 {
		return
	}
	m.CharIndex = ((m.CharIndex+d)%n + n) % n
	m.StatsScroll = 0
}

func (m *Model) charScroll(d int) {
	m.StatsScroll = max(0, min(m.StatsScroll+d, m.charMaxScroll()))
}

// handleCharacterKey — клавиши окна «Персонажи». Пока окно открыто, оно забирает все клавиши,
// кроме выхода и смены языка, чтобы ничего не происходило в игре за ним.
func (m *Model) handleCharacterKey(key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+c", "q", "l":
		return nil, false
	case "p", "h", "esc":
		m.State = StatePlaying
		return m.restartTicks(m.SpeedMs), true
	case "tab", "right":
		m.charStep(1)
	case "shift+tab", "left":
		m.charStep(-1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if idx := int(key[0] - '1'); idx < len(m.Party) {
			m.CharIndex = idx
			m.StatsScroll = 0
		}
	case "up", "k":
		m.charScroll(-1)
	case "down", "j":
		m.charScroll(1)
	case "pgup":
		m.charScroll(-10)
	case "pgdown":
		m.charScroll(10)
	}
	return nil, true
}
