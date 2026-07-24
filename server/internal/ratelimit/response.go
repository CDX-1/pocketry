package ratelimit

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

func WriteExceeded(w http.ResponseWriter, policy Policy) {
	retryAfter := int(math.Ceil(policy.RetryAfter.Seconds()))

	if retryAfter < 1 {
		retryAfter = 1
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "too many requests",
	})
}