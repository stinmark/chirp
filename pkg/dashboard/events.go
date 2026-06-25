package dashboard

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/austinemk/sigcat/pkg/helpers"
)

// handleGlobalKeys intercepts keys that apply regardless of what screen the user is viewing.
func (m dashboardModel) handleGlobalKeys(msg tea.KeyPressMsg) (dashboardModel, bool, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, true, tea.Quit
	case "q":
		// Only quit on 'q' if we are not actively filling a form or typing a search filter
		if m.state == viewTasks && !m.taskList.SettingFilter() {
			return m, true, tea.Quit
		}
	case "escape":
		if m.state == createTask {
			m.state = viewTasks // Safely abandon form input
			return m, true, nil
		}
	}
	return m, false, nil
}

// handleViewTasksKeys processes interactions for navigating, toggling, and deleting tasks.
func (m dashboardModel) handleViewTasksKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If bubbles list is in filtering mode, pass key events straight through to the list directly
	if m.taskList.SettingFilter() {
		m.taskList, cmd = m.taskList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "space":
		if len(m.taskList.Items()) > 0 {
			idx := m.taskList.Index()
			if task, ok := m.taskList.SelectedItem().(helpers.BreakTask); ok {
				task.IsActive = !task.IsActive
				if task.IsActive {
					task.NextRun = time.Now().Add(time.Duration(task.DurationMin) * time.Minute)
					helpers.StartDaemon()
				}
				m.taskList.SetItem(idx, task)
				_ = helpers.SaveTasks(m.getTasks())
			}
		}
	case "s":
		if m.daemonRunning {
			helpers.StopDaemon()
		} else {
			helpers.StartDaemon()
		}
		time.Sleep(50 * time.Millisecond)
		m.daemonRunning = helpers.IsDaemonRunning()
	case "n":
		if len(m.taskList.Items()) < 50 {
			m.state = createTask
			m.inputIndex = 0
			for i := range m.inputs {
				m.inputs[i].Reset()
			}
			m.inputs[0].Focus()
			m.errMessage = ""
		}
	case "d":
		if len(m.taskList.Items()) > 0 {
			m.taskList.RemoveItem(m.taskList.Index())
			_ = helpers.SaveTasks(m.getTasks())
		}
	default:
		// Forward arrow keys / standard list keys to the bubble list engine natively
		m.taskList, cmd = m.taskList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleCreateTaskKeys manages multi-step form focus toggles and structural validation.
func (m dashboardModel) handleCreateTaskKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "shift+tab":
		if m.inputIndex > 0 {
			m.inputs[m.inputIndex].Blur()
			m.inputIndex--
			m.inputs[m.inputIndex].Focus()
		}
	case "down", "tab", "enter":
		if m.inputIndex < 3 {
			m.inputs[m.inputIndex].Blur()
			m.inputIndex++
			m.inputs[m.inputIndex].Focus()
		} else {
			// Validate and process the final form submission
			return m.submitNewTask()
		}
	default:
		// Route letters/numbers straight into the currently focused text field package
		m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m dashboardModel) submitNewTask() (tea.Model, tea.Cmd) {
	rawMins := strings.TrimSpace(m.inputs[2].Value())
	if rawMins == "" {
		rawMins = "20"
	}

	mins, err := strconv.Atoi(rawMins)
	if err != nil || mins <= 0 {
		m.errMessage = "Duration must be a valid positive integer."
		return m, nil
	}

	repeatVal := strings.ToLower(strings.TrimSpace(m.inputs[3].Value()))
	isRepeat := repeatVal == "y" || repeatVal == "yes" || repeatVal == ""

	newTask := helpers.BreakTask{
		ID:          helpers.GenerateShortID(),
		Title:       m.inputs[0].Value(),
		Message:     m.inputs[1].Value(),
		DurationMin: mins,
		AutoRepeat:  isRepeat,
		IsActive:    true,
		NextRun:     time.Now().Add(time.Duration(mins) * time.Minute),
	}

	m.taskList.InsertItem(len(m.taskList.Items()), newTask)
	_ = helpers.SaveTasks(m.getTasks())
	helpers.StartDaemon()

	m.state = viewTasks
	return m, nil
}
