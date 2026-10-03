Changelog
All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog,
and this project adheres to Semantic Versioning.

[2.8.6] - 2026-10-03
Added
Manual control mode (M, flag -manual): full direct party movement using arrow keys.

Combat commands:

A — basic strike without mana cost against the selected target.

Space — guaranteed class skill activation bypassing percentage chance gates.

D — defensive guard stance (activates block, absorbs damage, regains +3 MP, relieves stress).

P — potion belt menu (1–9) with ally targeting via ←/→.

←/→ — manual focus targeting among living monsters (active focus highlighted with ▶).

F — manual retreat/flee attempt from active encounters.

Dungeon survival actions:

C — camp making (+8 HP/MP, stress relief, 25-step cooldown, proximity monster check).

X — voluntary town return when standing on the dungeon exit tile (<).

Context controls bar: dynamic bottom hints displaying available hotkeys during exploration, turn waiting, and potion selection (ctl.bar.*).

Test suite: comprehensive fuzzing and unit tests for player decisions and movement in manual_test.go.

Changed
step() cycle refactored: tile, container, combat, and transition processing extracted into a unified moveTo() pipeline shared by autopilot and manual control.

Autopilot constraints: in manual mode, automatic retreat, fleeing, and in-combat potion drinking are disabled; town visits remain automated.

[2.8.5] - 2026-10-03
Fixed
Memorial Book & Church sync (LostInAbyss): resolved bug where heroes left behind in the Abyss during retreat could still be resurrected in the Temple; abandoned heroes are now permanently lost and replaced by Guild recruits.

Uniform party cards grid: standardized all 6 cards (5 heroes + party banner) to exact identical width and height across all states and layouts using normalizeLines() and cardContentHeight().

Border alignment: fixed border collisions, padding overlaps, and line wrapping in map box, sidebar, and chronicler logs across Landscape, Portrait, and PortraitWide layouts.

[2.8.1] - 2026-10-02
Changed
Synchronized release automation workflows and metadata after the v2.8.0 release.

Embedded tag versions into builds via -ldflags and added automated sha256sums.txt generation.

[2.8.0] - 2026-10-02
Added
Autopilot tactics overlay (T): customizable thresholds for retreating, fleeing, potion drinking, and avoidance toggles for trapped chests, blood altars, and risky encounters.

Room events (?): dynamic tiles featuring traveler campfires, wandering alchemical merchants, whispering idols, locked vaults (lockpicked by rogues or forced by bruisers), and ambushes.

State persistence: cross-session save system storing town legacy, capital investments, language, and tactics in the user configuration directory (save.json).

CLI flags: -seed N, -no-save, -reset-save, -version, and -report.

Capital market trading: automatic procurement of missing basic equipment and gear upgrades during town visits.

Smart loot priority: unequipped heroes actively navigate around monsters to prioritize chests and equipment containers.

Test suite expansion: parity tests for RU/EN dictionary alignment, room events, tactic bounds, and persistence roundtrips.

Fixed
Stale tick accumulation (TickGen) after closing UI windows or unpausing that caused unintended simulation speed doubling.

Prevented accidental game resets via R inside Armory and Codex screens (restricted exclusively to Defeat and Glory screens).

Fixed missing localization key combat.log.cleric_heal and missing damage argument in combat.log.mage_storm.

Centralized kill credit handling in onMonsterKilled: gold, experience, contracts, and stats now credit properly on all lethal skill types across all 10 classes.

Fixed UI defect where Tavern level displayed Tannery level.

Aligned format verb counts across all Russian and English localization string pairs.

Corrected Blood Altar text (+3 Atk granted instead of outdated +2 Atk).

Fixed telemetry report average_atk calculation across multiple runs.

Changed
Baseline Go version upgraded to 1.24.

Localization dictionaries decoupled into separate modular files (i18n_ru.go and i18n_en.go).

Cleaned up .gitignore and added official MIT License (LICENSE).

Rewrote README.md with multi-platform binary support and build verification steps.

[2.7.0] - 2026-09-28
Added
Militia meta-progression (militia.go): 5-tier scaling system for free Guild recruits (Regular, Sergeant, Veteran, Elite, Capital Guard) funded by cumulative Capital investments (TotalInvested).

Biomes & floor scaling: 5 cyclical procedural biomes (Rotting Catacombs, Flooded Grottos, Ashen Deeps, Crystal Labyrinth, Throne of the Void) with environmental hazards and decennial boss encounters (Ash Dragon on floor 10).

Guild contracts: automated multi-type quest system (monster hunts, exploration, chest gathering, relic recovery, blood pacts).

Legendary relics: relic socketing system (Compass of Greed, Martyr's Crown, Holy Grail) providing squad-wide passive modifiers.

Telemetry engine: headless deterministic simulation runner (sim_test.go) and JSON report collector (-report).

[2.6.0] - 2026-09-20
Added
Capital hub cycle (stepTown): multi-phase automated town visit routine (Market, Magistrate, Temple, Tavern, Guild, Blacksmith, Tannery, Alchemist) with dynamic budget allocation (AllocateBudget).

Gear equipment slots: expanded loadout system with 4 individual item slots (Weapon, Head, Chest, Legs) categorized by material types (metals, woods, leathers, cloths) and rarity tiers.

Prefix & suffix affix generator: elemental enchantments (Fire, Poison, Frost, Lightning) and passive suffixes (Leech, Fury, Meditation, Titan).

Workshop upgrades: Blacksmith weapon/plate tempering and Tannery bag tailoring/leather crafting (up to +6).

Alchemical laboratory: 5 mutation types (Chimera, Fury, Titan, Aether, Bastion) with algorithmic hero-vulnerability targeting.

Potion belt system: expandable potion capacity (MaxPotionSlots) governed by Tannery progression.

Strategic evacuation: casualty carrying capacity on flee attempts (attemptFlee), allowing survivors to rescue fallen bodies for Temple resurrection.

Dynamic titles: accomplishment titles awarded dynamically for combat feats (blocks, crits, kills, clutch escapes).

[2.5.0] - 2026-09-12
Added
Extended hero roster (10 classes): Tank, Warrior, Rogue, Mage, Cleric, Paladin, Ranger, Monk, Bard, and Warlock with individual aggro weights (AggroWeight) and 3-tier skill kits (major, medium, minor).

Racial heritage (4 races): Human, Elf, Beastman, and Olongr featuring innate stat passives (mana regen, lifesteal, speed, stun immunity).

Darkest Dungeon stress mechanics: 0–200 sanity meter, psychological afflictions (Paranoia, Selfishness, Maniac), clutch virtues, and lethal heart attacks at maximum stress.

Expanded bestiary: 15 distinct monster types spread across procedural environments.

[2.4.3] - 2026-09-05
Added
Revamped combat damage formula: floor-scaled lethal monster attack calculations factoring in defense absorption and mitigations.

Tank guardian mechanics: passive protection triggers for Tanks and Paladins (CanGuard, ApplyDamage) mitigating incoming damage to fragile backline allies.

Monster enrage: progressive +15% per-round enrage damage bonus activating after round 15 to eliminate combat stalemates.

[2.4.0] - 2026-08-28
Added
Turn initiative system: D20 dice roll combined with hero speed and equipment modifiers driving a sorted turn order queue (TurnQueue).

Combat stances: temporary combat flags for guarding, berserk rage, stealth evasion, and auras.

Combat escape mechanics: fleeing chance calculations factoring in floor depths and class agility perks.

[2.1.0] - 2026-08-15
Added
Bubble Tea & Lipgloss migration: terminal UI rebuilt on [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) and lipgloss.

Modular interface layout: colorized tactical view featuring responsive dungeon viewport, scout status sidebar, hero status cards, and real-time combat logs.

Telemetry & run end summary: dedicated Memorial Book and end-of-run achievement statistics screen.

[1.0.0] - 2026-08-01
Added
Initial release of Dungeon Crawler Console Auto Game (DCCAG).

Terminal console auto-battler engine written in Go.

Procedural room and corridor dungeon generator with fog-of-war exploration.

4 foundational hero classes (Warrior, Mage, Rogue, Cleric) with automated basic turn progression.

Turn-based melee combat with basic monster packs.
- Tagged releases build Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 binaries, embed the tag version, and publish `sha256sums.txt`.
