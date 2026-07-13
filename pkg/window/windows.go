//go:build windows

package window

import (
	"fmt"
	"os/exec"
	"strings"
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
	// Resolve to an absolute path the same way exec.Command would, so
	// behavior matches what callers expect when they pass a bare name.
	resolved, err := exec.LookPath(executable)
	if err != nil {
		resolved = executable
	}

	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	cmdLine := strings.Join([]string{
		quoteWindowsArg(resolved),
		quoteWindowsArg(uiArg),
		quoteWindowsArg(idArg),
	}, " ")

	appNamePtr, err := windows.UTF16PtrFromString(resolved)
	if err != nil {
		return fmt.Errorf("invalid executable path %q: %w", resolved, err)
	}
	cmdLinePtr, err := windows.UTF16PtrFromString(cmdLine)
	if err != nil {
		return fmt.Errorf("invalid command line: %w", err)
	}

	si := new(windows.StartupInfo)
	si.Cb = uint32(unsafe.Sizeof(*si))
	// Deliberately leaving si.Flags at 0 (i.e. NOT setting
	// STARTF_USESTDHANDLES) is the whole fix here. os/exec's Cmd.Start()
	// always sets that flag on Windows and points StdInput/StdOutput/
	// StdErr at the null device whenever Stdin/Stdout/Stderr are nil —
	// which is why the popup window opened but never showed any Bubble
	// Tea output: its console existed, but stdout was wired to NUL
	// instead of to that console's screen buffer. With Flags left at 0,
	// CreateProcess ignores the (unset) Std* fields entirely and lets the
	// console freshly allocated by CREATE_NEW_CONSOLE become the child's
	// own stdin/stdout/stderr, exactly like double-clicking the exe would.
	pi := new(windows.ProcessInformation)

	err = windows.CreateProcess(
		appNamePtr,
		cmdLinePtr,
		nil,   // default process security
		nil,   // default thread security
		false, // don't inherit the caller's handles
		windows.CREATE_NEW_CONSOLE,
		nil, // nil env -> inherit the calling process's environment
		nil, // nil dir -> inherit the calling process's working directory
		si,
		pi,
	)
	if err != nil {
		return fmt.Errorf("CreateProcess failed: %w", err)
	}
	// We're not tracking/waiting on this process (fire-and-forget popup),
	// so release the handles CreateProcess gave us.
	defer windows.CloseHandle(pi.Thread)
	defer windows.CloseHandle(pi.Process)

	return nil
}

// quoteWindowsArg quotes a single argument using the same escaping rules
// the Windows C runtime (and CreateProcess's command-line parser) expects:
// wrap in quotes if the argument contains whitespace or a quote, doubling
// any backslashes that immediately precede a quote (or the closing quote).
func quoteWindowsArg(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\n\v\"") {
		return s
	}

	var b strings.Builder
	b.WriteByte('"')
	slashes := 0
	for _, r := range s {
		switch r {
		case '\\':
			slashes++
			b.WriteRune(r)
		case '"':
			for ; slashes > 0; slashes-- {
				b.WriteByte('\\')
			}
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			slashes = 0
			b.WriteRune(r)
		}
	}
	for ; slashes > 0; slashes-- {
		b.WriteByte('\\')
	}
	b.WriteByte('"')
	return b.String()
}
