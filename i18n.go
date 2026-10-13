package main

import (
	"fmt"
)

type Language string

const (
	LangRU Language = "ru"
	LangEN Language = "en"
	LangBE Language = "be"
)

// SupportedLanguages содержит список всех поддерживаемых языков в порядке переключения.
var SupportedLanguages = []Language{LangRU, LangEN, LangBE}

// NextLanguage переключает язык циклически: RU -> EN -> BE -> RU.
func NextLanguage(cur Language) Language {
	for i, l := range SupportedLanguages {
		if l == cur {
			return SupportedLanguages[(i+1)%len(SupportedLanguages)]
		}
	}
	return LangRU
}

func T(lang Language, key string, args ...any) string {
	var dict map[string]string
	switch lang {
	case LangEN:
		dict = dictEN
	case LangBE:
		dict = dictBE
	default:
		dict = dictRU
	}

	val, ok := dict[key]
	if !ok {
		// Каскадный fallback при отсутствии перевода:
		// для BE и EN сначала пробуем взять из dictRU, затем dictEN
		if lang == LangBE {
			val, ok = dictRU[key]
			if !ok {
				val, ok = dictEN[key]
			}
		} else if lang == LangEN {
			val, ok = dictRU[key]
		}

		if !ok {
			return key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// TVerb возвращает глагол в нужном роде по ключу: verb.<ключ>.m (мужской) и verb.<ключ>.f
// (женский).
func TVerb(lang Language, gender Gender, key string) string {
	suffix := "m"
	if gender == GenderFemale {
		suffix = "f"
	}
	return T(lang, "verb."+key+"."+suffix)
}

func TranslateEnum(lang Language, prefix string, val string) string {
	key := fmt.Sprintf("%s.%s", prefix, val)
	return T(lang, key)
}