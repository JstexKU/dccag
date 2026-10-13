package main

// --- Расы и их свойства ---
type RaceType string

const (
	RaceHuman     RaceType = "human"
	RaceElf       RaceType = "elf"
	RaceBeastman  RaceType = "beastman"
	RaceOlongr    RaceType = "olongr"
	RaceDwarf     RaceType = "dwarf"
	RaceDemonkin  RaceType = "demonkin"
	RaceDragonkin RaceType = "dragonkin"
)

var AllRaces = []RaceType{
	RaceHuman, RaceElf, RaceBeastman, RaceOlongr,
	RaceDwarf, RaceDemonkin, RaceDragonkin,
}

type RaceModifiers struct {
	ExpBonusPercent   float64
	SpeedFlat         int
	CritChanceBonus   int
	LifeStealPercent  float64
	DefBonusPercent   float64
	MaxHPBonusPercent float64
	ManaRegen         int
	StunImmune        bool
	StressResPercent  int  // Процент снижения получаемого стресса (Дварфы)
	LowHPAtkBonus     int  // Бонус к атаке при HP < 50% (Демониты)
	FireImmune        bool // Иммунитет к ожогам/стихии огня (Дракониты)
	RetaliateFire     int  // Чистый урон огнем атакующему врагу (Дракониты)
}

func GetRaceModifiers(r RaceType) RaceModifiers {
	switch r {
	case RaceElf:
		return RaceModifiers{SpeedFlat: 4, CritChanceBonus: 2, ManaRegen: 2}
	case RaceBeastman:
		return RaceModifiers{SpeedFlat: 2, LifeStealPercent: 0.10, DefBonusPercent: -0.10}
	case RaceOlongr:
		return RaceModifiers{SpeedFlat: -3, DefBonusPercent: 0.20, MaxHPBonusPercent: 0.15, StunImmune: true}
	case RaceDwarf:
		return RaceModifiers{SpeedFlat: -2, DefBonusPercent: 0.12, StressResPercent: 15}
	case RaceDemonkin:
		return RaceModifiers{SpeedFlat: 1, LowHPAtkBonus: 4, CritChanceBonus: 2}
	case RaceDragonkin:
		return RaceModifiers{SpeedFlat: -1, MaxHPBonusPercent: 0.12, DefBonusPercent: 0.08, FireImmune: true, RetaliateFire: 4}
	default: // Human
		return RaceModifiers{ExpBonusPercent: 0.15}
	}
}

// PreferredRacesForClass возвращает архетипичные расы для каждого класса
func PreferredRacesForClass(class HeroClass) []RaceType {
	switch class {
	case ClassTank:
		return []RaceType{RaceOlongr, RaceDwarf}
	case ClassWarrior:
		return []RaceType{RaceDwarf, RaceDemonkin}
	case ClassPaladin:
		return []RaceType{RaceHuman, RaceDragonkin}
	case ClassRogue:
		return []RaceType{RaceBeastman}
	case ClassRanger:
		return []RaceType{RaceBeastman, RaceElf}
	case ClassMage:
		return []RaceType{RaceElf}
	case ClassWarlock:
		return []RaceType{RaceDemonkin}
	case ClassCleric:
		return []RaceType{RaceHuman}
	case ClassBard:
		return []RaceType{RaceElf, RaceHuman}
	case ClassMonk:
		return []RaceType{RaceDragonkin, RaceOlongr}
	default:
		return []RaceType{RaceHuman}
	}
}

// PickRaceForClass выбирает расу: в 65% случаев профильную, в 35% — любую доступную
func PickRaceForClass(class HeroClass) RaceType {
	favored := PreferredRacesForClass(class)
	if rng.Intn(100) < 65 && len(favored) > 0 {
		return favored[rng.Intn(len(favored))]
	}
	return AllRaces[rng.Intn(len(AllRaces))]
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
	Race    RaceType
}

// 140 имён: по 10 мужских и 10 женских на каждую из 7 рас
var HeroNames = []HeroNameDef{
	// 1. Human (Люди)
	{NameKey: "hero.name.brand", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.aldos", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.walter", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.cassian", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.elias", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.duncan", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.gottfried", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.roland", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.godric", Gender: GenderMale, Race: RaceHuman},
	{NameKey: "hero.name.marcus", Gender: GenderMale, Race: RaceHuman},

	{NameKey: "hero.name.selina", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.brigitte", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.agnes", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.hilda", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.eleanor", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.beatrice", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.martha", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.constance", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.sybilla", Gender: GenderFemale, Race: RaceHuman},
	{NameKey: "hero.name.adele", Gender: GenderFemale, Race: RaceHuman},

	// 2. Elf (Эльфы)
	{NameKey: "hero.name.silran", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.veldor", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.aeron", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.falcon", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.zarvin", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.illirion", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.celeborn", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.faelin", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.talandril", Gender: GenderMale, Race: RaceElf},
	{NameKey: "hero.name.elorien", Gender: GenderMale, Race: RaceElf},

	{NameKey: "hero.name.lirianna", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.velara", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.mirael", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.celestina", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.kaelina", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.tirianna", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.nayra", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.elmira", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.silvana", Gender: GenderFemale, Race: RaceElf},
	{NameKey: "hero.name.elaria", Gender: GenderFemale, Race: RaceElf},

	// 3. Beastman (Зверолюды)
	{NameKey: "hero.name.wulf", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.ragnar", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.ingvar", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.sigurd", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.lyutobor", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.groh", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.fenris", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.grimm", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.torval", Gender: GenderMale, Race: RaceBeastman},
	{NameKey: "hero.name.brokk", Gender: GenderMale, Race: RaceBeastman},

	{NameKey: "hero.name.morrigan", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.zeyra", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.keldra", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.verissa", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.skadi", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.freya", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.runa", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.brigg", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.astrid", Gender: GenderFemale, Race: RaceBeastman},
	{NameKey: "hero.name.gerda", Gender: GenderFemale, Race: RaceBeastman},

	// 4. Olongr (Олонгры)
	{NameKey: "hero.name.yaropolk", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.radomir", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.dobrynya", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.bronislav", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.xandr", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.volimir", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.svyatoyar", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.gorynya", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.tverdislav", Gender: GenderMale, Race: RaceOlongr},
	{NameKey: "hero.name.stoyan", Gender: GenderMale, Race: RaceOlongr},

	{NameKey: "hero.name.lada", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.miroslava", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.radmila", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.bogdana", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.zvenislava", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.vsemila", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.gorislava", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.darina", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.dobronega", Gender: GenderFemale, Race: RaceOlongr},
	{NameKey: "hero.name.yaroslava", Gender: GenderFemale, Race: RaceOlongr},

	// 5. Dwarf (Дварфы)
	{NameKey: "hero.name.thorin", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.balin", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.dain", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.dvalin", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.baldur", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.grombor", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.thorgrim", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.khazad", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.kadrin", Gender: GenderMale, Race: RaceDwarf},
	{NameKey: "hero.name.rungar", Gender: GenderMale, Race: RaceDwarf},

	{NameKey: "hero.name.dis", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.helga", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.brinya", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.torra", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.gertrud", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.dagna", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.sigrun", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.hildur", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.verena", Gender: GenderFemale, Race: RaceDwarf},
	{NameKey: "hero.name.groa", Gender: GenderFemale, Race: RaceDwarf},

	// 6. Demonkin (Демониты)
	{NameKey: "hero.name.draven", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.kalar", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.zordan", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.azazel", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.malahor", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.valakar", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.balthazar", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.mordox", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.nergal", Gender: GenderMale, Race: RaceDemonkin},
	{NameKey: "hero.name.lucian", Gender: GenderMale, Race: RaceDemonkin},

	{NameKey: "hero.name.lilith", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.carmilla", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.bellatrix", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.morticia", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.zazira", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.hecate", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.sekhmet", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.velveta", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.astarte", Gender: GenderFemale, Race: RaceDemonkin},
	{NameKey: "hero.name.raziel", Gender: GenderFemale, Race: RaceDemonkin},

	// 7. Dragonkin (Дракониты)
	{NameKey: "hero.name.tarion", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.ignis", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.drakor", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.valirion", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.pyros", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.sargon", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.vulkanion", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.ragor", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.drakaris", Gender: GenderMale, Race: RaceDragonkin},
	{NameKey: "hero.name.smorg", Gender: GenderMale, Race: RaceDragonkin},

	{NameKey: "hero.name.veynara", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.sindra", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.ember", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.salamandra", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.pyrra", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.veigar", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.drakonis", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.tiamat", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.sovereign", Gender: GenderFemale, Race: RaceDragonkin},
	{NameKey: "hero.name.caldera", Gender: GenderFemale, Race: RaceDragonkin},
}

// getRandomHeroName выбирает имя под конкретную расу и пол
func getRandomHeroName(race RaceType, gender Gender) HeroNameDef {
	var matched []HeroNameDef
	for _, n := range HeroNames {
		if n.Race == race && n.Gender == gender {
			matched = append(matched, n)
		}
	}
	if len(matched) > 0 {
		return matched[rng.Intn(len(matched))]
	}
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
	CustomName     string
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
	IsDowned       bool
	LostInAbyss    bool
	IsGuarding     bool
	IsBerserk      bool
	IsStealthed    bool
	IsCharged      bool
	IsAura         bool
	ReviveCooldown int
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

	raceMod := GetRaceModifiers(h.Race)
	if raceMod.LowHPAtkBonus > 0 && float64(h.HP)/float64(h.MaxHP) <= 0.50 {
		b += raceMod.LowHPAtkBonus
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