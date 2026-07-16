//go:build !windows
// +build !windows

package static

import (
	"os/exec"
	"syscall"
)

func configureServerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
