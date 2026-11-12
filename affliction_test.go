package main

import (
	"strings"
	"testing"
)

// Новый герой начинает без состояния: AfflictionNone, а не пустая строка. Пустое значение
// ломало проверки «== AfflictionNone» (психоз не наступал) и ключ перевода «affliction.».
func TestNewHeroesStartWithoutAffliction(t *testing.T) {
	seedRNG(21)
	for _, class := range AllClasses {
		for _, race := range AllRaces {
			h := buildHero(class, race, "", GenderMale, 1, 0)
			if h.Affliction != AfflictionNone {
				t.Fatalf("%s/%s: Affliction=%q", class, race, h.Affliction)
			}
		}
	}
	bp := HeroBlueprint{Name: "Тест", Race: RaceHuman, Class: ClassWarrior}
	if h := createLeaderHero(bp, 0); h.Affliction != AfflictionNone {
		t.Fatalf("лидер: Affliction=%q", h.Affliction)
	}
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		key := "affliction." + string(AfflictionNone)
		if got := T(lang, key); got == key || got == "" {
			t.Errorf("%s: нет перевода %q", lang, key)
		}
	}
}

// Накопленный стресс у свежего героя приводит к добродетели или психозу.
func TestFreshHeroCanBreakUnderStress(t *testing.T) {
	seedRNG(22)
	m := initialModel()
	m.State = StatePlaying
	h := m.Party[1]
	h.Stress = 0
	for i := 0; i < 12 && h.Affliction == AfflictionNone; i++ {
		m.addStress(h, 100)
	}
	if h.Affliction == AfflictionNone {
		t.Fatal("после сильного стресса состояние героя не изменилось")
	}
}

// Предложения храма не содержат сырых ключей перевода.
func TestChurchOffersHaveNoRawKeys(t *testing.T) {
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		seedRNG(23)
		m := initialModel()
		m.Lang = lang
		m.State = StatePlaying
		m.Gold = 5000
		m.Party[0].Stress = 50
		m.Party[1].Stress = 90
		m.Party[1].Affliction = AfflictionParanoid

		for _, o := range m.GenerateChurchOffers() {
			if strings.Contains(o.Desc, "affliction.") {
				t.Errorf("%s: сырой ключ в описании %q", lang, o.Desc)
			}
		}
	}
}

// Алхимик не навязывает зелье от стресса здоровым героям и называет зелья по-человечески.
func TestAlchemistPotionOffers(t *testing.T) {
	for _, lang := range []Language{LangRU, LangEN, LangBE} {
		seedRNG(24)
		m := initialModel()
		m.Lang = lang
		m.State = StatePlaying
		m.MaxFloor = 5
		for _, h := range m.Party {
			h.Stress = 0
			h.IsDowned = false
			h.IsDead = false
		}
		for _, o := range m.GenerateAlchemistOffers() {
			if o.Type != AlchemistActionPotion {
				continue
			}
			if o.PotType == PotionStress {
				t.Errorf("%s: зелье от стресса предложено герою %d без стресса", lang, o.HeroIdx)
			}
			if strings.Contains(o.ItemTitle, "(") || strings.Contains(o.ItemTitle, "potion.") {
				t.Errorf("%s: сырое название зелья %q", lang, o.ItemTitle)
			}
		}
	}
}
