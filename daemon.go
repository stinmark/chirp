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
		// Read what is inside the lock file
		if data, err := os.ReadFile(lockPath); err == nil {
			status := string(data)

			// Case A: The UI window is actively open right now
			if status == "active" {
				log.Println("TUI window is open. Skipping this tick.")
				continue
			}

			// Case B: The UI window closed, checking if our grace period timestamp is still active
			if pauseTime, err := time.Parse(time.RFC3339, status); err == nil {
				if time.Now().Before(pauseTime) {
					log.Printf("Daemon is paused until %s. Skipping tick.\n", pauseTime.Format("15:04:05"))
					continue
				}
				// If the pause time has expired, clear the file out cleanly
				os.Remove(lockPath)
			}
		}

		log.Println("Timer hit! Opening break terminal...")

		// 1. Lock immediately before spawning to prevent a race condition on the next tick
		os.MkdirAll(os.Getenv("HOME")+"/.config/sigcat", 0o755)
		_ = os.WriteFile(lockPath, []byte("active"), 0o644)

		executable, err := os.Executable()
		if err != nil {
			executable = os.Args[0]
		}

		// 2. Pass your full system environment (keeps display tokens and window managers happy)
		cmd := exec.Command(terminalApp, "--", executable, "--ui=break")
		cmd.Env = os.Environ()

		err = cmd.Start()
		if err != nil {
			log.Printf("Spawn error: %v\n", err)
			// If spawning failed entirely, clear the lock so it can try again next tick
			os.Remove(lockPath)
		} else {
			log.Println("UI successfully dispatched in background.")
		}
	}
}
