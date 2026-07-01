package popup

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/helpers"
)

func (m PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg: // 👈 Catch the window size event
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil

	// Handle the simple text frame step
	case FrameMsg:
		if len(m.frames) == 0 {
			return m, nil
		}

		m.currentFrame = (m.currentFrame + 1) % len(m.frames)

		// Adjust the duration (e.g., 125ms) to change the animation speed
		return m, tea.Tick(125*time.Millisecond, func(t time.Time) tea.Msg {
			return FrameMsg{}
		})

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r", "s", "q", "escape":
			chirps, err := helpers.LoadChirps()
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
				_ = helpers.SaveChirps(chirps)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}
