package modules

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec > 1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
	}
	if bytesPerSec > 1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

func adoptiumURL() string {
	osName := "linux"
	if runtime.GOOS == "windows" { osName = "windows" }
	arch := "x64"
	return fmt.Sprintf(
		"https://api.adoptium.net/v3/binary/latest/25/ga/%s/%s/jre/hotspot/normal/eclipse?project=jdk",
		osName, arch,
	)
}

func downloadAndExtractJava(progressCb func(downloaded, total int64, percent int, speed string)) error {
	url := adoptiumURL()
	Log(fmt.Sprintf("[JAVA] Downloading from Adoptium: %s", url))

	javaDir := GetJavaDir()
	os.RemoveAll(javaDir)
	os.MkdirAll(javaDir, 0755)

	client := &http.Client{
		Timeout: 30 * time.Minute,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: 15 * time.Second, KeepAlive: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("adoptium connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("adoptium HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength
	var downloaded int64

	tmpDir, _ := os.MkdirTemp(javaDir, "java-tmp-*")
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "java-download")
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	buf := make([]byte, 256*1024)
	lastTime := time.Now()
	var lastBytes int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)

			if total > 0 && time.Since(lastTime) > 200*time.Millisecond {
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct < 1 { pct = 1 }
				elapsed := time.Since(lastTime).Seconds()
				spd := ""
				if elapsed > 0 { spd = formatSpeed(float64(downloaded-lastBytes) / elapsed) }
				if progressCb != nil { progressCb(downloaded, total, pct, spd) }
				lastTime = time.Now()
				lastBytes = downloaded
			}
		}
		if readErr != nil {
			if readErr == io.EOF { break }
			out.Close()
			return fmt.Errorf("download: %w", readErr)
		}
	}
	out.Close()

	// Extract while temp file still exists (defer RemoveAll runs after this function)
	Log("[JAVA] Extracting...")
	if err := extractJava(tmpFile, javaDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if runtime.GOOS != "windows" {
		binPath := GetJavaBinary()
		if binPath != "" { os.Chmod(binPath, 0755) }
	}

	return nil
}

func PrepareRuntimeWithStatus(statusCb func(key string, pct int, speed string, dl int64, total int64)) error {
	statusCb("javaChecking", 0, "", 0, 0)

	if err := verifyJavaIntegrity(); err == nil {
		Log("[RUNTIME] Java integrity OK")
		statusCb("javaOk", 100, "", 0, 0)
		return nil
	}

	Log("[RUNTIME] Java not found or corrupted, downloading...")
	statusCb("javaDownloading", 0, "", 0, 0)

	if err := downloadAndExtractJava(func(downloaded, total int64, pct int, speed string) {
		statusCb("javaDownloading", pct, speed, downloaded, total)
	}); err != nil {
		return fmt.Errorf("java download failed: %w", err)
	}

	statusCb("javaUnpacking", 100, "", 0, 0)
	os.WriteFile(filepath.Join(GetJavaDir(), ".version"), []byte("25.0"), 0644)
	Log("[JAVA] Java 25 installed successfully")
	return nil
}

func verifyJavaIntegrity() error {
	binPath := GetJavaBinary()
	if binPath == "" { return fmt.Errorf("java binary not found") }
	info, err := os.Stat(binPath)
	if err != nil { return fmt.Errorf("cannot stat: %w", err) }
	if info.Size() == 0 { return fmt.Errorf("binary is empty") }
	Log(fmt.Sprintf("[RUNTIME] Java binary verified: %s (%d bytes)", binPath, info.Size()))
	return nil
}

func extractJava(archivePath, destDir string) error {
	if runtime.GOOS == "linux" {
		return extractTarGz(archivePath, destDir)
	} else if runtime.GOOS == "windows" {
		return extractZip(archivePath, destDir)
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil { return err }
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil { return err }
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var rootDir string
	for {
		header, err := tr.Next()
		if err == io.EOF { break }
		if err != nil { return err }

		if rootDir == "" && header.Typeflag == tar.TypeDir {
			parts := strings.Split(strings.TrimSuffix(header.Name, "/"), "/")
			if len(parts) >= 1 { rootDir = parts[0] + "/" }
		}

		targetPath := header.Name
		if rootDir != "" && strings.HasPrefix(targetPath, rootDir) {
			targetPath = strings.TrimPrefix(targetPath, rootDir)
		}
		if targetPath == "" { continue }

		target := filepath.Join(destDir, targetPath)
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			mode := os.FileMode(0644)
			if header.Mode&0111 != 0 { mode = 0755 }
			outFile, _ := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if outFile != nil {
				io.Copy(outFile, tr)
				outFile.Close()
			}
		}
	}
	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil { return err }
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		outFile, _ := os.Create(target)
		if outFile != nil {
			rc, _ := f.Open()
			if rc != nil {
				io.Copy(outFile, rc)
				rc.Close()
			}
			outFile.Close()
		}
	}
	return nil
}
