package modules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type LatestRelease struct {
	Version string `json:"version"`
	Linux   struct {
		Sha256   string `json:"sha256"`
		Filename string `json:"filename"`
	} `json:"linux"`
	Windows struct {
		Sha256   string `json:"sha256"`
		Filename string `json:"filename"`
	} `json:"windows"`
}

type UpdateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Sha256    string `json:"sha256"`
	Size      int64  `json:"size"`
}

const latestJSONURL = "https://github.com/WooonderkinG33/CraftopiaMC-Launcher/releases/latest/download/latest.json"

func currentBinaryHash() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	f, err := os.Open(exe)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CheckForUpdate() (*UpdateInfo, error) {
	currentHash, err := currentBinaryHash()
	if err != nil {
		return nil, fmt.Errorf("current hash: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(latestJSONURL)
	if err != nil {
		return nil, fmt.Errorf("fetch latest.json: %w", err)
	}
	defer resp.Body.Close()

	var rel LatestRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode latest.json: %w", err)
	}

	remoteHash := ""
	if runtime.GOOS == "linux" {
		remoteHash = rel.Linux.Sha256
	} else if runtime.GOOS == "windows" {
		remoteHash = rel.Windows.Sha256
	}

	if remoteHash == "" {
		return nil, fmt.Errorf("no binary for %s in latest.json", runtime.GOOS)
	}

	if strings.EqualFold(currentHash, remoteHash) {
		return &UpdateInfo{Available: false}, nil
	}

	return &UpdateInfo{
		Available: true,
		Version:   rel.Version,
		Sha256:    remoteHash,
	}, nil
}

func downloadURL() string {
	if runtime.GOOS == "linux" {
		return "https://github.com/WooonderkinG33/CraftopiaMC-Launcher/releases/latest/download/CraftopiaMC-Launcher-linux"
	}
	return "https://github.com/WooonderkinG33/CraftopiaMC-Launcher/releases/latest/download/CraftopiaMC-Launcher-windows.exe"
}

func DownloadUpdate(progressCb func(pct int)) (string, error) {
	destDir := filepath.Join(OSInfo.DataDir, "runtime")
	os.MkdirAll(destDir, 0755)
	destPath := filepath.Join(destDir, "launcher.update")

	client := &http.Client{Timeout: 0}
	resp, err := client.Get(downloadURL())
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath + ".tmp")
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 256*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if total > 0 && progressCb != nil {
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct > 100 { pct = 100 }
				progressCb(pct)
			}
		}
		if readErr == io.EOF { break }
		if readErr != nil {
			out.Close()
			os.Remove(destPath + ".tmp")
			return "", fmt.Errorf("read: %w", readErr)
		}
	}
	out.Close()

	// Verify hash matches the latest.json (re-fetch)
	info, err := CheckForUpdate()
	if err != nil {
		os.Remove(destPath + ".tmp")
		return "", fmt.Errorf("re-verify: %w", err)
	}
	if info.Available {
		// Compute hash of download
		f, _ := os.Open(destPath + ".tmp")
		h := sha256.New()
		io.Copy(h, f)
		f.Close()
		actualHash := hex.EncodeToString(h.Sum(nil))
		if !strings.EqualFold(actualHash, info.Sha256) {
			os.Remove(destPath + ".tmp")
			return "", fmt.Errorf("SHA256 mismatch after download")
		}
	}

	os.Rename(destPath+".tmp", destPath)
	Log(fmt.Sprintf("[UPDATE] Downloaded to %s", destPath))
	return destPath, nil
}

func ApplyUpdate(downloadedPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	// Write PID so script can wait for us to exit
	pidFile := filepath.Join(OSInfo.DataDir, "runtime", "launcher.pid")
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	if runtime.GOOS == "windows" {
		return applyUpdateWindows(exe, downloadedPath, pidFile)
	}
	return applyUpdateLinux(exe, downloadedPath, pidFile)
}

func applyUpdateLinux(exe, downloadedPath, pidFile string) error {
	scriptPath := filepath.Join(OSInfo.DataDir, "runtime", "update.sh")
	script := fmt.Sprintf(`#!/bin/bash
sleep 1
BIN="%s"
UPDATE="%s"
PIDFILE="%s"
while [ -f "$PIDFILE" ]; do sleep 0.3; done
mv "$BIN" "${BIN}.old"
mv "$UPDATE" "$BIN"
chmod +x "$BIN"
"$BIN" &
sleep 2
rm -f "${BIN}.old"
`, exe, downloadedPath, pidFile)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}
	return exec.Command("/bin/bash", scriptPath).Start()
}

func applyUpdateWindows(exe, downloadedPath, pidFile string) error {
	scriptPath := filepath.Join(OSInfo.DataDir, "runtime", "update.bat")
	script := fmt.Sprintf(`@echo off
setlocal
set "EXE=%s"
set "NEW=%s"
set "PIDFILE=%s"
set /p PID=<"%%PIDFILE%%"
:wait
tasklist /fi "pid eq %%PID%%" 2>nul | find "%%PID%%" >nul
if not errorlevel 1 ping -n 2 127.0.0.1 >nul & goto wait
move /y "%%EXE%%" "%%EXE%%.old"
move /y "%%NEW%%" "%%EXE%%"
start "" "%%EXE%%"
ping -n 3 127.0.0.1 >nul
del "%%EXE%%.old" 2>nul
del "%%PIDFILE%%" 2>nul
`, exe, downloadedPath, pidFile)
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return err
	}
	return exec.Command("cmd", "/c", scriptPath).Start()
}

func CleanupPidFile() {
	pidFile := filepath.Join(OSInfo.DataDir, "runtime", "launcher.pid")
	os.Remove(pidFile)
}
