package main

import "testing"

// Таблица биомов цельная: у каждого биома есть мобы, мини-босс и дневник в библиотеке,
// названия не повторяются, а длина цикла равна длине таблицы.
func TestBiomeTableIsConsistent(t *testing.T) {
	seen := map[BiomeType]bool{}
	for i, def := range biomeTable {
		if def.Name == "" || seen[def.Name] {
			t.Errorf("биом %d: пустое или повторное название %q", i, def.Name)
		}
		seen[def.Name] = true

		if len(def.Mobs) == 0 {
			t.Errorf("%s: нет мобов", def.Name)
		}
		for _, m := range def.Mobs {
			if m.Type == "" || m.Glyph == 0 || m.HP <= 0 || m.Atk <= 0 {
				t.Errorf("%s: неполный моб %+v", def.Name, m)
			}
		}
		if got := len(def.MobTypes()); got != len(def.Mobs) {
			t.Errorf("%s: MobTypes отдаёт %d из %d", def.Name, got, len(def.Mobs))
		}

		mb := def.MiniBoss
		if mb.Type == "" || mb.NameKey == "" || mb.Glyph == 0 || mb.Atk <= 0 || mb.Def <= 0 {
			t.Errorf("%s: неполный мини-босс %+v", def.Name, mb)
		}

		if def.EnvHazardKey == "" || def.WallColor == "" || def.FloorColor == "" || def.FloorRune == 0 {
			t.Errorf("%s: неполный внешний вид", def.Name)
		}

		vol := FindLoreVolume("diary_" + string(def.Name))
		if vol == nil {
			t.Errorf("%s: нет дневника в библиотеке", def.Name)
		} else if vol.BiomeOrigin != def.Name || vol.VolumeNum != i+1 {
			t.Errorf("%s: дневник не совпал с биомом (том %d, ждали %d)", def.Name, vol.VolumeNum, i+1)
		}
	}
	if len(seen) != biomeCount {
		t.Errorf("биомов %d, длина цикла %d", len(seen), biomeCount)
	}
}
