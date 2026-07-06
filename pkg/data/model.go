// Package data manages data used by the system
package data

import (
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
	OpenedChirp  string       `json:"opened_chirp"`
	Chirps       []ChirpModel `json:"chirps"`
}

const CurrentSchemaVersion = 1

type ChirpModel struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"` // The user only types this
	PlaySound   bool      `json:"play_sound"`
	DurationMin int       `json:"duration_min"`
	AutoRepeat  bool      `json:"auto_repeat"`
	IsActive    bool      `json:"is_active"`
	NextRun     time.Time `json:"next_run"`
	IsOpened    bool      `json:"is_opened"`
}

// GetChirpsFilePath Uses AppData on Windows, ~/.config on Linux automatically
func GetChirpsFilePath() string {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		baseDir = "."
	}
	dir := filepath.Join(baseDir, "chirp")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "chirps.json")
}
