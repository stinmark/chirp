package popup

import (
	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/storage"
)

type FrameMsg struct{}

type PopupModel struct {
	Chirp          storage.ChirpModel
	DaemonRunning  bool
	currentFrame   int
	TerminalWidth  int
	TerminalHeight int
}

func InitialPopupModel(id string) PopupModel {
	store, _ := storage.Load()
	var targeted storage.ChirpModel

	for _, c := range store.Chirps {
		if c.ID == id {
			targeted = c
			break
		}
	}

	if targeted.ID == "" {
		targeted = storage.ChirpModel{Message: "Time to stretch."}
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
