package config

const CurrentVersion = 1

type Config struct {
	Version  int            `toml:"version"`
	Server   ServerConfig   `toml:"server"`
	Security SecurityConfig `toml:"security"`
	CORS     CORSConfig     `toml:"cors"`
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