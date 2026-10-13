package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTownTemplateShape(t *testing.T) {
	for i, row := range townTemplate {
		if n := len([]rune(row)); n != townBldW {
			t.Fatalf("строка шаблона %d: ширина %d, ожидалась %d", i, n, townBldW)
		}
	}
	if n := len([]rune(townGateRow6)); n != townBldW {
		t.Fatalf("строка окон Врат: ширина %d", n)
	}
	if got := []rune(townTemplate[townBldH-1])[townDoorCol]; got != '▓' {
		t.Fatalf("дверь должна стоять в колонке %d нижней строки, а там %q", townDoorCol, got)
	}
	for _, row := range []int{townSignRow1, townSignRow2, townPlaqRow} {
		runes := []rune(townTemplate[row])
		for x := townSignFrom; x < townSignFrom+townSignLen; x++ {
			if runes[x] != ' ' {
				t.Fatalf("клетки табличек в шаблоне должны быть пустыми: строка %d, колонка %d", row, x)
			}
		}
	}
	if len(townTemplate) != townBldH {
		t.Fatalf("высота шаблона %d", len(townTemplate))
	}
}

func TestGenerateTownInvariants(t *testing.T) {
	for seed := int64(1); seed <= 60; seed++ {
		seedRNG(seed)
		tm := generateTown()

		if tm.W != townW || tm.H != townH || len(tm.Cells) != townH {
			t.Fatalf("seed %d: размеры карты %dx%d", seed, tm.W, tm.H)
		}
		if len(tm.Buildings) != len(townSpecs) {
			t.Fatalf("seed %d: зданий %d, ожидалось %d", seed, len(tm.Buildings), len(townSpecs))
		}
		if tm.Buildings[tm.Gate].Phase != TownPhaseDepart {
			t.Fatalf("seed %d: Врата должны быть последним зданием", seed)
		}
		for i, b := range tm.Buildings {
			if b.Phase != townSpecs[i].Phase {
				t.Fatalf("seed %d: порядок зданий нарушен на %d", seed, i)
			}
		}

		for x := 0; x < tm.W; x++ {
			if tm.walkable(Point{x, 0}) || tm.walkable(Point{x, tm.H - 1}) {
				t.Fatalf("seed %d: городская стена пробита (x=%d)", seed, x)
			}
		}
		for y := 0; y < tm.H; y++ {
			if tm.walkable(Point{0, y}) || tm.walkable(Point{tm.W - 1, y}) {
				t.Fatalf("seed %d: городская стена пробита (y=%d)", seed, y)
			}
		}

		occupied := map[Point]int{}
		for i, b := range tm.Buildings {
			if b.X < 1 || b.Y < 1 || b.X+b.W > tm.W-1 || b.Y+b.H > tm.H-1 {
				t.Fatalf("seed %d: здание %d вышло за карту: %+v", seed, i, b)
			}
			for dy := 0; dy < b.H; dy++ {
				for dx := 0; dx < b.W; dx++ {
					p := Point{b.X + dx, b.Y + dy}
					if j, taken := occupied[p]; taken {
						t.Fatalf("seed %d: здания %d и %d пересекаются в %v", seed, j, i, p)
					}
					occupied[p] = i
					if p != b.Door && tm.walkable(p) {
						t.Fatalf("seed %d: стена здания %d проходима в %v", seed, i, p)
					}
				}
			}
			if b.Door != (Point{b.X + townDoorCol, b.Y + townBldH - 1}) {
				t.Fatalf("seed %d: дверь здания %d не там: %v", seed, i, b.Door)
			}
			if !tm.walkable(b.Door) || !tm.walkable(b.Front) {
				t.Fatalf("seed %d: к двери здания %d не подойти", seed, i)
			}
			if tm.doorAt(b.Door) != i || tm.doorAt(b.Front) != -1 {
				t.Fatalf("seed %d: дверь здания %d определяется неверно", seed, i)
			}
			if tm.buildingFor(b.Phase) != i {
				t.Fatalf("seed %d: buildingFor(%v) ошибается", seed, b.Phase)
			}
		}

		from := tm.Buildings[tm.Gate].Front
		for i, b := range tm.Buildings {
			if _, ok := tm.pathStep(from, b.Door); !ok {
				t.Fatalf("seed %d: до двери здания %d не добраться", seed, i)
			}
		}

		slots := map[[2]int]int{}
		for y := range tm.Cells {
			for _, c := range tm.Cells[y] {
				if c.Slot != 0 {
					slots[[2]int{int(c.Bld), int(c.Slot)}]++
				}
			}
		}
		for i := range tm.Buildings {
			for slot := 1; slot <= 3; slot++ {
				if n := slots[[2]int{i, slot}]; n != townSignLen {
					t.Fatalf("seed %d: у здания %d на табличке %d клеток подписи: %d", seed, i, slot, n)
				}
			}
		}
	}
}

func TestTownTourIsWalkable(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		seedRNG(seed)
		tm := generateTown()
		pos := tm.Buildings[tm.Gate].Front
		steps := 0
		for _, b := range tm.Buildings {
			for pos != b.Door {
				next, ok := tm.pathStep(pos, b.Door)
				if !ok {
					t.Fatalf("seed %d: нет пути из %v к двери %v", seed, pos, b.Door)
				}
				if absInt(next.X-pos.X)+absInt(next.Y-pos.Y) != 1 || !tm.walkable(next) {
					t.Fatalf("seed %d: некорректный шаг %v → %v", seed, pos, next)
				}
				pos = next
				if steps++; steps > 2000 {
					t.Fatalf("seed %d: обход не кончается", seed)
				}
			}
			pos = b.Front
		}
		if steps < 150 || steps > 500 {
			t.Fatalf("seed %d: длина полного обхода %d шагов — вне ожидаемых пределов", seed, steps)
		}
	}
}

func TestPathStepEdgeCases(t *testing.T) {
	seedRNG(2)
	tm := generateTown()
	b := tm.Buildings[0]
	if _, ok := tm.pathStep(b.Front, b.Front); ok {
		t.Error("путь из клетки в неё же не нужен")
	}
	if _, ok := tm.pathStep(Point{0, 0}, b.Door); ok {
		t.Error("из стены пути нет")
	}
	if _, ok := tm.pathStep(b.Front, Point{-5, 3}); ok {
		t.Error("за карту пути нет")
	}
	if next, ok := tm.pathStep(b.Front, b.Door); !ok || next != b.Door {
		t.Errorf("от фасада до двери один шаг, получено %v/%v", next, ok)
	}
}

func TestCenterRunes(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "         "},
		{"Рынок", "  Рынок  "},
		{"АБВГДЕЖЗИ", "АБВГДЕЖЗИ"},
		{"АБВГДЕЖЗИК", "АБВГДЕЖЗ…"},
	}
	for _, c := range cases {
		got := string(centerRunes(c.in, townSignLen))
		if got != c.want {
			t.Errorf("centerRunes(%q) = %q, ожидалось %q", c.in, got, c.want)
		}
		if n := len([]rune(got)); n != townSignLen {
			t.Errorf("centerRunes(%q): длина %d", c.in, n)
		}
	}
}

func townRenderModel(t *testing.T) *Model {
	t.Helper()
	useTempSave(t)
	seedRNG(4)
	m := initialModel()
	m.InTown = true
	m.TownWalking = true
	return &m
}

func TestRenderTownMapViewport(t *testing.T) {
	m := townRenderModel(t)
	for _, size := range [][2]int{{40, 12}, {80, 24}, {120, 40}, {200, 60}} {
		for _, bi := range []int{0, 4, m.Town.Gate} {
			m.TownPos = m.Town.Buildings[bi].Front
			out := m.renderTownMap(size[0], size[1])
			lines := strings.Split(out, "\n")
			if len(lines) != size[1] {
				t.Fatalf("окно %v: %d строк", size, len(lines))
			}
			for y, line := range lines {
				if w := lipgloss.Width(line); w != size[0] {
					t.Fatalf("окно %v, строка %d: ширина %d", size, y, w)
				}
			}
			if !strings.Contains(out, "@") {
				t.Fatalf("окно %v: отряд не виден на фасаде здания %d", size, bi)
			}
		}
	}
}

func TestRenderTownMapShowsSignsInBothLanguages(t *testing.T) {
	m := townRenderModel(t)
	m.TownPos = m.Town.Buildings[m.Town.Gate].Front
	for _, lang := range []Language{LangRU, LangEN} {
		m.Lang = lang
		out := m.renderTownMap(townW+10, townH+4)
		for _, b := range m.Town.Buildings {
			t1, _ := m.townBuildingType(b)
			want := strings.TrimSpace(string(centerRunes(t1, townSignLen)))
			if want != "" && !strings.Contains(out, want) {
				t.Fatalf("%s: тип здания %q не нарисован", lang, want)
			}
		}
		if !strings.Contains(out, T(lang, "town.gate_floor", 1)) {
			t.Fatalf("%s: у Врат должна быть табличка с этажом", lang)
		}
	}
}

func TestRenderTownMapHighlightsAutopilotTargetOnly(t *testing.T) {
	m := townRenderModel(t)
	m.TownPhase = TownPhaseChurch
	m.TownPos = m.Town.Buildings[0].Front
	if out := m.renderTownMap(townW, townH); out == "" {
		t.Fatal("карта не нарисована")
	}
	if idx := m.Town.buildingFor(m.TownPhase); idx != 2 {
		t.Fatalf("храм — третье здание обхода, а получено %d", idx)
	}
}