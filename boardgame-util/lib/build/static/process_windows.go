//go:build windows
// +build windows

package static

import (
	"os/exec"
	"syscall"
)

func configureServerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
