package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func padRightTruncate(s string, targetWidth int) string {
	if targetWidth <= 0 {
		return ""
	}
	w := runewidth.StringWidth(stripANSI(s))
	if w > targetWidth {
		return truncatePlain(s, targetWidth)
	}
	if w == targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

func stripANSI(s string) string {
	var sb strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if c == 'm' {
				inEsc = false
			}
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

func truncatePlain(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	plain := stripANSI(s)
	if runewidth.StringWidth(plain) <= maxW {
		return plain
	}
	var sb strings.Builder
	curW := 0
	target := maxW - 1
	for _, r := range plain {
		rw := runewidth.RuneWidth(r)
		if curW+rw > target {
			break
		}
		sb.WriteRune(r)
		curW += rw
	}
	sb.WriteString("…")
	return sb.String()
}

func truncateLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// normalizeLines добивает или урезает текст строго до targetCount строк одинаковой ширины innerWidth
func normalizeLines(content string, innerWidth, targetCount int) string {
	rawLines := strings.Split(content, "\n")
	res := make([]string, targetCount)
	for i := 0; i < targetCount; i++ {
		if i < len(rawLines) {
			res[i] = padRightTruncate(rawLines[i], innerWidth)
		} else {
			res[i] = strings.Repeat(" ", innerWidth)
		}
	}
	return strings.Join(res, "\n")
}

func renderBar(current, max int, totalBars int, filledColor, emptyColor lipgloss.Color) string {
	if totalBars < 2 {
		totalBars = 2
	}
	if max <= 0 {
		max = 1
	}
	ratio := float64(current) / float64(max)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(totalBars))
	empty := totalBars - filled
	fStr := lipgloss.NewStyle().Foreground(filledColor).Render(strings.Repeat("█", filled))
	eStr := lipgloss.NewStyle().Foreground(emptyColor).Render(strings.Repeat("░", empty))
	return "[" + fStr + eStr + "]"
}

func shortenItemName(name string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	w := runewidth.StringWidth(name)
	if w <= maxLen {
		return name
	}
	runes := []rune(name)
	res := ""
	curW := 0
	for _, r := range runes {
		rw := runewidth.RuneWidth(r)
		if curW+rw >= maxLen {
			break
		}
		res += string(r)
		curW += rw
	}
	return res + "…"
}

func padRight(s string, targetWidth int) string {
	w := runewidth.StringWidth(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}
