package main

import (
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// Эти тесты не дают вернуть в код захардкоженный пользовательский текст:
// всё, что видит игрок, должно лежать в словарях и доставаться через T / TVerb.

func hasCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

func firstCyrillicLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if hasCyrillic(line) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

type srcLiteral struct {
	File  string
	Line  int
	Value string
}

// productionSources — файлы игры без тестов и без самих словарей.
func productionSources(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "i18n_ru.go" || f == "i18n_en.go" {
			continue
		}
		out = append(out, f)
	}
	return out
}

// stringLiteralsOf возвращает строковые литералы файла (комментарии не учитываются).
func stringLiteralsOf(t *testing.T, path string) []srcLiteral {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file := fset.AddFile(path, fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(file, src, nil, 0)

	var out []srcLiteral
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.STRING {
			continue
		}
		v, err := strconv.Unquote(lit)
		if err != nil {
			v = lit
		}
		out = append(out, srcLiteral{path, fset.Position(pos).Line, v})
	}
	return out
}

func TestNoHardcodedCyrillicInSource(t *testing.T) {
	for _, f := range productionSources(t) {
		for _, l := range stringLiteralsOf(t, f) {
			if hasCyrillic(l.Value) {
				t.Errorf("%s:%d: захардкоженный русский текст %q — вынесите его в словари", l.File, l.Line, l.Value)
			}
		}
	}
}

var langBranch = regexp.MustCompile(`\b[lL]ang\s*[=!]=\s*Lang(EN|RU)\b`)

// Переводить «веткой по языку» нельзя: тексты берутся через T. Исключения — переключатель
// языка, проверка сохранения и сама функция перевода.
func TestNoInlineLanguageBranchesInSource(t *testing.T) {
	allowed := map[string]bool{"i18n.go": true, "save.go": true, "creator.go": true, "update.go": true}
	for _, f := range productionSources(t) {
		if allowed[f] {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if langBranch.MatchString(line) {
				t.Errorf("%s:%d: ветка по языку вместо ключа словаря: %s", f, i+1, strings.TrimSpace(line))
			}
		}
	}
}

var tverbCall = regexp.MustCompile(`TVerb\([^,]+,[^,]+,\s*"([\w]+)"\)`)

func TestTVerbKeysExistInBothDictionaries(t *testing.T) {
	calls := 0
	for _, f := range productionSources(t) {
		if f == "i18n.go" {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range tverbCall.FindAllStringSubmatch(string(src), -1) {
			calls++
			for _, suffix := range []string{"m", "f"} {
				key := "verb." + m[1] + "." + suffix
				if _, ok := dictRU[key]; !ok {
					t.Errorf("%s: нет ключа %q в русском словаре", f, key)
				}
				if _, ok := dictEN[key]; !ok {
					t.Errorf("%s: нет ключа %q в английском словаре", f, key)
				}
			}
		}
	}
	if calls == 0 {
		t.Fatal("в коде не найдено ни одного вызова TVerb — тест перестал что-либо проверять")
	}
}

func TestTVerbPicksGenderAndLanguage(t *testing.T) {
	cases := []struct {
		lang   Language
		gender Gender
		want   string
	}{
		{LangRU, GenderMale, "достиг"},
		{LangRU, GenderFemale, "достигла"},
		{LangEN, GenderMale, "reached"},
		{LangEN, GenderFemale, "reached"},
	}
	for _, c := range cases {
		if got := TVerb(c.lang, c.gender, "reach"); got != c.want {
			t.Errorf("TVerb(%s, %v, reach) = %q, ожидалось %q", c.lang, c.gender, got, c.want)
		}
	}
}

func TestEveryVerbHasBothGendersInBothLanguages(t *testing.T) {
	for key := range dictRU {
		if !strings.HasPrefix(key, "verb.") {
			continue
		}
		base := strings.TrimSuffix(strings.TrimSuffix(key, ".m"), ".f")
		for _, suffix := range []string{".m", ".f"} {
			if _, ok := dictRU[base+suffix]; !ok {
				t.Errorf("у глагола %s нет варианта %s в русском словаре", base, suffix)
			}
			if _, ok := dictEN[base+suffix]; !ok {
				t.Errorf("у глагола %s нет варианта %s в английском словаре", base, suffix)
			}
		}
	}
}

func TestSpendCategoriesAreTranslatable(t *testing.T) {
	ids := []string{spendGear, spendMagistrate, spendChurch, spendTavernInn, spendTavernBarn, spendRecruit, spendForge, spendTannery, spendBags, spendAlchemy}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Errorf("категория %q повторяется", id)
		}
		seen[id] = true
		if !strings.HasPrefix(id, "spend.") || hasCyrillic(id) {
			t.Errorf("идентификатор категории %q должен быть языконезависимым", id)
		}
		if dictRU[id] == "" || dictEN[id] == "" {
			t.Errorf("у категории %q нет названия в словарях", id)
		}
	}
}

func TestTownSignsFitTheSignboard(t *testing.T) {
	for _, key := range townSignKeys {
		for lang, dict := range map[Language]map[string]string{LangRU: dictRU, LangEN: dictEN} {
			text, ok := dict[key]
			if !ok || text == "" {
				t.Errorf("%s: нет подписи %q", lang, key)
				continue
			}
			lines := strings.Split(text, "\n")
			if len(lines) > 2 {
				t.Errorf("%s: подпись %q больше двух строк", lang, key)
			}
			for _, line := range lines {
				if n := utf8.RuneCountInString(line); n > townSignLen {
					t.Errorf("%s: строка %q подписи %q длиннее вывески (%d > %d)", lang, line, key, n, townSignLen)
				}
			}
		}
	}
	if len(townSignKeys) != len(townSpecs) {
		t.Errorf("подписей %d, а зданий %d", len(townSignKeys), len(townSpecs))
	}
}

func TestCodexTabsAndTiersAreTranslated(t *testing.T) {
	for _, lang := range []Language{LangRU, LangEN} {
		names := codexTabNames(lang)
		if len(names) != codexTabCount {
			t.Fatalf("%s: вкладок %d, ожидалось %d", lang, len(names), codexTabCount)
		}
		for i, n := range names {
			if n == "" || strings.HasPrefix(n, "codex.tab.") {
				t.Errorf("%s: у вкладки %d нет названия", lang, i)
			}
		}
	}
	if got := T(LangEN, "codex.tier", 3); got != "T3" {
		t.Errorf("английский ярус: %q", got)
	}
	if got := T(LangRU, "codex.tier", 3); got != "Т3" {
		t.Errorf("русский ярус: %q", got)
	}
}

func TestMenuFollowsTheLanguage(t *testing.T) {
	useTempSave(t)
	seedRNG(5)
	russianOnly := []string{"Quit", "Start Expedition", "Language", "Auto-play", "Manual control", "[Space] Start", "[Q] Quit"}
	for _, size := range [][2]int{{140, 46}, {90, 30}, {50, 22}} {
		m := initialModel()
		m.TermWidth, m.TermHeight = size[0], size[1]

		m.Lang = LangEN
		m.relocalizeStart()
		if out := m.View(); hasCyrillic(out) {
			t.Errorf("%v: в английском меню остался русский текст: %q", size, firstCyrillicLine(out))
		}

		m.Lang = LangRU
		m.relocalizeStart()
		out := m.View()
		for _, bad := range russianOnly {
			if strings.Contains(out, bad) {
				t.Errorf("%v: в русском меню остался английский текст %q", size, bad)
			}
		}
	}
}

var allTownPhases = []TownPhase{
	TownPhaseSellLoot, TownPhaseMagistrate, TownPhaseChurch, TownPhaseTavern, TownPhaseGuild,
	TownPhaseSmithy, TownPhaseTannery, TownPhaseAlchemist, TownPhaseDepart,
}

func townCardModel(t *testing.T, lang Language, phase TownPhase) *Model {
	t.Helper()
	useTempSave(t)
	seedRNG(6)
	m := initialModel()
	m.Lang = lang
	m.relocalizeStart()
	m.State = StatePlaying
	m.Gold = 3000
	m.MaxFloor = 5
	m.Party[0].IsDowned = true
	m.Party[1].Stress = 90
	m.Party[2].HP = 1
	m.InTown, m.TownVisit, m.TownPhase = true, true, phase
	m.initTownScreen(phase)
	return &m
}

func TestTownCardsHaveNoRussianInEnglish(t *testing.T) {
	for _, phase := range allTownPhases {
		m := townCardModel(t, LangEN, phase)
		out := m.renderTownHub(110, 40)
		if hasCyrillic(out) {
			t.Errorf("%v: русский текст в английской карточке: %q", phase, firstCyrillicLine(out))
		}
	}
}

func TestTownCardsHaveNoEnglishLeftoversInRussian(t *testing.T) {
	englishOnly := []string{"AUTOPILOT", "UPGRADED", "HEALED", "FREE", "Cleansing", "Places to rest", "Filling the ranks",
		"Available upgrades", "Mutations and potions", "Veteran", "Militiaman", "Stabilization", "Mutation", "Potion"}
	for _, phase := range allTownPhases {
		m := townCardModel(t, LangRU, phase)
		out := m.renderTownHub(110, 40)
		for _, bad := range englishOnly {
			if strings.Contains(out, bad) {
				t.Errorf("%v: английский текст %q в русской карточке", phase, bad)
			}
		}
	}
}

func TestAutopilotTownSummariesFollowTheLanguage(t *testing.T) {
	m := townModel(t)
	m.Lang = LangEN
	m.relocalizeStart()
	m.Tactics.DisableTraining = true
	m.Gold = 3000
	for _, h := range m.Party {
		h.HP = max(1, h.MaxHP/3)
		h.Stress = 150
	}
	m.Party[1].IsDowned = true
	m.enterTown(RetreatLowHP)
	runTownAuto(t, m, 6000)

	for _, line := range m.TownHistory {
		if hasCyrillic(line) {
			t.Errorf("в английской истории города русский текст: %q", line)
		}
	}
	for _, line := range m.Logs {
		if hasCyrillic(line) {
			t.Errorf("в английском журнале русский текст: %q", line)
		}
	}
}

func TestChurchOfferDescriptionsAreTranslated(t *testing.T) {
	for _, lang := range []Language{LangRU, LangEN} {
		useTempSave(t)
		seedRNG(7)
		m := initialModel()
		m.Lang = lang
		m.Gold = 3000
		m.Party[1].Stress = 90
		m.Party[1].Affliction = AfflictionParanoid
		for _, off := range m.GenerateChurchOffers() {
			if strings.Contains(off.Desc, "paranoid") {
				t.Errorf("%s: в описании сырой идентификатор состояния: %q", lang, off.Desc)
			}
			if lang == LangEN && hasCyrillic(off.Desc) {
				t.Errorf("английское описание на русском: %q", off.Desc)
			}
		}
	}
}
