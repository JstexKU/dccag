package main

import "strconv"

// ============================================================
// ЗАЧАРОВАНИЕ ОРУЖИЯ (КУЗНИЦА)
// ============================================================
//
// Кузнец зачаровывает оружие любого класса: пламя, яд, лёд, молния, бронебой. Эффект
// срабатывает на каждом обычном ударе оружием. Виды открываются с ростом кузницы, ступень
// ограничена уровнем кузницы, цена высокая и растёт со ступенью и материалом клинка.
// Оружие со стихийным префиксом несёт слабую версию той же стихии «от рождения».

type EnchantKind string

const (
	EnchantFire      EnchantKind = "fire"
	EnchantPoison    EnchantKind = "poison"
	EnchantFrost     EnchantKind = "frost"
	EnchantLightning EnchantKind = "lightning"
	EnchantPiercing  EnchantKind = "piercing"
)

// Enchant — зачарование на предмете. Структуру не меняют на месте: при улучшении
// на предмет ставится новая (копии предмета в сумке делят указатель).
type Enchant struct {
	Kind  EnchantKind
	Level int
}

const (
	maxEnchantLevel     = 3
	monsterPoisonTurns  = 3  // сколько ходов монстра действует яд клинка
	monsterChillTurns   = 2  // сколько ходов монстра действует холод клинка
	chillDamagePenalty  = 30 // на сколько холод ослабляет удары монстра, %
	innatePrefixPercent = 60 // сила стихийного префикса относительно зачарования 1-й ступени, %
	replaceSurchargePct = 50 // надбавка за смену вида зачарования, %
	enchantDiscountStep = 5  // скидка за каждый уровень кузницы, %
	enchantDiscountCap  = 25
)

// Порядок видов — порядок их открытия с ростом кузницы.
var enchantKinds = []EnchantKind{EnchantFire, EnchantPoison, EnchantFrost, EnchantLightning, EnchantPiercing}

// Цена зачарования по ступеням (для железного клинка).
var enchantBaseCost = [maxEnchantLevel + 1]int{0, 350, 700, 1300}

// Сила эффекта по ступеням, % от характеристики оружия (у бронебоя — доля игнорируемой защиты).
var enchantPct = map[EnchantKind][maxEnchantLevel + 1]int{
	EnchantFire:      {0, 30, 45, 60},
	EnchantPoison:    {0, 10, 15, 20},
	EnchantFrost:     {0, 15, 22, 30},
	EnchantLightning: {0, 40, 55, 70},
	EnchantPiercing:  {0, 40, 55, 70},
}

func enchantIcon(k EnchantKind) string {
	switch k {
	case EnchantFire:
		return "🔥"
	case EnchantPoison:
		return "☣️"
	case EnchantFrost:
		return "❄️"
	case EnchantLightning:
		return "⚡"
	case EnchantPiercing:
		return "🗡️"
	}
	return ""
}

// enchantLabel — значок и название вида зачарования на языке игры.
func enchantLabel(lang Language, k EnchantKind) string {
	return enchantIcon(k) + " " + T(lang, "enchant."+string(k))
}

// Tag — короткая метка для названия предмета: значок и ступень.
func (e *Enchant) Tag() string {
	return enchantIcon(e.Kind) + strconv.Itoa(e.Level)
}

func kindIn(list []EnchantKind, k EnchantKind) bool {
	for _, x := range list {
		if x == k {
			return true
		}
	}
	return false
}

// ---------- Доступность и цена ----------

// enchantKindsAvailable — виды зачарования, открытые текущим уровнем кузницы (минимум два).
func (m *Model) enchantKindsAvailable() []EnchantKind {
	n := min(len(enchantKinds), 2+max(0, m.Legacy.SmithyLevel))
	return enchantKinds[:n]
}

// enchantLevelCap — высшая доступная ступень зачарования: 1 + уровень кузницы, не выше III.
func (m *Model) enchantLevelCap() int {
	return min(maxEnchantLevel, 1+max(0, m.Legacy.SmithyLevel))
}

// enchantCost — цена зачарования до ступени newLevel. Растёт со ступенью и материалом клинка,
// смена вида дороже на replaceSurchargePct, а развитая кузница даёт скидку.
func (m *Model) enchantCost(it *EquipItem, newLevel int, replacing bool) int {
	level := max(1, min(maxEnchantLevel, newLevel))
	mat := max(1, it.Material.ValueMult)
	cost := enchantBaseCost[level] * (3 + mat) / 4
	if replacing {
		cost += cost * replaceSurchargePct / 100
	}
	discount := min(enchantDiscountCap, enchantDiscountStep*max(0, m.Legacy.SmithyLevel))
	return cost * (100 - discount) / 100
}

// preferredEnchantKind — какое зачарование выбирает автопилот для класса героя.
func preferredEnchantKind(c HeroClass) EnchantKind {
	switch c {
	case ClassWarrior, ClassCleric:
		return EnchantFire
	case ClassRogue, ClassWarlock:
		return EnchantPoison
	case ClassMage, ClassBard:
		return EnchantFrost
	case ClassPaladin, ClassMonk:
		return EnchantLightning
	}
	return EnchantPiercing // танк, следопыт и прочие
}

// enchantKindFor — вид зачарования, который предлагается герою: в ручном режиме тот, что
// выбрал игрок, в автопилоте — любимый для класса (если уже открыт).
func (m *Model) enchantKindFor(h *Hero) EnchantKind {
	avail := m.enchantKindsAvailable()
	want := preferredEnchantKind(h.Class)
	if m.ManualMode {
		want = m.SmithyState.EnchantKind
	}
	if kindIn(avail, want) {
		return want
	}
	return avail[0]
}

// generateEnchantOffers — по одному предложению на оружие каждого живого героя.
func (m *Model) generateEnchantOffers() []UpgradeOffer {
	var offers []UpgradeOffer
	capLevel := m.enchantLevelCap()
	for hIdx, h := range m.Party {
		if h.IsDead || h.IsDowned || h.Weapon == nil {
			continue
		}
		it := h.Weapon
		kind := m.enchantKindFor(h)

		newLevel, replaces := 1, ""
		if cur := it.Enchant; cur != nil {
			if cur.Kind == kind {
				newLevel = cur.Level + 1
			} else {
				replaces = cur.Tag()
			}
		}
		if newLevel > capLevel {
			continue
		}

		cost := m.enchantCost(it, newLevel, replaces != "")
		offers = append(offers, UpgradeOffer{
			HeroIdx:      hIdx,
			Slot:         SlotWeapon,
			ItemName:     it.DisplayName(m.Lang),
			Level:        newLevel - 1,
			Cost:         cost,
			CanAfford:    m.Gold >= cost,
			Enchant:      true,
			EnchantKind:  kind,
			EnchantLevel: newLevel,
			Replaces:     replaces,
		})
	}
	return offers
}

// executeEnchant проводит покупку зачарования; устаревшее или недоступное предложение отклоняется.
func (m *Model) executeEnchant(offer UpgradeOffer) bool {
	if !offer.Enchant || offer.HeroIdx < 0 || offer.HeroIdx >= len(m.Party) {
		return false
	}
	h := m.Party[offer.HeroIdx]
	it := h.Weapon
	if it == nil || h.IsDead || h.IsDowned || m.Gold < offer.Cost {
		return false
	}
	if !kindIn(m.enchantKindsAvailable(), offer.EnchantKind) {
		return false
	}

	wantLevel, replacing := 1, false
	if cur := it.Enchant; cur != nil {
		if cur.Kind == offer.EnchantKind {
			wantLevel = cur.Level + 1
		} else {
			replacing = true
		}
	}
	if offer.EnchantLevel != wantLevel || wantLevel > m.enchantLevelCap() ||
		offer.Cost != m.enchantCost(it, wantLevel, replacing) {
		return false
	}

	m.Gold -= offer.Cost
	it.Enchant = &Enchant{Kind: offer.EnchantKind, Level: offer.EnchantLevel}
	m.Stats.UpgradesForged++
	globalDebugReport.GoldSpentBreakdown[spendForge] += offer.Cost

	smithyName := T(m.Lang, m.TownEst.SmithyKey)
	kindName := T(m.Lang, "enchant."+string(offer.EnchantKind))
	m.addLog(goldStyle.Render(T(m.Lang, "town.log.enchant_done", smithyName, it.DisplayName(m.Lang), kindName, offer.EnchantLevel, offer.Cost)))
	m.logTownAction("✨", smithyName, T(m.Lang, "town.log.enchant_hist", kindName, offer.EnchantLevel, it.DisplayName(m.Lang), offer.Cost))
	return true
}

// smithyOfferWithinBudget — укладывается ли предложение в бюджет автопилота: зачарование
// редкая и дорогая покупка, ему отводится двойной бюджет кузницы.
func smithyOfferWithinBudget(off UpgradeOffer, forgeBudget int) bool {
	if off.Enchant {
		return off.Cost <= forgeBudget*2
	}
	return off.Cost <= forgeBudget
}

// cycleEnchantKind переключает вид зачарования в ручной кузнице (dir = ±1).
func (m *Model) cycleEnchantKind(dir int) {
	st := &m.SmithyState
	avail := m.enchantKindsAvailable()
	cur := 0
	for i, k := range avail {
		if k == st.EnchantKind {
			cur = i
		}
	}
	st.EnchantKind = avail[(cur+dir+len(avail))%len(avail)]
	st.Offers = m.GenerateSmithyOffers()
	if st.Cursor >= len(st.Offers) {
		st.Cursor = max(0, len(st.Offers)-1)
	}
}

// ---------- Эффекты в бою ----------

// weaponEffect — действующий эффект оружия: зачарование или слабая стихия префикса.
type weaponEffect struct {
	Kind   EnchantKind
	Level  int
	Innate bool
}

// pct — сила эффекта, % от характеристики оружия.
func (e weaponEffect) pct() int {
	level := max(1, min(maxEnchantLevel, e.Level))
	p := enchantPct[e.Kind][level]
	if e.Innate {
		p = p * innatePrefixPercent / 100
	}
	return p
}

func prefixEnchantKind(e ElementType) (EnchantKind, bool) {
	switch e {
	case ElemFire:
		return EnchantFire, true
	case ElemPoison:
		return EnchantPoison, true
	case ElemFrost:
		return EnchantFrost, true
	case ElemLightning:
		return EnchantLightning, true
	}
	return "", false
}

// weaponEffects — эффекты оружия героя: зачарование и стихия префикса (если не совпадает с зачарованием).
func (h *Hero) weaponEffects() []weaponEffect {
	w := h.Weapon
	if w == nil {
		return nil
	}
	var out []weaponEffect
	if w.Enchant != nil {
		out = append(out, weaponEffect{Kind: w.Enchant.Kind, Level: w.Enchant.Level})
	}
	if w.Prefix != nil {
		if kind, ok := prefixEnchantKind(w.Prefix.Element); ok && (w.Enchant == nil || w.Enchant.Kind != kind) {
			out = append(out, weaponEffect{Kind: kind, Level: 1, Innate: true})
		}
	}
	return out
}

// enchantPierce — сколько защиты монстра игнорирует удар (бронебой).
func (h *Hero) enchantPierce(mob *Monster) int {
	best := 0
	for _, ef := range h.weaponEffects() {
		if ef.Kind == EnchantPiercing {
			best = max(best, ef.pct())
		}
	}
	return mob.Defense * best / 100
}

// otherLivingMonster — другой живой монстр текущего боя (для цепной молнии).
func (m *Model) otherLivingMonster(mob *Monster) *Monster {
	if m.Combat == nil || m.Combat.Pack == nil {
		return nil
	}
	for _, other := range m.Combat.Pack.Members {
		if other != mob && !other.IsDead && other.HP > 0 {
			return other
		}
	}
	return nil
}

// enchantStrike — дополнительный урон зачарования по монстру.
func (m *Model) enchantStrike(h *Hero, mob *Monster, amount int) {
	mob.HP -= amount
	h.Feats.DamageDealt += amount
}

// applyWeaponEffects — эффекты зачарования после обычного удара героя по монстру.
// Смерть цели от дополнительного урона обрабатывает вызывающий код (он проверяет HP после вызова),
// а смерть соседа от молнии — сам эффект.
func (m *Model) applyWeaponEffects(h *Hero, mob *Monster, dmg int) {
	w := h.Weapon
	if w == nil || mob == nil {
		return
	}
	stat := w.TotalStat()
	hName := h.DisplayName(m.Lang)

	for _, ef := range h.weaponEffects() {
		if mob.HP <= 0 {
			return
		}
		amount := max(1, stat*ef.pct()/100)

		switch ef.Kind {
		case EnchantFire:
			if mob.Affix == AffixFrost {
				amount = amount * 3 / 2
			}
			m.enchantStrike(h, mob, amount)
			m.addLog(fireStyle.Render(T(m.Lang, "combat.log.ench_fire", hName, amount, m.monsterName(mob))))

		case EnchantFrost:
			if mob.Affix == AffixFire {
				amount = amount * 3 / 2
			}
			m.enchantStrike(h, mob, amount)
			mob.ChilledTurns = monsterChillTurns
			m.addLog(fountStyle.Render(T(m.Lang, "combat.log.ench_frost", hName, m.monsterName(mob), amount)))

		case EnchantPoison:
			mob.PoisonTurns = monsterPoisonTurns
			mob.PoisonDmg = amount
			mob.PoisonBy = h
			m.addLog(stressStyle.Render(T(m.Lang, "combat.log.ench_poison", hName, m.monsterName(mob), amount)))

		case EnchantLightning:
			if other := m.otherLivingMonster(mob); other != nil {
				other.HP -= amount
				h.Feats.DamageDealt += amount
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.ench_lightning", hName, m.monsterName(other), amount)))
				if other.HP <= 0 {
					m.onMonsterKilled(h, other)
				}
			} else {
				shock := max(1, amount/2)
				m.enchantStrike(h, mob, shock)
				m.addLog(accentStyle.Render(T(m.Lang, "combat.log.ench_shock", hName, m.monsterName(mob), shock)))
			}
		}
	}
}

// tickMonsterStatuses — начало хода монстра: яд клинка снимает здоровье, холод ослабляет удар.
// Возвращает true, если монстр погиб от яда (его ход на этом заканчивается).
func (m *Model) tickMonsterStatuses(mob *Monster) bool {
	mob.ChillWeak = mob.ChilledTurns > 0
	if mob.ChilledTurns > 0 {
		mob.ChilledTurns--
	}

	if mob.PoisonTurns > 0 {
		mob.PoisonTurns--
		dmg := max(1, mob.PoisonDmg)
		mob.HP -= dmg
		m.addLog(stressStyle.Render(T(m.Lang, "combat.log.ench_poison_tick", m.monsterName(mob), dmg)))

		if mob.HP <= 0 {
			owner := mob.PoisonBy
			if owner == nil {
				owner = m.getRandomLivingHero()
			}
			mob.PoisonTurns, mob.PoisonDmg, mob.PoisonBy = 0, 0, nil
			if owner != nil {
				m.onMonsterKilled(owner, mob)
			} else {
				mob.HP, mob.IsDead = 0, true
			}
			return true
		}
		if mob.PoisonTurns == 0 {
			mob.PoisonDmg, mob.PoisonBy = 0, nil
		}
	}
	return false
}
