package main

import (
	_ "embed"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed cat.txt
var catASCII string

type breakModel struct {
	task BreakTask
}

func initialBreakModel(id string) breakModel {
	tasks, _ := LoadTasks()
	var targeted BreakTask
	for _, t := range tasks {
		if t.ID == id {
			targeted = t
			break
		}
	}
	if targeted.ID == "" {
		targeted = BreakTask{Title: "Take a Break!", Message: "Time to stretch and look away."}
	}
	return breakModel{
		task: targeted,
	}
}

func (m breakModel) Init() tea.Cmd { return nil } // No timer needed anymore

func (m breakModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "s", "q", "escape":
			// Just quit/close the window
			return m, tea.Quit

		case "r":
			// Manually repeat/postpone the task:
			// Reload task list, find this task, and push its next run forward.
			tasks, err := LoadTasks()
			if err == nil {
				for i, t := range tasks {
					if t.ID == m.task.ID {
						tasks[i].IsActive = true
						tasks[i].NextRun = time.Now().Add(time.Duration(t.DurationMin) * time.Minute)
						break
					}
				}
				_ = SaveTasks(tasks)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m breakModel) View() tea.View {
	catStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("#AEB6FC")).Render(catASCII)
	bannerTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true).Render(m.task.Title)

	var panel string
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB8D1")).Bold(true).Render(m.task.Message) + "\n\n"
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render("✨ Status: Completed / Standing By\n\n")

	// Updated dynamic controls footer
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("[s] Close Window  •  [r] Repeat Task Timer")

	uiLayout := lipgloss.JoinHorizontal(lipgloss.Bottom, panel, catStyled)
	combinedView := lipgloss.NewStyle().Padding(2, 4).Render(lipgloss.JoinVertical(lipgloss.Center, bannerTitle, uiLayout))

	v := tea.NewView(combinedView)
	v.AltScreen = true
	return v
}
