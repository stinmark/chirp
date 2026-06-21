package popup

import (
	"embed"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/austinemk/sigcat/pkg/helpers"
)

// 1. Tell Go to embed the entire assets directory automatically
//
//go:embed cats/*.txt
var assetFiles embed.FS

type FrameMsg struct{}

type PopupModel struct {
	Task          helpers.BreakTask
	DaemonRunning bool
	frames        []string
	currentFrame  int
}

func InitialPopupModel(id string) PopupModel {
	tasks, _ := helpers.LoadTasks()
	var targeted helpers.BreakTask

	for _, t := range tasks {
		if t.ID == id {
			targeted = t
			break
		}
	}

	if targeted.ID == "" {
		targeted = helpers.BreakTask{Title: "Take a Break!", Message: "Time to stretch."}
	}

	var loadedFrames []string

	// 2. Cycle through your exported text files and read them into memory
	// Adjust the loop count to match your total number of frames (e.g., 9 frames)
	for i := 0; i < 6; i++ {
		filename := fmt.Sprintf("cats/frame_%d.txt", i)
		data, err := assetFiles.ReadFile(filename)
		if err == nil {
			loadedFrames = append(loadedFrames, string(data))
		}
	}

	return PopupModel{
		Task:          targeted,
		DaemonRunning: helpers.IsDaemonRunning(),
		frames:        loadedFrames,
		currentFrame:  0,
	}
}

func (m PopupModel) Init() tea.Cmd {
	if len(m.frames) > 0 {
		return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
			return FrameMsg{}
		})
	}
	return nil
}
