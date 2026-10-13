package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLibraryScreen(viewW, viewH int) string {
	innerW := max(50, viewW)
	halfW := max(22, (innerW-5)/2)
	rightW := max(26, innerW-halfW-5)

	var sb strings.Builder

	// 1. Верхний заголовок и счетчик коллекции
	titleStr := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("📖 " + T(m.Lang, "library.header"))
	unlockedCount := 0
	if m.Legacy.UnlockedLore != nil {
		unlockedCount = m.Legacy.UnlockedLore.TotalUnlocked()
	}
	totalCount := len(AllLoreVolumes)
	progressStr := goldStyle.Render(fmt.Sprintf("%s: %d / %d", T(m.Lang, "library.progress"), unlockedCount, totalCount))

	sb.WriteString(fmt.Sprintf("%s  |  %s\n", titleStr, progressStr))

	// 2. Подсказка по клавишам управления
	controlsStr := T(m.Lang, "library.controls")
	sb.WriteString(subtleStyle.Render(shortenItemName(controlsStr, innerW)) + "\n")
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	// 3. Таб-бар 6 разделов библиотеки
	catNames := []string{
		T(m.Lang, "library.cat.mastery"),
		T(m.Lang, "library.cat.genesis"),
		T(m.Lang, "library.cat.ethnography"),
		T(m.Lang, "library.cat.belles"),
		T(m.Lang, "library.cat.diaries"),
		T(m.Lang, "library.cat.bestiary"),
	}
	var tabs []string
	for i, name := range catNames {
		tabNum := fmt.Sprintf("[%d]", i+1)
		if LoreCategory(i) == m.LibraryState.ActiveCategory {
			tabs = append(tabs, accentStyle.Render(tabNum+" "+name))
		} else {
			tabs = append(tabs, subtleStyle.Render(tabNum+" "+name))
		}
	}
	sb.WriteString(strings.Join(tabs, "  ") + "\n")
	sb.WriteString(subtleStyle.Render(strings.Repeat("─", innerW)) + "\n")

	// 4. Формирование списков
	availRows := max(6, viewH-8)
	volumes := GetVolumesByCategory(m.LibraryState.ActiveCategory)

	// --- ЛЕВАЯ КОЛОНКА: СПИСОК ТОМОВ ---
	var leftLines []string
	leftHeader := lipgloss.NewStyle().Bold(true).Render("📜 " + T(m.Lang, "library.volumes_title"))
	if !m.LibraryState.ViewingDetails {
		leftHeader = accentStyle.Render("▶ 📜 " + T(m.Lang, "library.volumes_title"))
	}
	leftLines = append(leftLines, leftHeader, subtleStyle.Render(strings.Repeat("─", halfW)))

	if len(volumes) == 0 {
		leftLines = append(leftLines, subtleStyle.Render("  ("+T(m.Lang, "ui.none")+")"))
	} else {
		pageSize := max(3, availRows-3)
		m.LibraryState.AdjustScroll(len(volumes), pageSize)

		if m.LibraryState.ScrollOffset > 0 {
			leftLines = append(leftLines, subtleStyle.Render(T(m.Lang, "ui.scroll_more_above", m.LibraryState.ScrollOffset)))
		}

		endIdx := min(len(volumes), m.LibraryState.ScrollOffset+pageSize)
		for i := m.LibraryState.ScrollOffset; i < endIdx; i++ {
			vol := volumes[i]
			isUnlocked := m.Legacy.UnlockedLore != nil && m.Legacy.UnlockedLore.IsUnlocked(vol.ID)

			marker := "  "
			itemStyle := subtleStyle
			if m.LibraryState.Cursor == i {
				marker = accentStyle.Render("▶ ")
				itemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
			}

			titleDisplay := T(m.Lang, "library.locked_tome")
			statusTag := subtleStyle.Render("[?]")
			if isUnlocked {
				titleDisplay = T(m.Lang, vol.TitleKey)
				statusTag = healStyle.Render("[✓]")
			}

			cleanTitle := shortenItemName(titleDisplay, max(8, halfW-9))
			leftLines = append(leftLines, fmt.Sprintf("%s%s %s", marker, statusTag, itemStyle.Render(padRight(cleanTitle, halfW-9))))
		}

		if endIdx < len(volumes) {
			leftLines = append(leftLines, subtleStyle.Render(T(m.Lang, "ui.scroll_more_below", len(volumes)-endIdx)))
		}
	}

	// --- ПРАВАЯ КОЛОНКА: ОКНО ЧТЕНИЯ ---
	var rightLines []string
	rightHeader := lipgloss.NewStyle().Bold(true).Render("📖 " + T(m.Lang, "library.reading_room"))
	if m.LibraryState.ViewingDetails {
		rightHeader = accentStyle.Render("▶ 📖 " + T(m.Lang, "library.reading_room"))
	}
	rightLines = append(rightLines, rightHeader, subtleStyle.Render(strings.Repeat("─", rightW)))

	if len(volumes) > 0 && m.LibraryState.Cursor >= 0 && m.LibraryState.Cursor < len(volumes) {
		curVol := volumes[m.LibraryState.Cursor]
		isUnlocked := m.Legacy.UnlockedLore != nil && m.Legacy.UnlockedLore.IsUnlocked(curVol.ID)

		if !isUnlocked {
			rightLines = append(rightLines,
				"",
				subtleStyle.Render("  "+T(m.Lang, "library.not_yet_found")),
				subtleStyle.Render("  "+T(m.Lang, "library.explore_hint")),
			)
		} else {
			// Название и автор
			titleLine := lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Bold(true).Render(T(m.Lang, curVol.TitleKey))
			authorLine := subtleStyle.Render(T(m.Lang, "library.author_prefix") + ": " + T(m.Lang, curVol.AuthorKey))
			rightLines = append(rightLines, titleLine, authorLine, subtleStyle.Render(strings.Repeat("┄", rightW)))

			// Цитата курсивом / акцентом
			excerptLine := lipgloss.NewStyle().Foreground(lipgloss.Color("216")).Italic(true).Render("\"" + T(m.Lang, curVol.ExcerptKey) + "\"")
			rightLines = append(rightLines, excerptLine, "")

			// Полный текст с учетом TextScrollY
			fullText := T(m.Lang, curVol.ContentKey)
			wrapped := wordWrap(fullText, rightW-2)
			startLine := min(m.LibraryState.TextScrollY, max(0, len(wrapped)-1))

			maxTextLines := max(2, availRows-len(rightLines)-3)
			for i := startLine; i < min(len(wrapped), startLine+maxTextLines); i++ {
				rightLines = append(rightLines, "  "+wrapped[i])
			}

			// Индикатор бонуса тома
			if curVol.Bonus.Type != LoreBonusNone && curVol.Bonus.DescKey != "" {
				rightLines = append(rightLines, "", healStyle.Render("★ "+T(m.Lang, curVol.Bonus.DescKey)))
			}
		}
	}

	maxRows := max(len(leftLines), len(rightLines))
	for len(leftLines) < maxRows {
		leftLines = append(leftLines, strings.Repeat(" ", halfW))
	}
	for len(rightLines) < maxRows {
		rightLines = append(rightLines, strings.Repeat(" ", rightW))
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(halfW).Render(strings.Join(leftLines, "\n")),
		" │ ",
		lipgloss.NewStyle().Width(rightW).Render(strings.Join(rightLines, "\n")),
	))

	return padTownLines(sb.String(), viewW, viewH)
}

// wordWrap разбивает длинные абзацы текста по границам слов под ширину колонки
func wordWrap(text string, limit int) []string {
	if limit <= 4 {
		limit = 4
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	curr := words[0]

	for _, w := range words[1:] {
		if len([]rune(curr))+1+len([]rune(w)) <= limit {
			curr += " " + w
		} else {
			lines = append(lines, curr)
			curr = w
		}
	}
	lines = append(lines, curr)
	return lines
}