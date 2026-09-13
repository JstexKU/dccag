package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const LegacyTaxRate = 0.15

type Language string

const (
	LangRU Language = "ru"
	LangEN Language = "en"
)

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

// --- Расы и их свойства ---
type RaceType string

const (
	RaceHuman    RaceType = "human"
	RaceElf      RaceType = "elf"
	RaceBeastman RaceType = "beastman"
	RaceOlongr   RaceType = "olongr"
)

var AllRaces = []RaceType{RaceHuman, RaceElf, RaceBeastman, RaceOlongr}

type RaceModifiers struct {
	ExpBonusPercent   float64
	SpeedFlat         int
	CritChanceBonus   int
	LifeStealPercent  float64
	DefBonusPercent   float64
	MaxHPBonusPercent float64
	ManaRegen         int
	StunImmune        bool
}

func GetRaceModifiers(r RaceType) RaceModifiers {
	switch r {
	case RaceElf:
		return RaceModifiers{SpeedFlat: 4, CritChanceBonus: 2, ManaRegen: 2}
	case RaceBeastman:
		return RaceModifiers{SpeedFlat: 2, LifeStealPercent: 0.10, DefBonusPercent: -0.10}
	case RaceOlongr:
		return RaceModifiers{SpeedFlat: -3, DefBonusPercent: 0.20, MaxHPBonusPercent: 0.15, StunImmune: true}
	default: // Human
		return RaceModifiers{ExpBonusPercent: 0.15}
	}
}

// --- Предметы, Зелья, Мутации ---
type PotionType string

const (
	PotionHP     PotionType = "hp"
	PotionMP     PotionType = "mp"
	PotionStress PotionType = "stress"
)

type PotionSize string

const (
	SizeSmall  PotionSize = "small"
	SizeMedium PotionSize = "medium"
	SizeLarge  PotionSize = "large"
	SizeGrand  PotionSize = "grand"
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
	MutChimera MutationType = "chimera"
	MutFury    MutationType = "fury"
	MutTitan   MutationType = "titan"
	MutAether  MutationType = "aether"
	MutBastion MutationType = "bastion"
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
	SlotWeapon EquipSlot = "weapon"
	SlotHead   EquipSlot = "head"
	SlotChest  EquipSlot = "chest"
	SlotLegs   EquipSlot = "legs"
)

type ArmorCategory string

const (
	ArmorHeavy  ArmorCategory = "heavy"
	ArmorMedium ArmorCategory = "medium"
	ArmorLight  ArmorCategory = "light"
	ArmorNone   ArmorCategory = "none"
)

type MaterialTier struct {
	Key       string
	BonusMult int
	ValueMult int
}

var MetalMaterials = []MaterialTier{
	{Key: "mat.iron", BonusMult: 1, ValueMult: 1},
	{Key: "mat.steel", BonusMult: 2, ValueMult: 2},
	{Key: "mat.mithril", BonusMult: 3, ValueMult: 4},
	{Key: "mat.adamant", BonusMult: 4, ValueMult: 7},
}

var MageWeaponMaterials = []MaterialTier{
	{Key: "mat.yew", BonusMult: 1, ValueMult: 1},
	{Key: "mat.ash", BonusMult: 2, ValueMult: 2},
	{Key: "mat.crystal", BonusMult: 3, ValueMult: 4},
	{Key: "mat.aether", BonusMult: 4, ValueMult: 7},
}

var LeatherMaterials = []MaterialTier{
	{Key: "mat.raw_leather", BonusMult: 1, ValueMult: 1},
	{Key: "mat.boiled_leather", BonusMult: 2, ValueMult: 2},
	{Key: "mat.basilisk_skin", BonusMult: 3, ValueMult: 4},
	{Key: "mat.dragon_scale", BonusMult: 4, ValueMult: 7},
}

var ClothMaterials = []MaterialTier{
	{Key: "mat.linen", BonusMult: 1, ValueMult: 1},
	{Key: "mat.silk", BonusMult: 2, ValueMult: 2},
	{Key: "mat.brocade", BonusMult: 3, ValueMult: 4},
	{Key: "mat.void_cloth", BonusMult: 4, ValueMult: 7},
}

type ElementType string

const (
	ElemNone      ElementType = "none"
	ElemFire      ElementType = "fire"
	ElemPoison    ElementType = "poison"
	ElemFrost     ElementType = "frost"
	ElemLightning ElementType = "lightning"
)

type SuffixType string

const (
	SuffNone      SuffixType = "none"
	SuffVampirism SuffixType = "vampirism"
	SuffFury      SuffixType = "fury"
	SuffMana      SuffixType = "mana"
	SuffTitan     SuffixType = "titan"
)

type PrefixDef struct {
	Key     string
	Element ElementType
	Bonus   int
}

type SuffixDef struct {
	Key    string
	Effect SuffixType
	Bonus  int
}

var Prefixes = []PrefixDef{
	{Key: "prefix.fire", Element: ElemFire, Bonus: 3},
	{Key: "prefix.poison", Element: ElemPoison, Bonus: 2},
	{Key: "prefix.frost", Element: ElemFrost, Bonus: 2},
	{Key: "prefix.lightning", Element: ElemLightning, Bonus: 4},
	{Key: "prefix.none", Element: ElemNone, Bonus: 2},
}

var Suffixes = []SuffixDef{
	{Key: "suffix.vampirism", Effect: SuffVampirism, Bonus: 25},
	{Key: "suffix.fury", Effect: SuffFury, Bonus: 4},
	{Key: "suffix.mana", Effect: SuffMana, Bonus: 3},
	{Key: "suffix.titan", Effect: SuffTitan, Bonus: 4},
}

type EquipItem struct {
	BaseNameKey  string
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

func (e *EquipItem) DisplayName(lang Language) string {
	var parts []string
	if e.Prefix != nil {
		parts = append(parts, T(lang, e.Prefix.Key))
	}
	parts = append(parts, T(lang, e.Material.Key), T(lang, e.BaseNameKey))
	if e.UpgradeLevel > 0 {
		parts = append(parts, e.UpgradeLevelStr())
	}
	if e.Suffix != nil {
		parts = append(parts, T(lang, e.Suffix.Key))
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
	NameKey string
	Gender  Gender
}

// Имена героев локализуются через NameKey
var HeroNames = []HeroNameDef{
	{NameKey: "hero.name.brand", Gender: GenderMale},
	{NameKey: "hero.name.thorin", Gender: GenderMale},
	{NameKey: "hero.name.lyra", Gender: GenderFemale},
	{NameKey: "hero.name.aldos", Gender: GenderMale},
	{NameKey: "hero.name.selina", Gender: GenderFemale},
	{NameKey: "hero.name.ragnar", Gender: GenderMale},
	{NameKey: "hero.name.ingvar", Gender: GenderMale},
	{NameKey: "hero.name.wulf", Gender: GenderMale},
	{NameKey: "hero.name.sigurd", Gender: GenderMale},
	{NameKey: "hero.name.morgan", Gender: GenderMale},
	{NameKey: "hero.name.elias", Gender: GenderMale},
	{NameKey: "hero.name.duncan", Gender: GenderMale},
	{NameKey: "hero.name.gottfried", Gender: GenderMale},
	{NameKey: "hero.name.walter", Gender: GenderMale},
	{NameKey: "hero.name.cassian", Gender: GenderMale},
	{NameKey: "hero.name.iris", Gender: GenderFemale},
	{NameKey: "hero.name.morrigan", Gender: GenderFemale},
	{NameKey: "hero.name.brigitte", Gender: GenderFemale},
	{NameKey: "hero.name.agnes", Gender: GenderFemale},
	{NameKey: "hero.name.hilda", Gender: GenderFemale},
	{NameKey: "hero.name.yaropolk", Gender: GenderMale},
	{NameKey: "hero.name.radomir", Gender: GenderMale},
	{NameKey: "hero.name.dobrynya", Gender: GenderMale},
	{NameKey: "hero.name.lyutobor", Gender: GenderMale},
	{NameKey: "hero.name.bronislav", Gender: GenderMale},
	{NameKey: "hero.name.aeron", Gender: GenderMale},
	{NameKey: "hero.name.draven", Gender: GenderMale},
	{NameKey: "hero.name.kalar", Gender: GenderMale},
	{NameKey: "hero.name.zordan", Gender: GenderMale},
	{NameKey: "hero.name.tarion", Gender: GenderMale},
	{NameKey: "hero.name.veldor", Gender: GenderMale},
	{NameKey: "hero.name.falcon", Gender: GenderMale},
	{NameKey: "hero.name.yorick", Gender: GenderMale},
	{NameKey: "hero.name.edan", Gender: GenderMale},
	{NameKey: "hero.name.zarvin", Gender: GenderMale},
	{NameKey: "hero.name.lirianna", Gender: GenderFemale},
	{NameKey: "hero.name.velara", Gender: GenderFemale},
	{NameKey: "hero.name.mirael", Gender: GenderFemale},
	{NameKey: "hero.name.celestina", Gender: GenderFemale},
	{NameKey: "hero.name.kaelina", Gender: GenderFemale},
	{NameKey: "hero.name.tirianna", Gender: GenderFemale},
	{NameKey: "hero.name.nayra", Gender: GenderFemale},
	{NameKey: "hero.name.elmira", Gender: GenderFemale},
	{NameKey: "hero.name.zeyra", Gender: GenderFemale},
	{NameKey: "hero.name.xandr", Gender: GenderMale},
	{NameKey: "hero.name.verissa", Gender: GenderFemale},
	{NameKey: "hero.name.orwin", Gender: GenderMale},
	{NameKey: "hero.name.silran", Gender: GenderMale},
	{NameKey: "hero.name.keldra", Gender: GenderFemale},
	{NameKey: "hero.name.veynara", Gender: GenderFemale},
}

func getRandomHeroName() HeroNameDef {
	return HeroNames[rand.Intn(len(HeroNames))]
}

// --- Классы героев (10 шт.) ---
type HeroClass string

const (
	ClassTank     HeroClass = "tank"
	ClassWarrior  HeroClass = "warrior"
	ClassRogue    HeroClass = "rogue"
	ClassMage     HeroClass = "mage"
	ClassCleric   HeroClass = "cleric"
	ClassPaladin  HeroClass = "paladin"
	ClassRanger   HeroClass = "ranger"
	ClassMonk     HeroClass = "monk"
	ClassBard     HeroClass = "bard"
	ClassWarlock  HeroClass = "warlock"
)

var AllClasses = []HeroClass{
	ClassTank, ClassWarrior, ClassRogue, ClassMage, ClassCleric,
	ClassPaladin, ClassRanger, ClassMonk, ClassBard, ClassWarlock,
}

type CombatRole struct {
	AggroWeight int
	CanGuard    bool
	CanHeal     bool
}

func GetClassRole(class HeroClass) CombatRole {
	switch class {
	case ClassTank:
		return CombatRole{AggroWeight: 70, CanGuard: true, CanHeal: false}
	case ClassPaladin:
		return CombatRole{AggroWeight: 60, CanGuard: true, CanHeal: true}
	case ClassWarrior:
		return CombatRole{AggroWeight: 35, CanGuard: true, CanHeal: false}
	case ClassMonk:
		return CombatRole{AggroWeight: 30, CanGuard: false, CanHeal: false}
	case ClassRogue:
		return CombatRole{AggroWeight: 15, CanGuard: false, CanHeal: false}
	case ClassRanger:
		return CombatRole{AggroWeight: 18, CanGuard: false, CanHeal: false}
	case ClassMage:
		return CombatRole{AggroWeight: 20, CanGuard: false, CanHeal: false}
	case ClassWarlock:
		return CombatRole{AggroWeight: 22, CanGuard: false, CanHeal: false}
	case ClassCleric:
		return CombatRole{AggroWeight: 20, CanGuard: false, CanHeal: true}
	case ClassBard:
		return CombatRole{AggroWeight: 18, CanGuard: false, CanHeal: true}
	default:
		return CombatRole{AggroWeight: 25, CanGuard: false, CanHeal: false}
	}
}

type AfflictionType string

const (
	AfflictionNone     AfflictionType = "none"
	AfflictionParanoid AfflictionType = "paranoid"
	AfflictionSelfish  AfflictionType = "selfish"
	AfflictionManiac   AfflictionType = "maniac"
	AfflictionVirtuous AfflictionType = "virtuous"
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
	NameKey      string
	Race         RaceType
	Gender       Gender
	TitleKey     string
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
	SkillNameKey string
	SkillCost    int
	Mutations    HeroMutations
	Feats        HeroHeroics

	Weapon *EquipItem
	Head   *EquipItem
	Chest  *EquipItem
	Legs   *EquipItem

	Potions []*Potion
}

func (h *Hero) DisplayName(lang Language) string {
	return T(lang, h.NameKey)
}

func (h *Hero) RaceName(lang Language) string {
	return T(lang, "race."+string(h.Race)+".name")
}

func (h *Hero) Verb(male, female string) string {
	if h.Gender == GenderFemale {
		return female
	}
	return male
}

func (h *Hero) ShortClass(lang Language) string {
	return T(lang, "class."+string(h.Class)+".short")
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
		case ClassPaladin:
			h.MaxHP += 10
			h.MaxMP += 4
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
		case ClassMonk:
			h.MaxHP += 7
			h.MaxMP += 3
			h.BaseAtk += 2
			h.Speed += 1
		case ClassRogue:
			h.MaxHP += 5
			h.MaxMP += 4
			h.BaseAtk += 2
			h.Speed += 1
		case ClassRanger:
			h.MaxHP += 6
			h.MaxMP += 4
			h.BaseAtk += 2
			if h.Level%2 == 0 {
				h.Speed += 1
			}
		case ClassMage:
			h.MaxHP += 4
			h.MaxMP += 8
			h.BaseAtk += 3
		case ClassWarlock:
			h.MaxHP += 6
			h.MaxMP += 7
			h.BaseAtk += 2
		case ClassCleric:
			h.MaxHP += 6
			h.MaxMP += 6
			h.BaseAtk += 1
		case ClassBard:
			h.MaxHP += 5
			h.MaxMP += 6
			h.BaseAtk += 1
			if h.Level%2 == 0 {
				h.Speed += 1
			}
		}

		h.HP = h.MaxHP
		h.MP = h.MaxMP
	}
	return leveledUp
}

func (h *Hero) FullName(lang Language) string {
	name := h.DisplayName(lang)
	if h.TitleKey != "" {
		return name + " «" + T(lang, h.TitleKey) + "»"
	}
	return name
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
	total := h.BaseDef + b
	raceMod := GetRaceModifiers(h.Race)
	if raceMod.DefBonusPercent != 0 {
		total += int(float64(total) * raceMod.DefBonusPercent)
	}
	if total < 0 {
		total = 0
	}
	return total
}

func (h *Hero) TotalSpeed() int {
	spd := h.Speed + GetRaceModifiers(h.Race).SpeedFlat
	for _, it := range []*EquipItem{h.Weapon, h.Head, h.Chest, h.Legs} {
		if it != nil {
			spd += it.SpeedBonus
		}
	}
	if spd < 1 {
		spd = 1
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
	AffixNone     MonsterAffix = "none"
	AffixFire     MonsterAffix = "fire"
	AffixPoison   MonsterAffix = "poison"
	AffixFrost    MonsterAffix = "frost"
	AffixStone    MonsterAffix = "stone"
	AffixVampiric MonsterAffix = "vampiric"
)

type MonsterType string

const (
	MobRat         MonsterType = "rat"
	MobGoblin      MonsterType = "goblin"
	MobSkeleton    MonsterType = "skeleton"
	MobSlime       MonsterType = "slime"
	MobDrowned     MonsterType = "drowned"
	MobLizard      MonsterType = "lizard"
	MobImp         MonsterType = "imp"
	MobOrc         MonsterType = "orc"
	MobSalamander  MonsterType = "salamander"
	MobGargoyle    MonsterType = "gargoyle"
	MobGolem       MonsterType = "golem"
	MobPhantom     MonsterType = "phantom"
	MobVoidDemon   MonsterType = "void_demon"
	MobDeathKnight MonsterType = "death_knight"
	MobDragon      MonsterType = "dragon"
)

type Monster struct {
	ID      int
	Type    MonsterType
	NameKey string
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
	BiomeCatacombs BiomeType = "catacombs"
	BiomeGrotto    BiomeType = "grotto"
	BiomeInferno   BiomeType = "inferno"
	BiomeCrystal   BiomeType = "crystal"
	BiomeAbyss     BiomeType = "abyss"
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
	TitleKey    string
	DescKey     string
	TargetMob   MonsterType
	TargetCount int
	Current     int
	RewardGold  int
	Completed   bool
}

type BagUpgrade struct {
	Level    int
	NameKey  string
	Capacity int
	Cost     int
}

var bagUpgrades = []BagUpgrade{
	{Level: 1, NameKey: "bag.tier_1", Capacity: 5, Cost: 0},
	{Level: 2, NameKey: "bag.tier_2", Capacity: 8, Cost: 150},
	{Level: 3, NameKey: "bag.tier_3", Capacity: 12, Cost: 380},
	{Level: 4, NameKey: "bag.tier_4", Capacity: 16, Cost: 750},
	{Level: 5, NameKey: "bag.tier_5", Capacity: 20, Cost: 1400},
	{Level: 6, NameKey: "bag.tier_6", Capacity: 25, Cost: 2600},
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
	SmithyKey    string
	TanneryKey   string
	TavernKey    string
	GuildKey     string
	AlchemistKey string
	ChurchKey    string
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
	NameKey     string
	Level       int
	DescKey     string
	GoldMult    float64
	EnemyDmgMod float64
	MartyrFury  bool
	StressRes   int
}

type Model struct {
	Lang             Language
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