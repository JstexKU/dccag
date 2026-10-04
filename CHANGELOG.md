# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed
- Large source files were split by responsibility (`types_*.go`, `ui_*.go`, `combat_*.go`, `party.go`, `autopilot.go`, `step.go`, `tiles.go`, `loot.go`, `economy.go`, `update.go`, `model.go` and others). Declarations were moved without changes; see the structure in the README.
- The interface takes the version from `version` through `displayVersion()` instead of hard-coded strings; release builds that pass a tag with the `v` prefix no longer show `vv2.9.0`.

### Removed
- Unused code: `Hero.Verb`, `blankLines`, `wallStyle`, `floorStyle`, `ArmorNone`, `SuffNone`.
- 45 unused dictionary keys in RU and EN (`ui.btn_*`, `codex.base_*`, `status.*` and others).
- A stray copy of `main.go` that was committed under the file name `\`.
- Citation markers (`[cite: ...]`) left in the README.

### Known issues
- The cleric title `title.resurrector` can never be earned: `Feats.Revives` is not incremented anywhere (`Hero.AddRevive` is never called); revival happens in the Church, not through a Cleric action.

## [2.9.0]

### Added
- Six new biomes extend the dungeon cycle to 11 floors (Deadwood, Fungal Caverns, Archives, Mines, Sanctuary, Astral) with new monsters and contract pools for each of them.
- Mini-boss hunt contract.

## [2.8.6] - 2026-10-03

### Added
- **Manual control mode** (`M`, flag `-manual`): direct party movement with the arrow keys.
- Combat commands:
  - `A` — basic strike without mana cost against the selected target.
  - `Space` — class skill; the percentage chance gates are bypassed.
  - `D` — guard stance (block, +3 MP, stress relief).
  - `P` — potion belt menu (`1`–`9`), ally targeting with `←`/`→`.
  - `←`/`→` — choose the target among living monsters (marked with `▶`).
  - `F` — flee attempt.
- Dungeon actions: `C` — camp (+8 HP/MP, stress relief, 25-step cooldown, blocked near monsters); `X` — return to town while standing on the exit tile (`<`).
- Context controls bar showing the available keys during exploration, while waiting for a command and in the potion menu (`ctl.bar.*`).
- Test suite for manual control, including fuzzing (`manual_test.go`).

### Changed
- `step()` refactored: tile, container, combat and transition handling moved into `moveTo()`, shared by the autopilot and manual control.
- In manual mode the automatic retreat, fleeing and in-combat potion drinking are disabled; town visits stay automatic.

## [2.8.5] - 2026-10-03

### Fixed
- Memorial Book and Church (`LostInAbyss`): heroes left behind in the Abyss during a retreat could still be resurrected in the Church. They are now permanently lost and replaced by Guild recruits.
- Uniform card grid: all 6 cards (5 heroes and the party banner) now have exactly the same width and height in every state and layout (`normalizeLines()`, `cardContentHeight()`).
- Border alignment: fixed border collisions, padding overlaps and line wrapping in the map box, sidebar and chronicle log in the Landscape, Portrait and PortraitWide layouts.

## [2.8.1] - 2026-10-02

### Changed
- Synchronized release workflows and metadata after v2.8.0.
- Tag versions are embedded into builds via `-ldflags`; `sha256sums.txt` is generated automatically.

## [2.8.0] - 2026-10-02

### Added
- Autopilot tactics screen (`T`): thresholds for retreating, fleeing and drinking potions, plus toggles for trapped chests, blood altars and risky bargains.
- Room events (`?`): campfires, wandering merchants, whispering idols, locked vaults (picked by rogues or forced by bruisers) and ambushes.
- Persistent save (`save.json` in the user config directory): town legacy, capital investments, language and tactics.
- CLI flags: `-seed N`, `-no-save`, `-reset-save`, `-version`, `-report`.
- Capital market: automatic purchase of missing basic equipment and gear upgrades during town visits.
- Smart loot priority: a party with unequipped heroes avoids monsters and goes for chests and containers first.
- Tests: RU/EN dictionary parity, room events, tactics bounds, save round-trip.

### Fixed
- Stale tick chains (`TickGen`) accumulating after closing windows or unpausing, which doubled the simulation speed.
- Accidental resets with `R` in the Armory and Codex; restart is now limited to the Defeat and Glory screens.
- Missing localization key `combat.log.cleric_heal` and missing damage argument in `combat.log.mage_storm`.
- Kill credit centralized in `onMonsterKilled`: gold, experience, contracts and statistics are now credited for kills with any skill of all 10 classes.
- The Tavern level was displayed as the Tannery level.
