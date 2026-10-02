package main

import "os"

// replaceSaveFile atomically replaces the destination when possible.
func replaceSaveFile(tmp, dst string) error {
	return os.Rename(tmp, dst)
}
