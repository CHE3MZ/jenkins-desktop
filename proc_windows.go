//go:build windows

package main

import (
	"os"
	"os/exec"
)

func findProcess(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}

// setProcessGroup is a no-op on Windows; the child is signalled directly.
func setProcessGroup(cmd *exec.Cmd) {}

// signalProcessGroup terminates the child process.
func signalProcessGroup(pid int) error {
	proc, err := findProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// killProcessGroup force-kills the child process.
func killProcessGroup(pid int) error {
	return signalProcessGroup(pid)
}
