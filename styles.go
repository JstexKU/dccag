package main

import (
	"github.com/charmbracelet/lipgloss"
)

// --- Стили интерфейса ---
var (
	partyStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	dangerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	healStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	accentStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	goldStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	questStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	subtleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	townArtStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Bold(true)
	altarStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("165")).Bold(true)
	fountStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	trappedChestStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	barrelStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("172")).Bold(true)
	potionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true)
	relicTileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	eventTileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)

	stressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("135")).Bold(true)
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	fireStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)

	heroCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	heroCardActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("226")).
			Padding(0, 1)

	heroCardDead = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)

	statsBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(1, 2).
			Align(lipgloss.Left)

	menuBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 3).
			Align(lipgloss.Center)
)
