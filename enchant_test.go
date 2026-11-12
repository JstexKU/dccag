package main

import "testing"

func enchantModel(t *testing.T) Model {
	t.Helper()
	seedRNG(81)
	m := initialModel()
	m.State = StatePlaying
	m.Lang = LangRU
	m.Gold = 1_000_000
	for _, h := range m.Party {
		h.IsDead, h.IsDowned = false, false
	}
	return m
}

// plainWeapon выдаёт герою «чистое» оружие без префикса, суффикса и зачарования.
func plainWeapon(h *Hero) *EquipItem {
	w := generateItemForClassSlot(h.Class, SlotWeapon, 5)
	w.Prefix, w.Suffix, w.Enchant = nil, nil, nil
	h.SetItemInSlot(SlotWeapon, &w)
	return &w
}

func TestEnchantAvailabilityFollowsSmithyLevel(t *testing.T) {
	m := enchantModel(t)
	for level, want := range map[int][2]int{0: {2, 1}, 1: {3, 2}, 2: {4, 3}, 3: {5, 3}, 10: {5, 3}} {
		m.Legacy.SmithyLevel = level
		if got := len(m.enchantKindsAvailable()); got != want[0] {
			t.Errorf("кузница %d: видов %d, ожидалось %d", level, got, want[0])
		}
		if got := m.enchantLevelCap(); got != want[1] {
			t.Errorf("кузница %d: ступень до %d, ожидалось %d", level, got, want[1])
		}
	}
}

func TestEnchantCostIsSteep(t *testing.T) {
	m := enchantModel(t)
	w := plainWeapon(m.Party[0])
	w.Material.ValueMult = 1

	c1, c2, c3 := m.enchantCost(w, 1, false), m.enchantCost(w, 2, false), m.enchantCost(w, 3, false)
	if !(c1 >= 350 && c1 < c2 && c2 < c3) {
		t.Errorf("цены по ступеням %d, %d, %d: должны расти, а первая — не дешевле 350", c1, c2, c3)
	}
	if m.enchantCost(w, 1, true) <= c1 {
		t.Error("смена вида должна стоить дороже нового зачарования")
	}

	rich := *w
	rich.Material.ValueMult = 4
	if m.enchantCost(&rich, 1, false) <= c1 {
		t.Error("дорогой материал должен удорожать зачарование")
	}

	m.Legacy.SmithyLevel = 3
	if m.enchantCost(w, 1, false) >= c1 {
		t.Error("развитая кузница должна давать скидку")
	}
}

func TestEnchantOffers(t *testing.T) {
	m := enchantModel(t)
	for _, h := range m.Party {
		plainWeapon(h)
	}

	offers := m.generateEnchantOffers()
	if len(offers) != len(m.Party) {
		t.Fatalf("предложений %d при %d героях", len(offers), len(m.Party))
	}
	avail := m.enchantKindsAvailable()
	for _, o := range offers {
		w := m.Party[o.HeroIdx].Weapon
		if !o.Enchant || !kindIn(avail, o.EnchantKind) || o.EnchantLevel != 1 || o.Level != 0 || o.Replaces != "" {
			t.Errorf("предложение %+v", o)
		}
		if o.Cost != m.enchantCost(w, 1, false) {
			t.Errorf("цена %d, ожидалась %d", o.Cost, m.enchantCost(w, 1, false))
		}
	}

	// Ручной режим: вид выбирает игрок для всех сразу.
	m.ManualMode = true
	m.SmithyState.EnchantKind = EnchantPoison
	for _, o := range m.generateEnchantOffers() {
		if o.EnchantKind != EnchantPoison {
			t.Errorf("в ручном режиме выбран яд, а предложен %s", o.EnchantKind)
		}
	}
	m.ManualMode = false

	// Другой вид на оружии: предложение заменяет его и стоит дороже.
	h := m.Party[1]
	kind := m.enchantKindFor(h)
	other := EnchantFire
	if kind == EnchantFire {
		other = EnchantPoison
	}
	h.Weapon.Enchant = &Enchant{Kind: other, Level: 1}
	for _, o := range m.generateEnchantOffers() {
		if o.HeroIdx == 1 {
			if o.Replaces == "" || o.EnchantLevel != 1 || o.Cost != m.enchantCost(h.Weapon, 1, true) {
				t.Errorf("замена: %+v", o)
			}
		}
	}

	// Предел ступени: предложения этому герою больше нет.
	h.Weapon.Enchant = &Enchant{Kind: kind, Level: m.enchantLevelCap()}
	for _, o := range m.generateEnchantOffers() {
		if o.HeroIdx == 1 {
			t.Error("на пределе ступени предложения быть не должно")
		}
	}
}

func TestExecuteEnchant(t *testing.T) {
	m := enchantModel(t)
	m.Legacy.SmithyLevel = 1 // ступень до II
	h := m.Party[0]
	w := plainWeapon(h)

	var offer UpgradeOffer
	for _, o := range m.GenerateSmithyOffers() {
		if o.Enchant && o.HeroIdx == 0 {
			offer = o
		}
	}
	if !offer.Enchant {
		t.Fatal("нет предложения зачарования")
	}

	gold := m.Gold
	if !m.ExecuteSmithyUpgrade(offer) {
		t.Fatal("покупка не удалась")
	}
	if w.Enchant == nil || w.Enchant.Kind != offer.EnchantKind || w.Enchant.Level != 1 || m.Gold != gold-offer.Cost {
		t.Errorf("после покупки: %+v, золото %d (ожидалось %d)", w.Enchant, m.Gold, gold-offer.Cost)
	}

	// Устаревшее предложение не списывает золото и не меняет оружие.
	gold = m.Gold
	if m.ExecuteSmithyUpgrade(offer) || m.Gold != gold || w.Enchant.Level != 1 {
		t.Error("устаревшее предложение должно отклоняться")
	}

	// Следующая ступень доступна, а без золота не продаётся.
	var next UpgradeOffer
	for _, o := range m.GenerateSmithyOffers() {
		if o.Enchant && o.HeroIdx == 0 {
			next = o
		}
	}
	if next.EnchantLevel != 2 || next.EnchantKind != offer.EnchantKind {
		t.Fatalf("следующее предложение: %+v", next)
	}
	m.Gold = next.Cost - 1
	if m.ExecuteSmithyUpgrade(next) {
		t.Error("без золота зачарование не продаётся")
	}
	m.Gold = next.Cost
	if !m.ExecuteSmithyUpgrade(next) || w.Enchant.Level != 2 || m.Gold != 0 {
		t.Errorf("вторая ступень: %+v, золото %d", w.Enchant, m.Gold)
	}
}

func TestEnchantedItemNameShowsTag(t *testing.T) {
	m := enchantModel(t)
	w := plainWeapon(m.Party[0])
	plain := w.DisplayName(LangRU)
	w.Enchant = &Enchant{Kind: EnchantLightning, Level: 2}
	if got := w.DisplayName(LangRU); got != plain+" ⚡2" {
		t.Errorf("название %q, ожидалось %q", got, plain+" ⚡2")
	}
}

func combatWith(m *Model, mobs ...*Monster) {
	m.Combat = &ActiveCombat{Pack: &MonsterPack{Members: mobs}}
}

func newMob(hp int, aff MonsterAffix) *Monster {
	return &Monster{NameKey: "mob.rat", HP: hp, MaxHP: hp, Affix: aff}
}

func TestWeaponEffectFlameAndFrost(t *testing.T) {
	m := enchantModel(t)
	h := m.Party[0]
	w := plainWeapon(h)
	stat := w.TotalStat()

	w.Enchant = &Enchant{Kind: EnchantFire, Level: 1}
	mob := newMob(500, AffixNone)
	combatWith(&m, mob)
	m.applyWeaponEffects(h, mob, 10)
	if want := max(1, stat*30/100); mob.HP != 500-want {
		t.Errorf("пламя: HP %d, ожидалось %d", mob.HP, 500-want)
	}
	icy := newMob(500, AffixFrost)
	m.applyWeaponEffects(h, icy, 10)
	if want := max(1, stat*30/100) * 3 / 2; icy.HP != 500-want {
		t.Errorf("пламя по ледяному: HP %d, ожидалось %d", icy.HP, 500-want)
	}

	w.Enchant = &Enchant{Kind: EnchantFrost, Level: 2}
	mob = newMob(500, AffixNone)
	m.applyWeaponEffects(h, mob, 10)
	if want := max(1, stat*22/100); mob.HP != 500-want || mob.ChilledTurns != monsterChillTurns {
		t.Errorf("лёд: HP %d (ждали %d), холод %d", mob.HP, 500-want, mob.ChilledTurns)
	}
	fiery := newMob(500, AffixFire)
	m.applyWeaponEffects(h, fiery, 10)
	if want := max(1, stat*22/100) * 3 / 2; fiery.HP != 500-want {
		t.Errorf("лёд по огненному: HP %d, ожидалось %d", fiery.HP, 500-want)
	}
}

func TestWeaponEffectPoisonAndLightning(t *testing.T) {
	m := enchantModel(t)
	h := m.Party[0]
	w := plainWeapon(h)
	stat := w.TotalStat()

	w.Enchant = &Enchant{Kind: EnchantPoison, Level: 1}
	mob := newMob(500, AffixNone)
	combatWith(&m, mob)
	m.applyWeaponEffects(h, mob, 10)
	if mob.PoisonTurns != monsterPoisonTurns || mob.PoisonDmg != max(1, stat*10/100) || mob.PoisonBy != h || mob.HP != 500 {
		t.Errorf("яд: ходов %d, урон %d, HP %d", mob.PoisonTurns, mob.PoisonDmg, mob.HP)
	}

	w.Enchant = &Enchant{Kind: EnchantLightning, Level: 1}
	target, neighbour := newMob(500, AffixNone), newMob(500, AffixNone)
	combatWith(&m, target, neighbour)
	m.applyWeaponEffects(h, target, 10)
	amount := max(1, stat*40/100)
	if target.HP != 500 || neighbour.HP != 500-amount {
		t.Errorf("молния: цель %d, сосед %d (ждали %d)", target.HP, neighbour.HP, 500-amount)
	}

	// Одиночная цель получает слабый разряд.
	lone := newMob(500, AffixNone)
	combatWith(&m, lone)
	m.applyWeaponEffects(h, lone, 10)
	if want := max(1, amount/2); lone.HP != 500-want {
		t.Errorf("разряд по одиночной цели: HP %d, ожидалось %d", lone.HP, 500-want)
	}

	// Сосед, павший от молнии, засчитывается герою.
	weak := newMob(1, AffixNone)
	combatWith(&m, newMob(500, AffixNone), weak)
	kills := h.Feats.Kills
	m.applyWeaponEffects(h, m.Combat.Pack.Members[0], 10)
	if !weak.IsDead || h.Feats.Kills != kills+1 {
		t.Errorf("павший от молнии: мёртв=%v, убийств %d (было %d)", weak.IsDead, h.Feats.Kills, kills)
	}
}

func TestWeaponEffectSourcesAndPierce(t *testing.T) {
	m := enchantModel(t)
	h := m.Party[0]
	w := plainWeapon(h)
	if len(h.weaponEffects()) != 0 || h.enchantPierce(&Monster{Defense: 50}) != 0 {
		t.Fatal("у чистого оружия эффектов нет")
	}

	// Стихийный префикс — слабая версия: 60% от первой ступени.
	w.Prefix = &PrefixDef{Key: "prefix.fire", Element: ElemFire, Bonus: 3}
	effects := h.weaponEffects()
	if len(effects) != 1 || !effects[0].Innate || effects[0].pct() != 30*innatePrefixPercent/100 {
		t.Errorf("врождённый эффект: %+v", effects)
	}

	// Зачарование той же стихии вытесняет префикс, другой — работает вместе с ним.
	w.Enchant = &Enchant{Kind: EnchantFire, Level: 2}
	if got := h.weaponEffects(); len(got) != 1 || got[0].Innate || got[0].pct() != 45 {
		t.Errorf("зачарование поверх префикса: %+v", got)
	}
	w.Enchant = &Enchant{Kind: EnchantPiercing, Level: 1}
	if got := h.weaponEffects(); len(got) != 2 {
		t.Errorf("зачарование и префикс разных видов: %+v", got)
	}
	if got := h.enchantPierce(&Monster{Defense: 50}); got != 20 {
		t.Errorf("бронебой: игнорируется %d, ожидалось 20", got)
	}
}

func TestMonsterStatusTicks(t *testing.T) {
	m := enchantModel(t)
	h := m.Party[0]

	mob := newMob(20, AffixNone)
	combatWith(&m, mob)
	mob.PoisonTurns, mob.PoisonDmg, mob.PoisonBy = 2, 5, h
	if m.tickMonsterStatuses(mob) || mob.HP != 15 || mob.PoisonTurns != 1 {
		t.Errorf("первый ход яда: HP %d, ходов %d", mob.HP, mob.PoisonTurns)
	}
	if m.tickMonsterStatuses(mob) || mob.HP != 10 || mob.PoisonTurns != 0 || mob.PoisonDmg != 0 || mob.PoisonBy != nil {
		t.Errorf("второй ход яда: HP %d, ходов %d", mob.HP, mob.PoisonTurns)
	}

	dying := newMob(3, AffixNone)
	combatWith(&m, dying)
	dying.PoisonTurns, dying.PoisonDmg, dying.PoisonBy = 3, 5, h
	kills := h.Feats.Kills
	if !m.tickMonsterStatuses(dying) || !dying.IsDead || h.Feats.Kills != kills+1 {
		t.Errorf("смерть от яда: мёртв=%v, убийств %d (было %d)", dying.IsDead, h.Feats.Kills, kills)
	}

	chilled := newMob(100, AffixNone)
	chilled.ChilledTurns = 2
	for i, want := range []bool{true, true, false} {
		m.tickMonsterStatuses(chilled)
		if chilled.ChillWeak != want {
			t.Errorf("ход %d: ослабление %v, ожидалось %v", i+1, chilled.ChillWeak, want)
		}
	}
}

// Автопилот зачаровывает по классу героя, но только в пределах двойного бюджета кузницы.
func TestAutopilotEnchantBudget(t *testing.T) {
	m := enchantModel(t)
	m.Gold = 100000
	for _, h := range m.Party {
		plainWeapon(h)
	}
	budget := AllocateBudget(m.Gold).Forge
	for _, o := range m.generateEnchantOffers() {
		if !smithyOfferWithinBudget(o, budget) {
			t.Errorf("при золоте %d зачарование за %d не влезает в бюджет %d", m.Gold, o.Cost, budget)
		}
		if o.EnchantKind != m.enchantKindFor(m.Party[o.HeroIdx]) {
			t.Error("автопилот должен выбирать вид по классу")
		}
	}

	m.Gold = 800
	budget = AllocateBudget(m.Gold).Forge
	for _, o := range m.generateEnchantOffers() {
		if smithyOfferWithinBudget(o, budget) {
			t.Errorf("при золоте %d зачарование за %d не должно покупаться автопилотом", m.Gold, o.Cost)
		}
	}
	if !smithyOfferWithinBudget(UpgradeOffer{Cost: budget}, budget) || smithyOfferWithinBudget(UpgradeOffer{Cost: budget + 1}, budget) {
		t.Error("обычное улучшение должно укладываться ровно в бюджет")
	}
}
