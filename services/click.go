package services

import (
	"sync"

	"adserving/models"
)

// ClickService handles click statistics
type ClickService struct {
	mu    sync.Mutex
	stats map[models.ClickStatKey]int64
}

// NewClickService creates a new click service
func NewClickService() *ClickService {
	return &ClickService{
		stats: make(map[models.ClickStatKey]int64),
	}
}

// IncrementClick increments the click count for a given key and returns the new count
func (s *ClickService) IncrementClick(key models.ClickStatKey) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats[key]++
	return s.stats[key]
}

// GetClickCount returns the click count for a given key
func (s *ClickService) GetClickCount(key models.ClickStatKey) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats[key]
}
