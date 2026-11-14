# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [3.4.1] - 2026-11-13

### Fixed
- Fixed save file migration for pre-3.3 saves (training gap defaults).
- Fixed hero sanitization logic for corrupted save files.
- Fixed companion validation errors on game load.

## [3.4.0] - 2026-11-12

### Added
- Hero creator system with blueprint saving.
- Companion system (support for up to 4 companions per leader).
- Save format extended with `Hero` and `Companions` fields.

### Changed
- `persistState()` now accepts hero and companions parameters for accurate state persistence.

## [3.3.3] - 2026-11-12

### Fixed
- Fixed localization slot validation (empty string checks).
- Fixed save migration logic for pre-3.3 tactics configurations.

## [3.3.2] - 2026-11-11

### Added
- Expanded Guild Contract System with 19 contract types (reconnaissance, combat challenges, etc.).
- Smart filtering for contracts by depth, boss, and biome.
- Dynamic viewport for inventory in Market UI with auto-follow cursor.
- Adaptive scroll indicators for better navigation.

### Changed
- Contract system now supports advanced filtering logic.
- Market UI viewport dynamically adjusts to content size.

## [3.3.1] - 2026-11-10

### Fixed
- Fixed duplicate `revealFog` calls during dungeon generation.
- Fixed undefined boss/elite attribute errors.
- Migrated ~180 hardcoded UI strings to localization dictionaries.
- Updated `TVerb` to accept dynamic keys.
- Made telemetry spending categories language-independent (`spend.*`).

### Changed
- Refactored localization system for better maintainability and scalability.

## [3.3.0] - 2026-11-09

### Added
- 7 playable races: Human, Elf, Beastman, Olongr, Dwarf, Demonkin, Dragonkin.
- Unique racial traits and class synergies.
- Belarusian localization (RU/EN/BE) with in-game switching via `L` key.
- 10 legendary relics with active mechanics: Compass of Greed, Martyr's Crown, Holy Grail, Pioneer's Eye, Quicksilver Vial, Obsidian Anvil, Whispering Skull, Phoenix Core, Cornucopia, Inquisitor's Seal.
- Normalized hotkey handling (works correctly on any keyboard layout).
- Extended Codex with a new 6th tab.

### Changed
- Complete localization overhaul with three-language support framework.
- Relic system redesigned to include active abilities and deeper gameplay impact.

## [3.2.0] - 2026-11-06

### Added
- Capital Library and Lore system.
- Six-category registry of 80 books and manuscripts.
- Interactive lore reader activated via `B` key.
- Lore tiles hidden in dungeons and Forgotten Archives.
- Lore discovery rewards: class perks, racial bonuses, experience, and stress relief.

### Changed
- Integrated lore discovery directly with dungeon exploration mechanics.

## [3.1.0] - 2026-11-03

### Added
- Hero blueprint saving integrated into save files.
- Save format version 1 with dedicated hero support.
- Hero data sanitization on load (validates name, race, class, calling, and stat points).

### Changed
- `SaveData` structure extended with the `Hero` field.
- `persistState()` updated to accept and serialize hero parameters.

## [3.0.0] - 2026-10-31

### Added
- Full capital map rework with manual movement mechanics.
- Expansion to 7 races (introducing Dwarf, Demonkin, Dragonkin).
- 10 legendary relics with active mechanics.
- Complete multi-language localization system (RU/EN/BE).
- Manual mode toggle for player-controlled dungeon navigation.
- Deterministic simulation support via `-seed` command-line flag.

### Changed
- Major UI overhaul for capital navigation and town interactions.
- Relic system completely redesigned from passive bonuses to active mechanics.
- Core localization framework implemented across all UI and log elements.

## [2.9.2] - 2026-10-27

### Fixed
- Improved save file sanitization and legacy value clamping.
- Fixed tick generation logic to prevent stale updates after closing overlays (e.g., Armory).

## [2.9.1] - 2026-10-26

### Added
- `SaveData` structure with explicit version, language, legacy, and tactics fields.
- Atomic save writing via temporary files to prevent corruption.
- `persistState()` function for safe and reliable state saving.

### Changed
- Refactored the entire save system for improved reliability and crash resistance.

## [2.9.0] - 2026-10-25

### Added
- RNG seeding with `-seed` flag for fully reproducible runs.
- `seedRNG()` function to enable deterministic generation for testing and speedrunning.

### Changed
- Random number generation now supports a deterministic mode alongside the default time-based seeding.

## [2.8.4 – 2.8.9] - 2026-10-19 – 2026-10-23

### Fixed
- Save file format stability improvements and legacy tax calculation fixes.
- Tactics clamping on load to prevent out-of-bounds values.
- Language validation in save files to prevent crashes on corrupted data.
- Multiple stabilization patches preparing the codebase for the 2.9.0 release.

## [2.8.3] - 2026-10-18

### Fixed
- Verified legacy calculation logic after a completed run.
- Added robust save overwrite tests.

## [2.8.1] - 2026-10-17

### Added
- `TestSaveOverwrite` and `TestLegacyAfterRun` test suites.

### Fixed
- Save file round-trip verification to ensure data integrity.

## [2.8.0] - 2026-10-16

### Changed
- Synchronized `var version` with the menu display string to eliminate version mismatch confusion.

## [2.7.2] - 2026-10-15

### Added
- Modular codebase split into separate files: `rng.go`, `save.go`, `militia.go`, `tactics.go`.
- Comprehensive test coverage: `save_test.go`, `sim_test.go`, `tactics_test.go`.
- New data types: `RetreatReason`, `HealingUrgency`, `LostInAbyss`, `TickGen`, `Tactics`, `TileEvent`, `EventsSeen`, `TotalInvested`.
- `buyMissingOrBetterGear()` function in town for auto-buying missing or superior equipment.
- Adaptive town display modes (`renderTownHubWide`, `renderTownHubMedium`, `renderTownHubCompact`).

### Changed
- Major architectural refactoring into dedicated modules.
- Town UI now dynamically adapts to terminal dimensions.

## [2.7.1] - 2026-10-13

### Added
- `renderEnemyStrip` for clear enemy health visualization.
- `renderLogBox` with adjustable `boxH` parameter.
- Localized `renderControls` for consistent UI.
- Militia tier and progress display integrated into the stats screen.

## [2.7.0] - 2026-10-12

### Added
- `RunCounted` flag in `Model` to prevent double incrementing of `TotalRuns`.
- `RestTurnsLeft` mechanic for camp and rest management.
- `runtime.GOMAXPROCS(1)` in `main()` for stable execution on 32-bit architectures and iSH environments.

### Fixed
- Resolved double run counting bug.
- Improved overall stability on constrained environments.

## [2.6.0] - 2026-10-09

### Added
- Adaptive interface system.
- Display layout modes: `LayoutLandscape`, `LayoutPortraitWide`, `LayoutPortrait`.
- Adaptive hero cards: `CardWide`, `CardMedium`, `CardCompact`.

### Changed
- UI rendering now dynamically adapts to terminal width and height, ensuring readability on all screen sizes.

## [2.5.0] - 2026-10-06

### Added
- 4 playable races: Humans, Elves, Beastmen, Olongrs, each with unique `RaceModifiers`.
- Expanded class roster to 10 classes (added Paladin, Ranger, Monk, Bard, Warlock).

### Changed
- Implemented the foundational race system with stat modifiers.
- Significantly expanded class variety and role diversity.

## [2.4.3] - 2026-10-03

### Fixed
- `GetClassMutationPreference` now correctly considers fallen heroes' history.
- Critical HP drops now properly prioritize "Titan's Blood" mutation for survival.

## [2.4.2] - 2026-10-02

### Changed
- Improved alchemical mutation logic by incorporating historical run data for smarter recommendations.

## [2.4.1] - 2026-10-01

### Changed
- Tightened rest mechanics (camping now only works up to floor 5; deeper floors are pure hardcore).
- Adjusted base HP stats for better early-game balance.
- Rebalanced church and alchemist service costs.

## [2.4.0] - 2026-09-30

### Added
- `Language` type (`LangRU`, `LangEN`).
- Dictionary key system: `NameKey`, `TitleKey`, `BaseNameKey`.
- `T()` and `TVerb()` translation functions.

### Changed
- **Major localization refactor**: Replaced all hardcoded UI and log strings with dictionary lookups, paving the way for multi-language support.

## [2.3.0] - 2026-09-27

### Added
- Gender-aware verb support (`Verb`) for grammatically correct logs in Russian.
- Procedural generation of named capital establishments (unique forge, tavern, and guild names).

## [2.2.0] - 2026-09-24

### Added
- `Tannery` crafting specialization (`TanneryLevel`).
- Leather and cloth material tiers (`LeatherMaterials`).
- Movement speed bonuses (`SpeedBonus`).
- Multi-potion belt system (`[]*Potion`).

### Changed
- Split the crafting system into distinct specializations (Smithy vs. Tannery).
- Expanded the potion system from single slots to belt-based inventory management.

## [2.1.0] - 2026-09-21

### Added
- Town budgeting and allocation system (`AllocateBudget`).
- Alchemical mutations with class-specific priority logic (`GetClassMutationPreference`).

## [2.0.1] - 2026-09-18

### Added
- Full level and experience progression system (`Level`, `Exp`, `GainExp`, `distributePartyExp`) for heroes.

### Changed
- Introduced a major progression overhaul: heroes now gain levels and permanently improve stats.

## [1.1.2] - 2026-09-15

### Added
- Alchemical mutations: Chimera, Rage, Titan, Ether, Bastion.
- Refined material naming conventions via the `Adj` system.

## [1.1. asdasd1] - 2026-09-14

### Added
- Extended combat logging capacity (up to 100 entries).
- Log scrolling functionality (`LogScroll`).
- Improved visual layout of the Arsenal screen.

## [1.1.0] - 2026-09-13

### Added
- Aggro and threat system (`CombatRole`).
- Damage interception and guarding mechanics (`ApplyDamage`).
- Town budgeting foundation (`TownBudget`).
- Dynamic potion pricing based on floor depth.

### Changed
- Combat system now actively tracks aggro and character roles.
- Town economy mechanics implemented.

## [1.0.0] - 2026-09-11

### Added
- Initial public release.
- Core game structure: biomes, relics, basic combat, and equipment.
- Quest system with relic-finding objectives.
- Potion system with auto-drink logic.
- Gold tracking and run statistics.
