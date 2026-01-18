package actions

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestKillProcess(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("KillProcess test only runs on macOS")
	}

	// Start a dummy process we can kill
	cmd := exec.Command("sleep", "30")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}

	pid := cmd.Process.Pid
	t.Logf("Started test process with PID: %d", pid)

	// Kill the process
	err = KillProcess(pid)
	if err != nil {
		t.Fatalf("KillProcess(%d) error = %v", pid, err)
	}

	// Wait a bit for process to die
	time.Sleep(100 * time.Millisecond)

	// Verify process is dead (sending signal 0 checks existence)
	process, _ := os.FindProcess(pid)
	err = process.Signal(os.Signal(nil))

	// On macOS, FindProcess always succeeds, so we check Wait instead
	waitErr := cmd.Wait()
	if waitErr == nil {
		t.Error("Process still running after KillProcess")
	}
}

func TestKillProcess_InvalidPID(t *testing.T) {
	// Try to kill non-existent process
	err := KillProcess(999999)
	if err == nil {
		t.Error("KillProcess(999999) expected error for invalid PID, got nil")
	}
}
