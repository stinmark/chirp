// Package dashboard for apps entry
package dashboard

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/storage"
)

type sessionState int

const (
	viewChirps sessionState = iota
	createChirp
)

type dashboardModel struct {
	state          sessionState
	chirpList      list.Model
	inputIndex     int
	inputs         []textinput.Model
	errMessage     string
	daemonRunning  bool
	terminalWidth  int // Added to track width for centering
	terminalHeight int // Added to track height for centering
}

type chirpDelegate struct{}

func (d chirpDelegate) Height() int                               { return 2 }
func (d chirpDelegate) Spacing() int                              { return 1 }
func (d chirpDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func InitialDashboardModel() dashboardModel {
	store, _ := storage.Load()

	listItems := make([]list.Item, len(store.Chirps))
	for i, chirp := range store.Chirps {
		listItems[i] = chirp
	}

	chirpGrid := list.New(listItems, chirpDelegate{}, 0, 0)
	chirpGrid.SetShowTitle(false)
	chirpGrid.SetShowStatusBar(false)
	chirpGrid.SetShowHelp(false)
	chirpGrid.SetSize(80, 14)

	// Maintains 3 distinct form elements tailored to the actual model structure
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = ""
		inputs[i].Prompt = ""
	}
	inputs[0].Focus()

	return dashboardModel{
		state:         viewChirps,
		chirpList:     chirpGrid,
		inputs:        inputs,
		daemonRunning: daemon.IsDaemonRunning(),
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m dashboardModel) getChirps() []storage.ChirpModel {
	items := m.chirpList.Items()
	chirps := make([]storage.ChirpModel, len(items))
	for i, item := range items {
		chirps[i] = item.(storage.ChirpModel)
	}
	return chirps
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.daemonRunning = daemon.IsDaemonRunning()
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg: // Added to capture terminal resizing dynamically
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil

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
