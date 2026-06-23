package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var OSInfo struct {
	Name    string
	Version string
	Arch    string
	HomeDir string
	MainDir string
	DataDir string
}

func InitPaths() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}
	OSInfo.HomeDir = home

	exe, err := os.Executable()
	if err == nil {
		OSInfo.MainDir = filepath.Dir(exe)
	} else {
		OSInfo.MainDir = home
	}

	switch runtime.GOOS {
	case "linux":
		OSInfo.Name = "Linux"
		OSInfo.DataDir = filepath.Join(home, ".craftopiamc")
	case "windows":
		OSInfo.Name = "Windows"
		OSInfo.DataDir = filepath.Join(home, ".craftopiamc")
	case "darwin":
		OSInfo.Name = "macOS"
		OSInfo.DataDir = filepath.Join(home, ".craftopiamc")
	default:
		OSInfo.Name = runtime.GOOS
		OSInfo.DataDir = filepath.Join(home, ".craftopiamc")
	}

	OSInfo.Arch = runtime.GOARCH

	Log(fmt.Sprintf("[PATHS] HomeDir: %s", OSInfo.HomeDir))
	Log(fmt.Sprintf("[PATHS] DataDir: %s", OSInfo.DataDir))
	Log(fmt.Sprintf("[PATHS] MainDir: %s", OSInfo.MainDir))

	return nil
}

func SetupEnvironment() {
	EnsureDirectoryStructure()

	if runtime.GOOS == "windows" {
		if err := setFileHidden(OSInfo.DataDir); err != nil {
			Log(fmt.Sprintf("[HIDE] Failed to hide data dir: %v", err))
		}
	}
}

func InstallIcon() {
	if runtime.GOOS != "linux" {
		return
	}

	iconDestDir := filepath.Join(OSInfo.HomeDir, ".local", "share", "icons")
	iconDest := filepath.Join(iconDestDir, "craftopiamc-launcher.png")

	os.MkdirAll(iconDestDir, 0755)
	os.WriteFile(iconDest, trayIcon, 0644)

	desktopDir := filepath.Join(OSInfo.HomeDir, ".local", "share", "applications")
	os.MkdirAll(desktopDir, 0755)

	exePath, _ := os.Executable()

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=CraftopiaMC Launcher
Comment=CraftopiaMC Game Launcher
Exec=%s
Icon=%s
Type=Application
Categories=Game;
Terminal=false
StartupWMClass=launcher
`, exePath, iconDest)

	os.WriteFile(filepath.Join(desktopDir, "craftopiamc-launcher.desktop"), []byte(desktopContent), 0644)
	Log("[ICON] .desktop file created/updated")
}
