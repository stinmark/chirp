package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ToggleAutostart explicitly flips the configuration flag and syncs it with the OS
func ToggleAutostart() (bool, error) {
	path := getChirpsFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var storage ChirpStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return false, err
	}

	// Flip the user setting explicitly
	storage.RunOnStartup = !storage.RunOnStartup

	// Save back to disk
	newData, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, newData, 0o644); err != nil {
		return false, err
	}

	// Apply change to the OS
	err = SyncAutostartWithOS(storage.RunOnStartup, storage.Chirps)
	return storage.RunOnStartup, err
}

// IsAutostartEnabled reads the saved configuration preference
func IsAutostartEnabled() bool {
	path := getChirpsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var storage ChirpStorage
	_ = json.Unmarshal(data, &storage)
	return storage.RunOnStartup
}

// SyncAutostartWithOS creates or removes the desktop entry strictly on parameters
func SyncAutostartWithOS(enabled bool, chirps []ChirpModel) error {
	home, _ := os.UserHomeDir()
	desktopFilePath := filepath.Join(home, ".config", "autostart", "chirp.desktop")

	// Look for at least one active task
	hasActive := false
	for _, t := range chirps {
		if t.IsActive {
			hasActive = true
			break
		}
	}

	// Must be explicitly enabled by user AND have active tasks to write the file
	if enabled && hasActive {
		_ = os.MkdirAll(filepath.Dir(desktopFilePath), 0o755)
		executable, err := os.Executable()
		if err != nil {
			return err
		}

		desktopContent := "[Desktop Entry]\n" +
			"Type=Application\n" +
			"Name=Chirp Daemon\n" +
			"Exec=" + executable + " --run-daemon\n" +
			"Terminal=false\n" +
			"X-GNOME-Autostart-enabled=true\n"

		return os.WriteFile(desktopFilePath, []byte(desktopContent), 0o644)
	}

	// Otherwise remove it
	if _, err := os.Stat(desktopFilePath); err == nil {
		return os.Remove(desktopFilePath)
	}
	return nil
}
