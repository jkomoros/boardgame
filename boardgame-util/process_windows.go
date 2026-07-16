//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func configureChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func terminateChildProcess(process *os.Process) error {
	return killWindowsProcessTree(process)
}

func killChildProcess(process *os.Process) error {
	return killWindowsProcessTree(process)
}

func killWindowsProcessTree(process *os.Process) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(process.Pid), "/T", "/F").Run()
}
