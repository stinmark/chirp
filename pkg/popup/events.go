package popup

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/austinemk/sigcat/pkg/helpers"
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
		// In events.go (inside the tea.KeyPressMsg switch statement)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "s", "q", "escape":
			tasks, err := helpers.LoadTasks()
			if err == nil {
				for i, t := range tasks {
					if t.ID == m.Task.ID {
						// 👈 The window is closing, clear the opened flag
						tasks[i].IsOpened = false

						if t.AutoRepeat {
							// Reschedule the next run from the MOMENT they close it
							tasks[i].NextRun = time.Now().Add(time.Duration(t.DurationMin) * time.Minute)
							tasks[i].IsActive = true
						} else {
							tasks[i].IsActive = false
						}
						break
					}
				}
				_ = helpers.SaveTasks(tasks)
			}
			return m, tea.Quit
		case "r":
			// Your existing 'r' key logic for manual repetition overrides...
			tasks, err := helpers.LoadTasks()
			if err == nil {
				for i, t := range tasks {
					if t.ID == m.Task.ID {
						tasks[i].IsActive = true
						tasks[i].NextRun = time.Now().Add(time.Duration(t.DurationMin) * time.Minute)
						break
					}
				}
				_ = helpers.SaveTasks(tasks)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}
