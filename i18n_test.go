package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var formatVerb = regexp.MustCompile(`%[-+#0]*[0-9]*(?:\.[0-9]+)?[sdvfqxc]`)

func countVerbs(s string) int {
	n := 0
	for _, v := range formatVerb.FindAllString(s, -1) {
		if v != "%%" {
			n++
		}
	}
	return n
}

// Ключи во всех поддерживаемых словарях (RU, EN, BE) должны совпадать один в один.
func TestDictKeyParity(t *testing.T) {
	dicts := map[Language]map[string]string{
		LangRU: dictRU,
		LangEN: dictEN,
		LangBE: dictBE,
	}

	for l1, d1 := range dicts {
		for l2, d2 := range dicts {
			if l1 == l2 {
				continue
			}
			for k := range d1 {
				if _, ok := d2[k]; !ok {
					t.Errorf("ключ %q есть в %s, но отсутствует в %s", k, l1, l2)
				}
			}
		}
	}
}

// Переводы одного ключа во всех словарях должны принимать одинаковое число аргументов,
// иначе fmt.Sprintf выведет «%!s(MISSING)» или «%!(EXTRA ...)».
func TestFormatVerbParity(t *testing.T) {
	targetDicts := map[Language]map[string]string{
		LangEN: dictEN,
		LangBE: dictBE,
	}

	for k, ru := range dictRU {
		for lang, dict := range targetDicts {
			val, ok := dict[k]
			if !ok {
				continue
			}
			if countVerbs(ru) != countVerbs(val) {
				t.Errorf("ключ %q: в RU %d аргументов, в %s %d\n  RU: %s\n  %s: %s",
					k, countVerbs(ru), lang, countVerbs(val), ru, lang, val)
			}
		}
	}
}

func goSources(t *testing.T) map[string]string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || strings.HasPrefix(f, "i18n_") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		out[f] = string(raw)
	}
	return out
}

var literalKeyCall = regexp.MustCompile(`\bT\(\s*[\w.]+\s*,\s*"([\w.]+)"\s*[,)]`)

// Любой ключ, который код запрашивает литералом, обязан существовать во всех словарях.
func TestUsedKeysExist(t *testing.T) {
	dicts := map[Language]map[string]string{
		LangRU: dictRU,
		LangEN: dictEN,
		LangBE: dictBE,
	}

	for file, src := range goSources(t) {
		for _, m := range literalKeyCall.FindAllStringSubmatch(src, -1) {
			key := m[1]
			for lang, dict := range dicts {
				if _, ok := dict[key]; !ok {
					t.Errorf("%s: ключ %q не найден в словаре %s", file, key, lang)
				}
			}
		}
	}
}

func TestTranslateMissingKeyDoesNotPanic(t *testing.T) {
	_ = T(LangRU, "no.such.key")
	_ = T(LangEN, "no.such.key", 1, 2)
	_ = T(LangBE, "no.such.key", 1, 2)
}