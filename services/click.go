package services

import (
	"sync"

	"adserving/models"
)

type ClickService struct {
	mu    sync.Mutex
	stats map[models.ClickStatKey]int64
}

func NewClickService() *ClickService {
	return &ClickService{stats: make(map[models.ClickStatKey]int64)}
}

func (s *ClickService) IncrementClick(key models.ClickStatKey) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats[key]++
	return s.stats[key]
}
