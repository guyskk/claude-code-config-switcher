//go:build unix

package cli

import "syscall"

// executeProcess replaces the current process with the specified command.
// On Unix systems, this uses syscall.Exec which does not return on success.
func executeProcess(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}
