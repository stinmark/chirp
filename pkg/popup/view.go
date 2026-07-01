package popup

import (
	_ "embed" // Required for embedding files

	"github.com/stinmark/chirp/pkg/helpers"
	"github.com/stinmark/chirp/pkg/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed bird.txt
var birdArt string

func (m PopupModel) View() tea.View {
	// 1. Fallback to embedded ASCII art if m.frames is empty

	// 2. Render your layout elements
	artStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true).Render(birdArt)
	bannerTitle := theme.GenerateTexturedShadowTitle(theme.TruncateString(m.Chirp.Title, 10, false), "1")

	var panel []string
	panel = append(panel, theme.AccentStyle.Width(30).Render(m.Chirp.Message)+"\n")
	panel = append(panel, theme.MutedStyle.Render("status: "+helpers.Ternary(m.DaemonRunning, "● daemon active", "● daemon stopped")))

	if m.Chirp.AutoRepeat {
		panel = append(panel, theme.ActiveStye.Render("mode: AutoRepeat"))
		panel = append(panel, theme.MutedStyle.Render("[r] stop repeat [s/q] quit"))
	} else {
		panel = append(panel, theme.ActiveStye.Render("mode: Run Once"))
		panel = append(panel, theme.MutedStyle.Render("[r] repeat [s/q] quit"))
	}

	// 3. Join layout elements together
	uiLayout := lipgloss.JoinVertical(lipgloss.Bottom, bannerTitle, lipgloss.JoinVertical(lipgloss.Left, panel...))
	combinedView := lipgloss.JoinHorizontal(lipgloss.Center, artStyled, uiLayout)

	// 4. DYNAMICALLY CENTER THE CONTENT
	var finalRender string
	if m.TerminalWidth > 0 && m.TerminalHeight > 0 {
		finalRender = lipgloss.Place(
			m.TerminalWidth,
			m.TerminalHeight,
			lipgloss.Center,
			lipgloss.Center,
			combinedView,
		)
	} else {
		finalRender = lipgloss.NewStyle().Padding(2, 4).Render(combinedView)
	}

	v := tea.NewView(finalRender)
	v.AltScreen = true
	return v
}
