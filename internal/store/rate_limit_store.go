package store

import (
	"sync"
	"time"

	"github.com/wricheek84/source_asia_backend/internal/model"
)


type RateLimitStore struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	stats    map[string]*model.UserStats
}


func NewRateLimitStore() *RateLimitStore {
	return &RateLimitStore{
		requests: make(map[string][]time.Time),
		stats:    make(map[string]*model.UserStats),
	}
}
func (s *RateLimitStore) IncrementAndCheck(userID string) (bool, model.UserStats) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	oneMinuteAgo := now.Add(-60 * time.Second)

	
	if s.stats[userID] == nil {
		s.stats[userID] = &model.UserStats{}
	}
	userStat := s.stats[userID]

	
	var validTimestamps []time.Time
	for _, t := range s.requests[userID] {
		if t.After(oneMinuteAgo) {
			validTimestamps = append(validTimestamps, t)
		}
	}

	
	if len(validTimestamps) < 5 {
		
		validTimestamps = append(validTimestamps, now)
		s.requests[userID] = validTimestamps

		userStat.AcceptedCurrentWindow = len(validTimestamps)
		return true, *userStat
	}

	
	s.requests[userID] = validTimestamps
	userStat.AcceptedCurrentWindow = len(validTimestamps)
	userStat.RejectedCumulative++
	return false, *userStat
}

func (s *RateLimitStore) GetStats(userID string) model.UserStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	userStat := s.stats[userID]
	if userStat == nil {
		return model.UserStats{}
	}

	now := time.Now()
	oneMinuteAgo := now.Add(-60 * time.Second)

	
	var validTimestamps []time.Time
	for _, t := range s.requests[userID] {
		if t.After(oneMinuteAgo) {
			validTimestamps = append(validTimestamps, t)
		}
	}

	s.requests[userID] = validTimestamps
	userStat.AcceptedCurrentWindow = len(validTimestamps)

	return *userStat
}