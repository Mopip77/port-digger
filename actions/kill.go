package actions

import (
	"fmt"
	"os/exec"
)

// KillProcess attempts to terminate a process by PID
// First tries graceful SIGTERM, then falls back to SIGKILL with sudo
func KillProcess(pid int) error {
	pidStr := fmt.Sprintf("%d", pid)

	// Try graceful kill first (SIGTERM)
	cmd := exec.Command("kill", "-15", pidStr)
	err := cmd.Run()

	if err == nil {
		return nil
	}

	// If graceful kill failed, try force kill with sudo via osascript
	// This will prompt user for password via native macOS dialog
	script := fmt.Sprintf("kill -9 %d", pid)
	cmd = exec.Command("osascript", "-e",
		fmt.Sprintf(`do shell script "%s" with administrator privileges`, script))

	return cmd.Run()
}
