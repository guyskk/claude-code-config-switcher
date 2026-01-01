//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// executeProcess runs the specified command as a subprocess.
// On Windows, we use exec.Command since there's no exec syscall equivalent.
func executeProcess(path string, args []string, env []string) error {
	cmd := exec.Command(path, args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
