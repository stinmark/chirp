package helpers

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

func RunDaemon() {
	log.Println("🐱 sigcat background runtime scheduler listening...")

	terminalApp := FindTerminal()
	executable, _ := os.Executable()
	now := time.Now()

	// -------------------------------------------------------------------------
	// INITIALIZATION PASS: Reset overdue active tasks so they don't fire all at once
	// -------------------------------------------------------------------------
	if tasks, err := LoadTasks(); err == nil {
		changed := false
		for i, task := range tasks {
			// If a task is active but its scheduled time has already passed while
			// the daemon was stopped, push its next run forward from *now*.
			if task.IsActive && now.After(task.NextRun) {
				log.Printf("🔄 Resetting stale schedule for profile [%s] (%s) to avoid pile-up\n", task.ID, task.Title)
				tasks[i].NextRun = now.Add(time.Duration(task.DurationMin) * time.Minute)
				changed = true
			}
		}
		if changed {
			_ = SaveTasks(tasks)
		}
	}

	// Start the regular tracking interval
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tasks, err := LoadTasks()
		if err != nil {
			continue
		}

		changed := false
		now = time.Now()
		activeCount := 0

		for i, task := range tasks {
			if !task.IsActive {
				continue
			}

			activeCount++ // Found an active profile! Keep it counted.

			// 👈 If the window is already opened, do not fire it again or exit!
			if task.IsOpened {
				continue
			}

			if now.After(task.NextRun) {
				log.Printf("⏰ Target hit for profile [%s]: %s\n", task.ID, task.Title)

				// Set IsOpened to true so we don't spawn duplicate windows next tick
				tasks[i].IsOpened = true
				changed = true

				_ = SpawnFloatingWindow(terminalApp, executable, task.ID)
			}
		}

		if changed {
			_ = SaveTasks(tasks)
		}

		// Self-termination safety logic is now 100% safe.
		// activeCount remains > 0 because your task stays active while open!
		if activeCount == 0 {
			log.Println("💤 No active profiles found running. Giving workspace windows a second to map before exit...")
			time.Sleep(2 * time.Second)
			log.Println("💤 Shutting down daemon context automatically.")
			return
		}
	}
}
