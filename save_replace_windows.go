//go:build windows

package main

import "golang.org/x/sys/windows"

func replaceSaveFile(tmp, dst string) error {
	// Unlike os.Rename, MoveFileEx with REPLACE_EXISTING can replace an
	// existing file on Windows while retaining the temp-file write strategy.
	return windows.MoveFileEx(tmp, dst, windows.MOVEFILE_REPLACE_EXISTING)
}
