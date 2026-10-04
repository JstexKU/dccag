package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var globalDebugReport = DebugReportData{
	ClassDeaths:        make(map[string]int),
	DeathCauses:        make(map[string]int),
	DeathsByFloor:      make(map[int]int),
	GoldSpentBreakdown: make(map[string]int),
	AverageAtk:         make(map[string]float64),
}

type DebugReportData struct {
	TotalRuns          int                `json:"total_runs"`
	MaxFloorReached    int                `json:"max_floor_reached"`
	TotalGoldSpent     int                `json:"total_gold_spent"`
	GoldSpentBreakdown map[string]int     `json:"gold_spent_breakdown"`
	ClassDeaths        map[string]int     `json:"class_deaths"`
	DeathCauses        map[string]int     `json:"death_causes"`
	DeathsByFloor      map[int]int        `json:"deaths_by_floor"`
	FleeAttempts       int                `json:"flee_attempts"`
	FleeSuccesses      int                `json:"flee_successes"`
	DamageGuarded      int                `json:"damage_guarded_by_tanks"`
	EnrageProcs        int                `json:"enrage_triggered_count"`
	AverageAtk         map[string]float64 `json:"average_atk"`
}

// Накопители для честного среднего значения атаки по классам за все забеги.
var (
	atkSum   = map[string]float64{}
	atkCount = map[string]int{}
)

func (m *Model) accumulateDebugReport() {
	// Защита от двойного учёта рана.
	if m.RunCounted {
		return
	}
	m.RunCounted = true

	globalDebugReport.TotalRuns++
	if m.Floor > globalDebugReport.MaxFloorReached {
		globalDebugReport.MaxFloorReached = m.Floor
	}

	globalDebugReport.TotalGoldSpent += (m.Stats.TotalGoldEarned - m.Gold)

	for _, fallen := range m.Stats.FallenHeroes {
		globalDebugReport.DeathCauses[fallen.Cause]++
		globalDebugReport.ClassDeaths[string(fallen.Class)]++
		globalDebugReport.DeathsByFloor[fallen.Floor]++
	}

	for _, h := range m.Party {
		if !h.IsDead {
			c := string(h.Class)
			atkSum[c] += float64(h.TotalAtk())
			atkCount[c]++
			globalDebugReport.AverageAtk[c] = atkSum[c] / float64(atkCount[c])
		}
	}
}

func saveDebugReportToFile() {
	if !*debugReportFlag {
		return
	}
	fileData, err := json.MarshalIndent(globalDebugReport, "", "  ")
	if err == nil {
		_ = os.WriteFile("dccag_report.json", fileData, 0o644)
		fmt.Println("\n[DEBUG] Telemetry report saved to: dccag_report.json")
	}
}
