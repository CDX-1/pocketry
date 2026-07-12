package vault

import (
	"strings"
	"testing"
)

// verifies that envelope validation fully validates every parameter
func TestValidateEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(*Envelope)
		wantErr     bool
		wantMessage string
	}{
		{
			name:    "valid envelope",
			modify:  func(env *Envelope) {},
			wantErr: false,
		},
		{
			name: "unsupported version",
			modify: func(env *Envelope) {
				env.Version = 2
			},
			wantErr:     true,
			wantMessage: "unsupported vault envelope version",
		},
		{
			name: "unsupported cipher",
			modify: func(env *Envelope) {
				env.Cipher = "aes-256-gcm"
			},
			wantErr:     true,
			wantMessage: "unsupported vault cipher",
		},
		{
			name: "unsupported KDF",
			modify: func(env *Envelope) {
				env.KDF = "scrypt"
			},
			wantErr:     true,
			wantMessage: "unsupported vault KDF",
		},
		{
			name: "zero KDF memory",
			modify: func(env *Envelope) {
				env.KDFParams.Memory = 0
			},
			wantErr:     true,
			wantMessage: "invalid KDF parameters",
		},
		{
			name: "zero KDF iterations",
			modify: func(env *Envelope) {
				env.KDFParams.Iterations = 0
			},
			wantErr:     true,
			wantMessage: "invalid KDF parameters",
		},
		{
			name: "zero KDF parallelism",
			modify: func(env *Envelope) {
				env.KDFParams.Parallelism = 0
			},
			wantErr:     true,
			wantMessage: "invalid KDF parameters",
		},
		{
			name: "missing salt",
			modify: func(env *Envelope) {
				env.Salt = ""
			},
			wantErr:     true,
			wantMessage: "salt is required",
		},
		{
			name: "invalid salt base64url",
			modify: func(env *Envelope) {
				env.Salt = "not!valid"
			},
			wantErr:     true,
			wantMessage: "salt must be base64url encoded",
		},
		{
			name: "missing nonce",
			modify: func(env *Envelope) {
				env.Nonce = ""
			},
			wantErr:     true,
			wantMessage: "nonce is required",
		},
		{
			name: "invalid nonce base64url",
			modify: func(env *Envelope) {
				env.Nonce = "not!valid"
			},
			wantErr:     true,
			wantMessage: "nonce must be base64url encoded",
		},
		{
			name: "missing ciphertext",
			modify: func(env *Envelope) {
				env.Ciphertext = ""
			},
			wantErr:     true,
			wantMessage: "ciphertext is required",
		},
		{
			name: "invalid ciphertext base64url",
			modify: func(env *Envelope) {
				env.Ciphertext = "not!valid"
			},
			wantErr:     true,
			wantMessage: "ciphertext must be base64url encoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validTestEnvelope()
			tt.modify(&env)

			err := ValidateEnvelope(env)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantMessage {
				t.Fatalf(
					"expected error %q, got %q",
					tt.wantMessage,
					err.Error(),
				)
			}
		})
	}
}

// verifies that base64URL validation functions as expected
func TestValidateBase64URL(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		field       string
		wantErr     bool
		wantMessage string
	}{
		{
			name:    "valid encoded value",
			value:   "ZmFrZS12YWx1ZQ",
			field:   "salt",
			wantErr: false,
		},
		{
			name:    "valid URL safe characters",
			value:   "abc-DEF_123",
			field:   "nonce",
			wantErr: false,
		},
		{
			name:        "empty value",
			value:       "",
			field:       "salt",
			wantErr:     true,
			wantMessage: "salt is required",
		},
		{
			name:        "standard base64 padding rejected",
			value:       "ZmFrZQ==",
			field:       "nonce",
			wantErr:     true,
			wantMessage: "nonce must be base64url encoded",
		},
		{
			name:        "invalid punctuation",
			value:       "not!valid",
			field:       "ciphertext",
			wantErr:     true,
			wantMessage: "ciphertext must be base64url encoded",
		},
		{
			name:        "spaces rejected",
			value:       "abc def",
			field:       "salt",
			wantErr:     true,
			wantMessage: "salt must be base64url encoded",
		},
		{
			name:        "newline rejected",
			value:       "abc\ndef",
			field:       "nonce",
			wantErr:     true,
			wantMessage: "nonce must be base64url encoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBase64URL(tt.value, tt.field)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantMessage {
				t.Fatalf(
					"expected error %q, got %q",
					tt.wantMessage,
					err.Error(),
				)
			}
		})
	}
}

// verifies that ValidateEnvelope stops processing at the first encountered validation error
func TestValidateEnvelopeStopsAtFirstError(t *testing.T) {
	env := validTestEnvelope()

	env.Version = 99
	env.Cipher = "wrong"
	env.KDF = "wrong"
	env.Salt = ""

	err := ValidateEnvelope(env)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "unsupported vault envelope version" {
		t.Fatalf(
			"expected first validation error %q, got %q",
			"unsupported vault envelope version",
			err.Error(),
		)
	}
}

// verifies that ValidateBase64URL includes the field name in the error message
func TestValidateBase64URLErrorUsesFieldName(t *testing.T) {
	field := "custom_field"

	err := validateBase64URL("", field)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), field) {
		t.Fatalf(
			"expected error to contain field name %q, got %q",
			field,
			err.Error(),
		)
	}
}