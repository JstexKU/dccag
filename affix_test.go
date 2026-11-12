package main

import (
	"strings"
	"testing"
)

func affixModel(t *testing.T) Model {
	t.Helper()
	seedRNG(51)
	m := initialModel()
	m.State = StatePlaying
	m.Lang = LangRU
	m.Floor = 8
	m.Relic = nil
	return m
}

// Огненный игнорирует броню жертвы, остальные вычитают половину защиты.
func TestAffixRawDamage(t *testing.T) {
	for _, aff := range []MonsterAffix{AffixNone, AffixPoison, AffixFrost, AffixStone, AffixVampiric} {
		if got := affixRawDamage(&Monster{Atk: 20, Affix: aff}, 10); got != 20 {
			t.Errorf("%s: урон %d, ожидалось 20 (25 минус половина защиты)", aff, got)
		}
	}
	if got := affixRawDamage(&Monster{Atk: 20, Affix: AffixFire}, 10); got != 25 {
		t.Errorf("огонь: урон %d, ожидалось 25 (броня игнорируется)", got)
	}
}

func TestPoisonHitPoisonsAndStresses(t *testing.T) {
	m := affixModel(t)
	victim := m.Party[1]
	victim.HP, victim.Stress = victim.MaxHP, 0
	mob := &Monster{Affix: AffixPoison, NameKey: "mob.rat", HP: 10, MaxHP: 10}

	m.applyAffixOnHit(mob, victim, 12)

	if victim.PoisonTurns != poisonTurns || victim.PoisonDmg != poisonDamage(m.Floor) {
		t.Errorf("яд: ходов %d, урон %d", victim.PoisonTurns, victim.PoisonDmg)
	}
	if victim.Stress <= 0 {
		t.Error("ядовитый удар должен повышать стресс")
	}

	// Павшего героя не отравляют.
	dead := m.Party[2]
	dead.HP = 0
	m.applyAffixOnHit(mob, dead, 12)
	if dead.PoisonTurns != 0 {
		t.Error("яд на павшем герое")
	}
}

func TestFrostHitChills(t *testing.T) {
	m := affixModel(t)
	victim := m.Party[1]
	victim.HP = victim.MaxHP
	m.applyAffixOnHit(&Monster{Affix: AffixFrost, NameKey: "mob.rat"}, victim, 10)
	if victim.ChilledTurns != chillTurns {
		t.Errorf("холод: ходов %d", victim.ChilledTurns)
	}
}

func TestVampiricHitHealsMonster(t *testing.T) {
	m := affixModel(t)
	victim := m.Party[1]
	victim.HP = victim.MaxHP
	mob := &Monster{Affix: AffixVampiric, NameKey: "mob.rat", HP: 10, MaxHP: 30}

	m.applyAffixOnHit(mob, victim, 20)
	if mob.HP != 20 {
		t.Errorf("вампир: HP %d, ожидалось 20 (половина урона)", mob.HP)
	}
	m.applyAffixOnHit(mob, victim, 100)
	if mob.HP != 30 {
		t.Errorf("вампир: HP %d, лечение не должно превышать максимум", mob.HP)
	}

	mob.HP, mob.IsDead = 0, true
	m.applyAffixOnHit(mob, victim, 100)
	if mob.HP != 0 {
		t.Error("мёртвый вампир не лечится")
	}
}

func TestOtherAffixesLeaveNoStatuses(t *testing.T) {
	m := affixModel(t)
	victim := m.Party[1]
	victim.HP = victim.MaxHP
	for _, aff := range []MonsterAffix{AffixNone, AffixFire, AffixStone} {
		m.applyAffixOnHit(&Monster{Affix: aff, NameKey: "mob.rat"}, victim, 10)
		if victim.PoisonTurns != 0 || victim.ChilledTurns != 0 {
			t.Errorf("%s: появились состояния", aff)
		}
	}
}

// Яд снимает здоровье каждый ход, но не убивает.
func TestPoisonTickNeverKills(t *testing.T) {
	m := affixModel(t)
	h := m.Party[1]
	h.HP, h.PoisonTurns, h.PoisonDmg = 20, 2, 5

	if m.tickHeroStatuses(h) || h.HP != 15 || h.PoisonTurns != 1 {
		t.Errorf("первый ход яда: HP=%d ходов=%d", h.HP, h.PoisonTurns)
	}
	m.tickHeroStatuses(h)
	if h.HP != 10 || h.PoisonTurns != 0 || h.PoisonDmg != 0 {
		t.Errorf("второй ход яда: HP=%d ходов=%d урон=%d", h.HP, h.PoisonTurns, h.PoisonDmg)
	}

	h.HP, h.PoisonTurns, h.PoisonDmg = 3, 3, 5
	m.tickHeroStatuses(h)
	if h.HP != 1 {
		t.Errorf("яд не должен убивать: HP=%d", h.HP)
	}
	m.tickHeroStatuses(h)
	if h.HP != 1 {
		t.Errorf("при 1 HP яд ничего не отнимает: HP=%d", h.HP)
	}
}

// Скованный герой теряет ход примерно в chillSkipChance% случаев.
func TestChilledHeroSometimesLosesTurn(t *testing.T) {
	m := affixModel(t)
	h := m.Party[1]
	h.ChilledTurns = 0
	if m.tickHeroStatuses(h) {
		t.Fatal("без холода ход потерян")
	}

	skipped := 0
	for i := 0; i < 1000; i++ {
		h.ChilledTurns = 1
		if m.tickHeroStatuses(h) {
			skipped++
		}
		if h.ChilledTurns != 0 {
			t.Fatalf("холод должен убывать: %d", h.ChilledTurns)
		}
	}
	if skipped < 220 || skipped > 380 {
		t.Errorf("потеряно ходов %d из 1000, ожидалось около %d%%", skipped, chillSkipChance)
	}
}

func TestCleanseAndClearStatuses(t *testing.T) {
	m := affixModel(t)
	for _, h := range m.Party {
		h.PoisonTurns, h.PoisonDmg, h.ChilledTurns = 2, 3, 2
	}
	m.cleanseStatuses(m.Party[0])
	if p := m.Party[0]; p.PoisonTurns != 0 || p.ChilledTurns != 0 || p.PoisonDmg != 0 {
		t.Error("cleanseStatuses не снял состояния")
	}
	if m.Party[1].PoisonTurns == 0 {
		t.Error("cleanseStatuses должен трогать только одного героя")
	}
	m.clearCombatStatuses()
	for i, h := range m.Party {
		if h.PoisonTurns != 0 || h.ChilledTurns != 0 {
			t.Errorf("герой %d: состояния пережили бой", i)
		}
	}
}

func TestFrostInitiativePenalty(t *testing.T) {
	mk := func(affs ...MonsterAffix) *MonsterPack {
		p := &MonsterPack{}
		for _, a := range affs {
			p.Members = append(p.Members, &Monster{Affix: a})
		}
		return p
	}
	if frostPenalty(nil) != 0 || frostPenalty(mk(AffixFire, AffixNone)) != 0 {
		t.Error("без ледяных штрафа нет")
	}
	if got := frostPenalty(mk(AffixFrost, AffixFire, AffixFrost)); got != 2 {
		t.Errorf("два ледяных: штраф %d", got)
	}
	if got := frostPenalty(mk(AffixFrost, AffixFrost, AffixFrost, AffixFrost, AffixFrost)); got != frostInitiativeCap {
		t.Errorf("штраф %d должен быть ограничен %d", got, frostInitiativeCap)
	}
	dead := mk(AffixFrost)
	dead.Members[0].IsDead = true
	if frostPenalty(dead) != 0 {
		t.Error("мёртвый ледяной не замедляет")
	}
}

func TestMonsterNameShowsAffix(t *testing.T) {
	m := affixModel(t)
	plain := &Monster{NameKey: "mob.rat", Affix: AffixNone}
	if got := m.monsterName(plain); got != T(LangRU, "mob.rat") {
		t.Errorf("обычный монстр: %q", got)
	}
	for _, aff := range []MonsterAffix{AffixFire, AffixPoison, AffixFrost, AffixStone, AffixVampiric} {
		got := m.monsterName(&Monster{NameKey: "mob.rat", Affix: aff})
		if !strings.HasPrefix(got, affixIcon(aff)) || !strings.HasSuffix(got, T(LangRU, "mob.rat")) {
			t.Errorf("%s: %q", aff, got)
		}
		if T(LangRU, affixTitleKey(aff)) == affixTitleKey(aff) {
			t.Errorf("%s: нет перевода названия аспекта", aff)
		}
	}
}

func TestAnnounceAffixesLogsPackMutants(t *testing.T) {
	m := affixModel(t)
	pack := &MonsterPack{Members: []*Monster{
		{NameKey: "mob.rat", Affix: AffixFire},
		{NameKey: "mob.rat", Affix: AffixFire},
		{NameKey: "mob.rat", Affix: AffixFrost},
		{NameKey: "mob.rat", Affix: AffixNone},
	}}
	before := len(m.Logs)
	m.announceAffixes(pack, frostPenalty(pack))
	logs := strings.Join(m.Logs[before:], "\n")
	if !strings.Contains(logs, T(LangRU, "codex.affix.1.title")) || !strings.Contains(logs, T(LangRU, "codex.affix.3.title")) {
		t.Errorf("в журнале нет названий аспектов: %q", logs)
	}
	if strings.Count(logs, T(LangRU, "codex.affix.1.title")) != 1 {
		t.Error("одинаковые аспекты должны называться один раз")
	}

	before = len(m.Logs)
	m.announceAffixes(&MonsterPack{Members: []*Monster{{NameKey: "mob.rat"}}}, 0)
	if len(m.Logs) != before {
		t.Error("без мутантов журнал должен молчать")
	}
}

func TestStatusBadgesOnHeroCard(t *testing.T) {
	m := affixModel(t)
	h := m.Party[1]
	h.IsDowned, h.ReviveCooldown, h.Affliction = false, 0, AfflictionNone
	h.PoisonTurns = 2
	if !strings.Contains(stripANSI(buffBadge(h)), "☣") {
		t.Error("нет значка яда")
	}
	h.PoisonTurns, h.ChilledTurns = 0, 1
	if !strings.Contains(stripANSI(buffBadge(h)), "❄") {
		t.Error("нет значка холода")
	}
}
