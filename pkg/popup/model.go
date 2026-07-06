package popup

import (
	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/data"
)

// 1. Tell Go to embed the entire assets directory automatically

type FrameMsg struct{}

type PopupModel struct {
	Chirp          data.ChirpModel
	DaemonRunning  bool
	currentFrame   int
	TerminalWidth  int
	TerminalHeight int
}

func InitialPopupModel(id string) PopupModel {
	chirps, _ := data.LoadChirps()
	var targeted data.ChirpModel

	for _, c := range chirps {
		if c.ID == id {
			targeted = c
			break
		}
	}

	if targeted.ID == "" {
		targeted = data.ChirpModel{Title: "Take a Break!", Message: "Time to stretch."}
	}

	return PopupModel{
		Chirp:         targeted,
		DaemonRunning: daemon.IsDaemonRunning(),
		currentFrame:  0,
	}
}

func (m PopupModel) Init() tea.Cmd {
	return nil
}
