package ratelimit

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func testPolicy(burst int) Policy {
	return Policy{
		Limit: 		rate.Every(time.Hour),
		Burst: 		burst,
		RetryAfter: 30 * time.Second,
	}
}

// ensures that the ratelimit store allows requests up to the burst limit
func TestStoreAllowsRequestsUpToBurst(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	policy := testPolicy(2)

	if !store.Allow("login:192.0.0.67", policy) {
		t.Fatal("first request was rejected")
	}

	if !store.Allow("login:192.0.0.67", policy) {
		t.Fatal("second request was rejected")
	}

	if store.Allow("login:192.0.0.67", policy) {
		t.Error("third request was allowed after burst was exhausted")
	}
}

// ensures that ratelimit buckets are separated by key
func TestStoreSeparatesKeys(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	policy := testPolicy(1)

	if !store.Allow("login:192.0.0.67", policy) {
		t.Fatal("first key was rejected")
	}

	if store.Allow("login:192.0.0.67", policy) {
		t.Fatal("first key was not ratelimited")
	}

	if !store.Allow("login:192.0.0.68", policy) {
		t.Fatal("second key incorrectly shared first key's bucket")
	}
}

// ensures that ratelimit buckets are separated by namespace
func TestStoreSeparatesNamespaces(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	policy := testPolicy(1)

	if !store.Allow("login:192.0.0.67", policy) {
		t.Fatal("login request was rejected")
	}

	if !store.Allow("register:192.0.0.67", policy) {
		t.Fatal("registration namespace shared the login bucket")
	}
}

// ensures that requests with empty keys are rejected
func TestStoreRejectsEmptyKey(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	if store.Allow("", testPolicy(1)) {
		t.Fatal("empty key was allowed")
	}
}

// ensures that the store rejects an invalid ratelimit policy
func TestStoreRejectsInvalidPolicy(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	tests := []struct {
		name   string
		policy Policy
	}{
		{
			name: "zero limit",
			policy: Policy{
				Limit: 0,
				Burst: 1,
			},
		},
		{
			name: "negative limit",
			policy: Policy{
				Limit: -1,
				Burst: 1,
			},
		},
		{
			name: "zero burst",
			policy: Policy{
				Limit: rate.Every(time.Minute),
				Burst: 0,
			},
		},
		{
			name: "negative burst",
			policy: Policy{
				Limit: rate.Every(time.Minute),
				Burst: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if store.Allow("test-key", tt.policy) {
				t.Fatal("invalid policy was allowed")
			}
		})
	}
}

// ensures that the store removes expired entries when requested
func TestStoreRemovesExpiredEntries(t *testing.T) {
	now := time.Date(2026, time.July, 22, 16, 0, 0, 0, time.UTC)
	currentTime := now
	
	store := newStore(10 * time.Minute, time.Hour, func() time.Time {
		return currentTime
	})
	t.Cleanup(store.Close)

	if !store.Allow("expired-key", testPolicy(1)) {
		t.Fatal("request was rejected")
	}

	if len(store.entries) != 1 {
		t.Fatalf("entry counmt = %d, want 1", len(store.entries))
	}

	currentTime = currentTime.Add(11 * time.Minute)
	store.removeExpired()

	if len(store.entries) != 0 {
		t.Fatalf("entry count = %d, want 0", len(store.entries))
	}
}

// ensures that the store does not remove recently added entries
func TestStoreKeepsRecentEntries(t *testing.T) {
	now := time.Date(2026, time.July, 22, 16, 0, 0, 0, time.UTC)
	currentTime := now

	store := newStore(10 * time.Minute, time.Hour, func() time.Time {
		return currentTime
	})
	t.Cleanup(store.Close)

	if !store.Allow("recent-key", testPolicy(1)) {
		t.Fatal("request was rejected")
	}

	currentTime = currentTime.Add(9 * time.Minute)
	store.removeExpired()

	if len(store.entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(store.entries))
	}
}

// ensures that the store updates the policy for existing keys
func TestStoreUpdatesPolicy(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	key := "login:192.0.2.10"

	initialPolicy := Policy{
		Limit:      rate.Every(time.Hour),
		Burst:      1,
		RetryAfter: time.Hour,
	}

	if !store.Allow(key, initialPolicy) {
		t.Fatal("first request was rejected")
	}

	updatedPolicy := Policy{
		Limit:      rate.Every(30 * time.Minute),
		Burst:      3,
		RetryAfter: 30 * time.Minute,
	}

	_ = store.Allow(key, updatedPolicy)

	store.mu.Lock()
	current := store.entries[key]
	store.mu.Unlock()

	if current.policy.Limit != updatedPolicy.Limit {
		t.Fatalf(
			"stored limit = %v, want %v",
			current.policy.Limit,
			updatedPolicy.Limit,
		)
	}

	if current.policy.Burst != updatedPolicy.Burst {
		t.Fatalf(
			"stored burst = %d, want %d",
			current.policy.Burst,
			updatedPolicy.Burst,
		)
	}

	if current.limiter.Limit() != updatedPolicy.Limit {
		t.Fatalf(
			"limiter limit = %v, want %v",
			current.limiter.Limit(),
			updatedPolicy.Limit,
		)
	}

	if current.limiter.Burst() != updatedPolicy.Burst {
		t.Fatalf(
			"limiter burst = %d, want %d",
			current.limiter.Burst(),
			updatedPolicy.Burst,
		)
	}
}

// ensures that a store with zero duration uses the default duration
func TestStoreUsesDefaultDuration(t *testing.T) {
	store := NewStore(0, 0)
	t.Cleanup(store.Close)

	if store.entryTTL != 15 * time.Minute {
		t.Fatalf("entryTTL = %v, want %v", store.entryTTL, 15 * time.Minute)
	}
}

// ensures that close can be called more than once without panicking
func TestStoreCloseCanBeCalledMoreThanOnce(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)

	store.Close()
	store.Close()
}