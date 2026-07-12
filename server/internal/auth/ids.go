package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func NewFlowID() (string, error) {
	var b [32]byte;

	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate flow id: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}