//go:build !windows

package modules

import _ "embed"

//go:embed tray-icon.png
var trayIcon []byte
