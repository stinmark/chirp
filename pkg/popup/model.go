package popup

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/gif"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/austinemk/sigcat/pkg/helpers"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/draw"
)

//go:embed animation.gif
var gifData []byte

type FrameMsg struct{}

type PopupModel struct {
	Task          helpers.BreakTask
	DaemonRunning bool
	frames        []string
	delays        []time.Duration
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
		targeted = helpers.BreakTask{
			Title:   "Take a Break!",
			Message: "Time to stretch and look away.",
		}
	}

	model := PopupModel{
		Task:          targeted,
		DaemonRunning: helpers.IsDaemonRunning(),
	}

	// Read and parse the embedded GIF frames natively
	gifImage, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err == nil {
		targetWidth := 36
		targetHeight := 16

		for i, frame := range gifImage.Image {
			// 1. Downscale the frame dimensions cleanly using standard draw routines
			resizedImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
			draw.NearestNeighbor.Scale(resizedImg, resizedImg.Bounds(), frame, frame.Bounds(), draw.Over, nil)

			// 2. Translate the pixel color array into an ANSI block string matrix
			var sb strings.Builder
			for y := 0; y < targetHeight; y++ {
				for x := 0; x < targetWidth; x++ {
					r, g, b, _ := resizedImg.At(x, y).RGBA()

					// Convert color specs to true color elements via colorful
					c := colorful.Color{
						R: float64(r) / 65535.0,
						G: float64(g) / 65535.0,
						B: float64(b) / 65535.0,
					}

					// Append a solid block '█' formatted with the exact hex truecolor code
					sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm█", int(c.R*255), int(c.G*255), int(c.B*255)))
				}
				sb.WriteString("\x1b[0m\n") // Reset color attributes at the end of every row
			}

			model.frames = append(model.frames, sb.String())

			// Parse timing frames
			delay := time.Duration(gifImage.Delay[i]) * 10 * time.Millisecond
			if delay <= 0 {
				delay = 100 * time.Millisecond
			}
			model.delays = append(model.delays, delay)
		}
	}

	return model
}

func (m PopupModel) Init() tea.Cmd {
	if len(m.delays) > 0 {
		return tea.Tick(m.delays[0], func(t time.Time) tea.Msg {
			return FrameMsg{}
		})
	}
	return nil
}
