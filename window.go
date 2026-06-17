package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
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

func SpawnFloatingWindow(terminalApp, executable, taskID string) error {
	var cmd *exec.Cmd
	const uniqueTitle = "sigcat-break-popup"
	uiArg := "--ui=break"
	idArg := fmt.Sprintf("--task-id=%s", taskID)

	switch terminalApp {
	case "kitty", "gnome-terminal":
		cmd = exec.Command(terminalApp, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	case "alacritty", "foot":
		cmd = exec.Command(terminalApp, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	default:
		cmd = exec.Command(terminalApp, "--", executable, uiArg, idArg)
	}

	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}

	go applyHyprlandBatchRules(uniqueTitle)
	return nil
}

// applyHyprlandBatchRules batches all commands to prevent window mapping race conditions
func applyHyprlandBatchRules(titleName string) {
	// Give the terminal process 200ms to open its display canvas and bind the title
	time.Sleep(200 * time.Millisecond)

	// In modern Hyprland, targeting by precise title format string looks like this:
	target := fmt.Sprintf("title:^(%s)$", titleName)

	// Construct an official single-tick batch transaction string.
	// Commands are explicitly separated by semicolons within a single hyprctl payload.
	batchPayload := fmt.Sprintf(
		"dispatch togglefloating %s; "+
			"dispatch resizewindowpixel exact 650 420,%s; "+
			"dispatch centerwindow; "+
			"dispatch pin %s",
		target, target, target,
	)

	// Execute via the official utility with the --batch flag
	cmd := exec.Command("hyprctl", "--batch", batchPayload)

	if err := cmd.Run(); err != nil {
		log.Printf("⚠️ Batch configuration failed: %v\n", err)
	} else {
		log.Println("🎨 Window layer successfully updated via hyprctl --batch.")
	}
}
