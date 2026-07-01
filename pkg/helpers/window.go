package helpers

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	var cmd *exec.Cmd
	const uniqueTitle = "chirp-popup"
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	// 0. PREVENT CLUTTER: Force-kill any existing popup window matching our unique title
	// We use pkill -f to find the exact terminal layout title signature.
	_ = exec.Command("pkill", "-f", uniqueTitle).Run()

	// 1. Synchronously inject rules before layout mapping
	injectRuntimeHyprlandRules(uniqueTitle)

	// 2. Launch the terminal with clear, un-nested identifiers
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

// injectRuntimeHyprlandRules registers memory-resident rules inside the active compositor session
func injectRuntimeHyprlandRules(className string) {
	// We map every variant variant Hyprland uses to track surfaces (class, initialClass, and title).
	// This ensures that regardless of how fast your terminal updates its metadata, the
	// compositor grabs the window the exact millisecond it maps onto your workspace.
	rules := []string{
		// 1. Explicit Class mappings
		fmt.Sprintf("float, class:%s", className),
		fmt.Sprintf("pin, class:%s", className),
		fmt.Sprintf("center, class:%s", className),
		fmt.Sprintf("stayfocused, class:%s", className),
		fmt.Sprintf("dimaround, class:%s", className),

		// 2. Initial Class state mappings (Fixes race condition on terminal startup)
		fmt.Sprintf("float, initialClass:%s", className),
		fmt.Sprintf("pin, initialClass:%s", className),
		fmt.Sprintf("center, initialClass:%s", className),
		fmt.Sprintf("stayfocused, initialClass:%s", className),
		fmt.Sprintf("dimaround, initialClass:%s", className),

		// 3. Fallback explicit Title mappings
		fmt.Sprintf("float, title:%s", className),
		fmt.Sprintf("pin, title:%s", className),
		fmt.Sprintf("center, title:%s", className),
		fmt.Sprintf("stayfocused, title:%s", className),
		fmt.Sprintf("dimaround, title:%s", className),
	}

	for _, rule := range rules {
		// Pass each keyword option cleanly over the IPC execution boundaries
		cmd := exec.Command("hyprctl", "keyword", "windowrulev2", rule)
		if err := cmd.Run(); err != nil {
			log.Printf("⚠️ Failed to dynamically append rule [%s]: %v\n", rule, err)
		}
	}
	log.Println("🎨 Centered Hyprland windowrulev2 overlay rules successfully injected via IPC.")
}
