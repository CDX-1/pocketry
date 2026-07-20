package instance

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

func (i *Instance) LoadAccessTokenSecret() ([]byte, error) {
	data, err := os.ReadFile(i.AccessTokenSecretPath())
	if err != nil {
		return nil, fmt.Errorf("read access token secret: %w", err)
	}

	lines := strings.Split(
		strings.TrimSpace(string(data)),
		"\n",
	)

	if len(lines) != 2 {
		return nil, errors.New("invalid access token secret file format")
	}

	if lines[0] != accessTokenSecretHeader {
		return nil, fmt.Errorf("unsupported access token secret format %q", lines[0])
	}

	secret, err := base64.RawURLEncoding.DecodeString(
		strings.TrimSpace(lines[1]),
	)
	if err != nil {
		return nil, fmt.Errorf("decode access token secret: %w", err)
	}

	if len(secret) < 32 {
		return nil, fmt.Errorf(
			"access token secret must contain at least 32 bytes, got %d",
			len(secret),
		)
	}

	return secret, nil
}

func (i *Instance) LoadOpaqueServerKeyMaterial() ([]byte, error) {
	data, err := os.ReadFile(i.OpaqueKeyMaterialPath())
	if err != nil {
		return nil, fmt.Errorf("read OPAQUE server key material: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("OPAQUE server key material is empty")
	}

	return data, nil
}