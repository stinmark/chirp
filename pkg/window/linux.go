//go:build linux

package window

import (
	"fmt"
	"os"
	"os/exec"
)

func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	var cmd *exec.Cmd
	const uniqueTitle = "chirp-popup"
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	// Clean out existing active popups before creating a fresh window overlay
	_ = exec.Command("pkill", "-f", uniqueTitle).Run()
	injectRuntimeHyprlandRules(uniqueTitle)

	switch terminalApp {
	case "kitty":
		// Restored the "--" argument delimiter
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	case "alacritty":
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	case "foot":
		// Restored the missing "-e" flag required by foot
		cmd = exec.Command(terminalApp, "--app-id", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	default:
		// Restored the "--" argument delimiter for standard emulators
		cmd = exec.Command(terminalApp, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	}

	cmd.Env = os.Environ()
	return cmd.Start()
}

func injectRuntimeHyprlandRules(className string) {
	rules := []string{
		// Explicit Class mappings
		fmt.Sprintf("float, class:%s", className),
		fmt.Sprintf("pin, class:%s", className),
		fmt.Sprintf("center, class:%s", className),
		fmt.Sprintf("stayfocused, class:%s", className),
		fmt.Sprintf("dimaround, class:%s", className),

		// Initial Class state mappings (Fixes startup race conditions)
		fmt.Sprintf("float, initialClass:%s", className),
		fmt.Sprintf("pin, initialClass:%s", className),
		fmt.Sprintf("center, initialClass:%s", className),
		fmt.Sprintf("stayfocused, initialClass:%s", className),
		fmt.Sprintf("dimaround, initialClass:%s", className),

		// Fallback explicit Title mappings
		fmt.Sprintf("float, title:%s", className),
		fmt.Sprintf("pin, title:%s", className),
		fmt.Sprintf("center, title:%s", className),
		fmt.Sprintf("stayfocused, title:%s", className),
		fmt.Sprintf("dimaround, title:%s", className),
	}
	for _, rule := range rules {
		_ = exec.Command("hyprctl", "keyword", "windowrulev2", rule).Run()
	}
}
