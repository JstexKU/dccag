package main

// ============================================================
// КАРТА ГОРОДА
// ============================================================
//
// Столица — это карта, по которой ходит отряд. Девять зданий (восемь заведений и
// Врата подземелья) стоят кольцом в порядке обхода автопилота: рынок, магистрат, храм,
// таверна, гильдия, кузница, дубильня, алхимик и, наконец, Врата в центре города.
// Все здания обращены дверями на улицы; дома, трава и деревья непроходимы,
// поэтому отряд ходит только по улицам.

const (
	townW    = 90
	townH    = 44
	townBldW = 13
	townBldH = 9
)

// Геометрия: три ряда зданий, под каждым — двухклеточная улица; вертикальные
// улицы лежат между колонками и вдоль краёв, так что получается замкнутая сеть дорог.
var (
	townRowTops    = [3]int{3, 17, 31}
	townColCenters = [3]int{16, 45, 74}
	townStreetCols = [4]int{4, 30, 59, 84} // левая клетка двухклеточных вертикальных улиц
)

// Шаблон здания увеличен на 1 строку в высоту (9 строк).
// Строки 4 и 5 — название типа здания (с переносом).
// Строка 7 — имя заведения или уровень/этаж.
var townTemplate = [townBldH]string{
	"    ▄▄▄▄▄    ",
	"  ▄███████▄  ",
	" ▄█████████▄ ",
	"▄███████████▄",
	" ║         ║ ",
	" ║         ║ ",
	" ║ ◘     ◘ ║ ",
	" ║         ║ ",
	" ╚════▓════╝ ",
}

// У Врат вместо окон — тёмный зев подземелья.
const townGateRow6 = " ║ ▒▒▒▒▒▒▒ ║ "

const (
	townDoorCol   = 6 // колонка двери в шаблоне
	townSignFrom  = 2 // табличка занимает колонки 2..10 (9 клеток)
	townSignLen   = 9
	townSignRow1  = 4
	townSignRow2  = 5
	townPlaqRow   = 7
)

// TownCell — клетка карты города.
type TownCell struct {
	R     rune
	Color uint8 // цвет из 256-цветной палитры (0 — по умолчанию)
	Bold  bool
	Block bool  // непроходима
	Grass bool  // «чистая» трава: на ней можно посадить дерево или цветок
	Door  bool  // дверь здания Bld: вход
	Slot  uint8 // 1 — строка типа 1, 2 — строка типа 2, 3 — табличка с именем/уровнем; 0 — обычная
	SlotX uint8 // номер символа на табличке
	Bld   int8  // индекс здания для двери и табличек
}

// TownBuilding — здание на карте.
type TownBuilding struct {
	Phase TownPhase
	X, Y  int
	W, H  int
	Door  Point // клетка двери (вход)
	Front Point // клетка улицы перед дверью
	Roof  uint8
}

// TownMap — карта города.
type TownMap struct {
	W, H      int
	Cells     [][]TownCell
	Buildings []TownBuilding
	Gate      int // индекс Врат подземелья в Buildings
}

type townBuildingSpec struct {
	Phase    TownPhase
	Row, Col int
	Roof     uint8
	Emblem   rune
}

// townSpecs — здания в порядке обхода. Последнее — Врата подземелья.
var townSpecs = []townBuildingSpec{
	{TownPhaseSellLoot, 0, 0, 208, '$'},
	{TownPhaseMagistrate, 0, 1, 135, '▲'},
	{TownPhaseChurch, 0, 2, 252, '┼'},
	{TownPhaseTavern, 1, 2, 130, 'o'},
	{TownPhaseGuild, 2, 2, 124, 'X'},
	{TownPhaseSmithy, 2, 1, 202, '*'},
	{TownPhaseTannery, 2, 0, 137, '~'},
	{TownPhaseAlchemist, 1, 0, 71, '%'},
	{TownPhaseDepart, 1, 1, 240, '▼'},
}

// townDirs — четыре направления шага по карте города.
var townDirs = [4]Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

var (
	townTuftRunes   = []rune{'"', ',', '\''}
	townTuftColors  = []uint8{22, 28, 34}
	townTreeColors  = []uint8{28, 34, 70}
	townFlowerColor = []uint8{207, 220, 39, 213}
)

func townGrass() TownCell {
	c := TownCell{R: ' ', Block: true, Grass: true, Bld: -1}
	if rng.Intn(100) < 28 {
		c.R = townTuftRunes[rng.Intn(len(townTuftRunes))]
		c.Color = townTuftColors[rng.Intn(len(townTuftColors))]
	}
	return c
}

func townRoadCell() TownCell {
	if rng.Intn(100) < 22 {
		return TownCell{R: '▒', Color: 238, Bld: -1}
	}
	return TownCell{R: '░', Color: 240, Bld: -1}
}

// generateTown создаёт карту города: улицы, здания с вывесками и украшения.
func generateTown() *TownMap {
	t := &TownMap{W: townW, H: townH}
	t.Cells = make([][]TownCell, townH)
	for y := range t.Cells {
		t.Cells[y] = make([]TownCell, townW)
		for x := range t.Cells[y] {
			t.Cells[y][x] = townGrass()
		}
	}

	// Городская стена по периметру.
	for x := 0; x < townW; x++ {
		t.setWall(x, 0)
		t.setWall(x, townH-1)
	}
	for y := 0; y < townH; y++ {
		t.setWall(0, y)
		t.setWall(townW-1, y)
	}

	// Улицы.
	for _, top := range townRowTops {
		t.paintRoad(townStreetCols[0], top+townBldH, townStreetCols[3]+1, top+townBldH+1)
	}
	for _, x := range townStreetCols {
		t.paintRoad(x, townRowTops[0]+townBldH, x+1, townRowTops[2]+townBldH+1)
	}

	// Здания; колонны меняются местами при «зеркальном» городе, а сами здания слегка смещаются.
	mirror := rng.Intn(2) == 1
	for i, spec := range townSpecs {
		col := spec.Col
		if mirror {
			col = 2 - col
		}
		x := townColCenters[col] - townBldW/2 + rng.Intn(5) - 2
		t.placeBuilding(i, spec, x, townRowTops[spec.Row])
	}
	t.Gate = len(townSpecs) - 1

	// Фонтаны на перекрёстках средней улицы: по одной клетке из четырёх, проход остаётся.
	t.setDecor(townStreetCols[1], townRowTops[1]+townBldH, '≈', 45)
	t.setDecor(townStreetCols[2]+1, townRowTops[1]+townBldH+1, '≈', 45)

	// Деревья и цветы на траве.
	for y := range t.Cells {
		for x := range t.Cells[y] {
			c := &t.Cells[y][x]
			if !c.Grass {
				continue
			}
			switch roll := rng.Intn(100); {
			case roll < 6:
				*c = TownCell{R: '♣', Color: townTreeColors[rng.Intn(len(townTreeColors))], Block: true, Bld: -1}
			case roll < 8:
				*c = TownCell{R: '*', Color: townFlowerColor[rng.Intn(len(townFlowerColor))], Block: true, Bld: -1}
			}
		}
	}
	return t
}

func (t *TownMap) setWall(x, y int) {
	t.Cells[y][x] = TownCell{R: '▓', Color: 240, Block: true, Bld: -1}
}

func (t *TownMap) setDecor(x, y int, r rune, color uint8) {
	t.Cells[y][x] = TownCell{R: r, Color: color, Bold: true, Block: true, Bld: -1}
}

// paintRoad мостит прямоугольник улицей (границы включительно), не затрагивая стену.
func (t *TownMap) paintRoad(x1, y1, x2, y2 int) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if x <= 0 || y <= 0 || x >= townW-1 || y >= townH-1 {
				continue
			}
			t.Cells[y][x] = townRoadCell()
		}
	}
}

func (t *TownMap) placeBuilding(idx int, spec townBuildingSpec, x, y int) {
	rows := townTemplate
	if spec.Phase == TownPhaseDepart {
		rows[6] = townGateRow6
	}
	for ry := 0; ry < townBldH; ry++ {
		runes := []rune(rows[ry])
		for rx := 0; rx < townBldW; rx++ {
			c := TownCell{R: runes[rx], Block: true, Bld: int8(idx)}
			switch {
			case ry == 0 && rx == townDoorCol:
				c.R, c.Color, c.Bold = spec.Emblem, 229, true
			case ry <= 3 && c.R != ' ':
				c.Color = spec.Roof
			case ry == townSignRow1 && rx >= townSignFrom && rx < townSignFrom+townSignLen:
				c.Slot, c.SlotX = 1, uint8(rx-townSignFrom)
			case ry == townSignRow2 && rx >= townSignFrom && rx < townSignFrom+townSignLen:
				c.Slot, c.SlotX = 2, uint8(rx-townSignFrom)
			case ry == townPlaqRow && rx >= townSignFrom && rx < townSignFrom+townSignLen:
				c.Slot, c.SlotX = 3, uint8(rx-townSignFrom)
			case c.R == '◘':
				c.Color = 220
			case c.R == '▒':
				c.Color = 93
			case ry == townBldH-1 && rx == townDoorCol:
				c.Color, c.Bold, c.Block, c.Door = 94, true, false, true
			case c.R != ' ':
				c.Color = 250
			}
			t.Cells[y+ry][x+rx] = c
		}
	}
	door := Point{x + townDoorCol, y + townBldH - 1}
	t.Buildings = append(t.Buildings, TownBuilding{
		Phase: spec.Phase, X: x, Y: y, W: townBldW, H: townBldH,
		Door: door, Front: Point{door.X, door.Y + 1}, Roof: spec.Roof,
	})
}

// ---------------------------------------------------------------------------
// Навигация
// ---------------------------------------------------------------------------

func (t *TownMap) inBounds(p Point) bool {
	return p.X >= 0 && p.X < t.W && p.Y >= 0 && p.Y < t.H
}

func (t *TownMap) walkable(p Point) bool {
	return t.inBounds(p) && !t.Cells[p.Y][p.X].Block
}

// doorAt возвращает индекс здания, чья дверь находится в клетке p (или -1).
func (t *TownMap) doorAt(p Point) int {
	if !t.inBounds(p) {
		return -1
	}
	c := t.Cells[p.Y][p.X]
	if !c.Door {
		return -1
	}
	return int(c.Bld)
}

// buildingFor возвращает индекс здания, которому принадлежит фаза города (или -1).
func (t *TownMap) buildingFor(phase TownPhase) int {
	for i, b := range t.Buildings {
		if b.Phase == phase {
			return i
		}
	}
	return -1
}

// pathStep возвращает первую клетку кратчайшего пути from → to по проходимым клеткам.
func (t *TownMap) pathStep(from, to Point) (Point, bool) {
	if from == to || !t.walkable(from) || !t.walkable(to) {
		return from, false
	}
	n := t.W * t.H
	prev := make([]int32, n)
	for i := range prev {
		prev[i] = -2
	}
	start := int32(from.Y*t.W + from.X)
	goal := int32(to.Y*t.W + to.X)
	prev[start] = -1
	queue := []int32{start}

	for head := 0; head < len(queue); head++ {
		ci := queue[head]
		if ci == goal {
			step := ci
			for prev[step] != start {
				step = prev[step]
			}
			return Point{int(step) % t.W, int(step) / t.W}, true
		}
		cx, cy := int(ci)%t.W, int(ci)/t.W
		for _, d := range townDirs {
			np := Point{cx + d.X, cy + d.Y}
			if !t.walkable(np) {
				continue
			}
			ni := int32(np.Y*t.W + np.X)
			if prev[ni] != -2 {
				continue
			}
			prev[ni] = ci
			queue = append(queue, ni)
		}
	}
	return from, false
}