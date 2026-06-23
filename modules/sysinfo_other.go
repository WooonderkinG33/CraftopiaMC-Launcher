//go:build !windows && !linux

package modules

func GetTotalRAMMB() int {
	return 8192
}
