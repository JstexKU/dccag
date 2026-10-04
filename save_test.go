package main

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempSave(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dccag", "save.json")
	savePathOverride = path
	saveEnabled = true
	t.Cleanup(func() { savePathOverride = ""; saveEnabled = true })
	return path
}

func TestSaveRoundTrip(t *testing.T) {
	useTempSave(t)

	tc := DefaultTactics()
	tc.SkipTraps = true
	tc.RetreatHPPct = 50
	legacy := TownLegacy{TreasuryGold: 120, SmithyLevel: 3, TanneryLevel: 2, ChurchLevel: 1, TavernLevel: 4, TotalInvested: 900}

	persistState(LangEN, legacy, tc, nil)

	got, ok := loadSave()
	if !ok {
		t.Fatal("сохранение не прочиталось")
	}
	if got.Legacy != legacy {
		t.Errorf("наследие изменилось: %+v != %+v", got.Legacy, legacy)
	}
	if got.Tactics != tc {
		t.Errorf("тактика изменилась: %+v != %+v", got.Tactics, tc)
	}
	if got.Lang != LangEN {
		t.Errorf("язык потерян: %q", got.Lang)
	}
}

func TestSaveOverwrite(t *testing.T) {
	useTempSave(t)

	first := TownLegacy{TreasuryGold: 10, SmithyLevel: 1}
	second := TownLegacy{TreasuryGold: 999, SmithyLevel: 7}

	if err := writeSave(SaveData{Lang: LangRU, Legacy: first, Tactics: DefaultTactics()}); err != nil {
		t.Fatalf("первое сохранение: %v", err)
	}
	if err := writeSave(SaveData{Lang: LangEN, Legacy: second, Tactics: DefaultTactics()}); err != nil {
		t.Fatalf("повторное сохранение: %v", err)
	}

	got, ok := loadSave()
	if !ok {
		t.Fatal("повторно сохранённый файл не прочитался")
	}
	if got.Legacy != second {
		t.Errorf("сохранение не было заменено: %+v != %+v", got.Legacy, second)
	}
	if got.Lang != LangEN {
		t.Errorf("язык не был заменён: %q", got.Lang)
	}
}

func TestLoadSaveMissingAndCorrupted(t *testing.T) {
	path := useTempSave(t)

	if _, ok := loadSave(); ok {
		t.Fatal("несуществующий файл не должен читаться как сохранение")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadSave(); ok {
		t.Fatal("битый JSON не должен читаться как сохранение")
	}
}

func TestLoadSaveSanitizesValues(t *testing.T) {
	path := useTempSave(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"version":1,"lang":"xx","legacy":{"TreasuryGold":-50,"SmithyLevel":100000,"TotalInvested":-1},` +
		`"tactics":{"retreat_hp_pct":999,"retreat_min_alive":-3,"flee_hp_pct":1,"potion_hp_pct":1}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := loadSave()
	if !ok {
		t.Fatal("файл должен читаться")
	}
	if got.Legacy.TreasuryGold != 0 || got.Legacy.TotalInvested != 0 || got.Legacy.SmithyLevel != 99 {
		t.Errorf("наследие не очищено: %+v", got.Legacy)
	}
	if got.Lang != LangRU {
		t.Errorf("неизвестный язык должен сброситься на RU, получено %q", got.Lang)
	}
	if got.Tactics != got.Tactics.Clamped() {
		t.Errorf("тактика не приведена к диапазонам: %+v", got.Tactics)
	}
}

func TestSaveDisabledWritesNothing(t *testing.T) {
	path := useTempSave(t)
	saveEnabled = false
	persistState(LangRU, TownLegacy{TreasuryGold: 5}, DefaultTactics(), nil)
	if _, err := os.Stat(path); err == nil {
		t.Fatal("при -no-save файл создаваться не должен")
	}
}

func TestLegacyAfterRun(t *testing.T) {
	seedRNG(3)
	m := initialModel()
	m.Gold = 1000
	m.Legacy.TotalInvested = 500
	got := legacyAfterRun(m)
	wantTax := int(1000 * LegacyTaxRate)
	if got.TreasuryGold != wantTax {
		t.Errorf("казна: %d, ожидалось %d", got.TreasuryGold, wantTax)
	}
	if got.TotalInvested != 500+wantTax {
		t.Errorf("вклады: %d, ожидалось %d", got.TotalInvested, 500+wantTax)
	}
}

func TestSaveRoundTripWithHero(t *testing.T) {
	useTempSave(t)
	bp := HeroBlueprint{Name: "Эдан", Gender: GenderFemale, Race: RaceElf, Class: ClassMage, Calling: CallingStrategist}
	bp.AdjustPoint(StatAtk, 1)
	bp.AdjustPoint(StatHP, 2)

	persistState(LangRU, TownLegacy{TreasuryGold: 9}, DefaultTactics(), &bp)
	got, ok := loadSave()
	if !ok || got.Hero == nil {
		t.Fatal("герой должен прочитаться из сохранения")
	}
	if *got.Hero != bp {
		t.Fatalf("герой изменился: %+v != %+v", *got.Hero, bp)
	}

	persistState(LangRU, TownLegacy{}, DefaultTactics(), nil)
	got, _ = loadSave()
	if got.Hero != nil {
		t.Fatal("без героя в сохранении его быть не должно")
	}
}

func TestLoadSaveSanitizesHero(t *testing.T) {
	path := useTempSave(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	raw := `{"version":1,"lang":"ru","hero":{"name":"Bob%d<>","gender":9,"race":"dragon","class":"hacker","calling":"king","points":[99,99,99,99,99]}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := loadSave()
	if !ok || got.Hero == nil {
		t.Fatal("испорченный герой с корректным именем должен исправляться, а не пропадать")
	}
	if got.Hero.Name != "Bobd" || !validRace(got.Hero.Race) || !validClass(got.Hero.Class) || !validCalling(got.Hero.Calling) {
		t.Fatalf("герой не очищен: %+v", *got.Hero)
	}
	if got.Hero.PointsUsed() > creatorPoints {
		t.Fatalf("лимит очков превышен: %d", got.Hero.PointsUsed())
	}

	raw = `{"version":1,"lang":"ru","hero":{"name":"%%","race":"elf","class":"mage"}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok = loadSave()
	if !ok || got.Hero != nil {
		t.Fatal("герой без пригодного имени должен отбрасываться")
	}
}

func TestOldSaveWithoutHeroStillLoads(t *testing.T) {
	path := useTempSave(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"version":1,"lang":"en","legacy":{"TreasuryGold":40}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := loadSave()
	if !ok || got.Hero != nil || got.Legacy.TreasuryGold != 40 || got.Lang != LangEN {
		t.Fatalf("сохранение версии 2.x должно читаться: %+v", got)
	}
}
