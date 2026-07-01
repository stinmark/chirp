package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ==========================================
// Core Domain Structural Types
// ==========================================

type ChirpStorage struct {
	Version      int          `json:"version"`
	RunOnStartup bool         `json:"run_on_startup"`
	Chirps       []ChirpModel `json:"tasks"` // 👈 Storing ChirpModel here
}

const CurrentSchemaVersion = 1

type ChirpModel struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	DurationMin int       `json:"duration_min"`
	AutoRepeat  bool      `json:"auto_repeat"`
	IsActive    bool      `json:"is_active"`
	NextRun     time.Time `json:"next_run"`
	IsOpened    bool      `json:"is_opened"` // 👈 Track open state
}

// FilterValue satisfies the charm.land/bubbles/list.Item interface
func (c ChirpModel) FilterValue() string {
	return c.Title + " " + c.ID
}

// GenerateShortID creates a 4-character unique random hex string (e.g., "a2f9")
func GenerateShortID() string {
	bytes := make([]byte, 2) // 2 bytes = 4 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("0504")
	}
	return hex.EncodeToString(bytes)
}

func getChirpsFilePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "chirp")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "chirps.json")
}

// Updated return type to use ChirpModel instead of BreakTask
func LoadChirps() ([]ChirpModel, error) {
	path := getChirpsFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []ChirpModel{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 1. Try to unmarshal into the new versioned wrapper structure
	var storage ChirpStorage
	err = json.Unmarshal(data, &storage)

	// 2. Fallback check for legacy migration
	if err != nil || storage.Version == 0 {
		// Try parsing old raw array format directly into ChirpModel
		var legacyChirps []ChirpModel
		if legacyErr := json.Unmarshal(data, &legacyChirps); legacyErr == nil {
			log.Println("🔄 Old tasks.json format detected. Migrating schema to Version 1...")

			// Fill in any new default values for features here if needed:
			/*for i := range legacyTasks {
				// Example: legacyTasks[i].IsOpened = false
			}*/
			_ = SaveChirps(legacyChirps)
			return legacyChirps, nil
		}

		return []ChirpModel{}, nil
	}

	// 👈 Fixed: storage uses the field name 'Chirps', not 'Tasks'
	return storage.Chirps, nil
}

// Updated parameter type to use ChirpModel instead of BreakTask
func SaveChirps(chirps []ChirpModel) error {
	path := getChirpsFilePath()

	// Always wrap tasks with the current version when saving
	storage := ChirpStorage{
		Version: CurrentSchemaVersion,
		Chirps:  chirps, // 👈 Fixed: mapped to Chirps field
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
