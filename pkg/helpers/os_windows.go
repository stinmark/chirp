//go:build windows

package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// IsDaemonRunning checks tasklist to see if a background daemon process is alive
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
	}
}

// SpawnFloatingWindow utilizes native cmd.exe to launch an independent popup console window
func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	// Combine target execution arguments cleanly
	targetPayload := fmt.Sprintf(`"%s" %s %s`, executable, uiArg, idArg)

	// FIX: We explicitly pass "" as the first arg after start.
	// This satisfies the window title signature requirement so targetPayload runs correctly.
	cmd := exec.Command("cmd.exe", "/c", "start", "", "cmd.exe", "/c", targetPayload)

	cmd.Env = os.Environ()
	return cmd.Start()
}
