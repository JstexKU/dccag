package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// replayIterations — сколько итераций автопилота в каждом эталонном забеге.
const replayIterations = 20000

// fingerprint — сжатый отпечаток итогового состояния забега. Нужен, чтобы
// рефакторинг step()/moveTo() можно было доказать «поведение не изменилось».
func fingerprint(m Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, "state=%v floor=%d cleared=%d town=%v gold=%d steps=%d chests=%d barrels=%d logs=%d pos=%d,%d\n",
		m.State, m.Floor, m.Stats.FloorsCleared, m.InTown, m.Gold, m.Stats.TotalSteps,
		m.Stats.ChestsOpened, m.Stats.BarrelsBlown, len(m.Logs), m.PartyPos.X, m.PartyPos.Y)
	for _, h := range m.Party {
		fmt.Fprintf(&b, "hero lvl=%d exp=%d hp=%d/%d mp=%d stress=%d dead=%v\n",
			h.Level, h.Exp, h.HP, h.MaxHP, h.MP, h.Stress, h.IsDead)
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(b.String())))[:16] + "\n" + b.String()
}

// Эталонные отпечатки хранятся в testdata/replay_golden.txt. Если файла нет, тест
// создаёт его и пропускается. Файл нужно создать ДО рефакторинга и закоммитить;
// намеренное изменение баланса — удалить файл и перегенерировать.
func TestReplayMatchesGolden(t *testing.T) {
	var got strings.Builder
	// Забеги длинные: на широких картах один этаж занимает тысячи шагов, а проверять
	// нужно и спуски, и бои с боссами, и отступления в город.
	for _, seed := range []int64{1, 7, 42, 99} {
		fmt.Fprintf(&got, "seed %d\n%s\n", seed, fingerprint(runHeadless(seed, replayIterations)))
	}

	path := filepath.Join("testdata", "replay_golden.txt")
	want, err := os.ReadFile(path)
	if err != nil {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Skipf("создан эталон %s — закоммить его и запусти тест снова", path)
	}
	if string(want) != got.String() {
		t.Fatalf("забег с зерном разошёлся с эталоном (изменился порядок операций или баланс).\nожидалось:\n%s\nполучено:\n%s", want, got.String())
	}
}

// Предпросмотр защиты обязан совпадать с TotalDef героя без снаряжения.
func TestPreviewMatchesTotalDefWithoutGear(t *testing.T) {
	seedRNG(11)
	for _, class := range AllClasses {
		for _, race := range AllRaces {
			bp := HeroBlueprint{Name: "Тест", Race: race, Class: class}
			bp.AdjustPoint(StatDef, 1)
			bp.AdjustPoint(StatDef, 1)
			h := createLeaderHero(bp, 0)
			h.Weapon, h.Head, h.Chest, h.Legs = nil, nil, nil, nil
			if p := bp.Preview(0); p.Def != h.TotalDef() {
				t.Errorf("%s/%s: предпросмотр Def=%d, TotalDef=%d", class, race, p.Def, h.TotalDef())
			}
		}
	}
}

// Биомы идут по кругу: этаж biomeCount+1 — снова первый биом.
func TestBiomeIndexCycles(t *testing.T) {
	if biomeIndex(1) != 0 || biomeIndex(biomeCount) != biomeCount-1 || biomeIndex(biomeCount+1) != 0 {
		t.Fatal("цикл биомов нарушен")
	}
	seen := map[BiomeType]bool{}
	for f := 1; f <= biomeCount; f++ {
		seen[getBiome(f).Name] = true
	}
	if len(seen) != biomeCount {
		t.Fatalf("getBiome отдаёт %d разных биомов из %d", len(seen), biomeCount)
	}
}
