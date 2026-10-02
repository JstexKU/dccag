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

// Ключи в русском и английском словарях должны совпадать один в один.
func TestDictKeyParity(t *testing.T) {
	for k := range dictRU {
		if _, ok := dictEN[k]; !ok {
			t.Errorf("ключ %q есть в RU, но отсутствует в EN", k)
		}
	}
	for k := range dictEN {
		if _, ok := dictRU[k]; !ok {
			t.Errorf("ключ %q есть в EN, но отсутствует в RU", k)
		}
	}
}

// Переводы одного ключа должны принимать одинаковое число аргументов,
// иначе fmt.Sprintf выведет «%!s(MISSING)» или «%!(EXTRA ...)».
func TestFormatVerbParity(t *testing.T) {
	for k, ru := range dictRU {
		en, ok := dictEN[k]
		if !ok {
			continue
		}
		if countVerbs(ru) != countVerbs(en) {
			t.Errorf("ключ %q: в RU %d аргументов, в EN %d\n  RU: %s\n  EN: %s",
				k, countVerbs(ru), countVerbs(en), ru, en)
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

// Любой ключ, который код запрашивает литералом, обязан существовать в словарях.
func TestUsedKeysExist(t *testing.T) {
	for file, src := range goSources(t) {
		for _, m := range literalKeyCall.FindAllStringSubmatch(src, -1) {
			key := m[1]
			if _, ok := dictRU[key]; !ok {
				t.Errorf("%s: ключ %q не найден в RU-словаре", file, key)
			}
			if _, ok := dictEN[key]; !ok {
				t.Errorf("%s: ключ %q не найден в EN-словаре", file, key)
			}
		}
	}
}

func TestTranslateMissingKeyDoesNotPanic(t *testing.T) {
	_ = T(LangRU, "no.such.key")
	_ = T(LangEN, "no.such.key", 1, 2)
}
