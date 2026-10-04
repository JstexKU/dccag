package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// version подставляется при релизной сборке: -ldflags "-X main.version=v1.2.3".
// В исходниках хранится номер без префикса «v».
var version = "3.0.0"

// displayVersion возвращает версию для интерфейса в виде «v2.9.0»:
// релизная сборка подставляет тег целиком, локальная — только номер.
func displayVersion() string {
	return "v" + strings.TrimPrefix(version, "v")
}

var (
	showVersionFlag = flag.Bool("version", false, "Show version and exit")
	noSaveFlag      = flag.Bool("no-save", false, "Do not read or write the save file")
	resetSaveFlag   = flag.Bool("reset-save", false, "Delete the save file and exit")
	seedFlag        = flag.Int64("seed", 0, "Random seed for reproducible runs (0 = random)")
	manualFlag      = flag.Bool("manual", false, "Start in manual control mode (toggle with M)")
	debugReportFlag = flag.Bool("report", false, "Collect telemetry and write dccag_report.json")
)

func main() {
	// Фиксация 1 потока для стабильной работы на 32-битных архитектурах и в iSH
	runtime.GOMAXPROCS(1)

	flag.Parse()

	if *showVersionFlag {
		fmt.Println("dccag", displayVersion())
		return
	}
	if *noSaveFlag {
		saveEnabled = false
	}
	if *resetSaveFlag {
		if path, err := savePath(); err == nil {
			_ = os.Remove(path)
			fmt.Println("Save file removed:", path)
		}
		return
	}
	if *seedFlag != 0 {
		seedRNG(*seedFlag)
	}

	m := initialModel()
	if sd, ok := loadSave(); ok {
		m = initialModelWith(sd.Legacy, sd.Hero)
		m.Lang = sd.Lang
		m.Tactics = sd.Tactics
		m.relocalizeStart()
	}
	m.ManualMode = *manualFlag

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()

	// Выход из игры завершает экспедицию: наследие и настройки сохраняются.
	if fm, ok := finalModel.(Model); ok {
		persistState(fm.Lang, legacyAfterRun(fm), fm.Tactics, fm.Blueprint)
		if *debugReportFlag {
			fm.accumulateDebugReport()
			saveDebugReportToFile()
		}
	}

	if err != nil {
		fmt.Printf("Startup error: %v\n", err)
		os.Exit(1)
	}
}
