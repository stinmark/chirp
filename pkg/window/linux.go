//go:build linux

package window

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CenterAndFocusWindow remains an empty stub because on Linux, window positioning
func CenterAndFocusWindow() {}

func SpawnFloatingWindow(terminalApp, executable, chirpID string) error {
	var cmd *exec.Cmd
	const uniqueTitle = "chirp-popup"
	uiArg := "--ui=popup"
	idArg := fmt.Sprintf("--chirp-id=%s", chirpID)

	// Clean out existing active popups before creating a fresh window overlay
	_ = exec.Command("pkill", "-f", uniqueTitle).Run()

	// 1. Detect the current Desktop Environment and inject specific rules
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	if strings.Contains(desktop, "hyprland") {
		injectRuntimeHyprlandRules(uniqueTitle)
	} else if strings.Contains(desktop, "gnome") {
		injectGNOMERules(uniqueTitle)
	} else if strings.Contains(desktop, "kde") {
		injectKDERules(uniqueTitle)
	}

	// 2. Build the command matching your terminal
	switch terminalApp {
	case "kitty":
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	case "alacritty":
		cmd = exec.Command(terminalApp, "--class", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	case "foot":
		cmd = exec.Command(terminalApp, "--app-id", uniqueTitle, "--title", uniqueTitle, "-e", executable, uiArg, idArg)
	default:
		cmd = exec.Command(terminalApp, "--title", uniqueTitle, "--", executable, uiArg, idArg)
	}

	cmd.Env = os.Environ()
	return cmd.Start()
}

// hyprlandRuleOnce ensures we only ever register the anonymous chirp-popup rule once per process lifetime. Since v0.55, hl.window_rule() calls are
var hyprlandRuleOnce sync.Once

// injectRuntimeHyprlandRules registers a window rule for uniqueTitle at
// For anyone still on Hyprland < 0.55, `eval` won't exist, so we fall back
// to the legacy windowrulev2 keyword syntax if the eval call fails.
func injectRuntimeHyprlandRules(className string) {
	hyprlandRuleOnce.Do(func() {
		if err := applyHyprlandLuaRule(className); err != nil {
			applyLegacyHyprlandRules(className)
		}
	})
}

// applyHyprlandLuaRule
//
//	hl.window_rule({
//	  name = "chirp-center",
//	  match = { title = "^(chirp-popup)$", class = "^(chirp-popup)$" },
//	  float = true,
//	  size = "500 420",
//	  pin = true,
//	  center = true,
//	  dim_around = true,
//	  stay_focused = true,
//	})
func applyHyprlandLuaRule(className string) error {
	regex := fmt.Sprintf("^(%s)$", className)

	// Single quotes for the Lua string literals so we don't have to escape
	// double quotes; this whole thing is passed as ONE argv element to
	// exec.Command (no shell involved), so there's no shell-quoting to
	// worry about either.
	luaRule := fmt.Sprintf(
		`hl.window_rule({ name = 'chirp-center', match = { title = '%s', class = '%s' }, float = true, size = '500 420', pin = true, center = true, dim_around = true, stay_focused = true })`,
		regex, regex,
	)

	out, err := exec.Command("hyprctl", "eval", luaRule).CombinedOutput()
	if err != nil {
		return fmt.Errorf("hyprctl eval failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	result := strings.TrimSpace(string(out))
	if !strings.EqualFold(result, "ok") {
		// hyprctl eval returns "ok" on success; anything else is a Lua
		// error raised from inside hl.window_rule (bad match table, etc).
		return fmt.Errorf("hyprctl eval returned unexpected result: %s", result)
	}
	return nil
}

// applyLegacyHyprlandRules is the pre-0.55 fallback, using the deprecated
// windowrulev2 keyword syntax. Only used if hyprctl eval isn't understood
// (i.e. an older Hyprland build without the Lua config system).
func applyLegacyHyprlandRules(className string) {
	rules := []string{
		fmt.Sprintf("float, class:^(%s)$", className),
		fmt.Sprintf("pin, class:^(%s)$", className),
		fmt.Sprintf("center, class:^(%s)$", className),
		fmt.Sprintf("stayfocused, class:^(%s)$", className),
		fmt.Sprintf("dimaround, class:^(%s)$", className),
		fmt.Sprintf("size 500 420, class:^(%s)$", className),

		fmt.Sprintf("float, title:^(%s)$", className),
		fmt.Sprintf("pin, title:^(%s)$", className),
		fmt.Sprintf("center, title:^(%s)$", className),
		fmt.Sprintf("stayfocused, title:^(%s)$", className),
		fmt.Sprintf("dimaround, title:^(%s)$", className),
		fmt.Sprintf("size 500 420, title:^(%s)$", className),
	}
	for _, rule := range rules {
		_ = exec.Command("hyprctl", "keyword", "windowrulev2", rule).Run()
	}
}

func injectGNOMERules(className string) {
	go centerAndPinX11Window(className, 500, 420)
}

func injectKDERules(className string) {
	go applyKDERuntimeRule(className)
}

// centerAndPinX11Window polls for a window by class (retrying briefly since SpawnFloatingWindow's cmd.Start() returns before the window is mapped),
func centerAndPinX11Window(className string, width, height int) {
	if _, err := exec.LookPath("xdotool"); err != nil {
		return // no X11/XWayland tooling available; nothing we can do
	}

	var winID string
	for i := 0; i < 20; i++ { // retry for ~2s while the window is created/mapped
		out, err := exec.Command("xdotool", "search", "--class", "^"+className+"$").Output()
		if err == nil {
			if id := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0]); id != "" {
				winID = id
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if winID == "" {
		return // window never showed up under XWayland — likely a native Wayland client
	}

	screenW, screenH := screenSizeX11()
	x := (screenW - width) / 2
	y := (screenH - height) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	_ = exec.Command("xdotool", "windowsize", winID, strconv.Itoa(width), strconv.Itoa(height)).Run()
	_ = exec.Command("xdotool", "windowmove", winID, strconv.Itoa(x), strconv.Itoa(y)).Run()
	_ = exec.Command("xdotool", "windowactivate", winID).Run()

	if _, err := exec.LookPath("wmctrl"); err == nil {
		// Approximates Hyprland's "pin"/"stayfocused": keeps the popup
		// above every other window for as long as it's open.
		_ = exec.Command("wmctrl", "-i", "-r", winID, "-b", "add,above").Run()
	}
}

// screenSizeX11 returns the primary display's resolution, falling back to a
// common default if it can't be determined.
func screenSizeX11() (int, int) {
	out, err := exec.Command("xdotool", "getdisplaygeometry").Output()
	if err == nil {
		if parts := strings.Fields(strings.TrimSpace(string(out))); len(parts) == 2 {
			w, errW := strconv.Atoi(parts[0])
			h, errH := strconv.Atoi(parts[1])
			if errW == nil && errH == nil {
				return w, h
			}
		}
	}
	return 1920, 1080
}

// applyKDERuntimeRule writes (or reuses) a static KWin window rule for“ className into kwinrulesrc and asks KWin to reload it via D-Bus. This is
func applyKDERuntimeRule(className string) {
	writeTool := firstAvailable("kwriteconfig6", "kwriteconfig5")
	readTool := firstAvailable("kreadconfig6", "kreadconfig5")
	dbusTool := firstAvailable("qdbus6", "qdbus")
	if writeTool == "" || readTool == "" {
		return // Plasma config tools not found; not a KDE session (or minimal install)
	}

	// Deterministic (not random UUID) group name: lets us detect "already
	// registered" across process restarts and avoids piling up duplicate
	// rule groups in kwinrulesrc every time the app runs.
	const groupName = "chirp-popup-rule"

	existing, _ := exec.Command(readTool, "--file", "kwinrulesrc", "--group", "General", "--key", "rules").Output()
	rulesList := strings.TrimSpace(string(existing))
	if strings.Contains(rulesList, groupName) {
		return // already registered from a previous run
	}

	countOut, _ := exec.Command(readTool, "--file", "kwinrulesrc", "--group", "General", "--key", "count", "--default", "0").Output()
	count, _ := strconv.Atoi(strings.TrimSpace(string(countOut)))
	count++

	newRulesList := groupName
	if rulesList != "" {
		newRulesList = rulesList + "," + groupName
	}

	fields := [][2]string{
		{"Description", "chirp-popup centered + kept above"},
		{"wmclass", className},
		{"wmclassmatch", "1"},  // ExactMatch
		{"placement", "6"},     // Placement::Policy::Centered
		{"placementrule", "3"}, // Apply (only on initial mapping)
		{"above", "true"},
		{"aboverule", "2"}, // Force (keep above enforced for the window's whole life)
		{"acceptfocus", "true"},
		{"acceptfocusrule", "2"}, // Force
	}
	for _, kv := range fields {
		_ = exec.Command(writeTool, "--file", "kwinrulesrc", "--group", groupName, "--key", kv[0], kv[1]).Run()
	}
	_ = exec.Command(writeTool, "--file", "kwinrulesrc", "--group", "General", "--key", "count", strconv.Itoa(count)).Run()
	_ = exec.Command(writeTool, "--file", "kwinrulesrc", "--group", "General", "--key", "rules", newRulesList).Run()

	if dbusTool != "" {
		_ = exec.Command(dbusTool, "org.kde.KWin", "/KWin", "reconfigure").Run()
	}
}

func firstAvailable(candidates ...string) string {
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	return ""
}
