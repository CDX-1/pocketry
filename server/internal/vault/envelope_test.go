package vault

import (
	"encoding/base64"
	"testing"
)

func validTestEnvelope() Envelope {
	return Envelope{
		Version: SupportedEnvelopeVersion,
		Cipher:  SupportedCipher,
		KDF:     SupportedKDF,
		KDFParams: KDFParams{
			Memory:      65536,
			Iterations:  3,
			Parallelism: 2,
		},
		Salt:       base64.RawURLEncoding.EncodeToString([]byte("fake-salt")),
		Nonce:      base64.RawURLEncoding.EncodeToString([]byte("fake-nonce")),
		Ciphertext: base64.RawURLEncoding.EncodeToString([]byte("fake-ciphertext")),
	}
}

// check envelope validation
func TestValidateEnvelopeAcceptsValidEnvelope(t *testing.T) {
	env := validTestEnvelope()

	if err := ValidateEnvelope(env); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}
}

// check envelope version validation
func TestValidateEnvelopeRejectsUnsupportedVersion(t *testing.T) {
	env := validTestEnvelope()
	env.Version = 999

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for unsupported envelope version")
	}
}

// check envelope cipher validation

func TestValidateEnvelopeRejectsUnsupportedCipher(t *testing.T) {
	env := validTestEnvelope()
	env.Cipher = "aes-256-gcm"

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for unsupported cipher")
	}
}

// check envelope KDF validation
func TestValidateEnvelopeRejectsUnsupportedKDF(t *testing.T) {
	env := validTestEnvelope()
	env.KDF = "pbkdf2"

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for unsupported kdf")
	}
}

// check missing KDF memory parameter
func TestValidateEnvelopeRejectsMissingKDFMemory(t *testing.T) {
	env := validTestEnvelope()
	env.KDFParams.Memory = 0

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing kdf memory")
	}
}

// check missing KDF iterations parameter
func TestValidateEnvelopeRejectsMissingKDFIterations(t *testing.T) {
	env := validTestEnvelope()
	env.KDFParams.Iterations = 0

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing kdf iterations")
	}
}

// check missing KDF parallelism parameter
func TestValidateEnvelopeRejectsMissingKDFParallelism(t *testing.T) {
	env := validTestEnvelope()
	env.KDFParams.Parallelism = 0

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing kdf parallelism")
	}
}

// check salt validation
func TestValidateEnvelopeRejectsInvalidSaltBase64(t *testing.T) {
	env := validTestEnvelope()
	env.Salt = "not base64"

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for invalid salt base64")
	}
}

// check nonce validation
func TestValidateEnvelopeRejectsInvalidNonceBase64(t *testing.T) {
	env := validTestEnvelope()
	env.Nonce = "not base64"

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for invalid nonce base64")
	}
}


// check ciphertext validation
func TestValidateEnvelopeRejectsInvalidCiphertextBase64(t *testing.T) {
	env := validTestEnvelope()
	env.Ciphertext = "not base64!!!"

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for invalid ciphertext base64")
	}
}

// check missing salt
func TestValidateEnvelopeRejectsMissingSalt(t *testing.T) {
	env := validTestEnvelope()
	env.Salt = ""

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing salt")
	}
}

// check missing nonce
func TestValidateEnvelopeRejectsMissingNonce(t *testing.T) {
	env := validTestEnvelope()
	env.Nonce = ""

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing nonce")
	}
}

// check missing ciphertext
func TestValidateEnvelopeRejectsMissingCiphertext(t *testing.T) {
	env := validTestEnvelope()
	env.Ciphertext = ""

	if err := ValidateEnvelope(env); err == nil {
		t.Fatal("expected error for missing ciphertext")
	}
}