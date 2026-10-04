package main

import (
	"strings"
	"unicode"
)

// ============================================================
// ЧЕРТЁЖ ГЕРОЯ-ЛИДЕРА
// ============================================================
//
// HeroBlueprint — то, что игрок задаёт на экране создания: имя, пол, раса, класс,
// призвание и распределение очков характеристик. Чертёж хранится в сохранении,
// а в каждой новой экспедиции из него заново собирается герой первого уровня.

// LeaderCalling — призвание лидера: небольшой постоянный бонус всему отряду,
// пока лидер на ногах.
type LeaderCalling string

const (
	CallingInspirer   LeaderCalling = "inspirer"   // союзники меньше нервничают
	CallingStrategist LeaderCalling = "strategist" // больше золота с добычи
	CallingMentor     LeaderCalling = "mentor"     // больше опыта всему отряду
)

var AllCallings = []LeaderCalling{CallingInspirer, CallingStrategist, CallingMentor}

const (
	callingInspirerStressCut = 15 // % снижения стресса у союзников лидера
	callingStrategistGoldPct = 10 // % прибавки к золоту
	callingMentorExpPct      = 10 // % прибавки к опыту
)

// Распределяемые характеристики.
const (
	StatHP = iota
	StatMP
	StatAtk
	StatDef
	StatSpd
	statCount
)

const (
	creatorPoints  = 6  // всего свободных очков
	creatorNameMin = 2  // длина имени, в символах
	creatorNameMax = 16 //
	heroBaseAtk    = 6  // базовая атака героя 1-го уровня (без кузницы и оружия)
)

type statSpec struct {
	perPoint  int // сколько единиц характеристики даёт одно очко
	maxPoints int // сколько очков можно вложить в одну характеристику
	labelKey  string
}

var statSpecs = [statCount]statSpec{
	StatHP:  {perPoint: 5, maxPoints: 4, labelKey: "creator.stat.hp"},
	StatMP:  {perPoint: 5, maxPoints: 4, labelKey: "creator.stat.mp"},
	StatAtk: {perPoint: 1, maxPoints: 3, labelKey: "creator.stat.atk"},
	StatDef: {perPoint: 1, maxPoints: 3, labelKey: "creator.stat.def"},
	StatSpd: {perPoint: 1, maxPoints: 3, labelKey: "creator.stat.spd"},
}

type HeroBlueprint struct {
	Name    string         `json:"name"`
	Gender  Gender         `json:"gender"`
	Race    RaceType       `json:"race"`
	Class   HeroClass      `json:"class"`
	Calling LeaderCalling  `json:"calling"`
	Points  [statCount]int `json:"points"`
}

func defaultBlueprint() HeroBlueprint {
	return HeroBlueprint{
		Gender:  GenderMale,
		Race:    RaceHuman,
		Class:   ClassWarrior,
		Calling: CallingInspirer,
	}
}

// allowedNameRune — в имени допускаются буквы, цифры, пробел, дефис и апостроф.
// Остальное (управляющие символы, «%», кавычки) отсекается, чтобы имя безопасно
// попадало в журнал и карточки.
func allowedNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '\''
}

// cleanName убирает недопустимые символы, схлопывает пробелы и обрезает имя до creatorNameMax.
func cleanName(s string) string {
	var out []rune
	prevSpace := true // срезает ведущие пробелы
	for _, r := range s {
		if !allowedNameRune(r) {
			continue
		}
		if r == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
		} else {
			prevSpace = false
		}
		out = append(out, r)
		if len(out) >= creatorNameMax {
			break
		}
	}
	return strings.TrimRight(string(out), " ")
}

func validRace(r RaceType) bool {
	for _, v := range AllRaces {
		if v == r {
			return true
		}
	}
	return false
}

func validClass(c HeroClass) bool {
	for _, v := range AllClasses {
		if v == c {
			return true
		}
	}
	return false
}

func validCalling(c LeaderCalling) bool {
	for _, v := range AllCallings {
		if v == c {
			return true
		}
	}
	return false
}

// Sanitized приводит чертёж к допустимому виду. Нужен как для ввода игрока,
// так и для файла сохранения, который мог быть испорчен или отредактирован вручную.
func (bp HeroBlueprint) Sanitized() HeroBlueprint {
	bp.Name = cleanName(bp.Name)
	if bp.Gender != GenderFemale {
		bp.Gender = GenderMale
	}
	if !validRace(bp.Race) {
		bp.Race = RaceHuman
	}
	if !validClass(bp.Class) {
		bp.Class = ClassWarrior
	}
	if !validCalling(bp.Calling) {
		bp.Calling = CallingInspirer
	}
	left := creatorPoints
	for i := 0; i < statCount; i++ {
		p := clampInt(bp.Points[i], 0, statSpecs[i].maxPoints)
		if p > left {
			p = left
		}
		bp.Points[i] = p
		left -= p
	}
	return bp
}

// Valid — чертёж пригоден для создания героя (имя достаточной длины).
func (bp HeroBlueprint) Valid() bool {
	return len([]rune(cleanName(bp.Name))) >= creatorNameMin
}

func (bp HeroBlueprint) PointsUsed() int {
	n := 0
	for _, p := range bp.Points {
		n += p
	}
	return n
}

func (bp HeroBlueprint) PointsLeft() int {
	return creatorPoints - bp.PointsUsed()
}

// AdjustPoint вкладывает (dir > 0) или забирает (dir < 0) одно очко.
// Возвращает false, если нельзя: нет свободных очков или достигнут предел.
func (bp *HeroBlueprint) AdjustPoint(stat, dir int) bool {
	if stat < 0 || stat >= statCount || dir == 0 {
		return false
	}
	step := 1
	if dir < 0 {
		step = -1
	}
	v := bp.Points[stat] + step
	if v < 0 || v > statSpecs[stat].maxPoints {
		return false
	}
	if step > 0 && bp.PointsLeft() <= 0 {
		return false
	}
	bp.Points[stat] = v
	return true
}

// HeroPreview — итоговые характеристики героя 1-го уровня без снаряжения.
type HeroPreview struct {
	HP, MP, Atk, Def, Spd int
	SkillKey              string
	SkillCost             int
	Role                  CombatRole
}

// Preview считает характеристики по той же формуле, по которой собирается герой
// (см. buildHero и createLeaderHero), не трогая генератор случайных чисел.
func (bp HeroBlueprint) Preview(smithyLvl int) HeroPreview {
	base := baseStatsFor(bp.Class)
	raceMod := GetRaceModifiers(bp.Race)

	maxHP := base.MaxHP
	if raceMod.MaxHPBonusPercent != 0 {
		maxHP += int(float64(maxHP) * raceMod.MaxHPBonusPercent)
	}
	pt := func(stat int) int { return bp.Points[stat] * statSpecs[stat].perPoint }

	return HeroPreview{
		HP:        maxHP + pt(StatHP),
		MP:        base.MaxMP + pt(StatMP),
		Atk:       heroBaseAtk + smithyLvl + pt(StatAtk),
		Def:       base.BaseDef + pt(StatDef),
		Spd:       base.Speed + raceMod.SpeedFlat + pt(StatSpd),
		SkillKey:  base.SkillNameKey,
		SkillCost: base.SkillCost,
		Role:      GetClassRole(bp.Class),
	}
}

// createLeaderHero собирает героя-лидера из чертежа.
func createLeaderHero(bp HeroBlueprint, smithyLvl int) *Hero {
	bp = bp.Sanitized()
	h := buildHero(bp.Class, bp.Race, "", bp.Gender, 1, smithyLvl)
	h.CustomName = bp.Name
	h.IsLeader = true
	h.Calling = bp.Calling

	h.MaxHP += bp.Points[StatHP] * statSpecs[StatHP].perPoint
	h.MaxMP += bp.Points[StatMP] * statSpecs[StatMP].perPoint
	h.HP = h.MaxHP
	h.MP = h.MaxMP
	h.BaseAtk += bp.Points[StatAtk] * statSpecs[StatAtk].perPoint
	h.BaseDef += bp.Points[StatDef] * statSpecs[StatDef].perPoint
	h.Speed += bp.Points[StatSpd] * statSpecs[StatSpd].perPoint
	return h
}

// buildStartingParty собирает стартовый отряд из пяти героев разных классов.
// С чертежом первым идёт лидер, а его класс исключается из случайных спутников.
func buildStartingParty(smithyLvl int, bp *HeroBlueprint) []*Hero {
	if bp == nil || !bp.Valid() {
		// Перемешиваем все 10 классов и отбираем ровно 5 уникальных
		shuffled := make([]HeroClass, len(AllClasses))
		copy(shuffled, AllClasses)
		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		var heroes []*Hero
		for _, c := range shuffled[:5] {
			heroes = append(heroes, createHero(c, 1, smithyLvl))
		}
		return heroes
	}

	var pool []HeroClass
	for _, c := range AllClasses {
		if c != bp.Class {
			pool = append(pool, c)
		}
	}
	rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	heroes := []*Hero{createLeaderHero(*bp, smithyLvl)}
	for _, c := range pool[:4] {
		heroes = append(heroes, createHero(c, 1, smithyLvl))
	}
	return heroes
}

// randomBlueprint — случайный, но полностью допустимый герой (кнопка «Случайно»).
func randomBlueprint(lang Language) HeroBlueprint {
	gender := GenderMale
	if rng.Intn(2) == 1 {
		gender = GenderFemale
	}
	var names []HeroNameDef
	for _, n := range HeroNames {
		if n.Gender == gender {
			names = append(names, n)
		}
	}

	bp := HeroBlueprint{
		Gender:  gender,
		Race:    AllRaces[rng.Intn(len(AllRaces))],
		Class:   AllClasses[rng.Intn(len(AllClasses))],
		Calling: AllCallings[rng.Intn(len(AllCallings))],
	}
	if len(names) > 0 {
		bp.Name = cleanName(T(lang, names[rng.Intn(len(names))].NameKey))
	}
	for left := creatorPoints; left > 0; {
		if bp.AdjustPoint(rng.Intn(statCount), 1) {
			left--
		}
	}
	return bp.Sanitized()
}
