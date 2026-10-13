package main

import (
	"strings"
	"testing"
)

// TestLoreDatabaseCompleteness проверяет целостность реестра фолиантов и наличие переводов
func TestLoreDatabaseCompleteness(t *testing.T) {
	if len(AllLoreVolumes) == 0 {
		t.Fatal("база фолиантов AllLoreVolumes пуста")
	}

	seenIDs := make(map[string]bool)
	dicts := map[Language]map[string]string{
		LangRU: dictRU,
		LangEN: dictEN,
		LangBE: dictBE,
	}

	for _, v := range AllLoreVolumes {
		if v.ID == "" {
			t.Fatal("обнаружен фолиант с пустым ID")
		}
		if seenIDs[v.ID] {
			t.Fatalf("дубликат ID фолианта: %s", v.ID)
		}
		seenIDs[v.ID] = true

		if v.Category < LoreCatMastery || v.Category > LoreCatBestiary {
			t.Fatalf("фолиант %s имеет недопустимую категорию %v", v.ID, v.Category)
		}

		// Проверка ключей локализации во всех словарях
		keysToCheck := []string{v.TitleKey, v.AuthorKey, v.ExcerptKey, v.ContentKey}
		if v.Bonus.Type != LoreBonusNone && v.Bonus.DescKey != "" {
			keysToCheck = append(keysToCheck, v.Bonus.DescKey)
		}

		for _, k := range keysToCheck {
			if k == "" {
				t.Fatalf("у фолианта %s пустой ключ локализации", v.ID)
			}
			for lang, dict := range dicts {
				val, ok := dict[k]
				if !ok || val == "" {
					t.Errorf("[%s] отсутствует перевод для ключа %q (фолиант %s)", lang, k, v.ID)
				}
			}
		}
	}
}

// TestLoreTileStepAndBonus проверяет наступление на клетку TileLore и начисление бонусов
func TestLoreTileStepAndBonus(t *testing.T) {
	seedRNG(101)
	m := initialModel()
	m.State = StatePlaying
	m.ManualMode = true

	target := Point{X: m.PartyPos.X + 1, Y: m.PartyPos.Y}
	if target.X >= m.MapWidth {
		target = Point{X: m.PartyPos.X - 1, Y: m.PartyPos.Y}
	}
	m.Grid[target.Y][target.X] = TileLore

	initialFound := m.Stats.LoreFound
	unlockedBefore := m.Legacy.UnlockedLore.TotalUnlocked()

	// Герои получают немного стресса для проверки лечения
	for _, h := range m.Party {
		h.Stress = 40
	}

	m.moveTo(target)

	if m.Grid[target.Y][target.X] != TileFloor {
		t.Fatal("тайл TileLore должен заменяться на TileFloor после изучения")
	}
	if m.Stats.LoreFound != initialFound+1 {
		t.Fatalf("счетчик LoreFound не увеличился: было %d, стало %d", initialFound, m.Stats.LoreFound)
	}
	if m.Legacy.UnlockedLore.TotalUnlocked() <= unlockedBefore {
		t.Fatal("новый фолиант должен быть добавлен в UnlockedLore")
	}
}

// TestLibraryKeyNavigation тестирует хоткеи навигации и переключение экранов
func TestLibraryKeyNavigation(t *testing.T) {
	seedRNG(102)
	m := initialModel()
	m.State = StatePlaying

	// 1. Открытие экрана по [B]
	press(&m, "b")
	if m.State != StateLibrary {
		t.Fatalf("клавиша B должна открывать StateLibrary, текущее состояние: %v", m.State)
	}

	// 2. Переключение вкладок цифрами 1..6
	press(&m, "2")
	if m.LibraryState.ActiveCategory != LoreCatGenesis {
		t.Fatalf("клавиша 2 должна активировать LoreCatGenesis, текущая: %v", m.LibraryState.ActiveCategory)
	}
	press(&m, "5")
	if m.LibraryState.ActiveCategory != LoreCatDiaries {
		t.Fatalf("клавиша 5 должна активировать LoreCatDiaries, текущая: %v", m.LibraryState.ActiveCategory)
	}

	// 3. Переключение Tab / Shift+Tab
	press(&m, "tab")
	if m.LibraryState.ActiveCategory != LoreCatBestiary {
		t.Fatalf("Tab должен переключать на LoreCatBestiary, текущая: %v", m.LibraryState.ActiveCategory)
	}

	// 4. Скролл и выбор книги
	press(&m, "down")
	if m.LibraryState.Cursor != 1 {
		t.Fatalf("стрелка вниз должна сдвигать курсор, текущий: %d", m.LibraryState.Cursor)
	}

	// 5. Переход к детальному чтению по Enter
	press(&m, "enter")
	if !m.LibraryState.ViewingDetails {
		t.Fatal("Enter должен переводить фокус на чтение текста (ViewingDetails=true)")
	}

	// 6. Выход из чтения назад к списку по Esc
	press(&m, "esc")
	if m.LibraryState.ViewingDetails || m.State != StateLibrary {
		t.Fatal("Esc в режиме чтения должен возвращать фокус к списку")
	}

	// 7. Выход из Библиотеки в игру по Esc
	press(&m, "esc")
	if m.State != StatePlaying {
		t.Fatalf("Esc в списке должен закрывать Библиотеку, текущее состояние: %v", m.State)
	}
}

// TestLoreSavePersistence проверяет сохранение и загрузку UnlockedLore через save.json
func TestLoreSavePersistence(t *testing.T) {
	useTempSave(t)
	seedRNG(103)
	m := initialModel()

	testID := "mastery_tank_1"
	m.Legacy.UnlockedLore.Unlock(testID)

	persistState(m.Lang, m.Legacy, m.Tactics, m.Blueprint, m.Companions...)

	loaded, ok := loadSave()
	if !ok {
		t.Fatal("не удалось загрузить сохранение")
	}
	if !loaded.Legacy.UnlockedLore.IsUnlocked(testID) {
		t.Fatalf("фолиант %s не восстановился из сохранения", testID)
	}
}

// TestLibraryRenderScreens проверяет отрисовку экрана Библиотеки во всех размерах и языках
func TestLibraryRenderScreens(t *testing.T) {
	seedRNG(104)
	sizes := [][2]int{{50, 20}, {80, 25}, {120, 36}, {160, 48}}
	langs := []Language{LangRU, LangEN, LangBE}

	m := initialModel()
	m.Legacy.UnlockedLore.Unlock("mastery_tank_1")
	m.Legacy.UnlockedLore.Unlock("genesis_vol_1")

	for _, lang := range langs {
		m.Lang = lang
		for _, sz := range sizes {
			m.TermWidth, m.TermHeight = sz[0], sz[1]
			out := m.renderLibraryScreen(sz[0], sz[1])
			if out == "" {
				t.Fatalf("[%s %v] экран Библиотеки пуст", lang, sz)
			}
			headerText := T(lang, "library.header")
			if !strings.Contains(out, headerText) {
				t.Fatalf("[%s %v] на экране нет заголовка %q", lang, sz, headerText)
			}
		}
	}
}