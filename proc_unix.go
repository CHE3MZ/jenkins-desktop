//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts the child in its own process group so a single
// signal on shutdown reaches it (and any wrapper children).
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalProcessGroup asks the whole group to terminate gracefully.
func signalProcessGroup(pid int) error {
	return syscall.Kill(-pid, syscall.SIGTERM)
}

// killProcessGroup force-kills the whole group.
func killProcessGroup(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}
