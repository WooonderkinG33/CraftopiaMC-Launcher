package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const assetBaseURL = "https://resources.download.minecraft.net/"
const parallelWorkers = 16
const fabricMavenURL = "https://maven.fabricmc.net/net/fabricmc/fabric-loader/%s/fabric-loader-%s.jar"

type downloadTask struct {
	URL  string
	Path string
	Hash string
}

// Один общий клиент для всех batch-запросов (переиспользует соединения)
var sharedBatchClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        32,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	},
}

func PrepareMinecraft(statusCb func(key string, pct int, speed string, dl int64, total int64)) error {
	mcDir := GetMinecraftDir()
	os.MkdirAll(mcDir, 0755)

	statusCb("mcFetch", 0, "", 0, 0)
	mData, err := FindMojangVersion(GameVersion)
	if err != nil {
		return fmt.Errorf("mojang: %w", err)
	}
	statusCb("mcFetch", 100, "", 0, 0)

	if err := processClient(mData, mcDir, statusCb); err != nil {
		return err
	}
	if err := processLibs(mData, mcDir, statusCb); err != nil {
		return err
	}
	if err := processAssets(mData, statusCb, mcDir); err != nil {
		return err
	}
	if err := downloadFabric(mcDir, statusCb); err != nil {
		return err
	}
	ensureFabricDeps()

	os.WriteFile(filepath.Join(mcDir, ".craftopia-version"), []byte(GameVersion), 0644)
	Log(fmt.Sprintf("[MINECRAFT] %s + Fabric ready", GameVersion))
	statusCb("mcOk", 100, "", 0, 0)
	return nil
}

func processClient(m *MojangData, mcDir string, statusCb func(string, int, string, int64, int64)) error {
	clientPath := filepath.Join(mcDir, "versions", GameVersion, GameVersion+".jar")
	os.MkdirAll(filepath.Dir(clientPath), 0755)

	statusCb("mcClientCheck", 0, "", 0, 0)

	if m.ClientSHA1 != "" && FileMatchesHash(clientPath, m.ClientSHA1) {
		Log("[MINECRAFT] Client OK")
		statusCb("mcClientOk", 100, "", 0, 0)
		return nil
	}

	Log("[MINECRAFT] Client missing or corrupted, downloading...")
	statusCb("mcClientDownloading", 0, "", 0, 0)
	for {
		err := downloadFile(sharedBatchClient, m.ClientURL, clientPath, func(dl, total int64, pct int, speed string) {
			statusCb("mcClientDownloading", pct, speed, dl, total)
		})
		if err == nil && (m.ClientSHA1 == "" || FileMatchesHash(clientPath, m.ClientSHA1)) {
			break
		}
		if err != nil {
			Log(fmt.Sprintf("[MINECRAFT] Client download failed: %v, retrying...", err))
			statusCb("mcClientDownloading", 0, "", 0, 0)
			time.Sleep(2 * time.Second)
		}
	}

	Log("[MINECRAFT] Client ready")
	statusCb("mcClientOk", 100, "", 0, 0)
	return nil
}

func processLibs(m *MojangData, mcDir string, statusCb func(string, int, string, int64, int64)) error {
	if len(m.Libs) == 0 {
		statusCb("mcLibsOk", 100, "", 0, 0)
		return nil
	}

	statusCb("mcLibsCheck", 0, "", 0, int64(len(m.Libs)))

	type libCheck struct {
		url  string
		path string
		sha1 string
	}
	libsDir := filepath.Join(mcDir, "libraries")
	var needDownload []libCheck

	for _, lib := range m.Libs {
		jarPath := filepath.Join(libsDir, libJarPath(lib.URL))
		if lib.SHA1 != "" && FileMatchesHash(jarPath, lib.SHA1) {
			continue
		}
		if fileExists(jarPath) {
			os.Remove(jarPath)
		}
		os.MkdirAll(filepath.Dir(jarPath), 0755)
		needDownload = append(needDownload, libCheck{url: lib.URL, path: jarPath, sha1: lib.SHA1})
	}

	total := int64(len(needDownload))
	if total == 0 {
		Log("[MINECRAFT] All libs OK")
		statusCb("mcLibsOk", 100, "", 0, 0)
		return nil
	}

	Log(fmt.Sprintf("[MINECRAFT] Downloading %d libs...", total))
	runBatch("mcLibsDownloading", total, 0, 20, statusCb, func(idx int64, progressCb func(pct int, speed string, dl, total int64)) {
		l := needDownload[idx]
		for {
			sz := downloadOne(sharedBatchClient, l.url, l.path)
			if sz > 0 && (l.sha1 == "" || FileMatchesHash(l.path, l.sha1)) {
				progressCb(100, "", sz, 1)
				return
			}
			if sz > 0 {
				os.Remove(l.path)
			}
			time.Sleep(2 * time.Second)
		}
	})
	return nil
}

func ensureFabricDeps() {
	libsDir := GetLibrariesDir()
	type dep struct{ path, artifact, version, sha1 string }
	deps := []dep{
		{"org/ow2/asm/asm/9.9", "asm", "9.9", ""},
		{"org/ow2/asm/asm-tree/9.9", "asm-tree", "9.9", ""},
		{"org/ow2/asm/asm-commons/9.9", "asm-commons", "9.9", ""},
		{"org/ow2/asm/asm-analysis/9.9", "asm-analysis", "9.9", ""},
		{"org/ow2/asm/asm-util/9.9", "asm-util", "9.9", ""},
		{"net/fabricmc/sponge-mixin/0.17.2+mixin.0.8.7", "sponge-mixin", "0.17.2+mixin.0.8.7", ""},
	}

	allPresent := true
	for _, d := range deps {
		jarPath := filepath.Join(libsDir, d.path, d.artifact+"-"+d.version+".jar")
		if !fileExists(jarPath) {
			allPresent = false
			break
		}
	}

	if allPresent {
		Log("[FABRIC] All deps already present")
		return
	}

	// Only remove and re-download if something is missing
	oldAsm := filepath.Join(libsDir, "org", "ow2", "asm")
	if _, err := os.Stat(oldAsm); err == nil {
		os.RemoveAll(oldAsm)
	}
	oldMixin := filepath.Join(libsDir, "net", "fabricmc", "sponge-mixin")
	if _, err := os.Stat(oldMixin); err == nil {
		os.RemoveAll(oldMixin)
	}

	for _, d := range deps {
		jarPath := filepath.Join(libsDir, d.path, d.artifact+"-"+d.version+".jar")
		if fileExists(jarPath) {
			continue
		}
		url := fmt.Sprintf("https://maven.fabricmc.net/%s/%s-%s.jar", d.path, d.artifact, d.version)
		Log(fmt.Sprintf("[FABRIC] Downloading %s...", d.artifact))
		os.MkdirAll(filepath.Dir(jarPath), 0755)
		for {
			sz := downloadOne(sharedBatchClient, url, jarPath)
			if sz > 0 {
				Log(fmt.Sprintf("[FABRIC] %s downloaded (%d bytes)", d.artifact, sz))
				break
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func processAssets(m *MojangData, statusCb func(string, int, string, int64, int64), mcDir string) error {
	statusCb("mcAssetsCheck", 0, "", 0, 0)

	resp, err := http.Get(m.AssetURL)
	if err != nil {
		return fmt.Errorf("asset index: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read asset index: %w", err)
	}

	var ai AssetIndex
	if err := json.Unmarshal(body, &ai); err != nil {
		return fmt.Errorf("asset index decode: %w", err)
	}

	// Save index JSON для Minecraft
	if m.AssetIndexVersion != "" {
		AssetVersion = m.AssetIndexVersion
		indexDir := filepath.Join(mcDir, "assets", "indexes")
		os.MkdirAll(indexDir, 0755)
		indexPath := filepath.Join(indexDir, m.AssetIndexVersion+".json")
		os.WriteFile(indexPath, body, 0644)
		Log(fmt.Sprintf("[MINECRAFT] Asset index saved: %s (%s)", m.AssetIndexVersion, indexPath))
	}

	Log(fmt.Sprintf("[MINECRAFT] Asset index has %d objects", len(ai.Objects)))
	if len(ai.Objects) == 0 {
		statusCb("mcAssetsOk", 100, "", 0, 0)
		return nil
	}

	objectsDir := filepath.Join(mcDir, "assets", "objects")
	var needDownload []downloadTask

	// Parallel SHA1 verification of existing assets
	type assetCheck struct {
		hash      string
		assetPath string
	}
	checkCh := make(chan assetCheck, len(ai.Objects))
	resultCh := make(chan string, len(ai.Objects))

	for w := 0; w < 8; w++ {
		go func() {
			for ac := range checkCh {
				if fileExists(ac.assetPath) && FileMatchesHash(ac.assetPath, ac.hash) {
					resultCh <- ""
				} else {
					resultCh <- ac.hash
				}
			}
		}()
	}

	for _, obj := range ai.Objects {
		h := obj.Hash
		checkCh <- assetCheck{hash: h, assetPath: filepath.Join(objectsDir, h[:2], h)}
	}
	close(checkCh)

	for i := 0; i < len(ai.Objects); i++ {
		if hash := <-resultCh; hash != "" {
			assetPath := filepath.Join(objectsDir, hash[:2], hash)
			if fileExists(assetPath) {
				os.Remove(assetPath)
			}
			os.MkdirAll(filepath.Dir(assetPath), 0755)
			needDownload = append(needDownload, downloadTask{
				URL: assetBaseURL + hash[:2] + "/" + hash, Path: assetPath, Hash: hash,
			})
		}
	}

	total := int64(len(needDownload))
	Log(fmt.Sprintf("[MINECRAFT] Assets: %d objects, %d cached, %d to download",
		len(ai.Objects), len(ai.Objects)-len(needDownload), total))
	if total == 0 {
		statusCb("mcAssetsOk", 100, "", 0, 0)
		Log("[MINECRAFT] All assets present")
		return nil
	}

	Log(fmt.Sprintf("[MINECRAFT] Downloading %d assets...", total))
	runBatch("Downloading assets", total, 0, 100, statusCb, func(idx int64, progressCb func(pct int, speed string, dl, total int64)) {
		t := needDownload[idx]
		for {
			sz := downloadOne(sharedBatchClient, t.URL, t.Path)
			if sz > 0 {
				progressCb(100, "", sz, 1)
				return
			}
			time.Sleep(2 * time.Second)
		}
	})
	return nil
}

// runBatch запускает параллельную загрузку с per-file retry
func runBatch(key string, total int64, minPct, maxPct int,
	statusCb func(string, int, string, int64, int64),
	workFn func(idx int64, progressCb func(pct int, speed string, dl, total int64))) {

	type jobResult struct {
		idx      int64
		bytesDL  int64
	}

	work := make(chan int64, total)
	for i := int64(0); i < total; i++ {
		work <- i
	}
	close(work)

	var completed int64
	var bytesDL int64
	var wg sync.WaitGroup

	// Progress ticker
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
				pct := minPct
				if total > 0 {
					pct = minPct + int(float64(c)/float64(total)*float64(maxPct-minPct))
				}
				if pct > maxPct {
					pct = maxPct
				}
				b := atomic.LoadInt64(&bytesDL)
				elapsed := time.Since(lastT).Seconds()
				speed := ""
				if elapsed > 0 && b > lastB {
					speed = formatSpeed(float64(b-lastB) / elapsed)
				}
				statusCb(key, pct, speed, c, total)
				lastB = b
				lastT = time.Now()
			case <-done:
				return
			}
		}
	}()

	for w := 0; w < parallelWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				workFn(idx, func(pct int, speed string, dl, jobTotal int64) {
					atomic.AddInt64(&completed, 1)
					atomic.AddInt64(&bytesDL, dl)
				})
			}
		}()
	}
	wg.Wait()
	close(done)
	Log(fmt.Sprintf("[MINECRAFT] %s done (%d/%d)", key, atomic.LoadInt64(&completed), total))
}

func downloadFabric(mcDir string, statusCb func(string, int, string, int64, int64)) error {
	fabricJar := filepath.Join(mcDir, "versions", "fabric-loader-"+FabricLoaderVersion+"-"+GameVersion,
		"fabric-loader-"+FabricLoaderVersion+"-"+GameVersion+".jar")
	os.MkdirAll(filepath.Dir(fabricJar), 0755)

	fabricURL := fmt.Sprintf(fabricMavenURL, FabricLoaderVersion, FabricLoaderVersion)
	Log(fmt.Sprintf("[FABRIC] URL: %s", fabricURL))

	if fileExists(fabricJar) {
		Log("[FABRIC] Already installed")
		return nil
	}

	statusCb("fabricDownloading", 0, "", 0, 0)
	for {
		err := downloadFile(sharedBatchClient, fabricURL, fabricJar, func(dl, total int64, pct int, speed string) {
			statusCb("fabricDownloading", pct, speed, dl, total)
		})
		if err == nil {
			break
		}
		Log(fmt.Sprintf("[FABRIC] Download failed: %v, retrying...", err))
		statusCb("fabricDownloading", 0, "", 0, 0)
		time.Sleep(2 * time.Second)
	}

	Log("[FABRIC] Downloaded successfully")
	return nil
}

func libJarPath(libURL string) string {
	parts := strings.Split(libURL, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return libURL
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func downloadOne(cli *http.Client, url, destPath string) int64 {
	resp, err := cli.Get(url)
	if err != nil {
		Log(fmt.Sprintf("[DL] FAIL: %s — %v", url[:min(len(url), 80)], err))
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		Log(fmt.Sprintf("[DL] HTTP %d: %s", resp.StatusCode, url[:min(len(url), 80)]))
		return 0
	}
	out, err := os.Create(destPath + ".tmp")
	if err != nil {
		Log(fmt.Sprintf("[DL] CREATE FAIL: %s — %v", destPath, err))
		return 0
	}
	written, copyErr := io.Copy(out, resp.Body)
	out.Close()
	if copyErr != nil {
		Log(fmt.Sprintf("[DL] COPY FAIL: %s — %v", url[:min(len(url), 80)], copyErr))
		os.Remove(destPath + ".tmp")
		return 0
	}
	if written > 0 {
		os.Rename(destPath+".tmp", destPath)
	}
	return written
}

func downloadFile(cli *http.Client, url, destPath string, progressCb func(downloaded, total int64, percent int, speed string)) error {
	resp, err := cli.Get(url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath + ".tmp")
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer out.Close()

	var downloaded int64
	total := resp.ContentLength
	buf := make([]byte, 512*1024)
	lastTime := time.Now()
	var lastBytes int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if progressCb != nil && total > 0 && time.Since(lastTime) > 200*time.Millisecond {
				elapsed := time.Since(lastTime).Seconds()
				speed := ""
				if elapsed > 0 {
					speed = formatSpeed(float64(downloaded-lastBytes) / elapsed)
				}
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct < 1 {
					pct = 1
				}
				progressCb(downloaded, total, pct, speed)
				lastTime = time.Now()
				lastBytes = downloaded
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read: %w", readErr)
		}
	}
	out.Close()
	os.Rename(destPath+".tmp", destPath)
	if progressCb != nil {
		progressCb(downloaded, total, 100, "")
	}
	return nil
}
