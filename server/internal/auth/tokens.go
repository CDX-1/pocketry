package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type AccessTokenClaims struct {
	UserID	  int64
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type AccessTokenPayload struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type AccessTokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

var accessTokenSecret []byte

func SetAccessTokenSecret(secret []byte) error {
	if len(secret) < 32 {
		return fmt.Errorf("access token secret must be at least 32 bytes")
	}

	accessTokenSecret = append([]byte(nil), secret...)
	return nil
}

func GenerateAccessTokenSecret() ([]byte, error) {
	secret := make([]byte, 32)

	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate access token secret: %w", err)
	}

	return secret, nil
}

func IssueAccessToken(userID int64, ttl time.Duration) (string, error) {
	if len(accessTokenSecret) == 0 {
		return "", errors.New("access token secret not configured")
	}

	now := time.Now().UTC()

	header := AccessTokenHeader{
		Algorithm: "HS256",
		Type:	   "JWT",
	}

	payload := AccessTokenPayload{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := encodedHeader + "." + encodedPayload
	signature := signAccessToken(signingInput)

	return signingInput + "." + signature, nil
}

func VerifyAccessToken(token string) (AccessTokenClaims, error) {
	if len(accessTokenSecret) == 0 {
		return AccessTokenClaims{}, errors.New("access token secret not configured")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signAccessToken(signingInput)

	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedSignature)) != 1 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	var header AccessTokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	var payload AccessTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(payload.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	now := time.Now().UTC()
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()

	if !now.Before(expiresAt) {
		return AccessTokenClaims{}, ErrExpiredToken
	}

	return AccessTokenClaims{
		UserID:    userID,
		IssuedAt:  time.Unix(payload.IssuedAt, 0).UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func signAccessToken(signingInput string) string {
	mac := hmac.New(sha256.New, accessTokenSecret)
	mac.Write([]byte(signingInput))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}