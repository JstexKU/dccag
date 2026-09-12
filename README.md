# Dungeon Crawler Console Auto Game (DCCAG)

A grim tactical text-based dungeon crawler featuring autonomous squad mechanics and runtime internationalization, built with Go and Bubble Tea.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Release](https://img.shields.io/badge/release-v2.4.0-orange)](https://github.com/JstexKU/dccag/releases)

---

## 🎮 About The Project

**DCCAG** is an autonomous console roguelike where you command a party of 5 distinct hero classes (Tank, Warrior, Rogue, Mage, and Cleric). Your party explores infinite procedural dungeon floors, battles monster packs, manages stress and psychological afflictions, trades in the Capital, and crafts powerful gear—all inside an immersive terminal user interface.

### ✨ Key Features
* **Autonomous & Step-by-Step Modes**: Let your squad clear rooms automatically or step in for full tactical control at any time.
* **Dynamic Bilingual Engine**: Switch seamlessly between Russian (**RU**) and English (**EN**) on the fly using the `[L]` key.
* **Darkest Dungeon-Inspired Mechanics**: Manage party stress levels, dangerous afflictions (Paranoia, Selfishness, Maniac), heart attacks, and rare virtues.
* **The Capital & Economy**: Invest taxes into the Magistrate, temper weapons and heavy armor at the Blacksmith, craft light gear at the Tannery, brew alchemical mutations, and rest at the Tavern.
* **Deep Progression**: Unlock legendary relics, level up skills, complete dynamic contracts, and build a permanent kingdom legacy for future runs.

---

## ⌨️ Controls & Shortcuts

| Key | Action |
| :--- | :--- |
| **`Space`** / **`Enter`** | Toggle auto-pilot / step-by-step mode (or start game) |
| **`L`** | Switch game language (RU / EN) on the fly |
| **`E`** | Open Party Armory & Alchemical Mutations |
| **`I`** | Open Codex Knowledge Base (use `Tab` & `1-4` to switch tabs) |
| **`S`** | View Expedition Statistics & Memorial Book |
| **`F`** | Attempt to Flee from combat (Rogue & Tank bonuses) |
| **`+` / `-`** | Adjust simulation speed |
| **`↑` / `↓`** | Scroll expedition logs / modals |
| **`Q` / `Ctrl+C`** | Quit game |

---

## 🚀 Getting Started

### Option 1: Download Pre-built Binaries (No Go required)
If you don't have Go installed, you can download ready-to-run binaries for your operating system directly from the **[Releases](https://github.com/JstexKU/dccag/releases)** page.

1. Go to **Releases** and download the archive matching your OS:
   * **Linux**: `dccag-linux-amd64` or `dccag-linux-arm64`
   * **macOS**: `dccag-darwin-amd64` or `dccag-darwin-arm64`
   * **Windows**: `dccag-windows-amd64.exe`

2. **Running the binary:**
   * **Linux / macOS:** Open your terminal in the download directory, make it executable, and run it:
     ```bash
     chmod +x dccag-<os>-<arch>
     ./dccag-<os>-<arch>
     ```
   * **Windows:** Double-click the `dccag-windows-amd64.exe` file, or run it via Command Prompt / PowerShell:
     ```cmd
     dccag-windows-amd64.exe
     ```

---

### Option 2: Run from Source (For Developers)

**Prerequisites:**
* Go 1.21 or higher installed on your system.

1. Clone the repository:
   ```bash
   git clone [https://github.com/JstexKU/dccag.git](https://github.com/JstexKU/dccag.git)
   cd dccag
2. Run the game:
     ```bash
    go run .
  * (Optional) To generate a detailed telemetry JSON report upon exiting, run with the report flag:
      ```bash
    go run . -report
 
### 📂 Project Structure
* main.go — Core Bubble Tea application loop, state machine, and timer ticks.
* types.go — Data structures for Heroes, Items, Monsters, Stats, and Lipgloss styles.
* i18n.go — Centralized bilingual dictionary (RU/EN strings, UI labels, names, and logs).
* dungeon.go — Procedural dungeon generation, pathfinding logic, and floor scaling.
* combat.go — Tactical turn queue, aggro weight formulas, critical hits, and status effects.
* town.go — Capital management pipeline, budgeting, crafting, and tavern rest.
* ui.go — Terminal UI layout rendering, hero status cards, map grid, and codex tabs.

### 📜 License
*** Distributed under the MIT License. See LICENSE for more information. ***
