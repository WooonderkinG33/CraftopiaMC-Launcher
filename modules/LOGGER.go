package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"craftopiamc-launcher/core"
)

var LogFilePath string

func InitLogger() error {
	logsDir := filepath.Join(OSInfo.DataDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs dir: %w", err)
	}

	LogFilePath = filepath.Join(logsDir, "launcher.log")

	file, err := os.Create(LogFilePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	file.Close()

	return nil
}

func WriteHeader(screenW, screenH int) {
	now := time.Now()
	header := fmt.Sprintf("CraftopiaMC Launcher v%s launched\n\n", core.AppVersion)
	header += fmt.Sprintf("System: %s %s\n", OSInfo.Name, OSInfo.Arch)
	header += fmt.Sprintf("Timestamp: %s\n", now.Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("Screen resolution: %dx%d\n\n", screenW, screenH)

	os.WriteFile(LogFilePath, []byte(header), 0644)
}

func Log(msg string) {
	if LogFilePath == "" {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	f, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(entry)
}
