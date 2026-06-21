package popup

import (
	"fmt"

	"github.com/austinemk/sigcat/pkg/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m PopupModel) View() tea.View {
	// 1. Fetch current frame string or fallback text safely
	activeFrame := "Loading animation..."
	if len(m.frames) > 0 {
		activeFrame = m.frames[m.currentFrame]
	}

	// 2. Lip Gloss styles will safely style raw ANSI character sequences out-of-the-box
	gifStyled := lipgloss.NewStyle().Foreground(theme.PurpleColor).Render(activeFrame)

	bannerTitle := theme.GenerateTexturedShadowTitle(theme.TruncateString(m.Task.Title, 7, false), "#AEB6FC")

	daemonStatus := lipgloss.NewStyle().Foreground(theme.RedColor).Render("● Daemon Offline")
	if m.DaemonRunning {
		daemonStatus = lipgloss.NewStyle().Foreground(theme.GreenColor).Render("● Daemon Active")
	}

	var panel string
	panel += lipgloss.NewStyle().Foreground(theme.PinkColor).Width(40).Italic(true).Bold(true).Render(m.Task.Message) + "\n\n"
	panel += fmt.Sprintf("⚙️ Context: %s\n", daemonStatus)

	if m.Task.AutoRepeat {
		panel += lipgloss.NewStyle().Foreground(theme.PurpleColor).Render("🔄 Status: Loop Mode Engaged\n\n")
	} else {
		panel += lipgloss.NewStyle().Foreground(theme.SubtleColor).Render("⏳ Status: Manual Run Executed\n\n")
	}

	// 3. Align the animation horizontally next to text payload
	uiLayout := lipgloss.JoinHorizontal(lipgloss.Bottom, panel, gifStyled)
	combinedView := lipgloss.NewStyle().Padding(2, 4).Render(lipgloss.JoinVertical(lipgloss.Center, bannerTitle, uiLayout))

	v := tea.NewView(combinedView)
	v.AltScreen = true
	return v
}
