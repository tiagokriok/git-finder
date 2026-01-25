//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// detachProcess configures the command to run in a new session,
// detached from the parent process group (Unix/Linux/macOS)
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
