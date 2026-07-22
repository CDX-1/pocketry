package ratelimit

import (
	"math"
	"time"

	"github.com/CDX-1/pocketry/internal/config"
	"golang.org/x/time/rate"
)

type Policy struct {
	Limit	   rate.Limit
	Burst	   int
	RetryAfter time.Duration
}

func PolicyFromConfig(cfg config.RateLimitPolicyConfig) Policy {
	window := time.Duration(cfg.WindowSeconds) * time.Second

	eventsPerSecond := float64(cfg.Requests) / window.Seconds()
	retryAfterSeconds := window.Seconds() / float64(cfg.Requests)
	retryAfter := time.Duration(math.Ceil(retryAfterSeconds)) * time.Second

	if retryAfter < time.Second {
		retryAfter = time.Second
	}

	return Policy{
		Limit:      rate.Limit(eventsPerSecond),
		Burst:      cfg.Burst,
		RetryAfter: retryAfter,
	}
}