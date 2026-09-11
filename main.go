package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

func (m *Model) accumulateDebugReport() {
	globalDebugReport.TotalRuns++
	if m.Floor > globalDebugReport.MaxFloorReached {
		globalDebugReport.MaxFloorReached = m.Floor
	}

	globalDebugReport.TotalGoldSpent += (m.Stats.TotalGoldEarned - m.Gold)

	for _, fallen := range m.Stats.FallenHeroes {
		globalDebugReport.DeathCauses[fallen.Cause]++
		globalDebugReport.ClassDeaths[string(fallen.Class)]++
		globalDebugReport.DeathsByFloor[fallen.Floor]++
	}

	for _, h := range m.Party {
		if !h.IsDead {
			globalDebugReport.AverageAtk[string(h.Class)] = float64(h.TotalAtk())
		}
	}
}

func saveDebugReportToFile() {
	if !*debugReportFlag {
		return
	}
	fileData, err := json.MarshalIndent(globalDebugReport, "", "  ")
	if err == nil {
		_ = os.WriteFile("dccag_report.json", fileData, 0o644)
		fmt.Println("\n[DEBUG] Подробный отчет телеметрии сохранен в файл: dccag_report.json")
	}
}

// --- Стили ---
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

func renderBar(current, max int, totalBars int, filledColor, emptyColor lipgloss.Color) string {
	if totalBars < 2 {
		totalBars = 2
	}
	if max <= 0 {
		max = 1
	}
	ratio := float64(current) / float64(max)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(totalBars))
	empty := totalBars - filled
	fStr := lipgloss.NewStyle().Foreground(filledColor).Render(strings.Repeat("█", filled))
	eStr := lipgloss.NewStyle().Foreground(emptyColor).Render(strings.Repeat("░", empty))
	return "[" + fStr + eStr + "]"
}

func shortenItemName(name string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	replacer := strings.NewReplacer(
		"Пылающий", "Пыл.",
		"Ядовитый", "Ядов.",
		"Леденящий", "Лед.",
		"Громовой", "Гром.",
		"Закаленный", "Зак.",
		"Кровопийцы", "Кров.",
		"Медитации", "Мед.",
		"Ярости", "Яр.",
		"Титана", "Тит.",
		"Железн.", "Жел.",
		"Стальн.", "Стал.",
		"Мифрил.", "Мифр.",
		"Адамант.", "Адам.",
	)
	res := replacer.Replace(name)
	runes := []rune(res)
	if len(runes) > maxLen {
		if maxLen <= 1 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-1]) + "…"
	}
	return res
}

func padRight(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// --- Зелья и Экономика ---
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

func getPotionName(p Potion) string {
	return fmt.Sprintf("%s %s", p.Size, p.Type)
}

func CalculatePotionCost(basePrice, floor int) int {
	mult := 1.0 + (float64(floor) * 0.08)
	cost := int(float64(basePrice) * mult)
	maxCost := basePrice * 5
	if cost > maxCost {
		return maxCost
	}
	return cost
}

func createPotion(pType PotionType, size PotionSize, floor int) Potion {
	baseCost := 40
	power := 20
	switch size {
	case SizeSmall:
		power = 20
		baseCost = 35
	case SizeMedium:
		power = 45
		baseCost = 65
	case SizeLarge:
		power = 80
		baseCost = 110
	case SizeGrand:
		power = 140
		baseCost = 190
	}

	cost := CalculatePotionCost(baseCost, floor)
	symbol := "🟢"
	if pType == PotionMP {
		symbol = "🔵"
	} else if pType == PotionStress {
		symbol = "🟣"
	}

	return Potion{Type: pType, Size: size, Power: power, Cost: cost, Symbol: symbol}
}

// --- Мутации Алхимика ---
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

func GetClassMutationPreference(class HeroClass, m HeroMutations) MutationType {
	switch class {
	case ClassTank:
		if m.TitanCount <= m.BastionCount {
			return MutTitan
		}
		if m.BastionCount < 3 {
			return MutBastion
		}
		return MutChimera
	case ClassWarrior:
		if m.FuryCount <= m.TitanCount {
			return MutFury
		}
		return MutTitan
	case ClassRogue:
		if m.FuryCount <= m.ChimeraCount*2 {
			return MutFury
		}
		return MutChimera
	case ClassMage:
		if m.AetherCount <= m.FuryCount {
			return MutAether
		}
		return MutFury
	case ClassCleric:
		if m.AetherCount <= m.BastionCount {
			return MutAether
		}
		if m.BastionCount <= m.ChimeraCount {
			return MutBastion
		}
		return MutChimera
	default:
		return MutChimera
	}
}

func CalculateMutationCost(basePrice, floor, heroMutationsCount int) int {
	floorMult := 1.0 + (float64(floor) * 0.07)
	heroMult := 1.0 + (float64(heroMutationsCount) * 0.12)
	cost := int(float64(basePrice) * floorMult * heroMult)

	hardCap := basePrice * 6
	if cost > hardCap {
		cost = hardCap
	}
	return cost
}

// --- Биомы ---
type BiomeType string

const (
	BiomeCatacombs BiomeType = "Гнилые Катакомбы"
	BiomeGrotto    BiomeType = "Затопленные Гроты"
	BiomeInferno   BiomeType = "Пепельные Недра"
	BiomeCrystal   BiomeType = "Кристальный Лабиринт"
	BiomeAbyss     BiomeType = "Трон Бездны"
)

type BiomeConfig struct {
	Name       BiomeType
	WallColor  lipgloss.Color
	FloorColor lipgloss.Color
	FloorRune  rune
	EnvHazard  string
}

func getBiome(floor int) BiomeConfig {
	cycle := (floor - 1) % 5
	switch cycle {
	case 0:
		return BiomeConfig{Name: BiomeCatacombs, WallColor: lipgloss.Color("240"), FloorColor: lipgloss.Color("236"), FloorRune: '·', EnvHazard: "Сырость и гниль"}
	case 1:
		return BiomeConfig{Name: BiomeGrotto, WallColor: lipgloss.Color("31"), FloorColor: lipgloss.Color("24"), FloorRune: '~', EnvHazard: "Затопленные плиты (-2 Скор)"}
	case 2:
		return BiomeConfig{Name: BiomeInferno, WallColor: lipgloss.Color("124"), FloorColor: lipgloss.Color("52"), FloorRune: '≈', EnvHazard: "Палящий зной (+Огонь)"}
	case 3:
		return BiomeConfig{Name: BiomeCrystal, WallColor: lipgloss.Color("141"), FloorColor: lipgloss.Color("54"), FloorRune: '◊', EnvHazard: "Искажение эфира (+3 к MP)"}
	default:
		return BiomeConfig{Name: BiomeAbyss, WallColor: lipgloss.Color("89"), FloorColor: lipgloss.Color("233"), FloorRune: '×', EnvHazard: "Дыхание Бездны (+10% стресса)"}
	}
}

type TownLegacy struct {
	TreasuryGold int
	SmithyLevel  int
	ChurchLevel  int
	TavernLevel  int
}

type TownBudget struct {
	Treasury int
	Bags     int
	Recovery int
	Recruit  int
	Forge    int
	Alchemy  int
}

func AllocateBudget(gold int) TownBudget {
	treasury := int(float64(gold) * 0.10)
	rem := gold - treasury
	return TownBudget{
		Treasury: treasury,
		Bags:     int(float64(rem) * 0.05),
		Recovery: int(float64(rem) * 0.25),
		Recruit:  int(float64(rem) * 0.15),
		Forge:    int(float64(rem) * 0.25),
		Alchemy:  int(float64(rem) * 0.20),
	}
}

// --- Реликвии ---
type PartyRelic struct {
	Name        string
	Level       int
	Description string
	GoldMult    float64
	EnemyDmgMod float64
	MartyrFury  bool
	StressRes   int
}

func generateRelic(level int) PartyRelic {
	relicTypes := []string{"Компас Алчности", "Корона Мученика", "Священный Грааль"}
	chosen := relicTypes[rand.Intn(len(relicTypes))]

	prefix := "Потускневший"
	if level == 2 {
		prefix = "Освященный"
	} else if level == 3 {
		prefix = "Древний"
	}

	fullName := fmt.Sprintf("%s %s (Ур.%d)", prefix, chosen, level)

	switch chosen {
	case "Компас Алчности":
		gMult := 1.15 + (float64(level) * 0.10)
		dmgMod := 1.15 - (float64(level) * 0.03)
		return PartyRelic{
			Name: fullName, Level: level, GoldMult: gMult, EnemyDmgMod: dmgMod, MartyrFury: false,
			Description: fmt.Sprintf("+%.0f%% золота, урон врагов %.2fx", (gMult-1)*100, dmgMod),
		}
	case "Корона Мученика":
		return PartyRelic{
			Name: fullName, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: true,
			Description: fmt.Sprintf("При гибели живые получают +%d Atk", level*3),
		}
	default:
		res := 15 + (level * 15)
		return PartyRelic{
			Name: fullName, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: false, StressRes: res,
			Description: fmt.Sprintf("Снижает стресс отряда на %d%%", res),
		}
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

var FirstNames = []string{
	"Бранд", "Торин", "Лира", "Алдос", "Селина",
	"Рагнар", "Ингвар", "Вульф", "Сигурд", "Морган",
	"Элиас", "Дункан", "Готфрид", "Вальтер", "Кассиан",
	"Айрис", "Морриган", "Бригитта", "Агнес", "Хильда",
	"Ярополк", "Радомир", "Добрыня", "Лютобор", "Бронислав",
	"Аэрон", "Дрейвен", "Кэлар", "Зордан", "Тарион",
	"Велдор", "Фалькон", "Йоррик", "Эдан", "Зарвин",
	"Лирианна", "Велара", "Мираэль", "Селестина", "Каэлина",
	"Тирианна", "Найра", "Эльмира", "Зейра", "Ксандр",
	"Верисса", "Орвин", "Сильран", "Келдра", "Вейнара",
}

func getRandomName() string {
	return FirstNames[rand.Intn(len(FirstNames))]
}

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

var Materials = []MaterialTier{
	{Name: "Железо", Adj: "Железн.", BonusMult: 1, ValueMult: 1},
	{Name: "Сталь", Adj: "Стальн.", BonusMult: 2, ValueMult: 2},
	{Name: "Мифрил", Adj: "Мифрил.", BonusMult: 3, ValueMult: 4},
	{Name: "Адамант", Adj: "Адамант.", BonusMult: 4, ValueMult: 7},
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
		parts = append(parts, fmt.Sprintf("+%d", e.UpgradeLevel))
	}
	if e.Suffix != nil {
		parts = append(parts, e.Suffix.Name)
	}
	return strings.Join(parts, " ")
}

// --- Роли и Герои ---
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

type Hero struct {
	Name         string
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

	Potion *Potion
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

func (h *Hero) FullName() string {
	name := h.Name
	if h.Title != "" {
		name = fmt.Sprintf("%s «%s»", h.Name, h.Title)
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
	if b < 0 {
		b = 0
	}
	return h.BaseDef + b
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

func generateItemForClassSlot(class HeroClass, slot EquipSlot, floor int) EquipItem {
	matIdx := floor / 3
	if matIdx >= len(Materials) {
		matIdx = len(Materials) - 1
	}
	mat := Materials[rand.Intn(matIdx+1)]

	upg := 0
	if rand.Intn(100) > 75 {
		upg = rand.Intn(floor/3 + 1)
		if upg > 3 {
			upg = 3
		}
	}

	var pfx *PrefixDef
	if rand.Intn(100) > 60 {
		p := Prefixes[rand.Intn(len(Prefixes))]
		pfx = &p
	}

	var sfx *SuffixDef
	if rand.Intn(100) > 70 {
		s := Suffixes[rand.Intn(len(Suffixes))]
		sfx = &s
	}

	tier := floor / 4
	if tier > 3 {
		tier = 3
	}
	if rand.Intn(100) < 25 && tier < 3 {
		tier++
	}

	name := "Снаряжение"
	baseStat := 2
	bonusMP := 0
	bonusHP := 0
	critBonus := 0
	blockBonus := 0
	stressRes := 0
	cat := ArmorMedium

	switch class {
	case ClassTank:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			names := []string{"Гладиус с баклером", "Палаш с щитом", "Моргенштерн с павезой", "Бастионный меч"}
			name = names[tier]
			baseStat = 3 + tier*2
			blockBonus = 2 + tier*2
		case SlotChest:
			names := []string{"Бригантина", "Полудоспех", "Кираса бастиона", "Панцирь цитадели"}
			name = names[tier]
			baseStat = 4 + tier*3
			bonusHP = 10 + tier*10
		case SlotHead:
			names := []string{"Топфхельм", "Салад", "Армет", "Бацинет бастиона"}
			name = names[tier]
			baseStat = 2 + tier*2
			blockBonus = 1 + tier
		case SlotLegs:
			names := []string{"Наголенники", "Шарнирные поножи", "Латные поножи", "Протекторы цитадели"}
			name = names[tier]
			baseStat = 2 + tier*2
			bonusHP = 5 + tier*5
		}

	case ClassWarrior:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			names := []string{"Эспадон", "Клеймор", "Боевой топор", "Фальшион"}
			name = names[tier]
			baseStat = 5 + tier*3
			critBonus = 1 + tier
		case SlotChest:
			names := []string{"Хауберк", "Кираса ярости", "Чешуйчатый доспех", "Нагрудник витязя"}
			name = names[tier]
			baseStat = 3 + tier*2
			bonusHP = 8 + tier*8
		case SlotHead:
			names := []string{"Норманнский шлем", "Бацинет", "Барбют", "Шишак"}
			name = names[tier]
			baseStat = 2 + tier*2
			critBonus = 1
		case SlotLegs:
			names := []string{"Чешуйчатые гетры", "Чулки", "Пластины", "Поножи витязя"}
			name = names[tier]
			baseStat = 2 + tier*2
		}

	case ClassRogue:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			names := []string{"Охотничьи ножи", "Парные стилеты", "Зазубренные кинжалы", "Воровские кортики"}
			name = names[tier]
			baseStat = 4 + tier*2
			critBonus = 2 + tier*2
		case SlotChest:
			names := []string{"Колет", "Гамбезон", "Куртка теневика", "Плащ ассасина"}
			name = names[tier]
			baseStat = 2 + tier*2
			critBonus = 1 + tier
		case SlotHead:
			names := []string{"Тканевая маска", "Капюшон", "Бандана", "Маска теней"}
			name = names[tier]
			baseStat = 1 + tier*2
			critBonus = 1
		case SlotLegs:
			names := []string{"Краги", "Плотные гетры", "Мягкие сапоги", "Поножи бесшумности"}
			name = names[tier]
			baseStat = 1 + tier*2
		}

	case ClassMage:
		cat = ArmorLight
		switch slot {
		case SlotWeapon:
			names := []string{"Рунная трость", "Посох искр", "Кристаллический жезл", "Архимагический скипетр"}
			name = names[tier]
			baseStat = 6 + tier*3
			bonusMP = 10 + tier*10
		case SlotChest:
			names := []string{"Роба ученика", "Мантия чародея", "Одеяние эфира", "Астральная мантия"}
			name = names[tier]
			baseStat = 2 + tier*2
			bonusMP = 15 + tier*10
		case SlotHead:
			names := []string{"Остроконечная шляпа", "Обруч магии", "Диадема фокуса", "Капюшон магистра"}
			name = names[tier]
			baseStat = 1 + tier*2
			bonusMP = 8 + tier*6
		case SlotLegs:
			names := []string{"Обмотки", "Шёлковые поножи", "Ленты левитации", "Штаны чародея"}
			name = names[tier]
			baseStat = 1 + tier*2
			bonusMP = 6 + tier*4
		}

	case ClassCleric:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			names := []string{"Окованная дубина", "Боевой молот", "Шестопёр", "Булава света"}
			name = names[tier]
			baseStat = 4 + tier*2
			bonusMP = 8 + tier*6
		case SlotChest:
			names := []string{"Сутана", "Кираса инквизитора", "Пресвитерский панцирь", "Священный доспех"}
			name = names[tier]
			baseStat = 3 + tier*2
			stressRes = 10 + tier*5
		case SlotHead:
			names := []string{"Митра", "Койф", "Капеллина", "Венец правосудия"}
			name = names[tier]
			baseStat = 2 + tier*2
			stressRes = 5 + tier*5
		case SlotLegs:
			names := []string{"Наголенники веры", "Сапоги паломника", "Инквизиторские сапоги", "Наколенники света"}
			name = names[tier]
			baseStat = 2 + tier*2
			bonusHP = 6 + tier*6
		}
	}

	val := (baseStat + tier*3) * mat.ValueMult * 12

	return EquipItem{
		BaseName: name, Slot: slot, Category: cat, AllowedClass: class,
		Material: mat, UpgradeLevel: upg, BaseStat: baseStat,
		BonusMP: bonusMP, BonusHP: bonusHP, CritBonus: critBonus,
		BlockBonus: blockBonus, StressRes: stressRes, Value: val,
		Prefix: pfx, Suffix: sfx,
	}
}

func generateItemForClass(class HeroClass, floor int) EquipItem {
	slots := []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs}
	chosenSlot := slots[rand.Intn(len(slots))]
	return generateItemForClassSlot(class, chosenSlot, floor)
}

func (m *Model) checkAndAwardTitle(h *Hero) {
	if h.Title != "" && h.Title != "Ополченец" {
		return
	}

	newTitle := ""
	switch {
	case h.Feats.BossKills >= 3:
		newTitle = "Истребитель чудовищ"
	case h.Feats.BossKills >= 1:
		newTitle = "Драконоборец"
	case h.Feats.NearDeathEscapes >= 6:
		newTitle = "Проклятый удачей"
	case h.Feats.NearDeathEscapes >= 3:
		newTitle = "Бессмертный"
	case h.Feats.TreasureFound >= 5:
		newTitle = "Искатель кладов"
	case h.Feats.SecretsRevealed >= 4:
		newTitle = "Хранитель тайн"
	case h.Class == ClassTank && h.Feats.Blocks >= 10:
		newTitle = "Непробиваемый"
	case h.Class == ClassTank && h.Feats.AggroPulled >= 8:
		newTitle = "Грозовой щит"
	case h.Class == ClassTank && h.Feats.DamageTaken >= 75:
		newTitle = "Стена"
	case h.Class == ClassWarrior && h.Feats.Kills >= 12:
		newTitle = "Кровавый клинок"
	case h.Class == ClassWarrior && h.Feats.Kills >= 5:
		newTitle = "Палач"
	case h.Class == ClassWarrior && h.Feats.DamageDealt >= 90:
		newTitle = "Топор"
	case h.Class == ClassWarrior && h.Feats.CriticalStrikes >= 4:
		newTitle = "Разрыватель рядов"
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 6:
		newTitle = "Призрачный удар"
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 3:
		newTitle = "Тень"
	case h.Class == ClassRogue && h.Feats.Kills >= 4:
		newTitle = "Клинок"
	case h.Class == ClassRogue && h.Feats.Backstabs >= 5:
		newTitle = "Нож за спиной"
	case h.Class == ClassRogue && h.Feats.LootStolen >= 3:
		newTitle = "Ловкая рука"
	case h.Class == ClassMage && h.Feats.DamageDealt >= 150:
		newTitle = "Буревестник"
	case h.Class == ClassMage && h.Feats.DamageDealt >= 90:
		newTitle = "Пепел"
	case h.Class == ClassMage && h.Feats.CCDuration >= 40:
		newTitle = "Оковы пустоты"
	case h.Class == ClassMage && h.Feats.ManaBursts >= 3:
		newTitle = "Вспышка"
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 120:
		newTitle = "Благодать"
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 70:
		newTitle = "Святой"
	case h.Class == ClassCleric && h.Feats.Revives >= 3:
		newTitle = "Воскреситель"
	case h.Class == ClassCleric && h.Feats.DoTsRemoved >= 4:
		newTitle = "Очиститель"
	}

	if newTitle != "" {
		h.Title = newTitle
		m.addLog(titleStyle.Render(fmt.Sprintf("👑 [СЛАВА] %s заслужил титул «%s»!", h.Name, newTitle)))
	}
}

func (m *Model) addStress(h *Hero, amt int) {
	if h.IsDead {
		return
	}
	if m.Relic != nil && m.Relic.StressRes > 0 {
		amt = amt * (100 - m.Relic.StressRes) / 100
	}

	itemRes := 0
	for _, it := range []*EquipItem{h.Head, h.Chest, h.Legs} {
		if it != nil {
			itemRes += it.StressRes
		}
	}
	if itemRes > 60 {
		itemRes = 60
	}
	amt = amt * (100 - itemRes) / 100

	h.Stress += amt

	if amt >= 15 {
		splash := amt / 3
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += splash
			}
		}
	}

	if h.Stress >= 100 && h.Affliction == AfflictionNone {
		if rand.Intn(100) < 30 {
			h.Affliction = AfflictionVirtuous
			h.Stress = 0
			h.HP = h.MaxHP
			m.addLog(healStyle.Render(fmt.Sprintf("🌟 [ВООДУШЕВЛЕНИЕ] %s превозмог страх и обрел второе дыхание!", h.FullName())))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress = max(0, ally.Stress-30)
				}
			}
		} else {
			affs := []AfflictionType{AfflictionParanoid, AfflictionSelfish, AfflictionManiac}
			h.Affliction = affs[rand.Intn(len(affs))]
			m.addLog(stressStyle.Render(fmt.Sprintf("👁️ [ПСИХОЗ] %s сломлен: %s!", h.FullName(), h.Affliction)))
		}
	}

	if h.Stress >= 200 && h.Affliction != AfflictionVirtuous {
		h.Stress = 200
		if h.HP <= 1 {
			h.HP = 0
			h.CauseOfDeath = fmt.Sprintf("Сердечный приступ (%s)", h.Affliction)
			m.recordFallenHero(h)
			m.addLog(dangerStyle.Render(fmt.Sprintf("💔 [ИНФАРКТ] Сердце %s разорвалось от безумия! Смерть!", h.Name)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress += 25
				}
			}
			return
		}

		h.HP = 1
		h.Stress = 160
		m.addLog(dangerStyle.Render(fmt.Sprintf("💔 [СЕРДЕЧНЫЙ ПРИСТУП] %s схватился за сердце! HP упало до 1!", h.FullName())))
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += 15
			}
		}
	}
}

// --- Монстры и Аффиксы ---
type MonsterAffix string

const (
	AffixNone     MonsterAffix = "Обычный"
	AffixFire     MonsterAffix = "Огненный"
	AffixPoison   MonsterAffix = "Ядовитый"
	AffixFrost    MonsterAffix = "Ледяной"
	AffixStone    MonsterAffix = "Каменный"
	AffixVampiric MonsterAffix = "Вампир"
)

func getAffixIcon(a MonsterAffix) string {
	switch a {
	case AffixFire:
		return "🔥"
	case AffixPoison:
		return "☣️"
	case AffixFrost:
		return "❄️"
	case AffixStone:
		return "🪨"
	case AffixVampiric:
		return "🩸"
	default:
		return ""
	}
}

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

func spawnMonsterPack(isBoss bool, floor int) *MonsterPack {
	pack := &MonsterPack{IsBoss: isBoss}
	scaleMult := 1 + (floor / 10)

	if isBoss {
		if floor%10 == 0 {
			dragonLvl := floor
			dragonHP := (260 + (floor * 20)) * scaleMult
			dragon := &Monster{
				ID: 1, Type: MobDragon, Name: fmt.Sprintf("Пепельный Дракон (БОСС Этажа %d)", floor), Level: dragonLvl, Affix: AffixFire,
				Glyph: 'D', Color: "196", HP: dragonHP, MaxHP: dragonHP,
				Atk: (25 + floor) * scaleMult, Defense: (9 + floor/2) * scaleMult, Speed: 12, Exp: 350 * scaleMult,
			}
			pack.Members = append(pack.Members, dragon)

			for i := 1; i <= 2; i++ {
				guard := &Monster{
					ID: i + 1, Type: MobDeathKnight, Name: fmt.Sprintf("Рыцарь Смерти #%d", i), Level: dragonLvl - 1, Affix: AffixVampiric,
					Glyph: 'K', Color: "89", HP: (65 + floor*5) * scaleMult, MaxHP: (65 + floor*5) * scaleMult,
					Atk: (18 + floor) * scaleMult, Defense: (6 + floor/3) * scaleMult, Speed: 10, Exp: 55 * scaleMult,
				}
				pack.Members = append(pack.Members, guard)
			}
			return pack
		}

		bossLvl := floor
		hp := (70 + (bossLvl * 15)) * scaleMult
		bType := MobOrc
		bGlyph := 'B'
		bColor := "202"
		bName := "Вождь Орков"
		bAtk := (14 + (bossLvl * 2)) * scaleMult
		bDef := (4 + bossLvl/3) * scaleMult

		if floor >= 7 {
			bType = MobGolem
			bName = "Алмазный Колосс"
			bGlyph = 'G'
			bColor = "141"
			bDef = (8 + floor/4) * scaleMult
		} else if floor >= 4 {
			bType = MobDrowned
			bName = "Глубинный Левиафан"
			bGlyph = 'L'
			bColor = "31"
		}

		pack.Members = append(pack.Members, &Monster{
			ID: 1, Type: bType, Name: bName, Level: bossLvl, Affix: AffixStone,
			Glyph: bGlyph, Color: bColor, HP: hp, MaxHP: hp,
			Atk: bAtk, Defense: bDef, Speed: 10, Exp: (70 + bossLvl*8) * scaleMult,
		})
		for i := 1; i <= 2; i++ {
			pack.Members = append(pack.Members, &Monster{
				ID: i + 1, Type: MobSkeleton, Name: fmt.Sprintf("Прислужник #%d", i), Level: floor, Affix: AffixNone,
				Glyph: 's', Color: "245", HP: (24 + floor*5) * scaleMult, MaxHP: (24 + floor*5) * scaleMult,
				Atk: (9 + floor*2) * scaleMult, Defense: 3 * scaleMult, Speed: 9, Exp: (18 + floor*3) * scaleMult,
			})
		}
		return pack
	}

	packSize := rand.Intn(3) + 3
	affixes := []MonsterAffix{AffixNone, AffixFire, AffixPoison, AffixFrost, AffixStone, AffixVampiric}

	for i := 1; i <= packSize; i++ {
		mobLvl := floor
		if rand.Intn(100) < 30 {
			mobLvl++
		}

		aff := AffixNone
		if floor >= 3 && rand.Intn(100) < (20+(floor*5)) {
			aff = affixes[rand.Intn(len(affixes))]
		}

		var mType MonsterType
		var glyph rune
		var color string
		var baseAtk, baseDef, baseHP int

		biomeCycle := (floor - 1) % 5

		if biomeCycle == 0 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobRat, 'r', "137", 5, 0, 14
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGoblin, 'g', "118", 7, 1, 17
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSkeleton, 's', "252", 8, 3, 22
			}
		} else if biomeCycle == 1 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSlime, 'c', "43", 9, 1, 26
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDrowned, 'u', "31", 11, 2, 34
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobLizard, 'l', "29", 12, 3, 30
			}
		} else if biomeCycle == 2 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobImp, 'i', "208", 13, 2, 36
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobOrc, 'o', "130", 15, 4, 46
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSalamander, 'm', "196", 16, 3, 40
			}
		} else if biomeCycle == 3 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGargoyle, 'G', "102", 17, 6, 52
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGolem, 'C', "141", 18, 7, 60
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobPhantom, 'p', "159", 19, 2, 44
			}
		} else {
			roll := rand.Intn(2)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobVoidDemon, 'V', "161", 21, 5, 66
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDeathKnight, 'K', "89", 22, 7, 76
			}
		}

		hp := (baseHP + (mobLvl * 5)) * scaleMult
		atk := (baseAtk + (mobLvl * 2)) * scaleMult
		def := (baseDef + (mobLvl / 3)) * scaleMult

		if aff == AffixFire {
			atk += 3 * scaleMult
		} else if aff == AffixStone {
			def += 3 * scaleMult
			hp += 12 * scaleMult
		}

		pack.Members = append(pack.Members, &Monster{
			ID: i, Type: mType, Name: fmt.Sprintf("%s #%d", mType, i), Level: mobLvl, Affix: aff,
			Glyph: glyph, Color: color, HP: hp, MaxHP: hp, Atk: atk, Defense: def, Speed: 8 + mobLvl/2, Exp: (10 + mobLvl*4) * scaleMult,
		})
	}
	return pack
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

// --- Квесты ---
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

func generateAutoQuest(curFloor int) AutoQuest {
	qType := QuestType(rand.Intn(6))
	switch qType {
	case QuestHuntMonster:
		var pool []MonsterType
		biomeCycle := (curFloor - 1) % 5
		switch biomeCycle {
		case 0:
			pool = []MonsterType{MobRat, MobGoblin, MobSkeleton}
		case 1:
			pool = []MonsterType{MobSlime, MobDrowned, MobLizard}
		case 2:
			pool = []MonsterType{MobImp, MobOrc, MobSalamander}
		case 3:
			pool = []MonsterType{MobGargoyle, MobGolem, MobPhantom}
		default:
			pool = []MonsterType{MobVoidDemon, MobDeathKnight}
		}
		target := pool[rand.Intn(len(pool))]
		count := rand.Intn(3) + 3
		return AutoQuest{
			Type: QuestHuntMonster, TargetMob: target, TargetCount: count,
			Title: fmt.Sprintf("Охота: %s", target), Description: fmt.Sprintf("Истребить %d [%s]", count, target),
			RewardGold: count * (20 + curFloor*5),
		}
	case QuestOpenChests:
		count := rand.Intn(2) + 2
		return AutoQuest{
			Type: QuestOpenChests, TargetCount: count,
			Title: "Сбор сокровищ", Description: fmt.Sprintf("Вскрыть %d сундуков", count),
			RewardGold: count * (30 + curFloor*3),
		}
	case QuestReachFloor:
		tf := curFloor + 1
		return AutoQuest{
			Type: QuestReachFloor, TargetCount: tf,
			Title:       fmt.Sprintf("Освоение глубин: Этаж %d", tf),
			Description: fmt.Sprintf("Подготовиться и спуститься на этаж %d", tf),
			RewardGold:  tf * 50,
		}
	case QuestFindRelic:
		return AutoQuest{
			Type: QuestFindRelic, TargetCount: 1,
			Title: "Поиск Реликвии", Description: "Найти древний реликварий в глубинах",
			RewardGold: 100 + curFloor*15,
		}
	case QuestEscapeTrap:
		return AutoQuest{
			Type: QuestEscapeTrap, TargetCount: 1,
			Title: "Побег из засады", Description: "Исследовать запечатанный зал и найти выход",
			RewardGold: 110 + curFloor*20,
		}
	default:
		return AutoQuest{
			Type: QuestUseAltar, TargetCount: 1,
			Title: "Кровавый пакт", Description: "Принести жертву у Алтаря",
			RewardGold: 80 + curFloor*15,
		}
	}
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
	TownPhaseUpgradeBag
	TownPhaseChurch
	TownPhaseTavern
	TownPhaseGuild
	TownPhaseSmithy
	TownPhaseAlchemist
	TownPhaseDepart
)

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
	Combat           *ActiveCombat
	Logs             []string
	AutoMode         bool
	SpeedMs          int
	TownDelayMs      int
	StatsScroll      int
	LogScroll        int
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

func tickCmd(speedMs int) tea.Cmd {
	return tea.Tick(time.Duration(speedMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func restartTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return RestartTickMsg(t)
	})
}

func menuTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return MenuTickMsg(t)
	})
}

func createHero(class HeroClass, floor int, smithyLvl int) *Hero {
	targetLevel := max(1, floor/2)

	maxHP := 45
	baseDef := 2
	speed := 10
	skillName := "Удар"
	skillCost := 10
	maxMP := 35

	switch class {
	case ClassTank:
		maxHP = 65
		baseDef = 4
		speed = 8
		skillName = "Стойка"
		skillCost = 8
		maxMP = 30
	case ClassWarrior:
		maxHP = 50
		baseDef = 2
		speed = 10
		skillName = "Ярость"
		skillCost = 10
		maxMP = 25
	case ClassRogue:
		maxHP = 38
		baseDef = 1
		speed = 15
		skillName = "Тень"
		skillCost = 12
		maxMP = 35
	case ClassMage:
		maxHP = 30
		baseDef = 0
		speed = 11
		skillName = "Заряд"
		skillCost = 15
		maxMP = 45
	case ClassCleric:
		maxHP = 36
		baseDef = 2
		speed = 9
		skillName = "Аура"
		skillCost = 12
		maxMP = 40
	}

	h := &Hero{
		Name:      getRandomName(),
		Class:     class,
		Role:      GetClassRole(class),
		Level:     1,
		Exp:       0,
		MaxHP:     maxHP,
		HP:        maxHP,
		MaxMP:     maxMP,
		MP:        maxMP,
		BaseAtk:   6 + smithyLvl,
		BaseDef:   baseDef,
		Speed:     speed,
		SkillName: skillName,
		SkillCost: skillCost,
		IsDead:    false,
	}

	for h.Level < targetLevel {
		h.GainExp(h.NextLevelExp())
	}

	itemFloor := max(1, floor)
	w := generateItemForClassSlot(class, SlotWeapon, itemFloor)
	w.UpgradeLevel = smithyLvl
	h.Weapon = &w

	head := generateItemForClassSlot(class, SlotHead, itemFloor)
	h.Head = &head
	ch := generateItemForClassSlot(class, SlotChest, itemFloor)
	h.Chest = &ch
	legs := generateItemForClassSlot(class, SlotLegs, itemFloor)
	h.Legs = &legs

	return h
}

func initialModelWithLegacy(legacy TownLegacy) Model {
	rand.Seed(time.Now().UnixNano())

	classes := []HeroClass{ClassTank, ClassWarrior, ClassRogue, ClassMage, ClassCleric}
	var heroes []*Hero
	for _, c := range classes {
		heroes = append(heroes, createHero(c, 1, legacy.SmithyLevel))
	}

	activeRelic := generateRelic(1)

	m := Model{
		State:            StateMenu,
		MenuCountdown:    20,
		Party:            heroes,
		Bag:              []EquipItem{},
		BagLevel:         0,
		Gold:             50 + legacy.TreasuryGold,
		Floor:            1,
		InTown:           false,
		TownPhase:        TownPhaseSellLoot,
		TownDialog:       "Отряд вошел через ворота Столицы на привал.",
		Combat:           nil,
		Logs:             []string{"[Хроники] Отряд ступил во мрак подземелья dccag."},
		AutoMode:         true,
		SpeedMs:          260,
		TownDelayMs:      2200,
		StatsScroll:      0,
		LogScroll:        0,
		RestartCountdown: 10,
		CurrentQuest:     generateAutoQuest(1),
		Relic:            &activeRelic,
		Legacy:           legacy,
		Stats:            newStats(),
		Packs:            make(map[Point]*MonsterPack),
		TermWidth:        120,
		TermHeight:       36,
	}
	m.Stats.TotalGoldEarned = 50 + legacy.TreasuryGold
	m.initDungeonForFloor(1)
	return m
}

func initialModel() Model {
	return initialModelWithLegacy(TownLegacy{TreasuryGold: 0, SmithyLevel: 0, ChurchLevel: 0, TavernLevel: 0})
}

func (m *Model) currentBagCapacity() int {
	return bagUpgrades[m.BagLevel].Capacity
}

func (m *Model) calculateDungeonSize(floor int) (int, int) {
	w := 70 + (floor-1)*3
	h := 30 + (floor-1)*1
	if w > 130 {
		w = 130
	}
	if h > 55 {
		h = 55
	}
	return w, h
}

func (m *Model) initDungeonForFloor(floor int) {
	w, h := m.calculateDungeonSize(floor)
	m.MapWidth = w
	m.MapHeight = h
	m.Grid = make([][]Tile, h)
	m.Explored = make([][]bool, h)
	for y := 0; y < h; y++ {
		m.Grid[y] = make([]Tile, w)
		m.Explored[y] = make([]bool, w)
	}
	m.generateDungeon()
	m.revealFog()
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, menuTickCmd())
}

func (m *Model) generateDungeon() {
	m.Packs = make(map[Point]*MonsterPack)
	m.Combat = nil
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			m.Grid[y][x] = TileWall
			m.Explored[y][x] = false
		}
	}

	type Rect struct{ X, Y, W, H int }
	rooms := []Rect{}
	attempts := (m.MapWidth * m.MapHeight) / 50
	if attempts < 8 {
		attempts = 8
	}

	for i := 0; i < attempts; i++ {
		w := rand.Intn(7) + 7
		h := rand.Intn(4) + 4
		if m.MapWidth-w-2 <= 1 || m.MapHeight-h-2 <= 1 {
			continue
		}
		x := rand.Intn(m.MapWidth-w-2) + 1
		y := rand.Intn(m.MapHeight-h-2) + 1
		rooms = append(rooms, Rect{x, y, w, h})
		for ry := y; ry < y+h; ry++ {
			for rx := x; rx < x+w; rx++ {
				m.Grid[ry][rx] = TileFloor
			}
		}
	}

	if len(rooms) == 0 {
		return
	}

	carveWideCorridor := func(x1, y1, x2, y2 int) {
		cx, cy := x1, y1
		for cx != x2 {
			m.Grid[cy][cx] = TileFloor
			if cy+1 < m.MapHeight {
				m.Grid[cy+1][cx] = TileFloor
			}
			if x2 > cx {
				cx++
			} else {
				cx--
			}
		}
		for cy != y2 {
			m.Grid[cy][cx] = TileFloor
			if cx+1 < m.MapWidth {
				m.Grid[cy+1][cx] = TileFloor
			}
			if y2 > cy {
				cy++
			} else {
				cy--
			}
		}
	}

	for i := 0; i < len(rooms)-1; i++ {
		x1, y1 := rooms[i].X+rooms[i].W/2, rooms[i].Y+rooms[i].H/2
		x2, y2 := rooms[i+1].X+rooms[i+1].W/2, rooms[i+1].Y+rooms[i+1].H/2
		carveWideCorridor(x1, y1, x2, y2)
	}

	exitPos := Point{rooms[0].X + 1, rooms[0].Y + 1}
	if m.CurrentQuest.Type == QuestEscapeTrap && !m.CurrentQuest.Completed {
		m.Grid[exitPos.Y][exitPos.X] = TileFloor
	} else {
		m.Grid[exitPos.Y][exitPos.X] = TileExit
	}
	if exitPos.X+1 < rooms[0].X+rooms[0].W-1 {
		m.PartyPos = Point{exitPos.X + 1, exitPos.Y}
	} else {
		m.PartyPos = Point{exitPos.X, exitPos.Y + 1}
	}
	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileFloor

	endRoom := rooms[len(rooms)-1]
	m.Grid[endRoom.Y+endRoom.H/2][endRoom.X+endRoom.W/2] = TileStairs

	relicSpawned := false

	for i := 1; i < len(rooms); i++ {
		r := rooms[i]
		isBoss := (m.Floor%10 == 0 && i == len(rooms)-1)
		pos := Point{r.X + r.W/2, r.Y + r.H/2}
		if isBoss {
			m.Packs[pos] = spawnMonsterPack(true, m.Floor)
		} else {
			m.Packs[pos] = spawnMonsterPack(false, m.Floor)
			eventRoll := rand.Intn(13)
			cornerPos := Point{r.X + 1, r.Y + 1}
			switch {
			case (!relicSpawned && m.CurrentQuest.Type == QuestFindRelic) || (eventRoll == 12 && !relicSpawned):
				m.Grid[cornerPos.Y][cornerPos.X] = TileRelic
				relicSpawned = true
			case eventRoll == 0:
				m.Grid[cornerPos.Y][cornerPos.X] = TileAltar
			case eventRoll == 1:
				m.Grid[cornerPos.Y][cornerPos.X] = TileFountain
			case eventRoll == 2:
				m.Grid[cornerPos.Y][cornerPos.X] = TileTrappedChest
			case eventRoll == 3:
				m.Grid[cornerPos.Y][cornerPos.X] = TileBarrel
			case eventRoll >= 7:
				m.Grid[cornerPos.Y][cornerPos.X] = TileChest
			}
		}
	}
}

func (m *Model) revealFog() {
	radius := 6
	for y := m.PartyPos.Y - radius; y <= m.PartyPos.Y+radius; y++ {
		for x := m.PartyPos.X - radius; x <= m.PartyPos.X+radius; x++ {
			if x >= 0 && x < m.MapWidth && y >= 0 && y < m.MapHeight {
				dx := x - m.PartyPos.X
				dy := y - m.PartyPos.Y
				if dx*dx+dy*dy <= radius*radius {
					m.Explored[y][x] = true
				}
			}
		}
	}
}

func (m *Model) addLog(msg string) {
	m.Logs = append(m.Logs, msg)
	if len(m.Logs) > 100 {
		m.Logs = m.Logs[len(m.Logs)-100:]
	}
	m.LogScroll = 0
}

func (m *Model) distributePartyExp(expAmt int) {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return
	}

	expPerHero := expAmt / len(living)
	if expPerHero < 1 {
		expPerHero = 1
	}

	for _, h := range living {
		if h.GainExp(expPerHero) {
			m.addLog(healStyle.Render(fmt.Sprintf("⭐ [УРОВЕНЬ] %s достиг Ур.%d! Характеристики возросли!", h.Name, h.Level)))
		}
	}
}

func (m *Model) needsHealing() bool {
	for _, h := range m.Party {
		if !h.IsDead && (float64(h.HP)/float64(h.MaxHP) <= 0.50 || h.Stress >= 70) {
			return true
		}
	}
	return false
}

func (m *Model) checkRetreat() bool {
	if len(m.Bag) >= m.currentBagCapacity() {
		return true
	}

	living := 0
	for _, h := range m.Party {
		if !h.IsDead {
			living++
			if float64(h.HP)/float64(h.MaxHP) <= 0.35 || h.Stress >= 130 {
				return true
			}
		}
	}
	if living <= 2 {
		return true
	}
	return false
}

func (m *Model) isPartyWiped() bool {
	for _, h := range m.Party {
		if !h.IsDead {
			return false
		}
	}
	return true
}

func (m *Model) checkQuestProgress(action QuestType, targetMob MonsterType, val int) {
	if m.CurrentQuest.Completed {
		return
	}
	if m.CurrentQuest.Type == action {
		if action == QuestHuntMonster && m.CurrentQuest.TargetMob != targetMob {
			return
		}
		if action == QuestReachFloor {
			if val >= m.CurrentQuest.TargetCount {
				m.CurrentQuest.Current = m.CurrentQuest.TargetCount
				m.CurrentQuest.Completed = true
				m.addLog(questStyle.Render(fmt.Sprintf("📜 [КОНТРАКТ] %s выполнен!", m.CurrentQuest.Title)))
			}
			return
		}
		m.CurrentQuest.Current += val
		if m.CurrentQuest.Current >= m.CurrentQuest.TargetCount {
			m.CurrentQuest.Current = m.CurrentQuest.TargetCount
			m.CurrentQuest.Completed = true
			m.addLog(questStyle.Render(fmt.Sprintf("📜 [КОНТРАКТ] %s выполнен!", m.CurrentQuest.Title)))
		}
	}
}

func (m *Model) findNextStep() Point {
	retreat := m.checkRetreat()
	seekingFountain := m.needsHealing()
	forceDeeper := (m.CurrentQuest.Type == QuestEscapeTrap) && !m.CurrentQuest.Completed

	hasFountainOnMap := false
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			if m.Grid[y][x] == TileFountain {
				hasFountainOnMap = true
				break
			}
		}
		if hasFountainOnMap {
			break
		}
	}

	findPath := func(avoidMonsters bool, targetCondition func(Point, Tile) bool) (Point, bool) {
		queue := []Point{m.PartyPos}
		visited := make(map[Point]bool)
		cameFrom := make(map[Point]Point)
		visited[m.PartyPos] = true
		dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr != m.PartyPos && targetCondition(curr, m.Grid[curr.Y][curr.X]) {
				step := curr
				for cameFrom[step] != m.PartyPos {
					step = cameFrom[step]
				}
				return step, true
			}

			for _, d := range dirs {
				next := Point{curr.X + d.X, curr.Y + d.Y}
				if next.X >= 0 && next.X < m.MapWidth && next.Y >= 0 && next.Y < m.MapHeight {
					if !visited[next] && m.Grid[next.Y][next.X] != TileWall {
						if avoidMonsters {
							if _, hasMob := m.Packs[next]; hasMob {
								continue
							}
						}
						visited[next] = true
						cameFrom[next] = curr
						queue = append(queue, next)
					}
				}
			}
		}
		return m.PartyPos, false
	}

	if seekingFountain && hasFountainOnMap {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileFountain
		})
		if found {
			return step
		}
	}

	if retreat && !forceDeeper {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
		step, found = findPath(false, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
	}

	isExplorationQuest := m.CurrentQuest.Type == QuestReachFloor && !m.CurrentQuest.Completed

	isTarget := func(p Point, t Tile) bool {
		_, hasPack := m.Packs[p]
		if hasPack {
			return true
		}
		if seekingFountain && t == TileAltar {
			return false
		}
		if t == TileStairs && isExplorationQuest {
			return true
		}
		if t == TileStairs && !forceDeeper {
			hasVisibleLoot := false
			for y := 0; y < m.MapHeight; y++ {
				for x := 0; x < m.MapWidth; x++ {
					if m.Explored[y][x] {
						tile := m.Grid[y][x]
						if tile == TileChest || tile == TileRelic {
							hasVisibleLoot = true
							break
						}
					}
				}
			}
			if hasVisibleLoot {
				return false
			}
		}
		return t == TileChest || t == TileStairs || t == TileAltar || t == TileFountain || t == TileTrappedChest || t == TileBarrel || t == TileRelic
	}

	if seekingFountain {
		step, found := findPath(true, isTarget)
		if found {
			return step
		}
	}

	step, found := findPath(false, isTarget)
	if found {
		return step
	}

	return m.PartyPos
}

func (m *Model) equipOrBag(item EquipItem) {
	var bestHero *Hero
	maxDiff := -99999

	for _, h := range m.Party {
		if h.IsDead {
			continue
		}
		if item.AllowedClass != "" && item.AllowedClass != h.Class {
			continue
		}

		currItem := h.GetItemInSlot(item.Slot)
		currStat := 0
		if currItem != nil {
			currStat = currItem.TotalStat()
		}

		diff := item.TotalStat() - currStat
		if diff > maxDiff {
			maxDiff = diff
			bestHero = h
		}
	}

	if bestHero != nil && maxDiff > 0 {
		oldItem := bestHero.GetItemInSlot(item.Slot)
		if oldItem != nil {
			m.Bag = append(m.Bag, *oldItem)
		}
		newItem := item
		bestHero.SetItemInSlot(newItem.Slot, &newItem)
		m.addLog(healStyle.Render(fmt.Sprintf("✨ %s сменил [%s] на [%s] (Мощь: %d)!",
			bestHero.Name, item.Slot, newItem.DisplayName(), newItem.TotalStat())))
	} else {
		m.Bag = append(m.Bag, item)
		m.addLog(subtleStyle.Render(fmt.Sprintf("📦 %s сложен в сумку.", item.DisplayName())))
	}
}

func (m *Model) recordFallenHero(h *Hero) {
	m.Stats.FallenHeroes = append(m.Stats.FallenHeroes, FallenHeroRecord{
		FullName: h.FullName(),
		Class:    h.Class,
		Cause:    h.CauseOfDeath,
		Floor:    m.Floor,
	})

	h.IsDead = true
	h.IsGuarding = false
	h.IsBerserk = false
	h.IsStealthed = false
	h.IsCharged = false
	h.IsAura = false
	h.Weapon = nil
	h.Head = nil
	h.Chest = nil
	h.Legs = nil

	if m.Relic != nil && m.Relic.MartyrFury {
		m.addLog(fireStyle.Render("👑 [Корона] Ярость павшего усилила живых (+4 Atk)!"))
		for _, ally := range m.Party {
			if !ally.IsDead {
				ally.BaseAtk += 4
			}
		}
	}
}

func (m *Model) handleAltar() {
	m.Stats.AltarsUsed++
	m.checkQuestProgress(QuestUseAltar, "", 1)
	d20 := rand.Intn(20) + 1
	isEven := d20%2 == 0
	target := m.getRandomLivingHero()

	if target != nil {
		bloodCost := 10
		if !isEven {
			bloodCost = 16
		}
		target.HP -= bloodCost
		target.BaseAtk += 2
		m.addStress(target, 15)

		if target.HP <= 0 {
			target.HP = 0
			target.CauseOfDeath = "Принесен в жертву Алтарю"
			m.recordFallenHero(target)
			m.addLog(dangerStyle.Render(fmt.Sprintf("🩸 %s пал жертвой Алтаря!", target.Name)))
		} else {
			m.addLog(altarStyle.Render(fmt.Sprintf("🩸 %s пожертвовал %d HP (+2 Atk)!", target.Name, bloodCost)))
		}
	}
}

func (m *Model) handleFountain() {
	m.Stats.FountainsUsed++
	for _, h := range m.Party {
		if !h.IsDead {
			h.HP = h.MaxHP
			h.MP = h.MaxMP
			h.Stress = 0
			h.Affliction = AfflictionNone
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
		}
	}
	m.addLog(fountStyle.Render("💧 [Источник] Здоровье, мана и рассудок отряда восстановлены!"))
}

func (m *Model) handleTrappedChest() {
	m.Stats.ChestsOpened++
	m.checkQuestProgress(QuestOpenChests, "", 1)
	d20 := rand.Intn(20) + 1

	rogueBonus := 0
	for _, h := range m.Party {
		if h.Class == ClassRogue && !h.IsDead {
			rogueBonus = 3
			break
		}
	}

	rollSuccess := (d20+rogueBonus >= 12)

	if rollSuccess {
		m.Stats.TrapsDisarmed++
		gold := int(float64(rand.Intn(25)+15) * m.Relic.GoldMult * 2)
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			lh.RevealSecret()
			m.checkAndAwardTitle(lh)
		}
		rareItem := generateItemForClass(targetClass, m.Floor+1)
		m.addLog(goldStyle.Render(fmt.Sprintf("🗝️ [Ларь] Ловушка снята: +%dG и [%s]!", gold, rareItem.DisplayName())))
		m.equipOrBag(rareItem)
	} else {
		m.addLog(dangerStyle.Render(fmt.Sprintf("💥 [Ловушка!] (D20=%d) Взрыв нанес урон!", d20)))
		for _, h := range m.Party {
			if !h.IsDead {
				trapDmg := rand.Intn(8) + 6
				h.HP -= trapDmg
				m.addStress(h, 20)
				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = "Взорван ловушкой"
					m.recordFallenHero(h)
				}
			}
		}
	}
}

func (m *Model) handleRelicTile() {
	m.checkQuestProgress(QuestFindRelic, "", 1)
	newLevel := 2
	if m.Floor >= 6 {
		newLevel = 3
	}
	newRelic := generateRelic(newLevel)

	if m.Relic == nil || newRelic.Level > m.Relic.Level {
		m.Relic = &newRelic
		m.addLog(relicTileStyle.Render(fmt.Sprintf("✨ [РЕЛИКВИЯ] Отряд нашел %s!", newRelic.Name)))
	} else {
		m.Gold += 80
		m.Stats.TotalGoldEarned += 80
		m.addLog(goldStyle.Render("✨ [Реликварий] Реликвия разобрана на +80G!"))
	}
}

func (m *Model) checkAndDrinkPotions(h *Hero) {
	if h.IsDead || h.Potion == nil || m.InTown {
		return
	}

	shouldDrink := false
	p := *h.Potion

	switch p.Type {
	case PotionHP:
		missingHP := h.MaxHP - h.HP
		if float64(h.HP)/float64(h.MaxHP) <= 0.45 || missingHP >= p.Power {
			shouldDrink = true
		}
	case PotionMP:
		missingMP := h.MaxMP - h.MP
		skillNeeded := h.SkillCost
		if h.MP < skillNeeded || (h.MaxMP > 0 && float64(h.MP)/float64(h.MaxHP) <= 0.35) || missingMP >= p.Power {
			shouldDrink = true
		}
	case PotionStress:
		if h.Stress >= 60 || h.Affliction != AfflictionNone {
			shouldDrink = true
		}
	}

	if shouldDrink {
		h.Potion = nil
		switch p.Type {
		case PotionHP:
			h.HP += p.Power
			if h.HP > h.MaxHP {
				h.HP = h.MaxHP
			}
			m.addLog(potionStyle.Render(fmt.Sprintf("🧪 %s выпил [%s] (+%d HP)!", h.Name, getPotionName(p), p.Power)))
		case PotionMP:
			h.MP += p.Power
			if h.HP > h.MaxHP {
				h.HP = h.MaxHP
			}
			m.addLog(potionStyle.Render(fmt.Sprintf("🧪 %s выпил [%s] (+%d MP)!", h.Name, getPotionName(p), p.Power)))
		case PotionStress:
			h.Stress -= p.Power
			if h.Stress < 0 {
				h.Stress = 0
			}
			h.Affliction = AfflictionNone
			m.addLog(potionStyle.Render(fmt.Sprintf("🧪 %s принял [%s] (-%d стресса)!", h.Name, getPotionName(p), p.Power)))
		}
	}
}

// --- Интеллектуальный выбор цели монстром и защита соратников ---

func (m *Model) SelectTarget() *Hero {
	var candidates []*Hero
	totalWeight := 0

	for _, h := range m.Party {
		if !h.IsDead {
			w := h.Role.AggroWeight
			if h.HP < h.MaxHP/3 {
				w += 20
			}
			candidates = append(candidates, h)
			totalWeight += w
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	r := rand.Intn(totalWeight)
	curr := 0
	for _, h := range candidates {
		w := h.Role.AggroWeight
		if h.HP < h.MaxHP/3 {
			w += 20
		}
		curr += w
		if r < curr {
			return h
		}
	}
	return candidates[0]
}

func (m *Model) ApplyDamage(target *Hero, rawDmg int) (actual *Hero, finalDmg int, guarded bool) {
	if target.Class != ClassTank {
		for _, guard := range m.Party {
			if !guard.IsDead && guard.Role.CanGuard && guard != target && guard.HP > guard.MaxHP/4 {
				chance := 35
				mitigation := 0.8
				if guard.Class == ClassTank {
					chance = 65
					mitigation = 0.55
				}

				if rand.Intn(100) < chance {
					mitigated := int(float64(rawDmg) * mitigation)
					guard.HP -= mitigated
					globalDebugReport.DamageGuarded += (rawDmg - mitigated)
					guard.AddBlock()
					return guard, mitigated, true
				}
			}
		}
	}

	target.HP -= rawDmg
	return target, rawDmg, false
}

func (m *Model) startCombat(pos Point, pack *MonsterPack) {
	combat := &ActiveCombat{
		Pos:          pos,
		Pack:         pack,
		HasBarrel:    rand.Intn(100) < 35,
		FleeCooldown: 0,
		Round:        1,
	}

	biome := getBiome(m.Floor)
	speedPenalty := 0
	if biome.Name == BiomeGrotto {
		speedPenalty = 2
	}

	for _, h := range m.Party {
		if !h.IsDead {
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
			initRoll := rand.Intn(20) + 1 + h.Speed - speedPenalty
			combat.TurnQueue = append(combat.TurnQueue, TurnOrderEntry{
				Type: CombatantHero, HeroRef: h, Initiative: initRoll,
			})
		}
	}

	for _, mob := range pack.Members {
		if !mob.IsDead {
			initRoll := rand.Intn(20) + 1 + mob.Speed
			combat.TurnQueue = append(combat.TurnQueue, TurnOrderEntry{
				Type: CombatantMonster, MonsterRef: mob, Initiative: initRoll,
			})
		}
	}

	sort.Slice(combat.TurnQueue, func(i, j int) bool {
		return combat.TurnQueue[i].Initiative > combat.TurnQueue[j].Initiative
	})

	m.Combat = combat
	m.addLog(accentStyle.Render(fmt.Sprintf("⚔️ СХВАТКА! Вражеский отряд (%d тварей)!", pack.LivingCount())))
}

func (m *Model) shouldAttemptFlee() bool {
	if m.Combat == nil || m.Combat.FleeCooldown > 0 {
		return false
	}
	curHP, maxHP := 0, 0
	livingCount := 0
	for _, h := range m.Party {
		if !h.IsDead {
			curHP += h.HP
			maxHP += h.MaxHP
			livingCount++
		}
	}
	if maxHP == 0 || livingCount == 0 {
		return false
	}
	hpPercent := float64(curHP) / float64(maxHP)
	return hpPercent < 0.25 || (livingCount == 1 && hpPercent < 0.50)
}

func (m *Model) attemptFlee() {
	globalDebugReport.FleeAttempts++
	chance := 50

	for _, h := range m.Party {
		if !h.IsDead {
			if h.Class == ClassRogue {
				chance += 25
			}
			if h.Class == ClassTank {
				chance += 10
			}
		}
	}

	chance -= m.Floor / 3
	if chance < 35 {
		chance = 35
	}

	roll := rand.Intn(100)
	if roll < chance {
		globalDebugReport.FleeSuccesses++
		m.addLog(healStyle.Render("💨 [ПОБЕГ] Успех! Отряд оторвался от погони под прикрытием завесы."))

		evacuatedCount := 0
		for _, h := range m.Party {
			if h.IsDead {
				h.HP = 1
				h.IsDead = false
				h.CauseOfDeath = ""
				evacuatedCount++
			}
		}
		if evacuatedCount > 0 {
			m.addLog(altarStyle.Render(fmt.Sprintf("🕊️ [Эвакуация] Отряд вынес с поля боя %d павших героев!", evacuatedCount)))
		}

		m.Combat = nil
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = "Отряд отступил в Столицу."
		m.addLog(dangerStyle.Render("🏰 Отряд укрылся за стенами Города!"))
	} else {
		m.addLog(dangerStyle.Render("💥 [ПРОВАЛ ПОБЕГА] Монстры перекрыли отход! Отряд перегруппировался под градом скользящих ударов."))

		for _, h := range m.Party {
			if !h.IsDead {
				chipDamage := int(float64(h.HP) * 0.15)
				if chipDamage < 1 {
					chipDamage = 1
				}
				h.HP -= chipDamage
				h.Feats.DamageTaken += chipDamage
				m.addStress(h, 8)

				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = "Зарублен при неудачном отходе"
					m.recordFallenHero(h)
				}
			}
		}
		m.Combat.FleeCooldown = 3
	}
}

func (m *Model) executeCombatTurn() {
	if m.Combat == nil {
		return
	}

	if m.Combat.FleeCooldown > 0 {
		m.Combat.FleeCooldown--
	}

	if m.shouldAttemptFlee() {
		m.attemptFlee()
		if m.Combat == nil {
			return
		}
	}

	if m.Combat.Pack.LivingCount() == 0 {
		m.addLog(healStyle.Render("💀 Вражеский отряд повержен!"))
		delete(m.Packs, m.Combat.Pos)
		m.PartyPos = m.Combat.Pos
		m.Combat = nil
		m.revealFog()
		return
	}

	if m.isPartyWiped() {
		m.State = StateDefeat
		m.Combat = nil
		m.RestartCountdown = 10
		return
	}

	if m.Combat.TurnIdx >= len(m.Combat.TurnQueue) {
		m.Combat.TurnIdx = 0
		m.Combat.Round++
	}
	current := m.Combat.TurnQueue[m.Combat.TurnIdx]
	m.Combat.TurnIdx++

	biome := getBiome(m.Floor)

	enrageMult := 1.0
	if m.Combat.Round > 20 {
		globalDebugReport.EnrageProcs++
		enrageMult += float64(m.Combat.Round-20) * 0.10
	}

	if current.Type == CombatantHero {
		h := current.HeroRef
		if h.IsDead {
			return
		}

		m.checkAndDrinkPotions(h)

		if h.Affliction == AfflictionParanoid && rand.Intn(100) < 30 {
			m.addLog(stressStyle.Render(fmt.Sprintf("👁️ %s забился в угол (Паранойя)!", h.Name)))
			return
		}

		skillCost := h.SkillCost
		if biome.Name == BiomeCrystal {
			skillCost += 3
		}

		targetMob := m.Combat.Pack.GetLowestHPFocus()

		// 1. ТАНК
		if h.Class == ClassTank {
			if h.MP >= skillCost && !h.IsGuarding {
				h.MP -= skillCost
				h.IsGuarding = true
				h.AddBlock()
				h.PullAggro()
				m.checkAndAwardTitle(h)
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s принимает [Оборонительную Стойку] (+5 Def, блок)!", h.Name)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.Atk >= 12 && rand.Intn(100) < 45 {
				h.MP -= 6
				bashDmg := h.TotalDef() + 4
				targetMob.HP -= bashDmg
				targetMob.Atk = max(2, targetMob.Atk-3)
				h.AddBlock()
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s проводит [Удар щитом] (-%d HP, враг ослаблен на -3 Atk)!", h.Name, bashDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 2. ВОИН
		if h.Class == ClassWarrior {
			if h.MP >= skillCost && !h.IsBerserk {
				h.MP -= skillCost
				h.IsBerserk = true
				m.addLog(fireStyle.Render(fmt.Sprintf("⚔️ %s входит в [Состояние Ярости] (+5 Atk, -2 Def)!", h.Name)))
			} else if h.MP >= 7 && m.Combat.Pack.LivingCount() >= 2 && rand.Intn(100) < 55 {
				h.MP -= 7
				cleaveDmg := h.TotalAtk() + 2
				hitCount := 0
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead && hitCount < 2 {
						mob.HP -= cleaveDmg
						h.Feats.DamageDealt += cleaveDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
						}
						hitCount++
					}
				}
				m.addLog(dangerStyle.Render(fmt.Sprintf("⚔️ %s выполняет [Рассечение] по %d врагам (-%d HP)!", h.Name, hitCount, cleaveDmg)))
				return
			}
		}

		// 3. РАЗБОЙНИК
		if h.Class == ClassRogue {
			if h.MP >= skillCost && !h.IsStealthed {
				h.MP -= skillCost
				h.IsStealthed = true
				h.AddBackstab()
				m.addLog(accentStyle.Render(fmt.Sprintf("🗡️ %s растворился в тенях [Скрытность] (100%% крит)!", h.Name)))
			} else if h.MP >= 6 && targetMob != nil && targetMob.HP > 20 && rand.Intn(100) < 50 {
				h.MP -= 6
				poisonDmg := h.TotalAtk() + 6
				targetMob.HP -= poisonDmg
				h.Feats.DamageDealt += poisonDmg
				m.addLog(stressStyle.Render(fmt.Sprintf("☣️ %s наносит [Отравленный выпад] (-%d HP)!", h.Name, poisonDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 4. КЛИРИК
		if h.Class == ClassCleric {
			var criticalAlly *Hero
			for _, ally := range m.Party {
				if !ally.IsDead && float64(ally.HP)/float64(ally.MaxHP) <= 0.50 {
					criticalAlly = ally
					break
				}
			}

			if criticalAlly != nil && h.MP >= skillCost && h.Affliction != AfflictionSelfish {
				h.MP -= skillCost
				hAmt := rand.Intn(10) + 16 + (m.Floor * 3)
				if m.Combat.Round > 20 {
					hAmt /= 2
				}
				criticalAlly.HP += hAmt
				if criticalAlly.HP > criticalAlly.MaxHP {
					criticalAlly.HP = criticalAlly.MaxHP
				}
				criticalAlly.Stress = max(0, criticalAlly.Stress-12)
				h.Feats.HealsGiven += hAmt
				m.addLog(healStyle.Render(fmt.Sprintf("✨ %s исцелил %s (+%d HP)!", h.Name, criticalAlly.Name, hAmt)))
				m.checkAndAwardTitle(h)
				return
			} else if !h.IsAura && h.MP >= (skillCost+6) {
				h.MP -= skillCost
				h.IsAura = true
				m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s раскрывает [Ауру Защиты] (+3 Def отряду)!", h.Name)))
			} else if h.MP >= 7 && targetMob != nil && rand.Intn(100) < 40 {
				h.MP -= 7
				smiteDmg := h.TotalAtk() + 4
				targetMob.HP -= smiteDmg
				if lowest := m.getRandomLivingHero(); lowest != nil && lowest.HP < lowest.MaxHP {
					lowest.HP = min(lowest.MaxHP, lowest.HP+6)
				}
				m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s обрушил [Священную кару] (-%d HP)!", h.Name, smiteDmg)))
				if targetMob.HP <= 0 {
					targetMob.HP = 0
					targetMob.IsDead = true
					h.Feats.Kills++
					m.distributePartyExp(targetMob.Exp)
				}
				return
			}
		}

		// 5. МАГ
		if h.Class == ClassMage {
			if m.Combat.HasBarrel && h.MP >= skillCost {
				h.MP -= skillCost
				m.Combat.HasBarrel = false
				m.Stats.BarrelsBlown++
				h.AddManaBurst()
				barrelDmg := 22 + (m.Floor * 3)
				m.addLog(barrelStyle.Render(fmt.Sprintf("💥 %s ПОДРЫВАЕТ БОЧКУ СО СМОЛОЙ (-%d HP отряду врагов)!", h.Name, barrelDmg)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= barrelDmg
						h.Feats.DamageDealt += barrelDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult * 2)
							m.Gold += g
							m.Stats.TotalGoldEarned += g
							m.Stats.MonsterKills[mob.Type]++
							m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)
						}
					}
				}
				m.checkAndAwardTitle(h)
				return
			} else if m.Combat.Pack.LivingCount() >= 2 && h.MP >= skillCost {
				h.MP -= skillCost
				aoeDmg := h.TotalAtk() + 6
				h.AddManaBurst()
				h.AddCCDuration()
				m.addLog(fireStyle.Render(fmt.Sprintf("🔥 %s накрыл врагов [Огненной Бурей]!", h.Name)))
				for _, mob := range m.Combat.Pack.Members {
					if !mob.IsDead {
						mob.HP -= aoeDmg
						h.Feats.DamageDealt += aoeDmg
						if mob.HP <= 0 {
							mob.HP = 0
							mob.IsDead = true
							h.Feats.Kills++
							m.distributePartyExp(mob.Exp)
							g := int(float64(mob.Exp) * m.Relic.GoldMult * 2)
							m.Gold += g
							m.Stats.TotalGoldEarned += g
							m.Stats.MonsterKills[mob.Type]++
							m.checkQuestProgress(QuestHuntMonster, mob.Type, 1)
						}
					}
				}
				m.checkAndAwardTitle(h)
				return
			} else {
				if eliteMob := m.Combat.Pack.GetHighestHPFocus(); eliteMob != nil {
					targetMob = eliteMob
				}
			}
		}

		if targetMob == nil {
			return
		}

		if (targetMob.Type == MobSkeleton || targetMob.Type == MobGolem || targetMob.Type == MobGargoyle) && rand.Intn(100) < 20 {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ [%s] отразил выпад монолитным блоком!", targetMob.Name)))
			return
		}

		d20 := rand.Intn(20) + 1
		hitRoll := d20 + (h.TotalAtk() / 3)
		targetAC := 10 + targetMob.Defense
		critThreshold := 19
		for _, it := range []*EquipItem{h.Weapon, h.Head, h.Chest} {
			if it != nil {
				critThreshold -= it.CritBonus
			}
		}
		if h.IsStealthed {
			critThreshold = 1
		}
		if critThreshold < 14 && !h.IsStealthed {
			critThreshold = 14
		}
		bonusDmg := 0
		armorPierce := 0

		switch h.Class {
		case ClassMage:
			if h.MP >= 5 {
				h.MP -= 5
				bonusDmg += 5
				m.addLog(fireStyle.Render(fmt.Sprintf("🔥 %s выпускает [Огненную стрелу] (+5 ур)!", h.Name)))
			}
		case ClassCleric:
			if h.MP >= 4 {
				h.MP -= 4
				for _, ally := range m.Party {
					if !ally.IsDead && ally.Stress > 0 {
						ally.Stress = max(0, ally.Stress-6)
						h.RemoveDot()
						m.addLog(fountStyle.Render(fmt.Sprintf("✨ %s: [Благословение] (-6 стресса %s)!", h.Name, ally.Name)))
						break
					}
				}
			}
		case ClassWarrior:
			if h.MP >= 4 && h.MP < skillCost {
				h.MP -= 4
				bonusDmg += 3
				armorPierce = 3
				m.addLog(dangerStyle.Render(fmt.Sprintf("⚔️ %s: [Сокрушающий выпад]!", h.Name)))
			}
		case ClassRogue:
			if h.MP >= 3 && h.MP < skillCost {
				h.MP -= 3
				critThreshold = 17
				h.StealLoot()
				m.addLog(accentStyle.Render(fmt.Sprintf("🗡️ %s: [Быстрый порез]!", h.Name)))
			}
		case ClassTank:
			if h.MP >= 4 {
				h.MP -= 4
				h.BaseDef += 2
				h.PullAggro()
				m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s: [Провокация] (+2 Защ)!", h.Name)))
			}
		}

		isCrit := d20 >= critThreshold || h.IsStealthed
		isFumble := d20 == 1 && !h.IsStealthed

		if isFumble {
			m.addLog(subtleStyle.Render(fmt.Sprintf("💨 %s промахнулся (D20=1)!", h.Name)))
			return
		}
		if hitRoll < targetAC && !isCrit {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ Броня [%s] отразила удар %s.", targetMob.Name, h.Name)))
			return
		}

		effectiveDef := max(0, targetMob.Defense-armorPierce)
		varRand := max(4, h.TotalAtk()/2)
		dmg := h.TotalAtk() + rand.Intn(varRand) - (effectiveDef / 2) + bonusDmg
		dmg = max(4+(m.Floor/2), dmg)

		if isCrit {
			dmg = int(float64(dmg) * 1.8)
			h.Feats.CritsLanded++
			h.AddCriticalStrike()
			m.checkAndAwardTitle(h)
			m.addLog(dangerStyle.Render(fmt.Sprintf("💥 КРИТ (D20=%d)! %s сокрушает врага!", d20, h.Name)))
			h.IsStealthed = false
		}

		targetMob.HP -= dmg
		h.Feats.DamageDealt += dmg
		m.checkAndAwardTitle(h)
		m.addLog(fmt.Sprintf("⚔️ %s нанес %d урона [%s] (%d HP).", h.Name, dmg, targetMob.Name, targetMob.HP))

		if targetMob.HP <= 0 {
			targetMob.HP = 0
			targetMob.IsDead = true
			h.Feats.Kills++
			m.distributePartyExp(targetMob.Exp)
			g := int(float64(targetMob.Exp) * m.Relic.GoldMult * 2)
			m.Gold += g
			m.Stats.TotalGoldEarned += g
			m.Stats.MonsterKills[targetMob.Type]++
			m.checkQuestProgress(QuestHuntMonster, targetMob.Type, 1)

			if targetMob.Type == MobDragon {
				h.Feats.BossKills++
				bossItem := generateItemForClass(h.Class, m.Floor+2)
				bossItem.UpgradeLevel = 4
				m.addLog(accentStyle.Render(fmt.Sprintf("👑 Реликвия Дракона: [%s]!", bossItem.DisplayName())))
				m.equipOrBag(bossItem)
			}
			m.checkAndAwardTitle(h)
		}
		return
	}

	if current.Type == CombatantMonster {
		mob := current.MonsterRef
		if mob.IsDead {
			return
		}

		victim := m.SelectTarget()
		if victim == nil {
			return
		}

		m.checkAndDrinkPotions(victim)

		mobRoll := rand.Intn(20) + 1
		mobHit := mobRoll + (mob.Atk / 3)
		heroAC := 10 + victim.TotalDef()

		if mobRoll == 1 {
			m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s парировал выпад [%s]!", victim.Name, mob.Name)))
			return
		}
		if mobHit < heroAC && mobRoll < 19 {
			m.addLog(subtleStyle.Render(fmt.Sprintf("🛡️ Доспехи %s выдержали удар.", victim.Name)))
			return
		}

		rawDmg := mob.Atk - (victim.TotalDef() / 2)
		if victim.IsGuarding {
			rawDmg = int(float64(rawDmg) * 0.6)
		}

		if m.Combat.Pack.LivingCount() >= 4 {
			rawDmg = int(float64(rawDmg) * 0.78)
		}

		if mobRoll >= 19 {
			rawDmg = int(float64(rawDmg) * 1.5)
			m.addStress(victim, 25)
		}

		inDmg := max(3, int(float64(rawDmg)*m.Relic.EnemyDmgMod*enrageMult))

		actualHero, dealtDmg, guarded := m.ApplyDamage(victim, inDmg)

		if guarded {
			m.addLog(healStyle.Render(fmt.Sprintf("🛡️ %s ПРИКРЫЛ СОБОЙ %s, приняв %d урона!", actualHero.Name, victim.Name, dealtDmg)))
		} else {
			m.addLog(dangerStyle.Render(fmt.Sprintf("🩸 [%s] нанес %d урона %s!", mob.Name, dealtDmg, actualHero.Name)))
		}

		actualHero.Feats.DamageTaken += dealtDmg
		m.addStress(actualHero, 8)

		if actualHero.HP <= 0 {
			actualHero.HP = 0
			actualHero.CauseOfDeath = fmt.Sprintf("Сражен монстром [%s]", mob.Name)
			m.recordFallenHero(actualHero)
			m.addLog(dangerStyle.Render(fmt.Sprintf("☠️ %s пал в бою от удара [%s]!", actualHero.Name, mob.Name)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					m.addStress(ally, 20)
				}
			}
		} else {
			if float64(actualHero.HP)/float64(actualHero.MaxHP) <= 0.20 {
				actualHero.Feats.NearDeathEscapes++
			}
			m.checkAndAwardTitle(actualHero)
		}
	}
}

func (m *Model) step() {
	if m.State != StatePlaying {
		return
	}
	if m.isPartyWiped() {
		m.State = StateDefeat
		m.RestartCountdown = 10
		return
	}

	m.Stats.TotalSteps++

	if m.Stats.TotalSteps > 0 && m.Stats.TotalSteps%25 == 0 {
		healedCount := 0
		for _, h := range m.Party {
			if h.IsDead {
				continue
			}
			changed := false
			if h.HP < h.MaxHP {
				h.HP++
				changed = true
			}
			if h.MP < h.MaxMP {
				h.MP++
				changed = true
			}
			if h.Stress > 0 {
				h.Stress--
				changed = true
			}
			if changed {
				healedCount++
			}
		}
		if healedCount > 0 {
			m.addLog(healStyle.Render("🌿 [Привал] В тишине подземелья отряд немного передохнул (+1 HP/MP, -1 Стресс)."))
		}
	}

	for _, h := range m.Party {
		m.checkAndDrinkPotions(h)
	}

	if m.Combat != nil {
		m.executeCombatTurn()
		return
	}

	if m.InTown {
		m.stepTown()
		return
	}

	if m.checkRetreat() && m.Grid[m.PartyPos.Y][m.PartyPos.X] == TileExit && m.Stats.TotalSteps > 0 {
		m.InTown = true
		m.TownPhase = TownPhaseSellLoot
		m.TownDialog = "Отряд вернулся в город."
		return
	}

	next := m.findNextStep()

	loopHit := false
	for _, p := range m.PathHistory {
		if p == next {
			m.LoopDetectCount++
			loopHit = true
			break
		}
	}
	if !loopHit {
		m.LoopDetectCount = 0
	}

	if m.LoopDetectCount > 4 {
		m.addLog(dangerStyle.Render("⚠️ [Коллизия] Экстренный прорыв к свободному залу."))
		foundSafeSpot := false
		for y := 0; y < m.MapHeight && !foundSafeSpot; y++ {
			for x := 0; x < m.MapWidth && !foundSafeSpot; x++ {
				if m.Grid[y][x] == TileFloor || m.Grid[y][x] == TileExit {
					dx := x - m.PartyPos.X
					dy := y - m.PartyPos.Y
					if dx*dx+dy*dy <= 25 && dx*dx+dy*dy > 1 {
						m.PartyPos = Point{x, y}
						foundSafeSpot = true
					}
				}
			}
		}
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.revealFog()
		return
	} else {
		m.PathHistory = append(m.PathHistory, next)
		if len(m.PathHistory) > 10 {
			m.PathHistory = m.PathHistory[1:]
		}
	}

	if pack, exists := m.Packs[next]; exists {
		m.startCombat(next, pack)
		return
	}

	if next == m.PartyPos {
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(fmt.Sprintf("🌟 ЭТАЖ %d ПРОЙДЕН! Бездна зовет...", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(fmt.Sprintf("Этаж %d пройден! Спуск глубже.", m.Floor)))
		}

		m.initDungeonForFloor(m.Floor)
		return
	}

	switch m.Grid[next.Y][next.X] {
	case TileChest:
		m.Stats.ChestsOpened++
		m.checkQuestProgress(QuestOpenChests, "", 1)
		gold := int(float64(rand.Intn(16)+10+(m.Floor*2)) * m.Relic.GoldMult * 2)
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			m.checkAndAwardTitle(lh)
		}
		item := generateItemForClass(targetClass, m.Floor)
		m.addLog(goldStyle.Render(fmt.Sprintf("🎁 Сундук: +%dG и [%s].", gold, item.DisplayName())))
		m.equipOrBag(item)
		m.Grid[next.Y][next.X] = TileFloor

	case TileRelic:
		m.handleRelicTile()
		m.Grid[next.Y][next.X] = TileFloor

	case TileAltar:
		m.handleAltar()
		m.Grid[next.Y][next.X] = TileFloor

	case TileFountain:
		m.handleFountain()
		m.Grid[next.Y][next.X] = TileFloor

	case TileTrappedChest:
		m.handleTrappedChest()
		m.Grid[next.Y][next.X] = TileFloor

	case TileBarrel:
		m.Grid[next.Y][next.X] = TileFloor

	case TileStairs:
		m.Floor++
		m.Stats.FloorsCleared++
		m.checkQuestProgress(QuestReachFloor, "", m.Floor)
		m.checkQuestProgress(QuestEscapeTrap, "", 1)

		if m.Floor%10 == 0 {
			m.addLog(accentStyle.Render(fmt.Sprintf("🌟 ЭТАЖ %d ПРОЙДЕН! Бездна зовет...", m.Floor)))
		} else {
			m.addLog(accentStyle.Render(fmt.Sprintf("Спуск на этаж %d!", m.Floor)))
		}

		m.initDungeonForFloor(m.Floor)
		return
	}

	m.PartyPos = next
	m.revealFog()
}

func (m *Model) getRandomLivingHero() *Hero {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return nil
	}
	return living[rand.Intn(len(living))]
}

// --- Пайплайн Столицы с бюджетированием и мутациями ---

func (m *Model) stepTown() {
	switch m.TownPhase {
	case TownPhaseSellLoot:
		soldGold := 0
		for _, it := range m.Bag {
			soldGold += it.Value * 2
		}
		if soldGold > 0 {
			m.Gold += soldGold
			m.Stats.TotalGoldEarned += soldGold
			m.addLog(goldStyle.Render(fmt.Sprintf("⚖️ [Рынок] Трофеи проданы на +%dG.", soldGold)))
		}
		m.Bag = []EquipItem{}

		for _, h := range m.Party {
			if !h.IsDead {
				h.Stress = max(0, h.Stress-40)
				if h.Stress < 60 {
					h.Affliction = AfflictionNone
				}
			}
		}
		m.addLog(healStyle.Render("🏰 [Город] Безопасные стены Столицы восстановили дух отряда."))
		m.TownPhase = TownPhaseMagistrate

	case TownPhaseMagistrate:
		budget := AllocateBudget(m.Gold)
		investAmt := budget.Treasury
		if investAmt > 0 {
			m.Gold -= investAmt
			globalDebugReport.GoldSpentBreakdown["Инвестиции в Магистрат"] += investAmt

			type Building struct {
				Name  string
				Level int
				Cost  int
			}
			buildings := []Building{
				{Name: "Кузница", Level: m.Legacy.SmithyLevel, Cost: 250 + m.Legacy.SmithyLevel*150},
				{Name: "Церковь", Level: m.Legacy.ChurchLevel, Cost: 200 + m.Legacy.ChurchLevel*120},
				{Name: "Таверна", Level: m.Legacy.TavernLevel, Cost: 180 + m.Legacy.TavernLevel*100},
			}
			sort.Slice(buildings, func(i, j int) bool { return buildings[i].Level < buildings[j].Level })

			targetBld := &buildings[0]
			switch targetBld.Name {
			case "Кузница":
				m.Legacy.SmithyLevel++
			case "Церковь":
				m.Legacy.ChurchLevel++
			case "Таверна":
				m.Legacy.TavernLevel++
			}
			m.addLog(titleStyle.Render(fmt.Sprintf("🏛️ [Магистрат] Отчислено %dG на развитие города («%s» Ур.%d)!",
				investAmt, targetBld.Name, targetBld.Level+1)))
		}
		m.TownPhase = TownPhaseUpgradeBag

	case TownPhaseUpgradeBag:
		budget := AllocateBudget(m.Gold)
		if m.BagLevel < len(bagUpgrades)-1 {
			nextBag := bagUpgrades[m.BagLevel+1]
			if budget.Bags >= nextBag.Cost && m.Gold >= nextBag.Cost {
				m.Gold -= nextBag.Cost
				globalDebugReport.GoldSpentBreakdown["Улучшение сумок"] += nextBag.Cost
				m.BagLevel++
				m.addLog(goldStyle.Render(fmt.Sprintf("🎒 [Кожевник] Куплен %s (Вместимость: %d слотов) за %dG!", nextBag.Name, nextBag.Capacity, nextBag.Cost)))
			}
		}
		m.TownPhase = TownPhaseChurch

	case TownPhaseChurch:
		hasDead := false
		hasStress := false
		for _, h := range m.Party {
			if h.HP == 1 || h.IsDead {
				hasDead = true
			}
			if h.Stress > 0 {
				hasStress = true
			}
		}

		if hasDead || hasStress {
			blessCost := max(80, (100*m.Floor)-(m.Legacy.ChurchLevel*30))
			if m.Gold >= blessCost {
				m.Gold -= blessCost
				globalDebugReport.GoldSpentBreakdown["Церковь (исцеление/воскрешение)"] += blessCost
				revived := 0
				for _, h := range m.Party {
					if h.HP == 1 || h.IsDead {
						h.IsDead = false
						h.HP = h.MaxHP / 2
						revived++
					}
					h.Stress = 0
					h.Affliction = AfflictionNone
				}
				m.addLog(fountStyle.Render(fmt.Sprintf("⛪ [Церковь] Служба исцеления проведена (-%dG, поднято: %d)!", blessCost, revived)))
			}
		}
		m.TownPhase = TownPhaseTavern

	case TownPhaseTavern:
		tavernCost := max(35, (20*m.Floor)-(m.Legacy.TavernLevel*8))
		if m.Gold >= tavernCost {
			m.Gold -= tavernCost
			globalDebugReport.GoldSpentBreakdown["Таверна (ночлег)"] += tavernCost
			for _, h := range m.Party {
				if !h.IsDead {
					h.HP = h.MaxHP
					h.MP = h.MaxMP
				}
			}
			m.addLog(healStyle.Render(fmt.Sprintf("🍻 [Таверна] Полноценный отдых (-%dG). Отряд полон сил.", tavernCost)))
		} else {
			barnCost := max(5, tavernCost/4)
			if m.Gold >= barnCost {
				m.Gold -= barnCost
				globalDebugReport.GoldSpentBreakdown["Таверна (сарай)"] += barnCost
			}
			for _, h := range m.Party {
				if !h.IsDead {
					h.HP = max(1, int(float64(h.MaxHP)*0.45))
					h.MP = int(float64(h.MaxMP) * 0.45)
				}
			}
			m.addLog(dangerStyle.Render("🏚️ [Сарай] Казна истощена! Ночлег в сарае (45% сил)."))
		}
		m.TownPhase = TownPhaseGuild

	case TownPhaseGuild:
		if m.CurrentQuest.Completed {
			reward := m.CurrentQuest.RewardGold * 2
			m.Gold += reward
			m.Stats.TotalGoldEarned += reward
			m.Stats.QuestsCompleted++
			m.addLog(questStyle.Render(fmt.Sprintf("📜 [Гильдия] Контракт закрыт: +%dG!", reward)))
			m.CurrentQuest = generateAutoQuest(m.Floor)
		}

		recruitCost := max(35, 50+(m.Floor*22)-(m.Legacy.ChurchLevel*6))
		for i, h := range m.Party {
			if h.IsDead {
				classes := []HeroClass{ClassTank, ClassWarrior, ClassRogue, ClassMage, ClassCleric}
				newClass := classes[rand.Intn(len(classes))]

				if m.Gold >= recruitCost {
					m.Gold -= recruitCost
					globalDebugReport.GoldSpentBreakdown["Найм ветеранов"] += recruitCost
					m.Party[i] = createHero(newClass, m.Floor, m.Legacy.SmithyLevel)
					m.addLog(healStyle.Render(fmt.Sprintf("⚔️ [Гильдия] Нанят ветеран %s (%s, Ур.%d) за %dG!", m.Party[i].Name, newClass, m.Party[i].Level, recruitCost)))
				} else {
					m.Party[i] = createHero(newClass, 1, 0)
					m.Party[i].Title = "Ополченец"
					m.addLog(subtleStyle.Render(fmt.Sprintf("🤝 [Гильдия] Ополченец %s (%s) встал в строй бесплатно.", m.Party[i].Name, newClass)))
				}
			}
		}
		m.TownPhase = TownPhaseSmithy

	case TownPhaseSmithy:
		budget := AllocateBudget(m.Gold)
		smithyBudget := budget.Forge
		upgradesCount := 0
		totalSpent := 0

		getUpgradeCost := func(it *EquipItem) int {
			if it == nil {
				return 999999
			}
			matMult := max(1, it.Material.ValueMult)
			nextLvl := it.UpgradeLevel + 1
			cost := (nextLvl * nextLvl * 35 * matMult) - (m.Legacy.SmithyLevel * 15)
			return max(30*matMult, cost)
		}

		for smithyBudget > 0 {
			var bestHero *Hero
			var bestSlot EquipSlot
			var bestItem *EquipItem
			minCost := 999999

			for _, h := range m.Party {
				if h.IsDead {
					continue
				}
				for _, slot := range []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs} {
					it := h.GetItemInSlot(slot)
					if it != nil && it.UpgradeLevel < 6 {
						cost := getUpgradeCost(it)
						if cost < minCost && smithyBudget >= cost && m.Gold >= cost {
							minCost = cost
							bestHero = h
							bestSlot = slot
							bestItem = it
						}
					}
				}
			}

			if bestItem == nil {
				break
			}

			m.Gold -= minCost
			smithyBudget -= minCost
			totalSpent += minCost
			bestItem.UpgradeLevel++
			bestHero.SetItemInSlot(bestSlot, bestItem)
			upgradesCount++
			m.Stats.UpgradesForged++
		}

		if totalSpent > 0 {
			globalDebugReport.GoldSpentBreakdown["Кузница (заточки)"] += totalSpent
			m.addLog(goldStyle.Render(fmt.Sprintf("⚒️ [Кузнец] Заточено снаряжения: %d шт. (-%dG)!", upgradesCount, totalSpent)))
		}
		m.TownPhase = TownPhaseAlchemist

	case TownPhaseAlchemist:
		budget := AllocateBudget(m.Gold)
		alchBudget := budget.Alchemy
		spentAlch := 0

		for alchBudget > 0 {
			var candidates []*Hero
			for _, h := range m.Party {
				if !h.IsDead {
					candidates = append(candidates, h)
				}
			}
			if len(candidates) == 0 {
				break
			}

			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].Mutations.Total() < candidates[j].Mutations.Total()
			})
			target := candidates[0]

			pref := GetClassMutationPreference(target.Class, target.Mutations)
			baseCost := 220
			switch pref {
			case MutChimera:
				baseCost = 260
			case MutBastion:
				baseCost = 240
			case MutFury:
				baseCost = 220
			case MutTitan:
				baseCost = 210
			case MutAether:
				baseCost = 200
			}

			cost := CalculateMutationCost(baseCost, m.Floor, target.Mutations.Total())
			if alchBudget >= cost && m.Gold >= cost {
				m.Gold -= cost
				alchBudget -= cost
				spentAlch += cost

				switch pref {
				case MutChimera:
					target.Mutations.ChimeraCount++
					target.MaxHP += 8
					target.HP += 8
					target.MaxMP += 5
					target.MP += 5
					target.BaseAtk += 1
					target.BaseDef += 1
					m.addLog(accentStyle.Render(fmt.Sprintf("⚗️ %s принял [Сыворотку Химеры] (+8 HP, +5 MP, +1 Atk, +1 Def) за %dG!", target.Name, cost)))
				case MutFury:
					target.Mutations.FuryCount++
					target.BaseAtk += 3
					target.MaxHP += 2
					target.HP += 2
					m.addLog(fireStyle.Render(fmt.Sprintf("⚗️ %s принял [Эссенцию Ярости] (+3 Atk, +2 HP) за %dG!", target.Name, cost)))
				case MutTitan:
					target.Mutations.TitanCount++
					target.MaxHP += 20
					target.HP += 20
					m.addLog(healStyle.Render(fmt.Sprintf("⚗️ %s принял [Кровь Титана] (+20 HP) за %dG!", target.Name, cost)))
				case MutAether:
					target.Mutations.AetherCount++
					target.MaxMP += 14
					target.MP += 14
					target.BaseAtk += 1
					m.addLog(fountStyle.Render(fmt.Sprintf("⚗️ %s принял [Флюид Эфира] (+14 MP, +1 Atk) за %dG!", target.Name, cost)))
				case MutBastion:
					target.Mutations.BastionCount++
					target.BaseDef += 2
					target.MaxHP += 6
					target.HP += 6
					m.addLog(healStyle.Render(fmt.Sprintf("⚗️ %s принял [Эликсир Бастиона] (+2 Def, +6 HP) за %dG!", target.Name, cost)))
				}
			} else {
				break
			}
		}

		for _, h := range m.Party {
			if h.IsDead || h.Potion != nil {
				continue
			}
			pType := PotionHP
			if h.Stress > 25 || h.Affliction != AfflictionNone {
				pType = PotionStress
			} else if h.Class == ClassMage || h.Class == ClassCleric {
				pType = PotionMP
			}

			cand := createPotion(pType, SizeSmall, m.Floor)
			if alchBudget >= cand.Cost && m.Gold >= cand.Cost {
				m.Gold -= cand.Cost
				alchBudget -= cand.Cost
				spentAlch += cand.Cost
				h.Potion = &cand
			}
		}

		if spentAlch > 0 {
			globalDebugReport.GoldSpentBreakdown["Алхимия и зелья"] += spentAlch
		}
		m.TownPhase = TownPhaseDepart

	case TownPhaseDepart:
		m.InTown = false
		m.PathHistory = []Point{}
		m.LoopDetectCount = 0
		m.initDungeonForFloor(m.Floor)
		m.addLog(accentStyle.Render("🛡️ Отряд экипирован и спускается на глубину!"))
	}
}

func resetGameStatic(m Model) (Model, tea.Cmd) {
	m.accumulateDebugReport()
	saveDebugReportToFile()

	savedGold := int(float64(m.Gold) * LegacyTaxRate)
	newLegacy := m.Legacy
	newLegacy.TreasuryGold = savedGold

	fresh := initialModelWithLegacy(newLegacy)
	fresh.TermWidth = m.TermWidth
	fresh.TermHeight = m.TermHeight

	return fresh, tea.Batch(tickCmd(fresh.SpeedMs), menuTickCmd())
}

func (m Model) resetGame() (Model, tea.Cmd) {
	return resetGameStatic(m)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermWidth = msg.Width
		m.TermHeight = msg.Height
		if len(m.Grid) == 0 {
			m.initDungeonForFloor(m.Floor)
		}

	case MenuTickMsg:
		if m.State == StateMenu {
			m.MenuCountdown--
			if m.MenuCountdown <= 0 {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
			return m, menuTickCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.SpeedMs = 280
		case "2":
			m.SpeedMs = 120
		case "f":
			if m.State == StatePlaying && m.Combat != nil {
				m.attemptFlee()
			}
		case "enter", " ":
			if m.State == StateMenu {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
			if m.State == StatePlaying {
				m.AutoMode = !m.AutoMode
				if m.AutoMode {
					delay := m.SpeedMs
					if m.InTown {
						delay = m.TownDelayMs
					}
					return m, tickCmd(delay)
				}
			}
		case "r":
			if m.State != StatePlaying && m.State != StateMenu {
				return m.resetGame()
			}
		case "s":
			if m.State == StatePlaying {
				m.State = StateStatsManual
				m.StatsScroll = 0
			} else if m.State == StateStatsManual {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "i":
			if m.State == StatePlaying {
				m.State = StateInfoBook
				m.StatsScroll = 0
			} else if m.State == StateInfoBook {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "e":
			if m.State == StatePlaying {
				m.State = StateArmory
				m.StatsScroll = 0
			} else if m.State == StateArmory {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "esc":
			if m.State == StateArmory || m.State == StateStatsManual || m.State == StateInfoBook {
				m.State = StatePlaying
				return m, tickCmd(m.SpeedMs)
			}
		case "up", "k":
			if m.State == StatePlaying {
				if m.LogScroll < len(m.Logs)-3 {
					m.LogScroll++
				}
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				if m.StatsScroll > 0 {
					m.StatsScroll--
				}
			}
		case "down", "j":
			if m.State == StatePlaying {
				if m.LogScroll > 0 {
					m.LogScroll--
				}
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				if m.StatsScroll < 500 {
					m.StatsScroll++
				}
			}
		case "pgup":
			if m.State == StatePlaying {
				m.LogScroll = min(len(m.Logs)-3, m.LogScroll+5)
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				m.StatsScroll = max(0, m.StatsScroll-6)
			}
		case "pgdown":
			if m.State == StatePlaying {
				m.LogScroll = max(0, m.LogScroll-5)
			} else if m.State == StateStatsManual || m.State == StateDefeat || m.State == StateInfoBook || m.State == StateArmory {
				m.StatsScroll += 6
			}
		case "+", "=":
			if m.SpeedMs > 60 {
				m.SpeedMs -= 30
			}
		case "-", "_":
			if m.SpeedMs < 800 {
				m.SpeedMs += 40
			}
		case "n":
			if m.State == StatePlaying && !m.AutoMode {
				m.step()
			}
		}

	case RestartTickMsg:
		if m.State == StateDefeat {
			m.RestartCountdown--
			if m.RestartCountdown <= 0 {
				return m.resetGame()
			}
			return m, restartTickCmd()
		}

	case TickMsg:
		if m.State == StatePlaying && m.AutoMode {
			m.step()
			if m.State == StateDefeat {
				return m, restartTickCmd()
			}
			delay := m.SpeedMs
			if m.InTown {
				delay = m.TownDelayMs
			}
			return m, tickCmd(delay)
		}
	}
	return m, nil
}

func (m Model) renderMenuScreen() string {
	var sb strings.Builder

	banner := townArtStyle.Render(`
       ___,-------,,-----------------------,,-------,___
  _,-""             ''---------------------''             ""-,_
,'        /\                                           /\        '.
/        /  \         ______________________          /  \         \
|       / /\ \        |                    |         / /\ \        |
|      | |  | |       |     D C C A G      |        |  | | |       |
|      | |__| |       |____________________|        |__| | |       |
|      |______|         _||          ||_            |______|       |
|       |    |         (_||          ||_)           |    |         |
|       |    |          | |          | |            |    |         |
|      _||||_          _|_|_        _|_|_          _||||_          |
`)

	sb.WriteString(banner + "\n")
	sb.WriteString(titleStyle.Render("       Dungeon Crawler Console Auto Game (dccag)\n\n"))

	sb.WriteString("Добро пожаловать в мрачный подземельный рогалик!\n")
	sb.WriteString("Ваша задача — провести отряд сквозь БЕСКОНЕЧНЫЕ этажи опаснейших биомов.\n\n")

	sb.WriteString(questStyle.Render("ОСОБЕННОСТИ ИГРЫ:\n"))
	sb.WriteString(" • Система уровней и опыта: отряд прокачивает базовые характеристики в боях.\n")
	sb.WriteString(" • 5 специализированных алхимических мутаций под конкретные классовые роли.\n")
	sb.WriteString(" • Тактическая модель агро: танки прикрывают магов и клириков щитом.\n")
	sb.WriteString(" • Квотирование казны: сбалансированное развитие города и защита от переплат.\n")
	sb.WriteString(" • Мягкое отступление: тактический маневр [F] без ударов в спину.\n\n")

	sb.WriteString(dangerStyle.Render(fmt.Sprintf("⏳ Автоматический старт экспедиции через: %d сек...\n\n", m.MenuCountdown)))
	sb.WriteString(healStyle.Render("[ENTER] или [SPACE] — Начать экспедицию немедленно\n"))
	sb.WriteString(subtleStyle.Render("[Q] — Выход из игры"))

	box := menuBoxStyle.Render(sb.String())
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderStatsScreen(title string, titleColor lipgloss.Color) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", header))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("НАСЛЕДИЕ КОРОЛЕВСТВА:\n"))
	sb.WriteString(fmt.Sprintf(" • В Казну следующего поколения передано: %s\n", goldStyle.Render(fmt.Sprintf("%dG", int(float64(m.Gold)*LegacyTaxRate)))))
	sb.WriteString(fmt.Sprintf(" • Уровень Кузницы: %d | Церкви: %d | Таверны: %d\n\n", m.Legacy.SmithyLevel, m.Legacy.ChurchLevel, m.Legacy.TavernLevel))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("ДОСТИЖЕНИЯ И КОНТРАКТЫ:\n"))
	sb.WriteString(fmt.Sprintf(" • Зачищено этажей: %d | Шагов: %d | Золота: %s\n", m.Stats.FloorsCleared, m.Stats.TotalSteps, goldStyle.Render(fmt.Sprintf("%dG", m.Stats.TotalGoldEarned))))
	sb.WriteString(fmt.Sprintf(" • Выполнено квестов: %d | Заточек: %d | Вскрыто ларей: %d\n\n", m.Stats.QuestsCompleted, m.Stats.UpgradesForged, m.Stats.ChestsOpened))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("ВЫЖИВШИЕ БОЙЦЫ:\n"))
	for _, h := range m.Party {
		if !h.IsDead {
			heroName := h.FullName()
			if h.Title != "" {
				heroName = titleStyle.Render(heroName)
			}
			sb.WriteString(fmt.Sprintf(" • %-20s (%-9s, Ур.%2d) [%s] (Atk:%2d | Def:%2d)\n",
				heroName, h.Class, h.Level, healStyle.Render("ВЫЖИЛ"), h.TotalAtk(), h.TotalDef()))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("КНИГА ПАМЯТИ (ПАВШИЕ: %d):\n", len(m.Stats.FallenHeroes))))
	if len(m.Stats.FallenHeroes) == 0 {
		sb.WriteString(subtleStyle.Render(" Ни один боец не погиб в этом походе.\n"))
	} else {
		for _, f := range m.Stats.FallenHeroes {
			sb.WriteString(fmt.Sprintf(" ☠️ %-18s (%-9s) | Этаж: %d | %s\n",
				dangerStyle.Render(f.FullName), f.Class, f.Floor, subtleStyle.Render(f.Cause)))
		}
	}

	if m.State == StateDefeat {
		countdownStr := dangerStyle.Render(fmt.Sprintf("\n⏳ Автоматический перезапуск через: %d сек...", m.RestartCountdown))
		sb.WriteString(countdownStr + "\n")
	}

	sb.WriteString(subtleStyle.Render("\n[↑/↓/PgUp/PgDn] Прокрутка | [R] Перезапуск | [S] Назад | [Q] Выход"))

	fullText := sb.String()
	lines := strings.Split(fullText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderInfoBookScreen() string {
	boxW := m.TermWidth - 8
	if boxW > 108 {
		boxW = 108
	}
	if boxW < 60 {
		boxW = 60
	}
	innerW := boxW - 6

	cSec := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	cSub := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	cMob := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	cBoss := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	cHp := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	cAtk := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	cDef := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	cSpd := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	cTier := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
	cItem := lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
	cNote := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	cVal := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("➔ ")

	fmtMob := func(name string, nameWidth int, hp, atk, def, spd int, growth, extra string) string {
		nameStr := cMob.Render(padRight(name, nameWidth))
		statsStr := fmt.Sprintf(
			"%s %s | %s %s | %s %s | %s %s",
			cHp.Render("HP"), cVal.Render(fmt.Sprintf("%-2d", hp)),
			cAtk.Render("ATK"), cVal.Render(fmt.Sprintf("%-2d", atk)),
			cDef.Render("DEF"), cVal.Render(fmt.Sprintf("%-1d", def)),
			cSpd.Render("СКОР"), cVal.Render(fmt.Sprintf("%-2d", spd)),
		)
		tail := cNote.Render(fmt.Sprintf("| Рост: %s", growth))
		if extra != "" {
			tail += " " + extra
		}
		return fmt.Sprintf("   • %s: %s %s\n", nameStr, statsStr, tail)
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render("═══ КОДЕКС ЗНАНИЙ И БАЗА ДАННЫХ DCCAG [I] ═══\n\n"))

	sb.WriteString(cSec.Render("1. ЦИКЛИЧЕСКИЕ БИОМЫ ПОДЗЕМЕЛЬЯ:") + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: Базовые монстры, сырость и гниль.\n", cSub.Render("Этажи 1, 6, 11... (Гнилые Катакомбы)")))
	sb.WriteString(fmt.Sprintf(" • %s: Слизни, утопленники, ящеры (-2 к скорости отряда).\n", cSub.Render("Этажи 2, 7, 12... (Затопленные Гроты)")))
	sb.WriteString(fmt.Sprintf(" • %s: Пепельные бесы, орки, саламандры (+урон огнем).\n", cSub.Render("Этажи 3, 8, 13... (Пепельные Недра)")))
	sb.WriteString(fmt.Sprintf(" • %s: Големы, гаргульи (+3 MP к стоимости навыков).\n", cSub.Render("Этажи 4, 9, 14... (Кристальный Лабиринт)")))
	sb.WriteString(fmt.Sprintf(" • %s: Демоны, Рыцари Смерти, Дракон (+10%% стресса).\n\n", cSub.Render("Этажи 5, 10, 15... (Трон Бездны)")))

	sb.WriteString(cSec.Render("2. МОНСТРЫ И ПРОГРЕССИЯ ХАРАКТЕРИСТИК:") + "\n")
	sb.WriteString(cSub.Render(" [Гнилые Катакомбы]") + "\n")
	sb.WriteString(fmtMob("Чумная крыса", 14, 14, 5, 0, 8, "+15% HP, +10% ATK", ""))
	sb.WriteString(fmtMob("Гоблин", 14, 17, 7, 1, 8, "+15% HP, +12% ATK", ""))
	sb.WriteString(fmtMob("Скелет", 14, 22, 8, 3, 8, "+18% HP, +15% ATK", subtleStyle.Render("(Блок 20%)")))
	sb.WriteString("\n")

	sb.WriteString(cSub.Render(" [Затопленные Гроты]") + "\n")
	sb.WriteString(fmtMob("Болотный слизень", 16, 26, 9, 1, 9, "+20% HP, +12% ATK", stressStyle.Render("(+10 Стр)")))
	sb.WriteString(fmtMob("Утопленник", 16, 34, 11, 2, 9, "+20% HP, +16% ATK", stressStyle.Render("(+10 Стр)")))
	sb.WriteString(fmtMob("Болотный ящер", 16, 30, 12, 3, 9, "+18% HP, +18% ATK", accentStyle.Render("(Криты)")))
	sb.WriteString("\n")

	sb.WriteString(cSub.Render(" [Пепельные Недра]") + "\n")
	sb.WriteString(fmtMob("Пепельный бес", 14, 36, 13, 2, 10, "+22% HP, +20% ATK", fireStyle.Render("(Опаление)")))
	sb.WriteString(fmtMob("Орк-берсерк", 14, 46, 15, 4, 10, "+25% HP, +22% ATK", fireStyle.Render("(Ярость)")))
	sb.WriteString(fmtMob("Саламандра", 14, 40, 16, 3, 10, "+22% HP, +25% ATK", fireStyle.Render("(Ярость)")))
	sb.WriteString("\n")

	sb.WriteString(cSub.Render(" [Кристальный Лабиринт]") + "\n")
	sb.WriteString(fmtMob("Гаргулья", 17, 52, 17, 6, 11, "+25% HP, +22% ATK", subtleStyle.Render("(Блок 20%)")))
	sb.WriteString(fmtMob("Кристальный голем", 17, 60, 18, 7, 11, "+30% HP, +20% ATK", subtleStyle.Render("(Блок 20%)")))
	sb.WriteString(fmtMob("Фантом", 17, 44, 19, 2, 11, "+20% HP, +28% ATK", stressStyle.Render("(+18 Стр)")))
	sb.WriteString("\n")

	sb.WriteString(cSub.Render(" [Трон Бездны & Владыки]") + "\n")
	sb.WriteString(fmtMob("Демон Бездны", 16, 66, 21, 5, 12, "+30% HP, +25% ATK", stressStyle.Render("(+18 Стр)")))
	sb.WriteString(fmtMob("Рыцарь Смерти", 16, 76, 22, 7, 12, "+32% HP, +28% ATK", dangerStyle.Render("(Вампиризм)")))
	sb.WriteString(fmt.Sprintf("   • %s: %s 260+(Lvl*20) | %s 25+Lvl | %s\n\n",
		cBoss.Render("Пепельный Дракон (БОСС)"), cHp.Render("HP"), cAtk.Render("ATK"), fireStyle.Render("Дыхание по всей группе")))

	sb.WriteString(cSec.Render("3. АФФИКСЫ И МОДИФИКАТОРЫ ВРАГОВ:") + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fireStyle.Render("🔥 Огненный"), cNote.Render("+3 к базовой атаке; опаляет бойца на +4 урона")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", stressStyle.Render("☣️ Ядовитый"), cNote.Render("Отравляет раны: -3 чистого здоровья и +14 стресса")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fountStyle.Render("❄️ Ледяной"), cNote.Render("Замедляет инициативу группы и сковывает действия")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", subtleStyle.Render("🪨 Каменный"), cNote.Render("+3 к защите (DEF), +12 к максимальному здоровью")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n\n", dangerStyle.Render("🩸 Вампир"), cNote.Render("Крадет здоровье: исцеляет себе 50% нанесенного урона")))

	sb.WriteString(cSec.Render("4. ТАБЛИЦА СНАРЯЖЕНИЯ (МАТЕРИАЛЫ И БАЗОВЫЕ ТИПЫ):") + "\n")
	sb.WriteString(cNote.Render("   (Материалы: Железн. x1 | Стальн. x2 | Мифрил. x3 | Адамант. x4. Заточка: +2/ур)") + "\n\n")

	sb.WriteString(cSub.Render(" [ТАНК - Бастионное снаряжение]") + "\n")
	sb.WriteString(fmt.Sprintf("   • Оружие: %s Гладиус (%s:3) %s%s Палаш %s%s Моргенштерн %s%s Бастионный меч (%s:9, Блок:8)\n",
		cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
	sb.WriteString(fmt.Sprintf("   • Доспех: %s Бригантина (%s:4) %s%s Полудоспех %s%s Кираса бастиона %s%s Панцирь цитадели (%s:13, %s:+40)\n",
		cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))
	sb.WriteString(fmt.Sprintf("   • Шлем:   %s Топфхельм %s%s Салад %s%s Армет %s%s Бацинет бастиона (%s:8, Блок:4)\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Поножи: %s Наголенники %s%s Шарнирные поножи %s%s Латные поножи %s%s Протекторы цитадели (%s:8, %s:+20)\n\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))

	sb.WriteString(cSub.Render(" [ВОИН - Оружие прорыва и латы]") + "\n")
	sb.WriteString(fmt.Sprintf("   • Оружие: %s Эспадон (%s:5) %s%s Клеймор %s%s Боевой топор %s%s Фальшион (%s:14, Крит:4)\n",
		cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
	sb.WriteString(fmt.Sprintf("   • Доспех: %s Хауберк (%s:3) %s%s Кираса ярости %s%s Чешуйчатый доспех %s%s Нагрудник витязя (%s:9, %s:+32)\n",
		cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))
	sb.WriteString(fmt.Sprintf("   • Шлем:   %s Норманнский шлем %s%s Бацинет %s%s Барбют %s%s Шишак (%s:8, Крит:1)\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Поножи: %s Чешуйчатые гетры %s%s Чулки %s%s Пластины %s%s Поножи витязя (%s:8)\n\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))

	sb.WriteString(cSub.Render(" [РАЗБОЙНИК - Клинки скрытности и легкая кожа]") + "\n")
	sb.WriteString(fmt.Sprintf("   • Оружие: %s Охотничьи ножи (%s:4) %s%s Парные стилеты %s%s Кинжалы %s%s Воровские кортики (%s:10, Крит:8)\n",
		cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
	sb.WriteString(fmt.Sprintf("   • Доспех: %s Колет (%s:2) %s%s Гамбезон %s%s Куртка теневика %s%s Плащ ассасина (%s:8, Крит:4)\n",
		cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Шлем:   %s Тканевая маска %s%s Капюшон %s%s Бандана %s%s Маска теней (%s:7, Крит:1)\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Поножи: %s Краги %s%s Плотные гетры %s%s Мягкие сапоги %s%s Поножи бесшумности (%s:7)\n\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))

	sb.WriteString(cSub.Render(" [МАГ - Эфирные проводники и мантии]") + "\n")
	sb.WriteString(fmt.Sprintf("   • Оружие: %s Рунная трость (%s:6) %s%s Посох искр %s%s Жезл %s%s Архимагический скипетр (%s:15, MP:+40)\n",
		cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
	sb.WriteString(fmt.Sprintf("   • Доспех: %s Роба ученика (%s:2) %s%s Мантия чародея %s%s Одеяние эфира %s%s Астральная мантия (%s:8, MP:+45)\n",
		cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Шлем:   %s Остроконечная шляпа %s%s Обруч магии %s%s Диадема фокуса %s%s Капюшон магистра (%s:7, MP:+26)\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Поножи: %s Обмотки %s%s Шёлковые поножи %s%s Ленты левитации %s%s Штаны чародея (%s:7, MP:+18)\n\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))

	sb.WriteString(cSub.Render(" [КЛИРИК - Освященное оружие и облачения]") + "\n")
	sb.WriteString(fmt.Sprintf("   • Оружие: %s Окованная дубина (%s:4) %s%s Боевой молот %s%s Шестопёр %s%s Булава света (%s:10, MP:+26)\n",
		cTier.Render("Т1"), cAtk.Render("Atk"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cAtk.Render("Atk")))
	sb.WriteString(fmt.Sprintf("   • Доспех: %s Сутана (%s:3) %s%s Кираса инквизитора %s%s Пресвитерский панцирь %s%s Священный доспех (%s:9, Рез:25%%)\n",
		cTier.Render("Т1"), cDef.Render("Def"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Шлем:   %s Митра %s%s Койф %s%s Капеллина %s%s Венец правосудия (%s:8, Рез:20%%)\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def")))
	sb.WriteString(fmt.Sprintf("   • Поножи: %s Наголенники веры %s%s Сапоги паломника %s%s Инквизиторские сапоги %s%s Наколенники света (%s:8, %s:+24)\n\n",
		cTier.Render("Т1"), cArrow, cTier.Render("Т2"), cArrow, cTier.Render("Т3"), cArrow, cTier.Render("Т4"), cDef.Render("Def"), cHp.Render("HP")))

	sb.WriteString(cSec.Render("5. ПРОГРЕССИЯ УРОВНЕЙ И КЛАССОВЫЙ ОПЫТ (XP):") + "\n")
	sb.WriteString(" • Опыт от убитых врагов делится поровну между всеми живыми героями.\n")
	sb.WriteString(" • Прирост Танка: +12 HP, +2 MP, +1 Atk, +1 Def каждые 2 ур.\n")
	sb.WriteString(" • Прирост Воина: +8 HP, +3 MP, +2 Atk, +1 Def каждые 3 ур.\n")
	sb.WriteString(" • Прирост Разбойника: +5 HP, +4 MP, +2 Atk.\n")
	sb.WriteString(" • Прирост Мага: +4 HP, +8 MP, +3 Atk.\n")
	sb.WriteString(" • Прирост Клирика: +6 HP, +6 MP, +1 Atk.\n\n")

	sb.WriteString(cSec.Render("6. АЛХИМИЧЕСКИЕ МУТАЦИИ И ПРИОРИТЕТЫ:") + "\n")
	sb.WriteString(fmt.Sprintf(" • %s (База 260G): +8 HP, +5 MP, +1 Atk, +1 Def (Универсально)\n", accentStyle.Render("Сыворотка Химеры")))
	sb.WriteString(fmt.Sprintf(" • %s (База 220G): +3 Atk, +2 HP (Приоритет: Разбойник, Воин, Маг)\n", fireStyle.Render("Эссенция Ярости")))
	sb.WriteString(fmt.Sprintf(" • %s (База 210G): +20 MaxHP (Приоритет: Танк, Воин)\n", healStyle.Render("Кровь Титана")))
	sb.WriteString(fmt.Sprintf(" • %s (База 200G): +14 MaxMP, +1 Atk (Приоритет: Маг, Клирик)\n", fountStyle.Render("Флюид Эфира")))
	sb.WriteString(fmt.Sprintf(" • %s (База 240G): +2 Def, +6 HP (Приоритет: Танк, Клирик)\n\n", healStyle.Render("Эликсир Бастиона")))

	sb.WriteString(cSec.Render("7. РЕЛИКВИИ И АРТЕФАКТЫ ПОДЗЕМЕЛЬЯ:") + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: Увеличивает добычу золота до +45%%, но враги бьют больнее.\n", cItem.Render("Компас Алчности")))
	sb.WriteString(fmt.Sprintf(" • %s: При гибели героя все выжившие получают +3 Atk за каждый уровень реликвии.\n", cItem.Render("Корона Мученика")))
	sb.WriteString(fmt.Sprintf(" • %s: Пассивно снижает весь получаемый отрядом стресс вплоть до 60%%.\n\n", cItem.Render("Священный Грааль")))

	sb.WriteString(cNote.Render("[↑/↓/PgUp/PgDn] Прокрутка  |  [I / Esc / S] Закрыть кодекс"))

	wrappedText := lipgloss.NewStyle().Width(innerW).Render(sb.String())
	lines := strings.Split(wrappedText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Width(boxW).Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderArmoryScreen() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("═══ АРСЕНАЛ ОТРЯДА И АЛХИМИЧЕСКИЕ МУТАЦИИ [E] ═══")
	sb.WriteString(header + "\n\n")

	renderSlotInfo := func(slotName string, it *EquipItem) string {
		if it == nil {
			return fmt.Sprintf("   • %-7s: %s\n", slotName, subtleStyle.Render("Пусто"))
		}
		matStr := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("[%s, x%d]", it.Material.Name, it.Material.BonusMult))

		modStr := ""
		if it.Prefix != nil {
			modStr += fmt.Sprintf(" | %s (+%d %s)", it.Prefix.Name, it.Prefix.Bonus, it.Prefix.Element)
		}
		if it.Suffix != nil {
			modStr += fmt.Sprintf(" | %s (+%d %s)", it.Suffix.Name, it.Suffix.Bonus, it.Suffix.Effect)
		}

		statLabel := "Защ"
		if it.Slot == SlotWeapon {
			statLabel = "Атк"
		}

		upg := ""
		if it.UpgradeLevel > 0 {
			upg = fmt.Sprintf(" +%d", it.UpgradeLevel)
		}

		return fmt.Sprintf("   • %-7s: %s%s %s -> Итого: %s %d%s\n",
			slotName, goldStyle.Render(it.BaseName), upg, matStr, statLabel, it.TotalStat(), subtleStyle.Render(modStr))
	}

	for _, h := range m.Party {
		status := healStyle.Render("[В СТРОЮ]")
		if h.IsDead {
			status = dangerStyle.Render("[МЕРТВ]")
		}

		heroTitle := h.FullName()
		if h.Title != "" {
			heroTitle = titleStyle.Render(heroTitle)
		}

		mutStr := subtleStyle.Render("Нет")
		if h.Mutations.Total() > 0 {
			mutStr = accentStyle.Render(fmt.Sprintf("Всего: %d [Хим:%d Яр:%d Тит:%d Эфир:%d Баст:%d]",
				h.Mutations.Total(), h.Mutations.ChimeraCount, h.Mutations.FuryCount, h.Mutations.TitanCount, h.Mutations.AetherCount, h.Mutations.BastionCount))
		}

		potInfo := "Нет"
		if h.Potion != nil {
			potInfo = fmt.Sprintf("%s %s (Сила: %d)", h.Potion.Symbol, getPotionName(*h.Potion), h.Potion.Power)
		}

		sb.WriteString(fmt.Sprintf("👤 %s (%s, Ур.%d, XP:%d/%d, Агро: %d) %s\n",
			heroTitle, h.Class, h.Level, h.Exp, h.NextLevelExp(), h.Role.AggroWeight, status))
		sb.WriteString(fmt.Sprintf("   🧪 Мутации: %s | В поясе: %s\n", mutStr, potionStyle.Render(potInfo)))
		sb.WriteString(renderSlotInfo("Оружие", h.Weapon))
		sb.WriteString(renderSlotInfo("Шлем", h.Head))
		sb.WriteString(renderSlotInfo("Доспех", h.Chest))
		sb.WriteString(renderSlotInfo("Поножи", h.Legs))
		sb.WriteString("\n")
	}

	sb.WriteString(subtleStyle.Render("[↑/↓/PgUp/PgDn] Прокрутка  |  [E/Esc] Закрыть арсенал  |  [Space] Пауза  |  [Q] Выход"))

	fullText := sb.String()
	lines := strings.Split(fullText, "\n")

	maxVisibleLines := max(10, m.TermHeight-8)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]
	renderedContent := strings.Join(visibleLines, "\n")

	box := statsBoxStyle.Render(renderedContent)
	w := max(100, m.TermWidth)
	h := max(30, m.TermHeight)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderTownHub(viewW, viewH int) string {
	bMarket := "[ ⚖️ РЫНОК ]"
	bGuild := "[ ⚔️ ГИЛЬДИЯ ]"
	bAlchemist := "[ 🧪 АЛХИМИК ]"
	bSmithy := "[ ⚒️ КУЗНИЦА ]"
	bInvest := "[ 🏛️ МАГИСТРАТ ]"
	bChurch := "[ ⛪ ЦЕРКОВЬ ]"
	bTavern := "[ 🍻 ТАВЕРНА ]"

	switch m.TownPhase {
	case TownPhaseSellLoot, TownPhaseUpgradeBag:
		bMarket = accentStyle.Render("►[ ⚖️ РЫНОК ]◄")
	case TownPhaseGuild:
		bGuild = accentStyle.Render("►[ ⚔️ ГИЛЬДИЯ ]◄")
	case TownPhaseAlchemist:
		bAlchemist = accentStyle.Render("►[ 🧪 АЛХИМИК ]◄")
	case TownPhaseSmithy:
		bSmithy = accentStyle.Render("►[ ⚒️ КУЗНИЦА ]◄")
	case TownPhaseMagistrate:
		bInvest = accentStyle.Render("►[ 🏛️ МАГИСТРАТ ]◄")
	case TownPhaseChurch:
		bChurch = accentStyle.Render("►[ ⛪ ЦЕРКОВЬ ]◄")
	case TownPhaseTavern:
		bTavern = accentStyle.Render("►[ 🍻 ТАВЕРНА ]◄")
	}

	maxLineLen := max(20, viewW-2)
	centerIn := func(s string) string {
		visibleLen := lipgloss.Width(s)
		if visibleLen >= maxLineLen {
			return s
		}
		padding := (maxLineLen - visibleLen) / 2
		return strings.Repeat(" ", padding) + s
	}

	rawLines := []string{
		centerIn(townArtStyle.Render("═══ СТОЛИЧНЫЙ КВАРТАЛ (ПРИВАЛ) ═══")),
		"",
		centerIn(fmt.Sprintf("%s   %s   %s", bMarket, bGuild, bAlchemist)),
		centerIn("════════════════════════════════════"),
		centerIn(fmt.Sprintf("%s   %s   %s   %s", bSmithy, bInvest, bChurch, bTavern)),
		"",
		subtleStyle.Render(strings.Repeat("─", maxLineLen)),
		questStyle.Render("ДЕЙСТВИЕ: ") + shortenItemName(m.TownDialog, maxLineLen-10),
		subtleStyle.Render(strings.Repeat("─", maxLineLen)),
		subtleStyle.Render(shortenItemName("Квотирование казны: мутации, экипировка, отдых.", maxLineLen)),
	}

	var res []string
	padTop := max(0, (viewH-len(rawLines))/2)
	for i := 0; i < padTop; i++ {
		res = append(res, "")
	}
	res = append(res, rawLines...)
	for len(res) < viewH {
		res = append(res, "")
	}
	if len(res) > viewH {
		res = res[:viewH]
	}
	return strings.Join(res, "\n")
}

func (m Model) renderMap(viewW, viewH int) string {
	if m.InTown {
		return m.renderTownHub(viewW, viewH)
	}

	biome := getBiome(m.Floor)
	customWall := lipgloss.NewStyle().Foreground(biome.WallColor).Render("#")
	customFloor := lipgloss.NewStyle().Foreground(biome.FloorColor).Render(string(biome.FloorRune))

	startX := m.PartyPos.X - viewW/2
	startY := m.PartyPos.Y - viewH/2

	if startX+viewW > m.MapWidth {
		startX = m.MapWidth - viewW
	}
	if startY+viewH > m.MapHeight {
		startY = m.MapHeight - viewH
	}
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	var sb strings.Builder
	for y := 0; y < viewH; y++ {
		mapY := startY + y
		for x := 0; x < viewW; x++ {
			mapX := startX + x
			if mapX >= m.MapWidth || mapY >= m.MapHeight {
				sb.WriteString(" ")
				continue
			}

			if mapX == m.PartyPos.X && mapY == m.PartyPos.Y {
				sb.WriteString(partyStyle.Render("@"))
				continue
			}
			if !m.Explored[mapY][mapX] {
				sb.WriteString(" ")
				continue
			}
			pos := Point{mapX, mapY}
			if pack, exists := m.Packs[pos]; exists {
				first := pack.GetFirstLiving()
				if first != nil {
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(first.Color)).Bold(true).Render(string(first.Glyph)))
				} else {
					sb.WriteString(customFloor)
				}
				continue
			}
			switch m.Grid[mapY][mapX] {
			case TileWall:
				sb.WriteString(customWall)
			case TileFloor:
				sb.WriteString(customFloor)
			case TileChest:
				sb.WriteString(goldStyle.Render("$"))
			case TileRelic:
				sb.WriteString(relicTileStyle.Render("*"))
			case TileAltar:
				sb.WriteString(altarStyle.Render("_"))
			case TileFountain:
				sb.WriteString(fountStyle.Render("~"))
			case TileTrappedChest:
				sb.WriteString(trappedChestStyle.Render("T"))
			case TileBarrel:
				sb.WriteString(barrelStyle.Render("o"))
			case TileStairs:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render(">"))
			case TileExit:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render("<"))
			default:
				sb.WriteRune(rune(m.Grid[mapY][mapX]))
			}
		}
		if y < viewH-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) renderHeroCard(h *Hero, cardWidth int) string {
	isActiveTurn := false
	if m.Combat != nil && len(m.Combat.TurnQueue) > 0 {
		idx := m.Combat.TurnIdx
		if idx >= len(m.Combat.TurnQueue) {
			idx = 0
		}
		if m.Combat.TurnQueue[idx].Type == CombatantHero && m.Combat.TurnQueue[idx].HeroRef == h {
			isActiveTurn = true
		}
	}

	var sb strings.Builder
	innerWidth := max(18, cardWidth-4)

	nameRaw := fmt.Sprintf("%s L%d", h.FullName(), h.Level)
	nameRunes := []rune(nameRaw)
	if len(nameRunes) > innerWidth {
		nameRaw = string(nameRunes[:innerWidth-1]) + "…"
	}
	nameStr := nameRaw
	if h.Title != "" {
		nameStr = titleStyle.Render(nameRaw)
	}

	clsStr := string(h.Class)
	if innerWidth < 22 && len([]rune(clsStr)) > 3 {
		clsStr = string([]rune(clsStr)[:3])
	}
	classFormatted := subtleStyle.Render("(" + clsStr + ")")

	potSlot := subtleStyle.Render("[·]")
	if h.Potion != nil {
		potSlot = fmt.Sprintf("[%s]", h.Potion.Symbol)
	}

	buffSlot := subtleStyle.Render("[-]")
	switch {
	case h.Affliction != AfflictionNone:
		buffSlot = stressStyle.Render("[👁]")
	case h.IsGuarding:
		buffSlot = healStyle.Render("[🛡]")
	case h.IsBerserk:
		buffSlot = fireStyle.Render("[⚔ ]")
	case h.IsStealthed:
		buffSlot = accentStyle.Render("[🗡]")
	case h.IsCharged:
		buffSlot = accentStyle.Render("[🔮]")
	case h.IsAura:
		buffSlot = fountStyle.Render("[✨]")
	}

	turnSlot := subtleStyle.Render("[ ]")
	if isActiveTurn {
		turnSlot = accentStyle.Render("[⚡]")
	}

	if h.IsDead {
		sb.WriteString(fmt.Sprintf("%s\n", nameStr))
		sb.WriteString(fmt.Sprintf("%s %s\n", classFormatted, dangerStyle.Render("[☠️ ПАЛ]")))
		sb.WriteString(subtleStyle.Render(shortenItemName(h.CauseOfDeath, innerWidth)) + "\n")
		sb.WriteString(fmt.Sprintf("Этаж: %d\n", m.Floor))
		sb.WriteString(subtleStyle.Render("[Ждет воскрешения]"))
	} else {
		sb.WriteString(fmt.Sprintf("%s\n", nameStr))
		sb.WriteString(fmt.Sprintf("%s %s %s %s\n", classFormatted, potSlot, buffSlot, turnSlot))

		barLen := max(2, (innerWidth-18)/2)
		hpBar := renderBar(h.HP, h.MaxHP, barLen, lipgloss.Color("82"), lipgloss.Color("238"))
		mpBar := renderBar(h.MP, h.MaxMP, barLen, lipgloss.Color("39"), lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("HP:%s%2d MP:%s%2d\n", hpBar, h.HP, mpBar, h.MP))

		stressColor := lipgloss.Color("135")
		if h.Stress >= 140 {
			stressColor = lipgloss.Color("196")
		}
		stressBar := renderBar(h.Stress, 200, barLen, stressColor, lipgloss.Color("238"))
		sb.WriteString(fmt.Sprintf("СТ:%s%3d ⚔%-2d 🛡%-2d\n", stressBar, h.Stress, h.TotalAtk(), h.TotalDef()))

		slotLen := max(4, (innerWidth-6)/2)
		formatEquipSlot := func(item *EquipItem, maxLen int) string {
			if item == nil {
				return "-"
			}
			name := shortenItemName(item.BaseName, maxLen)
			if item.UpgradeLevel > 0 {
				return fmt.Sprintf("+%d%s", item.UpgradeLevel, shortenItemName(item.BaseName, maxLen-2))
			}
			return name
		}

		wName := padRight(formatEquipSlot(h.Weapon, slotLen), slotLen)
		hName := padRight(formatEquipSlot(h.Head, slotLen), slotLen)
		chName := padRight(formatEquipSlot(h.Chest, slotLen), slotLen)
		lName := padRight(formatEquipSlot(h.Legs, slotLen), slotLen)

		equipBlock := fmt.Sprintf("⚔%s 🪖%s\n🛡%s 🥾%s", wName, hName, chName, lName)
		sb.WriteString(subtleStyle.Render(equipBlock))
	}

	style := heroCardStyle.Width(innerWidth)
	if h.IsDead {
		style = heroCardDead.Width(innerWidth)
	} else if isActiveTurn {
		style = heroCardActive.Width(innerWidth)
	}
	return style.Render(sb.String())
}

func (m Model) renderPartyBanner(cardWidth int) string {
	innerWidth := max(18, cardWidth-4)
	var sb strings.Builder
	sb.WriteString(goldStyle.Render("👑 ОТРЯД И РЕЛИКВИЯ") + "\n")
	relicName := "Нет"
	relicDesc := "Реликвия не найдена"
	if m.Relic != nil {
		relicName = m.Relic.Name
		relicDesc = m.Relic.Description
	}
	sb.WriteString(fmt.Sprintf("Реликвия: %s\n", shortenItemName(relicName, innerWidth-10)))
	sb.WriteString(subtleStyle.Render(shortenItemName(relicDesc, innerWidth)) + "\n")
	sb.WriteString(fmt.Sprintf("🎒 Мешки: %d/%d слотов\n", len(m.Bag), m.currentBagCapacity()))
	legacyPart := int(float64(m.Gold) * LegacyTaxRate)
	sb.WriteString(healStyle.Render(fmt.Sprintf("🏛️ Наследие: +%dG", legacyPart)))

	return lipgloss.NewStyle().
		Width(innerWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Render(sb.String())
}

func (m Model) renderRightContentLines(innerRightW, maxLines int) []string {
	var lines []string
	if m.Combat != nil {
		barrelNotice := ""
		if m.Combat.HasBarrel {
			barrelNotice = " [🛢️ ПОРОХ]"
		}
		if m.Combat.Round > 20 {
			barrelNotice += dangerStyle.Render(fmt.Sprintf(" [ЯРОСТЬ R%d]", m.Combat.Round))
		}
		lines = append(lines, dangerStyle.Render(shortenItemName("ВРАЖЕСКАЯ СТАЯ"+barrelNotice+":", innerRightW-2)))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))

		availRows := max(1, maxLines-len(lines))
		displayed := 0
		for _, mob := range m.Combat.Pack.Members {
			if displayed >= availRows {
				break
			}

			statusWidth := 7
			stText := fmt.Sprintf("%d/%d", mob.HP, mob.MaxHP)
			stColor := lipgloss.Color("245")
			if mob.IsDead {
				stText = "УБИТ"
				stColor = lipgloss.Color("239")
			}
			stBadge := lipgloss.NewStyle().Foreground(stColor).Render(padRight(fmt.Sprintf("[%s]", stText), statusWidth))

			barLen := 3
			if innerRightW > 36 {
				barLen = 4
			}
			mobBar := renderBar(mob.HP, mob.MaxHP, barLen, lipgloss.Color("196"), lipgloss.Color("238"))

			affIcon := getAffixIcon(mob.Affix)
			if affIcon != "" {
				affIcon += " "
			}

			prefix := fmt.Sprintf("• %s", affIcon)
			fixedW := lipgloss.Width(prefix) + 1 + lipgloss.Width(mobBar) + 1 + lipgloss.Width(stBadge)
			availName := max(3, innerRightW-fixedW)
			mobName := shortenItemName(mob.Name, availName)
			namePadded := padRight(mobName, availName)

			lines = append(lines, fmt.Sprintf("%s%s %s %s", prefix, namePadded, mobBar, stBadge))
			displayed++
		}

		if len(m.Combat.Pack.Members) > displayed && len(lines) < maxLines {
			lines = append(lines, subtleStyle.Render(fmt.Sprintf("... и ещё %d", len(m.Combat.Pack.Members)-displayed)))
		}
	} else if m.InTown {
		lines = append(lines, townArtStyle.Render("СТОЛИЧНОЕ УПРАВЛЕНИЕ:"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("Кузница Ур.%-2d  | Церковь Ур.%-2d", m.Legacy.SmithyLevel, m.Legacy.ChurchLevel))
		lines = append(lines, fmt.Sprintf("Таверна Ур.%-2d  | Магистрат", m.Legacy.TavernLevel))
		lines = append(lines, "")
		lines = append(lines, subtleStyle.Render(shortenItemName("Квотирование казны активно.", innerRightW-2)))
	} else {
		lines = append(lines, accentStyle.Render("РАЗВЕДКА ТЕРРИТОРИИ:"))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, fmt.Sprintf("Сделано шагов:   %d", m.Stats.TotalSteps))
		lines = append(lines, fmt.Sprintf("Сундуков найдено: %d", m.Stats.ChestsOpened))
	}
	return lines
}

func (m Model) renderRightContent(innerRightW, infoInnerH int) string {
	lines := m.renderRightContentLines(innerRightW, infoInnerH)
	for len(lines) < infoInnerH {
		lines = append(lines, "")
	}
	if len(lines) > infoInnerH {
		lines = lines[:infoInnerH]
	}
	return strings.Join(lines, "\n")
}

func (m Model) View() string {
	if m.State == StateMenu {
		return m.renderMenuScreen()
	}
	if m.State == StateDefeat {
		return m.renderStatsScreen("ПОРАЖЕНИЕ: ВЕСЬ ОТРЯД ПАЛ ВО ТЬМЕ...", lipgloss.Color("196"))
	}
	if m.State == StateStatsManual {
		return m.renderStatsScreen("ЭКИПИРОВКА, НАСЛЕДИЕ И КНИГА ПАМЯТИ", lipgloss.Color("214"))
	}
	if m.State == StateInfoBook {
		return m.renderInfoBookScreen()
	}
	if m.State == StateArmory {
		return m.renderArmoryScreen()
	}

	termW := max(70, m.TermWidth)
	termH := max(20, m.TermHeight)
	usableW := termW - 2

	cardsPerRow := 3
	if usableW >= 150 || (termH < 34 && usableW >= 115) {
		cardsPerRow = 6
	} else if usableW < 90 {
		cardsPerRow = 2
	}

	singleCardWidth := (usableW - (cardsPerRow-1)*1) / cardsPerRow
	numRows := (6 + cardsPerRow - 1) / cardsPerRow

	var cards []string
	for _, h := range m.Party {
		cards = append(cards, m.renderHeroCard(h, singleCardWidth))
	}
	cards = append(cards, m.renderPartyBanner(singleCardWidth))

	var cardRowElements []string
	for i := 0; i < len(cards); i += cardsPerRow {
		end := min(len(cards), i+cardsPerRow)
		cardRowElements = append(cardRowElements, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	middleTier := lipgloss.JoinVertical(lipgloss.Left, cardRowElements...)

	cardsHeight := numRows * 8
	maxLogs := 5
	if termH < 32 {
		maxLogs = 3
	}
	if termH < 26 {
		maxLogs = 2
	}
	fixedBottomHeight := maxLogs + 4

	availableTopH := max(6, termH-cardsHeight-fixedBottomHeight)

	sideW := int(float64(usableW) * 0.34)
	if sideW < 28 {
		sideW = 28
	}
	if sideW > 46 {
		sideW = 46
	}
	mapBoxW := max(32, usableW-sideW-1)

	viewW := max(10, mapBoxW-4)
	viewH := max(4, availableTopH-2)

	mapStr := m.renderMap(viewW, viewH)
	leftMapBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(mapBoxW - 2).
		Height(viewH).
		Render(mapStr)

	innerRightW := max(22, sideW-4)
	biome := getBiome(m.Floor)
	floorTag := fmt.Sprintf("ЭТАЖ %d: %s", m.Floor, biome.Name)
	if m.InTown {
		floorTag = fmt.Sprintf("ЭТАЖ %d: [Лагерь]", m.Floor)
	}

	questTitleShort := shortenItemName(m.CurrentQuest.Title, innerRightW-12)
	var statusBadge string
	if m.CurrentQuest.Completed {
		statusBadge = healStyle.Render("[СДАТЬ]")
	} else {
		statusBadge = subtleStyle.Render(fmt.Sprintf("[%d/%d]", m.CurrentQuest.Current, m.CurrentQuest.TargetCount))
	}
	questLine := padRight(fmt.Sprintf("📜 %s %s", questTitleShort, statusBadge), innerRightW-4)

	var rightPane string
	if availableTopH >= 13 {
		floorBoxContent := fmt.Sprintf(
			"🏰 %s\n%s",
			titleStyle.Render(shortenItemName(floorTag, innerRightW-4)),
			subtleStyle.Render(shortenItemName("Опасность: "+biome.EnvHazard, innerRightW-4)),
		)
		floorBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Width(innerRightW).
			Height(2).
			Render(floorBoxContent)

		statusContent := fmt.Sprintf(
			"💰 Казна: %s\n%s",
			goldStyle.Render(fmt.Sprintf("%dG", m.Gold)),
			questLine,
		)
		statusBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Width(innerRightW).
			Height(2).
			Render(statusContent)

		infoInnerH := max(3, availableTopH-10)
		rightInfoBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(innerRightW).
			Height(infoInnerH).
			Render(m.renderRightContent(innerRightW, infoInnerH))

		rightPane = lipgloss.JoinVertical(lipgloss.Left, rightInfoBox, statusBox, floorBox)
	} else {
		compactH := max(4, availableTopH-2)
		var lines []string
		lines = append(lines, fmt.Sprintf("🏰 %s | 💰 %s", titleStyle.Render(shortenItemName(floorTag, 14)), goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
		lines = append(lines, fmt.Sprintf("📜 %s", questLine))
		lines = append(lines, subtleStyle.Render(strings.Repeat("─", innerRightW-2)))
		lines = append(lines, m.renderRightContentLines(innerRightW, compactH-3)...)

		rightPane = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(innerRightW).
			Height(compactH).
			Render(strings.Join(lines, "\n"))
	}

	topTier := lipgloss.JoinHorizontal(lipgloss.Top, leftMapBox, " ", rightPane)

	var logs strings.Builder
	scrollInfo := ""
	if m.LogScroll > 0 {
		scrollInfo = fmt.Sprintf(" (Архив: -%d)", m.LogScroll)
	}
	logs.WriteString(lipgloss.NewStyle().Bold(true).Render("ХРОНИКИ ЭКСПЕДИЦИИ" + scrollInfo + ":\n"))

	totalLogs := len(m.Logs)
	endIdx := totalLogs - m.LogScroll
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	if endIdx < maxLogs {
		endIdx = min(totalLogs, maxLogs)
	}
	startIdx := max(0, endIdx-maxLogs)

	for i := 0; i < maxLogs; i++ {
		curIdx := startIdx + i
		if curIdx < endIdx && curIdx < totalLogs {
			logs.WriteString(fmt.Sprintf("> %s\n", shortenItemName(m.Logs[curIdx], termW-8)))
		} else {
			logs.WriteString("\n")
		}
	}

	controls := subtleStyle.Render("[Space] Пауза  |  [↑/↓] Логи  |  [F] Побег  |  [+/-] Скор.  |  [1/2] Темп  |  [E] Арсенал  |  [I] Кодекс  |  [S] Слава  |  [Q] Выход")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topTier,
		middleTier,
		strings.TrimRight(logs.String(), "\n"),
		controls,
	)
}

func main() {
	flag.Parse()

	m := initialModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Ошибка запуска: %v\n", err)
		os.Exit(1)
	}

	if *debugReportFlag {
		if fm, ok := finalModel.(*Model); ok {
			fm.accumulateDebugReport()
		} else if fmVal, ok := finalModel.(Model); ok {
			fmVal.accumulateDebugReport()
		}
		saveDebugReportToFile()
	}
}
