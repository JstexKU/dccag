package main

import (
	"fmt"
)

// ============================================================
// 7. ЛАВКА АЛХИМИКА (ALCHEMIST)
// ============================================================

type AlchemistActionType int

const (
	AlchemistActionMutation AlchemistActionType = iota
	AlchemistActionPotion
	AlchemistActionElixir // Эликсир Усердия: опыт до конца вылазки
)

type AlchemistOffer struct {
	Type      AlchemistActionType
	HeroIdx   int
	HeroName  string
	ItemTitle string
	Cost      int
	CanAfford bool
	MutType   MutationType
	PotType   PotionType
	ExpPct    int // эликсир: прибавка опыта, %
	Gap       int // эликсир: на сколько уровней герой отстаёт от целевого (отрицательное — опережает)
}

type AlchemistServiceState struct {
	Cursor        int
	Offers        []AlchemistOffer
	ActionSummary string
	ChosenIdx     int
}

func (m *Model) GenerateAlchemistOffers() []AlchemistOffer {
	var offers []AlchemistOffer

	for idx, h := range m.Party {
		if h.IsDead || h.IsDowned {
			continue
		}
		pref := GetClassMutationPreference(h, m.Floor, m.Stats.FallenHeroes)
		baseCost := 240
		switch pref {
		case MutChimera:
			baseCost = 280
		case MutBastion:
			baseCost = 260
		case MutFury:
			baseCost = 240
		case MutTitan:
			baseCost = 230
		case MutAether:
			baseCost = 220
		}
		cost := CalculateMutationCost(baseCost, m.Floor, h.Mutations.Total())
		offers = append(offers, AlchemistOffer{
			Type:      AlchemistActionMutation,
			HeroIdx:   idx,
			HeroName:  h.DisplayName(m.Lang),
			ItemTitle: T(m.Lang, "mut."+string(pref)+".name"),
			Cost:      cost,
			CanAfford: m.Gold >= cost,
			MutType:   pref,
		})

		maxPots := m.MaxPotionSlots()
		if h.HasFreePotionSlot(maxPots) {
			pType := PotionHP
			if h.Stress > 30 || h.Affliction != AfflictionNone {
				pType = PotionStress
			} else if h.Class == ClassMage || h.Class == ClassCleric || h.Class == ClassWarlock || h.Class == ClassBard {
				pType = PotionMP
			}
			cand := createPotion(pType, SizeSmall, m.Floor)
			offers = append(offers, AlchemistOffer{
				Type:      AlchemistActionPotion,
				HeroIdx:   idx,
				HeroName:  h.DisplayName(m.Lang),
				ItemTitle: fmt.Sprintf("%s %s", cand.Symbol, getPotionName(cand, m.Lang)),
				Cost:      cand.Cost,
				CanAfford: m.Gold >= cand.Cost,
				PotType:   pType,
			})
		}

		if h.ExpElixirPct == 0 {
			offers = append(offers, m.elixirOffer(idx, h))
		}
	}
	return offers
}

func (m *Model) ExecuteAlchemistOffer(offer AlchemistOffer) bool {
	if m.Gold < offer.Cost || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	if offer.Type == AlchemistActionElixir && h.ExpElixirPct > 0 {
		return false // эликсир уже действует
	}
	m.Gold -= offer.Cost
	globalDebugReport.GoldSpentBreakdown[spendAlchemy] += offer.Cost

	if offer.Type == AlchemistActionMutation {
		verbApplied := TVerb(m.Lang, h.Gender, "take")
		hName := h.DisplayName(m.Lang)

		switch offer.MutType {
		case MutChimera:
			h.Mutations.ChimeraCount++
			h.MaxHP += 8
			h.HP += 8
			h.MaxMP += 5
			h.MP += 5
			h.BaseAtk += 1
			h.BaseDef += 1
			m.addLog(accentStyle.Render(T(m.Lang, "town.log.mut_chimera", hName, verbApplied, T(m.Lang, "mut.chimera.name"), offer.Cost)))
		case MutFury:
			h.Mutations.FuryCount++
			h.BaseAtk += 3
			h.MaxHP += 2
			h.HP += 2
			m.addLog(fireStyle.Render(T(m.Lang, "town.log.mut_fury", hName, verbApplied, T(m.Lang, "mut.fury.name"), offer.Cost)))
		case MutTitan:
			h.Mutations.TitanCount++
			h.MaxHP += 20
			h.HP += 20
			m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_titan", hName, verbApplied, T(m.Lang, "mut.titan.name"), offer.Cost)))
		case MutAether:
			h.Mutations.AetherCount++
			h.MaxMP += 14
			h.MP += 14
			h.BaseAtk += 1
			m.addLog(fountStyle.Render(T(m.Lang, "town.log.mut_aether", hName, verbApplied, T(m.Lang, "mut.aether.name"), offer.Cost)))
		case MutBastion:
			h.Mutations.BastionCount++
			h.BaseDef += 2
			h.MaxHP += 6
			h.HP += 6
			m.addLog(healStyle.Render(T(m.Lang, "town.log.mut_bastion", hName, verbApplied, T(m.Lang, "mut.bastion.name"), offer.Cost)))
		}
		return true
	}

	if offer.Type == AlchemistActionElixir {
		h.ExpElixirPct = offer.ExpPct
		m.addLog(potionStyle.Render(T(m.Lang, "town.log.elixir_drunk", h.DisplayName(m.Lang), offer.ExpPct, offer.Cost)))
		m.logTownAction("📚", T(m.Lang, m.TownEst.AlchemistKey), T(m.Lang, "town.log.elixir_hist", offer.ExpPct, offer.Cost))
		return true
	}

	if offer.Type == AlchemistActionPotion {
		pot := createPotion(offer.PotType, SizeSmall, m.Floor)
		h.Potions = append(h.Potions, &pot)
		m.logTownAction("🧪", T(m.Lang, m.TownEst.AlchemistKey), T(m.Lang, "town.log.alch_hist", offer.Cost))
		return true
	}
	return false
}
