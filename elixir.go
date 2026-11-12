package main

// ============================================================
// ЭЛИКСИР УСЕРДИЯ
// ============================================================
//
// Эликсир увеличивает опыт героя до конца вылазки (то есть до возвращения в город).
// Больше всего он даёт новичкам, отставшим от отряда, и стоит им дешевле всего: так отряд
// не вынужден часами сидеть на тренировочных этажах ради одного бойца. Опытным героям
// эликсир тоже доступен, но прибавка меньше, а цена выше.

const (
	elixirBaseCost     = 120 // цена эликсира при нулевом уровне героя, G
	elixirCostPerLevel = 30  // плюс столько за каждый уровень героя
)

// targetLevel — уровень, к которому подтягиваются отстающие: не ниже рекомендуемого для
// самого глубокого этажа и не ниже среднего уровня живого отряда.
func (m *Model) targetLevel() int {
	rec := recommendedLevel(m.MaxFloor)
	totalLvl, livingCount := 0, 0
	for _, h := range m.Party {
		if !h.IsDead && !h.LostInAbyss {
			totalLvl += h.Level
			livingCount++
		}
	}
	avgLvl := rec
	if livingCount > 0 {
		avgLvl = totalLvl / livingCount
	}
	return max(rec, avgLvl)
}

// elixirTier — прибавка опыта и множитель цены (оба в процентах) для героя, отстающего
// от целевого уровня на gap (отрицательное значение — герой опережает отряд).
func elixirTier(gap int) (bonusPct, priceMultPct int) {
	switch {
	case gap >= 4:
		return 150, 100
	case gap == 3:
		return 100, 100
	case gap == 2:
		return 80, 100
	case gap == 1:
		return 50, 130
	case gap == 0:
		return 25, 180
	}
	return 15, 250
}

// elixirCost — цена эликсира для героя уровня level при данном отставании.
func elixirCost(level, gap int) (cost, bonusPct int) {
	bonusPct, mult := elixirTier(gap)
	return (elixirBaseCost + elixirCostPerLevel*level) * mult / 100, bonusPct
}

// elixirOffer — предложение алхимика для героя; герою с действующим эликсиром оно не нужно.
func (m *Model) elixirOffer(idx int, h *Hero) AlchemistOffer {
	gap := m.targetLevel() - h.Level
	cost, bonus := elixirCost(h.Level, gap)
	return AlchemistOffer{
		Type:      AlchemistActionElixir,
		HeroIdx:   idx,
		HeroName:  h.DisplayName(m.Lang),
		ItemTitle: T(m.Lang, "elixir.exp.title", bonus),
		Cost:      cost,
		CanAfford: m.Gold >= cost,
		ExpPct:    bonus,
		Gap:       gap,
	}
}

// pickAutoElixir — какой эликсир купит автопилот: только отстающим (gap ≥ 1), самому
// слабому первым, в пределах двойного бюджета алхимика (это прямое средство против
// долгих тренировок). Возвращает индекс предложения или -1.
func (m *Model) pickAutoElixir(offers []AlchemistOffer) int {
	budget := AllocateBudget(m.Gold).Alchemy * 2
	best, bestLevel := -1, 1<<30
	for i, off := range offers {
		if off.Type != AlchemistActionElixir || !off.CanAfford || off.Cost > budget || off.Gap < 1 {
			continue
		}
		if off.HeroIdx < 0 || off.HeroIdx >= len(m.Party) {
			continue
		}
		if lvl := m.Party[off.HeroIdx].Level; lvl < bestLevel {
			best, bestLevel = i, lvl
		}
	}
	return best
}

// clearTripEffects снимает всё, что действует «до конца вылазки»: эликсир и боевые состояния.
func (m *Model) clearTripEffects() {
	for _, h := range m.Party {
		h.ExpElixirPct = 0
	}
	m.clearCombatStatuses()
}
