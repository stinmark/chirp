package daemon

import (
	"log"
	"os"
	"time"

	"github.com/stinmark/chirp/pkg/helpers"
	"github.com/stinmark/chirp/pkg/storage"
	"github.com/stinmark/chirp/pkg/window"
)

func RunDaemon() {
	log.Println("🐱 chirp background runtime scheduler listening...")

	terminalApp := helpers.FindTerminal()
	executable, _ := os.Executable()
	now := time.Now()

	// -------------------------------------------------------------------------
	// INITIALIZATION PASS: Reset overdue active tasks so they don't fire all at once
	// -------------------------------------------------------------------------
	if store, err := storage.Load(); err == nil {
		changed := false
		for i, chirp := range store.Chirps {
			// If a task is active but its scheduled time has already passed while
			// the daemon was stopped, push its next run forward from *now*.
			if chirp.IsActive && now.After(chirp.NextRun) {
				log.Printf("🔄 Resetting stale schedule for profile [%s] to avoid pile-up\n", chirp.ID)
				store.Chirps[i].NextRun = now.Add(time.Duration(chirp.DurationMin) * time.Minute)
				changed = true
			}
		}
		if changed {
			_ = storage.Save(store)
		}
	}

	// Start the regular tracking interval
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		store, err := storage.Load()
		if err != nil {
			continue
		}

		changed := false
		now = time.Now()
		activeCount := 0

		for _, chirp := range store.Chirps {
			if !chirp.IsActive {
				continue
			}

			activeCount++ // Found an active profile! Keep it counted.

			if store.IsChirpOpen(chirp.ID) {
				continue
			}

			if now.After(chirp.NextRun) {
				log.Printf("⏰ Target hit for profile [%s]\n", chirp.ID)

				store.SetOpenedChirp(chirp.ID)
				changed = true

				_ = window.SpawnFloatingWindow(terminalApp, executable, chirp.ID)
			}
		}

		if changed {
			_ = storage.Save(store)
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
