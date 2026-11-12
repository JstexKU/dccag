package main

import "strings"

// ============================================================
// АСПЕКТЫ МОНСТРОВ (аффиксы)
// ============================================================
//
// Огненный бьёт чистым уроном (броня не защищает), ядовитый отравляет и повышает стресс,
// ледяной сковывает бойцов и замедляет инициативу отряда, вампирный лечится от нанесённого
// урона, каменный — крепче обычных (характеристики заданы при создании стаи). Эффекты
// описаны в кодексе и видны игроку: значок у названия монстра, строки в журнале боя и
// значки состояний на карточках героев.

const (
	poisonTurns            = 3  // сколько ходов героя действует яд
	poisonStressOnHit      = 8  // стресс от ядовитого удара
	chillTurns             = 2  // сколько ходов героя действует холод
	chillSkipChance        = 30 // шанс скованного героя потерять ход, %
	frostInitiativePenalty = 1  // штраф инициативы героев за каждого ледяного монстра
	frostInitiativeCap     = 3  // но не больше
	vampiricHealPct        = 50 // доля нанесённого урона, возвращаемая вампиром, %
)

// affixIcon — значок аспекта; у обычных монстров пустой.
func affixIcon(a MonsterAffix) string {
	switch a {
	case AffixFire:
		return "🔥"
	case AffixPoison:
		return "☣️"
	case AffixFrost:
		return "❄️"
	case AffixStone:
		return "🪨"
	case AffixVampiric:
		return "🩸"
	}
	return ""
}

// affixTitleKey — ключ названия аспекта в кодексе (значок и прилагательное).
func affixTitleKey(a MonsterAffix) string {
	switch a {
	case AffixFire:
		return "codex.affix.1.title"
	case AffixPoison:
		return "codex.affix.2.title"
	case AffixFrost:
		return "codex.affix.3.title"
	case AffixStone:
		return "codex.affix.4.title"
	case AffixVampiric:
		return "codex.affix.5.title"
	}
	return ""
}

// monsterName — название монстра; у мутантов перед ним стоит значок аспекта.
func (m Model) monsterName(mob *Monster) string {
	name := T(m.Lang, mob.NameKey)
	if icon := affixIcon(mob.Affix); icon != "" {
		name = icon + name
	}
	if mob.PoisonTurns > 0 {
		name += " ☣"
	}
	if mob.ChilledTurns > 0 {
		name += " ❄"
	}
	return name
}

// poisonDamage — урон яда за ход героя; растёт с глубиной.
func poisonDamage(floor int) int {
	return 2 + floor/4
}

// affixRawDamage — урон монстра до множителей: огненный игнорирует броню жертвы,
// остальные вычитают половину её защиты.
func affixRawDamage(mob *Monster, victimDef int) int {
	raw := int(float64(mob.Atk) * 1.25)
	if mob.Affix != AffixFire {
		raw -= victimDef / 2
	}
	return raw
}

// applyAffixOnHit — эффекты аспекта после удара монстра по герою (dealt — нанесённый урон).
func (m *Model) applyAffixOnHit(mob *Monster, victim *Hero, dealt int) {
	if mob == nil || victim == nil || mob.Affix == AffixNone || mob.Affix == "" {
		return
	}
	mobName := T(m.Lang, mob.NameKey)
	victimName := victim.DisplayName(m.Lang)
	alive := victim.HP > 0

	switch mob.Affix {
	case AffixFire:
		m.addLog(fireStyle.Render(T(m.Lang, "combat.log.affix_fire", mobName, victimName)))

	case AffixPoison:
		if !alive {
			return
		}
		// «Ядовитая» броня сокращает яд (но не меньше одного хода)
		victim.PoisonTurns = max(1, poisonTurns-victim.elementPieces(ElemPoison))
		victim.PoisonDmg = poisonDamage(m.Floor)
		stressBefore := victim.Stress
		m.addStress(victim, poisonStressOnHit)
		m.addLog(stressStyle.Render(T(m.Lang, "combat.log.affix_poison", mobName, victimName, max(0, victim.Stress-stressBefore))))

	case AffixFrost:
		if !alive {
			return
		}
		// «Леденящая» броня даёт шанс устоять перед холодом
		if n := victim.elementPieces(ElemFrost); n > 0 && rng.Intn(100) < frostResistPerPiece*n {
			m.addLog(fountStyle.Render(T(m.Lang, "combat.log.resist_frost", victimName)))
			return
		}
		victim.ChilledTurns = chillTurns
		m.addLog(fountStyle.Render(T(m.Lang, "combat.log.affix_frost", mobName, victimName)))

	case AffixVampiric:
		heal := dealt * vampiricHealPct / 100
		if heal > 0 && !mob.IsDead {
			before := mob.HP
			mob.HP = min(mob.MaxHP, mob.HP+heal)
			if healed := mob.HP - before; healed > 0 {
				m.addLog(dangerStyle.Render(T(m.Lang, "combat.log.affix_vampiric", mobName, victimName, healed)))
			}
		}
	}
}

// tickHeroStatuses — начало хода героя: действует яд (он не убивает, оставляя 1 HP) и холод
// (скованный герой с некоторым шансом теряет ход). Возвращает true, если ход потерян.
func (m *Model) tickHeroStatuses(h *Hero) bool {
	name := h.DisplayName(m.Lang)

	if h.PoisonTurns > 0 {
		h.PoisonTurns--
		if dmg := min(h.PoisonDmg, h.HP-1); dmg > 0 {
			h.HP -= dmg
			m.addLog(stressStyle.Render(T(m.Lang, "combat.log.poison_tick", name, dmg)))
		}
		if h.PoisonTurns == 0 {
			h.PoisonDmg = 0
		}
	}

	if h.ChilledTurns > 0 {
		h.ChilledTurns--
		if rng.Intn(100) < chillSkipChance {
			m.addLog(fountStyle.Render(T(m.Lang, "combat.log.chilled_skip", name)))
			return true
		}
	}
	return false
}

// cleanseStatuses снимает с героя яд и холод.
func (m *Model) cleanseStatuses(h *Hero) {
	h.PoisonTurns, h.PoisonDmg, h.ChilledTurns = 0, 0, 0
}

// cleansePartyStatuses снимает яд и холод со всего отряда.
func (m *Model) cleansePartyStatuses() {
	for _, h := range m.Party {
		m.cleanseStatuses(h)
	}
}

// clearCombatStatuses — состояния боя не переживают сам бой.
func (m *Model) clearCombatStatuses() {
	m.cleansePartyStatuses()
}

// frostPenalty — на сколько холод мутантов в стае замедляет инициативу героев в начале боя.
func frostPenalty(pack *MonsterPack) int {
	if pack == nil {
		return 0
	}
	n := 0
	for _, mob := range pack.Members {
		if !mob.IsDead && mob.Affix == AffixFrost {
			n++
		}
	}
	return min(frostInitiativeCap, n*frostInitiativePenalty)
}

// announceAffixes пишет в журнал начала боя, какие аспекты есть в стае, и штраф инициативы от холода.
func (m *Model) announceAffixes(pack *MonsterPack, frost int) {
	var names []string
	seen := map[MonsterAffix]bool{}
	for _, mob := range pack.Members {
		if mob.IsDead || mob.Affix == AffixNone || mob.Affix == "" || seen[mob.Affix] {
			continue
		}
		seen[mob.Affix] = true
		names = append(names, T(m.Lang, affixTitleKey(mob.Affix)))
	}
	if len(names) > 0 {
		m.addLog(accentStyle.Render(T(m.Lang, "combat.log.affixes", strings.Join(names, ", "))))
	}
	if frost > 0 {
		m.addLog(fountStyle.Render(T(m.Lang, "combat.log.frost_aura", frost)))
	}
}
