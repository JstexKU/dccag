package main

// ============================================================
// СОБЫТИЯ В КОМНАТАХ (плитка '?')
// ============================================================
//
// В отличие от алтаря или фонтана, события зависят от состава отряда
// и от настроек тактики: взломщик вскрывает тайник, силач выбивает дверь,
// а рискованные сделки принимаются только при включённой «азартности».

type eventKind int

const (
	eventCampfire eventKind = iota
	eventMerchant
	eventIdol
	eventVault
	eventAmbush
)

func (m *Model) pickEventKind() eventKind {
	roll := rng.Intn(100)
	switch {
	case roll < 28:
		return eventCampfire
	case roll < 52:
		return eventMerchant
	case roll < 72:
		return eventIdol
	case roll < 90:
		return eventVault
	}
	if m.Floor >= 3 {
		return eventAmbush
	}
	return eventCampfire
}

func (m *Model) goldMult() float64 {
	if m.Relic != nil {
		return m.Relic.GoldMult
	}
	return 1.0
}

// handleEventTile запускает случайное событие. pos — клетка, на которой оно произошло.
func (m *Model) handleEventTile(pos Point) {
	m.Stats.EventsSeen++
	switch m.pickEventKind() {
	case eventCampfire:
		m.eventCampfire()
	case eventMerchant:
		m.eventMerchant()
	case eventIdol:
		m.eventIdol()
	case eventVault:
		m.eventVault()
	case eventAmbush:
		m.eventAmbush(pos)
	}
}

func (m *Model) eventCampfire() {
	const stressRelief = 25
	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		h.Stress = max(0, h.Stress-stressRelief)
		h.HP = min(h.MaxHP, h.HP+h.MaxHP*15/100)
		if h.Stress < 50 && h.Affliction != AfflictionVirtuous {
			h.Affliction = AfflictionNone
		}
	}
	m.addLog(fountStyle.Render(T(m.Lang, "event.campfire", stressRelief)))
}

func (m *Model) eventMerchant() {
	var buyer *Hero
	for _, h := range m.Party {
		if h.IsDead || !h.HasFreePotionSlot(m.MaxPotionSlots()) {
			continue
		}
		if buyer == nil || len(h.Potions) < len(buyer.Potions) {
			buyer = h
		}
	}
	if buyer == nil {
		m.addLog(subtleStyle.Render(T(m.Lang, "event.merchant_poor")))
		return
	}

	pType := PotionHP
	switch {
	case buyer.Stress >= 60:
		pType = PotionStress
	case buyer.Class == ClassMage || buyer.Class == ClassCleric || buyer.Class == ClassWarlock ||
		buyer.Class == ClassBard || buyer.Class == ClassPaladin:
		pType = PotionMP
	}

	potion := createPotion(pType, SizeMedium, m.Floor)
	price := potion.Cost * 3 / 4
	if m.Gold < price {
		m.addLog(subtleStyle.Render(T(m.Lang, "event.merchant_poor")))
		return
	}

	m.Gold -= price
	buyer.Potions = append(buyer.Potions, &potion)
	m.addLog(potionStyle.Render(T(m.Lang, "event.merchant_buy", getPotionName(potion, m.Lang), price, buyer.DisplayName(m.Lang))))
}

func (m *Model) eventIdol() {
	if !m.Tactics.GambleEvents {
		m.addLog(subtleStyle.Render(T(m.Lang, "event.idol_refuse")))
		return
	}
	h := m.getRandomLivingHero()
	if h == nil {
		return
	}
	name := h.DisplayName(m.Lang)

	if rng.Intn(100) < 65 {
		atk := 2 + rng.Intn(2)
		const stress = 15
		h.BaseAtk += atk
		m.addStress(h, stress)
		m.addLog(altarStyle.Render(T(m.Lang, "event.idol_boon", name, atk, stress)))
		return
	}

	dmg := max(6, h.HP*25/100)
	dmg = max(0, min(dmg, h.HP-1)) // проклятие идола не убивает напрямую
	h.HP -= dmg
	h.Feats.DamageTaken += dmg
	const curseStress = 35
	m.addLog(dangerStyle.Render(T(m.Lang, "event.idol_curse", name, dmg, curseStress)))
	m.addStress(h, curseStress)
}

func (m *Model) eventVault() {
	var opener, brute *Hero
	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		switch h.Class {
		case ClassRogue, ClassRanger:
			if opener == nil {
				opener = h
			}
		case ClassTank, ClassWarrior, ClassPaladin, ClassMonk:
			if brute == nil {
				brute = h
			}
		}
	}

	switch {
	case opener != nil:
		gold := int(float64(40+m.Floor*6+rng.Intn(20)) * m.goldMult())
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold
		opener.AddTreasure()
		opener.RevealSecret()
		m.checkAndAwardTitle(opener)

		lootClass := opener.Class
		if lh := m.getRandomLivingHero(); lh != nil {
			lootClass = lh.Class
		}
		item := generateItemForClass(lootClass, m.Floor+1)
		m.addLog(goldStyle.Render(T(m.Lang, "event.vault_pick", opener.DisplayName(m.Lang), gold, item.DisplayName(m.Lang))))
		m.equipOrBag(item)

	case brute != nil:
		dmg := 8 + m.Floor
		dmg = max(0, min(dmg, brute.HP-1))
		brute.HP -= dmg
		brute.Feats.DamageTaken += dmg
		gold := int(float64(20+m.Floor*4) * m.goldMult())
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold
		m.addLog(goldStyle.Render(T(m.Lang, "event.vault_smash", brute.DisplayName(m.Lang), gold, dmg)))

	default:
		m.addLog(subtleStyle.Render(T(m.Lang, "event.vault_locked")))
	}
}

func (m *Model) eventAmbush(pos Point) {
	m.addLog(dangerStyle.Render(T(m.Lang, "event.ambush")))
	m.startCombat(pos, spawnMonsterPack(false, m.Floor))
}
