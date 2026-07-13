// Package storage manages data used by the system
package storage

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
	Version     int          `json:"version"`
	OpenedChirp string       `json:"opened_chirp"`
	Chirps      []ChirpModel `json:"chirps"`
}

const CurrentSchemaVersion = 1

type ChirpModel struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	PlaySound   bool      `json:"play_sound"`
	DurationMin int       `json:"duration_min"`
	AutoRepeat  bool      `json:"auto_repeat"`
	IsActive    bool      `json:"is_active"`
	NextRun     time.Time `json:"next_run"`
}

// GetStorageFilePath Uses AppData on Windows, ~/.config on Linux automatically
func GetStorageFilePath() string {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		baseDir = "."
	}
	dir := filepath.Join(baseDir, "chirp")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "storage.json")
}

// Load reads the full storage structure from disk, handling migration if necessary.
func Load() (ChirpStorage, error) {
	path := GetStorageFilePath()
	var storage ChirpStorage

	if _, err := os.Stat(path); os.IsNotExist(err) {
		storage.Version = CurrentSchemaVersion
		return storage, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return storage, err
	}

	err = json.Unmarshal(data, &storage)
	if err != nil || storage.Version == 0 {
		var legacyChirps []ChirpModel
		if legacyErr := json.Unmarshal(data, &legacyChirps); legacyErr == nil {
			log.Println("🔄 Old tasks.json format detected. Migrating schema to Version 1...")

			storage.Version = CurrentSchemaVersion
			storage.Chirps = legacyChirps

			_ = Save(storage)
			return storage, nil
		}

		storage.Version = CurrentSchemaVersion
		return storage, nil
	}

	return storage, nil
}

// Save writes the entire ChirpStorage structure to disk.
func Save(storage ChirpStorage) error {
	path := GetStorageFilePath()
	storage.Version = CurrentSchemaVersion // Ensure the schema version stays accurate

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c ChirpModel) FilterValue() string {
	return c.Message + " " + c.ID
}

// GenerateShortID creates a 4-character unique random hex string (e.g., "a2f9")
func GenerateShortID() string {
	bytes := make([]byte, 2)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("0504")
	}
	return hex.EncodeToString(bytes)
}
