package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================
// Стили Кодекса (единый источник правды)
// ============================================================

var (
	codexHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	codexSecStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	codexSubStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	codexMobStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	codexBossStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	codexHpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	codexAtkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	codexDefStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	codexSpdStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	codexTierStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
	codexItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
	codexNoteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	codexValStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	codexTabActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Reverse(true)
	codexTabIdle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

const codexArrow = "→"
const codexTabCount = 7

func codexTabNames(lang Language) []string {
	if lang == LangEN {
		return []string{"World", "Classes", "Bestiary", "Affixes", "Crafts", "Alchemy", "Titles"}
	}
	return []string{"Мир", "Классы", "Бестиарий", "Аффиксы", "Ремёсла", "Алхимия", "Титулы"}
}

// ============================================================
// Таб-бар
// ============================================================

func (m Model) renderCodexTabBar(innerW int) string {
	names := codexTabNames(m.Lang)

	var parts []string
	for i, name := range names {
		label := fmt.Sprintf(" %d.%s ", i+1, name)
		if i == m.CodexTab {
			parts = append(parts, codexTabActive.Render(label))
		} else {
			parts = append(parts, codexTabIdle.Render(label))
		}
	}
	joined := strings.Join(parts, " ")
	return shortenItemName(joined, innerW)
}

// ============================================================
// Вкладка 1: Мир (биомы + расы)
// ============================================================

func (m Model) renderCodexTabWorld() string {
	var sb strings.Builder

	// Биомы
	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.1")) + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "codex.biome.1.title")), T(m.Lang, "codex.biome.1.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "codex.biome.2.title")), T(m.Lang, "codex.biome.2.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "codex.biome.3.title")), T(m.Lang, "codex.biome.3.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "codex.biome.4.title")), T(m.Lang, "codex.biome.4.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n\n", codexSubStyle.Render(T(m.Lang, "codex.biome.5.title")), T(m.Lang, "codex.biome.5.desc")))

	// Расы
	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.races")) + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "race.human.name")), T(m.Lang, "codex.race.human.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "race.elf.name")), T(m.Lang, "codex.race.elf.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "race.beastman.name")), T(m.Lang, "codex.race.beastman.desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexSubStyle.Render(T(m.Lang, "race.olongr.name")), T(m.Lang, "codex.race.olongr.desc")))

	return sb.String()
}

// ============================================================
// Вкладка 2: Классы (10 классов × 3 умения + прокачка)
// ============================================================

func (m Model) renderCodexTabClasses() string {
	var sb strings.Builder

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.spells")) + "\n")
	classes := []struct {
		nameKey string
		prefix  string
	}{
		{"class.tank.name", "tank"},
		{"class.paladin.name", "pala"},
		{"class.warrior.name", "warr"},
		{"class.monk.name", "monk"},
		{"class.rogue.name", "rogue"},
		{"class.ranger.name", "rang"},
		{"class.mage.name", "mage"},
		{"class.warlock.name", "lock"},
		{"class.cleric.name", "cleric"},
		{"class.bard.name", "bard"},
	}
	for _, c := range classes {
		sb.WriteString(fmt.Sprintf(" • %s: %s | %s | %s\n",
			codexSubStyle.Render(T(m.Lang, c.nameKey)),
			T(m.Lang, "codex.spell."+c.prefix+".major"),
			T(m.Lang, "codex.spell."+c.prefix+".med"),
			T(m.Lang, "codex.spell."+c.prefix+".minor")))
	}
	sb.WriteString("\n")

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.5")) + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.1") + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.2") + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.3") + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.4") + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.5") + "\n")
	sb.WriteString(T(m.Lang, "codex.xp.6") + "\n")

	return sb.String()
}

// ============================================================
// Вкладка 3: Бестиарий (15 монстров по биомам)
// ============================================================

func (m Model) renderCodexTabBestiary() string {
	var sb strings.Builder

	fmtMob := func(name string, hp, atk, def, spd int, extra string) string {
		nameStr := codexMobStyle.Render(padRight(name, 16))
		statsStr := fmt.Sprintf("%s%s %s%s %s%s %s%s",
			codexHpStyle.Render("H:"), codexValStyle.Render(fmt.Sprintf("%-3d", hp)),
			codexAtkStyle.Render("A:"), codexValStyle.Render(fmt.Sprintf("%-3d", atk)),
			codexDefStyle.Render("D:"), codexValStyle.Render(fmt.Sprintf("%-2d", def)),
			codexSpdStyle.Render("S:"), codexValStyle.Render(fmt.Sprintf("%-2d", spd)))
		if extra != "" {
			statsStr += " " + extra
		}
		return fmt.Sprintf(" • %s %s\n", nameStr, statsStr)
	}

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.2")) + "\n\n")

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.catacombs_header")) + "\n")
	sb.WriteString(fmtMob(T(m.Lang, "mob.rat"), 15, 6, 0, 8, ""))
	sb.WriteString(fmtMob(T(m.Lang, "mob.goblin"), 19, 8, 1, 8, ""))
	sb.WriteString(fmtMob(T(m.Lang, "mob.skeleton"), 25, 10, 3, 8, subtleStyle.Render(T(m.Lang, "codex.mob_block"))))
	sb.WriteString("\n")

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.grotto_header")) + "\n")
	sb.WriteString(fmtMob(T(m.Lang, "mob.slime"), 30, 11, 1, 9, stressStyle.Render(T(m.Lang, "codex.mob_stress_10"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.drowned"), 38, 13, 2, 9, stressStyle.Render(T(m.Lang, "codex.mob_stress_10"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.lizard"), 34, 14, 3, 9, accentStyle.Render(T(m.Lang, "codex.mob_crits"))))
	sb.WriteString("\n")

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.inferno_header")) + "\n")
	sb.WriteString(fmtMob(T(m.Lang, "mob.imp"), 40, 15, 2, 10, fireStyle.Render(T(m.Lang, "codex.mob_scorch"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.orc"), 52, 17, 4, 10, fireStyle.Render(T(m.Lang, "codex.mob_rage"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.salamander"), 46, 18, 3, 10, fireStyle.Render(T(m.Lang, "codex.mob_rage"))))
	sb.WriteString("\n")

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.crystal_header")) + "\n")
	sb.WriteString(fmtMob(T(m.Lang, "mob.gargoyle"), 58, 19, 6, 11, subtleStyle.Render(T(m.Lang, "codex.mob_block"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.golem"), 68, 20, 7, 11, subtleStyle.Render(T(m.Lang, "codex.mob_block"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.phantom"), 50, 22, 2, 11, stressStyle.Render(T(m.Lang, "codex.mob_stress_18"))))
	sb.WriteString("\n")

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.abyss_header")) + "\n")
	sb.WriteString(fmtMob(T(m.Lang, "mob.void_demon"), 75, 24, 5, 12, stressStyle.Render(T(m.Lang, "codex.mob_stress_18"))))
	sb.WriteString(fmtMob(T(m.Lang, "mob.death_knight"), 85, 26, 7, 12, dangerStyle.Render(T(m.Lang, "codex.mob_vamp"))))

	// Боссы — отдельным блоком, с подсветкой имени
	fmtBoss := func(name string, hp, atk, def, spd int, extra string) string {
		nameStr := codexBossStyle.Render(padRight(name, 16))
		statsStr := fmt.Sprintf("%s%s %s%s %s%s %s%s",
			codexHpStyle.Render("H:"), codexValStyle.Render(fmt.Sprintf("%-3d", hp)),
			codexAtkStyle.Render("A:"), codexValStyle.Render(fmt.Sprintf("%-3d", atk)),
			codexDefStyle.Render("D:"), codexValStyle.Render(fmt.Sprintf("%-2d", def)),
			codexSpdStyle.Render("S:"), codexValStyle.Render(fmt.Sprintf("%-2d", spd)))
		if extra != "" {
			statsStr += " " + extra
		}
		return fmt.Sprintf(" • %s %s\n", nameStr, statsStr)
	}
	sb.WriteString("\n")
	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.bosses_header")) + "\n")
	sb.WriteString(fmtBoss(T(m.Lang, "mob.boss_orc"), 90, 16, 5, 10, ""))
	sb.WriteString(fmtBoss(T(m.Lang, "mob.boss_golem"), 130, 18, 9, 10, ""))
	sb.WriteString(fmtBoss(T(m.Lang, "mob.boss_leviathan"), 115, 20, 7, 10, ""))
	sb.WriteString(fmtBoss(T(m.Lang, "mob.boss_dragon"), 300, 30, 10, 12,
		fireStyle.Render(T(m.Lang, "codex.mob_dragon_breath"))+
			" "+codexHpStyle.Render("(300+Lvl*20)")))
	sb.WriteString("\n")

	sb.WriteString(codexNoteStyle.Render(T(m.Lang, "codex.bestiary_note")) + "\n")

	return sb.String()
}

// ============================================================
// Вкладка 4: Аффиксы
// ============================================================

func (m Model) renderCodexTabAffixes() string {
	var sb strings.Builder

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.3")) + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fireStyle.Render(T(m.Lang, "codex.affix.1.title")), codexNoteStyle.Render(T(m.Lang, "codex.affix.1.desc"))))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", stressStyle.Render(T(m.Lang, "codex.affix.2.title")), codexNoteStyle.Render(T(m.Lang, "codex.affix.2.desc"))))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fountStyle.Render(T(m.Lang, "codex.affix.3.title")), codexNoteStyle.Render(T(m.Lang, "codex.affix.3.desc"))))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", subtleStyle.Render(T(m.Lang, "codex.affix.4.title")), codexNoteStyle.Render(T(m.Lang, "codex.affix.4.desc"))))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", dangerStyle.Render(T(m.Lang, "codex.affix.5.title")), codexNoteStyle.Render(T(m.Lang, "codex.affix.5.desc"))))

	return sb.String()
}

// ============================================================
// Вкладка 5: Ремёсла (оружие + броня)
// ============================================================

func (m Model) renderCodexTabCrafts() string {
	var sb strings.Builder

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.4")) + "\n")
	sb.WriteString(codexNoteStyle.Render("   "+T(m.Lang, "codex.forge_note")) + "\n")
	sb.WriteString(codexNoteStyle.Render("   "+T(m.Lang, "codex.tanner_note")) + "\n\n")

	// Оружие — по 1 строке на класс
	weapons := []struct {
		clsNameKey string
		itemPrefix string
		statLabel  string
	}{
		{"class.tank.name", "item.tank.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.paladin.name", "item.paladin.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.warrior.name", "item.warrior.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.monk.name", "item.monk.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.rogue.name", "item.rogue.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.ranger.name", "item.ranger.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.mage.name", "item.mage.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.warlock.name", "item.warlock.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.cleric.name", "item.cleric.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
		{"class.bard.name", "item.bard.weapon", codexAtkStyle.Render(T(m.Lang, "ui.atk_short"))},
	}

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.craft_header_weapons")) + "\n")
	for _, w := range weapons {
		sb.WriteString(fmt.Sprintf(" • %-12s: %s %s %s %s %s %s %s %s %s %s %s %s (%s)\n",
			T(m.Lang, w.clsNameKey),
			codexTierStyle.Render("Т1"), T(m.Lang, w.itemPrefix+".1"), codexArrow,
			codexTierStyle.Render("Т2"), T(m.Lang, w.itemPrefix+".2"), codexArrow,
			codexTierStyle.Render("Т3"), T(m.Lang, w.itemPrefix+".3"), codexArrow,
			codexTierStyle.Render("Т4"), T(m.Lang, w.itemPrefix+".4"),
			w.statLabel))
	}
	sb.WriteString("\n")

	// Броня — 3 строки на класс (chest, head, legs)
	armorClasses := []struct {
		clsNameKey string
		itemPrefix string
		statLabel  string
	}{
		{"class.tank.name", "item.tank", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.paladin.name", "item.paladin", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.warrior.name", "item.warrior", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.monk.name", "item.monk", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.rogue.name", "item.rogue", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.ranger.name", "item.ranger", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.mage.name", "item.mage", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.warlock.name", "item.warlock", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.cleric.name", "item.cleric", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
		{"class.bard.name", "item.bard", codexDefStyle.Render(T(m.Lang, "ui.def_short"))},
	}

	sb.WriteString(codexSubStyle.Render(T(m.Lang, "codex.craft_header_armor")) + "\n")
	for _, a := range armorClasses {
		sb.WriteString(fmt.Sprintf(" • %-12s:\n", T(m.Lang, a.clsNameKey)))
		sb.WriteString(fmt.Sprintf("     %s %s %s %s %s %s %s %s %s %s (%s)\n",
			codexSubStyle.Render(T(m.Lang, "slot.chest")),
			codexTierStyle.Render("Т1"), T(m.Lang, a.itemPrefix+".chest.1"), codexArrow,
			codexTierStyle.Render("Т2"), T(m.Lang, a.itemPrefix+".chest.2"), codexArrow,
			codexTierStyle.Render("Т3"), T(m.Lang, a.itemPrefix+".chest.3"), codexArrow,
			codexTierStyle.Render("Т4"), T(m.Lang, a.itemPrefix+".chest.4"),
			a.statLabel))
		sb.WriteString(fmt.Sprintf("     %s  %s %s %s %s %s %s %s %s %s (%s)\n",
			codexSubStyle.Render(T(m.Lang, "slot.head")),
			codexTierStyle.Render("Т1"), T(m.Lang, a.itemPrefix+".head.1"), codexArrow,
			codexTierStyle.Render("Т2"), T(m.Lang, a.itemPrefix+".head.2"), codexArrow,
			codexTierStyle.Render("Т3"), T(m.Lang, a.itemPrefix+".head.3"), codexArrow,
			codexTierStyle.Render("Т4"), T(m.Lang, a.itemPrefix+".head.4"),
			a.statLabel))
		sb.WriteString(fmt.Sprintf("     %s  %s %s %s %s %s %s %s %s %s (%s)\n",
			codexSubStyle.Render(T(m.Lang, "slot.legs")),
			codexTierStyle.Render("Т1"), T(m.Lang, a.itemPrefix+".legs.1"), codexArrow,
			codexTierStyle.Render("Т2"), T(m.Lang, a.itemPrefix+".legs.2"), codexArrow,
			codexTierStyle.Render("Т3"), T(m.Lang, a.itemPrefix+".legs.3"), codexArrow,
			codexTierStyle.Render("Т4"), T(m.Lang, a.itemPrefix+".legs.4"),
			a.statLabel))
	}

	return sb.String()
}

// ============================================================
// Вкладка 6: Алхимия и Реликвии
// ============================================================

func (m Model) renderCodexTabAlchemy() string {
	var sb strings.Builder

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.6")) + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", accentStyle.Render(T(m.Lang, "mut.chimera.name")), T(m.Lang, "codex.mut_chimera_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fireStyle.Render(T(m.Lang, "mut.fury.name")), T(m.Lang, "codex.mut_fury_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", healStyle.Render(T(m.Lang, "mut.titan.name")), T(m.Lang, "codex.mut_titan_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", fountStyle.Render(T(m.Lang, "mut.aether.name")), T(m.Lang, "codex.mut_aether_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n\n", healStyle.Render(T(m.Lang, "mut.bastion.name")), T(m.Lang, "codex.mut_bastion_desc")))

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.7")) + "\n")
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexItemStyle.Render(T(m.Lang, "relic.greed_compass.name.1")), T(m.Lang, "codex.relic_greed_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexItemStyle.Render(T(m.Lang, "relic.martyr_crown.name.1")), T(m.Lang, "codex.relic_martyr_desc")))
	sb.WriteString(fmt.Sprintf(" • %s: %s\n", codexItemStyle.Render(T(m.Lang, "relic.holy_grail.name.1")), T(m.Lang, "codex.relic_grail_desc")))

	return sb.String()
}

// ============================================================
// Вкладка 7: Титулы (по классам + общие)
// ============================================================

func (m Model) renderCodexTabTitles() string {
	var sb strings.Builder

	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.titles")) + "\n\n")

	titles := []struct {
		clsNameKey string
		entries    []struct{ nameKey, condKey string }
	}{
		{"class.tank.name", []struct{ nameKey, condKey string }{
			{"title.impenetrable", "codex.title_cond.blocks_10"},
			{"title.storm_shield", "codex.title_cond.aggro_8"},
			{"title.the_wall", "codex.title_cond.dmg_taken_75"},
		}},
		{"class.paladin.name", []struct{ nameKey, condKey string }{
			{"title.paladin_redeemer", "codex.title_cond.heals_80"},
			{"title.paladin_bastion", "codex.title_cond.blocks_8"},
			{"title.paladin_crusader", "codex.title_cond.kills_5"},
		}},
		{"class.warrior.name", []struct{ nameKey, condKey string }{
			{"title.blood_blade", "codex.title_cond.kills_12"},
			{"title.executioner", "codex.title_cond.kills_5"},
			{"title.axe", "codex.title_cond.dmg_90"},
			{"title.rank_cleaver", "codex.title_cond.crits_4"},
		}},
		{"class.monk.name", []struct{ nameKey, condKey string }{
			{"title.monk_calm", "codex.title_cond.cc_20"},
			{"title.monk_fist", "codex.title_cond.crits_6"},
			{"title.monk_wind", "codex.title_cond.dmg_85"},
		}},
		{"class.rogue.name", []struct{ nameKey, condKey string }{
			{"title.phantom_strike", "codex.title_cond.crits_6"},
			{"title.shadow", "codex.title_cond.crits_3"},
			{"title.blade", "codex.title_cond.kills_4"},
			{"title.knife_in_the_back", "codex.title_cond.backstabs_5"},
			{"title.deft_hand", "codex.title_cond.loot_3"},
		}},
		{"class.ranger.name", []struct{ nameKey, condKey string }{
			{"title.ranger_trapper", "codex.title_cond.cc_25"},
			{"title.ranger_sniper", "codex.title_cond.kills_6"},
			{"title.ranger_hawkeye", "codex.title_cond.crits_5"},
		}},
		{"class.mage.name", []struct{ nameKey, condKey string }{
			{"title.stormbringer", "codex.title_cond.dmg_150"},
			{"title.cinder", "codex.title_cond.dmg_90"},
			{"title.chains_of_the_void", "codex.title_cond.cc_40"},
			{"title.flash", "codex.title_cond.mana_3"},
		}},
		{"class.warlock.name", []struct{ nameKey, condKey string }{
			{"title.warlock_harvester", "codex.title_cond.dmg_110"},
			{"title.warlock_void", "codex.title_cond.mana_4"},
			{"title.warlock_curser", "codex.title_cond.kills_5"},
		}},
		{"class.cleric.name", []struct{ nameKey, condKey string }{
			{"title.grace", "codex.title_cond.heals_120"},
			{"title.holy", "codex.title_cond.heals_70"},
			{"title.resurrector", "codex.title_cond.revives_3"},
			{"title.purifier", "codex.title_cond.dots_4"},
		}},
		{"class.bard.name", []struct{ nameKey, condKey string }{
			{"title.bard_virtuoso", "codex.title_cond.stress_5"},
			{"title.bard_siren", "codex.title_cond.cc_20"},
			{"title.bard_rhapsodist", "codex.title_cond.dmg_60"},
		}},
	}

	for _, t := range titles {
		sb.WriteString(codexSubStyle.Render("["+T(m.Lang, t.clsNameKey)+"]") + "\n")
		for _, e := range t.entries {
			sb.WriteString(fmt.Sprintf("   • %s — %s\n",
				titleStyle.Render(T(m.Lang, e.nameKey)),
				T(m.Lang, e.condKey)))
		}
		sb.WriteString("\n")
	}

	// Общие титулы
	sb.WriteString(codexSecStyle.Render(T(m.Lang, "codex.sec.common_titles")) + "\n")
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.slayer_of_beasts")), T(m.Lang, "codex.title_cond.boss_kills_3")))
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.dragonslayer")), T(m.Lang, "codex.title_cond.boss_kills_1")))
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.luck_cursed")), T(m.Lang, "codex.title_cond.near_death_6")))
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.immortal")), T(m.Lang, "codex.title_cond.near_death_3")))
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.treasure_seeker")), T(m.Lang, "codex.title_cond.treasures_5")))
	sb.WriteString(fmt.Sprintf("   • %s — %s\n", titleStyle.Render(T(m.Lang, "title.secret_keeper")), T(m.Lang, "codex.title_cond.secrets_4")))

	return sb.String()
}

// ============================================================
// renderInfoBookScreen — координатор Кодекса
// ============================================================

func (m Model) renderInfoBookScreen() string {
	boxW := max(36, m.TermWidth-4)
	innerW := max(32, boxW-4)

	var sb strings.Builder
	sb.WriteString(codexHeaderStyle.Render(T(m.Lang, "codex.header")) + "\n")
	sb.WriteString(m.renderCodexTabBar(innerW) + "\n")
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n\n")

	switch m.CodexTab {
	case 0:
		sb.WriteString(m.renderCodexTabWorld())
	case 1:
		sb.WriteString(m.renderCodexTabClasses())
	case 2:
		sb.WriteString(m.renderCodexTabBestiary())
	case 3:
		sb.WriteString(m.renderCodexTabAffixes())
	case 4:
		sb.WriteString(m.renderCodexTabCrafts())
	case 5:
		sb.WriteString(m.renderCodexTabAlchemy())
	case 6:
		sb.WriteString(m.renderCodexTabTitles())
	}

	sb.WriteString("\n" + subtleStyle.Render(fmt.Sprintf("[Tab/Shift+Tab] %s | [1-7] %s | [↑/↓/PgUp/PgDn] %s | [I/Esc] %s",
		T(m.Lang, "codex.tab.switch"),
		T(m.Lang, "codex.tab.jump"),
		T(m.Lang, "ui.scroll"),
		T(m.Lang, "ui.btn_back"))))

	wrappedText := lipgloss.NewStyle().Width(innerW).Render(sb.String())
	lines := strings.Split(wrappedText, "\n")

	maxVisibleLines := max(8, m.TermHeight-4)
	maxScroll := max(0, len(lines)-maxVisibleLines)
	if m.StatsScroll > maxScroll {
		m.StatsScroll = maxScroll
	}
	if m.StatsScroll < 0 {
		m.StatsScroll = 0
	}

	endIdx := min(len(lines), m.StatsScroll+maxVisibleLines)
	visibleLines := lines[m.StatsScroll:endIdx]

	box := statsBoxStyle.Width(boxW).Render(strings.Join(visibleLines, "\n"))
	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}