//go:build linux || darwin

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// SetProcessGroup configures the command to run in its own process group.
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// HasProcessGroup returns true if the command was configured with its own process group.
func HasProcessGroup(cmd *exec.Cmd) bool {
	return cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid
}

// KillProcessGroup kills the entire process group for the given command.
func KillProcessGroup(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// KillProcess kills a single process. Returns an error if the kill fails,
// ignoring "process already done" and "no such process" errors.
func KillProcess(cmd *exec.Cmd) error {
	err := cmd.Process.Kill()
	if err != nil && !isProcessGoneError(err) {
		return err
	}
	return nil
}

// IsProcessGoneError returns true if the error indicates the process no longer exists.
func IsProcessGoneError(err error) bool {
	return isProcessGoneError(err)
}

func isProcessGoneError(err error) bool {
	if errors.Is(err, os.ErrProcessDone) {
		return true
	}
	// syscall.ESRCH = "no such process"
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.ESRCH
	}
	return false
}

// IsCrashSignal checks if the process exited due to a signal (e.g. SIGSEGV).
func IsCrashSignal(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return status.Signaled()
}

// WrapLimits wraps the command to suppress core dumps via ulimit on Unix.
func WrapLimits(cmd *exec.Cmd) {
	originalArgs := cmd.Args

	cmd.Path = "/bin/sh"

	// Format: sh -c "ulimit -c 0 && exec \"$@\"" -- original_bin args...
	shellCmd := `ulimit -c 0 && exec "$@"`
	newArgs := append([]string{"sh", "-c", shellCmd, "--"}, originalArgs...)
	cmd.Args = newArgs
}

// ProcessGroupPid returns the negative pid for use in log messages.
func ProcessGroupPid(cmd *exec.Cmd) string {
	return fmt.Sprintf("%d", -cmd.Process.Pid)
}
