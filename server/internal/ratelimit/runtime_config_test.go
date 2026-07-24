package ratelimit

import (
	"testing"
	"time"

	"github.com/CDX-1/pocketry/internal/config"
)

// ensures that the runtime config can be generated from the app config
func TestConfigFromAppConfig(t *testing.T) {
	appConfig := config.RateLimitConfig{
		Enabled:                true,
		EntryTTLSeconds:        900,
		CleanupIntervalSeconds: 60,

		Global: config.RateLimitPolicyConfig{
			Requests:      120,
			WindowSeconds: 60,
			Burst:         30,
		},

		RegistrationStart: config.RateLimitPolicyConfig{
			Requests:      5,
			WindowSeconds: 60,
			Burst:         3,
		},

		RegistrationFinish: config.RateLimitPolicyConfig{
			Requests:      10,
			WindowSeconds: 60,
			Burst:         5,
		},

		RegistrationUsername: config.RateLimitPolicyConfig{
			Requests:      5,
			WindowSeconds: 60,
			Burst:         3,
		},

		LoginStart: config.RateLimitPolicyConfig{
			Requests:      10,
			WindowSeconds: 60,
			Burst:         5,
		},

		LoginFinish: config.RateLimitPolicyConfig{
			Requests:      15,
			WindowSeconds: 60,
			Burst:         5,
		},

		LoginUsername: config.RateLimitPolicyConfig{
			Requests:      5,
			WindowSeconds: 60,
			Burst:         3,
		},

		VaultRead: config.RateLimitPolicyConfig{
			Requests:      120,
			WindowSeconds: 60,
			Burst:         30,
		},

		VaultWrite: config.RateLimitPolicyConfig{
			Requests:      30,
			WindowSeconds: 60,
			Burst:         10,
		},
	}

	got := ConfigFromAppConfig(appConfig)

	if !got.Enabled {
		t.Fatal("Enabled = false, want true")
	}

	if got.EntryTTL != 15*time.Minute {
		t.Fatalf(
			"EntryTTL = %v, want %v",
			got.EntryTTL,
			15*time.Minute,
		)
	}

	if got.CleanupInterval != time.Minute {
		t.Fatalf(
			"CleanupInterval = %v, want %v",
			got.CleanupInterval,
			time.Minute,
		)
	}

	tests := []struct {
		name       string
		policy     Policy
		wantBurst  int
		wantRetry  time.Duration
	}{
		{
			name:      "global",
			policy:    got.Global,
			wantBurst: 30,
			wantRetry: time.Second,
		},
		{
			name:      "registration start",
			policy:    got.RegistrationStart,
			wantBurst: 3,
			wantRetry: 12 * time.Second,
		},
		{
			name:      "registration finish",
			policy:    got.RegistrationFinish,
			wantBurst: 5,
			wantRetry: 6 * time.Second,
		},
		{
			name:      "registration username",
			policy:    got.RegistrationUsername,
			wantBurst: 3,
			wantRetry: 12 * time.Second,
		},
		{
			name:      "login start",
			policy:    got.LoginStart,
			wantBurst: 5,
			wantRetry: 6 * time.Second,
		},
		{
			name:      "login finish",
			policy:    got.LoginFinish,
			wantBurst: 5,
			wantRetry: 4 * time.Second,
		},
		{
			name:      "login username",
			policy:    got.LoginUsername,
			wantBurst: 3,
			wantRetry: 12 * time.Second,
		},
		{
			name:      "vault read",
			policy:    got.VaultRead,
			wantBurst: 30,
			wantRetry: time.Second,
		},
		{
			name:      "vault write",
			policy:    got.VaultWrite,
			wantBurst: 10,
			wantRetry: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.policy.Burst != tt.wantBurst {
				t.Fatalf(
					"Burst = %d, want %d",
					tt.policy.Burst,
					tt.wantBurst,
				)
			}

			if tt.policy.RetryAfter != tt.wantRetry {
				t.Fatalf(
					"RetryAfter = %v, want %v",
					tt.policy.RetryAfter,
					tt.wantRetry,
				)
			}
		})
	}
}