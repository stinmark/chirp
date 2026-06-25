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
//

type TaskStorage struct {
	Version int         `json:"version"`
	Tasks   []BreakTask `json:"tasks"`
}

const CurrentSchemaVersion = 1

type BreakTask struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	DurationMin int       `json:"duration_min"`
	AutoRepeat  bool      `json:"auto_repeat"`
	IsActive    bool      `json:"is_active"`
	NextRun     time.Time `json:"next_run"`
	IsOpened    bool      `json:"is_opened"` // 👈 Add this tracking flag
}

// FilterValue satisfies the charm.land/bubbles/list.Item interface
func (t BreakTask) FilterValue() string {
	return t.Title + " " + t.ID
}

// GenerateShortID creates a 4-character unique random hex string (e.g., "a2f9")
func GenerateShortID() string {
	bytes := make([]byte, 2) // 2 bytes = 4 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp string if crypto/rand fails
		return time.Now().Format("0504")
	}
	return hex.EncodeToString(bytes)
}

func getTasksFilePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "sigcat")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "tasks.json")
}

func LoadTasks() ([]BreakTask, error) {
	path := getTasksFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []BreakTask{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 1. Try to unmarshal into the new versioned wrapper structure
	var storage TaskStorage
	err = json.Unmarshal(data, &storage)

	// 2. Fallback check: If it failed, or if Version is 0, it means it's an old legacy format file!
	if err != nil || storage.Version == 0 {
		// Try parsing it as the old legacy format raw array []BreakTask
		var legacyTasks []BreakTask
		if legacyErr := json.Unmarshal(data, &legacyTasks); legacyErr == nil {
			log.Println("🔄 Old tasks.json format detected. Migrating schema to Version 1...")

			// Fill in any new default values for features here if needed:
			/*for i := range legacyTasks {
				// Example: legacyTasks[i].IsOpened = false
			}*/

			// Save it right back to disk in the brand new versioned format automatically
			_ = SaveTasks(legacyTasks)
			return legacyTasks, nil
		}

		// If it's completely corrupted, return an empty array instead of crashing
		return []BreakTask{}, nil
	}

	return storage.Tasks, nil
}

func SaveTasks(tasks []BreakTask) error {
	path := getTasksFilePath()

	// Always wrap tasks with the current version when saving
	storage := TaskStorage{
		Version: CurrentSchemaVersion,
		Tasks:   tasks,
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
