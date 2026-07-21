package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config

	meta, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Config{}, fmt.Errorf(
			"unknown config field: %s",
			undecoded[0],
		)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Version != CurrentVersion {
		return fmt.Errorf(
			"unsupported config version %d",
			c.Version,
		)
	}

	if c.Server.Host == "" {
		return errors.New("server.host is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}

	ttl, err := time.ParseDuration(c.Security.AccessTokenTTL)
	if err != nil {
		return fmt.Errorf(
			"invalid security.access_token_ttl: %w",
			err,
		)
	}

	if ttl <= 0 {
		return errors.New("security.access_token_ttl must be positive")
	}

	for _, origin := range c.CORS.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			return err
		}
	}

	if err := validateRateLimitConfig(c.RateLimit); err != nil {
		return err
	}

	return nil
}

func validateOrigin(origin string) error {
	parsed, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf(
			"invalid CORS origin %q: %w", origin, err,
		)
	}

	switch parsed.Scheme {
	case "http", "https", "chrome-extension", "moz-extension":
	default:
		return fmt.Errorf("unsupported CORS origin schema in %q", origin)
	}

	if parsed.Host == "" {
		return fmt.Errorf("CORS origin is missing a host: %q", origin)
	}
	
	if parsed.Path != "" && parsed.Path != "/" {
		return fmt.Errorf("CORS origin must not contain a path: %q", origin)
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("CORS origin must not contain a query or fragment: %q", origin)
	}

	return nil
}

func validateRateLimitConfig(cfg RateLimitConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if cfg.EntryTTLSeconds <= 0 {
		return errors.New("rate_limit.entry_ttl_seconds must be greater than zero")
	}

	if cfg.CleanupIntervalSeconds <= 0 {
		return errors.New("rate_limit.cleanup_interval_seconds must be greater than zero")
	}

	if cfg.CleanupIntervalSeconds > cfg.EntryTTLSeconds {
		return errors.New("rate_limit.cleanup_interval_secodns must not exceed entry_ttl_seconds")
	}

	policies := map[string]RateLimitPolicyConfig{
		"global":			 	 cfg.Global,
		"registration_start":	 cfg.RegistrationStart,
		"registration_finish":	 cfg.RegistrationFinish,
		"login_start":			 cfg.LoginStart,
		"login_finish":			 cfg.LoginFinish,
		"login_username":		 cfg.LoginUsername,
		"registration_username": cfg.RegistrationUsername,
		"vault_read":			 cfg.VaultRead,
		"vault_write":			 cfg.VaultWrite,
	}

	for name, policy := range policies {
		if err := validateRateLimitPolicy(name, policy); err != nil {
			return err
		}
	}

	return nil
}

func validateRateLimitPolicy(name string, policy RateLimitPolicyConfig) error {
	if policy.Requests <= 0 {
		return fmt.Errorf("rate_limit.%s.requests must be greater than zero", name)
	}

	if policy.WindowSeconds <= 0 {
		return fmt.Errorf("rate_limit.%s.window_seconds must be greater than zero", name)
	}

	if policy.Burst <= 0 {
		return fmt.Errorf("rate_limit.%s.burst must be greater than zero", name)
	}

	return nil
}