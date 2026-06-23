package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type LauncherSettings struct {
	Language string `json:"language"`
	RamMB    int    `json:"ram_mb"`
}

func DefaultSettings() LauncherSettings {
	return LauncherSettings{
		Language: "en",
		RamMB:    2048,
	}
}

func SettingsPath() string {
	return filepath.Join(OSInfo.DataDir, "settings.json")
}

func LoadSettings() LauncherSettings {
	path := SettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		Log(fmt.Sprintf("[SETTINGS] No settings file, using defaults: %v", err))
		def := DefaultSettings()
		SaveSettings(def)
		return def
	}

	var s LauncherSettings
	if err := json.Unmarshal(data, &s); err != nil {
		Log(fmt.Sprintf("[SETTINGS] Invalid JSON, resetting to defaults: %v", err))
		def := DefaultSettings()
		SaveSettings(def)
		return def
	}

	if !validateSettings(s) {
		Log("[SETTINGS] Invalid settings values, resetting to defaults")
		def := DefaultSettings()
		SaveSettings(def)
		return def
	}

	Log(fmt.Sprintf("[SETTINGS] Loaded: lang=%s ram=%dMB", s.Language, s.RamMB))
	return s
}

func SaveSettings(s LauncherSettings) error {
	if !validateSettings(s) {
		return fmt.Errorf("invalid settings")
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	path := SettingsPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	Log(fmt.Sprintf("[SETTINGS] Saved: lang=%s ram=%dMB", s.Language, s.RamMB))
	return nil
}

func validateSettings(s LauncherSettings) bool {
	if s.Language != "ru" && s.Language != "en" {
		return false
	}
	maxRam := GetMaxRamMB()
	if s.RamMB < 512 || s.RamMB > maxRam {
		return false
	}
	if s.RamMB%256 != 0 {
		return false
	}
	return true
}

func GetMaxRamMB() int {
	total := GetTotalRAMMB()
	if total <= 2048 {
		return 2048
	}
	max := total * 80 / 100
	max = (max / 256) * 256
	if max < 2048 {
		max = 2048
	}
	return max
}

func GetIconPath() string {
	return filepath.Join(OSInfo.DataDir, "runtime", "launcher-icon.png")
}

func ExtractIcon() {
	path := GetIconPath()
	if len(trayIcon) == 0 {
		return
	}
	if err := os.WriteFile(path, trayIcon, 0644); err != nil {
		Log(fmt.Sprintf("[ICON] Failed to extract: %v", err))
	} else {
		Log(fmt.Sprintf("[ICON] Extracted to %s (%d bytes)", path, len(trayIcon)))
	}
}

// Простейшая локализация для Go-кода
var currentLang string

func SetLanguage(lang string) {
	if lang != "ru" && lang != "en" {
		lang = "en"
	}
	currentLang = lang
}

func T(ru, en string) string {
	if currentLang == "ru" {
		return ru
	}
	return en
}
