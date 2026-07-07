//go:build windows

package window

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procGetConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

const (
	SM_CXSCREEN    = 0
	SM_CYSCREEN    = 1
	HWND_TOPMOST   = -1
	SWP_NOSIZE     = 0x0001
	SWP_SHOWWINDOW = 0x0040
)

// Define HWND_TOPMOST explicitly as a uintptr using a bitwise complement or signed conversion
// ^uintptr(0) flips all bits of 0 to 1, which represents -1 in memory
var hwndTopmost = ^uintptr(0)

type RECT struct {
	Left, Top, Right, Bottom int32
}

// CenterAndFocusWindow brings the current console to the center and pins it focused on top.
func CenterAndFocusWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	// 1. Get screen dimensions
	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	// 2. Get current window size
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	winWidth := rect.Right - rect.Left
	winHeight := rect.Bottom - rect.Top

	// 3. Calculate centered coordinates
	x := (int32(screenWidth) - winWidth) / 2
	y := (int32(screenHeight) - winHeight) / 2

	// 4. Center, set as TopMost (always on top), and pull focus
	// 4. Center, set as TopMost (always on top), and pull focus
	procSetWindowPos.Call(hwnd, hwndTopmost, uintptr(x), uintptr(y), 0, 0, SWP_NOSIZE|SWP_SHOWWINDOW)
	procSetForegroundWindow.Call(hwnd)
}

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
