package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// packDump — подробный отпечаток стаи для эталонного сравнения.
func packDump(p *MonsterPack) string {
	var b strings.Builder
	fmt.Fprintf(&b, "boss=%v mini=%v n=%d\n", p.IsBoss, p.IsMiniBoss, len(p.Members))
	for _, m := range p.Members {
		fmt.Fprintf(&b, "  #%d %s %s lvl=%d aff=%s %c %s hp=%d/%d atk=%d def=%d spd=%d exp=%d\n",
			m.ID, m.Type, m.NameKey, m.Level, m.Affix, m.Glyph, m.Color,
			m.HP, m.MaxHP, m.Atk, m.Defense, m.Speed, m.Exp)
	}
	return b.String()
}

// Стаи всех биомов (два полных цикла, с боссами и мини-боссами) на фиксированных зёрнах
// обязаны совпадать с эталоном. Эталон создаётся при первом запуске (testdata/biome_spawn_golden.txt),
// его нужно записывать ДО рефакторинга таблицы биомов.
func TestSpawnMatchesGolden(t *testing.T) {
	var got strings.Builder
	for floor := 1; floor <= 24; floor++ {
		for seed := int64(1); seed <= 4; seed++ {
			for _, boss := range []bool{false, true} {
				seedRNG(seed*1000 + int64(floor))
				fmt.Fprintf(&got, "floor %d seed %d boss %v\n%s", floor, seed, boss, packDump(spawnMonsterPack(boss, floor)))
			}
		}
	}

	path := filepath.Join("testdata", "biome_spawn_golden.txt")
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
		wl, gl := strings.Split(string(want), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(wl) && i < len(gl); i++ {
			if wl[i] != gl[i] {
				t.Fatalf("стаи разошлись с эталоном на строке %d:\nожидалось: %s\nполучено:  %s", i+1, wl[i], gl[i])
			}
		}
		t.Fatalf("стаи разошлись с эталоном по длине: %d != %d строк", len(wl), len(gl))
	}
}

// Контракт на охоту всегда нацелен на моба из биома этого этажа.
func TestHuntQuestTargetsBelongToBiome(t *testing.T) {
	pools := [][]MonsterType{
		{MobRat, MobGoblin, MobSkeleton},
		{MobSlime, MobDrowned, MobLizard},
		{MobImp, MobOrc, MobSalamander},
		{MobGargoyle, MobGolem, MobPhantom},
		{MobSproutSkeleton, MobDryad, MobBlightEnt},
		{MobSporling, MobTentacle, MobToxicBasil},
		{MobTomeBook, MobScrollMimic, MobArchiveKeeper},
		{MobObsidianBeetle, MobMinerGhoul, MobDeepTroll},
		{MobBloodCultist, MobShadowInquisitor, MobFallenCrusader},
		{MobAstralWeaver, MobChronoPhantom, MobEssenceDevour},
		{MobVoidDemon, MobDeathKnight},
	}
	for floor := 1; floor <= 2*len(pools); floor++ {
		pool := pools[(floor-1)%len(pools)]
		hunts := 0
		for seed := int64(1); seed <= 500; seed++ {
			seedRNG(seed)
			q := generateAutoQuest(floor)
			if q.Type != QuestHuntMonster {
				continue
			}
			hunts++
			ok := false
			for _, mt := range pool {
				if q.TargetMob == mt {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("этаж %d: цель охоты %q не из биома %v", floor, q.TargetMob, pool)
			}
		}
		if hunts == 0 {
			t.Fatalf("этаж %d: за 500 зёрен ни одного контракта на охоту", floor)
		}
	}
}
