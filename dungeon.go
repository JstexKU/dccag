package main

import (
	"fmt"
	"math/rand"
)

func generateRelic(level int) PartyRelic {
	relicKeys := []string{"greed_compass", "martyr_crown", "holy_grail"}
	chosen := relicKeys[rand.Intn(len(relicKeys))]

	nameKey := fmt.Sprintf("relic.%s.name.%d", chosen, level)
	descKey := fmt.Sprintf("relic.%s.desc", chosen)

	switch chosen {
	case "greed_compass":
		gMult := 1.15 + (float64(level) * 0.10)
		dmgMod := 1.20 - (float64(level) * 0.02)
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: gMult, EnemyDmgMod: dmgMod, MartyrFury: false,
			DescKey: descKey,
		}
	case "martyr_crown":
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: true,
			DescKey: descKey,
		}
	default:
		res := 10 + (level * 10)
		return PartyRelic{
			NameKey: nameKey, Level: level, GoldMult: 1.0, EnemyDmgMod: 1.0, MartyrFury: false, StressRes: res,
			DescKey: descKey,
		}
	}
}

func generateAutoQuest(curFloor int) AutoQuest {
	qType := QuestType(rand.Intn(6))
	switch qType {
	case QuestHuntMonster:
		var pool []MonsterType
		biomeCycle := (curFloor - 1) % 5
		switch biomeCycle {
		case 0:
			pool = []MonsterType{MobRat, MobGoblin, MobSkeleton}
		case 1:
			pool = []MonsterType{MobSlime, MobDrowned, MobLizard}
		case 2:
			pool = []MonsterType{MobImp, MobOrc, MobSalamander}
		case 3:
			pool = []MonsterType{MobGargoyle, MobGolem, MobPhantom}
		default:
			pool = []MonsterType{MobVoidDemon, MobDeathKnight}
		}
		target := pool[rand.Intn(len(pool))]
		count := rand.Intn(3) + 3
		return AutoQuest{
			Type: QuestHuntMonster, TargetMob: target, TargetCount: count,
			TitleKey: "quest.hunt.title", DescKey: "quest.hunt.desc",
			RewardGold: count * (20 + curFloor*5),
		}
	case QuestOpenChests:
		count := rand.Intn(2) + 2
		return AutoQuest{
			Type: QuestOpenChests, TargetCount: count,
			TitleKey: "quest.chest.title", DescKey: "quest.chest.desc",
			RewardGold: count * (30 + curFloor*3),
		}
	case QuestReachFloor:
		tf := curFloor + 1
		return AutoQuest{
			Type: QuestReachFloor, TargetCount: tf,
			TitleKey: "quest.floor.title", DescKey: "quest.floor.desc",
			RewardGold: tf * 50,
		}
	case QuestFindRelic:
		return AutoQuest{
			Type: QuestFindRelic, TargetCount: 1,
			TitleKey: "quest.relic.title", DescKey: "quest.relic.desc",
			RewardGold: 100 + curFloor*15,
		}
	case QuestEscapeTrap:
		return AutoQuest{
			Type: QuestEscapeTrap, TargetCount: 1,
			TitleKey: "quest.escape.title", DescKey: "quest.escape.desc",
			RewardGold: 110 + curFloor*20,
		}
	default:
		return AutoQuest{
			Type: QuestUseAltar, TargetCount: 1,
			TitleKey: "quest.altar.title", DescKey: "quest.altar.desc",
			RewardGold: 80 + curFloor*15,
		}
	}
}

func (m *Model) currentBagCapacity() int {
	return bagUpgrades[m.BagLevel].Capacity
}

func (m *Model) calculateDungeonSize(floor int) (int, int) {
	w := 70 + (floor-1)*3
	h := 30 + (floor-1)*1
	if w > 130 {
		w = 130
	}
	if h > 55 {
		h = 55
	}
	return w, h
}

func (m *Model) initDungeonForFloor(floor int) {
	w, h := m.calculateDungeonSize(floor)
	m.MapWidth = w
	m.MapHeight = h
	m.Grid = make([][]Tile, h)
	m.Explored = make([][]bool, h)
	for y := 0; y < h; y++ {
		m.Grid[y] = make([]Tile, w)
		m.Explored[y] = make([]bool, w)
	}
	m.generateDungeon()
	m.revealFog()
}

func (m *Model) generateDungeon() {
	m.Packs = make(map[Point]*MonsterPack)
	m.Combat = nil
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			m.Grid[y][x] = TileWall
			m.Explored[y][x] = false
		}
	}

	type Rect struct{ X, Y, W, H int }
	rooms := []Rect{}
	attempts := (m.MapWidth * m.MapHeight) / 50
	if attempts < 8 {
		attempts = 8
	}

	for i := 0; i < attempts; i++ {
		w := rand.Intn(7) + 7
		h := rand.Intn(4) + 4
		if m.MapWidth-w-2 <= 1 || m.MapHeight-h-2 <= 1 {
			continue
		}
		x := rand.Intn(m.MapWidth-w-2) + 1
		y := rand.Intn(m.MapHeight-h-2) + 1
		rooms = append(rooms, Rect{x, y, w, h})
		for ry := y; ry < y+h; ry++ {
			for rx := x; rx < x+w; rx++ {
				m.Grid[ry][rx] = TileFloor
			}
		}
	}

	if len(rooms) == 0 {
		return
	}

	carveWideCorridor := func(x1, y1, x2, y2 int) {
		cx, cy := x1, y1
		for cx != x2 {
			m.Grid[cy][cx] = TileFloor
			if cy+1 < m.MapHeight {
				m.Grid[cy+1][cx] = TileFloor
			}
			if x2 > cx {
				cx++
			} else {
				cx--
			}
		}
		for cy != y2 {
			m.Grid[cy][cx] = TileFloor
			if cx+1 < m.MapWidth {
				m.Grid[cy+1][cx] = TileFloor
			}
			if y2 > cy {
				cy++
			} else {
				cy--
			}
		}
	}

	for i := 0; i < len(rooms)-1; i++ {
		x1, y1 := rooms[i].X+rooms[i].W/2, rooms[i].Y+rooms[i].H/2
		x2, y2 := rooms[i+1].X+rooms[i+1].W/2, rooms[i+1].Y+rooms[i+1].H/2
		carveWideCorridor(x1, y1, x2, y2)
	}

	exitPos := Point{rooms[0].X + 1, rooms[0].Y + 1}
	if m.CurrentQuest.Type == QuestEscapeTrap && !m.CurrentQuest.Completed {
		m.Grid[exitPos.Y][exitPos.X] = TileFloor
	} else {
		m.Grid[exitPos.Y][exitPos.X] = TileExit
	}
	if exitPos.X+1 < rooms[0].X+rooms[0].W-1 {
		m.PartyPos = Point{exitPos.X + 1, exitPos.Y}
	} else {
		m.PartyPos = Point{exitPos.X, exitPos.Y + 1}
	}
	m.Grid[m.PartyPos.Y][m.PartyPos.X] = TileFloor

	endRoom := rooms[len(rooms)-1]
	m.Grid[endRoom.Y+endRoom.H/2][endRoom.X+endRoom.W/2] = TileStairs

	relicSpawned := false

	for i := 1; i < len(rooms); i++ {
		r := rooms[i]
		isBoss := (m.Floor%10 == 0 && i == len(rooms)-1)
		pos := Point{r.X + r.W/2, r.Y + r.H/2}
		if isBoss {
			m.Packs[pos] = spawnMonsterPack(true, m.Floor)
		} else {
			m.Packs[pos] = spawnMonsterPack(false, m.Floor)
			eventRoll := rand.Intn(13)
			cornerPos := Point{r.X + 1, r.Y + 1}
			switch {
			case (!relicSpawned && m.CurrentQuest.Type == QuestFindRelic) || (eventRoll == 12 && !relicSpawned):
				m.Grid[cornerPos.Y][cornerPos.X] = TileRelic
				relicSpawned = true
			case eventRoll == 0:
				m.Grid[cornerPos.Y][cornerPos.X] = TileAltar
			case eventRoll == 1:
				m.Grid[cornerPos.Y][cornerPos.X] = TileFountain
			case eventRoll == 2:
				m.Grid[cornerPos.Y][cornerPos.X] = TileTrappedChest
			case eventRoll == 3:
				m.Grid[cornerPos.Y][cornerPos.X] = TileBarrel
			case eventRoll >= 7:
				m.Grid[cornerPos.Y][cornerPos.X] = TileChest
			}
		}
	}
}

func (m *Model) revealFog() {
	radius := 6
	for y := m.PartyPos.Y - radius; y <= m.PartyPos.Y+radius; y++ {
		for x := m.PartyPos.X - radius; x <= m.PartyPos.X+radius; x++ {
			if x >= 0 && x < m.MapWidth && y >= 0 && y < m.MapHeight {
				dx := x - m.PartyPos.X
				dy := y - m.PartyPos.Y
				if dx*dx+dy*dy <= radius*radius {
					m.Explored[y][x] = true
				}
			}
		}
	}
}

func (m *Model) findNextStep() Point {
	retreat := m.checkRetreat()
	seekingFountain := m.needsHealing()
	forceDeeper := (m.CurrentQuest.Type == QuestEscapeTrap) && !m.CurrentQuest.Completed

	hasFountainOnMap := false
	for y := 0; y < m.MapHeight; y++ {
		for x := 0; x < m.MapWidth; x++ {
			if m.Grid[y][x] == TileFountain {
				hasFountainOnMap = true
				break
			}
		}
		if hasFountainOnMap {
			break
		}
	}

	findPath := func(avoidMonsters bool, targetCondition func(Point, Tile) bool) (Point, bool) {
		queue := []Point{m.PartyPos}
		visited := make(map[Point]bool)
		cameFrom := make(map[Point]Point)
		visited[m.PartyPos] = true
		dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr != m.PartyPos && targetCondition(curr, m.Grid[curr.Y][curr.X]) {
				step := curr
				for cameFrom[step] != m.PartyPos {
					step = cameFrom[step]
				}
				return step, true
			}

			for _, d := range dirs {
				next := Point{curr.X + d.X, curr.Y + d.Y}
				if next.X >= 0 && next.X < m.MapWidth && next.Y >= 0 && next.Y < m.MapHeight {
					if !visited[next] && m.Grid[next.Y][next.X] != TileWall {
						if avoidMonsters {
							if _, hasMob := m.Packs[next]; hasMob {
								continue
							}
						}
						visited[next] = true
						cameFrom[next] = curr
						queue = append(queue, next)
					}
				}
			}
		}
		return m.PartyPos, false
	}

	if seekingFountain && hasFountainOnMap {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileFountain
		})
		if found {
			return step
		}
	}

	if retreat && !forceDeeper {
		step, found := findPath(true, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
		step, found = findPath(false, func(p Point, t Tile) bool {
			return t == TileExit
		})
		if found {
			return step
		}
	}

	isExplorationQuest := m.CurrentQuest.Type == QuestReachFloor && !m.CurrentQuest.Completed

	isTarget := func(p Point, t Tile) bool {
		_, hasPack := m.Packs[p]
		if hasPack {
			return true
		}
		if seekingFountain && t == TileAltar {
			return false
		}
		if t == TileStairs && isExplorationQuest {
			return true
		}
		if t == TileStairs && !forceDeeper {
			hasVisibleLoot := false
			for y := 0; y < m.MapHeight; y++ {
				for x := 0; x < m.MapWidth; x++ {
					if m.Explored[y][x] {
						tile := m.Grid[y][x]
						if tile == TileChest || tile == TileRelic {
							hasVisibleLoot = true
							break
						}
					}
				}
			}
			if hasVisibleLoot {
				return false
			}
		}
		return t == TileChest || t == TileStairs || t == TileAltar || t == TileFountain || t == TileTrappedChest || t == TileBarrel || t == TileRelic
	}

	if seekingFountain {
		step, found := findPath(true, isTarget)
		if found {
			return step
		}
	}

	step, found := findPath(false, isTarget)
	if found {
		return step
	}

	return m.PartyPos
}

func (m *Model) needsHealing() bool {
	for _, h := range m.Party {
		if !h.IsDead && (float64(h.HP)/float64(h.MaxHP) <= 0.40 || h.Stress >= 80) {
			return true
		}
	}
	return false
}

func (m *Model) checkRetreat() bool {
	if len(m.Bag) >= m.currentBagCapacity() {
		return true
	}

	living := 0
	criticallyWounded := 0
	for _, h := range m.Party {
		if !h.IsDead {
			living++
			if float64(h.HP)/float64(h.MaxHP) <= 0.30 || h.Stress >= 140 {
				criticallyWounded++
			}
		}
	}
	if living <= 2 || criticallyWounded >= 2 {
		return true
	}
	return false
}

func (m *Model) isPartyWiped() bool {
	for _, h := range m.Party {
		if !h.IsDead {
			return false
		}
	}
	return true
}

func (m *Model) getRandomLivingHero() *Hero {
	var living []*Hero
	for _, h := range m.Party {
		if !h.IsDead {
			living = append(living, h)
		}
	}
	if len(living) == 0 {
		return nil
	}
	return living[rand.Intn(len(living))]
}

func (m *Model) addStress(h *Hero, amt int) {
	if h.IsDead {
		return
	}
	if m.Relic != nil && m.Relic.StressRes > 0 {
		amt = amt * (100 - m.Relic.StressRes) / 100
	}

	itemRes := 0
	for _, it := range []*EquipItem{h.Head, h.Chest, h.Legs} {
		if it != nil {
			itemRes += it.StressRes
		}
	}
	if itemRes > 50 {
		itemRes = 50
	}
	amt = amt * (100 - itemRes) / 100

	h.Stress += amt

	if amt >= 15 {
		splash := amt / 2
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += splash
			}
		}
	}

	if h.Stress >= 100 && h.Affliction == AfflictionNone {
		// Снижаем шанс воодушевления до 20%
		if rand.Intn(100) < 20 {
			h.Affliction = AfflictionVirtuous
			h.Stress = 0
			h.HP = h.MaxHP
			verb1 := TVerb(m.Lang, h.Gender, "превозмог", "превозмогла", "conquered")
			verb2 := TVerb(m.Lang, h.Gender, "обрел", "обрела", "gained")
			m.addLog(healStyle.Render(T(m.Lang, "dungeon.log.virtue", h.FullName(m.Lang), verb1, verb2)))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress = max(0, ally.Stress-20)
				}
			}
		} else {
			affs := []AfflictionType{AfflictionParanoid, AfflictionSelfish, AfflictionManiac}
			h.Affliction = affs[rand.Intn(len(affs))]
			verb := TVerb(m.Lang, h.Gender, "сломлен", "сломлена", "broken")
			affName := T(m.Lang, "affliction."+string(h.Affliction))
			m.addLog(stressStyle.Render(T(m.Lang, "dungeon.log.affliction", h.FullName(m.Lang), verb, affName)))
		}
	}

	if h.Stress >= 200 && h.Affliction != AfflictionVirtuous {
		h.Stress = 200
		if h.HP <= 1 {
			h.HP = 0
			h.CauseOfDeath = T(m.Lang, "dungeon.death.heart_attack", T(m.Lang, "affliction."+string(h.Affliction)))
			m.recordFallenHero(h)
			m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.heart_attack_death", h.DisplayName(m.Lang))))
			for _, ally := range m.Party {
				if !ally.IsDead {
					ally.Stress += 35
				}
			}
			return
		}

		h.HP = 1
		h.Stress = 160
		verb := TVerb(m.Lang, h.Gender, "схватился", "схватилась", "clutched")
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.heart_attack", h.FullName(m.Lang), verb)))
		for _, ally := range m.Party {
			if !ally.IsDead && ally != h {
				ally.Stress += 20
			}
		}
	}
}

func (m *Model) checkAndDrinkPotions(h *Hero) {
	if h.IsDead || len(h.Potions) == 0 || m.InTown {
		return
	}

	remainingPotions := []*Potion{}
	for _, p := range h.Potions {
		if p == nil {
			continue
		}

		shouldDrink := false
		switch p.Type {
		case PotionHP:
			missingHP := h.MaxHP - h.HP
			if float64(h.HP)/float64(h.MaxHP) <= 0.40 || missingHP >= p.Power {
				shouldDrink = true
			}
		case PotionMP:
			missingMP := h.MaxMP - h.MP
			skillNeeded := h.SkillCost
			if h.MP < skillNeeded || (h.MaxMP > 0 && float64(h.MP)/float64(h.MaxHP) <= 0.30) || missingMP >= p.Power {
				shouldDrink = true
			}
		case PotionStress:
			if h.Stress >= 70 || h.Affliction != AfflictionNone {
				shouldDrink = true
			}
		}

		if shouldDrink {
			verb := TVerb(m.Lang, h.Gender, "выпил", "выпила", "drank")
			pName := getPotionName(*p, m.Lang)
			hName := h.DisplayName(m.Lang)
			switch p.Type {
			case PotionHP:
				h.HP = min(h.MaxHP, h.HP+p.Power)
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_hp", hName, verb, pName, p.Power)))
			case PotionMP:
				h.MP = min(h.MaxMP, h.MP+p.Power)
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_mp", hName, verb, pName, p.Power)))
			case PotionStress:
				h.Stress = max(0, h.Stress-p.Power)
				h.Affliction = AfflictionNone
				verbStress := TVerb(m.Lang, h.Gender, "принял", "приняла", "took")
				m.addLog(potionStyle.Render(T(m.Lang, "dungeon.log.potion_stress", hName, verbStress, pName, p.Power)))
			}
		} else {
			remainingPotions = append(remainingPotions, p)
		}
	}
	h.Potions = remainingPotions
}

func (m *Model) checkAndAwardTitle(h *Hero) {
	if h.TitleKey != "" && h.TitleKey != "title.militia" {
		return
	}

	newTitleKey := ""
	switch {
	case h.Feats.BossKills >= 3:
		newTitleKey = "title.slayer_of_beasts"
	case h.Feats.BossKills >= 1:
		newTitleKey = "title.dragonslayer"
	case h.Feats.NearDeathEscapes >= 6:
		newTitleKey = "title.luck_cursed"
	case h.Feats.NearDeathEscapes >= 3:
		newTitleKey = "title.immortal"
	case h.Feats.TreasureFound >= 5:
		newTitleKey = "title.treasure_seeker"
	case h.Feats.SecretsRevealed >= 4:
		newTitleKey = "title.secret_keeper"
	case h.Class == ClassTank && h.Feats.Blocks >= 10:
		newTitleKey = "title.impenetrable"
	case h.Class == ClassTank && h.Feats.AggroPulled >= 8:
		newTitleKey = "title.storm_shield"
	case h.Class == ClassTank && h.Feats.DamageTaken >= 75:
		newTitleKey = "title.the_wall"
	case h.Class == ClassWarrior && h.Feats.Kills >= 12:
		newTitleKey = "title.blood_blade"
	case h.Class == ClassWarrior && h.Feats.Kills >= 5:
		newTitleKey = "title.executioner"
	case h.Class == ClassWarrior && h.Feats.DamageDealt >= 90:
		newTitleKey = "title.axe"
	case h.Class == ClassWarrior && h.Feats.CriticalStrikes >= 4:
		newTitleKey = "title.rank_cleaver"
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 6:
		newTitleKey = "title.phantom_strike"
	case h.Class == ClassRogue && h.Feats.CritsLanded >= 3:
		newTitleKey = "title.shadow"
	case h.Class == ClassRogue && h.Feats.Kills >= 4:
		newTitleKey = "title.blade"
	case h.Class == ClassRogue && h.Feats.Backstabs >= 5:
		newTitleKey = "title.knife_in_the_back"
	case h.Class == ClassRogue && h.Feats.LootStolen >= 3:
		newTitleKey = "title.deft_hand"
	case h.Class == ClassMage && h.Feats.DamageDealt >= 150:
		newTitleKey = "title.stormbringer"
	case h.Class == ClassMage && h.Feats.DamageDealt >= 90:
		newTitleKey = "title.cinder"
	case h.Class == ClassMage && h.Feats.CCDuration >= 40:
		newTitleKey = "title.chains_of_the_void"
	case h.Class == ClassMage && h.Feats.ManaBursts >= 3:
		newTitleKey = "title.flash"
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 120:
		newTitleKey = "title.grace"
	case h.Class == ClassCleric && h.Feats.HealsGiven >= 70:
		newTitleKey = "title.holy"
	case h.Class == ClassCleric && h.Feats.Revives >= 3:
		newTitleKey = "title.resurrector"
	case h.Class == ClassCleric && h.Feats.DoTsRemoved >= 4:
		newTitleKey = "title.purifier"
	}

	if newTitleKey != "" {
		h.TitleKey = newTitleKey
		verb := TVerb(m.Lang, h.Gender, "заслужил", "заслужила", "earned")
		m.addLog(titleStyle.Render(T(m.Lang, "dungeon.log.title_awarded", h.DisplayName(m.Lang), verb, T(m.Lang, newTitleKey))))
	}
}

func (m *Model) handleAltar() {
	m.Stats.AltarsUsed++
	m.checkQuestProgress(QuestUseAltar, "", 1)
	target := m.getRandomLivingHero()

	if target != nil {
		// Алтарь берет 35% от текущего HP, не меньше 15 урона
		bloodCost := max(15, target.HP*35/100)
		target.HP -= bloodCost
		target.BaseAtk += 3
		m.addStress(target, 25)

		hName := target.DisplayName(m.Lang)
		if target.HP <= 0 {
			target.HP = 0
			target.CauseOfDeath = T(m.Lang, "dungeon.death.altar")
			m.recordFallenHero(target)
			verb := TVerb(m.Lang, target.Gender, "пал", "пала", "fell")
			m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.altar_death", hName, verb)))
		} else {
			verb := TVerb(m.Lang, target.Gender, "пожертвовал", "пожертвовала", "sacrificed")
			m.addLog(altarStyle.Render(T(m.Lang, "dungeon.log.altar_success", hName, verb, bloodCost)))
		}
	}
}

func (m *Model) handleFountain() {
	m.Stats.FountainsUsed++
	for _, h := range m.Party {
		if !h.IsDead {
			h.HP = h.MaxHP
			h.MP = h.MaxMP
			h.Stress = max(0, h.Stress-60)
			if h.Stress == 0 {
				h.Affliction = AfflictionNone
			}
			h.IsGuarding = false
			h.IsBerserk = false
			h.IsStealthed = false
			h.IsCharged = false
			h.IsAura = false
		}
	}
	m.addLog(fountStyle.Render(T(m.Lang, "dungeon.log.fountain")))
}

func (m *Model) handleTrappedChest() {
	m.Stats.ChestsOpened++
	m.checkQuestProgress(QuestOpenChests, "", 1)
	d20 := rand.Intn(20) + 1

	rogueBonus := 0
	for _, h := range m.Party {
		if h.Class == ClassRogue && !h.IsDead {
			rogueBonus = 3
			break
		}
	}

	rollSuccess := (d20+rogueBonus >= 12)

	if rollSuccess {
		m.Stats.TrapsDisarmed++
		gold := int(float64(rand.Intn(25)+15) * m.Relic.GoldMult)
		m.Gold += gold
		m.Stats.TotalGoldEarned += gold

		targetClass := ClassWarrior
		if lh := m.getRandomLivingHero(); lh != nil {
			targetClass = lh.Class
			lh.AddTreasure()
			lh.RevealSecret()
			m.checkAndAwardTitle(lh)
		}
		rareItem := generateItemForClass(targetClass, m.Floor+1)
		m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.trapped_chest_success", gold, rareItem.DisplayName(m.Lang))))
		m.equipOrBag(rareItem)
	} else {
		m.addLog(dangerStyle.Render(T(m.Lang, "dungeon.log.trapped_chest_boom", d20)))
		for _, h := range m.Party {
			if !h.IsDead {
				trapDmg := rand.Intn(12) + 10
				h.HP -= trapDmg
				m.addStress(h, 25)
				if h.HP <= 0 {
					h.HP = 0
					h.CauseOfDeath = T(m.Lang, "dungeon.death.trap")
					m.recordFallenHero(h)
				}
			}
		}
	}
}

func (m *Model) handleRelicTile() {
	m.checkQuestProgress(QuestFindRelic, "", 1)
	newLevel := 2
	if m.Floor >= 6 {
		newLevel = 3
	}
	newRelic := generateRelic(newLevel)

	if m.Relic == nil || newRelic.Level > m.Relic.Level {
		m.Relic = &newRelic
		m.addLog(relicTileStyle.Render(T(m.Lang, "dungeon.log.relic_found", T(m.Lang, newRelic.NameKey))))
	} else {
		m.Gold += 80
		m.Stats.TotalGoldEarned += 80
		m.addLog(goldStyle.Render(T(m.Lang, "dungeon.log.relic_salvaged")))
	}
}

func spawnMonsterPack(isBoss bool, floor int) *MonsterPack {
	pack := &MonsterPack{IsBoss: isBoss}
	scaleMult := 1 + (floor / 8)

	if isBoss {
		if floor%10 == 0 {
			dragonLvl := floor
			dragonHP := (320 + (floor * 25)) * scaleMult
			dragon := &Monster{
				ID: 1, Type: MobDragon, NameKey: "mob.boss_dragon", Level: dragonLvl, Affix: AffixFire,
				Glyph: 'D', Color: "196", HP: dragonHP, MaxHP: dragonHP,
				Atk: (30 + floor*2) * scaleMult, Defense: (10 + floor/2) * scaleMult, Speed: 12, Exp: 450 * scaleMult,
			}
			pack.Members = append(pack.Members, dragon)

			for i := 1; i <= 2; i++ {
				guard := &Monster{
					ID: i + 1, Type: MobDeathKnight, NameKey: "mob.death_knight", Level: dragonLvl - 1, Affix: AffixVampiric,
					Glyph: 'K', Color: "89", HP: (85 + floor*6) * scaleMult, MaxHP: (85 + floor*6) * scaleMult,
					Atk: (22 + floor*2) * scaleMult, Defense: (7 + floor/3) * scaleMult, Speed: 10, Exp: 75 * scaleMult,
				}
				pack.Members = append(pack.Members, guard)
			}
			return pack
		}

		bossLvl := floor
		hp := (90 + (bossLvl * 18)) * scaleMult
		bType := MobOrc
		bGlyph := 'B'
		bColor := "202"
		bNameKey := "mob.boss_orc"
		bAtk := (16 + (bossLvl * 2)) * scaleMult
		bDef := (5 + bossLvl/3) * scaleMult

		if floor >= 7 {
			bType = MobGolem
			bNameKey = "mob.boss_golem"
			bGlyph = 'G'
			bColor = "141"
			bDef = (9 + floor/4) * scaleMult
		} else if floor >= 4 {
			bType = MobDrowned
			bNameKey = "mob.boss_leviathan"
			bGlyph = 'L'
			bColor = "31"
		}

		pack.Members = append(pack.Members, &Monster{
			ID: 1, Type: bType, NameKey: bNameKey, Level: bossLvl, Affix: AffixStone,
			Glyph: bGlyph, Color: bColor, HP: hp, MaxHP: hp,
			Atk: bAtk, Defense: bDef, Speed: 10, Exp: (90 + bossLvl*10) * scaleMult,
		})
		for i := 1; i <= 2; i++ {
			pack.Members = append(pack.Members, &Monster{
				ID: i + 1, Type: MobSkeleton, NameKey: "mob.skeleton", Level: floor, Affix: AffixNone,
				Glyph: 's', Color: "245", HP: (28 + floor*6) * scaleMult, MaxHP: (28 + floor*6) * scaleMult,
				Atk: (11 + floor*2) * scaleMult, Defense: 4 * scaleMult, Speed: 9, Exp: (22 + floor*3) * scaleMult,
			})
		}
		return pack
	}

	maxExtra := floor / 6
	if maxExtra > 4 {
		maxExtra = 4
	}
	packSize := rand.Intn(3) + 3 + maxExtra
	affixes := []MonsterAffix{AffixNone, AffixFire, AffixPoison, AffixFrost, AffixStone, AffixVampiric}

	for i := 1; i <= packSize; i++ {
		mobLvl := floor
		if rand.Intn(100) < 35 {
			mobLvl++
		}

		aff := AffixNone
		if floor >= 2 && rand.Intn(100) < (25+(floor*5)) {
			aff = affixes[rand.Intn(len(affixes))]
		}

		var mType MonsterType
		var glyph rune
		var color string
		var baseAtk, baseDef, baseHP int

		biomeCycle := (floor - 1) % 5

		if biomeCycle == 0 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobRat, 'r', "137", 6, 0, 15
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGoblin, 'g', "118", 8, 1, 19
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSkeleton, 's', "252", 10, 3, 25
			}
		} else if biomeCycle == 1 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSlime, 'c', "43", 11, 1, 30
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDrowned, 'u', "31", 13, 2, 38
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobLizard, 'l', "29", 14, 3, 34
			}
		} else if biomeCycle == 2 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobImp, 'i', "208", 15, 2, 40
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobOrc, 'o', "130", 17, 4, 52
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobSalamander, 'm', "196", 18, 3, 46
			}
		} else if biomeCycle == 3 {
			roll := rand.Intn(3)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGargoyle, 'G', "102", 19, 6, 58
			} else if roll == 1 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobGolem, 'C', "141", 20, 7, 68
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobPhantom, 'p', "159", 22, 2, 50
			}
		} else {
			roll := rand.Intn(2)
			if roll == 0 {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobVoidDemon, 'V', "161", 24, 5, 75
			} else {
				mType, glyph, color, baseAtk, baseDef, baseHP = MobDeathKnight, 'K', "89", 26, 7, 85
			}
		}

		hp := (baseHP + (mobLvl * 6)) * scaleMult
		atk := (baseAtk + (mobLvl * 3)) * scaleMult
		def := (baseDef + (mobLvl / 2)) * scaleMult

		if aff == AffixFire {
			atk += 4 * scaleMult
		} else if aff == AffixStone {
			def += 4 * scaleMult
			hp += 15 * scaleMult
		}

		pack.Members = append(pack.Members, &Monster{
			ID: i, Type: mType, NameKey: "mob." + string(mType), Level: mobLvl, Affix: aff,
			Glyph: glyph, Color: color, HP: hp, MaxHP: hp, Atk: atk, Defense: def, Speed: 8 + mobLvl/2, Exp: (12 + mobLvl*4) * scaleMult,
		})
	}
	return pack
}

func generateItemForClassSlot(class HeroClass, slot EquipSlot, floor int) EquipItem {
	matIdx := floor / 4
	if matIdx > 3 {
		matIdx = 3
	}

	var mat MaterialTier
	if slot == SlotWeapon {
		if class == ClassMage {
			mat = MageWeaponMaterials[rand.Intn(matIdx+1)]
		} else {
			mat = MetalMaterials[rand.Intn(matIdx+1)]
		}
	} else {
		switch class {
		case ClassTank, ClassWarrior:
			mat = MetalMaterials[rand.Intn(matIdx+1)]
		case ClassMage:
			mat = ClothMaterials[rand.Intn(matIdx+1)]
		default:
			mat = LeatherMaterials[rand.Intn(matIdx+1)]
		}
	}

	upg := 0
	if rand.Intn(100) > 75 {
		upg = rand.Intn(floor/3 + 1)
		if upg > 3 {
			upg = 3
		}
	}

	var pfx *PrefixDef
	if rand.Intn(100) > 60 {
		p := Prefixes[rand.Intn(len(Prefixes))]
		pfx = &p
	}

	var sfx *SuffixDef
	if rand.Intn(100) > 70 {
		s := Suffixes[rand.Intn(len(Suffixes))]
		if pfx != nil && pfx.Element == ElemFrost && s.Effect == SuffFury {
			s = Suffixes[2]
		}
		sfx = &s
	}

	tier := floor / 4
	if tier > 3 {
		tier = 3
	}
	if rand.Intn(100) < 25 && tier < 3 {
		tier++
	}

	baseKey := fmt.Sprintf("item.%s.%s.%d", string(class), string(slot), tier+1)
	baseStat := 2
	bonusMP := 0
	bonusHP := 0
	critBonus := 0
	blockBonus := 0
	stressRes := 0
	speedBonus := 0
	cat := ArmorMedium

	switch class {
	case ClassTank:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			baseStat = 3 + tier*2
			blockBonus = 2 + tier*2
		case SlotChest:
			baseStat = 4 + tier*3
			bonusHP = 10 + tier*10
		case SlotHead:
			baseStat = 2 + tier*2
			blockBonus = 1 + tier
		case SlotLegs:
			baseStat = 2 + tier*2
			bonusHP = 5 + tier*5
		}

	case ClassWarrior:
		cat = ArmorHeavy
		switch slot {
		case SlotWeapon:
			baseStat = 5 + tier*3
			critBonus = 1 + tier
		case SlotChest:
			baseStat = 3 + tier*2
			bonusHP = 8 + tier*8
		case SlotHead:
			baseStat = 2 + tier*2
			critBonus = 1
		case SlotLegs:
			baseStat = 2 + tier*2
		}

	case ClassRogue:
		cat = ArmorMedium
		if slot != SlotWeapon {
			speedBonus = 1 + tier
		}
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			critBonus = 2 + tier*2
		case SlotChest:
			baseStat = 2 + tier*2
			critBonus = 1 + tier
		case SlotHead:
			baseStat = 1 + tier*2
			critBonus = 1
		case SlotLegs:
			baseStat = 1 + tier*2
		}

	case ClassMage:
		cat = ArmorLight
		if slot != SlotWeapon {
			speedBonus = tier
		}
		switch slot {
		case SlotWeapon:
			baseStat = 6 + tier*3
			bonusMP = 10 + tier*10
		case SlotChest:
			baseStat = 2 + tier*2
			bonusMP = 15 + tier*10
		case SlotHead:
			baseStat = 1 + tier*2
			bonusMP = 8 + tier*6
		case SlotLegs:
			baseStat = 1 + tier*2
			bonusMP = 6 + tier*4
		}

	case ClassCleric:
		cat = ArmorMedium
		switch slot {
		case SlotWeapon:
			baseStat = 4 + tier*2
			bonusMP = 8 + tier*6
		case SlotChest:
			baseStat = 3 + tier*2
			stressRes = 10 + tier*5
		case SlotHead:
			baseStat = 2 + tier*2
			stressRes = 5 + tier*5
		case SlotLegs:
			baseStat = 2 + tier*2
			bonusHP = 6 + tier*6
		}
	}

	val := (baseStat + tier*3) * mat.ValueMult * 12

	return EquipItem{
		BaseNameKey: baseKey, Slot: slot, Category: cat, AllowedClass: class,
		Material: mat, UpgradeLevel: upg, BaseStat: baseStat,
		BonusMP: bonusMP, BonusHP: bonusHP, CritBonus: critBonus,
		BlockBonus: blockBonus, StressRes: stressRes, SpeedBonus: speedBonus, Value: val,
		Prefix: pfx, Suffix: sfx,
	}
}

func generateItemForClass(class HeroClass, floor int) EquipItem {
	slots := []EquipSlot{SlotWeapon, SlotHead, SlotChest, SlotLegs}
	chosenSlot := slots[rand.Intn(len(slots))]
	return generateItemForClassSlot(class, chosenSlot, floor)
}
