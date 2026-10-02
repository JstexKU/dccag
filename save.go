package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const saveVersion = 1

// SaveData — всё, что переживает закрытие игры: наследие столицы,
// язык интерфейса и настройки автопилота.
type SaveData struct {
	Version int        `json:"version"`
	Lang    Language   `json:"lang"`
	Legacy  TownLegacy `json:"legacy"`
	Tactics Tactics    `json:"tactics"`
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
	sd.Tactics = sd.Tactics.Clamped()
	sd.Legacy = sd.Legacy.sanitized()
	if sd.Lang != LangEN && sd.Lang != LangRU {
		sd.Lang = LangRU
	}
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
func persistState(lang Language, legacy TownLegacy, t Tactics) {
	if !saveEnabled {
		return
	}
	_ = writeSave(SaveData{Lang: lang, Legacy: legacy, Tactics: t})
}

// legacyAfterRun считает наследие, которое достанется следующему забегу:
// налог с добычи уходит в казну и учитывается как вклад в развитие столицы.
func legacyAfterRun(m Model) TownLegacy {
	saved := int(float64(m.Gold) * LegacyTaxRate)
	legacy := m.Legacy
	legacy.TreasuryGold = saved
	legacy.TotalInvested += saved
	return legacy
}
