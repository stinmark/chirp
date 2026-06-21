// Package helpers where the utility functions used by other parts live
package helpers

import (
	"os"
	"os/exec"
)

func FindTerminal() string {
	terminals := []string{"kitty", "alacritty", "foot", "gnome-terminal", "konsole", "xterm"}
	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}
	return "xterm"
}

func IsDaemonRunning() bool {
	cmd := exec.Command("pgrep", "-f", "sigcat --run-daemon")
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

func StopDaemon() {
	cmd := exec.Command("pkill", "-f", "sigcat --run-daemon")
	_ = cmd.Run()
}

func Ternary(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}
