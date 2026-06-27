package vault

import (
	"encoding/base64"
	"errors"
)

const (
	SupportedEnvelopeVersion = 1
	SupportedCipher          = "xchacha20-poly1305"
	SupportedKDF             = "argon2id"
)

func ValidateEnvelope(env Envelope) error {
	if env.Version != SupportedEnvelopeVersion {
		return errors.New("unsupported vault envelope version")
	}

	if env.Cipher != SupportedCipher {
		return errors.New("unsupported vault cipher")
	}

	if env.KDF != SupportedKDF {
		return errors.New("unsupported vault KDF")
	}

	if env.KDFParams.Memory == 0 || env.KDFParams.Iterations == 0 || env.KDFParams.Parallelism == 0 {
		return errors.New("invalid KDF parameters")
	}

	if err := validateBase64URL(env.Salt, "salt"); err != nil {
		return err
	}

	if err := validateBase64URL(env.Nonce, "nonce"); err != nil {
		return err
	}

	if err := validateBase64URL(env.Ciphertext, "ciphertext"); err != nil {
		return err
	}

	return nil
}

func validateBase64URL(value string, field string) error {
	if value == "" {
		return errors.New(field + " is required")
	}

	if _, err := base64.RawURLEncoding.DecodeString(value); err != nil {
		return errors.New(field + " must be base64url encoded")
	}

	return nil
}