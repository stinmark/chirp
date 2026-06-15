package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type breakModel struct {
	timeLeft time.Duration
}

func (m breakModel) Init() tea.Cmd {
	return tickCmd()
}

func (m breakModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "s" {
			return m, tea.Quit // Skip break
		}
	case tickMsg:
		if m.timeLeft > 0 {
			m.timeLeft -= time.Second
			return m, tickCmd()
		}
		return m, tea.Quit // Break complete, auto-close
	}
	return m, nil
}

func (m breakModel) View() tea.View {
	mins := int(m.timeLeft.Minutes())
	secs := int(m.timeLeft.Seconds()) % 60

	catASCII := `
    /\_/\   🧘 Break Time!
   ( o.o )  Step away from
    > ^ <   the screen.
   /     \ 
  (|  |  |)`

	timerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB8D1")).Bold(true)

	var panel string
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#AEB6FC")).Bold(true).Render("✨ REFRESH YOUR MIND ✨") + "\n\n"
	panel += fmt.Sprintf("Time Remaining: %s\n\n", timerStyle.Render(fmt.Sprintf("%02d:%02d", mins, secs)))
	panel += lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("[s] Skip Break")

	uiLayout := lipgloss.JoinVertical(lipgloss.Top, catASCII, "    ", panel)
	return tea.NewView(lipgloss.NewStyle().Padding(2, 4).Render(uiLayout))
}
