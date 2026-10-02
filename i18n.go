package main

import (
	"fmt"
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

func TVerb(lang Language, gender Gender, maleRU, femaleRU, eng string) string {
	if lang == LangEN {
		return eng
	}
	if gender == GenderFemale {
		return femaleRU
	}
	return maleRU
}

func TranslateEnum(lang Language, prefix string, val string) string {
	key := fmt.Sprintf("%s.%s", prefix, val)
	return T(lang, key)
}