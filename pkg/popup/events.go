package popup

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/austinemk/sigcat/pkg/helpers"
)

func (m PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// 1. Listen for the animation timer tick
	case FrameMsg:
		if len(m.frames) == 0 {
			return m, nil
		}
		m.currentFrame = (m.currentFrame + 1) % len(m.frames)

		// Schedule the loop tick for the subsequent frame
		return m, tea.Tick(m.delays[m.currentFrame], func(t time.Time) tea.Msg {
			return FrameMsg{}
		})

	case tea.KeyPressMsg:
		switch msg.String() {
		case "s", "q", "escape":
			return m, tea.Quit

		case "r":
			tasks, err := helpers.LoadTasks()
			if err == nil {
				for i, t := range tasks {
					if t.ID == m.Task.ID {
						if t.AutoRepeat {
							tasks[i].AutoRepeat = false
						} else {
							tasks[i].IsActive = true
							tasks[i].NextRun = time.Now().Add(time.Duration(t.DurationMin) * time.Minute)
						}
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
