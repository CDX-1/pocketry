package instance

import "path/filepath"

const (
	configFilename          = "pocketry.toml"
	markerFilename          = "pocketry.instance"
	databaseRelativePath    = "data/pocketry.db"
	accessTokenRelativePath = "secrets/access-token.key"
	opaqueRelativePath      = "secrets/opaque.key"
)

func (i *Instance) ConfigPath() string {
	return filepath.Join(i.RootDir, configFilename)
}

func (i *Instance) MarkerPath() string {
	return filepath.Join(i.RootDir, markerFilename)
}

func (i *Instance) DatabasePath() string {
	return filepath.Join(i.RootDir, databaseRelativePath)
}

func (i *Instance) AccessTokenSecretPath() string {
	return filepath.Join(i.RootDir, accessTokenRelativePath)
}

func (i *Instance) OpaqueServerKeyPath() string {
	return filepath.Join(i.RootDir, opaqueRelativePath)
}