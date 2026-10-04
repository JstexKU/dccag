package main

import (
	"github.com/charmbracelet/lipgloss"
)

func getBiome(floor int) BiomeConfig {
	cycle := (floor - 1) % 11
	switch cycle {
	case 0:
		return BiomeConfig{Name: BiomeCatacombs, WallColor: lipgloss.Color("240"), FloorColor: lipgloss.Color("236"), FloorRune: '·', EnvHazardKey: "hazard.catacombs"}
	case 1:
		return BiomeConfig{Name: BiomeGrotto, WallColor: lipgloss.Color("31"), FloorColor: lipgloss.Color("24"), FloorRune: '~', EnvHazardKey: "hazard.grotto"}
	case 2:
		return BiomeConfig{Name: BiomeInferno, WallColor: lipgloss.Color("124"), FloorColor: lipgloss.Color("52"), FloorRune: '≈', EnvHazardKey: "hazard.inferno"}
	case 3:
		return BiomeConfig{Name: BiomeCrystal, WallColor: lipgloss.Color("141"), FloorColor: lipgloss.Color("54"), FloorRune: '◊', EnvHazardKey: "hazard.crystal"}
	case 4:
		return BiomeConfig{Name: BiomeDeadwood, WallColor: lipgloss.Color("28"), FloorColor: lipgloss.Color("235"), FloorRune: '♣', EnvHazardKey: "hazard.deadwood"}
	case 5:
		return BiomeConfig{Name: BiomeFungal, WallColor: lipgloss.Color("100"), FloorColor: lipgloss.Color("58"), FloorRune: '§', EnvHazardKey: "hazard.fungal"}
	case 6:
		return BiomeConfig{Name: BiomeArchives, WallColor: lipgloss.Color("178"), FloorColor: lipgloss.Color("94"), FloorRune: '≡', EnvHazardKey: "hazard.archives"}
	case 7:
		return BiomeConfig{Name: BiomeMines, WallColor: lipgloss.Color("238"), FloorColor: lipgloss.Color("233"), FloorRune: '•', EnvHazardKey: "hazard.mines"}
	case 8:
		return BiomeConfig{Name: BiomeSanctuary, WallColor: lipgloss.Color("161"), FloorColor: lipgloss.Color("53"), FloorRune: '†', EnvHazardKey: "hazard.sanctuary"}
	case 9:
		return BiomeConfig{Name: BiomeAstral, WallColor: lipgloss.Color("69"), FloorColor: lipgloss.Color("17"), FloorRune: '¤', EnvHazardKey: "hazard.astral"}
	default:
		return BiomeConfig{Name: BiomeAbyss, WallColor: lipgloss.Color("89"), FloorColor: lipgloss.Color("232"), FloorRune: '×', EnvHazardKey: "hazard.abyss"}
	}
}

type BiomeConfig struct {
	Name         BiomeType
	WallColor    lipgloss.Color
	FloorColor   lipgloss.Color
	FloorRune    rune
	EnvHazardKey string
}
