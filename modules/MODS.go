package modules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CdnFile struct {
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	Hash         string `json:"hash"`
	LastModified int64  `json:"lastModified"`
}

const cdnRootURL = "https://download.craftopiamc.org/"

func SyncMods(statusCb func(key string, pct int, speed string, dl int64, total int64)) error {
	statusCb("modsCheck", 0, "", 0, 0)
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(cdnRootURL)
	if err != nil {
		return fmt.Errorf("cdn list: %w", err)
	}
	defer resp.Body.Close()

	var allFiles []CdnFile
	if err := json.NewDecoder(resp.Body).Decode(&allFiles); err != nil {
		return fmt.Errorf("cdn list decode: %w", err)
	}

	type modInfo struct{ path, hash string }
	var cdnMods []modInfo
	for _, f := range allFiles {
		if strings.HasPrefix(f.Path, "launcher/") && strings.HasSuffix(f.Path, ".jar") {
			cdnMods = append(cdnMods, modInfo{path: f.Path, hash: f.Hash})
		}
	}

	if len(cdnMods) == 0 {
		Log("[MODS] No mods in CDN")
		statusCb("modsOk", 100, "", 0, 0)
		return nil
	}

	Log(fmt.Sprintf("[MODS] CDN has %d mod(s)", len(cdnMods)))
	modsDir := filepath.Join(GetMinecraftDir(), "mods")
	os.MkdirAll(modsDir, 0755)

	// Build set of expected files from CDN
	type needEntry struct{ cdnPath, localPath, expectedHash string }
	var needDownload []needEntry
	expectedSet := make(map[string]string) // filename → hash

	for _, m := range cdnMods {
		localName := filepath.Base(m.path)
		expectedSet[localName] = m.hash
		localPath := filepath.Join(modsDir, localName)

		if fileExists(localPath) {
			h, err := sha256File(localPath)
			if err == nil && strings.EqualFold(h, m.hash) {
				Log(fmt.Sprintf("[MODS] OK: %s", localName))
				continue
			}
			Log(fmt.Sprintf("[MODS] Hash mismatch: %s", localName))
			os.Remove(localPath)
		}

		needDownload = append(needDownload, needEntry{
			cdnPath:      m.path,
			localPath:    localPath,
			expectedHash: m.hash,
		})
	}

	total := int64(len(needDownload))
	if total == 0 {
		// All good — just clean extra mods and done
		entries, _ := os.ReadDir(modsDir)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".jar") {
				if _, keep := expectedSet[e.Name()]; !keep {
					os.Remove(filepath.Join(modsDir, e.Name()))
					Log(fmt.Sprintf("[MODS] Removed extra: %s", e.Name()))
				}
			}
		}
		Log("[MODS] All up to date")
		statusCb("modsOk", 100, "", 0, 0)
		return nil
	}

	Log(fmt.Sprintf("[MODS] Downloading %d mod(s)...", total))
	statusCb("modsDownloading", 0, "", 0, total)

	var completed int64
	var bytesDL int64
	var wg sync.WaitGroup
	work := make(chan int64, total)
	for i := int64(0); i < total; i++ {
		work <- i
	}
	close(work)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		lastB := int64(0)
		lastT := time.Now()
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadInt64(&completed)
				b := atomic.LoadInt64(&bytesDL)
				pct := 0
				if total > 0 {
					pct = int(float64(c) / float64(total) * 100)
				}
				elapsed := time.Since(lastT).Seconds()
				speed := ""
				if elapsed > 0 && b > lastB {
					speed = formatSpeed(float64(b-lastB) / elapsed)
				}
				statusCb("modsDownloading", pct, speed, c, total)
				lastB = b
				lastT = time.Now()
			case <-done:
				return
			}
		}
	}()

	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cli := &http.Client{Timeout: 60 * time.Second}
			for idx := range work {
				entry := needDownload[idx]
				resp, err := cli.Get(cdnRootURL + entry.cdnPath)
				if err != nil {
					atomic.AddInt64(&completed, 1)
					continue
				}

				out, err := os.Create(entry.localPath + ".tmp")
				if err != nil {
					resp.Body.Close()
					atomic.AddInt64(&completed, 1)
					continue
				}

				written, _ := io.Copy(out, resp.Body)
				out.Close()
				resp.Body.Close()

				if written > 0 {
					h, _ := sha256File(entry.localPath + ".tmp")
					if strings.EqualFold(h, entry.expectedHash) {
						os.Rename(entry.localPath+".tmp", entry.localPath)
						atomic.AddInt64(&bytesDL, written)
						Log(fmt.Sprintf("[MODS] Downloaded: %s (%d bytes)", filepath.Base(entry.localPath), written))
					} else {
						Log(fmt.Sprintf("[MODS] Bad hash: %s", filepath.Base(entry.localPath)))
						os.Remove(entry.localPath + ".tmp")
					}
				}
				atomic.AddInt64(&completed, 1)
			}
		}()
	}

	wg.Wait()
	close(done)

	// Удаляем лишние моды ТОЛЬКО после успешной загрузки
	entries, _ := os.ReadDir(modsDir)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jar") {
			if _, keep := expectedSet[e.Name()]; !keep {
				os.Remove(filepath.Join(modsDir, e.Name()))
				Log(fmt.Sprintf("[MODS] Removed extra: %s", e.Name()))
			}
		}
	}

	Log(fmt.Sprintf("[MODS] Sync done (%d/%d)", atomic.LoadInt64(&completed), total))
	statusCb("modsOk", 100, "", 0, 0)
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}
