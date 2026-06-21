package popup

import (
	"github.com/austinemk/sigcat/pkg/helpers"
	"github.com/austinemk/sigcat/pkg/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m PopupModel) View() tea.View {
	// Safely retrieve the active Braille string frame
	activeArt := ""
	if len(m.frames) > 0 {
		activeArt = m.frames[m.currentFrame]
	}

	// Give the Braille art a clean, muted look like the screenshot (e.g., subtle blue/gray)
	artStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(activeArt)

	bannerTitle := theme.GenerateTexturedShadowTitle(theme.TruncateString(m.Task.Title, 7, false), "#AEB6FC")

	var panel string
	panel += theme.AccentStyle.Width(30).Render(m.Task.Message) + "\n\n"
	panel += theme.MutedStyle.Render("status: "+helpers.Ternary(m.DaemonRunning, "● daemon active", "● daemon stopped")) + "\n"

	if m.Task.AutoRepeat {
		panel += theme.ActiveStye.Render("mode: AutoRepeat\n")
		panel += theme.MutedStyle.Render("[r] stop repeat [s/q] quit")
	} else {
		panel += theme.ActiveStye.Render("mode: Run Once\n")
		panel += theme.MutedStyle.Render("[r] repeat [s/q] quit")
	}

	// Side-by-side alignment: info text panel on the left, Braille art on the right
	uiLayout := lipgloss.JoinHorizontal(lipgloss.Center, panel, artStyled)

	combinedView := lipgloss.NewStyle().Padding(2, 4).
		Render(lipgloss.JoinVertical(lipgloss.Center, bannerTitle, uiLayout))

	v := tea.NewView(combinedView)
	v.AltScreen = true
	return v
}
