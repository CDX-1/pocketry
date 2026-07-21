package config

import _ "embed"

//go:embed default.toml
var defaultConfig []byte

func DefaultConfig() []byte {
	config := make([]byte, len(defaultConfig))
	copy(config, defaultConfig)

	return config
}