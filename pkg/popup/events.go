package popup

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/storage"
)

func (m PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg: // 👈 Catch the window size event
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r", "s", "q", "escape":
			store, err := storage.Load()
			if err == nil {
				// 👈 The window is closing, clear the top-level active popup ID track
				store.SetOpenedChirp("")

				for i, c := range store.Chirps {
					if c.ID == m.Chirp.ID {
						// If 'r' is pressed, flip the autocomplete/autorepeat setting first
						if msg.String() == "r" {
							store.Chirps[i].AutoRepeat = !store.Chirps[i].AutoRepeat
						}

						// Handle scheduling based on the (potentially flipped) AutoRepeat state
						if store.Chirps[i].AutoRepeat {
							// Reschedule the next run from the MOMENT they close it
							store.Chirps[i].NextRun = time.Now().Add(time.Duration(store.Chirps[i].DurationMin) * time.Minute)
							store.Chirps[i].IsActive = true
						} else {
							store.Chirps[i].IsActive = false
						}
						break
					}
				}
				_ = storage.Save(store)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}
