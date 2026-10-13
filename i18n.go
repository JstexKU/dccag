package main

import (
	"fmt"
)

type Language string

const (
	LangRU Language = "ru"
	LangEN Language = "en"
)

func T(lang Language, key string, args ...any) string {
	var dict map[string]string
	if lang == LangEN {
		dict = dictEN
	} else {
		dict = dictRU
	}

	val, ok := dict[key]
	if !ok {
		if lang == LangEN {
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
// (женский). В английском словаре оба варианта совпадают.
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
