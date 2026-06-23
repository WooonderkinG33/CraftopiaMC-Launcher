package modules

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Hardcoded версии — меняем в обновах
const GameVersion = "26.1.2"
const FabricLoaderVersion = "0.19.2"

const MojangVersionManifest = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"

// Mojang structures
type MojangVersionList struct {
	Versions []MojangVersionEntry `json:"versions"`
}
type MojangVersionEntry struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type MojangVersionDetail struct {
	Downloads struct {
		Client struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"client"`
	} `json:"downloads"`
	Libraries []MojangLibrary `json:"libraries"`
	AssetIndex struct {
		URL  string `json:"url"`
		SHA1 string `json:"sha1"`
		Size int64  `json:"size"`
	} `json:"assetIndex"`
}

type MojangLibrary struct {
	Name string `json:"name"`
	Downloads struct {
		Artifact *struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"artifact"`
	} `json:"downloads"`
}

type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}
type AssetObject struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type MojangData struct {
	ClientURL        string
	ClientSHA1       string
	Libs             []MojangLibInfo
	AssetURL         string
	AssetSHA1        string
	AssetIndexVersion string
}

type MojangLibInfo struct {
	URL  string
	SHA1 string
}

var mojangCache *MojangData

func FindMojangVersion(version string) (*MojangData, error) {
	if mojangCache != nil {
		Log("[MOJANG] Using cached manifest")
		return mojangCache, nil
	}

	Log("[MOJANG] Fetching version list")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(MojangVersionManifest)
	if err != nil {
		return nil, fmt.Errorf("fetch mojang list: %w", err)
	}
	defer resp.Body.Close()

	var list MojangVersionList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode mojang list: %w", err)
	}

	var versionURL string
	for _, v := range list.Versions {
		if v.ID == version {
			versionURL = v.URL
			break
		}
	}
	if versionURL == "" {
		return nil, fmt.Errorf("version %s not found on Mojang", version)
	}

	Log("[MOJANG] Fetching version detail")
	resp2, err := client.Get(versionURL)
	if err != nil {
		return nil, fmt.Errorf("fetch detail: %w", err)
	}
	defer resp2.Body.Close()

	var detail MojangVersionDetail
	if err := json.NewDecoder(resp2.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("decode detail: %w", err)
	}

	data := &MojangData{
		ClientURL:  detail.Downloads.Client.URL,
		ClientSHA1: detail.Downloads.Client.SHA1,
		AssetURL:   detail.AssetIndex.URL,
		AssetSHA1:  detail.AssetIndex.SHA1,
	}

	// Извлекаем версию asset index из URL: .../{version}.json
	if idx := strings.LastIndexByte(data.AssetURL, '/'); idx >= 0 {
		v := data.AssetURL[idx+1:]
		data.AssetIndexVersion = strings.TrimSuffix(v, ".json")
	}
	Log(fmt.Sprintf("[MOJANG] Asset index version: %s", data.AssetIndexVersion))

	for _, lib := range detail.Libraries {
		if lib.Downloads.Artifact != nil && lib.Downloads.Artifact.URL != "" {
			data.Libs = append(data.Libs, MojangLibInfo{
				URL:  lib.Downloads.Artifact.URL,
				SHA1: lib.Downloads.Artifact.SHA1,
			})
		}
	}

	mojangCache = data
	Log(fmt.Sprintf("[MOJANG] Cached: %d libs, client SHA1: %s", len(data.Libs), data.ClientSHA1[:8]))
	return data, nil
}

// Path helpers
func GetJavaDir() string            { return filepath.Join(OSInfo.DataDir, "runtime", "java") }
func GetMinecraftDir() string        { return filepath.Join(OSInfo.DataDir, "minecraft") }

func GetJavaBinary() string {
	javaDir := GetJavaDir()
	binName := "java"
	if runtime.GOOS == "windows" { binName = "javaw.exe" }
	bin := filepath.Join(javaDir, "bin", binName)
	if _, err := os.Stat(bin); err == nil { return bin }
	entries, _ := os.ReadDir(javaDir)
	for _, e := range entries {
		if !e.IsDir() { continue }
		alt := filepath.Join(javaDir, e.Name(), "bin", binName)
		if _, err := os.Stat(alt); err == nil { return alt }
	}
	return ""
}

func IsJavaInstalled() bool        { return GetJavaBinary() != "" }
func IsMinecraftInstalled(version string) bool {
	if version == "" { return false }
	data, err := os.ReadFile(filepath.Join(GetMinecraftDir(), ".craftopia-version"))
	if err != nil { return false }
	return strings.TrimSpace(string(data)) == version
}

// SHA1 helpers
func Sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil { return "", err }
	defer f.Close()
	h := sha1.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func FileMatchesHash(path, expectedHash string) bool {
	if expectedHash == "" { return false }
	actual, err := Sha1File(path)
	if err != nil { return false }
	return strings.EqualFold(actual, expectedHash)
}

func GetLibrariesDir() string { return filepath.Join(GetMinecraftDir(), "libraries") }
func GetAssetsDir() string    { return filepath.Join(GetMinecraftDir(), "assets") }
func GetFabricJar() string {
	return filepath.Join(GetMinecraftDir(), "versions", "fabric-loader-"+FabricLoaderVersion+"-"+GameVersion,
		"fabric-loader-"+FabricLoaderVersion+"-"+GameVersion+".jar")
}
func GetClientJar() string {
	return filepath.Join(GetMinecraftDir(), "versions", GameVersion, GameVersion+".jar")
}
