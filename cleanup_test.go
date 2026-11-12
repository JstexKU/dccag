package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Снимок экрана статистики: три языка, обычный режим и поражение, разные размеры терминала
// и прокрутка. Эталон testdata/stats_screen_golden.txt создаётся при первом запуске.
func TestStatsScreenMatchesGolden(t *testing.T) {
	var b strings.Builder
	sizes := [][2]int{{100, 40}, {70, 20}}
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		for _, defeat := range []bool{false, true} {
			for _, size := range sizes {
				for _, scroll := range []int{0, 6} {
					seedRNG(41)
					m := initialModel()
					m.Lang = lang
					m.TermWidth, m.TermHeight = size[0], size[1]
					m.StatsScroll = scroll
					m.Gold = 1234
					m.Stats.FloorsCleared, m.Stats.TotalSteps, m.Stats.TotalGoldEarned = 7, 4200, 9876
					m.Stats.QuestsCompleted, m.Stats.UpgradesForged, m.Stats.ChestsOpened = 3, 5, 21
					m.Legacy.SmithyLevel, m.Legacy.TanneryLevel, m.Legacy.TotalInvested = 2, 1, 450
					m.Stats.FallenHeroes = []FallenHeroRecord{
						{FullName: "Тест Первый", Class: ClassWarrior, Cause: "яд", Floor: 4},
						{FullName: "Test Second", Class: ClassMage, Cause: "stress", Floor: 7, Revived: true},
						{FullName: "Третий с очень длинным именем героя", Class: ClassRogue, Cause: "ловушка", Floor: 9},
					}
					if defeat {
						m.State, m.RestartCountdown = StateDefeat, 7
					}
					out := stripANSI(m.renderStatsScreen("Итоги", "214"))
					fmt.Fprintf(&b, "=== lang=%s defeat=%v size=%dx%d scroll=%d\n", lang, defeat, size[0], size[1], scroll)
					for _, line := range strings.Split(out, "\n") {
						b.WriteString(strings.TrimRight(line, " "))
						b.WriteByte('\n')
					}
				}
			}
		}
	}
	got := b.String()

	path := filepath.Join("testdata", "stats_screen_golden.txt")
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
	if string(want) != got {
		wl, gl := strings.Split(string(want), "\n"), strings.Split(got, "\n")
		header := ""
		for i := 0; i < len(wl) && i < len(gl); i++ {
			if strings.HasPrefix(wl[i], "=== ") {
				header = wl[i]
			}
			if wl[i] != gl[i] {
				t.Fatalf("экран статистики разошёлся с эталоном\n%s\nстрока %d:\nожидалось: %q\nполучено:  %q", header, i+1, wl[i], gl[i])
			}
		}
		t.Fatalf("снимки разошлись по длине: %d != %d строк", len(wl), len(gl))
	}
}

// Перенос по словам считает ширину в ячейках терминала: широкие символы занимают две.
func TestWordWrapUsesDisplayWidth(t *testing.T) {
	if got := wordWrap("aa bb cc", 5); len(got) != 2 || got[0] != "aa bb" || got[1] != "cc" {
		t.Errorf("латиница: %q", got)
	}
	if got := wordWrap("日本 語", 6); len(got) != 2 {
		t.Errorf("широкие символы должны переноситься по ширине: %q", got)
	}
	if got := wordWrap("   ", 10); got != nil {
		t.Errorf("пустой текст: %q", got)
	}
}

// Ширина и режим карточки определяются в одном месте; явный режим главнее автоопределения.
func TestResolveCardMode(t *testing.T) {
	inner, text, auto := resolveCardMode(40, nil)
	if inner != 36 || text != 34 {
		t.Errorf("inner=%d text=%d", inner, text)
	}
	if auto != detectCardMode(inner) {
		t.Error("без явного режима он должен определяться по ширине")
	}
	for _, forced := range []CardMode{CardWide, CardMedium, CardCompact} {
		if _, _, got := resolveCardMode(40, []CardMode{forced}); got != forced {
			t.Errorf("явный режим %v проигнорирован: %v", forced, got)
		}
	}
	if inner, text, _ := resolveCardMode(1, nil); inner != 14 || text != 12 {
		t.Errorf("минимальные ширины: inner=%d text=%d", inner, text)
	}
}

// База фолиантов строится при инициализации пакета, без init().
func TestLoreDatabaseIsBuiltAtPackageInit(t *testing.T) {
	if len(AllLoreVolumes) == 0 {
		t.Fatal("база фолиантов пуста")
	}
	seen := map[string]bool{}
	for _, v := range AllLoreVolumes {
		if v.ID == "" || seen[v.ID] {
			t.Fatalf("пустой или повторный ID тома: %q", v.ID)
		}
		seen[v.ID] = true
	}
	if FindLoreVolume("diary_"+string(biomeTable[0].Name)) == nil {
		t.Error("дневник первого биома не найден")
	}
}
