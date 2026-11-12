package main

import "testing"

// Кузница и кожевник — две конфигурации одного механизма: цены считаются по своим
// формулам, а один и тот же предмет не предлагается обеим мастерским сразу.
func TestUpgradeServicesPriceAndPartition(t *testing.T) {
	seedRNG(5)
	m := initialModel()
	m.State = StatePlaying
	m.Gold = 1_000_000
	m.Legacy.SmithyLevel = 2
	m.Legacy.TanneryLevel = 1

	smithy, tannery := m.GenerateSmithyOffers(), m.GenerateTanneryOffers()
	if len(smithy)+len(tannery) == 0 {
		t.Fatal("ни одна мастерская ничего не предложила стартовому отряду")
	}

	type key struct {
		hero int
		slot EquipSlot
	}
	seen := map[key]bool{}
	check := func(name string, offers []UpgradeOffer, base, floor, discount int) {
		for _, o := range offers {
			if o.Enchant {
				continue // зачарования проверяет enchant_test.go
			}
			k := key{o.HeroIdx, o.Slot}
			if seen[k] {
				t.Errorf("%s: герой %d, слот %v предложен обеим мастерским", name, o.HeroIdx, o.Slot)
			}
			seen[k] = true

			it := m.Party[o.HeroIdx].GetItemInSlot(o.Slot)
			mat := max(1, it.Material.ValueMult)
			next := it.UpgradeLevel + 1
			want := max(floor*mat, next*next*base*mat-discount)
			if o.Cost != want || o.Level != it.UpgradeLevel || !o.CanAfford {
				t.Errorf("%s: предложение %+v, ожидалась цена %d", name, o, want)
			}
		}
	}
	check("кузница", smithy, 45, 35, 2*18)
	check("кожевник", tannery, 38, 30, 1*14)

	var upgrades []UpgradeOffer
	for _, o := range smithy {
		if !o.Enchant {
			upgrades = append(upgrades, o)
		}
	}
	if len(upgrades) > 0 {
		o := upgrades[0]
		it := m.Party[o.HeroIdx].GetItemInSlot(o.Slot)
		level, gold := it.UpgradeLevel, m.Gold
		forged, spent := m.Stats.UpgradesForged, globalDebugReport.GoldSpentBreakdown[spendForge]

		if !m.ExecuteSmithyUpgrade(o) {
			t.Fatal("улучшение не выполнено")
		}
		if it.UpgradeLevel != level+1 || m.Gold != gold-o.Cost || m.Stats.UpgradesForged != forged+1 ||
			globalDebugReport.GoldSpentBreakdown[spendForge] != spent+o.Cost {
			t.Error("после улучшения уровень, золото или счётчики не сошлись")
		}
	}

	// Недостаток золота и предел улучшения блокируют покупку.
	if len(tannery) > 0 {
		o := tannery[0]
		m.Gold = o.Cost - 1
		if m.ExecuteTanneryUpgrade(o) {
			t.Error("без золота улучшение не должно проходить")
		}
		m.Gold = 1_000_000
		m.Party[o.HeroIdx].GetItemInSlot(o.Slot).UpgradeLevel = maxUpgradeLevel
		if m.ExecuteTanneryUpgrade(o) {
			t.Error("предмет на пределе улучшения не должен улучшаться")
		}
	}
}
