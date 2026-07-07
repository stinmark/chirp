//go:build linux

package window

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CenterAndFocusWindow remains an empty stub because on Linux, window positioning
// must be managed externally by the compositor/WM before or during spawning.
func CenterAndFocusWindow() {}

func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	var cmd *exec.Cmd
	const uniqueTitle = "chirp-popup"
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	// Clean out existing active popups before creating a fresh window overlay
	_ = exec.Command("pkill", "-f", uniqueTitle).Run()

	// 1. Detect the current Desktop Environment and inject specific rules
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	if strings.Contains(desktop, "hyprland") {
		injectRuntimeHyprlandRules(uniqueTitle)
	} else if strings.Contains(desktop, "gnome") {
		injectGNOMERules(uniqueTitle)
	} else if strings.Contains(desktop, "kde") {
		injectKDERules(uniqueTitle)
	}

	// 2. Build the command matching your terminal
	switch terminalApp {
	case "kitty":
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	case "alacritty":
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	case "foot":
		cmd = exec.Command(terminalApp, "--app-id", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	default:
		cmd = exec.Command(terminalApp, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	}

	cmd.Env = os.Environ()
	return cmd.Start()
}

func injectRuntimeHyprlandRules(className string) {
	rules := []string{
		fmt.Sprintf("float, class:%s", className),
		fmt.Sprintf("pin, class:%s", className),
		fmt.Sprintf("center, class:%s", className),
		fmt.Sprintf("stayfocused, class:%s", className),
		fmt.Sprintf("dimaround, class:%s", className),

		fmt.Sprintf("float, initialClass:%s", className),
		fmt.Sprintf("pin, initialClass:%s", className),
		fmt.Sprintf("center, initialClass:%s", className),
		fmt.Sprintf("stayfocused, initialClass:%s", className),
		fmt.Sprintf("dimaround, initialClass:%s", className),

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

func injectGNOMERules(className string) {
	// GNOME (Mutter) strictly blocks applications from positioning themselves on Wayland.
	// If the user has 'wmctrl' or 'xdotool' installed (running via XWayland/X11 fallback),
	// we can try to force positioning, though Extension tools are typically preferred.
}

func injectKDERules(className string) {
	// KDE KWin allows injecting runtime rules using 'qdbus' or writing temporary
	// rule configs to ~/.config/kwinrulesrc.
}
