package ratelimit

import (
	"math"
	"testing"
	"time"

	"github.com/CDX-1/pocketry/internal/config"
)

// ensures that ratelimit policy configuration can be parsed correctly
func TestPolicyFromConfig(t *testing.T) {
	cfg := config.RateLimitPolicyConfig{
		Requests:	   5,
		WindowSeconds: 60,
		Burst:		   3,
	}

	got := PolicyFromConfig(cfg)
	wantLimit := 5.0 / 60.0

	if math.Abs(float64(got.Limit)-wantLimit) > 0.000001 {
		t.Fatalf("Limit = %v, want %v", got.Limit, wantLimit)
	}

	if got.Burst != 3 {
		t.Fatalf("Burst = %d, want %d", got.Burst, 3)
	}

	if got.RetryAfter != 12 * time.Second {
		t.Fatalf("RetryAfter = %v, want %v", got.RetryAfter, 12 * time.Second)
	}
}

// ensures that when interval seconds doesn't divide evenly into window seconds,
// the retry after is rounded up
func TestPolicyFromConfigRoundsRetryAfterUp(t *testing.T) {
	cfg := config.RateLimitPolicyConfig{
		Requests:      7,
		WindowSeconds: 60,
		Burst:         2,
	}

	got := PolicyFromConfig(cfg)

	if got.RetryAfter != 9 * time.Second {
		t.Fatalf("RetryAfter = %v, want %v", got.RetryAfter, 9 * time.Second)
	}
}

// ensures that even if the calculated retry after is less than one second,
// the retry after is at least one second
func TestPolicyFromConfigHasMinimumOneSecondRetryAfter(t *testing.T) {
	cfg := config.RateLimitPolicyConfig{
		Requests:	   120,
		WindowSeconds: 60,
		Burst:		   30,
	}

	got := PolicyFromConfig(cfg)

	if got.RetryAfter != time.Second {
		t.Fatalf("RetryAfter = %v, want %v", got.RetryAfter, time.Second)
	}
}