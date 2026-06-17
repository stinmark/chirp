package main

import (
	_ "embed"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed cat.txt
var catASCII string

type breakModel struct {
	task          BreakTask
	daemonRunning bool
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
		task:          targeted,
		daemonRunning: isDaemonRunning(), // Detect background state status directly
	}
}

func (m breakModel) Init() tea.Cmd { return nil }

func (m breakModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "s", "q", "escape":
			return m, tea.Quit

		case "r":
			tasks, err := LoadTasks()
			if err == nil {
				for i, t := range tasks {
					if t.ID == m.task.ID {
						if t.AutoRepeat {
							// If it's already on repeat, pressing 'r' turns the loop off completely
							tasks[i].AutoRepeat = false
						} else {
							// If it's not repeating, 'r' functions as a normal postpone/repeat trigger
							tasks[i].IsActive = true
							tasks[i].NextRun = time.Now().Add(time.Duration(t.DurationMin) * time.Minute)
						}
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

	// Contextual tracking indicators
	daemonStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("● Daemon Offline")
	if m.daemonRunning {
		daemonStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render("● Daemon Active")
	}

	var panel string
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB8D1")).Bold(true).Render(m.task.Message) + "\n\n"
	panel += fmt.Sprintf("⚙️ Context: %s\n", daemonStatus)

	if m.task.AutoRepeat {
		panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#AEB6FC")).Render("🔄 Status: Loop Mode Engaged\n\n")
	} else {
		panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("⏳ Status: Manual Run Executed\n\n")
	}

	// Informative helper keys menu change
	actionText := "[s] Dismiss Window  •  [r] Stop Repeat Loop"
	if !m.task.AutoRepeat {
		actionText = "[s] Dismiss Window  •  [r] Repeat Task Tracker"
	}
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(actionText)

	uiLayout := lipgloss.JoinHorizontal(lipgloss.Bottom, panel, catStyled)
	combinedView := lipgloss.NewStyle().Padding(2, 4).Render(lipgloss.JoinVertical(lipgloss.Center, bannerTitle, uiLayout))

	v := tea.NewView(combinedView)
	v.AltScreen = true
	return v
}
