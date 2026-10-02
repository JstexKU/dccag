package main

import (
	"math/rand"
	"time"
)

// rng — единый источник случайности игры.
// По умолчанию сидируется временем; флаг -seed делает генерацию воспроизводимой.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// seedRNG заменяет генератор на детерминированный с указанным зерном.
func seedRNG(seed int64) {
	rng = rand.New(rand.NewSource(seed))
}
