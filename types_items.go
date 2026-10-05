package main

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
	{Level: 7, NameKey: "bag.tier_7", Capacity: 30, Cost: 4800},
	{Level: 8, NameKey: "bag.tier_8", Capacity: 40, Cost: 8500},
	{Level: 9, NameKey: "bag.tier_9", Capacity: 50, Cost: 15000},
}
