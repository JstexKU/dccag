package main

import "testing"

func partyWith(t *testing.T, classes ...HeroClass) *Model {
	t.Helper()
	seedRNG(7)
	m := initialModel()
	m.State = StatePlaying
	m.Party = nil
	for _, c := range classes {
		m.Party = append(m.Party, createHero(c, 1, 0))
	}
	return &m
}

func TestCampfireRelievesStress(t *testing.T) {
	m := partyWith(t, ClassWarrior, ClassMage)
	for _, h := range m.Party {
		h.Stress = 60
		h.HP = 1
	}
	m.eventCampfire()
	for _, h := range m.Party {
		if h.Stress != 35 {
			t.Errorf("стресс должен упасть на 25, получено %d", h.Stress)
		}
		if h.HP <= 1 {
			t.Errorf("костёр должен подлечить героя, HP=%d", h.HP)
		}
	}
}

func TestVaultNeedsRogueOrBruiser(t *testing.T) {
	m := partyWith(t, ClassMage, ClassCleric)
	g := m.Gold
	m.eventVault()
	if m.Gold != g {
		t.Error("без взломщика и силача тайник не должен давать золото")
	}

	m = partyWith(t, ClassRogue, ClassMage)
	g = m.Gold
	m.eventVault()
	if m.Gold <= g {
		t.Error("взломщик должен получить золото из тайника")
	}

	m = partyWith(t, ClassTank, ClassMage)
	g = m.Gold
	hp := m.Party[0].HP
	m.eventVault()
	if m.Gold <= g || m.Party[0].HP >= hp {
		t.Error("силач должен получить золото и урон")
	}
	if m.Party[0].HP < 1 {
		t.Error("выламывание двери не должно убивать героя")
	}
}

func TestIdolRefusedWhenGamblingDisabled(t *testing.T) {
	m := partyWith(t, ClassWarrior, ClassMage)
	m.Tactics.GambleEvents = false
	atk := m.Party[0].BaseAtk + m.Party[1].BaseAtk
	hp := m.Party[0].HP + m.Party[1].HP
	for i := 0; i < 20; i++ {
		m.eventIdol()
	}
	if m.Party[0].BaseAtk+m.Party[1].BaseAtk != atk || m.Party[0].HP+m.Party[1].HP != hp {
		t.Error("при отключённой азартности идол ничего менять не должен")
	}
}

func TestIdolCurseNeverKills(t *testing.T) {
	m := partyWith(t, ClassMage)
	m.Party[0].HP = 2
	for i := 0; i < 200; i++ {
		m.Party[0].HP = 2
		m.Party[0].Stress = 0
		m.eventIdol()
		if m.Party[0].HP < 1 {
			t.Fatal("проклятие идола не должно убивать напрямую")
		}
	}
}

func TestAmbushStartsCombat(t *testing.T) {
	m := partyWith(t, ClassWarrior, ClassMage)
	m.eventAmbush(Point{1, 1})
	if m.Combat == nil || m.Combat.Pack.LivingCount() == 0 {
		t.Fatal("засада должна начать бой")
	}
}

func TestMerchantSellsPotion(t *testing.T) {
	m := partyWith(t, ClassMage, ClassWarrior)
	m.Gold = 10000
	for _, h := range m.Party {
		h.Potions = nil
		h.Stress = 0
	}
	m.eventMerchant()
	total := 0
	for _, h := range m.Party {
		total += len(h.Potions)
	}
	if total != 1 || m.Gold >= 10000 {
		t.Errorf("торговец должен продать одно зелье (куплено %d, золото %d)", total, m.Gold)
	}

	m.Gold = 0
	m.eventMerchant()
	after := 0
	for _, h := range m.Party {
		after += len(h.Potions)
	}
	if after != total {
		t.Error("без золота покупка невозможна")
	}
}
