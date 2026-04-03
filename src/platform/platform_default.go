//go:build !linux && !darwin

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// SetProcessGroup is a no-op on non-Unix platforms.
func SetProcessGroup(_ *exec.Cmd) {}

// HasProcessGroup always returns false on non-Unix platforms.
func HasProcessGroup(_ *exec.Cmd) bool { return false }

// KillProcessGroup is a no-op on non-Unix platforms; returns nil.
func KillProcessGroup(_ *exec.Cmd) error { return nil }

// KillProcess kills a single process.
func KillProcess(cmd *exec.Cmd) error {
	err := cmd.Process.Kill()
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

// IsProcessGoneError returns true if the error indicates the process no longer exists.
func IsProcessGoneError(err error) bool {
	return errors.Is(err, os.ErrProcessDone)
}

// IsCrashSignal always returns false on non-Unix platforms.
func IsCrashSignal(_ error) bool { return false }

// WrapLimits is a no-op on non-Unix platforms.
func WrapLimits(_ *exec.Cmd) {}

// ProcessGroupPid returns the pid as a string for log messages.
func ProcessGroupPid(cmd *exec.Cmd) string {
	return fmt.Sprintf("%d", cmd.Process.Pid)
}
