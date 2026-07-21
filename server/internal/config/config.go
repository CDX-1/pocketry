package config

const CurrentVersion = 1

type Config struct {
	Version   int             `toml:"version"`
	Server    ServerConfig    `toml:"server"`
	Security  SecurityConfig  `toml:"security"`
	CORS      CORSConfig      `toml:"cors"`
	RateLimit RateLimitConfig `toml:"ratelimit"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type SecurityConfig struct {
	AccessTokenTTL string `toml:"access_token_ttl"`
}

type CORSConfig struct {
	AllowedOrigins         []string `toml:"allowed_origins"`
	AllowFirefoxExtensions bool     `toml:"allow_firefox_extensions"`
}

type RateLimitConfig struct {
	Enabled bool `toml:"enabled"`

	EntryTTLSeconds        int64 `toml:"entry_ttl_seconds"`
	CleanupIntervalSeconds int64 `toml:"cleanup_interval_seconds"`

	Global RateLimitPolicyConfig `toml:"global"`

	RegistrationStart  RateLimitPolicyConfig `toml:"registration_start"`
	RegistrationFinish RateLimitPolicyConfig `toml:"registration_finish"`

	LoginStart  RateLimitPolicyConfig `toml:"login_start"`
	LoginFinish RateLimitPolicyConfig `toml:"login_finish"`

	LoginUsername        RateLimitPolicyConfig `toml:"login_username"`
	RegistrationUsername RateLimitPolicyConfig `toml:"registration_username"`

	VaultRead  RateLimitPolicyConfig `toml:"vault_read"`
	VaultWrite RateLimitPolicyConfig `toml:"vault_write"`
}

type RateLimitPolicyConfig struct {
	Requests      int   `toml:"requests"`
	WindowSeconds int64 `toml:"window_seconds"`
	Burst         int   `toml:"burst"`
}