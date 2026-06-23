//go:build windows

package modules

import (
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func GetTotalRAMMB() int {
	return getWindowsRAM()
}

func getWindowsRAM() int {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	var memInfo memoryStatusEx
	memInfo.cbSize = uint32(unsafe.Sizeof(memInfo))

	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		return 8192
	}

	totalMB := int(memInfo.ullTotalPhys / 1024 / 1024)
	if totalMB < 2048 {
		return 2048
	}
	return totalMB
}

func getLinuxRAM() int {
	return 8192
}

func getMacRAM() int {
	return 8192
}
