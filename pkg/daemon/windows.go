//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// IsDaemonRunning checks if the background daemon process is alive using the PID file
func IsDaemonRunning() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return false
	}

	pidPath := filepath.Join(configDir, "chirp", "daemon.pid")
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		// If the PID file doesn't exist, the daemon isn't running
		return false
	}

	pidStr := strings.TrimSpace(string(pidData))
	if pidStr == "" {
		return false
	}

	// Use tasklist to check if this exact PID is active
	// /FI "PID eq X" filters down to only the target process
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pidStr)).Output()
	// Fallback to absolute system path if environment PATH is broken
	if err != nil {
		out, err = exec.Command(`C:\Windows\System32\tasklist.exe`, "/FI", fmt.Sprintf("PID eq %s", pidStr)).Output()
		if err != nil {
			return false
		}
	}

	// If the PID is active, tasklist returns a table containing the PID string.
	// If the PID is dead, tasklist returns "INFO: No tasks are running which match the specified criteria."
	return strings.Contains(string(out), pidStr)
}

// StartDaemon handles background thread spawning on Windows
func StartDaemon() {
	if IsDaemonRunning() {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(executable, "--run-daemon")
	ConfigureBackgroundCmd(cmd)
	_ = cmd.Start()
}

// KillDaemon / StopDaemon stops the active chirp.exe background process on Windows
// KillDaemon stops ONLY the active chirp.exe background process via its saved PID lockfile
func KillDaemon() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not find config directory")
	}

	pidPath := filepath.Join(configDir, "chirp", "daemon.pid")
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		// Fallback: If no PID file exists, try a soft name check or return error
		return err
	}

	pidStr := strings.TrimSpace(string(pidData))

	// Execute taskkill exclusively on the target PID using the standard path
	err = exec.Command("taskkill", "/F", "/PID", pidStr).Run()
	if err != nil {
		// Fallback to absolute system path if environment PATH is broken
		err = exec.Command(`C:\Windows\System32\taskkill.exe`, "/F", "/PID", pidStr).Run()
	}

	// Clean up the lockfile after successful termination
	if err == nil {
		_ = os.Remove(pidPath)
	}

	return err
}

func StopDaemon() {
	_ = KillDaemon()
}

// ConfigureBackgroundCmd ensures the background daemon runs completely hidden without a console window flashing
func ConfigureBackgroundCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// DETACHED_PROCESS (0x8): child gets no console at all, so it's
		// no longer part of the parent's console session.
		// CREATE_NEW_PROCESS_GROUP (0x200): also stops it from receiving
		// Ctrl+C / Ctrl+Break signals meant for the parent's console.
		CreationFlags: 0x00000008 | 0x00000200,
	}
}
