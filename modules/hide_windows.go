//go:build windows

package modules

import (
	"syscall"
)

func setFileHidden(path string) error {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(ptr, syscall.FILE_ATTRIBUTE_HIDDEN)
}
