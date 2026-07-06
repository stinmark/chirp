//go:build linux

package daemon

import (
	"os"
	"os/exec"
)

func IsDaemonRunning() bool {
	cmd := exec.Command("pgrep", "-f", "chirp --run-daemon")
	return cmd.Run() == nil
}

func StartDaemon() {
	if IsDaemonRunning() {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(executable, "--run-daemon")
	_ = cmd.Start()
}

func KillDaemon() error {
	return exec.Command("pkill", "-f", "chirp --run-daemon").Run()
}

func StopDaemon() {
	_ = KillDaemon()
}

func ConfigureBackgroundCmd(cmd *exec.Cmd) {
	// No-op for Linux
}
