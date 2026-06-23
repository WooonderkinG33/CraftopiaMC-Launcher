//go:build !bindings

package main

import (
	"embed"
	"os"

	"craftopiamc-launcher/modules"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/appicon.png
var wailsIcon []byte

func main() {
	if err := modules.InitPaths(); err != nil {
		os.Exit(1)
	}

	// Enable GPU compositing for webkit2gtk — must be set BEFORE wails.Run()
	os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "0")
	os.Setenv("WEBKIT_FORCE_COMPOSITING_MODE", "1")

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "CraftopiaMC Launcher",
		Width:     1024,
		Height:    700,
		MinWidth:  640,
		MinHeight: 480,
		Frameless: true,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 21, A: 255},
		OnStartup:        app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "craftopiamc-launcher-v1",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				app.showWindowFromSecondInstance()
			},
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon: wailsIcon,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
