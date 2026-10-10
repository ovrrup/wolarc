//go:build windows

package system

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {
	// Windows does not support POSIX Setpgid
}
