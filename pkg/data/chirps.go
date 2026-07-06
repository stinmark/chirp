package data

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"time"
)

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

func LoadChirps() ([]ChirpModel, error) {
	path := GetChirpsFilePath()
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
	path := GetChirpsFilePath()

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
