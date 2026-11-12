package main

import (
	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// ТАБЛИЦА БИОМОВ
// ============================================================
//
// Всё, что зависит от биома, описано здесь один раз: внешний вид, мобы обычных стай,
// мини-босс. getBiome, spawnMonsterPack, генератор контрактов и дневники в библиотеке
// читают эту таблицу. Новый биом — одна запись в biomeTable (плюс переводы).
//
// Порядок мобов в записи влияет на воспроизводимость забегов: spawnMonsterPack выбирает
// моба как Mobs[rng.Intn(len(Mobs))].

// MobTemplate — базовые характеристики моба обычной стаи (до масштабирования по этажу).
type MobTemplate struct {
	Type  MonsterType
	Glyph rune
	Color string
	Atk   int
	Def   int
	HP    int
}

// MiniBossTemplate — мини-босс биома: Atk и Def — базовые значения до масштабирования.
type MiniBossTemplate struct {
	Type    MonsterType
	NameKey string
	Glyph   rune
	Color   string
	Affix   MonsterAffix
	Atk     int
	Def     int
}

// BiomeConfig — внешний вид биома.
type BiomeConfig struct {
	Name         BiomeType
	WallColor    lipgloss.Color
	FloorColor   lipgloss.Color
	FloorRune    rune
	EnvHazardKey string
}

// BiomeDef — полное описание биома.
type BiomeDef struct {
	BiomeConfig
	Mobs     []MobTemplate
	MiniBoss MiniBossTemplate
}

// MobTypes — типы мобов биома (пул для контрактов на охоту).
func (d BiomeDef) MobTypes() []MonsterType {
	types := make([]MonsterType, len(d.Mobs))
	for i, m := range d.Mobs {
		types[i] = m.Type
	}
	return types
}

// Мини-боссы; одного мини-босса делят несколько биомов.
var (
	miniExecutioner = MiniBossTemplate{MobMiniExecutioner, "mob.mini_executioner", 'E', "196", AffixPoison, 18, 6}
	miniReaver      = MiniBossTemplate{MobMiniReaver, "mob.mini_reaver", 'R', "208", AffixFire, 22, 4}
	miniColossus    = MiniBossTemplate{MobMiniColossus, "mob.mini_colossus", 'C', "51", AffixStone, 17, 9}
	miniBoneblight  = MiniBossTemplate{MobMiniBoneblight, "mob.mini_boneblight", 'W', "118", AffixPoison, 20, 7}
	miniStalker     = MiniBossTemplate{MobMiniStalker, "mob.mini_stalker", 'S', "93", AffixVampiric, 24, 5}
)

// biomeTable — биомы по порядку цикла этажей. Длина таблицы и есть длина цикла.
var biomeTable = [...]BiomeDef{
	{
		BiomeConfig: BiomeConfig{Name: BiomeCatacombs, WallColor: "240", FloorColor: "236", FloorRune: '·', EnvHazardKey: "hazard.catacombs"},
		Mobs: []MobTemplate{
			{MobRat, 'r', "137", 6, 0, 15},
			{MobGoblin, 'g', "118", 8, 1, 19},
			{MobSkeleton, 's', "252", 10, 3, 25},
		},
		MiniBoss: miniExecutioner,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeGrotto, WallColor: "31", FloorColor: "24", FloorRune: '~', EnvHazardKey: "hazard.grotto"},
		Mobs: []MobTemplate{
			{MobSlime, 'c', "43", 11, 1, 30},
			{MobDrowned, 'u', "31", 13, 2, 38},
			{MobLizard, 'l', "29", 14, 3, 34},
		},
		MiniBoss: miniExecutioner,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeInferno, WallColor: "124", FloorColor: "52", FloorRune: '≈', EnvHazardKey: "hazard.inferno"},
		Mobs: []MobTemplate{
			{MobImp, 'i', "208", 15, 2, 40},
			{MobOrc, 'o', "130", 17, 4, 52},
			{MobSalamander, 'm', "196", 18, 3, 46},
		},
		MiniBoss: miniReaver,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeCrystal, WallColor: "141", FloorColor: "54", FloorRune: '◊', EnvHazardKey: "hazard.crystal"},
		Mobs: []MobTemplate{
			{MobGargoyle, 'G', "102", 19, 6, 58},
			{MobGolem, 'C', "141", 20, 7, 68},
			{MobPhantom, 'p', "159", 22, 2, 50},
		},
		MiniBoss: miniColossus,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeDeadwood, WallColor: "28", FloorColor: "235", FloorRune: '♣', EnvHazardKey: "hazard.deadwood"},
		Mobs: []MobTemplate{
			{MobSproutSkeleton, 's', "106", 12, 3, 28},
			{MobDryad, 'd', "83", 16, 1, 38},
			{MobBlightEnt, 'T', "94", 18, 6, 56},
		},
		MiniBoss: miniBoneblight,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeFungal, WallColor: "100", FloorColor: "58", FloorRune: '§', EnvHazardKey: "hazard.fungal"},
		Mobs: []MobTemplate{
			{MobSporling, 'x', "142", 13, 2, 32},
			{MobTentacle, 't', "65", 15, 3, 42},
			{MobToxicBasil, 'b', "35", 19, 4, 48},
		},
		MiniBoss: miniBoneblight,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeArchives, WallColor: "178", FloorColor: "94", FloorRune: '≡', EnvHazardKey: "hazard.archives"},
		Mobs: []MobTemplate{
			{MobTomeBook, 'b', "221", 16, 1, 35},
			{MobScrollMimic, 'm', "178", 18, 4, 44},
			{MobArchiveKeeper, 'A', "75", 21, 5, 54},
		},
		MiniBoss: miniColossus,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeMines, WallColor: "238", FloorColor: "233", FloorRune: '•', EnvHazardKey: "hazard.mines"},
		Mobs: []MobTemplate{
			{MobObsidianBeetle, 'o', "238", 17, 8, 45},
			{MobMinerGhoul, 'z', "243", 20, 3, 48},
			{MobDeepTroll, 'O', "130", 23, 6, 65},
		},
		MiniBoss: miniReaver,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeSanctuary, WallColor: "161", FloorColor: "53", FloorRune: '†', EnvHazardKey: "hazard.sanctuary"},
		Mobs: []MobTemplate{
			{MobBloodCultist, 'c', "161", 19, 2, 42},
			{MobShadowInquisitor, 'i', "125", 22, 5, 52},
			{MobFallenCrusader, 'F', "196", 24, 7, 68},
		},
		MiniBoss: miniStalker,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeAstral, WallColor: "69", FloorColor: "17", FloorRune: '¤', EnvHazardKey: "hazard.astral"},
		Mobs: []MobTemplate{
			{MobAstralWeaver, 'w', "69", 20, 3, 46},
			{MobChronoPhantom, 'p', "183", 22, 2, 50},
			{MobEssenceDevour, 'D', "135", 25, 4, 60},
		},
		MiniBoss: miniStalker,
	},
	{
		BiomeConfig: BiomeConfig{Name: BiomeAbyss, WallColor: "89", FloorColor: "232", FloorRune: '×', EnvHazardKey: "hazard.abyss"},
		Mobs: []MobTemplate{
			{MobVoidDemon, 'V', "161", 25, 5, 78},
			{MobDeathKnight, 'K', "89", 27, 8, 88},
		},
		MiniBoss: miniStalker,
	},
}

// biomeCount — длина цикла биомов; после последнего биома этажи идут по кругу.
const biomeCount = len(biomeTable)

// biomeIndex — номер биома этажа в цикле (0..biomeCount-1).
func biomeIndex(floor int) int {
	return (floor - 1) % biomeCount
}

// getBiomeDef — полное описание биома этажа.
func getBiomeDef(floor int) *BiomeDef {
	return &biomeTable[biomeIndex(floor)]
}

// getBiome — внешний вид биома этажа.
func getBiome(floor int) BiomeConfig {
	return getBiomeDef(floor).BiomeConfig
}
