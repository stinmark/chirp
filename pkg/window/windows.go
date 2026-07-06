//go:build windows

package window

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/windows"
)

func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	cmd := exec.Command(executable, uiArg, idArg)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	return cmd.Start()
}
