package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

func main() {
	startFlag := flag.Bool("start", false, "Start a focused working session")
	uiMode := flag.String("ui", "", "Launch a specific TUI interface")
	flag.Parse()

	// 1. If user typed: ./sigcat --start
	if *startFlag {
		fmt.Println("🚀 Session active! Go start your task. Sigcat is watching.")
		runDaemon()
		return
	}

	// 2. If daemon executed: ./sigcat --ui=break
	if *uiMode == "break" {
		lockPath := os.Getenv("HOME") + "/.config/sigcat/ui.lock"

		// Create the config directory if it doesn't exist yet
		os.MkdirAll(os.Getenv("HOME")+"/.config/sigcat", 0o755)

		// Write an empty lock file to notify the daemon that the UI is active
		os.WriteFile(lockPath, []byte("active"), 0o644)

		// Start your break countdown
		p := tea.NewProgram(breakModel{timeLeft: 10 * time.Second})
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		// ✨ CRITICAL: When the TUI closes (or user hits 's'), remove the lock file!
		os.Remove(lockPath)
		return
	}

	// 3. If they typed nothing or wrong flags
	fmt.Println("Usage:\n  ./sigcat --start         (To begin working)\n  ./sigcat --ui=break      (Manual test)")
}
