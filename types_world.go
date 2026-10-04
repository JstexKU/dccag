package main

const LegacyTaxRate = 0.15

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
	TileEvent        Tile = '?'
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
	BiomeDeadwood  BiomeType = "deadwood"
	BiomeFungal    BiomeType = "fungal"
	BiomeArchives  BiomeType = "archives"
	BiomeMines     BiomeType = "mines"
	BiomeSanctuary BiomeType = "sanctuary"
	BiomeAstral    BiomeType = "astral"
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
	QuestHuntMiniBoss
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

type FallenHeroRecord struct {
	FullName string
	Class    HeroClass
	Cause    string
	Floor    int
	Revived  bool
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
	EventsSeen      int
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
	TreasuryGold  int
	SmithyLevel   int
	TanneryLevel  int
	ChurchLevel   int
	TavernLevel   int
	TotalInvested int // Кумулятивные вложения в столицу за все прогоны — мета-прогрессия ополчения
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

type PartyRelic struct {
	NameKey     string
	Level       int
	DescKey     string
	GoldMult    float64
	EnemyDmgMod float64
	MartyrFury  bool
	StressRes   int
}

// --- Retreat: причины отступления в город ---
type RetreatReason int

const (
	RetreatNone        RetreatReason = iota // Отступать не нужно
	RetreatBagFull                          // Мешок полон
	RetreatQuestDone                        // Квест выполнен
	RetreatTooFewAlive                      // Живых бойцов осталось мало
	RetreatLowHP                            // Суммарное HP отряда низкое
	RetreatNoResources                      // Нет зелий и высокий стресс
)

// --- Healing Urgency: срочность поиска источника ---
type HealingUrgency int

const (
	HealingNone     HealingUrgency = iota // Исцеление не нужно
	HealingOptional                       // Можно посетить, если по пути
	HealingUrgent                         // Стоит свернуть с маршрута
	HealingCritical                       // Идти прямо сейчас, любой ценой
)
