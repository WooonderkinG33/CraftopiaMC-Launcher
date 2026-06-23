//go:build windows

package modules

import _ "embed"

//go:embed tray_icon.ico
var trayIcon []byte
