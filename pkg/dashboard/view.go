package dashboard

import (
	"fmt"
	"io"

	"github.com/austinemk/sigcat/pkg/helpers"
	"github.com/austinemk/sigcat/pkg/theme"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	t, ok := listItem.(helpers.BreakTask)
	if !ok {
		return
	}

	// Dynamically determine prefix based on whether the task repeat
	taskRow := fmt.Sprintf(
		"%d. [%s] %s\n    %s %dm | %s",
		index+1,
		theme.MutedStyle.Render(t.ID),
		t.Title,
		theme.MutedStyle.Render(helpers.Ternary(t.AutoRepeat, "Every", "After")),
		t.DurationMin,
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

	segments = append(segments, theme.GenerateTexturedShadowTitle("SIGCAT HUB", "1"))
	segments = append(segments, lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("8")).Render("---"+helpers.Ternary(m.daemonRunning, "daemon running\n", "daemon stopped\n")))

	if m.state == viewTasks {
		if len(m.taskList.Items()) == 0 {
			segments = append(segments, theme.HelpStyle.Render("  No active profiles found. Press [n] to create one.\n"))
		} else {
			segments = append(segments, m.taskList.View())
		}

		segments = append(segments, theme.HelpStyle.Render("[n] New Task • [space] Toggle • [s] Start/Stop Daemon • [d] Delete • [/] Filter • [q] Quit"))
	} else {
		segments = append(segments, theme.TitleStyle.Render("CREATE NEW SCHEDULER PROFILE \n"))

		labels := []string{"Window Title:   ", "Sweet Message:  ", "Timeout (Mins): ", "AutoRepeat(y/n):"}
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
