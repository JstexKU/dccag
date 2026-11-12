package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Общие части карточек заведений: шапка, стиль строки с курсором, цена и значок
// «✓ …» после автоматического действия.

// renderShopTop — три верхние строки карточки: заголовок со счётом золота (и дополнительными
// полями через « | »), подсказка по клавишам либо баннер автопилота и разделитель.
func (m Model) renderShopTop(title, controlsKey, summary string, innerW int, extras ...string) string {
	var sb strings.Builder
	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(title)
	sb.WriteString(fmt.Sprintf("%s  |  💰 %s", titleStr, goldStyle.Render(fmt.Sprintf("%dG", m.Gold))))
	for _, extra := range extras {
		sb.WriteString("  |  " + extra)
	}
	sb.WriteString("\n")

	if m.ManualMode {
		controls := T(m.Lang, controlsKey)
		sb.WriteString(subtleStyle.Render(shortenItemName(controls, innerW)) + "\n")
	} else {
		sb.WriteString(m.renderAutoBanner(summary, innerW) + "\n")
	}
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")
	return sb.String()
}

// shopRowStyle — маркер и стиль названия строки списка; у строки под курсором они выделены.
func shopRowStyle(selected bool) (string, lipgloss.Style) {
	marker := "  "
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	if selected {
		marker = accentStyle.Render("▶ ")
		nameStyle = nameStyle.Bold(true).Underline(true)
	}
	return marker, nameStyle
}

// costText — цена в золоте: обычным цветом, если хватает денег, приглушённым — если нет.
func costText(cost int, affordable bool) string {
	text := fmt.Sprintf("%dG", cost)
	if affordable {
		return goldStyle.Render(text)
	}
	return subtleStyle.Render(text)
}

// autoPicked — автопилот только что выбрал строку i (в ручном режиме значка нет).
func (m Model) autoPicked(cursor, i int, summary string) bool {
	return !m.ManualMode && cursor == i && summary != ""
}

// badgeIf — значок «✓ <текст по ключу>» после выбранной строки; для остальных пусто.
func (m Model) badgeIf(chosen bool, key string) string {
	if !chosen {
		return ""
	}
	return " " + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ "+T(m.Lang, key))
}
