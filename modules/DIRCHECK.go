package modules

import (
	"fmt"
	"os"
	"path/filepath"
)

var allowedRootDirs = map[string]bool{
	"logs":      true,
	"minecraft": true,
	"runtime":   true,
}

var allowedFiles = map[string]map[string]bool{
	"logs": {
		"launcher.log": true,
	},
}

func EnsureDirectoryStructure() {
	Log("[DIRCHECK] Ensuring directory structure...")

	baseDir := OSInfo.DataDir

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		Log(fmt.Sprintf("[DIRCHECK] Failed to create base dir: %v", err))
		return
	}

	for dir := range allowedRootDirs {
		path := filepath.Join(baseDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			Log(fmt.Sprintf("[DIRCHECK] Failed to create %s: %v", dir, err))
		} else {
			Log(fmt.Sprintf("[DIRCHECK] Directory ensured: %s", dir))
		}
	}

	cleanDirectory(baseDir, allowedRootDirs, allowedFiles)

	Log("[DIRCHECK] Directory structure check complete")
}

func cleanDirectory(baseDir string, allowedDirs map[string]bool, allowedFiles map[string]map[string]bool) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		Log(fmt.Sprintf("[DIRCHECK] Failed to read %s: %v", baseDir, err))
		return
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			if !allowedDirs[name] {
				fullPath := filepath.Join(baseDir, name)
				if err := os.RemoveAll(fullPath); err != nil {
					Log(fmt.Sprintf("[DIRCHECK] Failed to remove dir %s: %v", name, err))
				} else {
					Log(fmt.Sprintf("[DIRCHECK] Removed unauthorized directory: %s", name))
				}
			} else {
				cleanSubDirectory(filepath.Join(baseDir, name), name, allowedFiles)
			}
		} else {
			if !isAllowedRootFile(name) {
				fullPath := filepath.Join(baseDir, name)
				if err := os.Remove(fullPath); err != nil {
					Log(fmt.Sprintf("[DIRCHECK] Failed to remove file %s: %v", name, err))
				} else {
					Log(fmt.Sprintf("[DIRCHECK] Removed unauthorized file: %s", name))
				}
			}
		}
	}
}

func cleanSubDirectory(dirPath string, parentDir string, allowedFiles map[string]map[string]bool) {
	allowed, exists := allowedFiles[parentDir]
	if !exists {
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		Log(fmt.Sprintf("[DIRCHECK] Failed to read %s: %v", dirPath, err))
		return
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			fullPath := filepath.Join(dirPath, name)
			if err := os.RemoveAll(fullPath); err != nil {
				Log(fmt.Sprintf("[DIRCHECK] Failed to remove nested dir %s: %v", name, err))
			} else {
				Log(fmt.Sprintf("[DIRCHECK] Removed unauthorized nested directory: %s/%s", parentDir, name))
			}
		} else if !allowed[name] {
			fullPath := filepath.Join(dirPath, name)
			Log(fmt.Sprintf("[DIRCHECK] Found unauthorized file in %s: %s", parentDir, name))
			if err := os.Remove(fullPath); err != nil {
				Log(fmt.Sprintf("[DIRCHECK] Failed to remove file %s/%s: %v", parentDir, name, err))
			} else {
				Log(fmt.Sprintf("[DIRCHECK] Removed unauthorized file: %s/%s", parentDir, name))
			}
		} else {
			Log(fmt.Sprintf("[DIRCHECK] Allowed file kept: %s/%s", parentDir, name))
		}
	}
}

func isAllowedRootFile(name string) bool {
	switch name {
	case ".storage.aes", "launcher", "launcher.exe", "launcher.pid", "launcher.new", "auth":
		return true
	}
	return false
}

func ResetAllData() error {
	Log("[DIRCHECK] Resetting all game data...")
	baseDir := OSInfo.DataDir
	for dir := range allowedRootDirs {
		path := filepath.Join(baseDir, dir)
		if err := os.RemoveAll(path); err != nil {
			Log(fmt.Sprintf("[DIRCHECK] Failed to remove %s: %v", dir, err))
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
		Log(fmt.Sprintf("[DIRCHECK] Removed: %s", dir))
	}
	// Recreate empty directories
	EnsureDirectoryStructure()
	Log("[DIRCHECK] Reset complete, directories recreated")
	return nil
}
