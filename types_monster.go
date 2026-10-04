package main

// --- Монстры и Схватки ---
type MonsterAffix string

const (
	AffixNone     MonsterAffix = "none"
	AffixFire     MonsterAffix = "fire"
	AffixPoison   MonsterAffix = "poison"
	AffixFrost    MonsterAffix = "frost"
	AffixStone    MonsterAffix = "stone"
	AffixVampiric MonsterAffix = "vampiric"
)

type MonsterType string

const (
	MobRat         MonsterType = "rat"
	MobGoblin      MonsterType = "goblin"
	MobSkeleton    MonsterType = "skeleton"
	MobSlime       MonsterType = "slime"
	MobDrowned     MonsterType = "drowned"
	MobLizard      MonsterType = "lizard"
	MobImp         MonsterType = "imp"
	MobOrc         MonsterType = "orc"
	MobSalamander  MonsterType = "salamander"
	MobGargoyle    MonsterType = "gargoyle"
	MobGolem       MonsterType = "golem"
	MobPhantom     MonsterType = "phantom"
	MobVoidDemon   MonsterType = "void_demon"
	MobDeathKnight MonsterType = "death_knight"
	MobDragon      MonsterType = "dragon"

	// 5. Затонувший Лес Мертвецов
	MobSproutSkeleton MonsterType = "sprout_skeleton"
	MobDryad          MonsterType = "dryad"
	MobBlightEnt      MonsterType = "blight_ent"

	// 6. Грибные Топи
	MobSporling   MonsterType = "sporling"
	MobTentacle   MonsterType = "tentacle"
	MobToxicBasil MonsterType = "toxic_basil"

	// 7. Забытые Архивы
	MobTomeBook      MonsterType = "tome_book"
	MobScrollMimic   MonsterType = "scroll_mimic"
	MobArchiveKeeper MonsterType = "archive_keeper"

	// 8. Обсидиановые Шахты
	MobObsidianBeetle MonsterType = "obsidian_beetle"
	MobDeepTroll      MonsterType = "deep_troll"
	MobMinerGhoul     MonsterType = "miner_ghoul"

	// 9. Осквернённый Санктуарий
	MobFallenCrusader   MonsterType = "fallen_crusader"
	MobBloodCultist     MonsterType = "blood_cultist"
	MobShadowInquisitor MonsterType = "shadow_inquisitor"

	// 10. Астральный Разлом
	MobAstralWeaver  MonsterType = "astral_weaver"
	MobChronoPhantom MonsterType = "chrono_phantom"
	MobEssenceDevour MonsterType = "essence_devourer"

	// Мини-боссы
	MobMiniBoneblight  MonsterType = "mini_boneblight"
	MobMiniExecutioner MonsterType = "mini_executioner"
	MobMiniColossus    MonsterType = "mini_colossus"
	MobMiniReaver      MonsterType = "mini_reaver"
	MobMiniStalker     MonsterType = "mini_stalker"
)

type Monster struct {
	ID      int
	Type    MonsterType
	NameKey string
	Level   int
	Affix   MonsterAffix
	Glyph   rune
	Color   string
	HP      int
	MaxHP   int
	Atk     int
	Defense int
	Speed   int
	Exp     int
	IsDead  bool
}

type MonsterPack struct {
	Members    []*Monster
	IsBoss     bool
	IsMiniBoss bool
}

func (p *MonsterPack) LivingCount() int {
	cnt := 0
	for _, m := range p.Members {
		if !m.IsDead {
			cnt++
		}
	}
	return cnt
}

func (p *MonsterPack) GetFirstLiving() *Monster {
	for _, m := range p.Members {
		if !m.IsDead {
			return m
		}
	}
	return nil
}

func (p *MonsterPack) GetLowestHPFocus() *Monster {
	var target *Monster
	minHP := 99999
	for _, m := range p.Members {
		if !m.IsDead && m.HP < minHP {
			minHP = m.HP
			target = m
		}
	}
	return target
}

func (p *MonsterPack) GetHighestHPFocus() *Monster {
	var target *Monster
	maxHP := -1
	for _, m := range p.Members {
		if !m.IsDead && m.HP > maxHP {
			maxHP = m.HP
			target = m
		}
	}
	return target
}

type CombatantType int

const (
	CombatantHero CombatantType = iota
	CombatantMonster
)

type TurnOrderEntry struct {
	Type       CombatantType
	HeroRef    *Hero
	MonsterRef *Monster
	Initiative int
}

type ActiveCombat struct {
	Pos          Point
	Pack         *MonsterPack
	TurnQueue    []TurnOrderEntry
	TurnIdx      int
	Round        int
	HasBarrel    bool
	FleeCooldown int

	// Ручное управление (см. manual.go)
	Focus      *Monster  // цель, выбранная игроком
	Cmd        ManualCmd // команда игрока для героя, чей ход сейчас
	PotionMenu bool      // открыт выбор зелья
	PotionTo   *Hero     // кому достанется зелье
}
