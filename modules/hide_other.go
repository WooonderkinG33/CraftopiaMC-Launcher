//go:build !windows

package modules

func setFileHidden(path string) error {
	return nil
}
