package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Снимки карточек заведений: все здания, три языка, два размера терминала, автопилот
// и ручной режим. Нужны, чтобы рефакторинг отрисовки (общие шапки, списки) был
// доказуемо побуквенным. Эталон создаётся при первом запуске (testdata/town_cards_golden.txt)
// и записывается ДО правки отрисовки.
func townCardsSnapshot(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	sizes := [][2]int{{110, 40}, {64, 24}}
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		for _, phase := range allTownPhases {
			for _, size := range sizes {
				for _, manual := range []bool{false, true} {
					m := townCardModel(t, lang, phase)
					m.ManualMode = manual
					out := stripANSI(m.renderTownHub(size[0], size[1]))

					fmt.Fprintf(&b, "=== lang=%s phase=%v size=%dx%d manual=%v\n", lang, phase, size[0], size[1], manual)
					for _, line := range strings.Split(out, "\n") {
						b.WriteString(strings.TrimRight(line, " "))
						b.WriteByte('\n')
					}
				}
			}
		}
	}
	return b.String()
}

func TestTownCardsMatchGolden(t *testing.T) {
	got := townCardsSnapshot(t)

	path := filepath.Join("testdata", "town_cards_golden.txt")
	want, err := os.ReadFile(path)
	if err != nil {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Skipf("создан эталон %s — закоммить его и запусти тест снова", path)
	}
	if string(want) == got {
		return
	}

	// Первое расхождение с названием снимка и номером строки.
	wl, gl := strings.Split(string(want), "\n"), strings.Split(got, "\n")
	header := ""
	for i := 0; i < len(wl) && i < len(gl); i++ {
		if strings.HasPrefix(wl[i], "=== ") {
			header = wl[i]
		}
		if wl[i] != gl[i] {
			t.Fatalf("карточка разошлась с эталоном\n%s\nстрока %d:\nожидалось: %q\nполучено:  %q", header, i+1, wl[i], gl[i])
		}
	}
	t.Fatalf("снимки разошлись по длине: %d != %d строк", len(wl), len(gl))
}
