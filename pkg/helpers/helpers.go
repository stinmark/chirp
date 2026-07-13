// Package helpers where the utility functions used by other parts live
package helpers

import (
	"os/exec"
)

func FindTerminal() string {
	terminals := []string{"kitty", "alacritty", "foot", "gnome-terminal", "konsole", "xterm"}
	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}
	return "xterm"
}

func Ternary(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}
