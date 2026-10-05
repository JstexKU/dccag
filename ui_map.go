package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// updateCamera вычисляет начало видимой области с учётом мёртвой зоны (deadzone),
// исключая скачки экрана вверх-вниз при локальных шагах отряда.
func (m *Model) updateCamera(viewW, viewH int) (int, int) {
	if m.CameraPos.X == 0 && m.CameraPos.Y == 0 {
		m.CameraPos = m.PartyPos
	}

	deadzoneX := max(2, viewW/6)
	deadzoneY := max(2, viewH/6)

	dx := m.PartyPos.X - m.CameraPos.X
	dy := m.PartyPos.Y - m.CameraPos.Y

	if dx > deadzoneX {
		m.CameraPos.X = m.PartyPos.X - deadzoneX
	} else if dx < -deadzoneX {
		m.CameraPos.X = m.PartyPos.X + deadzoneX
	}

	if dy > deadzoneY {
		m.CameraPos.Y = m.PartyPos.Y - deadzoneY
	} else if dy < -deadzoneY {
		m.CameraPos.Y = m.PartyPos.Y + deadzoneY
	}

	startX := m.CameraPos.X - viewW/2
	startY := m.CameraPos.Y - viewH/2

	if startX+viewW > m.MapWidth {
		startX = m.MapWidth - viewW
	}
	if startY+viewH > m.MapHeight {
		startY = m.MapHeight - viewH
	}
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	return startX, startY
}

func (m Model) renderMap(viewW, viewH int) string {
	if viewW <= 0 || viewH <= 0 {
		return ""
	}
	if m.InTown {
		return m.renderTownHub(viewW, viewH)
	}

	biome := getBiome(m.Floor)
	customWall := lipgloss.NewStyle().Foreground(biome.WallColor).Render("#")
	customFloor := lipgloss.NewStyle().Foreground(biome.FloorColor).Render(string(biome.FloorRune))

	startX, startY := m.updateCamera(viewW, viewH)

	out := make([]string, viewH)
	for y := 0; y < viewH; y++ {
		mapY := startY + y
		var sb strings.Builder
		for x := 0; x < viewW; x++ {
			mapX := startX + x
			if mapX >= m.MapWidth || mapY >= m.MapHeight {
				sb.WriteString(" ")
				continue
			}
			if mapX == m.PartyPos.X && mapY == m.PartyPos.Y {
				sb.WriteString(partyStyle.Render("@"))
				continue
			}
			if !m.Explored[mapY][mapX] {
				sb.WriteString(" ")
				continue
			}
			pos := Point{mapX, mapY}
			if pack, exists := m.Packs[pos]; exists {
				first := pack.GetFirstLiving()
				if first != nil {
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(first.Color)).Bold(true).Render(string(first.Glyph)))
				} else {
					sb.WriteString(customFloor)
				}
				continue
			}
			switch m.Grid[mapY][mapX] {
			case TileWall:
				sb.WriteString(customWall)
			case TileFloor:
				sb.WriteString(customFloor)
			case TileChest:
				sb.WriteString(goldStyle.Render("$"))
			case TileRelic:
				sb.WriteString(relicTileStyle.Render("*"))
			case TileAltar:
				sb.WriteString(altarStyle.Render("_"))
			case TileFountain:
				sb.WriteString(fountStyle.Render("~"))
			case TileTrappedChest:
				sb.WriteString(trappedChestStyle.Render("T"))
			case TileBarrel:
				sb.WriteString(barrelStyle.Render("o"))
			case TileEvent:
				sb.WriteString(eventTileStyle.Render("?"))
			case TileStairs:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render(">"))
			case TileExit:
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render("<"))
			default:
				sb.WriteRune(rune(m.Grid[mapY][mapX]))
			}
		}
		out[y] = sb.String()
	}
	return strings.Join(out, "\n")
}
