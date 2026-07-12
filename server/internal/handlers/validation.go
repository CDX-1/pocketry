package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/CDX-1/pocketry/internal/vault"
)

const (
	maxAuthBodyBytes  = 1 << 20  // 1 MB
	maxVaultBodyBytes = 10 << 20 // 10 MB
)

var rxUsername = regexp.MustCompile(`^[a-zA-Z0-9]+$`) // alphanumeric

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	uLen := len(username)
	if uLen < 3 || uLen > 24 {
		return errors.New("username must be between 3 and 24 characters")
	}

	if !rxUsername.MatchString(username) {
		return errors.New("username must be alphanumeric")
	}

	return nil
}

func validateOpaqueClientMessage(clientMessage string) error {
	if strings.TrimSpace(clientMessage) == "" {
		return errors.New("client_message is required")
	}

	return nil
}

func validateFlowID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("flow id is required")
	}

	return nil
}

func validateVaultRequest(req *VaultRequest) error {
	if req.ExpectedRevision < 0 {
		return errors.New("expected_revision cannot be negative")
	}

	if err := vault.ValidateEnvelope(req.EncryptedBlob); err != nil {
		return err
	}

	return nil
}