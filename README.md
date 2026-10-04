# Dungeon Crawler Console Auto Game (DCCAG)

A grim tactical text-based dungeon crawler featuring autonomous squad mechanics and runtime internationalization, built with Go and Bubble Tea.

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Release](https://img.shields.io/github/v/release/JstexKU/dccag?color=orange)](https://github.com/JstexKU/dccag/releases)

---

## 🎮 About The Project

**DCCAG** is an autonomous console roguelike. Each expedition you get a party of 5 heroes picked at random from **10 classes** (Tank, Warrior, Rogue, Mage, Cleric, Paladin, Ranger, Monk, Bard, Warlock) and 4 races (Human, Elf, Beastman, Olongr). The party explores infinite procedural floors, fights monster packs, manages stress and psychological afflictions, trades in the Capital and crafts gear, all inside a terminal UI. You can set its **tactics** and watch, take **manual control**, or create **your own leader hero** who leads the party.

### ✨ Key Features
* **Your Own Leader Hero** (`C` in the main menu): choose a name, gender, race, class and a *calling*, and spend 6 free points on health, mana, attack, defense and speed. The preview shows the final stats before you commit. The hero is saved to `save.json` and leads every new expedition (`★` on the card); the other four heroes are random and never share the leader's class. The leader is never left behind in the Abyss, the Church revives them for free and the Guild never replaces them, but their fall terrifies the party. Callings: *Inspirer* (allies take 15% less stress), *Strategist* (+10% gold), *Mentor* (+10% experience for the party). `Ctrl+R` in the editor rolls a random hero; `X` in the menu (pressed twice) removes the hero.
* **Manual Control** (`M`): Take direct command of the expedition. Move through the labyrinth with arrow keys, make camp (`C`), exit voluntarily (`X`), command heroes during combat (`A` strike, `Space` skill, `D` guard, `P` potions), and select enemy focus targets (←/→). Toggle seamlessly between autopilot and manual control at any time.
* **Autopilot Tactics** (`T`): Fine-tune squad thresholds for emergency retreat, combat flee, potion drinking, and decision rules for trapped chests, blood altars, and room events.
* **Room Events** (`?` on the map): Campfires, wandering alchemical merchants, whispering idols, ambushes, and locked vaults (picked by rogues or smashed open by bruisers).
* **Town Economy & Trading**: Automated visit pipeline (Market, Magistrate, Church, Tavern, Guild, Smithy, Tannery, Alchemist). Heroes proactively purchase missing gear and buy equipment upgrades from merchants.
* **Capital Militia Meta-Progression**: Cumulative municipal investments unlock 5 permanent tiers of free reinforcements (from local watchmen up to Citadel Keepers).
* **Permanent Legacy & Memorial Book**: Town treasury, infrastructure upgrades, tactics, and language settings persist across sessions in `save.json`. Fallen heroes are recorded in the Memorial Book, and unrecovered casualties left behind in the Abyss are permanently lost.
* **Dynamic Bilingual Engine**: switch between Russian (**RU**) and English (**EN**) on the fly with `L`.
* **Darkest Dungeon-Inspired Mechanics**: Sanity meter (0–200), afflictions (Paranoia, Selfishness, Maniac), virtues, and lethal heart attacks.
* **Uniform Normalized UI**: Grid system locking 6 uniform hero and party banner cards with consistent dimensions and borders across Landscape, Portrait, and PortraitWide layouts.
* **Reproducible Runs**: `-seed N` replays the same dungeon, handy for bug reports and balancing.

---

## ⌨️ Controls

| Key | Action |
| :--- | :--- |
| **`C`** | *Main menu:* create or edit your leader hero (`X` twice removes it) |
| **`Space`** / **`Enter`** | Pause / resume autopilot (or start the game from the menu) |
| **`N`** | One step while paused |
| **`M`** | Toggle manual control / autopilot |
| **`T`** | Autopilot tactics (`↑/↓` select, `←/→` change, `T`/`Esc` back) |
| **`F`** | Try to flee from combat |
| **`←↑↓→`** | *Manual:* move the party one tile |
| **`C`** / **`X`** | *Manual:* make camp (limited by a cooldown) / leave through the exit tile `<` |
| **`A`** / **`Space`** / **`D`** / **`P`** | *Manual combat:* plain strike / class skill / guard / potion (`1`-`9` pick, `←/→` choose who gets it) |
| **`←`** / **`→`** | *Manual combat:* choose the target (marked with `▶`) |
| **`E`** | Party Armory & mutations |
| **`I`** | Codex (`Tab` / `Shift+Tab` or `1`-`7` to switch tabs) |
| **`S`** | Expedition statistics & Memorial Book |
| **`R`** | Restart the expedition (only from the Defeat and Glory screens) |
| **`1` / `2`** | Normal / fast speed (outside the Codex) |
| **`+` / `-`** | Fine speed control |
| **`↑` / `↓`**, **`PgUp` / `PgDn`** | Scroll logs and windows |
| **`L`** | Switch language |
| **`Q`** / **`Ctrl+C`** | Quit (progress is saved) |

### Command-line flags

| Flag | Purpose |
| :--- | :--- |
| `-manual` | Start in manual control mode |
| `-seed N` | Deterministic run with the given seed (`0` = random) |
| `-report` | Collect telemetry and write `dccag_report.json` |
| `-no-save` | Neither read nor write the save file |
| `-reset-save` | Delete the save file and exit |
| `-version` | Print the version and exit |

The save file lives in your user config directory (`~/.config/dccag/save.json` on Linux, `%AppData%\\dccag\\save.json` on Windows, `~/Library/Application Support/dccag/save.json` on macOS).

---

## 🚀 Getting Started

### Option 1: Pre-built binaries (no Go required)
Download the binary for your system from **[Releases](https://github.com/JstexKU/dccag/releases)**.

The current release workflow publishes these targets:

* **Linux**: `dccag-linux-amd64`, `dccag-linux-arm64`
* **macOS**: `dccag-darwin-amd64`, `dccag-darwin-arm64`
* **Windows**: `dccag-windows-amd64.exe`

Each release also contains **`sha256sums.txt`** for verifying the downloaded binaries:

```bash
sha256sum -c sha256sums.txt
chmod +x dccag-<os>-<arch>
./dccag-<os>-<arch>
```

On Windows, verify the checksum with `Get-FileHash` in PowerShell, then run the `.exe`.

### Option 2: Run from source

Requires **Go 1.24+**.

```bash
git clone https://github.com/JstexKU/dccag.git
cd dccag
go run .
```

Or use `./install.sh` to build for your platform (or all of them).

---

## 🧪 Development

```bash
go vet ./...
go test ./...
gofmt -l .        # should print nothing
```

Tests cover dictionary parity (RU/EN keys and format arguments), the autopilot tactics, saving and loading, room events, tick-chain handling and a headless seeded simulation.

### 📂 Project Structure

**Entry point and loop**
* `main.go`: version, CLI flags, `main()`.
* `model.go`: `Model`, game states, town phases, model construction and restart.
* `update.go`: Bubble Tea `Update`, key handling, tick chain.
* `telemetry.go`: the `-report` telemetry collector.
* `save.go`, `rng.go`: persistence (legacy, language, tactics) and the single random source (`-seed`).

**Data types**
* `types_hero.go`: races, classes, afflictions, heroes and their stats.
* `types_items.go`: potions, mutations, equipment, materials, bag upgrades.
* `types_monster.go`: monsters, packs, turn order, active combat.
* `types_world.go`: tiles, biomes, quests, run statistics, town legacy and budget, relics.
* `styles.go`: Lipgloss styles.

**Dungeon**
* `dungeon_gen.go`: floor size, procedural generation, fog of war.
* `monsters.go`: monster pack spawning and scaling.
* `loot.go`: item generation.
* `tiles.go`: altars, fountains, trapped chests, relics.
* `events.go`: room events (`?` tiles).
* `quests.go`: contracts and quest progress.
* `step.go`: the main game step and movement onto a tile.
* `biomes.go`: biome configuration.

**Leader hero**
* `hero_blueprint.go`: hero blueprint, point allocation, stat preview, starting party assembly.
* `creator.go`: the creation screen, its input handling and the main-menu hero lines.
* `leader.go`: leader rules (rescue, Church, Guild, shock on fall).

**Party and decisions**
* `party.go`: hero creation, stress, potions, experience, equipment, fallen heroes.
* `titles.go`: glory titles.
* `autopilot.go`: retreat and healing decisions, pathfinding to targets.
* `tactics.go`: autopilot settings and their screen.
* `manual.go`: manual control (movement, hero commands, camp, key handling).

**Combat**
* `combat.go`: targeting, damage, combat start, kill credit.
* `combat_turn.go`: one combat turn (skills of all classes, monster turns).
* `combat_flee.go`: fleeing and emergency retreat.

**Town**
* `town.go`: the town visit pipeline (`stepTown`).
* `economy.go`: establishments, budget, potion and mutation prices, market.
* `militia.go`: Guild militia tiers.

**Interface**
* `ui_layout.go`, `ui_text.go`: layout detection and text helpers.
* `ui_view.go`: `View`, Landscape / Portrait layouts, log box, controls bar.
* `ui_cards.go`: hero cards and the party banner.
* `ui_map.go`, `ui_town.go`, `ui_screens.go`: map, town hub, menu / statistics / armory screens.
* `codex.go`: the Codex tabs.
* `i18n.go`, `i18n_ru.go`, `i18n_en.go`: translation lookup and the RU / EN dictionaries.

### 📜 License
Distributed under the MIT License. See [LICENSE](LICENSE).

---
