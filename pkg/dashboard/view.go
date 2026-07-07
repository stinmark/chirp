package dashboard

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/stinmark/chirp/pkg/helpers"
	"github.com/stinmark/chirp/pkg/storage"
	"github.com/stinmark/chirp/pkg/theme"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed banner.txt
var bannerText string

func (d chirpDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	t, ok := listItem.(storage.ChirpModel)
	if !ok {
		return
	}

	taskRow := fmt.Sprintf(
		"%d. [%s] %s\n    %s %dm | sound: %s | %s",
		index+1,
		theme.MutedStyle.Render(t.ID),
		theme.TruncateString(t.Message, 30, true),
		theme.MutedStyle.Render(helpers.Ternary(t.AutoRepeat, "Every", "After")),
		t.DurationMin,
		helpers.Ternary(t.PlaySound, "on", "off"),
		helpers.Ternary(t.IsActive, theme.ActiveStye.Render(fmt.Sprintf("next (%s)", t.NextRun.Format("15:04:05"))), theme.MutedStyle.Render("Inactive")),
	)

	if index == m.Index() {
		taskRow = theme.SelectedItemStyle.Render(taskRow)
	} else {
		taskRow = " " + taskRow
	}

	fmt.Fprint(w, taskRow)
}

func (m dashboardModel) View() tea.View {
	var segments []string

	segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(bannerText))
	statusLine := fmt.Sprintf(
		"daemon %s | launch on startup %s\n",
		helpers.Ternary(m.daemonRunning, "RUNNING", "STOPPED"),
		helpers.Ternary(m.autostartEnabled, "ENABLED", "DISABLED"),
	)
	segments = append(segments, lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("8")).
		Border(lipgloss.NormalBorder(), true, false, true, false).BorderForeground(lipgloss.Color("4")).Width(70).AlignHorizontal(lipgloss.Center).
		Render(statusLine))

	if m.state == viewChirps {
		if len(m.chirpList.Items()) == 0 {
			segments = append(segments, theme.HelpStyle.Render("  No active profiles found. Press [n] to create one.\n"))
		} else {
			segments = append(segments, m.chirpList.View())
		}

		segments = append(segments, theme.HelpStyle.Border(lipgloss.NormalBorder(), true, false, true, false).BorderForeground(lipgloss.Color("3")).
			AlignHorizontal(lipgloss.Center).Width(70).
			Render("[n] New Chirp • [space] Toggle • [o] Toggle Startup • [s] Start/Stop Daemon • [d] Delete • [/] Filter • [q] Quit"))
	} else {
		segments = append(segments, theme.TitleStyle.Render("CREATE NEW SCHEDULER PROFILE \n"))

		// Cleaned labels array reflecting structural field updates
		labels := []string{"Sweet Message:  ", "Play Sound(y/n):", "Timeout (Mins): ", "AutoRepeat(y/n):"}
		for i, label := range labels {
			rowText := fmt.Sprintf("%s %s", label, theme.InputStyle.Render(m.inputs[i].View()))
			if m.inputIndex == i {
				segments = append(segments, theme.SelectedItemStyle.Render(rowText)+"\n")
			} else {
				segments = append(segments, " "+rowText+"\n")
			}
		}

		if m.errMessage != "" {
			segments = append(segments, theme.ErrorStyle.Render("❌ "+m.errMessage)+"\n")
		}
		segments = append(segments, theme.HelpStyle.Render("[Esc] Cancel • [Tab/Arrows] Navigate • [Enter] Next / Save"))
	}

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, segments...))
	v.AltScreen = true
	return v
}
