//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// detachProcess configures the command to run detached from the parent (Windows)
// Uses CREATE_NEW_PROCESS_GROUP to detach from parent's console
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
