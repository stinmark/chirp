package storage

// IsChirpOpen checks if a specific chirp ID is currently the active popup
func (s *ChirpStorage) IsChirpOpen(id string) bool {
	return s.OpenedChirp == id
}

// SetOpenedChirp sets the currently active popup ID (pass "" to clear it)
func (s *ChirpStorage) SetOpenedChirp(id string) {
	s.OpenedChirp = id
}
