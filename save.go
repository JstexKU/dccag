package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const saveVersion = 1

// SaveData — всё, что переживает закрытие игры: наследие столицы,
// язык интерфейса, настройки автопилота и созданный игроком герой-лидер.
type SaveData struct {
	Version int            `json:"version"`
	Lang    Language       `json:"lang"`
	Legacy  TownLegacy     `json:"legacy"`
	Tactics Tactics        `json:"tactics"`
	Hero    *HeroBlueprint `json:"hero,omitempty"`

	// Спутники лидера, созданные игроком (до четырёх). Без лидера не используются.
	Companions []HeroBlueprint `json:"companions,omitempty"`
}

var (
	saveEnabled      = true // отключается флагом -no-save
	savePathOverride string // используется тестами
)

func savePath() (string, error) {
	if savePathOverride != "" {
		return savePathOverride, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dccag", "save.json"), nil
}

// sanitized страхует от отрицательных и «раздутых» значений в испорченном файле сохранения.
func (l TownLegacy) sanitized() TownLegacy {
	l.TreasuryGold = max(0, l.TreasuryGold)
	l.TotalInvested = max(0, l.TotalInvested)
	l.SmithyLevel = clampInt(l.SmithyLevel, 0, 99)
	l.TanneryLevel = clampInt(l.TanneryLevel, 0, 99)
	l.ChurchLevel = clampInt(l.ChurchLevel, 0, 99)
	l.TavernLevel = clampInt(l.TavernLevel, 0, 99)
	if l.UnlockedLore == nil {
		l.UnlockedLore = NewUnlockedLoreStore()
	}
	return l
}

// loadSave читает сохранение; второе значение false, если файла нет или он повреждён.
func loadSave() (SaveData, bool) {
	if !saveEnabled {
		return SaveData{}, false
	}
	path, err := savePath()
	if err != nil {
		return SaveData{}, false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return SaveData{}, false
	}
	var sd SaveData
	if err := json.Unmarshal(raw, &sd); err != nil {
		return SaveData{}, false
	}
	if sd.Tactics == (Tactics{}) {
		sd.Tactics = DefaultTactics()
	}
	if sd.Tactics.TrainingGap == 0 {
		sd.Tactics.TrainingGap = DefaultTactics().TrainingGap // сохранения до 3.3 не знали про тренировку
	}
	sd.Tactics = sd.Tactics.Clamped()
	sd.Legacy = sd.Legacy.sanitized()

	validLang := false
	for _, l := range SupportedLanguages {
		if sd.Lang == l {
			validLang = true
			break
		}
	}
	if !validLang {
		sd.Lang = LangRU
	}

	if sd.Hero != nil {
		hero := sd.Hero.Sanitized()
		if hero.Valid() {
			sd.Hero = &hero
		} else {
			sd.Hero = nil
		}
	}
	sd.Companions = sanitizeCompanions(sd.Hero, sd.Companions)
	return sd, true
}

// writeSave атомарно записывает сохранение (через временный файл).
func writeSave(sd SaveData) error {
	path, err := savePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	sd.Version = saveVersion
	raw, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// persistState сохраняет состояние, если сохранения включены. Ошибки записи не должны ронять игру.
func persistState(lang Language, legacy TownLegacy, t Tactics, hero *HeroBlueprint, companions ...HeroBlueprint) {
	if !saveEnabled {
		return
	}
	_ = writeSave(SaveData{Lang: lang, Legacy: legacy, Tactics: t, Hero: hero, Companions: sanitizeCompanions(hero, companions)})
}

// legacyAfterRun считает наследие, которое достанется следующему забегу:
// налог с добычи уходит в казну и учитывается как вклад в развитие столицы.
func legacyAfterRun(m Model) TownLegacy {
	saved := int(float64(m.Gold) * LegacyTaxRate)
	legacy := m.Legacy
	legacy.TreasuryGold = saved
	legacy.TotalInvested += saved

	if legacy.UnlockedLore == nil {
		legacy.UnlockedLore = NewUnlockedLoreStore()
	} else {
		cloned := NewUnlockedLoreStore()
		for k, v := range legacy.UnlockedLore {
			cloned[k] = v
		}
		legacy.UnlockedLore = cloned
	}
	return legacy
}

// sanitizeCompanions оставляет только пригодных спутников (не больше четырёх) и только
// при наличии лидера: без него отряд всё равно собирается случайно.
func sanitizeCompanions(leader *HeroBlueprint, list []HeroBlueprint) []HeroBlueprint {
	if leader == nil || len(list) == 0 {
		return nil
	}
	var out []HeroBlueprint
	for _, c := range list {
		c = c.Sanitized()
		if !c.Valid() || len(out) >= partySize-1 {
			continue
		}
		out = append(out, c)
	}
	return out
}
