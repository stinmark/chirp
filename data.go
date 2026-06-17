package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type BreakTask struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	DurationMin int       `json:"duration_min"`
	AutoRepeat  bool      `json:"auto_repeat"`
	IsActive    bool      `json:"is_active"`
	NextRun     time.Time `json:"next_run"`
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

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "sigcat")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "tasks.json")
}

func LoadTasks() ([]BreakTask, error) {
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []BreakTask{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tasks []BreakTask
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

func SaveTasks(tasks []BreakTask) error {
	path := getConfigPath()
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
