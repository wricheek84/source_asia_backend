package store

import (
	"sync"
	"time"

	"github.com/wricheek84/source_asia_backend/internal/model"
)

// RateLimitStore manages the in-memory request tracking and statistics.
type RateLimitStore struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	stats    map[string]*model.UserStats
}

// NewRateLimitStore initializes an empty, clean storage instance.
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

	// 1. Initialize stats if this is a brand-new user
	if s.stats[userID] == nil {
		s.stats[userID] = &model.UserStats{}
	}
	userStat := s.stats[userID]

	// 2. Filter out old timestamps (clean the basket)
	var validTimestamps []time.Time
	for _, t := range s.requests[userID] {
		if t.After(oneMinuteAgo) {
			validTimestamps = append(validTimestamps, t)
		}
	}

	// 3. Evaluate the rate limit rules
	if len(validTimestamps) < 5 {
		// Allowed! Add current time to their history
		validTimestamps = append(validTimestamps, now)
		s.requests[userID] = validTimestamps

		userStat.AcceptedCurrentWindow = len(validTimestamps)
		return true, *userStat
	}

	// Rejected! Keep the history the same, but bump the failure count
	s.requests[userID] = validTimestamps
	userStat.AcceptedCurrentWindow = len(validTimestamps)
	userStat.RejectedCumulative++
	return false, *userStat
}
// GetStats retrieves the current statistics for a user, keeping the window accurate.
func (s *RateLimitStore) GetStats(userID string) model.UserStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	userStat := s.stats[userID]
	if userStat == nil {
		return model.UserStats{}
	}

	now := time.Now()
	oneMinuteAgo := now.Add(-60 * time.Second)

	// Clean out old timestamps so the current window count isn't stale
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