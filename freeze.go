package main

import (
	"log"
	"os"
	"os/exec"
)

// IsWayland checks if the user is running a Wayland session
func IsWayland() bool {
	return os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

// FreezeDesktop blocks interaction with other windows universally
func FreezeDesktop() {
	log.Println("🔒 Freezing desktop interactions...")

	// 1. Check if we are running natively on Hyprland
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		_ = exec.Command("hyprctl", "dispatch", "submap", "void").Run()
		return
	}

	// 2. Fallback for X11 Window Managers (i3, bspwm, Xfce, etc.)
	if !IsWayland() {
		// Disable Master Pointer (ID 2) and Master Keyboard (ID 3)
		_ = exec.Command("xinput", "disable", "2").Run()
		_ = exec.Command("xinput", "disable", "3").Run()
	}
}

// UnfreezeDesktop restores input control back to all applications
func UnfreezeDesktop() {
	log.Println("🔓 Unfreezing desktop interactions...")

	// 1. Restore Hyprland
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		_ = exec.Command("hyprctl", "dispatch", "submap", "reset").Run()
		return
	}

	// 2. Restore X11 Window Managers
	if !IsWayland() {
		_ = exec.Command("xinput", "enable", "2").Run()
		_ = exec.Command("xinput", "enable", "3").Run()
	}
}
