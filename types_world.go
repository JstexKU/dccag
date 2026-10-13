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

// FloorState хранит сохранённое состояние конкретного этажа подземелья
type FloorState struct {
	Floor     int
	Width     int
	Height    int
	Grid      [][]Tile
	Explored  [][]bool
	Packs     map[Point]*MonsterPack
	ExitPos   Point
	StairsPos Point
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
	// --- Новые квесты v3.3.2 (всего 19 контрактов) ---
	QuestDisarmTraps     // Сапёрное дело (разминирование сундуков-ловушек)
	QuestUseFountain     // Благодать источника (испить святых вод)
	QuestFullMap         // Топография Бездны (разведка 90%+ этажа)
	QuestCollectGold     // Золотая лихорадка (сбор золота в подземелье)
	QuestSecrets         // Тайные тропы (поиск скрытых тайников и залов)
	QuestPurgeBiome      // Полная зачистка (уничтожение всех стай на этаже)
	QuestHuntBoss        // Истребление Владыки (победа над боссом/Драконом)
	QuestAffixBounty     // Охота на мутантов (враги с аффиксами)
	QuestPyromania       // Огненный шквал (подрыв бочек со смолой)
	QuestFlawlessFight   // Безупречный строй (бои без павших соратников)
	QuestCriticalStrikes // Точечные удары (критические попадания)
	QuestBreakArmor      // Несокрушимый оплот (блоки и парирования ударов)
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

// --- 10 уникальных типов реликвий ---
type RelicType string

const (
	RelicGreedCompass    RelicType = "greed_compass"    // Компас Алчности (+золото, урон мобов выше)
	RelicMartyrCrown     RelicType = "martyr_crown"     // Корона Мученика (+Atk при гибели союзника)
	RelicHolyGrail       RelicType = "holy_grail"       // Священный Грааль (сопротивление стрессу)
	RelicPioneerEye      RelicType = "pioneer_eye"      // Око Первопроходца (радиус обзора + автопоиск тайников)
	RelicQuicksilverVial RelicType = "quicksilver_vial" // Флакон Живой Ртути (+2 Скорость, право первого удара)
	RelicObsidianAnvil   RelicType = "obsidian_anvil"   // Обсидиановая Наковальня (закалка T2+ экипировки)
	RelicWhisperingSkull RelicType = "whispering_skull" // Шепчущий Череп (+крит отряда, входящие криты = стресс)
	RelicPhoenixCore     RelicType = "phoenix_core"     // Эссенция Феникса (спасение от гибели 1 раз за поход)
	RelicHornOfPlenty    RelicType = "horn_of_plenty"   // Рог Изобилия (+эффективность зелий и привала)
	RelicInquisitorSeal  RelicType = "inquisitor_seal"  // Печать Инквизитора (урон против нежити/демонов/элиты)
)

var AllRelics = []RelicType{
	RelicGreedCompass,
	RelicMartyrCrown,
	RelicHolyGrail,
	RelicPioneerEye,
	RelicQuicksilverVial,
	RelicObsidianAnvil,
	RelicWhisperingSkull,
	RelicPhoenixCore,
	RelicHornOfPlenty,
	RelicInquisitorSeal,
}

type PartyRelic struct {
	Type        RelicType
	NameKey     string
	Level       int
	DescKey     string
	GoldMult    float64
	EnemyDmgMod float64
	MartyrFury  bool
	StressRes   int

	// Активные свойства реликвий:
	VisionBonus    int     // Увеличение радиуса обзора тумана войны (Око Первопроходца)
	SpeedBonus     int     // Прибавка к скорости всего отряда (Флакон Живой Ртути)
	FirstStrike    bool    // Гарантированный первый ход в бою (Флакон Живой Ртути)
	GearTempered   bool    // Закалка: экипировка T2+ дает +1 Atk и +1 Def (Обсидиановая Наковальня)
	PartyCritBonus int     // Снижение порога крита для отряда (Шепчущий Череп)
	CritStressRisk int     // Доп. стресс при получении крита от моба (Шепчущий Череп)
	PhoenixCharged bool    // Заряд спасения от смертельного удара активен (Эссенция Феникса)
	PotionPotency  float64 // Множитель эффективности зелий и привала (Рог Изобилия)
	BaneDmgBonus   int     // Чистый доп. урон по демонам/нежити/боссам (Печать Инквизитора)
}

func (r PartyRelic) TotalGoldBonus() float64 {
	if r.GoldMult > 0 {
		return r.GoldMult
	}
	return 1.0
}

func (r PartyRelic) TotalEnemyMod() float64 {
	if r.EnemyDmgMod > 0 {
		return r.EnemyDmgMod
	}
	return 1.0
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
	RetreatTrainingDone                     // Тренировка закончена
)

// --- Healing Urgency: срочность поиска источника ---
type HealingUrgency int

const (
	HealingNone     HealingUrgency = iota // Исцеление не нужно
	HealingOptional                       // Можно посетить, если по пути
	HealingUrgent                         // Стоит свернуть с маршрута
	HealingCritical                       // Идти прямо сейчас, любой ценой
)