package main

type LayoutMode int

const (
	LayoutLandscape LayoutMode = iota
	LayoutPortraitWide
	LayoutPortrait
)

type CardMode int

const (
	CardWide CardMode = iota
	CardMedium
	CardCompact
)

func detectLayout(termW, termH int) LayoutMode {
	if termW <= 0 || termH <= 0 {
		return LayoutLandscape
	}
	ratio := float64(termW) / float64(termH)
	switch {
	case ratio >= 1.4 || termW >= 140:
		return LayoutLandscape
	case ratio >= 0.75 && termW >= 48:
		return LayoutPortraitWide
	default:
		return LayoutPortrait
	}
}

func detectCardMode(innerW int) CardMode {
	switch {
	case innerW >= 34:
		return CardWide
	case innerW >= 24:
		return CardMedium
	default:
		return CardCompact
	}
}

// cardContentHeight возвращает точное число строк содержимого внутри карточки (без рамки)
func cardContentHeight(mode CardMode) int {
	switch mode {
	case CardWide:
		return 10
	case CardMedium:
		return 8
	default:
		return 6
	}
}

// cardOuterHeight возвращает полную внешнюю высоту карточки с учетом рамки (top + bottom = +2 строки)
func cardOuterHeight(mode CardMode) int {
	return cardContentHeight(mode) + 2
}

func landscapeCardsPerRow(usableW int) int {
	switch {
	case usableW >= 115:
		return 6
	case usableW >= 74:
		return 3
	default:
		return 2
	}
}
