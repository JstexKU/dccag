package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const LegacyTaxRate = 0.15

var debugReportFlag = flag.Bool("report", false, "Включить сбор подробного отчета телеметрии")

var globalDebugReport = DebugReportData{
	ClassDeaths:        make(map[string]int),
	DeathCauses:        make(map[string]int),
	DeathsByFloor:      make(map[int]int),
	GoldSpentBreakdown: make(map[string]int),
	AverageAtk:         make(map[string]float64),
}

type DebugReportData struct {
	TotalRuns          int                `json:"total_runs"`
	MaxFloorReached    int                `json:"max_floor_reached"`
	TotalGoldSpent     int                `json:"total_gold_spent"`
	GoldSpentBreakdown map[string]int     `json:"gold_spent_breakdown"`
	ClassDeaths        map[string]int     `json:"class_deaths"`
	DeathCauses        map[string]int     `json:"death_causes"`
	DeathsByFloor      map[int]int        `json:"deaths_by_floor"`
	FleeAttempts       int                `json:"flee_attempts"`
	FleeSuccesses      int                `json:"flee_successes"`
	DamageGuarded      int                `json:"damage_guarded_by_tanks"`
	EnrageProcs        int                `json:"enrage_triggered_count"`
	AverageAtk         map[string]float64 `json:"average_atk"`
}

// --- Предметы, Зелья, Мутации ---
type PotionType string

const (
	PotionHP     PotionType = "Зелье HP"
	PotionMP     PotionType = "Зелье MP"
	PotionStress PotionType = "Настойка Рассудка"
)

type PotionSize string

const (
	SizeSmall  PotionSize = "Малое"
	SizeMedium PotionSize = "Среднее"
	SizeLarge  PotionSize = "Большое"
	SizeGrand  PotionSize = "Великое"
)

type Potion struct {
	Type   PotionType
	Size   PotionSize
	Power  int
	Cost   int
	Symbol string
}

type MutationType string

const (
	MutChimera MutationType = "Химера (Все)"
	MutFury    MutationType = "Ярость (Atk)"
	MutTitan   MutationType = "Титан (HP)"
	MutAether  MutationType = "Эфир (MP)"
	MutBastion MutationType = "Бастион (Def)"
)

type HeroMutations struct {
	ChimeraCount int
	FuryCount    int
	TitanCount   int
	AetherCount  int
	BastionCount int
}

func (m HeroMutations) Total() int {
	return m.ChimeraCount + m.FuryCount + m.TitanCount + m.AetherCount + m.BastionCount
}

type EquipSlot string

const (
	SlotWeapon EquipSlot = "Оружие"
	SlotHead   EquipSlot = "Шлем"
	SlotChest  EquipSlot = "Доспех"
	SlotLegs   EquipSlot = "Поножи"
)

type ArmorCategory string

const (
	ArmorHeavy  ArmorCategory = "Тяжелая"
	ArmorMedium ArmorCategory = "Средняя"
	ArmorLight  ArmorCategory = "Легкая"
	ArmorNone   ArmorCategory = "Нет"
)

type MaterialTier struct {
	Name      string
	Adj       string
	BonusMult int
	ValueMult int
}

var MetalMaterials = []MaterialTier{
	{Name: "Железо", Adj: "Железн.", BonusMult: 1, ValueMult: 1},
	{Name: "Сталь", Adj: "Стальн.", BonusMult: 2, ValueMult: 2},
	{Name: "Мифрил", Adj: "Мифрил.", BonusMult: 3, ValueMult: 4},
	{Name: "Адамант", Adj: "Адамант.", BonusMult: 4, ValueMult: 7},
}

var MageWeaponMaterials = []MaterialTier{
	{Name: "Тис", Adj: "Тисов.", BonusMult: 1, ValueMult: 1},
	{Name: "Ясень", Adj: "Ясенев.", BonusMult: 2, ValueMult: 2},
	{Name: "Кристалл", Adj: "Кристальн.", BonusMult: 3, ValueMult: 4},
	{Name: "Астральный эфир", Adj: "Эфирн.", BonusMult: 4, ValueMult: 7},
}

var LeatherMaterials = []MaterialTier{
	{Name: "Сыромятная кожа", Adj: "Сыромятн.", BonusMult: 1, ValueMult: 1},
	{Name: "Варёная кожа", Adj: "Варён.", BonusMult: 2, ValueMult: 2},
	{Name: "Кожа василиска", Adj: "Василиск.", BonusMult: 3, ValueMult: 4},
	{Name: "Шкура дракона", Adj: "Драконь.", BonusMult: 4, ValueMult: 7},
}

var ClothMaterials = []MaterialTier{
	{Name: "Плотный лён", Adj: "Льнян.", BonusMult: 1, ValueMult: 1},
	{Name: "Рунический шёлк", Adj: "Шёлков.", BonusMult: 2, ValueMult: 2},
	{Name: "Астральная парча", Adj: "Парчов.", BonusMult: 3, ValueMult: 4},
	{Name: "Ткань Бездны", Adj: "Эфирн.", BonusMult: 4, ValueMult: 7},
}

type ElementType string

const (
	ElemNone      ElementType = "Нет"
	ElemFire      ElementType = "Огонь"
	ElemPoison    ElementType = "Яд"
	ElemFrost     ElementType = "Мороз"
	ElemLightning ElementType = "Молния"
)

type SuffixType string

const (
	SuffNone      SuffixType = "Нет"
	SuffVampirism SuffixType = "Вампиризм"
	SuffFury      SuffixType = "Ярость"
	SuffMana      SuffixType = "Медитация"
	SuffTitan     SuffixType = "Титан"
)

type PrefixDef struct {
	Name    string
	Element ElementType
	Bonus   int
}

type SuffixDef struct {
	Name   string
	Effect SuffixType
	Bonus  int
}

var Prefixes = []PrefixDef{
	{Name: "Пылающий", Element: ElemFire, Bonus: 3},
	{Name: "Ядовитый", Element: ElemPoison, Bonus: 2},
	{Name: "Леденящий", Element: ElemFrost, Bonus: 2},
	{Name: "Громовой", Element: ElemLightning, Bonus: 4},
	{Name: "Закаленный", Element: ElemNone, Bonus: 2},
}

var Suffixes = []SuffixDef{
	{Name: "Кровопийцы", Effect: SuffVampirism, Bonus: 25},
	{Name: "Ярости", Effect: SuffFury, Bonus: 4},
	{Name: "Медитации", Effect: SuffMana, Bonus: 3},
	{Name: "Титана", Effect: SuffTitan, Bonus: 4},
}

type EquipItem struct {
	BaseName     string
	Slot         EquipSlot
	Category     ArmorCategory
	AllowedClass HeroClass
	Material     MaterialTier
	UpgradeLevel int
	BaseStat     int
	BonusMP      int
	BonusHP      int
	CritBonus    int
	BlockBonus   int
	StressRes    int
	SpeedBonus   int
	Value        int
	Prefix       *PrefixDef
	Suffix       *SuffixDef
}

func (e *EquipItem) TotalStat() int {
	stat := e.BaseStat * e.Material.BonusMult
	stat += e.UpgradeLevel * 2
	if e.Prefix != nil {
		stat += e.Prefix.Bonus
	}
	if e.Suffix != nil && e.Suffix.Effect == SuffTitan {
		stat += e.Suffix.Bonus
	}
	return stat
}

func (e *EquipItem) DisplayName() string {
	var parts []string
	if e.Prefix != nil {
		parts = append(parts, e.Prefix.Name)
	}
	parts = append(parts, e.Material.Adj, e.BaseName)
	if e.UpgradeLevel > 0 {
		parts = append(parts, e.UpgradeLevelStr())
	}
	if e.Suffix != nil {
		parts = append(parts, e.Suffix.Name)
	}
	return joinNonEmpty(parts, " ")
}

func (e *EquipItem) UpgradeLevelStr() string {
	if e.UpgradeLevel > 0 {
		return "+" + string(rune('0'+e.UpgradeLevel))
	}
	return ""
}

func joinNonEmpty(items []string, sep string) string {
	var res []string
	for _, it := range items {
		if it != "" {
			res = append(res, it)
		}
	}
	if len(res) == 0 {
		return ""
	}
	out := res[0]
	for _, s := range res[1:] {
		out += sep + s
	}
	return out
}

// --- Герои и Гендер ---
type Gender int

const (
	GenderMale Gender = iota
	GenderFemale
)

type HeroNameDef struct {
	Name   string
	Gender Gender
}

var HeroNames = []HeroNameDef{
	{Name: "Бранд", Gender: GenderMale},
	{Name: "Торин", Gender: GenderMale},
	{Name: "Лира", Gender: GenderFemale},
	{Name: "Алдос", Gender: GenderMale},
	{Name: "Селина", Gender: GenderFemale},
	{Name: "Рагнар", Gender: GenderMale},
	{Name: "Ингвар", Gender: GenderMale},
	{Name: "Вульф", Gender: GenderMale},
	{Name: "Сигурд", Gender: GenderMale},
	{Name: "Морган", Gender: GenderMale},
	{Name: "Элиас", Gender: GenderMale},
	{Name: "Дункан", Gender: GenderMale},
	{Name: "Готфрид", Gender: GenderMale},
	{Name: "Вальтер", Gender: GenderMale},
	{Name: "Кассиан", Gender: GenderMale},
	{Name: "Айрис", Gender: GenderFemale},
	{Name: "Морриган", Gender: GenderFemale},
	{Name: "Бригитта", Gender: GenderFemale},
	{Name: "Агнес", Gender: GenderFemale},
	{Name: "Хильда", Gender: GenderFemale},
	{Name: "Ярополк", Gender: GenderMale},
	{Name: "Радомир", Gender: GenderMale},
	{Name: "Добрыня", Gender: GenderMale},
	{Name: "Лютобор", Gender: GenderMale},
	{Name: "Бронислав", Gender: GenderMale},
	{Name: "Аэрон", Gender: GenderMale},
	{Name: "Дрейвен", Gender: GenderMale},
	{Name: "Кэлар", Gender: GenderMale},
	{Name: "Зордан", Gender: GenderMale},
	{Name: "Тарион", Gender: GenderMale},
	{Name: "Велдор", Gender: GenderMale},
	{Name: "Фалькон", Gender: GenderMale},
	{Name: "Йоррик", Gender: GenderMale},
	{Name: "Эдан", Gender: GenderMale},
	{Name: "Зарвин", Gender: GenderMale},
	{Name: "Лирианна", Gender: GenderFemale},
	{Name: "Велара", Gender: GenderFemale},
	{Name: "Мираэль", Gender: GenderFemale},
	{Name: "Селестина", Gender: GenderFemale},
	{Name: "Каэлина", Gender: GenderFemale},
	{Name: "Тирианна", Gender: GenderFemale},
	{Name: "Найра", Gender: GenderFemale},
	{Name: "Эльмира", Gender: GenderFemale},
	{Name: "Зейра", Gender: GenderFemale},
	{Name: "Ксандр", Gender: GenderMale},
	{Name: "Верисса", Gender: GenderFemale},
	{Name: "Орвин", Gender: GenderMale},
	{Name: "Сильран", Gender: GenderMale},
	{Name: "Келдра", Gender: GenderFemale},
	{Name: "Вейнара", Gender: GenderFemale},
}

func getRandomHeroName() HeroNameDef {
	return HeroNames[rand.Intn(len(HeroNames))]
}

type HeroClass string

const (
	ClassTank    HeroClass = "Танк"
	ClassWarrior HeroClass = "Воин"
	ClassRogue   HeroClass = "Разбойник"
	ClassMage    HeroClass = "Маг"
	ClassCleric  HeroClass = "Клирик"
)

type CombatRole struct {
	AggroWeight int
	CanGuard    bool
	CanHeal     bool
}

func GetClassRole(class HeroClass) CombatRole {
	switch class {
	case ClassTank:
		return CombatRole{AggroWeight: 70, CanGuard: true, CanHeal: false}
	case ClassWarrior:
		return CombatRole{AggroWeight: 35, CanGuard: true, CanHeal: false}
	case ClassRogue:
		return CombatRole{AggroWeight: 15, CanGuard: false, CanHeal: false}
	case ClassMage:
		return CombatRole{AggroWeight: 20, CanGuard: false, CanHeal: false}
	case ClassCleric:
		return CombatRole{AggroWeight: 20, CanGuard: false, CanHeal: true}
	default:
		return CombatRole{AggroWeight: 25, CanGuard: false, CanHeal: false}
	}
}

type AfflictionType string

const (
	AfflictionNone     AfflictionType = "Спокоен"
	AfflictionParanoid AfflictionType = "Параноик"
	AfflictionSelfish  AfflictionType = "Эгоист"
	AfflictionManiac   AfflictionType = "Безумец"
	AfflictionVirtuous AfflictionType = "Воодушевлен"
)

type HeroHeroics struct {
	DamageDealt      int
	DamageTaken      int
	Kills            int
	BossKills        int
	CritsLanded      int
	HealsGiven       int
	NearDeathEscapes int

	TreasureFound   int
	SecretsRevealed int
	Blocks          int
	AggroPulled     int
	CriticalStrikes int
	Backstabs       int
	LootStolen      int
	CCDuration      int
	ManaBursts      int
	Revives         int
	DoTsRemoved     int
}

type Hero struct {
	Name         string
	Gender       Gender
	Title        string
	Class        HeroClass
	Role         CombatRole
	Level        int
	Exp          int
	MaxHP        int
	HP           int
	MaxMP        int
	MP           int
	Stress       int
	Affliction   AfflictionType
	BaseAtk      int
	BaseDef      int
	Speed        int
	IsDead       bool
	IsGuarding   bool
	IsBerserk    bool
	IsStealthed  bool
	IsCharged    bool
	IsAura       bool
	CauseOfDeath string
	SkillName    string
	SkillCost    int
	Mutations    HeroMutations
	Feats        HeroHeroics

	Weapon *EquipItem
	Head   *EquipItem
	Chest  *EquipItem
	Legs   *EquipItem

	Potions []*Potion
}

func (h *Hero) Verb(male, female string) string {
	if h.Gender == GenderFemale {
		return female
	}
	return male
}

func (h *Hero) ShortClass() string {
	switch h.Class {
	case ClassRogue:
		return "Рога"
	case ClassCleric:
		return "Жрец"
	default:
		return string(h.Class)
	}
}

func (h *Hero) NextLevelExp() int {
	return h.Level*100 + (h.Level * h.Level * 25)
}

func (h *Hero) GainExp(amt int) bool {
	h.Exp += amt
	leveledUp := false
	for h.Exp >= h.NextLevelExp() {
		h.Exp -= h.NextLevelExp()
		h.Level++
		leveledUp = true

		switch h.Class {
		case ClassTank:
			h.MaxHP += 12
			h.MaxMP += 2
			h.BaseAtk += 1
			if h.Level%2 == 0 {
				h.BaseDef += 1
			}
		case ClassWarrior:
			h.MaxHP += 8
			h.MaxMP += 3
			h.BaseAtk += 2
			if h.Level%3 == 0 {
				h.BaseDef += 1
			}
		case ClassRogue:
			h.MaxHP += 5
			h.MaxMP += 4
			h.BaseAtk += 2
			h.Speed += 1
		case ClassMage:
			h.MaxHP += 4
			h.MaxMP += 8
			h.BaseAtk += 3
		case ClassCleric:
			h.MaxHP += 6
			h.MaxMP += 6
			h.BaseAtk += 1
		}

		h.HP = h.MaxHP
		h.MP = h.MaxMP
	}
	return leveledUp
}

func (h *Hero) FullName() string {
	if h.Title != "" {
		return h.Name + " «" + h.Title + "»"
	}
	return h.Name
}

func (h *Hero) TotalAtk() int {
	b := 0
	if h.Weapon != nil {
		b = h.Weapon.TotalStat()
	}
	if h.Affliction == AfflictionVirtuous {
		b += 4
	}
	if h.IsBerserk {
		b += 5
	}
	if h.IsCharged {
		b += 6
	}
	return h.BaseAtk + b
}

func (h *Hero) TotalDef() int {
	b := 0
	slots := []*EquipItem{h.Weapon, h.Head, h.Chest, h.Legs}
	for _, it := range slots {
		if it != nil {
			b += it.TotalStat()
			b += it.BlockBonus
		}
	}
	if h.IsGuarding {
		b += 5
	}
	if h.IsAura {
		b += 3
	}
	if h.IsBerserk {
		b -= 2
	}
	if b < 0 {
		b = 0
	}
	return h.BaseDef + b
}

func (h *Hero) TotalSpeed() int {
	spd := h.Speed
	for _, it := range []*EquipItem{h.Weapon, h.Head, h.Chest, h.Legs} {
		if it != nil {
			spd += it.SpeedBonus
		}
	}
	return spd
}

func (h *Hero) GetItemInSlot(slot EquipSlot) *EquipItem {
	switch slot {
	case SlotWeapon:
		return h.Weapon
	case SlotHead:
		return h.Head
	case SlotChest:
		return h.Chest
	case SlotLegs:
		return h.Legs
	}
	return nil
}

func (h *Hero) SetItemInSlot(slot EquipSlot, item *EquipItem) {
	switch slot {
	case SlotWeapon:
		h.Weapon = item
	case SlotHead:
		h.Head = item
	case SlotChest:
		h.Chest = item
	case SlotLegs:
		h.Legs = item
	}
}

func (h *Hero) HasFreePotionSlot(maxSlots int) bool {
	return len(h.Potions) < maxSlots
}

func (h *Hero) AddTreasure()       { h.Feats.TreasureFound++ }
func (h *Hero) RevealSecret()      { h.Feats.SecretsRevealed++ }
func (h *Hero) AddBlock()          { h.Feats.Blocks++ }
func (h *Hero) PullAggro()         { h.Feats.AggroPulled++ }
func (h *Hero) AddCriticalStrike() { h.Feats.CriticalStrikes++ }
func (h *Hero) AddBackstab()       { h.Feats.Backstabs++ }
func (h *Hero) StealLoot()         { h.Feats.LootStolen++ }
func (h *Hero) AddCCDuration()     { h.Feats.CCDuration++ }
func (h *Hero) AddManaBurst()      { h.Feats.ManaBursts++ }
func (h *Hero) AddRevive()         { h.Feats.Revives++ }
func (h *Hero) RemoveDot()         { h.Feats.DoTsRemoved++ }

// --- Монстры и Схватки ---
type MonsterAffix string

const (
	AffixNone     MonsterAffix = "Обычный"
	AffixFire     MonsterAffix = "Огненный"
	AffixPoison   MonsterAffix = "Ядовитый"
	AffixFrost    MonsterAffix = "Ледяной"
	AffixStone    MonsterAffix = "Каменный"
	AffixVampiric MonsterAffix = "Вампир"
)

type MonsterType string

const (
	MobRat         MonsterType = "Чумная крыса"
	MobGoblin      MonsterType = "Гоблин"
	MobSkeleton    MonsterType = "Скелет"
	MobSlime       MonsterType = "Болотный слизень"
	MobDrowned     MonsterType = "Утопленник"
	MobLizard      MonsterType = "Болотный ящер"
	MobImp         MonsterType = "Пепельный бес"
	MobOrc         MonsterType = "Орк-берсерк"
	MobSalamander  MonsterType = "Саламандра"
	MobGargoyle    MonsterType = "Гаргулья"
	MobGolem       MonsterType = "Кристальный голем"
	MobPhantom     MonsterType = "Фантом"
	MobVoidDemon   MonsterType = "Демон Бездны"
	MobDeathKnight MonsterType = "Рыцарь Смерти"
	MobDragon      MonsterType = "Пепельный Дракон"
)

type Monster struct {
	ID      int
	Type    MonsterType
	Name    string
	Level   int
	Affix   MonsterAffix
	Glyph   rune
	Color   string
	HP      int
	MaxHP   int
	Atk     int
	Defense int
	Speed   int
	Exp     int
	IsDead  bool
}

type MonsterPack struct {
	Members []*Monster
	IsBoss  bool
}

func (p *MonsterPack) LivingCount() int {
	cnt := 0
	for _, m := range p.Members {
		if !m.IsDead {
			cnt++
		}
	}
	return cnt
}

func (p *MonsterPack) GetFirstLiving() *Monster {
	for _, m := range p.Members {
		if !m.IsDead {
			return m
		}
	}
	return nil
}

func (p *MonsterPack) GetLowestHPFocus() *Monster {
	var target *Monster
	minHP := 99999
	for _, m := range p.Members {
		if !m.IsDead && m.HP < minHP {
			minHP = m.HP
			target = m
		}
	}
	return target
}

func (p *MonsterPack) GetHighestHPFocus() *Monster {
	var target *Monster
	maxHP := -1
	for _, m := range p.Members {
		if !m.IsDead && m.HP > maxHP {
			maxHP = m.HP
			target = m
		}
	}
	return target
}

type CombatantType int

const (
	CombatantHero CombatantType = iota
	CombatantMonster
)

type TurnOrderEntry struct {
	Type       CombatantType
	HeroRef    *Hero
	MonsterRef *Monster
	Initiative int
}

type ActiveCombat struct {
	Pos          Point
	Pack         *MonsterPack
	TurnQueue    []TurnOrderEntry
	TurnIdx      int
	Round        int
	HasBarrel    bool
	FleeCooldown int
}

// --- Карта, Квесты и Город ---
type Tile rune

const (
	TileWall         Tile = '#'
	TileFloor        Tile = '.'
	TileStairs       Tile = '>'
	TileChest        Tile = '$'
	TileExit         Tile = '<'
	TileAltar        Tile = '_'
	TileFountain     Tile = '~'
	TileTrappedChest Tile = 'T'
	TileBarrel       Tile = 'o'
	TileRelic        Tile = '*'
)

type Point struct {
	X, Y int
}

type BiomeType string

const (
	BiomeCatacombs BiomeType = "Гнилые Катакомбы"
	BiomeGrotto    BiomeType = "Затопленные Гроты"
	BiomeInferno   BiomeType = "Пепельные Недра"
	BiomeCrystal   BiomeType = "Кристальный Лабиринт"
	BiomeAbyss     BiomeType = "Трон Бездны"
)

type QuestType int

const (
	QuestHuntMonster QuestType = iota
	QuestOpenChests
	QuestReachFloor
	QuestUseAltar
	QuestFindRelic
	QuestEscapeTrap
)

type AutoQuest struct {
	Type        QuestType
	Title       string
	Description string
	TargetMob   MonsterType
	TargetCount int
	Current     int
	RewardGold  int
	Completed   bool
}

type BagUpgrade struct {
	Level    int
	Name     string
	Capacity int
	Cost     int
}

var bagUpgrades = []BagUpgrade{
	{Level: 1, Name: "Холщовый мешок", Capacity: 5, Cost: 0},
	{Level: 2, Name: "Кожаный ранец", Capacity: 8, Cost: 150},
	{Level: 3, Name: "Бездонная торба", Capacity: 12, Cost: 380},
	{Level: 4, Name: "Походный кофр", Capacity: 16, Cost: 750},
	{Level: 5, Name: "Обозный тюк", Capacity: 20, Cost: 1400},
	{Level: 6, Name: "Мешок иллюзий", Capacity: 25, Cost: 2600},
}

type FallenHeroRecord struct {
	FullName string
	Class    HeroClass
	Cause    string
	Floor    int
}

type RunStats struct {
	TotalGoldEarned int
	FloorsCleared   int
	ChestsOpened    int
	AltarsUsed      int
	FountainsUsed   int
	TrapsDisarmed   int
	Resurrections   int
	QuestsCompleted int
	UpgradesForged  int
	BarrelsBlown    int
	MonsterKills    map[MonsterType]int
	TotalSteps      int
	FallenHeroes    []FallenHeroRecord
}

func newStats() RunStats {
	return RunStats{
		MonsterKills: make(map[MonsterType]int),
		FallenHeroes: []FallenHeroRecord{},
	}
}

type TownLegacy struct {
	TreasuryGold int
	SmithyLevel  int
	TanneryLevel int
	ChurchLevel  int
	TavernLevel  int
}

type TownBudget struct {
	Treasury int
	Bags     int
	Recovery int
	Recruit  int
	Forge    int
	Tannery  int
	Alchemy  int
}

type TownEstablishments struct {
	SmithyName    string
	TanneryName   string
	TavernName    string
	GuildName     string
	AlchemistName string
	ChurchName    string
}

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateDefeat
	StateStatsManual
	StateArmory
	StateInfoBook
)

type TownPhase int

const (
	TownPhaseSellLoot TownPhase = iota
	TownPhaseMagistrate
	TownPhaseChurch
	TownPhaseTavern
	TownPhaseGuild
	TownPhaseSmithy
	TownPhaseTannery
	TownPhaseAlchemist
	TownPhaseDepart
)

type PartyRelic struct {
	Name        string
	Level       int
	Description string
	GoldMult    float64
	EnemyDmgMod float64
	MartyrFury  bool
	StressRes   int
}

type Model struct {
	State            GameState
	MenuCountdown    int
	TermWidth        int
	TermHeight       int
	MapWidth         int
	MapHeight        int
	Grid             [][]Tile
	Explored         [][]bool
	Packs            map[Point]*MonsterPack
	PartyPos         Point
	PathHistory      []Point
	LoopDetectCount  int
	Party            []*Hero
	Bag              []EquipItem
	BagLevel         int
	Gold             int
	Floor            int
	InTown           bool
	TownPhase        TownPhase
	TownDialog       string
	TownHistory      []string
	TownEst          TownEstablishments
	Combat           *ActiveCombat
	Logs             []string
	AutoMode         bool
	SpeedMs          int
	TownDelayMs      int
	StatsScroll      int
	LogScroll        int
	CodexTab         int
	RestartCountdown int
	CurrentQuest     AutoQuest
	Relic            *PartyRelic
	Legacy           TownLegacy
	Stats            RunStats
}

type (
	TickMsg        time.Time
	RestartTickMsg time.Time
	MenuTickMsg    time.Time
)

// --- Стили интерфейса ---
var (
	wallStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	floorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	partyStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	dangerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	healStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	accentStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	goldStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	questStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	subtleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	townArtStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Bold(true)
	altarStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("165")).Bold(true)
	fountStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	trappedChestStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	barrelStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("172")).Bold(true)
	potionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true)
	relicTileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)

	stressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("135")).Bold(true)
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	fireStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)

	heroCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	heroCardActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("226")).
			Padding(0, 1)

	heroCardDead = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)

	statsBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(1, 2).
			Align(lipgloss.Left)

	menuBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 3).
			Align(lipgloss.Center)
)
