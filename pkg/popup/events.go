package popup

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/data"
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
			chirps, err := data.LoadChirps()
			if err == nil {
				for i, c := range chirps {
					if c.ID == m.Chirp.ID {
						// 👈 The window is closing, clear the opened flag
						chirps[i].IsOpened = false

						// If 'r' is pressed, flip the autocomplete/autorepeat setting first
						if msg.String() == "r" {
							chirps[i].AutoRepeat = !chirps[i].AutoRepeat
						}

						// Handle scheduling based on the (potentially flipped) AutoRepeat state
						if chirps[i].AutoRepeat {
							// Reschedule the next run from the MOMENT they close it
							chirps[i].NextRun = time.Now().Add(time.Duration(chirps[i].DurationMin) * time.Minute)
							chirps[i].IsActive = true
						} else {
							chirps[i].IsActive = false
						}
						break
					}
				}
				_ = data.SaveChirps(chirps)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}
