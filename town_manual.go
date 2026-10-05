package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) initMarketScreen() {
	m.MarketState = MarketServiceState{
		Section: MarketSectionLoot,
		LootIdx: 0,
		ShopIdx: 0,
		Offers:  m.GenerateMarketOffers(),
	}
}

func (m *Model) initMagistrateScreen() {
	m.MagistrateState = MagistrateServiceState{
		Cursor: 0,
		Offers: m.GenerateMagistrateOffers(),
	}
}

func (m *Model) initChurchScreen() {
	m.ChurchState = ChurchServiceState{
		Cursor: 0,
		Offers: m.GenerateChurchOffers(),
	}
}

func (m *Model) initTavernScreen() {
	m.TavernState = TavernServiceState{
		Cursor: 0,
		Offers: m.GenerateTavernOffers(),
	}
}

func (m *Model) initGuildScreen() {
	m.GuildState = GuildServiceState{
		Cursor: 0,
		Offers: m.GenerateGuildOffers(),
	}
}

func (m *Model) initSmithyScreen() {
	m.SmithyState = ForgeServiceState{
		Cursor: 0,
		Offers: m.GenerateSmithyOffers(),
	}
}

func (m *Model) initTanneryScreen() {
	m.TanneryState = ForgeServiceState{
		Cursor: 0,
		Offers: m.GenerateTanneryOffers(),
	}
}

func (m *Model) initAlchemistScreen() {
	m.AlchemistState = AlchemistServiceState{
		Cursor: 0,
		Offers: m.GenerateAlchemistOffers(),
	}
}

func (m *Model) handleMarketKey(key string) (tea.Cmd, bool) {
	st := &m.MarketState
	switch key {
	case "tab":
		if st.Section == MarketSectionLoot {
			st.Section = MarketSectionShop
		} else {
			st.Section = MarketSectionLoot
		}
		return nil, true
	case "up", "k":
		if st.Section == MarketSectionLoot {
			maxIdx := len(m.Bag)
			if maxIdx > 0 {
				st.LootIdx = (st.LootIdx + maxIdx + 1) % (maxIdx + 1)
			}
		} else if len(st.Offers) > 0 {
			st.ShopIdx = (st.ShopIdx + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if st.Section == MarketSectionLoot {
			maxIdx := len(m.Bag)
			if maxIdx > 0 {
				st.LootIdx = (st.LootIdx + 1) % (maxIdx + 1)
			}
		} else if len(st.Offers) > 0 {
			st.ShopIdx = (st.ShopIdx + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if st.Section == MarketSectionLoot {
			if st.LootIdx == len(m.Bag) {
				m.ExecuteMarketSellAll()
				st.LootIdx = 0
			} else if len(m.Bag) > 0 && st.LootIdx < len(m.Bag) {
				m.ExecuteMarketSellLoot(st.LootIdx)
				if st.LootIdx >= len(m.Bag) && st.LootIdx > 0 {
					st.LootIdx = len(m.Bag) - 1
				}
			}
		} else if len(st.Offers) > 0 && st.ShopIdx < len(st.Offers) {
			if m.ExecuteMarketBuy(st.Offers[st.ShopIdx]) {
				st.Offers = m.GenerateMarketOffers()
				if st.ShopIdx >= len(st.Offers) && len(st.Offers) > 0 {
					st.ShopIdx = len(st.Offers) - 1
				}
			}
		}
		return nil, true
	case " ":
		m.ExecuteMarketSellAll()
		st.LootIdx = 0
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseMagistrate
		m.initMagistrateScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleMagistrateKey(key string) (tea.Cmd, bool) {
	st := &m.MagistrateState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteMagistrateInvest(st.Offers[st.Cursor]) {
				st.Offers = m.GenerateMagistrateOffers()
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseChurch
		m.initChurchScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleChurchKey(key string) (tea.Cmd, bool) {
	st := &m.ChurchState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteChurchAction(st.Offers[st.Cursor]) {
				st.Offers = m.GenerateChurchOffers()
				if st.Cursor >= len(st.Offers) && len(st.Offers) > 0 {
					st.Cursor = len(st.Offers) - 1
				}
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseTavern
		m.initTavernScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleTavernKey(key string) (tea.Cmd, bool) {
	st := &m.TavernState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteTavernRest(st.Offers[st.Cursor]) {
				m.TownPhase = TownPhaseGuild
				m.initGuildScreen()
				return nil, true
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseGuild
		m.initGuildScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleGuildKey(key string) (tea.Cmd, bool) {
	st := &m.GuildState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteGuildRecruit(st.Offers[st.Cursor]) {
				st.Offers = m.GenerateGuildOffers()
				if st.Cursor >= len(st.Offers) && len(st.Offers) > 0 {
					st.Cursor = len(st.Offers) - 1
				}
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseSmithy
		m.initSmithyScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleSmithyKey(key string) (tea.Cmd, bool) {
	st := &m.SmithyState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteSmithyUpgrade(st.Offers[st.Cursor]) {
				st.Offers = m.GenerateSmithyOffers()
				if st.Cursor >= len(st.Offers) && len(st.Offers) > 0 {
					st.Cursor = len(st.Offers) - 1
				}
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseTannery
		m.initTanneryScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleTanneryKey(key string) (tea.Cmd, bool) {
	st := &m.TanneryState
	hasBag := m.BagLevel < len(bagUpgrades)-1
	totalRows := len(st.Offers)
	if hasBag {
		totalRows++
	}

	switch key {
	case "up", "k":
		if totalRows > 0 {
			st.Cursor = (st.Cursor + totalRows - 1) % totalRows
		}
		return nil, true
	case "down", "j":
		if totalRows > 0 {
			st.Cursor = (st.Cursor + 1) % totalRows
		}
		return nil, true
	case "enter":
		if hasBag && st.Cursor == 0 {
			if m.ExecuteBagUpgrade() {
				st.Offers = m.GenerateTanneryOffers()
			}
			return nil, true
		}
		offerIdx := st.Cursor
		if hasBag {
			offerIdx--
		}
		if offerIdx >= 0 && offerIdx < len(st.Offers) {
			if m.ExecuteTanneryUpgrade(st.Offers[offerIdx]) {
				st.Offers = m.GenerateTanneryOffers()
				if st.Cursor >= len(st.Offers)+1 && len(st.Offers) > 0 {
					st.Cursor = len(st.Offers)
				}
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseAlchemist
		m.initAlchemistScreen()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleAlchemistKey(key string) (tea.Cmd, bool) {
	st := &m.AlchemistState
	switch key {
	case "up", "k":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + len(st.Offers) - 1) % len(st.Offers)
		}
		return nil, true
	case "down", "j":
		if len(st.Offers) > 0 {
			st.Cursor = (st.Cursor + 1) % len(st.Offers)
		}
		return nil, true
	case "enter":
		if len(st.Offers) > 0 && st.Cursor < len(st.Offers) {
			if m.ExecuteAlchemistOffer(st.Offers[st.Cursor]) {
				st.Offers = m.GenerateAlchemistOffers()
				if st.Cursor >= len(st.Offers) && len(st.Offers) > 0 {
					st.Cursor = len(st.Offers) - 1
				}
			}
		}
		return nil, true
	case "esc":
		m.TownPhase = TownPhaseDepart
		m.stepTown()
		return nil, true
	}
	return nil, false
}
