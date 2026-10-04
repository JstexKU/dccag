package main

import (
	"strings"
	"testing"
)

func TestCleanName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Эдан", "Эдан"},
		{"  Анна   Мария ", "Анна Мария"},
		{"Ann%s<b>", "Anns" + "b"},
		{"O'Neil-Smith", "O'Neil-Smith"},
		{"Рыцарь\n\tБез\x00Страха", "РыцарьБезСтраха"},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ", "ABCDEFGHIJKLMNOP"},
		{"   ", ""},
	}
	for _, c := range cases {
		if got := cleanName(c.in); got != c.want {
			t.Errorf("cleanName(%q) = %q, ожидалось %q", c.in, got, c.want)
		}
	}
}

func TestBlueprintValidName(t *testing.T) {
	bp := defaultBlueprint()
	for name, want := range map[string]bool{"": false, "A": false, " A ": false, "Al": true, "Алёша": true} {
		bp.Name = name
		if bp.Valid() != want {
			t.Errorf("имя %q: Valid() = %v, ожидалось %v", name, bp.Valid(), want)
		}
	}
}

func TestBlueprintSanitizedFixesGarbage(t *testing.T) {
	bp := HeroBlueprint{
		Name:    "Bob%d<>",
		Gender:  Gender(7),
		Race:    RaceType("dragon"),
		Class:   HeroClass("hacker"),
		Calling: LeaderCalling("king"),
		Points:  [statCount]int{99, -5, 99, 99, 99},
	}.Sanitized()

	if bp.Name != "Bobd" {
		t.Errorf("имя не очищено: %q", bp.Name)
	}
	if bp.Gender != GenderMale || bp.Race != RaceHuman || bp.Class != ClassWarrior || bp.Calling != CallingInspirer {
		t.Errorf("недопустимые значения не сброшены: %+v", bp)
	}
	if bp.PointsUsed() > creatorPoints {
		t.Errorf("потрачено %d очков при лимите %d", bp.PointsUsed(), creatorPoints)
	}
	for i, p := range bp.Points {
		if p < 0 || p > statSpecs[i].maxPoints {
			t.Errorf("характеристика %d вне диапазона: %d", i, p)
		}
	}
}

func TestAdjustPointRespectsLimits(t *testing.T) {
	bp := defaultBlueprint()

	if bp.AdjustPoint(StatAtk, -1) {
		t.Error("нельзя забрать очко, которого нет")
	}
	for i := 0; i < statSpecs[StatAtk].maxPoints; i++ {
		if !bp.AdjustPoint(StatAtk, +1) {
			t.Fatalf("очко %d должно вкладываться", i+1)
		}
	}
	if bp.AdjustPoint(StatAtk, +1) {
		t.Error("нельзя превысить предел характеристики")
	}

	// Вкладываем всё остальное и упираемся в общий лимит.
	for _, stat := range []int{StatHP, StatMP, StatDef, StatSpd} {
		for bp.AdjustPoint(stat, +1) {
		}
	}
	if bp.PointsUsed() != creatorPoints || bp.PointsLeft() != 0 {
		t.Errorf("использовано %d из %d очков", bp.PointsUsed(), creatorPoints)
	}
	if bp.AdjustPoint(StatHP, +1) && bp.PointsUsed() > creatorPoints {
		t.Error("общий лимит очков превышен")
	}
	if !bp.AdjustPoint(StatHP, -1) || bp.PointsLeft() != 1 {
		t.Error("очко должно возвращаться")
	}
	if bp.AdjustPoint(99, +1) || bp.AdjustPoint(-1, +1) || bp.AdjustPoint(StatHP, 0) {
		t.Error("неверные аргументы должны отклоняться")
	}
}

// Предпросмотр на экране создания обязан совпадать с реально созданным героем.
func TestPreviewMatchesCreatedHero(t *testing.T) {
	seedRNG(77)
	for _, class := range AllClasses {
		for _, race := range AllRaces {
			for _, smithy := range []int{0, 3} {
				bp := HeroBlueprint{Name: "Тест", Gender: GenderFemale, Race: race, Class: class, Calling: CallingMentor}
				bp.AdjustPoint(StatHP, 1)
				bp.AdjustPoint(StatHP, 1)
				bp.AdjustPoint(StatMP, 1)
				bp.AdjustPoint(StatAtk, 1)
				bp.AdjustPoint(StatDef, 1)
				bp.AdjustPoint(StatSpd, 1)

				p := bp.Preview(smithy)
				h := createLeaderHero(bp, smithy)
				name := string(class) + "/" + string(race)

				if p.HP != h.MaxHP || h.HP != h.MaxHP {
					t.Errorf("%s: HP %d != %d", name, p.HP, h.MaxHP)
				}
				if p.MP != h.MaxMP || h.MP != h.MaxMP {
					t.Errorf("%s: MP %d != %d", name, p.MP, h.MaxMP)
				}
				if p.Atk != h.BaseAtk {
					t.Errorf("%s: атака %d != %d", name, p.Atk, h.BaseAtk)
				}
				if p.Def != h.BaseDef {
					t.Errorf("%s: защита %d != %d", name, p.Def, h.BaseDef)
				}
				if p.Spd != h.Speed+GetRaceModifiers(race).SpeedFlat {
					t.Errorf("%s: скорость %d != %d", name, p.Spd, h.Speed+GetRaceModifiers(race).SpeedFlat)
				}
				if p.SkillKey != h.SkillNameKey || p.SkillCost != h.SkillCost {
					t.Errorf("%s: навык не совпал", name)
				}
			}
		}
	}
}

func TestBaseStatsForKeepsClassBalance(t *testing.T) {
	// Значения взяты из исходной таблицы createHero: рефакторинг не должен менять баланс.
	want := map[HeroClass][4]int{ // MaxHP, MaxMP, BaseDef, Speed
		ClassTank:    {60, 30, 4, 8},
		ClassPaladin: {54, 35, 3, 9},
		ClassWarrior: {48, 25, 2, 10},
		ClassMonk:    {44, 30, 1, 13},
		ClassRogue:   {35, 35, 1, 15},
		ClassRanger:  {38, 30, 1, 13},
		ClassMage:    {28, 45, 0, 11},
		ClassWarlock: {34, 40, 1, 10},
		ClassCleric:  {34, 40, 2, 9},
		ClassBard:    {36, 40, 1, 12},
	}
	for class, w := range want {
		b := baseStatsFor(class)
		got := [4]int{b.MaxHP, b.MaxMP, b.BaseDef, b.Speed}
		if got != w {
			t.Errorf("%s: %v, ожидалось %v", class, got, w)
		}
	}
}

func TestCreateLeaderHero(t *testing.T) {
	seedRNG(5)
	bp := HeroBlueprint{Name: "Эдан", Gender: GenderFemale, Race: RaceElf, Class: ClassMage, Calling: CallingStrategist}
	h := createLeaderHero(bp, 0)

	if !h.IsLeader || h.Calling != CallingStrategist {
		t.Error("герой должен быть лидером с выбранным призванием")
	}
	if h.DisplayName(LangRU) != "Эдан" || h.DisplayName(LangEN) != "Эдан" {
		t.Error("пользовательское имя не должно переводиться")
	}
	if h.Gender != GenderFemale || h.Race != RaceElf || h.Class != ClassMage || h.Level != 1 {
		t.Errorf("параметры героя не совпали: %+v", h)
	}
	if h.Weapon == nil || h.Head == nil || h.Chest == nil || h.Legs == nil {
		t.Error("лидер должен стартовать со снаряжением класса")
	}
	if h.IsDead || h.IsDowned || h.HP != h.MaxHP {
		t.Error("лидер должен стартовать здоровым")
	}
}

func TestBuildStartingParty(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		seedRNG(seed)
		bp := HeroBlueprint{Name: "Лидер", Gender: GenderMale, Race: RaceHuman, Class: ClassCleric, Calling: CallingMentor}
		party := buildStartingParty(0, &bp)
		if len(party) != 5 {
			t.Fatalf("seed %d: в отряде %d героев", seed, len(party))
		}
		if !party[0].IsLeader || party[0].Class != ClassCleric {
			t.Fatalf("seed %d: первым должен идти лидер", seed)
		}
		seen := map[HeroClass]bool{}
		leaders := 0
		for _, h := range party {
			if seen[h.Class] {
				t.Fatalf("seed %d: повтор класса %s", seed, h.Class)
			}
			seen[h.Class] = true
			if h.IsLeader {
				leaders++
			}
		}
		if leaders != 1 {
			t.Fatalf("seed %d: лидеров %d", seed, leaders)
		}
	}
}

func TestBuildStartingPartyWithoutLeader(t *testing.T) {
	seedRNG(3)
	for _, bp := range []*HeroBlueprint{nil, {Name: "A"}} {
		party := buildStartingParty(0, bp)
		if len(party) != 5 {
			t.Fatalf("в отряде %d героев", len(party))
		}
		seen := map[HeroClass]bool{}
		for _, h := range party {
			if h.IsLeader {
				t.Error("без валидного чертежа лидера быть не должно")
			}
			if seen[h.Class] {
				t.Errorf("повтор класса %s", h.Class)
			}
			seen[h.Class] = true
		}
	}
}

func TestRandomBlueprintIsValid(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		seedRNG(seed)
		for _, lang := range []Language{LangRU, LangEN} {
			bp := randomBlueprint(lang)
			if !bp.Valid() {
				t.Fatalf("seed %d: случайный герой невалиден: %+v", seed, bp)
			}
			if bp.PointsLeft() != 0 {
				t.Fatalf("seed %d: случайный герой не потратил очки (%d свободно)", seed, bp.PointsLeft())
			}
			if bp != bp.Sanitized() {
				t.Fatalf("seed %d: случайный чертёж не прошёл очистку: %+v", seed, bp)
			}
		}
	}
}

func TestDictionaryCoversCreatorOptions(t *testing.T) {
	for _, lang := range []Language{LangRU, LangEN} {
		for _, race := range AllRaces {
			for _, key := range []string{"race." + string(race) + ".name", "creator.race." + string(race)} {
				if got := T(lang, key); got == key {
					t.Errorf("%s: нет перевода %q", lang, key)
				}
			}
		}
		for _, class := range AllClasses {
			if got := T(lang, "class."+string(class)+".name"); strings.HasPrefix(got, "class.") {
				t.Errorf("%s: нет названия класса %s", lang, class)
			}
			if got := T(lang, baseStatsFor(class).SkillNameKey); strings.HasPrefix(got, "skill.") {
				t.Errorf("%s: нет названия навыка класса %s", lang, class)
			}
		}
		for _, c := range AllCallings {
			for _, suffix := range []string{".name", ".desc"} {
				key := "calling." + string(c) + suffix
				if got := T(lang, key); got == key {
					t.Errorf("%s: нет перевода %q", lang, key)
				}
			}
		}
	}
}
