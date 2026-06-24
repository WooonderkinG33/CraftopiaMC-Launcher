package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"craftopiamc-launcher/core"
	"craftopiamc-launcher/modules"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	targetW       int
	targetH       int
	windowVisible bool
	mcCmd         *exec.Cmd
}

func NewApp() *App {
	return &App{}
}

func killMC(mcCmd **exec.Cmd) {
	if *mcCmd != nil && (*mcCmd).Process != nil {
		(*mcCmd).Process.Kill()
		*mcCmd = nil
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.WindowSetTitle(a.ctx, "CraftopiaMC Launcher "+core.AppVersion)

	modules.InitPaths()
	modules.InitLogger()
	modules.SetupEnvironment()

	settings := modules.LoadSettings()
	modules.SetLanguage(settings.Language)
	modules.Log(fmt.Sprintf("[SETTINGS] lang=%s ram=%dMB", settings.Language, settings.RamMB))
	modules.ExtractIcon()

	screens, _ := runtime.ScreenGetAll(ctx)
	if len(screens) > 0 {
		s := screens[0]
		sw := s.Size.Width
		sh := s.Size.Height
    w := int(float64(sw) * 0.60)
    h := int(float64(sh) * 0.65)
    if w < 800 { w = 800 }
    if h < 480 { h = 480 }
		a.targetW = w
		a.targetH = h
		modules.WriteHeader(sw, sh)
		modules.Log(fmt.Sprintf("[STARTUP] Size: %dx%d", a.targetW, a.targetH))
		runtime.WindowSetSize(ctx, a.targetW, a.targetH)
		time.Sleep(30 * time.Millisecond)
		runtime.WindowSetPosition(ctx, (sw-a.targetW)/2, (sh-a.targetH)/2)
		runtime.WindowShow(ctx)
		a.windowVisible = true
		time.Sleep(10 * time.Millisecond)
		runtime.WindowSetPosition(ctx, (sw-a.targetW)/2, (sh-a.targetH)/2)
	} else {
		runtime.WindowShow(ctx)
		a.windowVisible = true
	}

	modules.SetTrayContext(ctx)
	modules.InitTray()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		killMC(&a.mcCmd)
		os.Exit(0)
	}()

	runtime.EventsOn(ctx, "showWindow", func(...interface{}) {
		runtime.WindowShow(ctx)
		runtime.WindowSetSize(ctx, a.targetW, a.targetH)
		runtime.WindowCenter(ctx)
		a.windowVisible = true
	})
	runtime.EventsOn(ctx, "hideWindow", func(...interface{}) {
		runtime.Hide(ctx)
		a.windowVisible = false
	})
	runtime.EventsOn(ctx, "toggleWindow", func(...interface{}) {
		if a.windowVisible {
			runtime.Hide(ctx)
			a.windowVisible = false
		} else {
			runtime.WindowShow(ctx)
			runtime.WindowSetSize(ctx, a.targetW, a.targetH)
			runtime.WindowCenter(ctx)
			a.windowVisible = true
		}
	})
	runtime.EventsOn(ctx, "quitApp", func(...interface{}) {
		killMC(&a.mcCmd)
		modules.CleanupPidFile()
		runtime.Quit(ctx)
	})

	modules.Log("[STARTUP] Launcher started")
}

func (a *App) GetVersion() string { return core.AppVersion }
func (a *App) GetTotalRAM() int   { return modules.GetTotalRAMMB() }

type SettingsResponse struct {
	Language string `json:"language"`
	RamMB    int    `json:"ram_mb"`
	MaxRam   int    `json:"max_ram"`
}

func (a *App) GetSettings() SettingsResponse {
	s := modules.LoadSettings()
	return SettingsResponse{
		Language: s.Language,
		RamMB:    s.RamMB,
		MaxRam:   modules.GetMaxRamMB(),
	}
}
func (a *App) GetRecommendedRAM() int {
	totalGB := modules.GetTotalRAMMB() / 1024
	switch {
	case totalGB <= 4: return 3
	case totalGB <= 8: return 4
	case totalGB <= 16: return 6
	default: return 8
	}
}
func (a *App) GetMaxRamMB() int { return modules.GetMaxRamMB() }

func (a *App) ResetLauncher() bool {
	if a.mcCmd != nil && a.mcCmd.Process != nil {
		modules.Log("[RESET] MC is running, cannot reset")
		return false
	}
	if err := modules.ResetAllData(); err != nil {
		modules.Log(fmt.Sprintf("[RESET] Failed: %v", err))
		return false
	}
	modules.Log("[RESET] All data cleared")
	return true
}

// SaveSettings — единственный способ сохранить настройки. Go сам валидирует и пишет файл.
func (a *App) SaveSettings(language string, ramMB int) bool {
	s := modules.LauncherSettings{Language: language, RamMB: ramMB}
	if err := modules.SaveSettings(s); err != nil {
		modules.Log(fmt.Sprintf("[SETTINGS] Save failed: %v", err))
		return false
	}
	modules.Log(fmt.Sprintf("[SETTINGS] Saved: lang=%s ram=%dMB", language, ramMB))
	return true
}

func (a *App) emit(msg string, pct int, speed string, dl, total int64) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "runtimeStatus", msg, pct, speed, dl, total)
	}
}

func (a *App) emitError(msg string, retrySec int) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "runtimeStatus", msg, -1, "ERROR", int64(retrySec), 0)
	}
}

func (a *App) emitFatal(msg string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "runtimeStatus", msg, -2, "ERROR", 0, 0)
	}
}

func (a *App) retryPhase(name string, fn func() error) error {
	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			modules.Log(fmt.Sprintf("[RETRY] %s retry #%d", name, attempt))
			for i := 3; i > 0; i-- {
				a.emitError("errConnection", i)
				time.Sleep(1 * time.Second)
			}
		}
		err := fn()
		if err == nil { return nil }
		modules.Log(fmt.Sprintf("[RETRY] %s failed: %v (retry #%d)", name, err, attempt+1))
	}
}

func (a *App) KillMinecraft() {
	modules.Log("[LAUNCH] KillMinecraft called")
	killMC(&a.mcCmd)
	a.emit("MC_KILLED", 0, "", 0, 0)
}

func (a *App) showWindowFromSecondInstance() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
		runtime.WindowSetSize(a.ctx, a.targetW, a.targetH)
		runtime.WindowCenter(a.ctx)
		a.windowVisible = true
	}
}

func (a *App) CheckForUpdate() *modules.UpdateInfo {
	info, err := modules.CheckForUpdate()
	if err != nil {
		modules.Log(fmt.Sprintf("[UPDATE] Check failed: %v", err))
		return &modules.UpdateInfo{Available: false}
	}
	return info
}

func (a *App) ApplyUpdate() bool {
	downloaded, err := modules.DownloadUpdate(func(pct int) {
		a.emit("updating", pct, "", 0, 0)
	})
	if err != nil {
		modules.Log(fmt.Sprintf("[UPDATE] Download failed: %v", err))
		a.emitFatal("errorLaunch")
		return false
	}
	a.emit("updating", 100, "", 0, 0)
	time.Sleep(500 * time.Millisecond)
	a.emit("updateDone", 0, "", 0, 0)
	time.Sleep(2 * time.Second)

	if err := modules.ApplyUpdate(downloaded); err != nil {
		modules.Log(fmt.Sprintf("[UPDATE] Apply failed: %v", err))
		return false
	}
	// Delete PID before quit so update script can proceed
	modules.CleanupPidFile()
	runtime.Quit(a.ctx)
	return true
}

func (a *App) prepareRuntime() {
	modules.Log("[RUNTIME] Starting pipeline")
	a.emit("init", 0, "", 0, 0)

	a.retryPhase("Java", func() error {
		return modules.PrepareRuntimeWithStatus(func(key string, pct int, speed string, dl int64, total int64) {
			a.emit(key, int(float64(pct)*0.20), speed, dl, total)
		})
	})
	if a.ctx == nil { return }

	modules.PrepareMinecraft(func(key string, pct int, speed string, dl int64, total int64) {
		a.emit(key, 20+int(float64(pct)*0.60), speed, dl, total)
	})
	if a.ctx == nil { return }

	// Mods 80-95
	a.retryPhase("Mods", func() error {
		return modules.SyncMods(func(key string, pct int, speed string, dl int64, total int64) {
			a.emit(key, 80+int(float64(pct)*0.15), speed, dl, total)
		})
	})
	if a.ctx == nil { return }

	settings := modules.LoadSettings()
	a.emit("launching", 95, "", 0, 0)
	modules.Log("[LAUNCH] Starting Minecraft")
	cmd, err := modules.LaunchMinecraft(settings.RamMB)
	if err != nil {
		a.emitFatal("errorLaunch")
		modules.Log(fmt.Sprintf("[LAUNCH] Failed: %v", err))
		return
	}
	a.mcCmd = cmd

	modules.WaitForLaunch(cmd,
		func() {
			a.emit("gameRunning", 100, "", 0, 0)
			// Hide launcher when MC window appears
			runtime.Hide(a.ctx)
			a.windowVisible = false
		},
		func(err error) {
			a.emitFatal("errorLaunch")
			modules.Log(fmt.Sprintf("[LAUNCH] Error: %v", err))
		},
	)
	a.mcCmd = nil
	// Show launcher when MC exits
	runtime.WindowShow(a.ctx)
	runtime.WindowSetSize(a.ctx, a.targetW, a.targetH)
	runtime.WindowCenter(a.ctx)
	a.windowVisible = true
	a.emit("MC_EXITED", 0, "", 0, 0)
	modules.Log("[RUNTIME] MC closed, ready")
}

func (a *App) StartRuntimePreparation() {
	go a.prepareRuntime()
}

func (a *App) HideToTray() {
	if a.ctx != nil {
		runtime.Hide(a.ctx)
		a.windowVisible = false
	}
}
