package dashboard

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/storage"
)

// handleGlobalKeys intercepts keys that apply regardless of what screen the user is viewing.
func (m dashboardModel) handleGlobalKeys(msg tea.KeyPressMsg) (dashboardModel, bool, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, true, tea.Quit
	case "q":
		if m.state == viewChirps && !m.chirpList.SettingFilter() {
			return m, true, tea.Quit
		}
	case "escape":
		if m.state == createChirp {
			m.state = viewChirps
			return m, true, nil
		}
	}
	return m, false, nil
}

// handleViewTasksKeys processes interactions for navigating, toggling, and deleting tasks.
func (m dashboardModel) handleViewTasksKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.chirpList.SettingFilter() {
		m.chirpList, cmd = m.chirpList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "s":
		if m.daemonRunning {
			daemon.StopDaemon()
		} else {
			daemon.StartDaemon()
		}
		time.Sleep(50 * time.Millisecond)
		m.daemonRunning = daemon.IsDaemonRunning()
	case "n":
		if len(m.chirpList.Items()) < 50 {
			m.state = createChirp
			m.inputIndex = 0
			for i := range m.inputs {
				m.inputs[i].Reset()
			}
			m.inputs[0].Focus()
			m.errMessage = ""
		}

	case "o":
		if store, err := storage.Load(); err == nil {
			store.RunOnStartup = !store.RunOnStartup
			_ = storage.Save(store)
			m.autostartEnabled = store.RunOnStartup
		}

	case "space":
		if len(m.chirpList.Items()) > 0 {
			idx := m.chirpList.Index()
			if chirp, ok := m.chirpList.SelectedItem().(storage.ChirpModel); ok {
				chirp.IsActive = !chirp.IsActive
				if chirp.IsActive {
					chirp.NextRun = time.Now().Add(time.Duration(chirp.DurationMin) * time.Minute)
					if !m.daemonRunning {
						daemon.StartDaemon()
					}
				}
				m.chirpList.SetItem(idx, chirp)

				if store, err := storage.Load(); err == nil {
					store.Chirps = m.getChirps()
					_ = storage.Save(store)
				}
			}
		}

	case "d":
		if len(m.chirpList.Items()) > 0 {
			m.chirpList.RemoveItem(m.chirpList.Index())
			if store, err := storage.Load(); err == nil {
				store.Chirps = m.getChirps()
				_ = storage.Save(store)
			}
		}
	default:
		m.chirpList, cmd = m.chirpList.Update(msg)
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
			return m.submitNewTask()
		}
	default:
		m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m dashboardModel) submitNewTask() (tea.Model, tea.Cmd) {
	// index 2 is now Timeout/Duration
	rawMins := strings.TrimSpace(m.inputs[2].Value())
	if rawMins == "" {
		rawMins = "20"
	}

	mins, err := strconv.Atoi(rawMins)
	if err != nil || mins <= 0 {
		m.errMessage = "Duration must be a valid positive integer."
		return m, nil
	}

	// Parsing boolean values for PlaySound (1) and AutoRepeat (3)
	soundVal := strings.ToLower(strings.TrimSpace(m.inputs[1].Value()))
	playSound := soundVal == "y" || soundVal == "yes" || soundVal == ""

	repeatVal := strings.ToLower(strings.TrimSpace(m.inputs[3].Value()))
	isRepeat := repeatVal == "y" || repeatVal == "yes" || repeatVal == ""

	newTask := storage.ChirpModel{
		ID:          storage.GenerateShortID(),
		Message:     m.inputs[0].Value(),
		PlaySound:   playSound,
		DurationMin: mins,
		AutoRepeat:  isRepeat,
		IsActive:    true,
		NextRun:     time.Now().Add(time.Duration(mins) * time.Minute),
	}

	m.chirpList.InsertItem(len(m.chirpList.Items()), newTask)

	if store, err := storage.Load(); err == nil {
		store.Chirps = m.getChirps()
		_ = storage.Save(store)
	}
	daemon.StartDaemon()

	m.state = viewChirps
	return m, nil
}
