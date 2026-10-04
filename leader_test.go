package main

import "testing"

func leaderParty(t *testing.T, calling LeaderCalling, classes ...HeroClass) *Model {
	t.Helper()
	seedRNG(21)
	m := initialModel()
	m.State = StatePlaying
	m.Relic = nil
	bp := HeroBlueprint{Name: "Лидер", Gender: GenderMale, Race: RaceElf, Class: ClassPaladin, Calling: calling}
	m.Blueprint = &bp
	m.Party = []*Hero{createLeaderHero(bp, 0)}
	for _, c := range classes {
		h := createHero(c, 1, 0)
		h.Race = RaceElf // без бонуса опыта у людей, чтобы числа были точными
		h.Weapon, h.Head, h.Chest, h.Legs = nil, nil, nil, nil
		m.Party = append(m.Party, h)
	}
	return &m
}

func TestLeaderLookup(t *testing.T) {
	m := leaderParty(t, CallingInspirer, ClassMage)
	if m.leader() != m.Party[0] {
		t.Fatal("leader() должен находить лидера")
	}
	if m.activeCalling() != CallingInspirer {
		t.Fatal("призвание активно, пока лидер на ногах")
	}
	m.Party[0].IsDowned = true
	if m.activeCalling() != "" {
		t.Fatal("упавший лидер не даёт бонусов призвания")
	}
	m.Party[0].IsDowned = false
	m.Party[0].IsDead = true
	if m.activeCalling() != "" {
		t.Fatal("погибший лидер не даёт бонусов призвания")
	}

	m.Party = m.Party[1:]
	if m.leader() != nil || m.activeCalling() != "" {
		t.Fatal("без лидера бонусов быть не должно")
	}
}

func TestStrategistGoldBonus(t *testing.T) {
	m := leaderParty(t, CallingStrategist, ClassMage)
	if got, want := m.goldMult(), 1.0+float64(callingStrategistGoldPct)/100; got != want {
		t.Fatalf("множитель золота %v, ожидалось %v", got, want)
	}
	m.Party[0].IsDowned = true
	if m.goldMult() != 1.0 {
		t.Fatal("упавший лидер не должен давать бонус золота")
	}

	m = leaderParty(t, CallingInspirer, ClassMage)
	if m.goldMult() != 1.0 {
		t.Fatal("другое призвание не должно менять золото")
	}
	m.Relic = &PartyRelic{GoldMult: 1.5}
	if m.goldMult() != 1.5 {
		t.Fatal("реликвия должна учитываться")
	}
	m = leaderParty(t, CallingStrategist, ClassMage)
	m.Relic = &PartyRelic{GoldMult: 2.0}
	if got, want := m.goldMult(), 2.0*(1.0+float64(callingStrategistGoldPct)/100); got != want {
		t.Fatalf("бонусы должны перемножаться: %v != %v", got, want)
	}
}

func TestMentorExpBonus(t *testing.T) {
	m := leaderParty(t, CallingMentor, ClassWarrior)
	ally := m.Party[1]
	m.distributePartyExp(100)
	if want := 50 + 50*callingMentorExpPct/100; ally.Exp != want {
		t.Fatalf("опыт союзника с наставником: %d, ожидалось %d", ally.Exp, want)
	}

	m = leaderParty(t, CallingInspirer, ClassWarrior)
	ally = m.Party[1]
	m.distributePartyExp(100)
	if ally.Exp != 50 {
		t.Fatalf("без наставника опыт должен быть 50, получено %d", ally.Exp)
	}
}

func TestInspirerStressCut(t *testing.T) {
	m := leaderParty(t, CallingInspirer, ClassWarrior)
	ally, leader := m.Party[1], m.Party[0]
	leader.Head, leader.Chest, leader.Legs = nil, nil, nil

	m.addStress(ally, 20)
	if want := 20 * (100 - callingInspirerStressCut) / 100; ally.Stress != want {
		t.Fatalf("стресс союзника с вдохновителем: %d, ожидалось %d", ally.Stress, want)
	}
	leader.Stress = 0 // отбрасываем «брызги» стресса от союзника
	m.addStress(leader, 20)
	if leader.Stress != 20 {
		t.Fatalf("сам лидер бонуса не получает: %d", leader.Stress)
	}

	m = leaderParty(t, CallingMentor, ClassWarrior)
	m.addStress(m.Party[1], 20)
	if m.Party[1].Stress != 20 {
		t.Fatalf("другое призвание не снижает стресс: %d", m.Party[1].Stress)
	}
}

func TestLeaderFallShocksParty(t *testing.T) {
	m := leaderParty(t, CallingStrategist, ClassWarrior, ClassMage)
	leader := m.Party[0]
	leader.IsDowned = true
	m.onLeaderDowned(leader)
	for _, ally := range m.Party[1:] {
		if ally.Stress != leaderShockStress {
			t.Errorf("падение лидера должно давать %d стресса, у союзника %d", leaderShockStress, ally.Stress)
		}
	}
	if leader.Stress != 0 {
		t.Error("сам лидер от своего падения стресса не получает")
	}

	before := m.Party[1].Stress
	m.onLeaderDowned(m.Party[1]) // не лидер — ничего не происходит
	if m.Party[1].Stress != before {
		t.Error("падение обычного героя не должно запускать шок лидера")
	}
}

func TestHazardDeathSparesLeader(t *testing.T) {
	m := leaderParty(t, CallingInspirer, ClassWarrior)
	leader, ally := m.Party[0], m.Party[1]
	ally.Weapon = nil

	m.hazardDeath(ally)
	if !ally.IsDead {
		t.Error("обычный герой от ловушки гибнет")
	}

	m.hazardDeath(leader)
	if leader.IsDead || !leader.IsDowned || leader.HP != 0 {
		t.Errorf("лидер должен лишь потерять сознание: dead=%v downed=%v hp=%d", leader.IsDead, leader.IsDowned, leader.HP)
	}
	if leader.Weapon == nil {
		t.Error("лидер не должен терять снаряжение")
	}
}

func TestChurchRevivesLeaderForFreeAndNeverAbandonsHim(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		m := leaderParty(t, CallingInspirer, ClassWarrior, ClassMage)
		seedRNG(seed)
		m.Gold = 0
		m.InTown = true
		m.TownPhase = TownPhaseChurch
		leader, warrior := m.Party[0], m.Party[1]
		leader.IsDowned, leader.HP = true, 0
		warrior.IsDowned, warrior.HP = true, 0
		// единственный живой — маг: переносчиков меньше, чем раненых

		m.stepTown()

		if leader.IsDead || leader.LostInAbyss {
			t.Fatalf("seed %d: лидера бросили в Бездне", seed)
		}
		if leader.IsDowned || leader.HP <= 0 {
			t.Fatalf("seed %d: Храм должен поставить лидера на ноги бесплатно (hp=%d)", seed, leader.HP)
		}
		if m.Gold != 0 {
			t.Fatalf("seed %d: лечение лидера не должно стоить золота, золото %d", seed, m.Gold)
		}
		if !warrior.IsDead {
			t.Fatalf("seed %d: без золота и рук воин остаётся в Бездне", seed)
		}
	}
}

func TestGuildNeverReplacesLeader(t *testing.T) {
	m := leaderParty(t, CallingInspirer, ClassWarrior, ClassMage)
	leader := m.Party[0]
	leader.IsDowned = true // например, не успели вылечить
	m.Party[1].IsDead = true
	m.Gold = 0
	m.InTown = true
	m.TownPhase = TownPhaseGuild

	m.stepTown()

	if m.Party[0] != leader || !m.Party[0].IsLeader {
		t.Fatal("Гильдия не должна заменять лидера")
	}
	if m.Party[1].IsDead {
		t.Fatal("павшего обычного героя Гильдия обязана заменить")
	}
	if m.Party[1].IsLeader {
		t.Fatal("новичок не может быть лидером")
	}
}

func TestLeaderSurvivesLongSimulation(t *testing.T) {
	for seed := int64(60); seed < 66; seed++ {
		seedRNG(seed)
		bp := HeroBlueprint{Name: "Лидер", Gender: GenderMale, Race: RaceOlongr, Class: ClassTank, Calling: CallingInspirer}
		m := initialModelWith(TownLegacy{}, &bp)
		m.State = StatePlaying

		for i := 0; i < 5000 && m.State == StatePlaying; i++ {
			m.step()
			l := m.leader()
			if l == nil {
				t.Fatalf("seed %d, шаг %d: лидер пропал из отряда", seed, i)
			}
			if l.IsDead {
				t.Fatalf("seed %d, шаг %d: лидер погиб навсегда", seed, i)
			}
		}
	}
}
