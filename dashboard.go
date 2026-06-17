package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type sessionState int

const (
	viewTasks sessionState = iota
	createTask
)

// LipGloss Styles Palette
var (
	subtleColor = lipgloss.Color("#64748B")
	purpleColor = lipgloss.Color("#AEB6FC")
	pinkColor   = lipgloss.Color("#FFB8D1")
	greenColor  = lipgloss.Color("#22C55E")
	redColor    = lipgloss.Color("#EF4444")

	titleStyle = lipgloss.NewStyle().Foreground(purpleColor).Bold(true).Padding(0, 1)
	cardStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(subtleColor).Padding(1, 2).MarginBottom(1)
	errorStyle = lipgloss.NewStyle().Foreground(redColor).Bold(true)
	helpStyle  = lipgloss.NewStyle().Foreground(subtleColor).Italic(true)
	focusStyle = lipgloss.NewStyle().Foreground(pinkColor).Bold(true)
)

type dashboardModel struct {
	state         sessionState
	tasks         []BreakTask
	cursor        int
	inputIndex    int
	inputs        []textinput.Model // Using real interactive text inputs
	errMessage    string
	daemonRunning bool
}

func isDaemonRunning() bool {
	cmd := exec.Command("pgrep", "-f", "sigcat --run-daemon")
	return cmd.Run() == nil
}

func startDaemon() {
	if isDaemonRunning() {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(executable, "--run-daemon")
	_ = cmd.Start()
}

func stopDaemon() {
	cmd := exec.Command("pkill", "-f", "sigcat --run-daemon")
	_ = cmd.Run()
}

func initialDashboardModel() dashboardModel {
	t, _ := LoadTasks()

	// Initialize the interactive text inputs
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Stretch Break"
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Look away for 20 seconds!"

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "20"

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "y/n"

	return dashboardModel{
		state:         viewTasks,
		tasks:         t,
		inputs:        inputs,
		daemonRunning: isDaemonRunning(),
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.daemonRunning = isDaemonRunning() //
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// 1. GLOBAL KEY INTERCEPTORS (Happens first, regardless of state)
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit //[cite: 1]
		case "q":
			if m.state == viewTasks { //[cite: 1]
				return m, tea.Quit //[cite: 1]
			}
		case "escape":
			if m.state == createTask {
				m.state = viewTasks // Switch safely back to list view
				return m, nil
			}
		}

		// 2. VIEW TASKS STATE
		if m.state == viewTasks { //[cite: 1]
			switch msg.String() { //[cite: 1]
			case "up", "k": //[cite: 1]
				if m.cursor > 0 { //[cite: 1]
					m.cursor-- //[cite: 1]
				}
			case "down", "j": //[cite: 1]
				if m.cursor < len(m.tasks)-1 { //[cite: 1]
					m.cursor++ //[cite: 1]
				}
			case "space": //[cite: 1]
				if len(m.tasks) > 0 { //[cite: 1]
					m.tasks[m.cursor].IsActive = !m.tasks[m.cursor].IsActive //[cite: 1]
					if m.tasks[m.cursor].IsActive {                          //[cite: 1]
						m.tasks[m.cursor].NextRun = time.Now().Add(time.Duration(m.tasks[m.cursor].DurationMin) * time.Minute) //[cite: 1]
						startDaemon()                                                                                          //[cite: 1]
					}
					_ = SaveTasks(m.tasks) //[cite: 1]
				}
			case "s": //[cite: 1]
				if m.daemonRunning { //[cite: 1]
					stopDaemon() //[cite: 1]
				} else {
					startDaemon() //[cite: 1]
				}
				time.Sleep(50 * time.Millisecond)   //[cite: 1]
				m.daemonRunning = isDaemonRunning() //[cite: 1]
			case "n": //[cite: 1]
				m.state = createTask      //[cite: 1]
				m.inputIndex = 0          //[cite: 1]
				for i := range m.inputs { //[cite: 1]
					m.inputs[i].Reset() //[cite: 1]
				}
				m.inputs[0].Focus() //[cite: 1]
				m.errMessage = ""   //[cite: 1]
			case "d": //[cite: 1]
				if len(m.tasks) > 0 { //[cite: 1]
					m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...) //[cite: 1]
					if m.cursor >= len(m.tasks) && m.cursor > 0 {                 //[cite: 1]
						m.cursor-- //[cite: 1]
					}
					_ = SaveTasks(m.tasks) //[cite: 1]
				}
			}
			// 3. CREATE TASK STATE
		} else if m.state == createTask { //[cite: 1]
			switch msg.String() { //[cite: 1]
			// 'escape' case has been moved safely up to the global block!
			case "up", "shift+tab": //[cite: 1]
				if m.inputIndex > 0 { //[cite: 1]
					m.inputs[m.inputIndex].Blur()  //[cite: 1]
					m.inputIndex--                 //[cite: 1]
					m.inputs[m.inputIndex].Focus() //[cite: 1]
				}
			case "down", "tab", "enter": //[cite: 1]
				if m.inputIndex < 3 { //[cite: 1]
					m.inputs[m.inputIndex].Blur()  //[cite: 1]
					m.inputIndex++                 //[cite: 1]
					m.inputs[m.inputIndex].Focus() //[cite: 1]
				} else { //[cite: 1]
					mins, err := strconv.Atoi(strings.TrimSpace(m.inputs[2].Value())) //[cite: 1]
					if err != nil || mins <= 0 {                                      //[cite: 1]
						m.errMessage = "Duration must be a valid positive integer." //[cite: 1]
						return m, nil                                               //[cite: 1]
					}

					autoID := GenerateShortID() //[cite: 1]
					newTask := BreakTask{       //[cite: 1]
						ID:          autoID,                                            //[cite: 1]
						Title:       m.inputs[0].Value(),                               //[cite: 1]
						Message:     m.inputs[1].Value(),                               //[cite: 1]
						DurationMin: mins,                                              //[cite: 1]
						AutoRepeat:  strings.ToLower(m.inputs[3].Value()) == "y",       //[cite: 1]
						IsActive:    true,                                              //[cite: 1]
						NextRun:     time.Now().Add(time.Duration(mins) * time.Minute), //[cite: 1]
					}
					m.tasks = append(m.tasks, newTask) //[cite: 1]
					_ = SaveTasks(m.tasks)             //[cite: 1]
					startDaemon()                      //[cite: 1]

					m.state = viewTasks //[cite: 1]
					return m, nil       //[cite: 1]
				}
			}
		}
	}

	// Update the focused textinput model safely
	if m.state == createTask { //[cite: 1]
		m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg) //[cite: 1]
		cmds = append(cmds, cmd)                                         //[cite: 1]
	}

	return m, tea.Batch(cmds...) //[cite: 1]
}

func (m dashboardModel) View() tea.View {
	// 1. Header block
	headerStr := titleStyle.Render("🐱 SIGCAT HUB")
	statusStr := lipgloss.NewStyle().Foreground(redColor).Render("[🔴 STOPPED]")
	if m.daemonRunning {
		statusStr = lipgloss.NewStyle().Foreground(greenColor).Render("[🟢 RUNNING]")
	}
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, headerStr, " ", statusStr) + "\n\n"

	var bodyContent string

	if m.state == viewTasks {
		// 2. Main View Mode
		bodyContent += "Active Automation Timers Matrix:\n\n"
		if len(m.tasks) == 0 {
			bodyContent += helpStyle.Render("  No active profiles found. Press [n] to create one.") + "\n"
		}

		for i, t := range m.tasks {
			cursor := "  "
			if m.cursor == i {
				cursor = focusStyle.Render("➔ ")
			}

			activeStr := lipgloss.NewStyle().Foreground(subtleColor).Render("❌ Inactive")
			if t.IsActive {
				activeStr = lipgloss.NewStyle().Foreground(greenColor).Render(fmt.Sprintf("🟢 Ready (%s)", t.NextRun.Format("15:04:05")))
			}

			taskRow := fmt.Sprintf(
				"%s %s %s\n    Every %s | %s",
				cursor,
				lipgloss.NewStyle().Foreground(purpleColor).Bold(true).Render("["+t.ID+"]"),
				t.Title,
				focusStyle.Render(strconv.Itoa(t.DurationMin)+"m"),
				activeStr,
			)
			bodyContent += taskRow + "\n\n"
		}

		bodyContent += "\n" + helpStyle.Render("[n] New Task • [space] Toggle • [s] Start/Stop Daemon • [d] Delete • [q] Quit")
	} else {
		// 3. Form Input View Mode
		bodyContent += titleStyle.Render("✨ CREATE NEW SCHEDULER PROFILE") + "\n\n"

		labels := []string{"Window Title:   ", "Sweet Message:  ", "Timeout (Mins): ", "AutoRepeat(y/n):"}
		for i, label := range labels {
			if m.inputIndex == i {
				bodyContent += focusStyle.Render("➔ "+label) + " " + m.inputs[i].View() + "\n"
			} else {
				bodyContent += "  " + label + " " + m.inputs[i].View() + "\n"
			}
		}

		if m.errMessage != "" {
			bodyContent += "\n" + errorStyle.Render("❌ "+m.errMessage) + "\n"
		}
		bodyContent += "\n" + helpStyle.Render("[Esc] Cancel • [Tab/Arrows] Navigate • [Enter] Next / Save")
	}

	// Layout packaging inside a modern TUI container box
	boxedLayout := cardStyle.Render(bodyContent)
	v := tea.NewView(headerBar + boxedLayout)
	v.AltScreen = true
	return v
}
