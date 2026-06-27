package handlers

import (
	"encoding/base64"

	"github.com/CDX-1/pocketry/internal/vault"
)

func validTestEnvelope() vault.Envelope {
	return vault.Envelope{
		Version: vault.SupportedEnvelopeVersion,
		Cipher:  vault.SupportedCipher,
		KDF:     vault.SupportedKDF,
		KDFParams: vault.KDFParams{
			Memory:      65536,
			Iterations:  3,
			Parallelism: 2,
		},
		Salt:       base64.RawURLEncoding.EncodeToString([]byte("fake-salt")),
		Nonce:      base64.RawURLEncoding.EncodeToString([]byte("fake-nonce")),
		Ciphertext: base64.RawURLEncoding.EncodeToString([]byte("fake-ciphertext")),
	}
}
