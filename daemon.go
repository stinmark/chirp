package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Helper function to find an available terminal on the system
func findTerminal() string {
	terminals := []string{
		"gnome-terminal", // Ubuntu/Fedora default
		"kitty",          // GPU accelerated popular terminal
		"alacritty",      // Cross-platform fast terminal
		"konsole",        // KDE default
		"xfce4-terminal", // XFCE default
		"tilix",          // Popular tiling terminal
		"foot",           // Wayland native minimalist terminal
		"wezterm",        // Lua-configured terminal
		"xterm",          // Universal fallback found on almost all Linux setups
	}

	for _, term := range terminals {
		_, err := exec.LookPath(term)
		if err == nil {
			return term // Found a working terminal emulator!
		}
	}
	return "xterm" // Fallback default if nothing else matches
}

func runDaemon() {
	log.Println("🐱 sigcat daemon monitoring session...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	terminalApp := findTerminal()
	log.Printf("Detected terminal emulator: %s\n", terminalApp)

	lockPath := os.Getenv("HOME") + "/.config/sigcat/ui.lock"

	for range ticker.C {
		// 🛑 NEW: Check if the TUI is already open
		if _, err := os.Stat(lockPath); err == nil {
			log.Println("TUI is currently active. Skipping timer tick to prevent duplicate popups.")
			continue // Skip this iteration and wait for the next 10s tick
		}

		log.Println("Timer hit! Opening break terminal...")

		executable, err := os.Executable()
		if err != nil {
			executable = os.Args[0]
		}

		cmd := exec.Command(terminalApp, "-e", executable, "--ui=break")
		cmd.Env = append(os.Environ(), "DISPLAY=:0")

		err = cmd.Run()
		if err != nil {
			log.Printf("Spawn error: %v\n", err)
		}
	}
}
