//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func configureChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateChildProcess(process *os.Process) error {
	return syscall.Kill(-process.Pid, syscall.SIGTERM)
}

func killChildProcess(process *os.Process) error {
	return syscall.Kill(-process.Pid, syscall.SIGKILL)
}
