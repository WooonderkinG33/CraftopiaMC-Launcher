//go:build !linux

package modules

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {}
