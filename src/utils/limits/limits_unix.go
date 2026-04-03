//go:build linux || darwin

package limits

import (
	"os/exec"
)

func WrapLimits(cmd *exec.Cmd) {
	// Reconstruct the shell command.
	originalArgs := cmd.Args

	// We use /bin/sh to set the limit and then execute the original command
	cmd.Path = "/bin/sh"

	// Format: sh -c "ulimit -c 0 && exec \"$@\"" -- original_bin args...
	// 'exec' ensures the student process replaces the shell process
	shellCmd := "ulimit -c 0 && exec \"$@\""

	newArgs := append([]string{"sh", "-c", shellCmd, "--"}, originalArgs...)
	cmd.Args = newArgs
}
