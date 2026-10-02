# Changelog

## 2.8.0

### Added
- Autopilot tactics screen (`T`): retreat / flee / potion thresholds, avoid trapped chests and altars, accept risky bargains.
- Room events (`?`): campfire, wandering merchant, whispering idol, locked vault, ambush.
- Persistent save (legacy, language, tactics) in the user config directory.
- Flags `-seed`, `-no-save`, `-reset-save`, `-version`.
- Test suite and CI on pull requests; release builds now publish `sha256sums.txt` and embed the version.

### Fixed
- Tick chains could stack after closing a window or resuming from pause, speeding the game up; stale ticks are now discarded.
- `R` reset the run from the Armory and Codex; it now works only on the Defeat and Glory screens.
- Missing `combat.log.cleric_heal` key (a raw key was printed when a Cleric healed).
- `combat.log.mage_storm` was missing its damage argument.
- Gold, kill statistics and hunt contracts were credited only for basic attacks and some mage spells; every kill is now handled in one place.
- Tavern level was displayed using the Tannery level.
- Several English log lines had fewer format arguments than their Russian originals (verbs were dropped); aligned.
- Altar text said "+2 Atk" while the altar grants +3.
- `average_atk` in the telemetry report now really averages across runs.
- Hard-coded Russian strings ("Ур.", hit log line) moved into the dictionaries.
- Progress (legacy, language, tactics) is now saved on quit and on every restart instead of being lost.

### Changed
- `go.mod` now declares Go 1.24 (was 1.27); CI reads the version from `go.mod`.
- Dictionaries split into `i18n_ru.go` and `i18n_en.go`.
- `.gitignore` cleaned up, `LICENSE` added, README rewritten.

## [2.8.0] - 2026-10-02

### Добавлено
- **Торговля снаряжением в Столице:** автоматическая покупка недостающей базовой экипировки и плановые апгрейды вещей у торговцев на рынке.
- **Интеллектуальный сбор снаряжения в подземелье:** отряд без оружия/брони целенаправленно обходит монстров и открывает сундуки/тайники в первую очередь.
- **События в залах (тайл `?`):** кострища путников, бродячие торговцы, шепчущие идолы, запертые хранилища и внезапные засады.
- **Экран тактики автопилота:** настройка порогов отступления, побега, применения зелий и реакции на алтари/ловушки.
- **Система сохранений (`save.go`):** сохранение прогресса инвестиций столицы, настроек тактики и выбранного языка между сессиями.
- **Автоматические тесты:** проверка паритета локализации (`i18n_test.go`), механик событий (`events_test.go`), тактик (`tactics_test.go`) и сохранений (`save_test.go`).
