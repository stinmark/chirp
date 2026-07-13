// Package app is the package entry
package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/daemon"
	"github.com/stinmark/chirp/pkg/dashboard"
	"github.com/stinmark/chirp/pkg/popup"
	"github.com/stinmark/chirp/pkg/window"
)

func Execute() {
	runFlag := flag.Bool("run-daemon", false, "Start the chirp engine context")
	stopFlag := flag.Bool("stop-daemon", false, "Stop the active background engine")
	uiMode := flag.String("ui", "", "Launch UI ('dashboard' or 'popup')")
	chirpID := flag.String("chirp-id", "", "Target task reference for the popup renderer")
	flag.Parse()

	// 1. DAEMON STOP HANDLER
	if *stopFlag {
		if err := daemon.KillDaemon(); err != nil {
			fmt.Println("❌ No running chirp daemon found.")
			return
		}
		fmt.Println("🛑 Chirp daemon stopped successfully.")
		return
	}

	// 2. DAEMON START HANDLER
	if *runFlag {
		if os.Getenv("CHIRP_BACKGROUND") != "true" {
			executable, _ := os.Executable()
			cmd := exec.Command(executable, "--run-daemon")
			cmd.Env = append(os.Environ(), "CHIRP_BACKGROUND=true")

			daemon.ConfigureBackgroundCmd(cmd)

			configDir, err := os.UserConfigDir()
			if err != nil {
				configDir = "."
			}
			logDir := filepath.Join(configDir, "chirp")
			_ = os.MkdirAll(logDir, 0o755)

			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Failed to split engine thread: %v\n", err)
				return
			}
			pidFile := filepath.Join(logDir, "daemon.pid")
			_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0o644)
			// =====================================================================

			fmt.Println("🚀 Chirp tracking platform engaged in background!")
			return
		}
		daemon.RunDaemon()
		return
	}

	// 3. POPUP UI ROUTER (Checked before Dashboard to prevent fallback loops)
	if *uiMode == "popup" && *chirpID != "" {
		// 1. Center the terminal and force focus right here
		window.CenterAndFocusWindow()

		// 2. Then run your Bubble Tea UI loop
		p := tea.NewProgram(popup.InitialPopupModel(*chirpID))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Popup interface crash: %v\n", err)
		}
		return
	}

	// 4. INTERACTIVE DASHBOARD DEFAULT FALLBACK
	p := tea.NewProgram(dashboard.InitialDashboardModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Dashboard crash: %v\n", err)
	}
}
