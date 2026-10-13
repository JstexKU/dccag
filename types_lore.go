package main

// LoreCategory — 6 тематических разделов глобальной Библиотеки Столицы
type LoreCategory int

const (
	LoreCatMastery      LoreCategory = iota // 1. Фолианты Мастерства и Родовые Свитки (30 классовых + 7 расовых)
	LoreCatGenesis                          // 2. Хроники Бытия (история мира, катаклизм, падение)
	LoreCatEthnography                      // 3. Этнография и Мифы Народов (заметки учёного и фольклор)
	LoreCatBellesLettres                    // 4. Беллетристика, Лирика и Юмор (снятие стресса)
	LoreCatDiaries                          // 5. Экспедиционные Дневники Биомов (11 глав о выживании)
	LoreCatBestiary                         // 6. Бестиарий и Трактаты об Аффиксах (уязвимости чудовищ)
)

// LoreBonusType — типы эффектов, активируемых при изучении фолианта
type LoreBonusType string

const (
	LoreBonusNone       LoreBonusType = "none"
	LoreBonusClassPerk  LoreBonusType = "class_perk"  // Пассивное усиление конкретного класса
	LoreBonusRacePerk   LoreBonusType = "race_perk"   // Пассивное усиление конкретной расы
	LoreBonusPartyExp   LoreBonusType = "party_exp"   // Мгновенный опыт отряду
	LoreBonusStressHeal LoreBonusType = "stress_heal" // Мгновенное снятие стресса
	LoreBonusBaneDmg    LoreBonusType = "bane_dmg"    // Доп. урон/крит по конкретному семейству монстров
)

// LoreBonus описывает игровую выгоду от нахождения фолианта
type LoreBonus struct {
	Type        LoreBonusType
	TargetClass HeroClass    // Заполняется, если Type == LoreBonusClassPerk
	TargetRace  RaceType     // Заполняется, если Type == LoreBonusRacePerk
	TargetMob   MonsterType  // Заполняется для бестиария (если урон по конкретному мобу)
	TargetAffix MonsterAffix // Заполняется для трактатов об аффиксах
	Value       int          // Значение бонуса (XP, стресс, урон, % блока)
	DescKey     string       // Ключ локализации описания эффекта
}

// LoreVolume представляет один конкретный фолиант, свиток или запись
type LoreVolume struct {
	ID          string       // Уникальный ID (например: "mastery_tank_1", "genesis_vol_1")
	Category    LoreCategory // Раздел библиотеки
	VolumeNum   int          // Номер тома в серии (1, 2, 3...)
	TitleKey    string       // Ключ локализации названия
	AuthorKey   string       // Ключ локализации автора
	ExcerptKey  string       // Ключ краткой цитаты для лога при нахождении
	ContentKey  string       // Ключ полного текста для экрана библиотеки
	BiomeOrigin BiomeType    // Биом преимущественного спавна (или "" для любого)
	Bonus       LoreBonus    // Игровой бонус
}

// LibraryState управляет отображением и навигацией в экране Библиотеки [B]
type LibraryState struct {
	ActiveCategory LoreCategory // Выбранный раздел (0..5)
	Cursor         int          // Индекс выбранного тома в списке раздела
	ScrollOffset   int          // Смещение списка томов (viewport)
	TextScrollY    int          // Прокрутка текста внутри самого окна чтения справа
	ViewingDetails bool         // Фокус переключен на правое окно для скролла длинного текста
}

// AdjustScroll удерживает курсор внутри видимого скользящего окна списка томов
func (s *LibraryState) AdjustScroll(totalItems, pageSize int) {
	if totalItems <= pageSize {
		s.ScrollOffset = 0
		return
	}
	if s.Cursor < s.ScrollOffset {
		s.ScrollOffset = s.Cursor
	} else if s.Cursor >= s.ScrollOffset+pageSize {
		s.ScrollOffset = s.Cursor - pageSize + 1
	}
	maxOffset := totalItems - pageSize
	if s.ScrollOffset > maxOffset {
		s.ScrollOffset = maxOffset
	}
	if s.ScrollOffset < 0 {
		s.ScrollOffset = 0
	}
}

// GlobalLoreRegister содержит статическую базу всех доступных в игре томов
var AllLoreVolumes = []LoreVolume{}

// UnlockedLoreStore — структура для сериализации в save.json и хранения в TownLegacy
type UnlockedLoreStore map[string]bool

func NewUnlockedLoreStore() UnlockedLoreStore {
	return make(UnlockedLoreStore)
}

func (s UnlockedLoreStore) IsUnlocked(id string) bool {
	return s[id]
}

func (s UnlockedLoreStore) Unlock(id string) bool {
	if s[id] {
		return false
	}
	s[id] = true
	return true
}

func (s UnlockedLoreStore) TotalUnlocked() int {
	return len(s)
}