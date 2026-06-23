package modules

import (
	"context"
	"time"

	"github.com/energye/systray"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var trayCtx context.Context
var trayQuitMenu *systray.MenuItem

func SetTrayContext(ctx context.Context) {
	trayCtx = ctx
}

func InitTray() {
	Log("[TRAY] Initializing system tray...")
	go func() {
		time.Sleep(500 * time.Millisecond)
		systray.Run(onTrayReady, onTrayExit)
	}()
}

func onTrayReady() {
	Log("[TRAY] onTrayReady called")

	systray.SetIcon(trayIcon)
	systray.SetTooltip("CraftopiaMC Launcher")

	trayQuitMenu = systray.AddMenuItem("Выход", "Exit")
	trayQuitMenu.Click(func() {
		if trayCtx != nil {
			wailsruntime.EventsEmit(trayCtx, "quitApp", nil)
		}
		systray.Quit()
	})

	Log("[TRAY] System tray initialized successfully")

	Log("[TRAY] System tray initialized successfully")
}

func onTrayExit() {
	Log("[TRAY] System tray exited")
}
