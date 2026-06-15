package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
)

func main() {
	startFlag := flag.Bool("start", false, "Start a focused working session")
	stopFlag := flag.Bool("stop-watch", false, "Stop the running daemon") // 👈 Added flag
	uiMode := flag.String("ui", "", "Launch a specific TUI interface")
	flag.Parse()

	// 1. If user typed: ./sigcat --stop-watch
	if *stopFlag {
		// Use pkill to kill any background sigcat instances running the daemon
		cmd := exec.Command("pkill", "-f", "sigcat --start")
		err := cmd.Run()
		if err != nil {
			fmt.Println("❌ No running sigcat daemon found.")
			return
		}
		fmt.Println("🛑 Sigcat daemon stopped successfully.")
		return
	}

	// 2. If user typed: ./sigcat --start
	if *startFlag {
		fmt.Println("🚀 Session active! Go start your task. Sigcat is watching.")
		runDaemon()
		return
	}

	// In main.go (Inside the if *uiMode == "break" block)

	// 2. If daemon executed: ./sigcat --ui=break
	if *uiMode == "break" {
		lockPath := os.Getenv("HOME") + "/.config/sigcat/ui.lock"

		// Run the TUI session
		p := tea.NewProgram(breakModel{timeLeft: 10 * time.Second})
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		// Write a timestamp telling the daemon to pause for 1 minute
		pauseDuration := 10 * time.Second
		pauseUntil := time.Now().Add(pauseDuration).Format(time.RFC3339)

		err := os.WriteFile(lockPath, []byte(pauseUntil), 0o644)
		if err != nil {
			fmt.Printf("Warning: Failed to set pause timestamp: %v\n", err)
		} else {
			fmt.Println("Daemon paused for 1 minute.")
		}
		return
	}

	// 4. Update usage instructions
	fmt.Println("Usage:\n  ./sigcat --start         (To begin working)\n  ./sigcat --stop-watch    (To stop the daemon)\n  ./sigcat --ui=break      (Manual test)")
}
