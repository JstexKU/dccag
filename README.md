# Dungeon Crawler Console Auto Game (DCCAG)

A grim tactical text-based dungeon crawler featuring autonomous squad mechanics and runtime internationalization, built with Go and Bubble Tea.

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Release](https://img.shields.io/github/v/release/JstexKU/dccag?color=orange)](https://github.com/JstexKU/dccag/releases)

---

## 🎮 About The Project

**DCCAG** is an autonomous console roguelike. Each expedition you get a party of 5 heroes picked at random from **10 classes** (Tank, Warrior, Rogue, Mage, Cleric, Paladin, Ranger, Monk, Bard, Warlock) and 4 races (Human, Elf, Beastman, Olongr). The party explores infinite procedural floors, fights monster packs, manages stress and psychological afflictions, trades in the Capital and crafts gear, all inside a terminal UI. You don't steer the party: you set its **tactics** and watch.

### ✨ Key Features
* **Autopilot Tactics** (`T`): decide when the party retreats, flees, drinks potions, and whether it touches trapped chests, blood altars and risky bargains.
* **Room Events** (`?` on the map): campfires, wandering merchants, whispering idols, locked vaults (rogues pick them, bruisers smash them) and ambushes.
* **Persistent Legacy**: the Capital's investments, your language and tactics are saved between launches.
* **Dynamic Bilingual Engine**: switch between Russian (**RU**) and English (**EN**) on the fly with `L`.
* **Darkest Dungeon-Inspired Mechanics**: stress, afflictions (Paranoia, Selfishness, Maniac), heart attacks, rare virtues.
* **The Capital & Economy**: Magistrate, Blacksmith, Tannery, Alchemist, Tavern, Church and the Guild militia.
* **Reproducible Runs**: `-seed N` replays the same dungeon, handy for bug reports and balancing.

---

## ⌨️ Controls

| Key | Action |
| :--- | :--- |
| **`Space`** / **`Enter`** | Pause / resume autopilot (or start the game from the menu) |
| **`N`** | One step while paused |
| **`T`** | Autopilot tactics (`↑/↓` select, `←/→` change, `T`/`Esc` back) |
| **`F`** | Try to flee from combat |
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
| `-seed N` | Deterministic run with the given seed (`0` = random) |
| `-report` | Collect telemetry and write `dccag_report.json` |
| `-no-save` | Neither read nor write the save file |
| `-reset-save` | Delete the save file and exit |
| `-version` | Print the version and exit |

The save file lives in your user config directory (`~/.config/dccag/save.json` on Linux, `%AppData%\dccag\save.json` on Windows, `~/Library/Application Support/dccag/save.json` on macOS).

---

## 🚀 Getting Started

### Option 1: Pre-built binaries (no Go required)
Download the binary for your system from **[Releases](https://github.com/JstexKU/dccag/releases)**:

* **Linux**: `dccag-linux-amd64`, `dccag-linux-arm64`, `dccag-linux-386`
* **macOS**: `dccag-darwin-amd64`, `dccag-darwin-arm64`
* **Windows**: `dccag-windows-amd64.exe`

Verify the download against `sha256sums.txt` from the same release, then:

```bash
chmod +x dccag-<os>-<arch>
./dccag-<os>-<arch>
```

On Windows, double-click the `.exe` or run it from PowerShell.

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
* `main.go`: Bubble Tea loop, state machine, tick chain, CLI flags.
* `types.go`: data structures for heroes, items, monsters, stats and Lipgloss styles.
* `tactics.go`: autopilot settings and their screen.
* `events.go`: room events (`?` tiles).
* `save.go`: persistence of legacy, language and tactics.
* `rng.go`: single random source (`-seed`).
* `i18n.go`, `i18n_ru.go`, `i18n_en.go`: translation lookup and the RU / EN dictionaries.
* `dungeon.go`: procedural generation, pathfinding, floor scaling, items.
* `combat.go`: turn queue, aggro, skills, critical hits, fleeing.
* `town.go`, `militia.go`: Capital pipeline, budgeting, crafting, militia tiers.
* `ui.go`, `codex.go`: layouts, hero cards, map, codex tabs.

### 📜 License
Distributed under the MIT License. See [LICENSE](LICENSE).
