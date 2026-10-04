package main

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
	return HeroNames[rng.Intn(len(HeroNames))]
}

// --- Классы героев (10 шт.) ---
type HeroClass string

const (
	ClassTank    HeroClass = "tank"
	ClassWarrior HeroClass = "warrior"
	ClassRogue   HeroClass = "rogue"
	ClassMage    HeroClass = "mage"
	ClassCleric  HeroClass = "cleric"
	ClassPaladin HeroClass = "paladin"
	ClassRanger  HeroClass = "ranger"
	ClassMonk    HeroClass = "monk"
	ClassBard    HeroClass = "bard"
	ClassWarlock HeroClass = "warlock"
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
	NameKey        string
	CustomName     string // имя, введённое игроком (для героя-лидера); если задано — вместо NameKey
	IsLeader       bool
	Calling        LeaderCalling
	Race           RaceType
	Gender         Gender
	TitleKey       string
	Class          HeroClass
	Role           CombatRole
	Level          int
	Exp            int
	MaxHP          int
	HP             int
	MaxMP          int
	MP             int
	Stress         int
	Affliction     AfflictionType
	BaseAtk        int
	BaseDef        int
	Speed          int
	IsDead         bool
	IsDowned       bool // Без сознания / при смерти (ожидает выноса или помощи Клирика)
	LostInAbyss    bool
	IsGuarding     bool
	IsBerserk      bool
	IsStealthed    bool
	IsCharged      bool
	IsAura         bool
	ReviveCooldown int // Шагов/этажей до повторного боевого поднятия на ноги (для Клирика)
	CauseOfDeath   string
	SkillNameKey   string
	SkillCost      int
	Mutations      HeroMutations
	Feats          HeroHeroics

	Weapon *EquipItem
	Head   *EquipItem
	Chest  *EquipItem
	Legs   *EquipItem

	Potions []*Potion
}

func (h *Hero) DisplayName(lang Language) string {
	if h.CustomName != "" {
		return h.CustomName
	}
	return T(lang, h.NameKey)
}

func (h *Hero) RaceName(lang Language) string {
	return T(lang, "race."+string(h.Race)+".name")
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
