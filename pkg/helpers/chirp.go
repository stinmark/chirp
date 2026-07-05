package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ==========================================
// Chirp Action Enum Definition
// ==========================================

type ChirpAction int

const (
	ActionPopup ChirpAction = iota // Default is 0
	ActionLockscreen
	ActionSound
	ActionPopupWithSound
)

// String representation for debugging/printing
func (a ChirpAction) String() string {
	return [...]string{"popup", "lockscreen"}[a]
}

// MarshalJSON converts the enum integer into its string value for the JSON file
func (a ChirpAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON parses the string from the JSON file back into the enum integer
func (a *ChirpAction) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "popup", "": // Empty string defaults to popup
		*a = ActionPopup
	case "lockscreen":
		*a = ActionLockscreen
	default:
		return fmt.Errorf("invalid chirp action: %s", s)
	}
	return nil
}

// ==========================================
// Core Domain Structural Types
// ==========================================

type ChirpStorage struct {
	Version      int          `json:"version"`
	RunOnStartup bool         `json:"run_on_startup"`
	Chirps       []ChirpModel `json:"tasks"`
}

const CurrentSchemaVersion = 1

type ChirpModel struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Message     string      `json:"message"`
	DurationMin int         `json:"duration_min"`
	AutoRepeat  bool        `json:"auto_repeat"`
	IsActive    bool        `json:"is_active"`
	NextRun     time.Time   `json:"next_run"`
	IsOpened    bool        `json:"is_opened"`
	Action      ChirpAction `json:"action"`
}

// FilterValue satisfies the charm.land/bubbles/list.Item interface
func (c ChirpModel) FilterValue() string {
	return c.Title + " " + c.ID
}

// GenerateShortID creates a 4-character unique random hex string (e.g., "a2f9")
func GenerateShortID() string {
	bytes := make([]byte, 2)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("0504")
	}
	return hex.EncodeToString(bytes)
}

// CROSS PLATFORM: Uses AppData on Windows, ~/.config on Linux automatically
func getChirpsFilePath() string {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		baseDir = "."
	}
	dir := filepath.Join(baseDir, "chirp")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "chirps.json")
}

func LoadChirps() ([]ChirpModel, error) {
	path := getChirpsFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []ChirpModel{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var storage ChirpStorage
	err = json.Unmarshal(data, &storage)

	if err != nil || storage.Version == 0 {
		var legacyChirps []ChirpModel
		if legacyErr := json.Unmarshal(data, &legacyChirps); legacyErr == nil {
			log.Println("🔄 Old tasks.json format detected. Migrating schema to Version 1...")

			// Go defaults integer fields to 0 during unmarshal,
			// so legacy tasks automatically gain ActionPopup (0) safely.
			_ = SaveChirps(legacyChirps)
			return legacyChirps, nil
		}

		return []ChirpModel{}, nil
	}

	return storage.Chirps, nil
}

func SaveChirps(chirps []ChirpModel) error {
	path := getChirpsFilePath()

	storage := ChirpStorage{
		Version: CurrentSchemaVersion,
		Chirps:  chirps,
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
