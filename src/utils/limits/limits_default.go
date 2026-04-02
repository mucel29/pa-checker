//go:build !linux && !darwin

package limits

import "os/exec"

func WrapLimits(cmd *exec.Cmd) {
	// No-op for Windows/other platforms
}
