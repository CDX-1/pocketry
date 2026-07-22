package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	policy   Policy
}

type Store struct {
	mu      sync.Mutex
	entries map[string]*entry

	entryTTL time.Duration
	now      func() time.Time

	stopOnce sync.Once
	stop     chan struct{}
	done     chan struct{}
}

func NewStore(entryTTL, cleanupInterval time.Duration) *Store {
	return newStore(entryTTL, cleanupInterval, time.Now)
}

func newStore(entryTTL, cleanupInterval time.Duration, now func() time.Time) *Store {
	if entryTTL <= 0 {
		entryTTL = 15 * time.Minute
	}

	if cleanupInterval <= 0 {
		cleanupInterval = time.Minute
	}

	store := &Store{
		entries:  make(map[string]*entry),
		entryTTL: entryTTL,
		now:      now,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	go store.cleanupLoop(cleanupInterval)

	return store
}

func (s *Store) Allow(key string, policy Policy) bool {
	if key == "" {
		return false
	}

	if policy.Limit <= 0 || policy.Burst <= 0 {
		return false
	}

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.entries[key]
	if !ok {
		current = &entry{
			limiter: rate.NewLimiter(
				policy.Limit,
				policy.Burst,
			),
			lastSeen: now,
			policy:   policy,
		}

		s.entries[key] = current
	} else {
		if current.policy.Limit != policy.Limit {
			current.limiter.SetLimitAt(now, policy.Limit)
		}

		if current.policy.Burst != policy.Burst {
			current.limiter.SetBurstAt(now, policy.Burst)
		}

		current.lastSeen = now
		current.policy = policy
	}

	return current.limiter.AllowN(now, 1)
}

func (s *Store) cleanupLoop(interval time.Duration) {
	defer close(s.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()

		case <-s.stop:
			return
		}
	}
}

func (s *Store) removeExpired() {
	cutoff := s.now().Add(-s.entryTTL)

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, current := range s.entries {
		if current.lastSeen.Before(cutoff) {
			delete(s.entries, key)
		}
	}
}

func (s *Store) Close() {
	s.stopOnce.Do(func() {
		close(s.stop)
		<-s.done
	})
}
