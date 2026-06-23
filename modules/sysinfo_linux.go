//go:build linux

package modules

import (
	"os"
	"strconv"
	"strings"
)

func GetTotalRAMMB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 8192
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.Atoi(fields[1])
				return kb / 1024
			}
		}
	}
	return 8192
}
