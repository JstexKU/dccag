# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed
- **All player-visible text now comes from the dictionaries.** About 180 hardcoded Russian and English strings were moved to `i18n_ru.go` and `i18n_en.go` (159 new keys per language): the whole main menu (full, medium and compact variants), the cards of every Capital building (headers, badges, hints, offer rows, the autopilot banner), the autopilot's town summaries, the town hub status line, the Codex tab names, tier labels and the dragon HP formula, the town signs, the armory and hero card labels (`HP`, `MP`, `ST`, `Atk`, `Def`), and the Church offer descriptions.
- `TVerb` now takes a key: `TVerb(lang, gender, "reach")` reads `verb.reach.m` / `verb.reach.f` from the dictionary. The 27 call sites carried their Russian and English words inline; they now use 22 dictionary verbs.
- Gold spending categories in the telemetry report are stable language-neutral identifiers (`spend.gear`, `spend.forge`, ...) instead of Russian phrases; the readable names live in the dictionaries under the same keys. **The keys of `GoldSpentBreakdown` in `dccag_report.json` changed accordingly.**
- `renderAutoBanner` became a method of `Model` so that it can translate its tag and idle text.
- The Capital's autopilot summaries (sold, invested, stabilized, hired, sharpened, ...) existed only in Russian; they now have English text.

### Fixed
- The medium and compact main menu showed English buttons (`[Q] Quit`, `[L] Language`) to Russian players; the town hub status line (`BAG ... SMITHY ...`) and the Codex tier labels (`Т1`..`Т4`, Cyrillic in English too) did not follow the language.
- Church offers printed the raw affliction identifier (`paranoid`); the translated name is used now. The Guild summary printed the raw class identifier (`tank`) the same way.
- The armory used the English `Atk`/`Def` in the Russian interface instead of the dictionary words.

### Added
- Tests that keep the code free of hardcoded text (`i18n_hardcode_test.go`): no Cyrillic literals outside the dictionaries, no inline language branches, every `TVerb` key exists for both genders in both languages, sign texts fit the signboards, the menu and every Capital card contain no foreign-language leftovers, and the autopilot's town history and log are fully English in the English interface.

### Not changed on purpose
- Command-line help and startup errors in `main.go`, the telemetry console message, and the key hints inside the bars (`[Enter]`, `[Esc]`, ...) stay as they are: they are developer-facing or name physical keys.
- The unit suffix `G` (gold), the `HP`/`MP`/`ST` values and the product name stay identical in both languages, but they now live in the dictionaries (`ui.hp_short`, `menu.title`, ...) and can be localized there.
[Unreleased]


## [3.3.2]

### 📜 Expanded Contract System (19 Quest Types)
* Added 12 new Guild contract types:
  * **Exploration:** "Trap Disarmer" (disarming traps), "Fountain Grace", "Abyss Cartography" (90%+ explored), "Gold Rush", "Secret Paths", "Full Clearance".
  * **Combat Challenges:** "Lord Slayer" (biome bosses), "Mutant Hunt" (affixed monsters), "Firestorm Havoc" (tar barrel detonations), "Flawless Vanguard" (victories without knockdowns), "Precision Strikes" (critical hits), "Unyielding Bulwark" (blocks and defensive stances).
* **Smart Filtering:** Contracts dynamically adjust to current dungeon depth, boss floors, and active biome mechanics.

### ⚖️ Market UI Improvements
* Implemented a dynamic sliding **Viewport** for backpack inventory: long item lists no longer overflow or break the shop card.
* Cursor focus automatically tracks the selected item in both manual mode (`↑`/`↓`, `PgUp`/`PgDown`) and during autopilot transaction previews.
* Adaptive scroll indicators (`▲ N more above...` / `▼ N more below...`) tailored to terminal window height.

### 🌐 Localization

## [3.3.1]

Added
7 Playable Races with Racial Perks and Class Synergies:
Complete race system integrated into character generation and party creation: Human, Elf, Beastman, Olongr, Dwarf, Demonkin, and Dragonkin.
Unique traits: Human exp boost & stress relief on level-up; Elf speed, crit & MP regeneration; Beastman lifesteal; Olongr health, defense & stun immunity; Dwarf defense, virtue bonus & reinforced block (55%); Demonkin low-health enrage & mana-on-kill; Dragonkin fire immunity & fire retaliation damage.
140 unique hero names (10 male and 10 female names for each of the 7 races).
10 Unique Legendary Relics with Active Gameplay Mechanics:
Expanded relic system from 3 to 10 relics with 3 leveling tiers each (types_world.go, tiles.go, combat_turn.go, combat.go, step.go, manual.go).
Compass of Greed (RelicGreedCompass): Up to +45% gold multiplier, increased monster damage.
Martyr's Crown (RelicMartyrCrown): Permanent attack buff granted to surviving party members upon ally death.
Holy Grail (RelicHolyGrail): Passive squad-wide stress mitigation.
Pioneer's Eye (RelicPioneerEye): Expanded field-of-view radius through fog of war and automatic hidden wall detection.
Quicksilver Vial (RelicQuicksilverVial): Squad speed bonus and guaranteed first-turn initiative in combat.
Obsidian Anvil (RelicObsidianAnvil): Gear tempering granting bonus attack and defense for equipped T2+ items.
Whispering Skull (RelicWhisperingSkull): Lowers party critical strike threshold; incoming enemy crits inflict psychological stress.
Phoenix Core (RelicPhoenixCore): One-time miraculous death prevention per dungeon trip, reviving downed hero at 35% HP with stress cleansed.
Horn of Plenty (RelicHornOfPlenty): Amplified resource recovery (+25% to +45%) from healing potions and camp resting.
Inquisitor's Seal (RelicInquisitorSeal): Pure bane damage bonus and 50% armor penetration against undead, demons, mini-bosses, and dragon bosses.
Belarusian Localization (BE):
Full Belarusian dictionary support added alongside Russian and English with on-the-fly switching via [L].
Complete parity across all menus, town buildings, combat logs, codex entries, race descriptions, and gendered verbs (TVerb).
Keyboard Layout Independence:
Transparent hotkey normalization: gameplay controls and shortcut keys (M, L, T, E, I, Q, A, D, P, C, X, arrows/WASD) function seamlessly regardless of active keyboard layout (QWERTY, JCUKEN/Cyrillic, Belarusian).
Text input fields (such as custom hero naming in the creator) preserve raw character runes.
Codex Tab 6 Expansion:
Full breakdown of all 10 relics with mechanics descriptions and color-coded rarity tiers.
Fixed
Resolved duplicate revealFog declaration across step.go and dungeon_gen.go.
Fixed undefined boss/elite attribute access in combat_turn.go by verifying pack properties (IsMiniBoss) and dragon monster types.
Changed
All player-visible text now comes from the dictionaries. About 180 hardcoded strings moved to dictionary files across all supported languages (RU, EN, BE).
TVerb accepts dynamic keys for gender-aware conjugation across all localized languages.
Telemetry spend categories use stable, language-neutral keys (spend.*).

## [3.3.0]

### Added
- **The Capital as a map.** A generated 90×41 town with nine buildings (Market, Magistrate, Church, Tavern, Guild, Forge, Tannery, Alchemist and the Dungeon Gate in the middle), cobbled streets, a stone wall, trees, flowers and fountains. Buildings have coloured roofs with an emblem, windows, a door and two signboards; the signs show the building's name and its level (or the deepest floor at the Gate) in the player's language, and the autopilot's current destination is highlighted. The town is mirrored randomly and the buildings shift a little, so every run looks different. Only streets are walkable.
- **Autopilot walks the town** in the order of the old pipeline. At a door the party enters, the building's card appears for a moment, the building does its work, the result stays on the card for one more tick, and the party steps out into the street and walks on. Buildings with nothing to do are skipped (a healthy party with no gold visits only the Market and the Gate); the Market and the Gate are always visited. Walking is faster than the work inside buildings.
- **Manual walking**: arrow keys move through the streets, stepping on a door opens the building's card, `Esc` goes back out (at the Gate, `Esc` first returns from the floor list). Old building handlers that used to advance to the next building now simply send the party out into the street.
- **Two entrances to the dungeon** (Gate card): *Continue the descent* (to the deepest floor reached) and *Pick a cleared floor* — the chosen floor is restored and refilled with small packs (up to 2 monsters, no mini-bosses) of the floor's own level, spaced apart and kept away from the entrance. A floor's list shows its biome, recommended level, living packs and explored percent.
- **Autopilot training logic**: recommended level of a floor is about `floor/2`. A fighter is *weak* when he lags behind the deepest floor's recommended level by the configured gap (default 2). The autopilot trains when two or more fighters are weak, or one is weak and the previous two trips ended in a retreat without progress (new *failure streak* counter); the floor is chosen from the weakest fighter's level. On a training floor the autopilot never takes the stairs and heads for the exit once the weak fighters have caught up or the floor is cleared (new retreat reason *training done*).
- **Tactics**: two new rows — *Send recruits to train* (on/off) and *Weak when behind by N levels* (1–6). Saves from earlier versions get the defaults.
- **A fully custom party.** The editor has five slots (`PgUp`/`PgDn`, or the first row). Slot 1 is the leader, slots 2–5 are companions without a calling. Slots without a name are filled with random fighters of classes not yet in the party. `Ctrl+T` rolls a random party, `Del` clears a slot. The save format has an optional `companions` list; the whole party survives restarts.
- Tests for the town map generator and rendering, walking and visits, skipping rules, the Gate card, the departure plan, repopulation, training runs, the new tactics rows and the custom party.

### Changed
- Every way back into town (exit tile, emergency retreat, flight, manual leave) now goes through `enterTown`, which also saves the floor. Before, emergency retreats and flights did not, so the explored state of that floor was lost.
- `MaxFloor` (deepest floor reached) is tracked separately from `Floor`; the town economy and recruits keep using the progress floor even after a training trip.
- `Esc` hints on the building cards now say *Leave* instead of *Next*.

### Fixed
- The Tavern's barn no longer lowers the health and mana of fighters who are already above 40%.

## [3.2.0]

### Added
- **Large-scale procedural dungeons**: floor dimensions increased up to 220x90 with 25–55+ rooms per level.
- Room topology variety: rectangular halls, organic oval grottos, and pillared chambers.
- Branched path network: minimum spanning tree (MST) corridor graph with 25% loopback corridors for flanking and shortcuts.
- **Floor state persistence**: visited floors, explored fog of war, opened chests, and remaining monster packs are preserved across floor transitions and town trips (`VisitedFloors`).
- **Fog of war exploration AI**: autopilot now explores unknown territory via frontier exploration and can no longer see through fog of war.

## [3.1.0]

### Added
- **Interactive Town Establishments:** Dedicated interactive shop/service cards for all 8 town locations (`ui_town_shops.go`, `town_services.go`, `town_manual.go`):
  - **Market (`TownPhaseSellLoot`):** Dual-section view for selling inventory loot (single or bulk with `Space`) and browsing upgrade/missing equipment offers (`Tab` switches sections).
  - **Magistrate (`TownPhaseMagistrate`):** Capital project management for upgrading Forge, Tannery, Church, and Tavern levels, plus treasury grants.
  - **Church (`TownPhaseChurch`):** Direct triage and cleansing interface for downed allies and afflicted/stressed party members.
  - **Tavern (`TownPhaseTavern`):** Room choice between luxury quarters (full recovery) and barn stay (budget partial rest).
  - **Guild (`TownPhaseGuild`):** Recruit replacements for fallen party members with veterans or free militia volunteers.
  - **Forge & Tannery (`TownPhaseSmithy`, `TownPhaseTannery`):** Item sharpening and armor tailoring with dynamic tier restrictions, plus backpack capacity upgrades at the Tannery.
  - **Alchemist (`TownPhaseAlchemist`):** Targeted mutation elixirs and potion belt restocking.
- **Autopilot Action Feedback:** In auto-mode, town shop cards visually reflect the squad's decisions: the cursor `▶` tracks the selected offer, rows are highlighted with green status badges (`✓ КУПЛЕНО`, `✓ УЛУЧШЕНО`, `✓ ИСЦЕЛЕНО`, etc.), and a bottom banner displays the autopilot rationale (`ActionSummary`).
- **Main Menu Mode Toggle:** Hotkey `[M]` in the main menu (`StateMenu`) toggles between fully autonomous (`⚙️ Auto-play`) and direct control (`🎮 Manual control`) before embarking, resetting the countdown timer on change.
- Comprehensive test coverage for manual town interactions, navigation, investments, upgrades, and menu mode switching (`town_manual_test.go`, `creator_test.go`).

### Changed
- Large source files were split by responsibility (`types_*.go`, `ui_*.go`, `combat_*.go`, `party.go`, `autopilot.go`, `step.go`, `tiles.go`, `loot.go`, `economy.go`, `update.go`, `model.go` and others). Declarations were moved without changes; see the structure in the README.
- The interface takes the version from `version` through `displayVersion()` instead of hard-coded strings; release builds that pass a tag with the `v` prefix no longer show `vv2.9.0`.
- Town visit loop refactored: step execution waits for player confirmation/navigation in manual mode while maintaining instant evaluation in auto-mode.

### Removed
- Unused code: `Hero.Verb`, `blankLines`, `wallStyle`, `floorStyle`, `ArmorNone`, `SuffNone`.
- 45 unused dictionary keys in RU and EN (`ui.btn_*`, `codex.base_*`, `status.*` and others).
- A stray copy of `main.go` that was committed under the file name `\`.
- Citation markers (``) left in the README.

### Known issues
- The cleric title `title.resurrector` can never be earned: `Feats.Revives` is not incremented anywhere (`Hero.AddRevive` is never called); revival happens in the Church, not through a Cleric action.

## [3.0.0]

### Added
- **Leader hero creation** (`C` in the main menu). The editor lets the player enter a name, choose gender, race, class and a calling, and distribute 6 free points (health, mana, attack, defense, speed; `+5/+5/+1/+1/+1` per point, with per-stat caps). A live preview shows the final stats, the class skill and the combat role. `Ctrl+R` rolls a random hero, `Ctrl+L` switches the language, `Esc` goes back.
- The hero is stored in `save.json` (`hero` section, optional, validated and sanitized on load) and rebuilt at level 1 for every new expedition, including restarts after a defeat. The other four party members are random and never share the leader's class. `X` pressed twice in the main menu removes the hero.
- **Leader rules** (`leader.go`): the leader is marked with `★`; is always carried out first and never lost in the Abyss; is revived by the Church for free; is never replaced by the Guild; only loses consciousness (instead of dying) from altars and trapped chests; and their fall in battle gives every ally +14 stress.
- **Callings:** *Inspirer* (allies take 15% less stress), *Strategist* (+10% gold from kills, chests and events), *Mentor* (+10% experience for the party). Bonuses work only while the leader is on their feet.
- Tests for the blueprint, the preview/hero consistency, the editor input (including pasted text and hotkeys-as-text), save round-trips, leader rules and a long leader simulation.

### Changed
- `createHero` was split into `baseStatsFor` (class table) and `buildHero` (assembly); the numbers are unchanged and covered by a test.
- Main-menu timer ticks carry a generation number (`MenuGen`), so a stale tick can no longer start the expedition from inside the editor.
- `persistState` takes the hero blueprint as an extra argument.
- Removed the `[I] Codex | [E] Armory | [S] Stats` hints from the main menu: those keys never worked there.

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
