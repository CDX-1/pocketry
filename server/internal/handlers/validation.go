package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

const (
	maxAuthBodyBytes  = 1 << 20  // 1 MB
	maxVaultBodyBytes = 10 << 20 // 10 MB
)

var rxUsername = regexp.MustCompile(`^[a-zA-Z0-9]+$`) // alphanumeric
var rxPassword = regexp.MustCompile(`^[\x20-\x7E]+$`) // printable char on std keyboard

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func validateAuthRequest(req *AuthRequest) error {
	req.Username = strings.TrimSpace(req.Username)

	// username lengtyh
	uLen := len(req.Username)
	if uLen < 3 || uLen > 24 {
		return errors.New("username must be between 3 and 24 characters")
	}

	// username regex (alphanumeric)
	if !rxUsername.MatchString(req.Username) {
		return errors.New("username must be alphanumeric")
	}

	// password length
	pLen := len(req.Password)
	if pLen < 12 || pLen > 1024 {
		return errors.New("password must be between 12 and 1024 characters")
	}
	
	// password regex
	if !rxPassword.MatchString(req.Password) {
		return errors.New("password must only contain letters, numbers, and standard symbols")
	}

	return nil
}

func validateVaultRequest(req *VaultRequest) error {
	req.EncryptedBlob = strings.TrimSpace(req.EncryptedBlob)

	if req.EncryptedBlob == "" {
		return errors.New("encrypted_blob is required")
	}

	if len(req.EncryptedBlob) > maxVaultBodyBytes {
		return errors.New("encrypted_blob is too large")
	}

	if req.ExpectedRevision < 0 {
		return errors.New("expected_revision cannot be negative")
	}

	return nil
}