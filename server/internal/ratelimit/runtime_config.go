package ratelimit

import (
	"time"

	"github.com/CDX-1/pocketry/internal/config"
)

type Config struct {
	Enabled bool

	EntryTTL       time.Duration
	CleanupInterval time.Duration

	Global Policy

	RegistrationStart  Policy
	RegistrationFinish Policy
	RegistrationUsername Policy

	LoginStart    Policy
	LoginFinish   Policy
	LoginUsername Policy

	VaultRead  Policy
	VaultWrite Policy
}

func ConfigFromAppConfig(cfg config.RateLimitConfig) Config {
	return Config{
		Enabled: cfg.Enabled,

		EntryTTL: time.Duration(
			cfg.EntryTTLSeconds,
		) * time.Second,

		CleanupInterval: time.Duration(
			cfg.CleanupIntervalSeconds,
		) * time.Second,

		Global: PolicyFromConfig(cfg.Global),

		RegistrationStart: PolicyFromConfig(
			cfg.RegistrationStart,
		),

		RegistrationFinish: PolicyFromConfig(
			cfg.RegistrationFinish,
		),

		RegistrationUsername: PolicyFromConfig(
			cfg.RegistrationUsername,
		),

		LoginStart: PolicyFromConfig(
			cfg.LoginStart,
		),

		LoginFinish: PolicyFromConfig(
			cfg.LoginFinish,
		),

		LoginUsername: PolicyFromConfig(
			cfg.LoginUsername,
		),

		VaultRead: PolicyFromConfig(
			cfg.VaultRead,
		),

		VaultWrite: PolicyFromConfig(
			cfg.VaultWrite,
		),
	}
}