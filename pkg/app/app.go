// Package app is the package entry
package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/stinmark/chirp/pkg/dashboard"
	"github.com/stinmark/chirp/pkg/helpers"
	"github.com/stinmark/chirp/pkg/popup"
)

func Execute() {
	runFlag := flag.Bool("run-daemon", false, "Start the chirp engine context")
	stopFlag := flag.Bool("stop-daemon", false, "Stop the active background engine")
	uiMode := flag.String("ui", "", "Launch UI ('dashboard' or 'popup')")
	chirpID := flag.String("chirp-id", "", "Target task reference for the popup renderer")
	flag.Parse()

	// 1. DAEMON STOP HANDLER
	if *stopFlag {
		if err := helpers.KillDaemon(); err != nil {
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

			helpers.ConfigureBackgroundCmd(cmd)

			configDir, err := os.UserConfigDir()
			if err != nil {
				configDir = "."
			}
			logDir := filepath.Join(configDir, "chirp")
			_ = os.MkdirAll(logDir, 0o755)

			// Write log output files
			logFile, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err == nil {
				cmd.Stdout = logFile
				cmd.Stderr = logFile
				defer logFile.Close()
			}

			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Failed to split engine thread: %v\n", err)
				return
			}

			// =====================================================================
			// NEW CODE: Record the specific background process PID to a lockfile
			// =====================================================================
			pidFile := filepath.Join(logDir, "daemon.pid")
			_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0o644)
			// =====================================================================

			fmt.Println("🚀 Chirp tracking platform engaged in background!")
			return
		}
		helpers.RunDaemon()
		return
	}
	// 3. POPUP UI ROUTER (Checked before Dashboard to prevent fallback loops)
	if *uiMode == "popup" && *chirpID != "" {
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
