package modules

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"craftopiamc-launcher/core"
)

const MinecraftPlayerName = "CRAFTOPIAMC"

var AssetVersion string // устанавливается из MINECRAFT.go после загрузки

func LaunchMinecraft(ramMB int) (*exec.Cmd, error) {
	Log("[LAUNCH] Launching Minecraft...")

	javaBinary := GetJavaBinary()
	if javaBinary == "" {
		return nil, fmt.Errorf("java binary not found")
	}

	fabricJar := GetFabricJar()
	if _, err := os.Stat(fabricJar); os.IsNotExist(err) {
		return nil, fmt.Errorf("fabric jar not found at %s", fabricJar)
	}
	clientJar := GetClientJar()
	if _, err := os.Stat(clientJar); os.IsNotExist(err) {
		return nil, fmt.Errorf("client jar not found at %s", clientJar)
	}

	classPath, err := buildClassPath()
	if err != nil {
		return nil, fmt.Errorf("classpath build: %w", err)
	}
	Log(fmt.Sprintf("[LAUNCH] Classpath (%d jars)", strings.Count(classPath, ":")+1))

	args := []string{}
	args = append(args, buildJvmArgs(ramMB)...)
	args = append(args, "-cp", classPath)
	args = append(args, "net.fabricmc.loader.impl.launch.knot.KnotClient")
	args = append(args, buildGameArgs()...)

	Log(fmt.Sprintf("[LAUNCH] Cmd: %s %s ...", javaBinary, strings.Join(args[:min(len(args), 8)], " ")))

	cmd := exec.Command(javaBinary, args...)
	cmd.Dir = GetMinecraftDir()
	setProcessGroup(cmd)

	// Pipe stdout/stderr → minecraft.log
	mcLogPath := filepath.Join(filepath.Dir(LogFilePath), "minecraft.log")
	mcLog, err := os.Create(mcLogPath)
	if err != nil {
		return nil, fmt.Errorf("minecraft log: %w", err)
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		mcLog.Close()
		return nil, fmt.Errorf("pipe: %w", err)
	}
	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Start(); err != nil {
		writer.Close(); reader.Close(); mcLog.Close()
		return nil, fmt.Errorf("start: %w", err)
	}
	writer.Close()

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			mcLog.WriteString(line + "\n")
			mcLog.Sync()
		}
		reader.Close()
		mcLog.Close()
	}()

	Log(fmt.Sprintf("[LAUNCH] PID: %d", cmd.Process.Pid))
	return cmd, nil
}

func WaitForLaunch(cmd *exec.Cmd, launchCb func(), failCb func(error)) {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// Ждём маркер "Setting user: CRAFTOPIAMC_PLAYER" или 15s
	mcLogPath := filepath.Join(filepath.Dir(LogFilePath), "minecraft.log")
	marker := "Setting user: " + MinecraftPlayerName
	timer := time.After(15 * time.Second)

	for {
		select {
		case err := <-done:
			// Process exited
			logData, _ := os.ReadFile(mcLogPath)
			logStr := string(logData)
			if logStr != "" {
				lines := strings.Split(logStr, "\n")
				errLines := []string{}
				for _, l := range lines {
					if strings.Contains(l, "Exception") || strings.Contains(l, "Error") {
						errLines = append(errLines, strings.TrimSpace(l))
					}
				}
				if len(errLines) > 0 {
					Log(fmt.Sprintf("[LAUNCH] %s", strings.Join(errLines[:min(len(errLines), 5)], " | ")))
				}
			}
			failCb(fmt.Errorf("minecraft exited: %v", err))
			return
		case <-timer:
			// Timeout — считаем что окно уже должно было появиться
			logData, _ := os.ReadFile(mcLogPath)
			if strings.Contains(string(logData), "Minecraft 26") || strings.Contains(string(logData), marker) {
				launchCb()
				<-done
				Log("[LAUNCH] Minecraft exited")
				return
			}
			failCb(fmt.Errorf("launch timeout (15s)"))
			return
		default:
			logData, _ := os.ReadFile(mcLogPath)
			if strings.Contains(string(logData), marker) {
				Log("[LAUNCH] Window detected via " + marker)
				launchCb()
				<-done
				Log("[LAUNCH] Minecraft exited")
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func buildJvmArgs(ramMB int) []string {
	base := []string{
		"-XX:+UseG1GC", "-XX:+ParallelRefProcEnabled", "-XX:MaxGCPauseMillis=200",
		"-XX:+UnlockExperimentalVMOptions", "-XX:+DisableExplicitGC", "-XX:+AlwaysPreTouch",
		"-XX:G1NewSizePercent=30", "-XX:G1MaxNewSizePercent=40", "-XX:G1HeapRegionSize=8M",
		"-XX:G1ReservePercent=20", "-XX:G1HeapWastePercent=5", "-XX:G1MixedGCCountTarget=4",
		"-XX:InitiatingHeapOccupancyPercent=15", "-XX:G1MixedGCLiveThresholdPercent=90",
		"-XX:G1RSetUpdatingPauseTimePercent=5", "-XX:SurvivorRatio=32",
		"-XX:+PerfDisableSharedMem", "-XX:MaxTenuringThreshold=1",
	}

	// Pass icon path + version to Minecraft (Fabric mod reads this)
	iconPath := GetIconPath()
	if _, err := os.Stat(iconPath); err == nil {
		base = append(base, "-Dcraftopiamc.iconPath="+iconPath)
	}
	base = append(base, "-Dcraftopiamc.launcherVersion="+core.AppVersion)
	return append([]string{fmt.Sprintf("-Xms%dM", ramMB), fmt.Sprintf("-Xmx%dM", ramMB)}, base...)
}

func buildClassPath() (string, error) {
	libsDir := GetLibrariesDir()
	fabricJar := GetFabricJar()
	clientJar := GetClientJar()

	var paths []string
	paths = append(paths, fabricJar)
	paths = append(paths, clientJar)

	err := filepath.Walk(libsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Verify critical JARs
	required := []string{"asm-9.9.jar", "sponge-mixin-0.17.2"}
	foundAll := true
	for _, req := range required {
		found := false
		for _, p := range paths {
			if strings.Contains(p, req) { found = true; break }
		}
		if !found {
			Log(fmt.Sprintf("[LAUNCH] MISSING REQUIRED JAR: %s", req))
			foundAll = false
		}
	}
	if !foundAll {
		Log("[LAUNCH] Some critical Fabric dependencies are missing!")
	}

	separator := ":"
	if runtime.GOOS == "windows" { separator = ";" }
	return strings.Join(paths, separator), nil
}

func buildGameArgs() []string {
	args := []string{
		"--username", MinecraftPlayerName,
		"--version", GameVersion,
		"--gameDir", GetMinecraftDir(),
		"--assetsDir", GetAssetsDir(),
		"--assetIndex", AssetVersion,
		"--uuid", generateUUID(MinecraftPlayerName),
		"--accessToken", "0",
		"--userType", "mojang",
		"--versionType", "CraftopiaMC",
	}
	if runtime.GOOS == "linux" {
		args = append(args, "--nativeLauncherVersion", "linux")
	}
	return args
}

func generateUUID(name string) string {
	hash := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}
