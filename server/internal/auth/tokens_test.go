package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func setTestAccessTokenSecret(t *testing.T) {
	t.Helper()

	original := accessTokenSecret

	if err := SetAccessTokenSecret(
		[]byte("0123456789abcdef0123456789abcdef"),
	); err != nil {
		t.Fatalf("SetAccessTokenSecret returned error: %v", err)
	}

	t.Cleanup(func() {
		accessTokenSecret = original
	})
}

func makeSignedTestToken(
	t *testing.T,
	header AccessTokenHeader,
	payload AccessTokenPayload,
) string {
	t.Helper()

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal test header: %v", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := encodedHeader + "." + encodedPayload

	return signingInput + "." + signAccessToken(signingInput)
}

// verifies that valid access tokens are accepted and invalid tokens are rejected
func TestVerifyAccessToken(t *testing.T) {
	setTestAccessTokenSecret(t)

	validHeader := AccessTokenHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}

	validPayload := AccessTokenPayload{
		Subject:   "67",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	tests := []struct {
		name    string
		token   func(t *testing.T) string
		wantErr error
	}{
		{
			name: "valid token",
			token: func(t *testing.T) string {
				return makeSignedTestToken(t, validHeader, validPayload)
			},
			wantErr: nil,
		},
		{
			name: "missing token parts",
			token: func(t *testing.T) string {
				return "invalid-token"
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "too many token parts",
			token: func(t *testing.T) string {
				return "one.two.three.four"
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "wrong algorithm",
			token: func(t *testing.T) string {
				header := validHeader
				header.Algorithm = "none"

				return makeSignedTestToken(t, header, validPayload)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "wrong token type",
			token: func(t *testing.T) string {
				header := validHeader
				header.Type = "NOT-JWT"

				return makeSignedTestToken(t, header, validPayload)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "non numeric user ID",
			token: func(t *testing.T) string {
				payload := validPayload
				payload.Subject = "abc"

				return makeSignedTestToken(t, validHeader, payload)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "zero user ID",
			token: func(t *testing.T) string {
				payload := validPayload
				payload.Subject = "0"

				return makeSignedTestToken(t, validHeader, payload)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "negative user ID",
			token: func(t *testing.T) string {
				payload := validPayload
				payload.Subject = "-1"

				return makeSignedTestToken(t, validHeader, payload)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "expired token",
			token: func(t *testing.T) string {
				payload := validPayload
				payload.ExpiresAt = time.Now().Add(-time.Second).Unix()

				return makeSignedTestToken(t, validHeader, payload)
			},
			wantErr: ErrExpiredToken,
		},
		{
			name: "invalid header base64",
			token: func(t *testing.T) string {
				payloadJSON, err := json.Marshal(validPayload)
				if err != nil {
					t.Fatalf("marshal payload: %v", err)
				}

				encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
				signingInput := "not!base64." + encodedPayload

				return signingInput + "." + signAccessToken(signingInput)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "invalid payload base64",
			token: func(t *testing.T) string {
				headerJSON, err := json.Marshal(validHeader)
				if err != nil {
					t.Fatalf("marshal header: %v", err)
				}

				encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
				signingInput := encodedHeader + ".not!base64"

				return signingInput + "." + signAccessToken(signingInput)
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "tampered signature",
			token: func(t *testing.T) string {
				token := makeSignedTestToken(
					t,
					validHeader,
					validPayload,
				)

				parts := strings.Split(token, ".")
				parts[2] = "tampered"

				return strings.Join(parts, ".")
			},
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.token(t)

			claims, err := VerifyAccessToken(token)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.wantErr,
						err,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if claims.UserID != 67 {
				t.Fatalf(
					"expected user ID 67, got %d",
					claims.UserID,
				)
			}
		})
	}
}

// ensures that the SetAccessTokenSecret function validates the secret length
func TestSetAccessTokenSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  []byte
		wantErr bool
	}{
		{
			name:    "exactly 32 bytes",
			secret:  make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "more than 32 bytes",
			secret:  make([]byte, 64),
			wantErr: false,
		},
		{
			name:    "31 bytes",
			secret:  make([]byte, 31),
			wantErr: true,
		},
		{
			name:    "empty secret",
			secret:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := accessTokenSecret

			t.Cleanup(func() {
				accessTokenSecret = original
			})

			err := SetAccessTokenSecret(tt.secret)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// ensures that the SetAccessTokenSecret function copies the input secret
func TestSetAccessTokenSecretCopiesInput(t *testing.T) {
	original := accessTokenSecret
	t.Cleanup(func() {
		accessTokenSecret = original
	})

	secret := []byte("0123456789abcdef0123456789abcdef")

	if err := SetAccessTokenSecret(secret); err != nil {
		t.Fatalf("SetAccessTokenSecret returned error: %v", err)
	}

	secret[0] = 'X'

	if accessTokenSecret[0] == 'X' {
		t.Fatal("expected access token secret to be copied")
	}
}

// ensures that the GenerateAccessTokenSecret function generates a 32-byte secret
func TestGenerateAccessTokenSecret(t *testing.T) {
	secret, err := GenerateAccessTokenSecret()
	if err != nil {
		t.Fatalf("GenerateAccessTokenSecret returned error: %v", err)
	}

	if len(secret) != 32 {
		t.Fatalf("expected 32-byte secret, got %d bytes", len(secret))
	}
}

// ensures that the GenerateAccessTokenSecret function generates unique secrets
func TestGenerateAccessTokenSecretProducesUniqueSecrets(t *testing.T) {
	first, err := GenerateAccessTokenSecret()
	if err != nil {
		t.Fatalf("first GenerateAccessTokenSecret returned error: %v", err)
	}

	second, err := GenerateAccessTokenSecret()
	if err != nil {
		t.Fatalf("second GenerateAccessTokenSecret returned error: %v", err)
	}

	if bytes.Equal(first, second) {
		t.Fatal("expected generated secrets to differ")
	}
}

// ensures that the IssueAccessToken and VerifyAccessToken functions work together correctly
func TestIssueAndVerifyAccessToken(t *testing.T) {
	setTestAccessTokenSecret(t)

	const userID int64 = 1
	const ttl = time.Hour

	beforeIssue := time.Now().UTC()

	token, err := IssueAccessToken(userID, ttl)
	if err != nil {
		t.Fatalf("IssueAccessToken returned error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty access token")
	}

	claims, err := VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken returned error: %v", err)
	}

	afterVerify := time.Now().UTC()

	if claims.UserID != userID {
		t.Fatalf(
			"expected user ID %d, got %d",
			userID,
			claims.UserID,
		)
	}

	if claims.IssuedAt.Before(beforeIssue.Add(-time.Second)) {
		t.Fatalf(
			"issued-at time is unexpectedly early: %v",
			claims.IssuedAt,
		)
	}

	if claims.IssuedAt.After(afterVerify) {
		t.Fatalf(
			"issued-at time is in the future: %v",
			claims.IssuedAt,
		)
	}

	expectedExpiry := claims.IssuedAt.Add(ttl)

	if !claims.ExpiresAt.Equal(expectedExpiry) {
		t.Fatalf(
			"expected expiry %v, got %v",
			expectedExpiry,
			claims.ExpiresAt,
		)
	}
}

// ensures that the IssueAccessToken function rejects missing secret
func TestIssueAccessTokenRejectsMissingSecret(t *testing.T) {
	original := accessTokenSecret
	accessTokenSecret = nil

	t.Cleanup(func() {
		accessTokenSecret = original
	})

	token, err := IssueAccessToken(1, time.Hour)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}

	const want = "access token secret not configured"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures that the VerifyAccessToken function rejects missing secret
func TestVerifyAccessTokenRejectsMissingSecret(t *testing.T) {
	original := accessTokenSecret
	accessTokenSecret = nil

	t.Cleanup(func() {
		accessTokenSecret = original
	})

	_, err := VerifyAccessToken("header.payload.signature")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	const want = "access token secret not configured"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}
