// Package dashboard for apps entry
package dashboard

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/helpers"
)

type sessionState int

const (
	viewChirps sessionState = iota
	createChirp
)

// ==========================================
// Core Domain Structural Types
// ==========================================

type dashboardModel struct {
	state            sessionState
	chirpList        list.Model
	inputIndex       int
	inputs           []textinput.Model
	errMessage       string
	daemonRunning    bool
	autostartEnabled bool
}

// ==========================================
// Localized Task List Delegate
// ==========================================

type chirpDelegate struct{}

func (d chirpDelegate) Height() int                               { return 2 }
func (d chirpDelegate) Spacing() int                              { return 1 }
func (d chirpDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// ==========================================
// Initialization Loop & Update Loop Router
// ==========================================

func InitialDashboardModel() dashboardModel {
	t, _ := helpers.LoadChirps()

	listItems := make([]list.Item, len(t))
	for i, chirp := range t {
		listItems[i] = chirp
	}

	chirpGrid := list.New(listItems, chirpDelegate{}, 0, 0)
	chirpGrid.SetShowTitle(false)
	chirpGrid.SetShowStatusBar(false)
	chirpGrid.SetShowHelp(false)

	// Fixes the layout to show exactly 5 items (5 items * 2 lines + 4 gaps * 1 line)
	chirpGrid.SetSize(80, 14)

	inputs := make([]textinput.Model, 4)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = ""
	inputs[0].Prompt = ""
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = ""
	inputs[1].Prompt = ""

	inputs[2] = textinput.New()
	inputs[2].Placeholder = ""
	inputs[2].Prompt = ""

	inputs[3] = textinput.New()
	inputs[3].Placeholder = ""
	inputs[3].Prompt = ""

	return dashboardModel{
		state:            viewChirps,
		chirpList:        chirpGrid,
		inputs:           inputs,
		daemonRunning:    daemon.IsDaemonRunning(),
		autostartEnabled: helpers.IsAutostartEnabled(),
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m dashboardModel) getChirps() []helpers.ChirpModel {
	items := m.chirpList.Items()
	chirps := make([]helpers.ChirpModel, len(items))
	for i, item := range items {
		chirps[i] = item.(helpers.ChirpModel)
	}
	return chirps
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.daemonRunning = daemon.IsDaemonRunning()
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if updatedModel, handled, globalCmd := m.handleGlobalKeys(msg); handled {
			return updatedModel, globalCmd
		}

		if m.state == viewChirps {
			return m.handleViewTasksKeys(msg)
		} else if m.state == createChirp {
			return m.handleCreateTaskKeys(msg)
		}
	}

	if m.state == createChirp {
		m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}
