package main

import (
	"log"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func runDaemon() {
	log.Println("🐱 sigcat background runtime scheduler listening...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	terminalApp := FindTerminal()
	executable, _ := os.Executable()

	for range ticker.C {
		tasks, err := LoadTasks()
		if err != nil {
			continue
		}

		changed := false
		now := time.Now()
		activeCount := 0

		for i, task := range tasks {
			if !task.IsActive {
				continue
			}

			activeCount++ // Found an active profile

			if now.After(task.NextRun) {
				log.Printf("⏰ Target hit for profile [%s]: %s\n", task.ID, task.Title)

				// Spawn explicit workspace overlay window
				_ = SpawnFloatingWindow(terminalApp, executable, task.ID)

				if task.AutoRepeat {
					tasks[i].NextRun = now.Add(time.Duration(task.DurationMin) * time.Minute)
				} else {
					tasks[i].IsActive = false
					activeCount-- // It just turned inactive
				}
				changed = true
			}
		}

		if changed {
			_ = SaveTasks(tasks)
		}

		// Self-termination safety logic if no automated tasks remain active
		if activeCount == 0 {
			log.Println("💤 No active profiles found running. Shutting down daemon context automatically.")
			return // Exiting main loop shuts down the background daemon process safely!
		}
	}
}
