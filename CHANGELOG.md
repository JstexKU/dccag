# Changelog

## [2.8.0] - 2026-10-02

### Added
- Autopilot tactics screen (`T`): retreat / flee / potion thresholds, avoid trapped chests and altars, accept risky bargains.
- Room events (`?`): campfire, wandering merchant, whispering idol, locked vault, ambush.
- Persistent save (legacy, language, tactics) in the user config directory.
- Flags `-seed`, `-no-save`, `-reset-save`, `-version`.
- Equipment trading in the Capital: automatic purchases of missing basic gear and planned upgrades.
- Intelligent dungeon equipment collection: unequipped heroes prioritize monsters and containers.
- Automated tests for localization parity, room events, tactics and save/load behavior.

### Fixed
- Tick chains could stack after closing a window or resuming from pause, speeding the game up; stale ticks are now discarded.
- `R` reset the run from the Armory and Codex; it now works only on the Defeat and Glory screens.
- Missing `combat.log.cleric_heal` key (a raw key was printed when a Cleric healed).
- `combat.log.mage_storm` was missing its damage argument.
- Gold, kill statistics and hunt contracts were credited only for basic attacks and some mage spells; every kill is now handled in one place.
- Tavern level was displayed using the Tannery level.
- Several English log lines had fewer format arguments than their Russian originals; format arguments are now aligned.
- Altar text said `+2 Atk` while the altar grants `+3`.
- `average_atk` in the telemetry report now really averages across runs.
- Hard-coded Russian strings (`Ур.`, hit log line) moved into the dictionaries.
- Progress (legacy, language, tactics) is now saved on quit and on every restart instead of being lost.

### Changed
- Go 1.24 is the project baseline, and CI/release builds read the version from `go.mod`.
- Dictionaries are split into `i18n_ru.go` and `i18n_en.go`.
- `.gitignore` was cleaned up and `LICENSE` added.
- README was rewritten to document the supported release targets and verification steps.
- Pull requests run formatting, vet and tests.
- Tagged releases build Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 binaries, embed the tag version, and publish `sha256sums.txt`.
