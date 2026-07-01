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
	log.Println("🐱 chirp background runtime scheduler listening...")

	terminalApp := FindTerminal()
	executable, _ := os.Executable()
	now := time.Now()

	// -------------------------------------------------------------------------
	// INITIALIZATION PASS: Reset overdue active tasks so they don't fire all at once
	// -------------------------------------------------------------------------
	if chirps, err := LoadChirps(); err == nil {
		changed := false
		for i, chirp := range chirps {
			// If a task is active but its scheduled time has already passed while
			// the daemon was stopped, push its next run forward from *now*.
			if chirp.IsActive && now.After(chirp.NextRun) {
				log.Printf("🔄 Resetting stale schedule for profile [%s] (%s) to avoid pile-up\n", chirp.ID, chirp.Title)
				chirps[i].NextRun = now.Add(time.Duration(chirp.DurationMin) * time.Minute)
				changed = true
			}
		}
		if changed {
			_ = SaveChirps(chirps)
		}
	}

	// Start the regular tracking interval
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		chirps, err := LoadChirps()
		if err != nil {
			continue
		}

		changed := false
		now = time.Now()
		activeCount := 0

		for i, chirp := range chirps {
			if !chirp.IsActive {
				continue
			}

			activeCount++ // Found an active profile! Keep it counted.

			// 👈 If the window is already opened, do not fire it again or exit!
			if chirp.IsOpened {
				continue
			}

			if now.After(chirp.NextRun) {
				log.Printf("⏰ Target hit for profile [%s]: %s\n", chirp.ID, chirp.Title)

				// Set IsOpened to true so we don't spawn duplicate windows next tick
				chirps[i].IsOpened = true
				changed = true

				_ = SpawnFloatingWindow(terminalApp, executable, chirp.ID)
			}
		}

		if changed {
			_ = SaveChirps(chirps)
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
